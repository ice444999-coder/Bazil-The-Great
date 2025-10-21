/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
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
