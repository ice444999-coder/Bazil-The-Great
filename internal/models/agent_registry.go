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

// AgentRegistry manages the agent swarm
type AgentRegistry struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name         string    `gorm:"uniqueIndex;not null"`
	Status       string    `gorm:"default:'idle'"` // idle, active, busy, offline
	LastActive   time.Time
	Capabilities string `gorm:"type:jsonb"` // JSON array of capabilities
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// BeforeCreate hooks
func (ar *AgentRegistry) BeforeCreate(tx *gorm.DB) error {
	if ar.ID == uuid.Nil {
		ar.ID = uuid.New()
	}
	return nil
}
