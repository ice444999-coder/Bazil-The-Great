package dto

import "time"

// ==========================
// All Coins List DTO
// ==========================
type CoinDTO struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

// ==========================
// Coin Market Details DTO
// ==========================
type CoinMarketDTO struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	PriceUSD    float64   `json:"price_usd"`
	MarketCap   float64   `json:"market_cap"`
	Change24h   float64   `json:"change_24h"`
	LastUpdated time.Time `json:"last_updated"`
}

// ==========================
// Top Movers DTO
// ==========================
type TopMoverDTO struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	PriceUSD    float64   `json:"price_usd"`
	MarketCap   float64   `json:"market_cap"`
	Change24h   float64   `json:"change_24h"`
	LastUpdated time.Time `json:"last_updated"`
}

//===========================
// Vs currency DTO
//===========================

type VsCurrencyDTO struct {
	Currency string `json:"currency"`
}
