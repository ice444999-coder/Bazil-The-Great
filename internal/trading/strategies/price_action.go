package strategies

import (
	"fmt"
	"math"
	"time"
)

// PriceActionStrategy implements price action trading using candlestick patterns and volume analysis
// Target: 2x bi-monthly returns, 5min-1h timeframe
// Focus: Candlestick patterns + volume confirmation + support/resistance
type PriceActionStrategy struct {
	config map[string]interface{}
}

// NewPriceActionStrategy creates a new price action strategy
func NewPriceActionStrategy() *PriceActionStrategy {
	return &PriceActionStrategy{
		config: map[string]interface{}{
			"minGain":                 0.005, // 0.5% min gain
			"maxHold":                 7200,  // 2 hours max hold
			"stopLoss":                -0.02, // 2% stop loss
			"targetGain":              0.04,  // 4% target gain
			"candlestickWeight":       0.6,   // Weight for candlestick patterns
			"volumeWeight":            0.4,   // Weight for volume confirmation
			"supportResistanceWeight": 0.3,   // Weight for S/R levels
			"patternThreshold":        0.7,   // Minimum pattern strength
			"volumeMultiplier":        1.3,   // Volume confirmation multiplier
		},
	}
}

// Generate generates trading signals based on price action analysis
func (s *PriceActionStrategy) Generate(data *MarketData) (*TradeSignal, error) {
	if len(data.PriceHistory) < 20 {
		return &TradeSignal{Action: "hold"}, nil
	}

	// Analyze candlestick patterns
	candlestickSignal := s.analyzeCandlestickPatterns(data)

	// Analyze volume confirmation
	volumeSignal := s.analyzeVolumeConfirmation(data)

	// Analyze support/resistance levels
	supportResistanceSignal := s.analyzeSupportResistance(data)

	// Calculate composite signal
	candlestickWeight := s.getConfigFloat("candlestickWeight")
	volumeWeight := s.getConfigFloat("volumeWeight")
	srWeight := s.getConfigFloat("supportResistanceWeight")

	compositeSignal := (candlestickSignal * candlestickWeight) +
		(volumeSignal * volumeWeight) +
		(supportResistanceSignal * srWeight)

	confidence := math.Abs(compositeSignal)

	// Check pattern threshold
	if confidence < s.getConfigFloat("patternThreshold") {
		return &TradeSignal{Action: "hold"}, nil
	}

	currentPrice := data.CurrentPrice
	targetGain := s.getConfigFloat("targetGain")
	stopLoss := s.getConfigFloat("stopLoss")

	action := "hold"
	reasoning := ""

	// BUY conditions: Bullish patterns + volume + support confirmation
	if compositeSignal > s.getConfigFloat("patternThreshold") {
		action = "buy"
		reasoning = fmt.Sprintf("Bullish price action: Candles(%.2f) + Volume(%.2f) + S/R(%.2f) = %.2f confidence",
			candlestickSignal, volumeSignal, supportResistanceSignal, confidence)
	}

	// SELL conditions: Bearish patterns + volume + resistance confirmation
	if compositeSignal < -s.getConfigFloat("patternThreshold") {
		action = "sell"
		reasoning = fmt.Sprintf("Bearish price action: Candles(%.2f) + Volume(%.2f) + S/R(%.2f) = %.2f confidence",
			candlestickSignal, volumeSignal, supportResistanceSignal, confidence)
	}

	// Additional exit conditions for open positions
	if action == "hold" {
		// Check for profit taking or stop loss based on price history
		if len(data.PriceHistory) >= 2 {
			prevPrice := data.PriceHistory[len(data.PriceHistory)-2]
			priceChange := (currentPrice - prevPrice) / prevPrice

			// Take profit if target reached
			if priceChange >= targetGain {
				action = "sell"
				confidence = 0.8
				reasoning = fmt.Sprintf("Profit target reached: +%.2f%%", priceChange*100)
			}

			// Stop loss if loss too large
			if priceChange <= stopLoss {
				action = "sell"
				confidence = 0.9
				reasoning = fmt.Sprintf("Stop loss triggered: %.2f%%", priceChange*100)
			}
		}
	}

	return &TradeSignal{
		Action:      action,
		Symbol:      data.Symbol,
		Confidence:  confidence,
		Reasoning:   reasoning,
		Strategy:    "price_action",
		Timestamp:   time.Now(),
		TargetGain:  targetGain,
		StopLoss:    stopLoss,
		MaxHoldTime: int(s.getConfigFloat("maxHold")),
		Priority:    6, // Medium-high priority for price action
	}, nil
}

