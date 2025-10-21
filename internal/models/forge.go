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

// ForgeConfidenceTracker tracks confidence growth for apprenticeship patterns
type ForgeConfidenceTracker struct {
	ID                uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PatternName       string    `gorm:"not null"`
	ConfidenceScore   float64
	ObservationsCount int `gorm:"default:0"`
	LastUpdated       time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// BeforeCreate hook
func (fct *ForgeConfidenceTracker) BeforeCreate(tx *gorm.DB) error {
	if fct.ID == uuid.Nil {
		fct.ID = uuid.New()
	}
	return nil
}
