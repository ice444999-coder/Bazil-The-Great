package models

import (
	
	"gorm.io/gorm"
)
type Chat struct {
	gorm.Model
	UserID      uint   `gorm:"index;not null"`
	Message     string `gorm:"type:text;not null"`  // user message
	Response    string `gorm:"type:text"`           // AI respons
}