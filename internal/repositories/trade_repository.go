package repositories

import (
	repo "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"time"

	"gorm.io/gorm"
)

type TradeRepository struct {
	db *gorm.DB
}

func NewTradeRepository(db *gorm.DB) repo.TradeRepository {
	return &TradeRepository{db: db}
}

func (r *TradeRepository) Create(trade *models.Trade) error {
	return r.db.Create(trade).Error
}

func (r *TradeRepository) GetByUserID(userID uint, limit int) ([]models.Trade, error) {
	var trades []models.Trade
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&trades).Error
	return trades, err
}

func (r *TradeRepository) GetOpenLimitOrders() ([]models.Trade, error) {
	var trades []models.Trade
	err := r.db.Where("type = ? AND status = ?", "limit", "open").Find(&trades).Error
	return trades, err
}

func (r *TradeRepository) MarkOrderFilled(tradeID uint) error {
	return r.db.Model(&models.Trade{}).Where("id = ?", tradeID).Updates(map[string]interface{}{
		"status":     "filled",
		"updated_at": time.Now(),
	}).Error
}

func (r *TradeRepository) GetOpenLimitOrdersByUser(userID uint) ([]models.Trade, error) {
	var trades []models.Trade
	err := r.db.Where("user_id = ? AND type = ? AND status = ?", userID, "limit", "open").Find(&trades).Error
	return trades, err
}
