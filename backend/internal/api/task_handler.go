package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/task-manager/internal/database"
	"github.com/yourusername/task-manager/internal/repository"
	"github.com/yourusername/task-manager/internal/websocket"
	"gorm.io/gorm"
)

type TaskHandler struct {
	taskRepo       *repository.TaskRepository
	dependencyRepo *repository.DependencyRepository
	hub            *websocket.Hub
}

func NewTaskHandler(db *gorm.DB, hub *websocket.Hub) *TaskHandler {
	return &TaskHandler{
		taskRepo:       repository.NewTaskRepository(db),
		dependencyRepo: repository.NewDependencyRepository(db),
		hub:            hub,
	}
}

type CreateTaskRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ProjectID   uuid.UUID `json:"project_id" binding:"required"`
}

type UpdateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	Version     int        `json:"version" binding:"required"`
}

type AddDependencyRequest struct {
	DependsOnID uuid.UUID `json:"depends_on_id" binding:"required"`
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	tasks, err := h.taskRepo.GetByProject(projectID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := h.taskRepo.GetByIDWithRelations(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user ID from JWT token
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	task := &database.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		ProjectID:   req.ProjectID,
		CreatedBy:   userID,
	}

	if req.Status == "" {
		task.Status = "todo"
	}

	if err := h.taskRepo.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast to WebSocket clients with full data for creates
	h.hub.Broadcast(&websocket.Message{
		Type:      "TASK_CREATED",
		ProjectID: req.ProjectID,
		TaskID:    &task.ID,
		FullData:  task,
	})

	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Validate status transition if status is being changed
	if req.Status != "" && req.Status != task.Status {
		if !database.IsValidStatusTransition(task.Status, req.Status) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid status transition",
				"from":  task.Status,
				"to":    req.Status,
			})
			return
		}

		// Check dependencies if transitioning to "done"
		if req.Status == "done" {
			canTransition, err := h.dependencyRepo.CanTransitionToStatus(task.ID, req.Status)
			if err != nil || !canTransition {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// Store old value for delta computation
	oldTask := *task

	// Update fields
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	task.Version = req.Version

	if err := h.taskRepo.Update(task); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Log update
	// TODO: Get user ID from JWT token
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	oldValue, _ := json.Marshal(oldTask)
	newValue, _ := json.Marshal(task)
	update := &database.TaskUpdate{
		TaskID:     task.ID,
		UserID:     userID,
		ChangeType: "updated",
		OldValue:   string(oldValue),
		NewValue:   string(newValue),
	}
	h.taskRepo.LogUpdate(update)

	// Broadcast delta update (only changed fields)
	delta, _ := ComputeDelta(&oldTask, task)
	h.hub.Broadcast(&websocket.Message{
		Type:      "TASK_UPDATED",
		ProjectID: task.ProjectID,
		TaskID:    &task.ID,
		Delta:     delta,
	})

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := h.taskRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast to WebSocket clients
	h.hub.Broadcast(&websocket.Message{
		Type:      "TASK_DELETED",
		ProjectID: task.ProjectID,
		TaskID:    &id,
	})

	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}

func (h *TaskHandler) GetDependencies(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	dependencies, err := h.dependencyRepo.GetByTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dependencies)
}

func (h *TaskHandler) AddDependency(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var req AddDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get task to get project ID
	task, err := h.taskRepo.GetByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	dependency := &database.TaskDependency{
		TaskID:      taskID,
		DependsOnID: req.DependsOnID,
	}

	if err := h.dependencyRepo.Create(dependency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast dependency addition
	h.hub.Broadcast(&websocket.Message{
		Type:      "TASK_DEPENDENCY_ADDED",
		ProjectID: task.ProjectID,
		TaskID:    &taskID,
		FullData:  dependency,
	})

	c.JSON(http.StatusCreated, dependency)
}

func (h *TaskHandler) RemoveDependency(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	dependsOnID, err := uuid.Parse(c.Param("depends_on_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid depends_on_id"})
		return
	}

	// Get task to get project ID
	task, err := h.taskRepo.GetByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	if err := h.dependencyRepo.Delete(taskID, dependsOnID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast dependency removal
	h.hub.Broadcast(&websocket.Message{
		Type:      "TASK_DEPENDENCY_REMOVED",
		ProjectID: task.ProjectID,
		TaskID:    &taskID,
		Delta:     map[string]interface{}{"depends_on_id": dependsOnID},
	})

	c.JSON(http.StatusOK, gin.H{"message": "dependency removed"})
}
