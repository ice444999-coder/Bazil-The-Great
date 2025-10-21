/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package models

import "gorm.io/gorm"

type Balance struct {
	gorm.Model
	UserID uint    `gorm:"index;not null"`
	Asset  string  `gorm:"size:10;not null"`
	Amount float64 `gorm:"not null"`

	// Trading-specific fields
	AutoTopup       bool    `gorm:"default:false" json:"auto_topup"`
	TopupThreshold  float64 `gorm:"type:decimal(18,8);default:1000.00" json:"topup_threshold"`
	TopupAmount     float64 `gorm:"type:decimal(18,8);default:10000.00" json:"topup_amount"`
	TotalDeposits   float64 `gorm:"type:decimal(18,8);default:10000.00" json:"total_deposits"`
	TotalWithdrawals float64 `gorm:"type:decimal(18,8);default:0.00" json:"total_withdrawals"`
	RealizedPnL     float64 `gorm:"type:decimal(18,8);default:0.00" json:"realized_pnl"`
	UnrealizedPnL   float64 `gorm:"type:decimal(18,8);default:0.00" json:"unrealized_pnl"`
}
