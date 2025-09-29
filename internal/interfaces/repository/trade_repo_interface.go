package Repositories

import "ares_api/internal/models"

type TradeRepository interface {
	Create(trade *models.Trade) error
	GetByUserID(userID uint, limit int) ([]models.Trade, error)
	GetOpenLimitOrders() ([]models.Trade, error)
	MarkOrderFilled(tradeID uint) error
	GetOpenLimitOrdersByUser(userID uint) ([]models.Trade, error)
}
