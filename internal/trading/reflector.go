/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package trading

import (
	"ares_api/internal/models"
	"fmt"
	"strings"
	"time"
)

// Reflector analyzes completed trades and extracts lessons
// This is the "learning" component of the ACE Framework
type Reflector struct{}

// NewReflector creates a new trade analyzer
func NewReflector() *Reflector {
	return &Reflector{}
}

// TradeOutcome represents the result of a completed trade
type TradeOutcome struct {
	TradeID       uint
	Symbol        string
	Action        string    // "BUY" or "SELL"
	EntryPrice    float64
	ExitPrice     float64
	Amount        float64
	ProfitLoss    float64
	Duration      time.Duration
	WasProfit     bool
	
	// Market conditions at entry
	EntryRSI      float64
	EntryVolume   float64
	EntryMA20     float64
	TimeOfDay     int       // Hour (0-23)
	DayOfWeek     int       // 0=Sunday, 6=Saturday
	
	// Rules that were active during this trade
	RulesUsed     []string
}

// DeltaUpdate represents a change to make to the playbook
type DeltaUpdate struct {
	RuleID      string
	Content     string
	Category    string
	IsHelpful   bool    // true if trade was profitable
	Conditions  map[string]interface{}
}

// AnalyzeTrade examines a completed trade and generates insights
func (r *Reflector) AnalyzeTrade(outcome TradeOutcome, activeRules []models.PlaybookRule) []DeltaUpdate {
	deltas := make([]DeltaUpdate, 0)
	
	// 1. Update existing rules that were used
	for _, rule := range activeRules {
		delta := DeltaUpdate{
			RuleID:     rule.RuleID,
			Content:    rule.Content,
			Category:   rule.Category,
			IsHelpful:  outcome.WasProfit,
			Conditions: make(map[string]interface{}),
		}
		deltas = append(deltas, delta)
	}
	
	// 2. Extract new patterns if trade was profitable
	if outcome.WasProfit {
		newPatterns := r.ExtractSuccessPatterns(outcome)
		deltas = append(deltas, newPatterns...)
	}
	
	// 3. Extract failure patterns if trade was a loss
	if !outcome.WasProfit {
		failurePatterns := r.ExtractFailurePatterns(outcome)
		deltas = append(deltas, failurePatterns...)
	}
	
	return deltas
}

// ExtractSuccessPatterns identifies what made a trade successful
func (r *Reflector) ExtractSuccessPatterns(outcome TradeOutcome) []DeltaUpdate {
	patterns := make([]DeltaUpdate, 0)
	
	// Pattern 1: RSI Oversold + Buy = Profit
	if outcome.Action == "BUY" && outcome.EntryRSI < 30 {
		patterns = append(patterns, DeltaUpdate{
			RuleID:    "buy_rsi_oversold",
			Content:   "Buy when RSI < 30 (oversold condition)",
			Category:  "entry",
			IsHelpful: true,
			Conditions: map[string]interface{}{
				"rsi_threshold": 30,
				"action":        "BUY",
			},
		})
	}
	
	// Pattern 2: RSI Overbought + Sell = Profit
	if outcome.Action == "SELL" && outcome.EntryRSI > 70 {
		patterns = append(patterns, DeltaUpdate{
			RuleID:    "sell_rsi_overbought",
			Content:   "Sell when RSI > 70 (overbought condition)",
			Category:  "exit",
			IsHelpful: true,
			Conditions: map[string]interface{}{
				"rsi_threshold": 70,
				"action":        "SELL",
			},
		})
	}
	
	// Pattern 3: High Volume + Profitable Trade
	if outcome.EntryVolume > 2.0 {
		patterns = append(patterns, DeltaUpdate{
			RuleID:    fmt.Sprintf("%s_high_volume", strings.ToLower(outcome.Action)),
			Content:   fmt.Sprintf("%s when volume > 2x average", outcome.Action),
			Category:  "entry",
			IsHelpful: true,
			Conditions: map[string]interface{}{
				"volume_multiplier": 2.0,
				"action":            outcome.Action,
			},
		})
	}
	
	// Pattern 4: Time of Day Success
	if outcome.TimeOfDay >= 9 && outcome.TimeOfDay <= 16 {
		patterns = append(patterns, DeltaUpdate{
			RuleID:    "trade_market_hours",
			Content:   "Trades during market hours (9am-4pm) tend to succeed",
			Category:  "risk_management",
			IsHelpful: true,
			Conditions: map[string]interface{}{
				"hour_start": 9,
				"hour_end":   16,
			},
		})
	}
	
	// Pattern 5: Price above MA20 (uptrend)
	if outcome.Action == "BUY" && outcome.EntryPrice > outcome.EntryMA20 {
		patterns = append(patterns, DeltaUpdate{
			RuleID:    "buy_uptrend_ma20",
			Content:   "Buy when price is above 20-period moving average (uptrend)",
			Category:  "entry",
			IsHelpful: true,
			Conditions: map[string]interface{}{
				"price_vs_ma": "above",
				"ma_period":   20,
			},
		})
	}
	
	return patterns
}

