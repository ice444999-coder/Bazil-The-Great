package services

import (
	repo "ares_api/internal/interfaces/repository"
	service "ares_api/internal/interfaces/service"
	"ares_api/internal/models"
	"errors"
)

var _ service.SettingsService = &SettingsService{}

type SettingsService struct {
	Repo repo.SettingsRepository
}

func NewSettingsService(repo repo.SettingsRepository) *SettingsService {
	return &SettingsService{Repo: repo}
}

func (s *SettingsService) SaveAPIKey(userID uint, apiKey string) error {
	if apiKey == "" {
		return errors.New("API key cannot be empty")
	}

	setting, err := s.Repo.GetByUserID(userID)
	if err != nil {
		setting = &models.Setting{UserID: userID, APIKey: apiKey} 
		return s.Repo.Save(setting)
	}

	setting.APIKey = apiKey
	return s.Repo.Save(setting)
}




