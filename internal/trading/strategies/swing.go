package strategies

import (
	"fmt"
	"math"
	"time"
)

// SwingStrategy implements swing trading with MACD crossovers and ATR trailing stops
// Target: 250% returns in trending markets, 4h-1d timeframe
// Focus: MACD crossovers + ATR trailing stops + trend confirmation
type SwingStrategy struct {
	config map[string]interface{}
}

// NewSwingStrategy creates a new swing strategy
func NewSwingStrategy() *SwingStrategy {
	return &SwingStrategy{
		config: map[string]interface{}{
			"minGain":         0.03,  // 3% min gain
			"maxHold":         86400, // 24 hours max hold
			"stopLoss":        -0.08, // 8% stop loss
			"targetGain":      0.15,  // 15% target gain
			"macdFast":        12,
			"macdSlow":        26,
			"macdSignal":      9,
			"atrPeriod":       14,
			"trailingATRMult": 2.0, // ATR multiplier for trailing stop
			"trendPeriod":     50,  // Period for trend calculation
		},
	}
}

// Generate generates trading signals based on swing analysis
func (s *SwingStrategy) Generate(data *MarketData) (*TradeSignal, error) {
	if len(data.PriceHistory) < 100 {
		return &TradeSignal{Action: "hold"}, nil
	}

	// Calculate indicators
	macd := s.calculateMACD(data.PriceHistory, 12, 26, 9)
	atr := s.calculateATR(data, 14)
	trend := s.calculateTrendStrength(data, 50)

	// Get current values
	currentPrice := data.CurrentPrice
	currentMACD := macd[len(macd)-1]
	currentATR := atr[len(atr)-1]
	trendStrength := trend

	// MACD crossover signals
	macdSignal := ""
	if len(macd) >= 3 {
		prevMACD := macd[len(macd)-2]
		prevSignal := macd[len(macd)-3].Signal

		// Bullish crossover: MACD crosses above signal line
		if prevMACD.MACD <= prevSignal && currentMACD.MACD > currentMACD.Signal {
			macdSignal = "bullish"
		}
		// Bearish crossover: MACD crosses below signal line
		if prevMACD.MACD >= prevSignal && currentMACD.MACD < currentMACD.Signal {
			macdSignal = "bearish"
		}
	}

	// Trend confirmation (strong uptrend or downtrend)
	uptrend := trendStrength > 0.6
	downtrend := trendStrength < -0.6

	// Volume confirmation (above average)
	avgVolume := s.calculateAverageFloat(data.VolumeHistory[len(data.VolumeHistory)-20:])
	volumeConfirmed := len(data.VolumeHistory) > 0 && data.VolumeHistory[len(data.VolumeHistory)-1] > avgVolume*1.2

	// Generate signal
	action := "hold"
	confidence := 0.0
	reasoning := ""
	targetGain := s.getConfigFloat("targetGain")
	stopLoss := s.getConfigFloat("stopLoss")

	// BUY conditions: Bullish MACD crossover + uptrend + volume + ATR confirmation
	if macdSignal == "bullish" && uptrend && volumeConfirmed && currentATR > 0 {
		action = "buy"
		confidence = 0.75
		reasoning = fmt.Sprintf("Bullish MACD crossover (%.4f) in uptrend (%.2f) with volume confirmation. ATR: %.4f",
			currentMACD.MACD, trendStrength, currentATR)
	}

	// SELL conditions: Bearish MACD crossover + downtrend + volume + ATR confirmation
	if macdSignal == "bearish" && downtrend && volumeConfirmed && currentATR > 0 {
		action = "sell"
		confidence = 0.75
		reasoning = fmt.Sprintf("Bearish MACD crossover (%.4f) in downtrend (%.2f) with volume confirmation. ATR: %.4f",
			currentMACD.MACD, trendStrength, currentATR)
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
		Strategy:    "swing",
		Timestamp:   time.Now(),
		TargetGain:  targetGain,
		StopLoss:    stopLoss,
		MaxHoldTime: int(s.getConfigFloat("maxHold")),
		Priority:    6, // Medium priority for swing trades
	}, nil
}

