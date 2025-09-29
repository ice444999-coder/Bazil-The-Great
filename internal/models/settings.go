package models

import "gorm.io/gorm"

type Setting struct {
	gorm.Model
	UserID   uint   `gorm:"not null;uniqueIndex"` // each user has one setting row
	APIKey   string `gorm:"size:255"`
}
