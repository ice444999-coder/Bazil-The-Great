/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package models

import (
	"time"

	"gorm.io/gorm"
)

type Strategy struct {
	gorm.Model
	Name        string    `gorm:"uniqueIndex;not null" json:"name"` // RSI_Oversold, MACD_Volume, etc.
	Description string    `gorm:"type:text" json:"description"`
	Mode        string    `gorm:"size:10;not null;default:'sandbox'" json:"mode"` // sandbox, live
	IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
	Config      JSONB     `gorm:"type:jsonb" json:"config"` // Strategy-specific parameters
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}
