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

type CommentHandler struct {
	commentRepo *repository.CommentRepository
	taskRepo    *repository.TaskRepository
	hub         *websocket.Hub
}

func NewCommentHandler(db *gorm.DB, hub *websocket.Hub) *CommentHandler {
	return &CommentHandler{
		commentRepo: repository.NewCommentRepository(db),
		taskRepo:    repository.NewTaskRepository(db),
		hub:         hub,
	}
}

type CreateCommentRequest struct {
	Content  string     `json:"content" binding:"required"`
	ParentID *uuid.UUID `json:"parent_id"` // For threaded comments
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *CommentHandler) GetComments(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	comments, err := h.commentRepo.GetByTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user ID from JWT token
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Get task to get project ID
	task, err := h.taskRepo.GetByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	comment := &database.Comment{
		TaskID:   taskID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := h.commentRepo.Create(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast comment creation with full data
	h.hub.Broadcast(&websocket.Message{
		Type:      "COMMENT_CREATED",
		ProjectID: task.ProjectID,
		TaskID:    &taskID,
		CommentID: &comment.ID,
		FullData:  comment,
	})

	c.JSON(http.StatusCreated, comment)
}

func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.commentRepo.GetByID(commentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	oldComment := *comment
	comment.Content = req.Content

	if err := h.commentRepo.Update(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get task to get project ID
	task, err := h.taskRepo.GetByID(comment.TaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "task not found"})
		return
	}

	// Broadcast delta update
	delta, _ := ComputeDelta(&oldComment, comment)
	h.hub.Broadcast(&websocket.Message{
		Type:      "COMMENT_UPDATED",
		ProjectID: task.ProjectID,
		TaskID:    &comment.TaskID,
		CommentID: &commentID,
		Delta:     delta,
	})

	c.JSON(http.StatusOK, comment)
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	comment, err := h.commentRepo.GetByID(commentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get task to get project ID
	task, err := h.taskRepo.GetByID(comment.TaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "task not found"})
		return
	}

	if err := h.commentRepo.Delete(commentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast comment deletion
	h.hub.Broadcast(&websocket.Message{
		Type:      "COMMENT_DELETED",
		ProjectID: task.ProjectID,
		TaskID:    &comment.TaskID,
		CommentID: &commentID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted"})
}
