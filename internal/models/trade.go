package models

import (
	"gorm.io/gorm"
)

type Trade struct {
	gorm.Model
	UserID   uint    `gorm:"not null;index" json:"user_id"`  // user placing the trade
	CoinID   string  `gorm:"size:10;not null;index" json:"coin_id"`
	Symbol   string  `gorm:"size:20;not null;index" json:"symbol"` // trading pair
	Side     string  `gorm:"size:10;not null" json:"side"`  // buy or sell
	Quantity float64 `gorm:"not null" json:"quantity"`
	Price    float64 `gorm:"not null" json:"price"`
	Type     string  `gorm:"size:10;not null" json:"type"`  // market or limit
	Status   string  `gorm:"size:20;not null" json:"status"` // filled, open, cancelled
}
