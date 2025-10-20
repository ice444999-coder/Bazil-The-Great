package models

import "gorm.io/gorm"

type BazilReward struct {
	gorm.Model
	FaultType string `gorm:"uniqueIndex" json:"fault_type"`
	Points    int    `json:"points"`
}
