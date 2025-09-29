package service

import "ares_api/internal/api/dto"
type AssetService interface {
	GetAllCoins(limit int) ([]dto.CoinDTO, error)
	GetCoinMarket(id string , vsCurrency string) (*dto.CoinMarketDTO, error)
	GetTopMovers(limit int) ([]dto.TopMoverDTO, error)
	GetSupportedVSCurrencies() ([]string, error)
}
