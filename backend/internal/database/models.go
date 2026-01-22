package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Status      string    `gorm:"default:'todo'" json:"status"` // todo, in_progress, done
	AssigneeID  *uuid.UUID `gorm:"type:uuid" json:"assignee_id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"workspace_id"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	Version     int       `gorm:"default:1" json:"version"` // For optimistic locking
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type TaskUpdate struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID      uuid.UUID `gorm:"type:uuid;not null;index" json:"task_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ChangeType  string    `gorm:"not null" json:"change_type"` // created, updated, deleted, status_changed
	OldValue    string    `gorm:"type:jsonb" json:"old_value"`
	NewValue    string    `gorm:"type:jsonb" json:"new_value"`
	Timestamp   time.Time `gorm:"default:now()" json:"timestamp"`
}

// BeforeCreate hook to generate UUID
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (tu *TaskUpdate) BeforeCreate(tx *gorm.DB) error {
	if tu.ID == uuid.Nil {
		tu.ID = uuid.New()
	}
	return nil
}
