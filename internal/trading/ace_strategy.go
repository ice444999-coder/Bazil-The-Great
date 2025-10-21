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
)

// ACEStrategy uses the trading playbook to make informed decisions
// This is the "Generator" component of the ACE Framework
type ACEStrategy struct {
	userID         uint
	curator        *Curator
	baseConfidence float64 // Default confidence without playbook
}

// NewACEStrategy creates a playbook-enhanced strategy
func NewACEStrategy(userID uint, curator *Curator) *ACEStrategy {
	return &ACEStrategy{
		userID:         userID,
		curator:        curator,
		baseConfidence: 50.0, // Start at 50% confidence
	}
}

func (s *ACEStrategy) Name() string {
	return "ACE (Agentic Context Engineering)"
}

func (s *ACEStrategy) Description() string {
	return "Learns from past trades and applies proven patterns. Confidence improves over time."
}

func (s *ACEStrategy) GetRiskLevel() string {
	return "Adaptive" // Risk adjusts based on learned patterns
}

// MarketConditions represents current market state
type MarketConditions struct {
	RSI            float64
	Volume         float64
	Price          float64
	MA20           float64
	TimeOfDay      int
	DayOfWeek      int
	Symbol         string
}

// Analyze generates a trade signal using playbook knowledge
func (s *ACEStrategy) Analyze(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	// Extract current market conditions
	conditions := s.extractConditions(symbol, marketData, history)
	
	// Query playbook for relevant rules
	relevantRules, err := s.QueryPlaybook(conditions)
	if err != nil {
		return nil, err
	}
	
	// Generate base signal using simple logic
	baseSignal := s.generateBaseSignal(conditions)
	
	// Apply playbook rules to enhance signal
	enhancedSignal := s.ApplyPlaybookRules(baseSignal, relevantRules, conditions)
	
	return enhancedSignal, nil
}

// extractConditions pulls relevant data from market
func (s *ACEStrategy) extractConditions(symbol string, marketData *MockMarketData, history []VirtualTrade) MarketConditions {
	// Get current price for symbol
	currentPrice := marketData.GetPrice(symbol)
	
	// Calculate RSI (simple approximation)
	rsi := s.calculateSimpleRSI(currentPrice, history)
	
	// Calculate volume ratio
	volumeRatio := 1.0 // Placeholder: would compare current volume to average
	
	// Calculate MA20
	ma20 := s.calculateMA20(history)
	
	return MarketConditions{
		RSI:       rsi,
		Volume:    volumeRatio,
		Price:     currentPrice,
		MA20:      ma20,
		TimeOfDay: 12, // Placeholder: would use actual time
		DayOfWeek: 3,  // Placeholder: would use actual day
		Symbol:    symbol,
	}
}

// QueryPlaybook finds rules matching current conditions
func (s *ACEStrategy) QueryPlaybook(conditions MarketConditions) ([]models.PlaybookRule, error) {
	// Get all reliable rules
	allRules, err := s.curator.GetReliableRules(s.userID)
	if err != nil {
		return nil, err
	}
	
	// Filter rules that match current conditions
	relevantRules := make([]models.PlaybookRule, 0)
	for _, rule := range allRules {
		if s.ruleMatchesConditions(rule, conditions) {
			relevantRules = append(relevantRules, rule)
		}
	}
	
	return relevantRules, nil
}

// ruleMatchesConditions checks if rule applies to current market
func (s *ACEStrategy) ruleMatchesConditions(rule models.PlaybookRule, conditions MarketConditions) bool {
	// Check RSI-based rules
	if strings.Contains(rule.RuleID, "rsi_oversold") {
		return conditions.RSI < 30
	}
	if strings.Contains(rule.RuleID, "rsi_overbought") {
		return conditions.RSI > 70
	}
	
	// Check volume-based rules
	if strings.Contains(rule.RuleID, "high_volume") {
		return conditions.Volume > 2.0
	}
	if strings.Contains(rule.RuleID, "low_volume") {
		return conditions.Volume < 0.5
	}
	
	// Check trend-based rules
	if strings.Contains(rule.RuleID, "uptrend") {
		return conditions.Price > conditions.MA20
	}
	if strings.Contains(rule.RuleID, "downtrend") {
		return conditions.Price < conditions.MA20
	}
	
	// Check time-based rules
	if strings.Contains(rule.RuleID, "weekend") {
		return conditions.DayOfWeek == 0 || conditions.DayOfWeek == 6
	}
	if strings.Contains(rule.RuleID, "market_hours") {
		return conditions.TimeOfDay >= 9 && conditions.TimeOfDay <= 16
	}
	
	return false
}

