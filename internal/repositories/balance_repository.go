/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package repositories

import (
	repository "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"

	"gorm.io/gorm"
)



type BalanceRepositoryImpl struct {
	DB *gorm.DB
}

func NewBalanceRepository(db *gorm.DB) repository.BalanceRepository {
	return &BalanceRepositoryImpl{DB: db}
}

func (r *BalanceRepositoryImpl) GetUSDBalanceModel(userID uint) (*models.Balance, error) {
	var balance models.Balance
	err := r.DB.Where("user_id = ? AND asset = ?", userID, "USD").First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *BalanceRepositoryImpl) UpdateUSDBalance(userID uint, delta float64) (*models.Balance, error) {
	var balance models.Balance
	if err := r.DB.Where("user_id = ? AND asset = ?", userID, "USD").First(&balance).Error; err != nil {
		return nil, err
	}

	balance.Amount += delta
	if balance.Amount < 0 {
		return nil, gorm.ErrInvalidData // insufficient funds
	}

	if err := r.DB.Save(&balance).Error; err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *BalanceRepositoryImpl) ResetUSDBalance(userID uint, defaultBalance float64) error {
	return r.DB.Model(&models.Balance{}).
		Where("user_id = ? AND asset = ?", userID, "USD").
		Update("amount", defaultBalance).Error
}

func (r *BalanceRepositoryImpl) CreateUSDBalance(userID uint, defaultBalance float64) (*models.Balance, error) {
	balance := models.Balance{
		UserID: userID,
		Asset:  "USD",
		Amount: defaultBalance,
		TotalDeposits: defaultBalance,
	}
	if err := r.DB.Create(&balance).Error; err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetUSDBalance returns just the USD amount as float64
func (r *BalanceRepositoryImpl) GetUSDBalance(userID uint) (float64, error) {
	var balance models.Balance
	err := r.DB.Where("user_id = ? AND asset = ?", userID, "USD").First(&balance).Error
	if err != nil {
		return 0, err
	}
	return balance.Amount, nil
}

// GetBalanceRecord returns the full balance record
func (r *BalanceRepositoryImpl) GetBalanceRecord(userID uint) (*models.Balance, error) {
	var balance models.Balance
	err := r.DB.Where("user_id = ? AND asset = ?", userID, "USD").First(&balance).Error
	return &balance, err
}

// UpdateBalance updates the USD balance amount
func (r *BalanceRepositoryImpl) UpdateBalance(userID uint, newAmount float64) error {
	return r.DB.Model(&models.Balance{}).
		Where("user_id = ? AND asset = ?", userID, "USD").
		Update("amount", newAmount).Error
}

// UpdateBalanceRecord saves the entire balance record
func (r *BalanceRepositoryImpl) UpdateBalanceRecord(balance *models.Balance) error {
	return r.DB.Save(balance).Error
}