// Analyze performs detailed analysis for the strategy
func (s *SwingStrategy) Analyze(data *MarketData) *StrategyAnalysis {
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
	if len(data.PriceHistory) >= 100 {
		macd := s.calculateMACD(data.PriceHistory, 12, 26, 9)
		atr := s.calculateATR(data, 14)
		trend := s.calculateTrendStrength(data, 50)

		if len(macd) > 0 {
			indicators["macd"] = macd[len(macd)-1].MACD
			indicators["macd_signal"] = macd[len(macd)-1].Signal
		}

		if len(atr) > 0 {
			indicators["atr"] = atr[len(atr)-1]
		}

		indicators["trend_strength"] = trend

		// Volume confirmation
		avgVolume := s.calculateAverageFloat(data.VolumeHistory[len(data.VolumeHistory)-20:])
		indicators["volume_confirmed"] = boolToFloat(len(data.VolumeHistory) > 0 && data.VolumeHistory[len(data.VolumeHistory)-1] > avgVolume*1.2)
	}

	return &StrategyAnalysis{
		StrategyName:   "swing",
		Score:          score,
		Indicators:     indicators,
		Recommendation: recommendation,
		Timestamp:      time.Now(),
	}
}

// GetConfig returns the strategy configuration
func (s *SwingStrategy) GetConfig() map[string]interface{} {
	return s.config
}

// UpdateConfig updates the strategy configuration
func (s *SwingStrategy) UpdateConfig(params map[string]interface{}) error {
	for k, v := range params {
		s.config[k] = v
	}
	return nil
}

// Helper functions

func (s *SwingStrategy) getConfigFloat(key string) float64 {
	if val, ok := s.config[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

// calculateMACD calculates MACD indicator
func (s *SwingStrategy) calculateMACD(prices []float64, fast, slow, signal int) []MACDData {
	if len(prices) < slow+signal {
		return nil
	}

	result := make([]MACDData, 0)

	// Calculate EMAs
	fastEMA := s.calculateEMAFloat(prices, fast)
	slowEMA := s.calculateEMAFloat(prices, slow)

	// Calculate MACD line
	macdLine := make([]float64, len(fastEMA))
	for i := 0; i < len(fastEMA); i++ {
		macdLine[i] = fastEMA[i] - slowEMA[i]
	}

	// Calculate signal line (EMA of MACD)
	signalEMA := s.calculateEMAFloat(macdLine, signal)

	// Calculate histogram
	histogram := make([]float64, len(macdLine))
	for i := 0; i < len(signalEMA); i++ {
		histogram[i] = macdLine[i] - signalEMA[i]
	}

	// Build result
	for i := 0; i < len(macdLine); i++ {
		data := MACDData{
			MACD:   macdLine[i],
			Signal: signalEMA[i],
		}
		if i < len(histogram) {
			data.Histogram = histogram[i]
		}
		result = append(result, data)
	}

	return result
}

// calculateATR calculates Average True Range
func (s *SwingStrategy) calculateATR(data *MarketData, period int) []float64 {
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

// calculateTrendStrength calculates trend strength using linear regression slope
func (s *SwingStrategy) calculateTrendStrength(data *MarketData, period int) float64 {
	if len(data.PriceHistory) < period {
		return 0.0
	}

	// Use recent prices for trend calculation
	prices := data.PriceHistory[len(data.PriceHistory)-period:]

	// Calculate linear regression slope
	n := float64(len(prices))
	sumX := n * (n - 1) / 2
	sumY := 0.0
	sumXY := 0.0
	sumXX := 0.0

	for i, price := range prices {
		x := float64(i)
		sumY += price
		sumXY += x * price
		sumXX += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)

	// Normalize slope by average price to get relative trend strength
	avgPrice := sumY / n
	if avgPrice == 0 {
		return 0.0
	}

	trendStrength := slope / avgPrice

	// Bound between -1 and 1
	if trendStrength > 1 {
		trendStrength = 1
	} else if trendStrength < -1 {
		trendStrength = -1
	}

	return trendStrength
}

// calculateEMAFloat calculates Exponential Moving Average for float64
func (s *SwingStrategy) calculateEMAFloat(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	result := make([]float64, 0, len(prices))
	multiplier := 2.0 / (float64(period) + 1.0)

	// First EMA is SMA
	sma := s.calculateAverageFloat(prices[:period])
	result = append(result, sma)

	// Calculate subsequent EMAs
	for i := period; i < len(prices); i++ {
		ema := (prices[i] * multiplier) + (result[len(result)-1] * (1 - multiplier))
		result = append(result, ema)
	}

	return result
}

// calculateAverageFloat calculates simple moving average for float64
func (s *SwingStrategy) calculateAverageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}
