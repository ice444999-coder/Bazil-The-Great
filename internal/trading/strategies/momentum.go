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

// MomentumStrategy - Ride price trends with volume confirmation
// Target: 30-80% annual returns by catching strong trends
// Timeframe: 15min - 4h
// Indicators: MACD, Volume, Price velocity, Trend strength
type MomentumStrategy struct {
	Name             string
	Description      string
	Enabled          bool
	MACDFast         int     // 12
	MACDSlow         int     // 26
	MACDSignal       int     // 9
	VolumeMultiplier float64 // 2.0 (200% of average)
	MinMomentumScore float64 // 0.6
}

// NewMomentumStrategy creates a momentum strategy instance
func NewMomentumStrategy() *MomentumStrategy {
	return &MomentumStrategy{
		Name:             "Momentum",
		Description:      "Ride price trends with volume confirmation",
		Enabled:          true,
		MACDFast:         12,
		MACDSlow:         26,
		MACDSignal:       9,
		VolumeMultiplier: 2.0,
		MinMomentumScore: 0.6,
	}
}

// Generate creates trade signals based on momentum indicators
func (m *MomentumStrategy) Generate(marketData *MarketData) (*TradeSignal, error) {
	if !m.Enabled {
		return nil, fmt.Errorf("strategy disabled")
	}

	// Calculate MACD
	macdLine, signalLine, histogram := m.calculateMACD(marketData.PriceHistory)

	// Calculate volume momentum
	avgVolume := m.calculateAvgVolume(marketData.VolumeHistory)
	volumeRatio := marketData.CurrentVolume / avgVolume

	// Calculate price velocity (rate of change)
	priceVelocity := m.calculatePriceVelocity(marketData.PriceHistory)

	// Calculate overall momentum score
	momentumScore := m.calculateMomentumScore(histogram, volumeRatio, priceVelocity)

	// Bullish momentum: MACD crossover + high volume + positive velocity
	if histogram > 0 && macdLine > signalLine && volumeRatio > m.VolumeMultiplier && priceVelocity > 0 {
		if momentumScore >= m.MinMomentumScore {
			return &TradeSignal{
				Action:      "buy",
				Symbol:      marketData.Symbol,
				Confidence:  momentumScore,
				Reasoning:   fmt.Sprintf("Strong bullish momentum: MACD crossover (+%.4f), %.1fx volume, velocity +%.2f%%", histogram, volumeRatio, priceVelocity*100),
				Strategy:    m.Name,
				Timestamp:   time.Now(),
				TargetGain:  0.08,  // 8% target for momentum trades
				StopLoss:    -0.03, // -3% stop loss
				MaxHoldTime: 14400, // 4 hours
				Priority:    7,
			}, nil
		}
	}

	// Bearish momentum: MACD cross under + high volume + negative velocity
	if histogram < 0 && macdLine < signalLine && volumeRatio > m.VolumeMultiplier && priceVelocity < 0 {
		if momentumScore >= m.MinMomentumScore {
			return &TradeSignal{
				Action:      "sell",
				Symbol:      marketData.Symbol,
				Confidence:  momentumScore,
				Reasoning:   fmt.Sprintf("Strong bearish momentum: MACD cross under (%.4f), %.1fx volume, velocity %.2f%%", histogram, volumeRatio, priceVelocity*100),
				Strategy:    m.Name,
				Timestamp:   time.Now(),
				TargetGain:  0.08,
				StopLoss:    -0.03,
				MaxHoldTime: 14400,
				Priority:    7,
			}, nil
		}
	}

	return nil, nil // No momentum signal
}

// Analyze evaluates momentum conditions
func (m *MomentumStrategy) Analyze(marketData *MarketData) *StrategyAnalysis {
	macdLine, signalLine, histogram := m.calculateMACD(marketData.PriceHistory)
	avgVolume := m.calculateAvgVolume(marketData.VolumeHistory)
	volumeRatio := marketData.CurrentVolume / avgVolume
	priceVelocity := m.calculatePriceVelocity(marketData.PriceHistory)

	momentumScore := m.calculateMomentumScore(histogram, volumeRatio, priceVelocity)

	return &StrategyAnalysis{
		StrategyName: m.Name,
		Score:        momentumScore,
		Indicators: map[string]float64{
			"macd_line":      macdLine,
			"signal_line":    signalLine,
			"histogram":      histogram,
			"volume_ratio":   volumeRatio,
			"price_velocity": priceVelocity,
		},
		Recommendation: m.getRecommendation(momentumScore),
		Timestamp:      time.Now(),
	}
}

