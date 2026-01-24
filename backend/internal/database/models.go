package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project represents a project that contains tasks
type Project struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Task represents a task within a project
type Task struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Status      string    `gorm:"default:'todo'" json:"status"` // todo, in_progress, done, blocked
	AssigneeID  *uuid.UUID `gorm:"type:uuid" json:"assignee_id"`
	ProjectID   uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	Version     int       `gorm:"default:1" json:"version"` // For optimistic locking
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Dependencies []TaskDependency `gorm:"foreignKey:TaskID" json:"dependencies,omitempty"`
	Comments     []Comment        `gorm:"foreignKey:TaskID" json:"comments,omitempty"`
}

// TaskDependency represents a dependency relationship between tasks
type TaskDependency struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID       uuid.UUID `gorm:"type:uuid;not null;index" json:"task_id"`
	DependsOnID  uuid.UUID `gorm:"type:uuid;not null;index" json:"depends_on_id"`
	CreatedAt    time.Time `json:"created_at"`
	
	// Relationships
	Task      Task `gorm:"foreignKey:TaskID" json:"-"`
	DependsOn Task `gorm:"foreignKey:DependsOnID" json:"depends_on,omitempty"`
}

// Comment represents a comment on a task
type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID    uuid.UUID `gorm:"type:uuid;not null;index" json:"task_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Content   string    `gorm:"not null" json:"content"`
	ParentID  *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"` // For threaded comments
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TaskUpdate represents an audit log of task changes
type TaskUpdate struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID      uuid.UUID `gorm:"type:uuid;not null;index" json:"task_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ChangeType  string    `gorm:"not null" json:"change_type"` // created, updated, deleted, status_changed
	OldValue    string    `gorm:"type:jsonb" json:"old_value"`
	NewValue    string    `gorm:"type:jsonb" json:"new_value"`
	Timestamp   time.Time `gorm:"default:now()" json:"timestamp"`
}

// Valid status transitions
var ValidStatusTransitions = map[string][]string{
	"todo":       {"in_progress", "blocked"},
	"in_progress": {"done", "blocked", "todo"},
	"done":       {}, // Terminal state
	"blocked":    {"todo", "in_progress"},
}

// IsValidStatusTransition checks if a status transition is valid
func IsValidStatusTransition(from, to string) bool {
	allowed, exists := ValidStatusTransitions[from]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}

// BeforeCreate hooks to generate UUIDs
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (td *TaskDependency) BeforeCreate(tx *gorm.DB) error {
	if td.ID == uuid.Nil {
		td.ID = uuid.New()
	}
	return nil
}

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (tu *TaskUpdate) BeforeCreate(tx *gorm.DB) error {
	if tu.ID == uuid.Nil {
		tu.ID = uuid.New()
	}
	return nil
}
