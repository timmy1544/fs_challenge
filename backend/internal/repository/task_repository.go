package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/task-manager/internal/database"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *database.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) GetByID(id uuid.UUID) (*database.Task, error) {
	var task database.Task
	if err := r.db.Where("id = ?", id).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) GetByWorkspace(workspaceID uuid.UUID, limit, offset int) ([]database.Task, error) {
	var tasks []database.Task
	query := r.db.Where("workspace_id = ?", workspaceID)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Order("updated_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepository) Update(task *database.Task) error {
	// Optimistic locking: check version
	var currentTask database.Task
	if err := r.db.Where("id = ?", task.ID).First(&currentTask).Error; err != nil {
		return err
	}

	if currentTask.Version != task.Version {
		return fmt.Errorf("version conflict: task was modified by another user")
	}

	task.Version++
	return r.db.Model(task).Updates(map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"assignee_id": task.AssigneeID,
		"version":     task.Version,
		"updated_at":  gorm.Expr("NOW()"),
	}).Error
}

func (r *TaskRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&database.Task{}, id).Error
}

func (r *TaskRepository) LogUpdate(update *database.TaskUpdate) error {
	return r.db.Create(update).Error
}
