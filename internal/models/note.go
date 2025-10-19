package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Note represents a daily note/journal entry
type Note struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Title      string         `gorm:"type:varchar(255);not null" json:"title"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Tags       string         `gorm:"type:varchar(500)" json:"tags"` // Comma-separated tags
	IsFavorite bool           `gorm:"default:false" json:"is_favorite"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for Note model
func (Note) TableName() string {
	return "notes"
}
