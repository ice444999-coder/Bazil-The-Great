package repositories

import (
	"ares_api/internal/models"
	repository "ares_api/internal/interfaces/repository"
	"gorm.io/gorm"
)

type settingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) repository.SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) GetByUserID(userID uint) (*models.Setting, error) {
	var setting models.Setting
	if err := r.db.Where("user_id = ?", userID).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *settingsRepository) Save(setting *models.Setting) error {
	return r.db.Save(setting).Error
}


