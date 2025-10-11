package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatMessage stores all chat messages for persistence and history loading
type ChatMessage struct {
	MessageID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"message_id"`
	SessionID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"session_id"`
	Timestamp         time.Time      `gorm:"not null;default:NOW();index:idx_chat_timestamp,sort:desc" json:"timestamp"`
	UserMessage       string         `gorm:"type:text;not null" json:"user_message"`
	AssistantResponse string         `gorm:"type:text" json:"assistant_response,omitempty"`
	Context           map[string]any `gorm:"type:jsonb" json:"context,omitempty"`
	UserID            *int           `gorm:"index" json:"user_id,omitempty"`
	CreatedAt         time.Time      `gorm:"default:NOW()" json:"created_at"`
}

// TableName override
func (ChatMessage) TableName() string {
	return "chat_messages"
}
