/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package models

import (
	"time"
)

// RepoFileCache represents cached information about repository files
type RepoFileCache struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	FilePath      string    `gorm:"uniqueIndex;not null" json:"file_path"`
	FileType      string    `gorm:"size:10;not null" json:"file_type"`
	ContentHash   string    `gorm:"size:64;not null" json:"content_hash"`
	LineCount     int       `gorm:"not null" json:"line_count"`
	LastInspected time.Time `gorm:"not null" json:"last_inspected"`
	LastModified  time.Time `gorm:"not null" json:"last_modified"`
	SizeBytes     int64     `gorm:"not null" json:"size_bytes"`
	IsTracked     bool      `gorm:"default:true" json:"is_tracked"`
	Metadata      string    `gorm:"type:jsonb" json:"metadata,omitempty"` // JSON string
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}
