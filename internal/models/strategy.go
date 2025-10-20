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
