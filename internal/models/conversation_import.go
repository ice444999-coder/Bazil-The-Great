package models

import "time"

// ConversationImport tracks imported conversations
type ConversationImport struct {
	ID           uint      `gorm:"primaryKey"`
	Source       string    `gorm:"type:varchar(100);not null"` // "manual_paste", "file_upload", etc
	ImportedAt   time.Time `gorm:"autoCreateTime;not null;index"`
	MessageCount int       `gorm:"default:0"`
	Tags         []string  `gorm:"type:text[]"`
	Metadata     JSONB     `gorm:"type:jsonb"`
	UserID       uint      `gorm:"index"`
}