// Analyze performs detailed analysis for the strategy
func (s *PriceActionStrategy) Analyze(data *MarketData) *StrategyAnalysis {
	signal, _ := s.Generate(data)

	score := 0.0
	recommendation := "NO_SIGNAL"

	// Calculate score based on signal confidence and market conditions
	if signal.Confidence > 0.7 {
		score = signal.Confidence * 100
		if signal.Action == "buy" {
			recommendation = "STRONG_BUY"
		} else if signal.Action == "sell" {
			recommendation = "STRONG_SELL"
		}
	} else if signal.Confidence > 0.5 {
		score = signal.Confidence * 80
		if signal.Action == "buy" {
			recommendation = "MODERATE_BUY"
		} else if signal.Action == "sell" {
			recommendation = "MODERATE_SELL"
		}
	}

	// Calculate current indicators for analysis
	indicators := map[string]float64{}
	if len(data.PriceHistory) >= 20 {
		candlestickSignal := s.analyzeCandlestickPatterns(data)
		volumeSignal := s.analyzeVolumeConfirmation(data)
		supportResistanceSignal := s.analyzeSupportResistance(data)

		indicators["candlestick_signal"] = candlestickSignal
		indicators["volume_signal"] = volumeSignal
		indicators["support_resistance_signal"] = supportResistanceSignal

		candlestickWeight := s.getConfigFloat("candlestickWeight")
		volumeWeight := s.getConfigFloat("volumeWeight")
		srWeight := s.getConfigFloat("supportResistanceWeight")

		indicators["composite_signal"] = (candlestickSignal*candlestickWeight +
			volumeSignal*volumeWeight +
			supportResistanceSignal*srWeight)
	}

	return &StrategyAnalysis{
		StrategyName:   "price_action",
		Score:          score,
		Indicators:     indicators,
		Recommendation: recommendation,
		Timestamp:      time.Now(),
	}
}

// GetConfig returns the strategy configuration
func (s *PriceActionStrategy) GetConfig() map[string]interface{} {
	return s.config
}

// UpdateConfig updates the strategy configuration
func (s *PriceActionStrategy) UpdateConfig(params map[string]interface{}) error {
	for k, v := range params {
		s.config[k] = v
	}
	return nil
}

// Helper functions

