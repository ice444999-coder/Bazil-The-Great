package service

import (
	"ares_api/internal/api/dto"
)

type BalanceService interface {
	GetUSDBalance(userID uint) (*dto.BalanceDTO, error)
	UpdateUSDBalance(userID uint, delta float64) (*dto.BalanceDTO, error)
	ResetUSDBalance(userID uint) (*dto.BalanceDTO, error)
	InitializeBalance(userID uint) (*dto.BalanceDTO, error)
}