// ExtractFailurePatterns identifies what made a trade fail
func (r *Reflector) ExtractFailurePatterns(outcome TradeOutcome) []DeltaUpdate {
	patterns := make([]DeltaUpdate, 0)
	
	// Anti-Pattern 1: Weekend Trading
	if outcome.DayOfWeek == 0 || outcome.DayOfWeek == 6 {
		patterns = append(patterns, DeltaUpdate{
			RuleID:    "avoid_weekend_trading",
			Content:   "Avoid trading on weekends (higher volatility, lower liquidity)",
			Category:  "risk_management",
			IsHelpful: false,
			Conditions: map[string]interface{}{
				"day_of_week": []int{0, 6},
			},
		})
	}
	
	// Anti-Pattern 2: Low Volume Trades
	if outcome.EntryVolume < 0.5 {
		patterns = append(patterns, DeltaUpdate{
			RuleID:    "avoid_low_volume",
			Content:   "Avoid trading when volume < 0.5x average (low liquidity)",
			Category:  "risk_management",
			IsHelpful: false,
			Conditions: map[string]interface{}{
				"volume_multiplier": 0.5,
			},
		})
	}
	
	// Anti-Pattern 3: Late Night Trading
	if outcome.TimeOfDay < 6 || outcome.TimeOfDay > 22 {
		patterns = append(patterns, DeltaUpdate{
			RuleID:    "avoid_late_night_trading",
			Content:   "Avoid trading between 10pm-6am (low volume, high spreads)",
			Category:  "risk_management",
			IsHelpful: false,
			Conditions: map[string]interface{}{
				"hour_start": 22,
				"hour_end":   6,
			},
		})
	}
	
	// Anti-Pattern 4: Counter-trend trading
	if outcome.Action == "BUY" && outcome.EntryPrice < outcome.EntryMA20 {
		patterns = append(patterns, DeltaUpdate{
			RuleID:    "avoid_buy_downtrend",
			Content:   "Avoid buying when price is below MA20 (downtrend)",
			Category:  "risk_management",
			IsHelpful: false,
			Conditions: map[string]interface{}{
				"price_vs_ma": "below",
				"action":      "BUY",
			},
		})
	}
	
	return patterns
}

// GenerateInsights creates human-readable analysis of trading patterns
func (r *Reflector) GenerateInsights(deltas []DeltaUpdate) []string {
	insights := make([]string, 0)
	
	helpfulCount := 0
	harmfulCount := 0
	
	for _, delta := range deltas {
		if delta.IsHelpful {
			helpfulCount++
		} else {
			harmfulCount++
		}
	}
	
	if helpfulCount > 0 {
		insights = append(insights, fmt.Sprintf("Identified %d successful trading patterns", helpfulCount))
	}
	
	if harmfulCount > 0 {
		insights = append(insights, fmt.Sprintf("Identified %d patterns to avoid", harmfulCount))
	}
	
	return insights
}
