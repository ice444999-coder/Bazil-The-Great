package services

import (
	"ares_api/internal/api/dto"
	repo "ares_api/internal/interfaces/repository"
	service "ares_api/internal/interfaces/service"
	"sync"
	"time"
)

type AssetServiceImpl struct {
	Repo       repo.AssetRepository
	coinCache  []dto.CoinDTO
	cacheMutex sync.RWMutex
	lastFetch  time.Time
	cacheTTL   time.Duration
}

func NewAssetService(r repo.AssetRepository) service.AssetService {
	return &AssetServiceImpl{
		Repo:     r,
		cacheTTL: 1 * time.Hour, 
	}
}

// GetAllCoins with caching
func (s *AssetServiceImpl) GetAllCoins(limit int) ([]dto.CoinDTO, error) {
	s.cacheMutex.RLock()
	if time.Since(s.lastFetch) < s.cacheTTL && len(s.coinCache) > 0 {
		cached := s.coinCache
		s.cacheMutex.RUnlock()
		if limit > 0 && limit < len(cached) {
			return cached[:limit], nil
		}
		return cached, nil
	}
	s.cacheMutex.RUnlock()

	coins, err := s.Repo.FetchAllCoins()
	if err != nil {
		return nil, err
	}

	s.cacheMutex.Lock()
	s.coinCache = coins
	s.lastFetch = time.Now()
	s.cacheMutex.Unlock()

	if limit > 0 && limit < len(coins) {
		return coins[:limit], nil
	}
	return coins, nil
}

// GetCoinMarket returns market data for a single coin
func (s *AssetServiceImpl) GetCoinMarket(id string , vsCurrency string) (*dto.CoinMarketDTO, error) {
	return s.Repo.FetchCoinMarket(id , vsCurrency)
}

// GetTopMovers returns top movers with optional limit
func (s *AssetServiceImpl) GetTopMovers(limit int) ([]dto.TopMoverDTO, error) {
	return s.Repo.FetchTopMovers(limit)
}

// GetSupportedVSCurrencies returns a list of supported virtual currencies
func (s *AssetServiceImpl) GetSupportedVSCurrencies() ([]string, error) {
    return s.Repo.FetchSupportedVSCurrencies()
}
