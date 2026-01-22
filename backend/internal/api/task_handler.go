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
	repo *repository.TaskRepository
	hub  *websocket.Hub
}

func NewTaskHandler(db *gorm.DB, hub *websocket.Hub) *TaskHandler {
	return &TaskHandler{
		repo: repository.NewTaskRepository(db),
		hub:  hub,
	}
}

type CreateTaskRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	WorkspaceID uuid.UUID `json:"workspace_id" binding:"required"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	Version     int    `json:"version" binding:"required"`
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	workspaceIDStr := c.Query("workspace_id")
	if workspaceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}

	// TODO: Get user ID from JWT token
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tasks, err := h.repo.GetByWorkspace(workspaceID, 100, 0)
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

	task, err := h.repo.GetByID(id)
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
		WorkspaceID: req.WorkspaceID,
		CreatedBy:   userID,
	}

	if req.Status == "" {
		task.Status = "todo"
	}

	if err := h.repo.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast to WebSocket clients
	h.hub.Broadcast(&websocket.Message{
		Type:        "TASK_CREATED",
		WorkspaceID: req.WorkspaceID,
		TaskID:      &task.ID,
		Data:        task,
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

	task, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Store old value for update log
	oldValue, _ := json.Marshal(task)

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

	if err := h.repo.Update(task); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Log update
	// TODO: Get user ID from JWT token
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	newValue, _ := json.Marshal(task)
	update := &database.TaskUpdate{
		TaskID:     task.ID,
		UserID:     userID,
		ChangeType: "updated",
		OldValue:   string(oldValue),
		NewValue:   string(newValue),
	}
	h.repo.LogUpdate(update)

	// Broadcast to WebSocket clients
	h.hub.Broadcast(&websocket.Message{
		Type:        "TASK_UPDATED",
		WorkspaceID: task.WorkspaceID,
		TaskID:      &task.ID,
		Data:        task,
	})

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast to WebSocket clients
	h.hub.Broadcast(&websocket.Message{
		Type:        "TASK_DELETED",
		WorkspaceID: task.WorkspaceID,
		TaskID:      &id,
		Data:        nil,
	})

	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}
