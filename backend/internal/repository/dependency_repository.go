package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/task-manager/internal/database"
	"gorm.io/gorm"
)

type DependencyRepository struct {
	db *gorm.DB
}

func NewDependencyRepository(db *gorm.DB) *DependencyRepository {
	return &DependencyRepository{db: db}
}

func (r *DependencyRepository) Create(dependency *database.TaskDependency) error {
	// Check for circular dependencies
	if err := r.checkCircularDependency(dependency.TaskID, dependency.DependsOnID); err != nil {
		return err
	}
	return r.db.Create(dependency).Error
}

func (r *DependencyRepository) checkCircularDependency(taskID, dependsOnID uuid.UUID) error {
	// Simple check: if dependsOnID depends on taskID, it's circular
	var existing database.TaskDependency
	if err := r.db.Where("task_id = ? AND depends_on_id = ?", dependsOnID, taskID).First(&existing).Error; err == nil {
		return fmt.Errorf("circular dependency detected")
	}
	return nil
}

func (r *DependencyRepository) GetByTask(taskID uuid.UUID) ([]database.TaskDependency, error) {
	var dependencies []database.TaskDependency
	if err := r.db.Preload("DependsOn").Where("task_id = ?", taskID).Find(&dependencies).Error; err != nil {
		return nil, err
	}
	return dependencies, nil
}

func (r *DependencyRepository) Delete(taskID, dependsOnID uuid.UUID) error {
	result := r.db.Where("task_id = ? AND depends_on_id = ?", taskID, dependsOnID).Delete(&database.TaskDependency{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("dependency not found")
	}
	return nil
}

func (r *DependencyRepository) CanTransitionToStatus(taskID uuid.UUID, newStatus string) (bool, error) {
	if newStatus == "done" {
		// Check if all dependencies are done
		var dependencies []database.TaskDependency
		if err := r.db.Preload("DependsOn").Where("task_id = ?", taskID).Find(&dependencies).Error; err != nil {
			return false, err
		}

		for _, dep := range dependencies {
			if dep.DependsOn.Status != "done" {
				return false, fmt.Errorf("cannot mark task as done: dependency '%s' is not done", dep.DependsOn.Title)
			}
		}
	}
	return true, nil
}
