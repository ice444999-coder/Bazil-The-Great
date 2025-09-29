package Repositories

import "ares_api/internal/models"

type SettingsRepository interface {
	GetByUserID(userID uint) (*models.Setting, error)
	Save(setting *models.Setting) error
	
}
