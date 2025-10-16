package models
import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username       string  `gorm:"uniqueIndex;not null"`
	PasswordHash   string  `gorm:"not null"`
	Email          string  `gorm:"uniqueIndex;not null"`
	IsActive       bool    `gorm:"default:true"`
	VirtualBalance float64 `gorm:"default:10000.00;not null"` // Starting balance for sandbox trading
}
