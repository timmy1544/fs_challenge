package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/task-manager/internal/database"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *database.Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepository) GetByID(id uuid.UUID) (*database.Project, error) {
	var project database.Project
	if err := r.db.Where("id = ?", id).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) GetByUser(userID uuid.UUID) ([]database.Project, error) {
	var projects []database.Project
	if err := r.db.Where("created_by = ?", userID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) Update(project *database.Project) error {
	return r.db.Model(project).Updates(map[string]interface{}{
		"name":        project.Name,
		"description": project.Description,
		"updated_at":   gorm.Expr("NOW()"),
	}).Error
}

func (r *ProjectRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&database.Project{}, id).Error
}
