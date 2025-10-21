/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package strategies

import (
	"fmt"
	"time"
)

// WhaleTrackingStrategy - Follow smart money by detecting large wallet movements
// Target: 20-50% annual returns by copying profitable whale trades
// Detection: >$1M transactions, order book 60/40 imbalances
// Data Source: WhaleAlert API + on-chain monitoring
type WhaleTrackingStrategy struct {
	Name               string
	Description        string
	Enabled            bool
	MinTransactionUSD  float64            // 1000000 ($1M threshold)
	ImbalanceThreshold float64            // 0.60 (60/40 ratio)
	FollowDelay        int                // 30 seconds (delay before mirroring)
	WhaleWallets       map[string]float64 // Wallet address -> success rate
}

// NewWhaleTrackingStrategy creates a whale tracking strategy instance
func NewWhaleTrackingStrategy() *WhaleTrackingStrategy {
	return &WhaleTrackingStrategy{
		Name:               "WhaleTracking",
		Description:        "Follow smart money by detecting large wallet movements",
		Enabled:            true,
		MinTransactionUSD:  1000000,
		ImbalanceThreshold: 0.60,
		FollowDelay:        30,
		WhaleWallets:       make(map[string]float64),
	}
}

// Generate creates trade signals based on whale activity
func (w *WhaleTrackingStrategy) Generate(marketData *MarketData) (*TradeSignal, error) {
	if !w.Enabled {
		return nil, fmt.Errorf("strategy disabled")
	}

	// Check for order book imbalance (whale accumulation/distribution)
	bidTotal := w.sumOrderBook(marketData.OrderBookBids)
	askTotal := w.sumOrderBook(marketData.OrderBookAsks)
	totalLiquidity := bidTotal + askTotal

	if totalLiquidity == 0 {
		return nil, nil // No liquidity data
	}

	bidRatio := bidTotal / totalLiquidity
	askRatio := askTotal / totalLiquidity

	// Heavy bid-side (60%+) = Whale accumulation = BUY
	if bidRatio >= w.ImbalanceThreshold {
		return &TradeSignal{
			Action:      "buy",
			Symbol:      marketData.Symbol,
			Confidence:  w.calculateConfidence(bidRatio, bidTotal),
			Reasoning:   fmt.Sprintf("Whale accumulation detected: %.1f%% bid-side liquidity ($%.2fM)", bidRatio*100, bidTotal/1000000),
			Strategy:    w.Name,
			Timestamp:   time.Now(),
			TargetGain:  0.05,  // 5% target (whales move markets)
			StopLoss:    -0.02, // -2% stop loss
			MaxHoldTime: 86400, // 24 hours (whales hold longer)
			Priority:    8,     // High priority
		}, nil
	}

	// Heavy ask-side (60%+) = Whale distribution = SELL
	if askRatio >= w.ImbalanceThreshold {
		return &TradeSignal{
			Action:      "sell",
			Symbol:      marketData.Symbol,
			Confidence:  w.calculateConfidence(askRatio, askTotal),
			Reasoning:   fmt.Sprintf("Whale distribution detected: %.1f%% ask-side liquidity ($%.2fM)", askRatio*100, askTotal/1000000),
			Strategy:    w.Name,
			Timestamp:   time.Now(),
			TargetGain:  0.05,
			StopLoss:    -0.02,
			MaxHoldTime: 86400,
			Priority:    8,
		}, nil
	}

	return nil, nil // No whale activity detected
}

// Analyze evaluates current whale activity
func (w *WhaleTrackingStrategy) Analyze(marketData *MarketData) *StrategyAnalysis {
	bidTotal := w.sumOrderBook(marketData.OrderBookBids)
	askTotal := w.sumOrderBook(marketData.OrderBookAsks)
	totalLiquidity := bidTotal + askTotal

	bidRatio := 0.0
	askRatio := 0.0
	if totalLiquidity > 0 {
		bidRatio = bidTotal / totalLiquidity
		askRatio = askTotal / totalLiquidity
	}

	// Calculate whale activity score
	score := 0.0
	maxRatio := bidRatio
	if askRatio > bidRatio {
		maxRatio = askRatio
	}

	if maxRatio >= w.ImbalanceThreshold {
		score = (maxRatio - 0.5) / 0.5 // 60% = 0.2 score, 100% = 1.0 score
	}

	// Bonus if total liquidity is very high (indicates whale presence)
	if totalLiquidity > 10000000 { // $10M+
		score += 0.2
	}

	if score > 1.0 {
		score = 1.0
	}

	return &StrategyAnalysis{
		StrategyName: w.Name,
		Score:        score,
		Indicators: map[string]float64{
			"bid_ratio":       bidRatio,
			"ask_ratio":       askRatio,
			"total_liquidity": totalLiquidity,
			"imbalance":       maxRatio,
		},
		Recommendation: w.getRecommendation(score),
		Timestamp:      time.Now(),
	}
}

