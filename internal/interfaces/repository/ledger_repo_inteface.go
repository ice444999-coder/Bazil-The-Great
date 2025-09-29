package Repositories

import "ares_api/internal/models"

// LedgerRepository defines database operations for the ledger
type LedgerRepository interface {
	// Append a new entry to the ledger
	Append(entry *models.Ledger) error

	// GetLast retrieves the last N entries for a given user
	GetLast(userID uint, limit int) ([]models.Ledger, error)
}
