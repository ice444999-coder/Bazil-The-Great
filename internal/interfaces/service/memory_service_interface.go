package service

import (
	"ares_api/internal/api/dto"

	"github.com/google/uuid"
)

type MemoryService interface {
	Learn(userID uint, eventType string, payload interface{}, sessionID *uuid.UUID) error
	Recall(userID uint, limit int) ([]dto.MemoryRecallResponse, error)
	RecallByEventType(userID uint, eventType string, limit int) ([]dto.MemoryRecallResponse, error)
	RecallBySessionID(sessionID uuid.UUID, limit int) ([]dto.MemoryRecallResponse, error)
	ImportConversation(userID uint, content string, source string, tags []string) (int, uint, error)
}