// calculateMACD computes Moving Average Convergence Divergence
func (m *MomentumStrategy) calculateMACD(prices []float64) (float64, float64, float64) {
	if len(prices) < m.MACDSlow {
		return 0, 0, 0 // Insufficient data
	}

	// Calculate EMA for fast and slow periods
	emaFast := m.calculateEMA(prices, m.MACDFast)
	emaSlow := m.calculateEMA(prices, m.MACDSlow)

	// MACD line = Fast EMA - Slow EMA
	macdLine := emaFast - emaSlow

	// Signal line = EMA of MACD line (simplified: use recent MACD values)
	signalLine := macdLine * 0.9 // Simplified for now

	// Histogram = MACD - Signal
	histogram := macdLine - signalLine

	return macdLine, signalLine, histogram
}

// calculateEMA computes Exponential Moving Average
func (m *MomentumStrategy) calculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return prices[len(prices)-1] // Return last price if insufficient data
	}

	multiplier := 2.0 / float64(period+1)
	ema := prices[len(prices)-period] // Start with first price in period

	for i := len(prices) - period + 1; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// calculatePriceVelocity computes rate of price change (momentum)
func (m *MomentumStrategy) calculatePriceVelocity(prices []float64) float64 {
	if len(prices) < 10 {
		return 0 // Insufficient data
	}

	// Calculate % change over last 10 periods
	oldPrice := prices[len(prices)-10]
	newPrice := prices[len(prices)-1]

	if oldPrice == 0 {
		return 0
	}

	return (newPrice - oldPrice) / oldPrice // Decimal change (0.05 = 5%)
}

// calculateAvgVolume computes average volume
func (m *MomentumStrategy) calculateAvgVolume(volumes []float64) float64 {
	if len(volumes) == 0 {
		return 1.0
	}

	sum := 0.0
	for _, v := range volumes {
		sum += v
	}

	return sum / float64(len(volumes))
}

// calculateMomentumScore combines all indicators into unified score
func (m *MomentumStrategy) calculateMomentumScore(histogram float64, volumeRatio float64, priceVelocity float64) float64 {
	// MACD histogram strength (0.0 - 0.4)
	histogramStrength := 0.0
	if histogram > 0.001 || histogram < -0.001 {
		histogramStrength = 0.4
	} else if histogram > 0.0005 || histogram < -0.0005 {
		histogramStrength = 0.2
	}

	// Volume strength (0.0 - 0.3)
	volumeStrength := 0.0
	if volumeRatio > m.VolumeMultiplier {
		volumeStrength = 0.3
	} else if volumeRatio > m.VolumeMultiplier*0.75 {
		volumeStrength = 0.15
	}

	// Price velocity strength (0.0 - 0.3)
	velocityStrength := 0.0
	absVelocity := priceVelocity
	if absVelocity < 0 {
		absVelocity = -absVelocity
	}

	if absVelocity > 0.05 { // 5%+ move
		velocityStrength = 0.3
	} else if absVelocity > 0.02 { // 2%+ move
		velocityStrength = 0.15
	}

	return histogramStrength + volumeStrength + velocityStrength
}

// getRecommendation translates score to action recommendation
func (m *MomentumStrategy) getRecommendation(score float64) string {
	if score >= 0.7 {
		return "STRONG_SIGNAL"
	} else if score >= 0.5 {
		return "MODERATE_SIGNAL"
	} else if score >= 0.3 {
		return "WEAK_SIGNAL"
	}
	return "NO_SIGNAL"
}

// GetConfig returns strategy configuration
func (m *MomentumStrategy) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":               m.Name,
		"enabled":            m.Enabled,
		"macd_fast":          m.MACDFast,
		"macd_slow":          m.MACDSlow,
		"macd_signal":        m.MACDSignal,
		"volume_multiplier":  m.VolumeMultiplier,
		"min_momentum_score": m.MinMomentumScore,
	}
}

// UpdateConfig updates strategy parameters
func (m *MomentumStrategy) UpdateConfig(params map[string]interface{}) error {
	if val, ok := params["macd_fast"].(int); ok {
		m.MACDFast = val
	}
	if val, ok := params["macd_slow"].(int); ok {
		m.MACDSlow = val
	}
	if val, ok := params["volume_multiplier"].(float64); ok {
		m.VolumeMultiplier = val
	}
	if val, ok := params["min_momentum_score"].(float64); ok {
		m.MinMomentumScore = val
	}
	return nil
}
