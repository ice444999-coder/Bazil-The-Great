package service

import "ares_api/internal/api/dto"

type TradeService interface {
	MarketOrder(userID uint, req dto.MarketOrderRequest) (*dto.TradeResponse, error)
	LimitOrder(userID uint, req dto.LimitOrderRequest) (*dto.TradeResponse, error)
	GetHistory(userID uint, limit int) ([]dto.TradeResponse, error)
	GetPendingLimitOrders(userID uint ) ([]dto.TradeResponse, error)
}
