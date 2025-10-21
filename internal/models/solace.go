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

// SolaceDecision represents autonomous decisions made by SOLACE
type SolaceDecision struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	DecisionType    string    `gorm:"not null"`
	Context         string    `gorm:"type:jsonb;not null"`
	Decision        string    `gorm:"not null"`
	ConfidenceScore float64
	Outcome         string    `gorm:"type:jsonb"`
	Embedding       []float64 `gorm:"type:vector(1536)"` // pgvector
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CognitivePattern represents learned patterns in SOLACE's cognition
type CognitivePattern struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Pattern     string    `gorm:"not null"`
	Confidence  float64
	Occurrences int `gorm:"default:0"`
	LastSeen    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BeforeCreate hooks
func (sd *SolaceDecision) BeforeCreate(tx *gorm.DB) error {
	if sd.ID == uuid.Nil {
		sd.ID = uuid.New()
	}
	return nil
}

func (cp *CognitivePattern) BeforeCreate(tx *gorm.DB) error {
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	return nil
}
