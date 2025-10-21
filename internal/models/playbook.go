/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package models

import (
	"time"
)

// PlaybookRule represents a learned trading pattern or strategy
// This is the core of the ACE Framework - rules that SOLACE learns from experience
type PlaybookRule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Rule Identity
	RuleID      string `gorm:"uniqueIndex;not null" json:"rule_id"`        // e.g., "buy_rsi_oversold_high_volume"
	Content     string `gorm:"type:text;not null" json:"content"`          // Human-readable description
	Category    string `gorm:"index" json:"category"`                      // "entry", "exit", "risk_management", "market_condition"
	
	// Performance Tracking
	HelpfulCount int     `gorm:"default:0" json:"helpful_count"`           // Times this rule led to profit
	HarmfulCount int     `gorm:"default:0" json:"harmful_count"`           // Times this rule led to loss
	Confidence   float64 `gorm:"default:0.0" json:"confidence"`            // helpful / (helpful + harmful)
	
	// Context (What conditions trigger this rule)
	Conditions JSONB `gorm:"type:jsonb" json:"conditions"`                // Market conditions when rule applies
	
	// Metadata
	UserID       uint    `gorm:"index" json:"user_id"`                     // Which user this rule belongs to
	IsActive     bool    `gorm:"default:true;index" json:"is_active"`      // Can be disabled if underperforming
	TotalUses    int     `gorm:"default:0" json:"total_uses"`              // How many times this rule was consulted
	AvgProfit    float64 `gorm:"default:0.0" json:"avg_profit"`            // Average profit when rule used
	LastUsedAt   *time.Time `json:"last_used_at"`                          // When rule was last applied
}

// TableName overrides default table name
func (PlaybookRule) TableName() string {
	return "trading_playbook"
}

// CalculateConfidence updates the confidence score based on outcomes
func (p *PlaybookRule) CalculateConfidence() {
	total := p.HelpfulCount + p.HarmfulCount
	if total == 0 {
		p.Confidence = 0.0
		return
	}
	p.Confidence = float64(p.HelpfulCount) / float64(total)
}

// ShouldPrune determines if this rule should be removed from playbook
// Prune if: confidence < 30% AND total observations > 20
func (p *PlaybookRule) ShouldPrune() bool {
	total := p.HelpfulCount + p.HarmfulCount
	return total >= 20 && p.Confidence < 0.30
}

// IsReliable returns true if rule has enough data and good performance
func (p *PlaybookRule) IsReliable() bool {
	total := p.HelpfulCount + p.HarmfulCount
	return total >= 10 && p.Confidence >= 0.60
}
