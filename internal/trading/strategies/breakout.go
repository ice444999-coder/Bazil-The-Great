package strategies

import (
	"fmt"
	"math"
	"time"
)

// BreakoutStrategy implements breakout trading based on volatility squeezes and ATR
// Target: 2x returns in volatile conditions, 15min-4h timeframe
// Focus: Bollinger Band squeezes + ATR breakouts
type BreakoutStrategy struct {
	config map[string]interface{}
}

// NewBreakoutStrategy creates a new breakout strategy
func NewBreakoutStrategy() *BreakoutStrategy {
	return &BreakoutStrategy{
		config: map[string]interface{}{
			"minGain":          0.01,  // 1% min gain
			"maxHold":          14400, // 4 hours max hold
			"stopLoss":         -0.03, // 3% stop loss
			"targetGain":       0.05,  // 5% target gain
			"bollingerPeriod":  20,
			"bollingerStdDev":  2.0,
			"atrPeriod":        14,
			"volumeMultiplier": 1.5,
			"squeezeThreshold": 0.8, // Band width threshold for squeeze
		},
	}
}

// Generate generates trading signals based on breakout analysis
func (s *BreakoutStrategy) Generate(data *MarketData) (*TradeSignal, error) {
	if len(data.PriceHistory) < 50 {
		return &TradeSignal{Action: "hold"}, nil
	}

	// Calculate indicators
	bollinger := s.calculateBollingerBands(data.PriceHistory, 20, 2.0)
	atr := s.calculateATR(data, 14)

	// Get current values
	currentPrice := data.CurrentPrice
	currentUpper := bollinger[len(bollinger)-1].Upper
	currentLower := bollinger[len(bollinger)-1].Lower
	currentMiddle := bollinger[len(bollinger)-1].Middle
	currentATR := atr[len(atr)-1]

	// Check for Bollinger Band squeeze (bands are tight)
	bandWidth := (currentUpper - currentLower) / currentMiddle
	squeezeDetected := bandWidth < s.getConfigFloat("squeezeThreshold")

	// Volume confirmation
	avgVolume := s.calculateAverageFloat(data.VolumeHistory[len(data.VolumeHistory)-20:])
	volumeConfirmed := len(data.VolumeHistory) > 0 && data.VolumeHistory[len(data.VolumeHistory)-1] > avgVolume*s.getConfigFloat("volumeMultiplier")

	// Check for breakout conditions
	breakoutUp := currentPrice > currentUpper
	breakoutDown := currentPrice < currentLower

	// Additional confirmation: price moved more than ATR
	priceMove := math.Abs(currentPrice-data.PriceHistory[len(data.PriceHistory)-2]) / data.PriceHistory[len(data.PriceHistory)-2]
	atrBreakout := priceMove > currentATR*0.5 // Price moved more than 0.5 ATR

	// Generate signal
	action := "hold"
	confidence := 0.0
	reasoning := ""
	targetGain := s.getConfigFloat("targetGain")
	stopLoss := s.getConfigFloat("stopLoss")

	// BUY conditions: Upper breakout after squeeze + volume + ATR confirmation
	if breakoutUp && squeezeDetected && volumeConfirmed && atrBreakout {
		action = "buy"
		confidence = 0.8
		reasoning = fmt.Sprintf("Bullish breakout above upper Bollinger Band (%.4f) after squeeze (width: %.3f) with volume and ATR confirmation",
			currentUpper, bandWidth)
	}

	// SELL conditions: Lower breakout after squeeze + volume + ATR confirmation
	if breakoutDown && squeezeDetected && volumeConfirmed && atrBreakout {
		action = "sell"
		confidence = 0.8
		reasoning = fmt.Sprintf("Bearish breakout below lower Bollinger Band (%.4f) after squeeze (width: %.3f) with volume and ATR confirmation",
			currentLower, bandWidth)
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
				confidence = 0.85
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
		Strategy:    "breakout",
		Timestamp:   time.Now(),
		TargetGain:  targetGain,
		StopLoss:    stopLoss,
		MaxHoldTime: int(s.getConfigFloat("maxHold")),
		Priority:    8, // High priority for breakout signals
	}, nil
}

