/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package models

import "gorm.io/gorm"

type BazilReward struct {
	gorm.Model
	FaultType string `gorm:"uniqueIndex" json:"fault_type"`
	Points    int    `json:"points"`
}
