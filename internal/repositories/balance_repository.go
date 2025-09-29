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

func (r *BalanceRepositoryImpl) GetUSDBalance(userID uint) (*models.Balance, error) {
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
	}
	if err := r.DB.Create(&balance).Error; err != nil {
		return nil, err
	}
	return &balance, nil
}