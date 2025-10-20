package strategies

import (
	"fmt"
	"time"
)

// ScalpingStrategy - High-frequency micro-trades on order book imbalances
// Target: 20-50% annual returns via 0.2-0.5% gains per trade
// Timeframe: 1-5 minutes
// Indicators: RSI(8), Volume spike (>150% avg), Bid/Ask spread
type ScalpingStrategy struct {
	Name            string
	Description     string
	Enabled         bool
	MinGain         float64 // 0.002 (0.2%)
	MaxHold         int     // 300 seconds (5 min)
	RSIPeriod       int     // 8
	VolumeThreshold float64 // 1.5 (150% of average)
}

// NewScalpingStrategy creates a scalping strategy instance
func NewScalpingStrategy() *ScalpingStrategy {
	return &ScalpingStrategy{
		Name:            "Scalping",
		Description:     "High-frequency micro-trades on order book imbalances",
		Enabled:         true,
		MinGain:         0.002,
		MaxHold:         300,
		RSIPeriod:       8,
		VolumeThreshold: 1.5,
	}
}

// Generate creates trade signals based on RSI oversold/overbought + volume spikes
func (s *ScalpingStrategy) Generate(marketData *MarketData) (*TradeSignal, error) {
	if !s.Enabled {
		return nil, fmt.Errorf("strategy disabled")
	}

	// Calculate RSI(8)
	rsi := s.calculateRSI(marketData.PriceHistory, s.RSIPeriod)

	// Check volume spike
	avgVolume := s.calculateAvgVolume(marketData.VolumeHistory)
	volumeRatio := marketData.CurrentVolume / avgVolume

	// Oversold + volume spike = BUY
	if rsi < 30 && volumeRatio > s.VolumeThreshold {
		return &TradeSignal{
			Action:      "buy",
			Symbol:      marketData.Symbol,
			Confidence:  s.calculateConfidence(rsi, volumeRatio),
			Reasoning:   fmt.Sprintf("RSI oversold (%.2f) + volume spike (%.1fx avg)", rsi, volumeRatio),
			Strategy:    s.Name,
			Timestamp:   time.Now(),
			TargetGain:  s.MinGain,
			StopLoss:    -0.005, // -0.5% stop loss (tight for scalping)
			MaxHoldTime: s.MaxHold,
		}, nil
	}

	// Overbought + volume spike = SELL
	if rsi > 70 && volumeRatio > s.VolumeThreshold {
		return &TradeSignal{
			Action:      "sell",
			Symbol:      marketData.Symbol,
			Confidence:  s.calculateConfidence(rsi, volumeRatio),
			Reasoning:   fmt.Sprintf("RSI overbought (%.2f) + volume spike (%.1fx avg)", rsi, volumeRatio),
			Strategy:    s.Name,
			Timestamp:   time.Now(),
			TargetGain:  s.MinGain,
			StopLoss:    -0.005,
			MaxHoldTime: s.MaxHold,
		}, nil
	}

	return nil, nil // No signal
}

// Analyze evaluates market conditions for scalping opportunities
func (s *ScalpingStrategy) Analyze(marketData *MarketData) *StrategyAnalysis {
	rsi := s.calculateRSI(marketData.PriceHistory, s.RSIPeriod)
	avgVolume := s.calculateAvgVolume(marketData.VolumeHistory)
	volumeRatio := marketData.CurrentVolume / avgVolume

	score := 0.0
	if rsi < 35 || rsi > 65 {
		score += 0.4 // RSI extreme
	}
	if volumeRatio > s.VolumeThreshold {
		score += 0.6 // Volume spike
	}

	return &StrategyAnalysis{
		StrategyName: s.Name,
		Score:        score,
		Indicators: map[string]float64{
			"rsi":          rsi,
			"volume_ratio": volumeRatio,
		},
		Recommendation: s.getRecommendation(score),
		Timestamp:      time.Now(),
	}
}

// calculateRSI computes Relative Strength Index
func (s *ScalpingStrategy) calculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50.0 // Neutral if insufficient data
	}

	gains := 0.0
	losses := 0.0

	for i := len(prices) - period; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// calculateAvgVolume computes average volume over recent periods
func (s *ScalpingStrategy) calculateAvgVolume(volumes []float64) float64 {
	if len(volumes) == 0 {
		return 1.0
	}

	sum := 0.0
	for _, v := range volumes {
		sum += v
	}

	return sum / float64(len(volumes))
}

// calculateConfidence determines signal strength (0.0 - 1.0)
func (s *ScalpingStrategy) calculateConfidence(rsi float64, volumeRatio float64) float64 {
	rsiStrength := 0.0
	if rsi < 30 {
		rsiStrength = (30 - rsi) / 30 // More oversold = higher confidence
	} else if rsi > 70 {
		rsiStrength = (rsi - 70) / 30 // More overbought = higher confidence
	}

	volumeStrength := (volumeRatio - 1.0) / 2.0 // 2x volume = 50% strength
	if volumeStrength > 1.0 {
		volumeStrength = 1.0
	}

	return (rsiStrength*0.6 + volumeStrength*0.4) // RSI weighted more heavily
}

// getRecommendation translates score to action recommendation
func (s *ScalpingStrategy) getRecommendation(score float64) string {
	if score > 0.7 {
		return "STRONG_SIGNAL"
	} else if score > 0.4 {
		return "MODERATE_SIGNAL"
	}
	return "WEAK_SIGNAL"
}

// GetConfig returns strategy configuration
func (s *ScalpingStrategy) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":             s.Name,
		"enabled":          s.Enabled,
		"min_gain":         s.MinGain,
		"max_hold_seconds": s.MaxHold,
		"rsi_period":       s.RSIPeriod,
		"volume_threshold": s.VolumeThreshold,
	}
}

// UpdateConfig updates strategy parameters (for GRPO optimization)
func (s *ScalpingStrategy) UpdateConfig(params map[string]interface{}) error {
	if val, ok := params["rsi_period"].(int); ok {
		s.RSIPeriod = val
	}
	if val, ok := params["volume_threshold"].(float64); ok {
		s.VolumeThreshold = val
	}
	if val, ok := params["min_gain"].(float64); ok {
		s.MinGain = val
	}
	return nil
}
