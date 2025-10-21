/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ToolRegistry represents a registered AI tool
type ToolRegistry struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ToolName    string    `gorm:"uniqueIndex;not null"`
	Description string
	Category    string
	Embedding   []float64 `gorm:"type:vector(1536)"` // pgvector for semantic search
	IsActive    bool      `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ToolPermission manages agent access to tools
type ToolPermission struct {
	ID                   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ToolID               uuid.UUID `gorm:"not null"`
	AgentName            string    `gorm:"not null"`
	DailyUsageLimit      int       `gorm:"default:100"`
	CurrentUsage         int       `gorm:"default:0"`
	CircuitBreakerActive bool      `gorm:"default:false"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	// Tool              ToolRegistry `gorm:"foreignKey:ToolID"` // Temporarily disabled
}

// ToolExecutionLog tracks tool usage for billing and analytics
type ToolExecutionLog struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ToolID          uuid.UUID `gorm:"not null"`
	AgentName       string    `gorm:"not null"`
	Success         bool      `gorm:"default:true"`
	ExecutionTimeMs int64
	CostUSD         float64
	ExecutedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	// Tool           ToolRegistry `gorm:"foreignKey:ToolID"` // Temporarily disabled
}

// BeforeCreate hook for ToolRegistry
func (tr *ToolRegistry) BeforeCreate(tx *gorm.DB) error {
	if tr.ID == uuid.Nil {
		tr.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for ToolPermission
func (tp *ToolPermission) BeforeCreate(tx *gorm.DB) error {
	if tp.ID == uuid.Nil {
		tp.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for ToolExecutionLog
func (tel *ToolExecutionLog) BeforeCreate(tx *gorm.DB) error {
	if tel.ID == uuid.Nil {
		tel.ID = uuid.New()
	}
	return nil
}
