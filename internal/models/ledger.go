package models

import (
	"gorm.io/gorm"
)

type Ledger struct {

	gorm.Model
	UserID    uint      `json:"user_id"`
	Action    string    `json:"action"`           // CHAT, TRADE_MARKET, TRADE_LIMIT, BALANCE_RESET, API_KEY_UPDATE, COMPILE, TEST
	Details   string    `json:"details"`          // JSON string with additional info
}
