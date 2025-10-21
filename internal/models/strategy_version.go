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

type StrategyVersion struct {
	gorm.Model
	StrategyID uint      `gorm:"not null;index" json:"strategy_id"`
	Strategy   Strategy  `gorm:"foreignKey:StrategyID" json:"-"`
	Version    string    `gorm:"size:64;not null" json:"version"` // SHA256 hash
	Code       string    `gorm:"type:text;not null" json:"code"`  // Serialized strategy code
	IsActive   bool      `gorm:"default:false" json:"is_active"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}
