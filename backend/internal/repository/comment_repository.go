package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/task-manager/internal/database"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *database.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) GetByID(id uuid.UUID) (*database.Comment, error) {
	var comment database.Comment
	if err := r.db.Preload("Replies").Where("id = ?", id).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) GetByTask(taskID uuid.UUID) ([]database.Comment, error) {
	var comments []database.Comment
	if err := r.db.Preload("Replies").
		Where("task_id = ? AND parent_id IS NULL", taskID).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *CommentRepository) Update(comment *database.Comment) error {
	return r.db.Model(comment).Updates(map[string]interface{}{
		"content":    comment.Content,
		"updated_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *CommentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&database.Comment{}, id).Error
}
