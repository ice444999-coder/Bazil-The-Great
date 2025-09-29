package services

import (
	 repository"ares_api/internal/interfaces/repository"
	"ares_api/internal/interfaces/service"
	"ares_api/internal/models"
	"encoding/json"
	"fmt"
)

type LedgerService struct {
	Repo repository.LedgerRepository
}

func NewLedgerService(repo repository.LedgerRepository) service.LedgerService {
	return &LedgerService{Repo: repo}
}

// Append a new ledger entry
func (s *LedgerService) Append(userID uint, action string, details interface{}) error {
	detailBytes, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	entry := &models.Ledger{
		UserID:  userID,
		Action:  action,
		Details: string(detailBytes),
	}

	return s.Repo.Append(entry)
}

// Get last N entries
func (s *LedgerService) GetLast(userID uint, limit int) ([]interface{}, error) {
	entries, err := s.Repo.GetLast(userID, limit)
	if err != nil {
		return nil, err
	}

	var result []interface{}
	for _, e := range entries {
		var d interface{}
		if err := json.Unmarshal([]byte(e.Details), &d); err != nil {
			d = e.Details
		}
		result = append(result, map[string]interface{}{
			"id":         e.ID,
			"user_id":    e.UserID,
			"action":     e.Action,
			"details":    d,
			"created_at": e.CreatedAt,
		})
	}

	return result, nil
}
