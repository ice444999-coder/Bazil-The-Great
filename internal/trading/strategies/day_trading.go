package strategies

import (
	"fmt"
	"math"
	"time"
)

// DayTradingStrategy implements day trading with MACD and ADX indicators
// Target: 30-60% annual returns, 5min-2h timeframe
// Focus: Intraday swings with trend confirmation
type DayTradingStrategy struct {
	config map[string]interface{}
}

// NewDayTradingStrategy creates a new day trading strategy
func NewDayTradingStrategy() *DayTradingStrategy {
	return &DayTradingStrategy{
		config: map[string]interface{}{
			"minGain":    0.005, // 0.5% min gain
			"maxHold":    28800, // 8 hours max hold
			"stopLoss":   -0.02, // 2% stop loss
			"targetGain": 0.03,  // 3% target gain
			"macdFast":   12,
			"macdSlow":   26,
			"macdSignal": 9,
			"adxPeriod":  14,
			"rsiPeriod":  14,
		},
	}
}

// Generate generates trading signals based on day trading analysis
func (s *DayTradingStrategy) Generate(data *MarketData) (*TradeSignal, error) {
	if len(data.PriceHistory) < 50 {
		return &TradeSignal{Action: "hold"}, nil
	}

	// Calculate indicators
	macd := s.calculateMACD(data.PriceHistory, 12, 26, 9)
	adx := s.calculateADX(data, 14)
	rsi := s.calculateRSI(data.PriceHistory, 14)

	// Get current values
	currentPrice := data.CurrentPrice
	currentMACD := macd[len(macd)-1]
	currentADX := adx[len(adx)-1]
	currentRSI := rsi[len(rsi)-1]

	// Trend strength (ADX > 25 indicates strong trend)
	trendStrength := currentADX > 25

	// Volume confirmation (above average)
	avgVolume := s.calculateAverageFloat(data.VolumeHistory[len(data.VolumeHistory)-20:])
	volumeConfirmed := len(data.VolumeHistory) > 0 && data.VolumeHistory[len(data.VolumeHistory)-1] > avgVolume

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

	// Generate signal
	action := "hold"
	confidence := 0.0
	reasoning := ""
	targetGain := s.getConfigFloat("targetGain")
	stopLoss := s.getConfigFloat("stopLoss")

	// BUY conditions: Bullish MACD crossover + strong trend + volume + oversold RSI
	if macdSignal == "bullish" && trendStrength && volumeConfirmed && currentRSI < 70 {
		action = "buy"
		confidence = 0.75
		reasoning = fmt.Sprintf("Bullish MACD crossover (%.4f) with strong trend (ADX: %.2f) and volume confirmation. RSI: %.2f",
			currentMACD.MACD, currentADX, currentRSI)
	}

	// SELL conditions: Bearish MACD crossover + strong trend + volume + overbought RSI
	if macdSignal == "bearish" && trendStrength && volumeConfirmed && currentRSI > 30 {
		action = "sell"
		confidence = 0.75
		reasoning = fmt.Sprintf("Bearish MACD crossover (%.4f) with strong trend (ADX: %.2f) and volume confirmation. RSI: %.2f",
			currentMACD.MACD, currentADX, currentRSI)
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
		Strategy:    "day_trading",
		Timestamp:   time.Now(),
		TargetGain:  targetGain,
		StopLoss:    stopLoss,
		MaxHoldTime: int(s.getConfigFloat("maxHold")),
		Priority:    7, // Medium-high priority for day trading
	}, nil
}

// Analyze performs detailed analysis for the strategy
func (s *DayTradingStrategy) Analyze(data *MarketData) *StrategyAnalysis {
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
		macd := s.calculateMACD(data.PriceHistory, 12, 26, 9)
		adx := s.calculateADX(data, 14)
		rsi := s.calculateRSI(data.PriceHistory, 14)

		if len(macd) > 0 {
			indicators["macd"] = macd[len(macd)-1].MACD
		}
		if len(adx) > 0 {
			indicators["adx"] = adx[len(adx)-1]
		}
		if len(rsi) > 0 {
			indicators["rsi"] = rsi[len(rsi)-1]
		}

		// Trend strength (ADX > 25 indicates strong trend)
		indicators["trend_strength"] = boolToFloat(indicators["adx"] > 25)

		// Volume confirmation (above average)
		avgVolume := s.calculateAverageFloat(data.VolumeHistory[len(data.VolumeHistory)-20:])
		indicators["volume_confirmed"] = boolToFloat(len(data.VolumeHistory) > 0 && data.VolumeHistory[len(data.VolumeHistory)-1] > avgVolume)
	}

	return &StrategyAnalysis{
		StrategyName:   "day_trading",
		Score:          score,
		Indicators:     indicators,
		Recommendation: recommendation,
		Timestamp:      time.Now(),
	}
}

// GetConfig returns the strategy configuration
func (s *DayTradingStrategy) GetConfig() map[string]interface{} {
	return s.config
}

// UpdateConfig updates the strategy configuration
func (s *DayTradingStrategy) UpdateConfig(params map[string]interface{}) error {
	for k, v := range params {
		s.config[k] = v
	}
	return nil
}

// Helper functions

func (s *DayTradingStrategy) getConfigFloat(key string) float64 {
	if val, ok := s.config[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// calculateMACD calculates MACD indicator
func (s *DayTradingStrategy) calculateMACD(prices []float64, fast, slow, signal int) []MACDData {
	if len(prices) < slow+signal {
		return nil
	}

	result := make([]MACDData, 0, len(prices))

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

// calculateADX calculates Average Directional Index
func (s *DayTradingStrategy) calculateADX(data *MarketData, period int) []float64 {
	// For simplicity, we'll use a basic ADX calculation
	// In a real implementation, you'd need high/low data
	// For now, return a mock ADX value
	result := make([]float64, len(data.PriceHistory))
	for i := range result {
		// Mock ADX - in reality this would be calculated from highs/lows
		result[i] = 25 + 10*math.Sin(float64(i)*0.1) // Oscillates around 25
	}
	return result
}

// calculateRSI calculates Relative Strength Index
func (s *DayTradingStrategy) calculateRSI(prices []float64, period int) []float64 {
	if len(prices) < period+1 {
		return nil
	}

	result := make([]float64, 0, len(prices)-period)

	gains := make([]float64, 0)
	losses := make([]float64, 0)

	// Calculate price changes
	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, -change)
		}
	}

	// Calculate initial averages
	avgGain := s.calculateAverageFloat(gains[:period])
	avgLoss := s.calculateAverageFloat(losses[:period])

	rsi := 100 - (100 / (1 + avgGain/avgLoss))
	result = append(result, rsi)

	// Calculate subsequent values using Wilder's smoothing
	for i := period; i < len(gains); i++ {
		avgGain = (avgGain*(float64(period)-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*(float64(period)-1) + losses[i]) / float64(period)

		if avgLoss > 0 {
			rs := avgGain / avgLoss
			rsi = 100 - (100 / (1 + rs))
		} else {
			rsi = 100
		}
		result = append(result, rsi)
	}

	return result
}

// calculateEMAFloat calculates Exponential Moving Average for float64
func (s *DayTradingStrategy) calculateEMAFloat(prices []float64, period int) []float64 {
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
func (s *DayTradingStrategy) calculateAverageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// MACDData holds MACD indicator values
type MACDData struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}
