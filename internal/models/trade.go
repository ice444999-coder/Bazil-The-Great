package models

import (
	"time"
	"gorm.io/gorm"
)

type Trade struct {
	gorm.Model
	UserID          uint       `gorm:"not null;index:idx_user_status" json:"user_id"` // user placing the trade
	TradeID         string     `gorm:"size:100;uniqueIndex" json:"trade_id"`  // Sandbox trade ID (e.g., SANDBOX_xxx)
	CoinID          string     `gorm:"size:10;index" json:"coin_id"`
	Symbol          string     `gorm:"size:20;not null;index:idx_symbol" json:"symbol"` // trading pair (e.g., SOL/USDC)
	Side            string     `gorm:"size:10;not null" json:"side"`  // buy or sell
	Amount          float64    `gorm:"not null" json:"amount"`  // Amount of base asset
	Quantity        float64    `gorm:"not null" json:"quantity"` // Legacy: same as Amount
	Price           float64    `gorm:"not null" json:"price"`   // Entry price
	ExitPrice       *float64   `json:"exit_price,omitempty"`    // Exit price (when closed)
	Type            string     `gorm:"size:10;not null;default:'market'" json:"type"`  // market or limit
	Status          string     `gorm:"size:20;not null;index:idx_user_status" json:"status"` // open, closed, cancelled
	Strategy        string     `gorm:"size:50" json:"strategy"` // Strategy name (e.g., Momentum)
	Reasoning       string     `gorm:"type:text" json:"reasoning"` // AI reasoning for trade
	ProfitLoss      float64    `json:"profit_loss"` // P&L in USDC
	ProfitLossPct   float64    `json:"profit_loss_pct"` // P&L percentage
	Fee             float64    `json:"fee"` // Trading fee paid
	TransactionHash string     `gorm:"size:64" json:"transaction_hash"` // SHA256 hash for audit
	ExecutedAt      time.Time  `gorm:"index" json:"executed_at"` // When trade opened
	ExitedAt        *time.Time `json:"exited_at,omitempty"` // When trade closed
}
