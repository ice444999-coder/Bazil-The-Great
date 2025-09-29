package service

import "ares_api/internal/api/dto"

type ChatService interface {
	SendMessage(userID uint, req dto.ChatRequest) (dto.ChatResponse, error)
	GetHistory(userID uint, limit int) ([]dto.ChatHistoryResponse, error)
}
