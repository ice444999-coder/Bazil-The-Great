/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package services

import (
	"ares_api/internal/api/dto"
	 repository "ares_api/internal/interfaces/repository"
	"ares_api/internal/interfaces/service"
)

const DefaultBalance = 10000.0 // Every user starts with 10k USD

type BalanceServiceImpl struct {
	Repo repository.BalanceRepository
}

func NewBalanceService(r repository.BalanceRepository) service.BalanceService {
	return &BalanceServiceImpl{Repo: r}
}

func (s *BalanceServiceImpl) GetUSDBalance(userID uint) (*dto.BalanceDTO, error) {
	b, err := s.Repo.GetUSDBalanceModel(userID)
	if err != nil {
		return nil, err
	}
	return &dto.BalanceDTO{UserID: b.UserID, Asset: b.Asset, Amount: b.Amount}, nil
}

func (s *BalanceServiceImpl) UpdateUSDBalance(userID uint, delta float64) (*dto.BalanceDTO, error) {
	b, err := s.Repo.UpdateUSDBalance(userID, delta)
	if err != nil {
		return nil, err
	}
	return &dto.BalanceDTO{UserID: b.UserID, Asset: b.Asset, Amount: b.Amount}, nil
}

func (s *BalanceServiceImpl) ResetUSDBalance(userID uint) (*dto.BalanceDTO, error) {
	if err := s.Repo.ResetUSDBalance(userID, DefaultBalance); err != nil {
		return nil, err
	}
	return &dto.BalanceDTO{UserID: userID, Asset: "USD", Amount: DefaultBalance}, nil
}

func (s *BalanceServiceImpl) InitializeBalance(userID uint) (*dto.BalanceDTO, error) {
	b, err := s.Repo.CreateUSDBalance(userID, DefaultBalance)
	if err != nil {
		return nil, err
	}
	return &dto.BalanceDTO{UserID: b.UserID, Asset: b.Asset, Amount: b.Amount}, nil
}
