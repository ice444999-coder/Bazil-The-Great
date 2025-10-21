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

// GRPOBias represents learned token biases for reinforcement learning
type GRPOBias struct {
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Token   string    `gorm:"not null"`
	Bias    float64
	Updated time.Time
}

// GRPOMetric tracks GRPO learning performance
type GRPOMetric struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	MetricName string    `gorm:"not null"`
	Value      float64
	Timestamp  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// BeforeCreate hooks
func (gb *GRPOBias) BeforeCreate(tx *gorm.DB) error {
	if gb.ID == uuid.Nil {
		gb.ID = uuid.New()
	}
	return nil
}

func (gm *GRPOMetric) BeforeCreate(tx *gorm.DB) error {
	if gm.ID == uuid.Nil {
		gm.ID = uuid.New()
	}
	return nil
}
