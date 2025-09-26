package Repositories
import "ares_api/internal/models"

type ChatRepository interface {
	Create(chat *models.Chat) error
	GetByUserID(userID uint, limit int) ([]models.Chat, error)
}
