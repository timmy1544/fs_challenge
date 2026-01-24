package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/task-manager/internal/database"
	"github.com/yourusername/task-manager/internal/repository"
	"github.com/yourusername/task-manager/internal/websocket"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	repo *repository.ProjectRepository
	hub  *websocket.Hub
}

func NewProjectHandler(db *gorm.DB, hub *websocket.Hub) *ProjectHandler {
	return &ProjectHandler{
		repo: repository.NewProjectRepository(db),
		hub:  hub,
	}
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *ProjectHandler) GetProjects(c *gin.Context) {
	// TODO: Get user ID from JWT token
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	projects, err := h.repo.GetByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user ID from JWT token
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	project := &database.Project{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := h.repo.Create(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast project creation
	h.hub.Broadcast(&websocket.Message{
		Type:      "PROJECT_CREATED",
		ProjectID: project.ID,
		FullData:  project,
	})

	c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	oldProject := *project

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}

	if err := h.repo.Update(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast delta update
	delta, _ := ComputeDelta(&oldProject, project)
	h.hub.Broadcast(&websocket.Message{
		Type:      "PROJECT_UPDATED",
		ProjectID: project.ID,
		Delta:     delta,
	})

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast project deletion
	h.hub.Broadcast(&websocket.Message{
		Type:      "PROJECT_DELETED",
		ProjectID: id,
	})

	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}
