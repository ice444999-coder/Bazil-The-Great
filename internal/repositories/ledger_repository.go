package repositories

import (
	"ares_api/internal/models"
	repository"ares_api/internal/interfaces/repository"

	"gorm.io/gorm"
)

type LedgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) repository.LedgerRepository {
	return &LedgerRepository{db: db}
}

// Append a new ledger entry
func (r *LedgerRepository) Append(entry *models.Ledger) error {
	return r.db.Create(entry).Error
}

// GetLast retrieves last N ledger entries for a user
func (r *LedgerRepository) GetLast(userID uint, limit int) ([]models.Ledger, error) {
	var entries []models.Ledger
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&entries).Error
	return entries, err
}
