package Repositories
import "ares_api/internal/models"

type BalanceRepository interface {
	GetUSDBalance(userID uint) (*models.Balance, error)
	UpdateUSDBalance(userID uint, delta float64) (*models.Balance, error)
	ResetUSDBalance(userID uint, defaultBalance float64) error
	CreateUSDBalance(userID uint, defaultBalance float64) (*models.Balance, error)
}

