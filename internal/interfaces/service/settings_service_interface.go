package service

type SettingsService interface {
	SaveAPIKey(userID uint, apiKey string) error
}
