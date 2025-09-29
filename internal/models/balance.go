package models

import "gorm.io/gorm"

type Balance struct {
	gorm.Model
	UserID uint    `gorm:"index;not null"`
	Asset  string  `gorm:"size:10;not null"`
	Amount float64 `gorm:"not null"`
}
