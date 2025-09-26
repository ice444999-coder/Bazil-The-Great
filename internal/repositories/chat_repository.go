package repositories

import (
	repository "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"gorm.io/gorm"
)

type ChatRepositoryImpl struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) repository.ChatRepository {
	return &ChatRepositoryImpl{db: db}
}

func (r *ChatRepositoryImpl) Create(chat *models.Chat) error {
	return r.db.Create(chat).Error
}

func (r *ChatRepositoryImpl) GetByUserID(userID uint, limit int) ([]models.Chat, error) {
	var chats []models.Chat
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&chats).Error
	return chats, err
}

