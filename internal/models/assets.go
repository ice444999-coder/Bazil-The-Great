package models

import "time"

// ==========================
// All Coins List
// ==========================
type Coin struct {
	ID     string `gorm:"primaryKey;size:100" json:"id"`
	Symbol string `gorm:"size:50" json:"symbol"`
	Name   string `gorm:"size:100" json:"name"`
}

// ==========================
// Coin Market Details
// ==========================
type CoinMarket struct {
	ID          string    `gorm:"primaryKey;size:100" json:"id"`
	Symbol      string    `gorm:"size:50" json:"symbol"`
	Name        string    `gorm:"size:100" json:"name"`
	PriceUSD    float64   `json:"price_usd"`
	MarketCap   float64   `json:"market_cap"`
	Change24h   float64   `json:"change_24h"`
	LastUpdated time.Time `json:"last_updated"`
}

// ==========================
// Top Movers (Gainers/Losers)
// ==========================
type TopMover struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	PriceUSD    float64   `json:"price_usd"`
	MarketCap   float64   `json:"market_cap"`
	Change24h   float64   `json:"price_change_percentage_24h"`
	LastUpdated time.Time `json:"last_updated"`
}
