/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package models
import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username       string  `gorm:"uniqueIndex;not null"`
	PasswordHash   string  `gorm:"not null"`
	Email          string  `gorm:"uniqueIndex;not null"`
	IsActive       bool    `gorm:"default:true"`
	VirtualBalance float64 `gorm:"default:10000.00;not null"` // Starting balance for sandbox trading
}
