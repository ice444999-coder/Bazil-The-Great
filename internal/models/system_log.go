package models

import (
	"time"
)

// SystemLog represents a centralized log entry for all services
type SystemLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Service   string    `gorm:"size:50;not null;index" json:"service"`
	Level     string    `gorm:"size:10;not null" json:"level"` // DEBUG, INFO, WARN, ERROR
	Message   string    `gorm:"type:text;not null" json:"message"`
	EventType string    `gorm:"size:50" json:"event_type,omitempty"`
	EventData string    `gorm:"type:text" json:"event_data,omitempty"` // JSON string
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}