// Analyze performs detailed analysis for the strategy
func (s *BreakoutStrategy) Analyze(data *MarketData) *StrategyAnalysis {
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
	if len(data.PriceHistory) >= 50 {
		bollinger := s.calculateBollingerBands(data.PriceHistory, 20, 2.0)
		atr := s.calculateATR(data, 14)

		if len(bollinger) > 0 {
			current := bollinger[len(bollinger)-1]
			indicators["bollinger_upper"] = current.Upper
			indicators["bollinger_middle"] = current.Middle
			indicators["bollinger_lower"] = current.Lower
			indicators["band_width"] = (current.Upper - current.Lower) / current.Middle
		}

		if len(atr) > 0 {
			indicators["atr"] = atr[len(atr)-1]
		}

		// Squeeze detection
		indicators["squeeze_detected"] = boolToFloat(indicators["band_width"] < s.getConfigFloat("squeezeThreshold"))

		// Volume confirmation
		avgVolume := s.calculateAverageFloat(data.VolumeHistory[len(data.VolumeHistory)-20:])
		indicators["volume_confirmed"] = boolToFloat(len(data.VolumeHistory) > 0 && data.VolumeHistory[len(data.VolumeHistory)-1] > avgVolume*s.getConfigFloat("volumeMultiplier"))
	}

	return &StrategyAnalysis{
		StrategyName:   "breakout",
		Score:          score,
		Indicators:     indicators,
		Recommendation: recommendation,
		Timestamp:      time.Now(),
	}
}

// GetConfig returns the strategy configuration
func (s *BreakoutStrategy) GetConfig() map[string]interface{} {
	return s.config
}

// UpdateConfig updates the strategy configuration
func (s *BreakoutStrategy) UpdateConfig(params map[string]interface{}) error {
	for k, v := range params {
		s.config[k] = v
	}
	return nil
}

// Helper functions

func (s *BreakoutStrategy) getConfigFloat(key string) float64 {
	if val, ok := s.config[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

// calculateBollingerBands calculates Bollinger Bands
func (s *BreakoutStrategy) calculateBollingerBands(prices []float64, period int, stdDev float64) []BollingerData {
	if len(prices) < period {
		return nil
	}

	result := make([]BollingerData, 0, len(prices)-period+1)

	for i := period - 1; i < len(prices); i++ {
		slice := prices[i-period+1 : i+1]
		middle := s.calculateAverageFloat(slice)

		// Calculate standard deviation
		sumSquares := 0.0
		for _, price := range slice {
			sumSquares += math.Pow(price-middle, 2)
		}
		std := math.Sqrt(sumSquares / float64(len(slice)))

		result = append(result, BollingerData{
			Upper:  middle + stdDev*std,
			Middle: middle,
			Lower:  middle - stdDev*std,
		})
	}

	return result
}

// calculateATR calculates Average True Range
func (s *BreakoutStrategy) calculateATR(data *MarketData, period int) []float64 {
	// For simplicity, we'll use a basic ATR calculation
	// In a real implementation, you'd need high/low data
	// For now, return a mock ATR based on price volatility
	result := make([]float64, len(data.PriceHistory))

	for i := range result {
		if i == 0 {
			result[i] = math.Abs(data.CurrentPrice * 0.02) // 2% initial ATR
		} else {
			// Calculate true range (simplified)
			tr := math.Abs(data.PriceHistory[i] - data.PriceHistory[i-1])
			// Smooth with previous ATR
			result[i] = (result[i-1]*(float64(period)-1) + tr) / float64(period)
		}
	}

	return result
}

// calculateAverageFloat calculates simple moving average for float64
func (s *BreakoutStrategy) calculateAverageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// BollingerData holds Bollinger Band values
type BollingerData struct {
	Upper  float64 `json:"upper"`
	Middle float64 `json:"middle"`
	Lower  float64 `json:"lower"`
}