func (s *PriceActionStrategy) getConfigFloat(key string) float64 {
	if val, ok := s.config[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

// analyzeCandlestickPatterns analyzes recent candlestick patterns
func (s *PriceActionStrategy) analyzeCandlestickPatterns(data *MarketData) float64 {
	if len(data.PriceHistory) < 5 {
		return 0.0
	}

	patterns := []float64{}

	// Analyze last 3 candles for patterns
	for i := len(data.PriceHistory) - 4; i < len(data.PriceHistory)-1; i++ {
		if i >= 0 {
			pattern := s.identifyCandlestickPattern(data, i)
			patterns = append(patterns, pattern)
		}
	}

	// Average pattern strength
	if len(patterns) == 0 {
		return 0.0
	}

	totalStrength := 0.0
	for _, pattern := range patterns {
		totalStrength += pattern
	}

	return totalStrength / float64(len(patterns))
}

// identifyCandlestickPattern identifies patterns in a single candle
func (s *PriceActionStrategy) identifyCandlestickPattern(data *MarketData, index int) float64 {
	if index < 0 || index >= len(data.PriceHistory)-1 {
		return 0.0
	}

	current := data.PriceHistory[index]
	next := data.PriceHistory[index+1]

	// Simple pattern recognition (in reality, would need OHLC data)
	// Using price change as proxy for candlestick patterns

	change := (next - current) / current

	// Bullish patterns
	if change > 0.01 { // Large upward move
		return 0.8 // Bullish engulfing or marubozu
	}
	if change > 0.005 { // Moderate upward move
		return 0.4 // Bullish candle
	}

	// Bearish patterns
	if change < -0.01 { // Large downward move
		return -0.8 // Bearish engulfing or marubozu
	}
	if change < -0.005 { // Moderate downward move
		return -0.4 // Bearish candle
	}

	// Doji or indecision
	return 0.0
}

// analyzeVolumeConfirmation analyzes volume for signal confirmation
func (s *PriceActionStrategy) analyzeVolumeConfirmation(data *MarketData) float64 {
	if len(data.VolumeHistory) < 10 {
		return 0.0
	}

	// Get recent volume
	currentVolume := data.VolumeHistory[len(data.VolumeHistory)-1]
	avgVolume := s.calculateAverageFloat(data.VolumeHistory[len(data.VolumeHistory)-10:])

	if avgVolume == 0 {
		return 0.0
	}

	volumeRatio := currentVolume / avgVolume
	threshold := s.getConfigFloat("volumeMultiplier")

	// Strong volume confirmation
	if volumeRatio > threshold*1.5 {
		return 0.9
	}

	// Moderate volume confirmation
	if volumeRatio > threshold {
		return 0.6
	}

	// Weak volume
	if volumeRatio < 0.7 {
		return -0.3 // Low volume = weak signal
	}

	return 0.0
}

// analyzeSupportResistance analyzes support and resistance levels
func (s *PriceActionStrategy) analyzeSupportResistance(data *MarketData) float64 {
	if len(data.PriceHistory) < 20 {
		return 0.0
	}

	currentPrice := data.CurrentPrice

	// Find recent swing highs and lows
	swingHighs, swingLows := s.findSwingPoints(data, 5)

	// Check proximity to support/resistance
	nearSupport := false
	nearResistance := false

	for _, low := range swingLows {
		if math.Abs(currentPrice-low)/low < 0.01 { // Within 1% of support
			nearSupport = true
			break
		}
	}

	for _, high := range swingHighs {
		if math.Abs(currentPrice-high)/high < 0.01 { // Within 1% of resistance
			nearResistance = true
			break
		}
	}

	// Calculate signal based on S/R proximity
	if nearSupport {
		return 0.7 // Bullish (near support)
	}
	if nearResistance {
		return -0.7 // Bearish (near resistance)
	}

	return 0.0 // Not near any significant level
}

// findSwingPoints finds recent swing highs and lows
func (s *PriceActionStrategy) findSwingPoints(data *MarketData, lookback int) ([]float64, []float64) {
	swingHighs := []float64{}
	swingLows := []float64{}

	for i := lookback; i < len(data.PriceHistory)-lookback; i++ {
		price := data.PriceHistory[i]

		// Check for swing high
		isHigh := true
		for j := i - lookback; j <= i+lookback; j++ {
			if j != i && data.PriceHistory[j] >= price {
				isHigh = false
				break
			}
		}
		if isHigh {
			swingHighs = append(swingHighs, price)
		}

		// Check for swing low
		isLow := true
		for j := i - lookback; j <= i+lookback; j++ {
			if j != i && data.PriceHistory[j] <= price {
				isLow = false
				break
			}
		}
		if isLow {
			swingLows = append(swingLows, price)
		}
	}

	return swingHighs, swingLows
}

// calculateAverageFloat calculates simple moving average for float64
func (s *PriceActionStrategy) calculateAverageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}