// ProcessWhaleTransaction processes a detected whale transaction from WhaleAlert API
func (w *WhaleTrackingStrategy) ProcessWhaleTransaction(tx *WhaleTransaction) (*TradeSignal, error) {
	// Check if transaction meets threshold
	if tx.Amount < w.MinTransactionUSD {
		return nil, nil // Below threshold
	}

	// Check if wallet is known whale with good track record
	successRate, isKnown := w.WhaleWallets[tx.WalletAddress]
	confidence := 0.6 // Base confidence for unknown whales

	if isKnown && successRate > 0.6 {
		confidence = successRate // Use historical success rate
	}

	// Mirror the whale's trade (with delay for confirmation)
	time.Sleep(time.Duration(w.FollowDelay) * time.Second)

	action := "hold"
	if tx.Direction == "buy" {
		action = "buy"
	} else if tx.Direction == "sell" {
		action = "sell"
	}

	return &TradeSignal{
		Action:     action,
		Symbol:     tx.Symbol,
		Confidence: confidence,
		Reasoning: fmt.Sprintf("Whale %s: $%.2fM %s by %s (success rate: %.1f%%)",
			tx.Direction, tx.Amount/1000000, tx.Direction, tx.WalletAddress[:10], successRate*100),
		Strategy:    w.Name,
		Timestamp:   time.Now(),
		TargetGain:  0.05,
		StopLoss:    -0.02,
		MaxHoldTime: 86400,
		Priority:    9, // Very high priority for confirmed whale moves
	}, nil
}

// UpdateWhaleWallet updates the success rate for a known whale wallet
func (w *WhaleTrackingStrategy) UpdateWhaleWallet(wallet string, successRate float64) {
	w.WhaleWallets[wallet] = successRate
}

// sumOrderBook calculates total liquidity in order book side
func (w *WhaleTrackingStrategy) sumOrderBook(orders []OrderBookEntry) float64 {
	total := 0.0
	for _, order := range orders {
		total += order.Price * order.Amount
	}
	return total
}

// calculateConfidence determines signal strength based on imbalance magnitude
func (w *WhaleTrackingStrategy) calculateConfidence(ratio float64, liquidityUSD float64) float64 {
	// Base confidence from imbalance ratio
	imbalanceStrength := (ratio - 0.5) / 0.5 // 60% = 0.2, 100% = 1.0

	// Bonus for high liquidity (indicates real whale, not small trader)
	liquidityBonus := 0.0
	if liquidityUSD > 5000000 { // $5M+
		liquidityBonus = 0.2
	} else if liquidityUSD > 1000000 { // $1M+
		liquidityBonus = 0.1
	}

	confidence := imbalanceStrength*0.7 + liquidityBonus
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// getRecommendation translates score to action recommendation
func (w *WhaleTrackingStrategy) getRecommendation(score float64) string {
	if score > 0.7 {
		return "STRONG_SIGNAL"
	} else if score > 0.4 {
		return "MODERATE_SIGNAL"
	} else if score > 0.2 {
		return "WEAK_SIGNAL"
	}
	return "NO_SIGNAL"
}

// GetConfig returns strategy configuration
func (w *WhaleTrackingStrategy) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":                 w.Name,
		"enabled":              w.Enabled,
		"min_transaction_usd":  w.MinTransactionUSD,
		"imbalance_threshold":  w.ImbalanceThreshold,
		"follow_delay_seconds": w.FollowDelay,
		"tracked_whales_count": len(w.WhaleWallets),
	}
}

// UpdateConfig updates strategy parameters
func (w *WhaleTrackingStrategy) UpdateConfig(params map[string]interface{}) error {
	if val, ok := params["min_transaction_usd"].(float64); ok {
		w.MinTransactionUSD = val
	}
	if val, ok := params["imbalance_threshold"].(float64); ok {
		w.ImbalanceThreshold = val
	}
	if val, ok := params["follow_delay_seconds"].(int); ok {
		w.FollowDelay = val
	}
	return nil
}
