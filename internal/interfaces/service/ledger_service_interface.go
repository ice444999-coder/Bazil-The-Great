package service

// LedgerService defines high-level operations for ledger
type LedgerService interface {
	// Append a new ledger entry
	Append(userID uint, action string, details interface{}) error

	// Get last N entries for a user
	GetLast(userID uint, limit int) ([]interface{}, error)
}
