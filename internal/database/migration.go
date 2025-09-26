package database

import (
	"ares_api/internal/models"

	"gorm.io/gorm"
)

// Function to auto-migrate everything
func AutoMigrateAll(db *gorm.DB) error {

	return db.AutoMigrate(
	// Add all your models here
	 &models.User{},
	 &models.Chat{},
	)
}
