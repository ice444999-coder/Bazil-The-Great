package Repositories

import (
	"ares_api/internal/api/dto"
)

type AssetRepository interface {
	FetchAllCoins() ([]dto.CoinDTO, error)
	FetchCoinMarket(id string , vsCurrency string) (*dto.CoinMarketDTO, error)
	FetchTopMovers(limit int) ([]dto.TopMoverDTO, error)
	FetchSupportedVSCurrencies() ([]string, error)
}