// generateBaseSignal creates initial recommendation
func (s *ACEStrategy) generateBaseSignal(conditions MarketConditions) *TradeSignal {
	signal := &TradeSignal{
		Action:     "hold",
		Confidence: s.baseConfidence,
		Reasoning:  "No strong signal detected",
		Strategy:   s.Name(),
		Symbol:     conditions.Symbol,
	}
	
	// Simple RSI logic
	if conditions.RSI < 30 {
		signal.Action = "buy"
		signal.Reasoning = fmt.Sprintf("RSI %.1f indicates oversold condition", conditions.RSI)
	} else if conditions.RSI > 70 {
		signal.Action = "sell"
		signal.Reasoning = fmt.Sprintf("RSI %.1f indicates overbought condition", conditions.RSI)
	}
	
	return signal
}

// ApplyPlaybookRules enhances signal with learned patterns
func (s *ACEStrategy) ApplyPlaybookRules(baseSignal *TradeSignal, rules []models.PlaybookRule, conditions MarketConditions) *TradeSignal {
	if len(rules) == 0 {
		// No playbook rules available yet
		baseSignal.Reasoning += " (No playbook data - learning phase)"
		return baseSignal
	}
	
	// Calculate confidence boost from rules
	totalConfidenceBoost := 0.0
	rulesApplied := 0
	reasoning := baseSignal.Reasoning
	
	for _, rule := range rules {
		// Track that we're using this rule
		s.curator.RecordRuleUsage(rule.RuleID, s.userID)
		
		// Check if rule supports or opposes current signal
		if s.ruleSupportsAction(rule, baseSignal.Action) {
			// Add confidence based on rule's historical performance
			boost := rule.Confidence * 20.0 // Max +20% confidence per rule
			totalConfidenceBoost += boost
			rulesApplied++
			
			reasoning += fmt.Sprintf("\n✓ Playbook: %s (%.0f%% confidence)", rule.Content, rule.Confidence*100)
		} else if s.ruleOpposesAction(rule, baseSignal.Action) {
			// Rule suggests avoiding this action
			totalConfidenceBoost -= rule.Confidence * 30.0 // Penalty
			reasoning += fmt.Sprintf("\n✗ Warning: %s (%.0f%% confidence)", rule.Content, rule.Confidence*100)
		}
	}
	
	// Apply confidence boost
	baseSignal.Confidence += totalConfidenceBoost
	
	// Cap confidence at 95% (never 100% certain)
	if baseSignal.Confidence > 95.0 {
		baseSignal.Confidence = 95.0
	}
	
	// Floor confidence at 5%
	if baseSignal.Confidence < 5.0 {
		baseSignal.Confidence = 5.0
	}
	
	baseSignal.Reasoning = reasoning
	
	if rulesApplied > 0 {
		baseSignal.Reasoning += fmt.Sprintf("\n📚 Applied %d playbook rules", rulesApplied)
	}
	
	return baseSignal
}

// ruleSupportsAction checks if rule encourages this action
func (s *ACEStrategy) ruleSupportsAction(rule models.PlaybookRule, action string) bool {
	actionLower := strings.ToLower(action)
	ruleIDLower := strings.ToLower(rule.RuleID)
	
	if actionLower == "buy" && strings.Contains(ruleIDLower, "buy") {
		return true
	}
	if actionLower == "sell" && strings.Contains(ruleIDLower, "sell") {
		return true
	}
	
	return false
}

// ruleOpposesAction checks if rule warns against this action
func (s *ACEStrategy) ruleOpposesAction(rule models.PlaybookRule, action string) bool {
	actionLower := strings.ToLower(action)
	ruleIDLower := strings.ToLower(rule.RuleID)
	
	// Avoidance rules
	if strings.Contains(ruleIDLower, "avoid") {
		return true
	}
	
	// Opposite action rules
	if actionLower == "buy" && strings.Contains(ruleIDLower, "sell") {
		return true
	}
	if actionLower == "sell" && strings.Contains(ruleIDLower, "buy") {
		return true
	}
	
	return false
}

// Helper functions for technical indicators

func (s *ACEStrategy) calculateSimpleRSI(currentPrice float64, history []VirtualTrade) float64 {
	if len(history) < 2 {
		return 50.0 // Neutral if not enough data
	}
	
	// Simple RSI approximation (real RSI requires more data)
	lastPrice := history[len(history)-1].Price
	priceChange := ((currentPrice - lastPrice) / lastPrice) * 100
	
	// Map price change to RSI scale (very simplified)
	rsi := 50.0 + (priceChange * 5.0)
	
	if rsi < 0 {
		rsi = 0
	}
	if rsi > 100 {
		rsi = 100
	}
	
	return rsi
}

func (s *ACEStrategy) calculateMA20(history []VirtualTrade) float64 {
	if len(history) == 0 {
		return 0
	}
	
	// Average last 20 trades (or all if less than 20)
	count := 20
	if len(history) < 20 {
		count = len(history)
	}
	
	sum := 0.0
	for i := len(history) - count; i < len(history); i++ {
		sum += history[i].Price
	}
	
	return sum / float64(count)
}
