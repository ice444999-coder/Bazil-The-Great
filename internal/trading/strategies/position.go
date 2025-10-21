/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package strategies

import (
	"fmt"
	"math"
	"time"
)

// PositionStrategy implements position trading based on macro trends and AI forecasts
// Target: 4x returns over years, weekly-monthly timeframe
// Focus: Macro trends + AI forecasts + fundamental analysis
type PositionStrategy struct {
	config map[string]interface{}
}

// NewPositionStrategy creates a new position strategy
func NewPositionStrategy() *PositionStrategy {
	return &PositionStrategy{
		config: map[string]interface{}{
			"minGain":           0.05,    // 5% min gain
			"maxHold":           2592000, // 30 days max hold
			"stopLoss":          -0.15,   // 15% stop loss
			"targetGain":        0.50,    // 50% target gain
			"trendPeriod":       200,     // Long-term trend period
			"forecastWeight":    0.6,     // AI forecast weight in decision
			"fundamentalWeight": 0.4,     // Fundamental analysis weight
			"momentumThreshold": 0.7,     // Momentum threshold for entry
			"volatilityFilter":  0.3,     // Maximum volatility for position trades
		},
	}
}

// Generate generates trading signals based on position analysis
func (s *PositionStrategy) Generate(data *MarketData) (*TradeSignal, error) {
	if len(data.PriceHistory) < 200 {
		return &TradeSignal{Action: "hold"}, nil
	}

	// Calculate long-term indicators
	longTrend := s.calculateLongTermTrend(data, 200)
	momentum := s.calculateMomentumScore(data, 100)
	volatility := s.calculateVolatility(data, 50)

	// Get AI forecast and fundamental analysis (mock for now)
	aiForecast := s.getAIForecast(data)
	fundamentalScore := s.getFundamentalScore(data)

	// Calculate composite score
	trendScore := longTrend * 0.3
	momentumScore := momentum * 0.3
	forecastScore := aiForecast * s.getConfigFloat("forecastWeight")
	fundamentalScoreWeighted := fundamentalScore * s.getConfigFloat("fundamentalWeight")

	compositeScore := trendScore + momentumScore + forecastScore + fundamentalScoreWeighted

	// Check volatility filter
	volatilityThreshold := s.getConfigFloat("volatilityFilter")
	volatilityOK := volatility < volatilityThreshold

	currentPrice := data.CurrentPrice
	targetGain := s.getConfigFloat("targetGain")
	stopLoss := s.getConfigFloat("stopLoss")

	action := "hold"
	confidence := 0.0
	reasoning := ""

	// BUY conditions: Strong bullish composite score + low volatility + positive momentum
	if compositeScore > s.getConfigFloat("momentumThreshold") && volatilityOK && momentum > 0.5 {
		action = "buy"
		confidence = math.Min(compositeScore*0.8, 0.9) // Cap at 90%
		reasoning = fmt.Sprintf("Strong position setup: Trend(%.2f) + Momentum(%.2f) + AI(%.2f) + Fundamental(%.2f) = %.2f. Volatility: %.2f",
			longTrend, momentum, aiForecast, fundamentalScore, compositeScore, volatility)
	}

	// SELL conditions: Strong bearish composite score + low volatility + negative momentum
	if compositeScore < -s.getConfigFloat("momentumThreshold") && volatilityOK && momentum < -0.5 {
		action = "sell"
		confidence = math.Min(math.Abs(compositeScore)*0.8, 0.9) // Cap at 90%
		reasoning = fmt.Sprintf("Strong position exit: Trend(%.2f) + Momentum(%.2f) + AI(%.2f) + Fundamental(%.2f) = %.2f. Volatility: %.2f",
			longTrend, momentum, aiForecast, fundamentalScore, compositeScore, volatility)
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
				confidence = 0.95
				reasoning = fmt.Sprintf("Stop loss triggered: %.2f%%", priceChange*100)
			}
		}
	}

	return &TradeSignal{
		Action:      action,
		Symbol:      data.Symbol,
		Confidence:  confidence,
		Reasoning:   reasoning,
		Strategy:    "position",
		Timestamp:   time.Now(),
		TargetGain:  targetGain,
		StopLoss:    stopLoss,
		MaxHoldTime: int(s.getConfigFloat("maxHold")),
		Priority:    4, // Lower priority for long-term positions
	}, nil
}

// Analyze performs detailed analysis for the strategy
func (s *PositionStrategy) Analyze(data *MarketData) *StrategyAnalysis {
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
	if len(data.PriceHistory) >= 200 {
		longTrend := s.calculateLongTermTrend(data, 200)
		momentum := s.calculateMomentumScore(data, 100)
		volatility := s.calculateVolatility(data, 50)
		aiForecast := s.getAIForecast(data)
		fundamentalScore := s.getFundamentalScore(data)

		indicators["long_term_trend"] = longTrend
		indicators["momentum_score"] = momentum
		indicators["volatility"] = volatility
		indicators["ai_forecast"] = aiForecast
		indicators["fundamental_score"] = fundamentalScore

		// Composite score calculation
		trendScore := longTrend * 0.3
		momentumScore := momentum * 0.3
		forecastScore := aiForecast * s.getConfigFloat("forecastWeight")
		fundamentalScoreWeighted := fundamentalScore * s.getConfigFloat("fundamentalWeight")
		indicators["composite_score"] = trendScore + momentumScore + forecastScore + fundamentalScoreWeighted
	}

	return &StrategyAnalysis{
		StrategyName:   "position",
		Score:          score,
		Indicators:     indicators,
		Recommendation: recommendation,
		Timestamp:      time.Now(),
	}
}

// GetConfig returns the strategy configuration
func (s *PositionStrategy) GetConfig() map[string]interface{} {
	return s.config
}

// UpdateConfig updates the strategy configuration
func (s *PositionStrategy) UpdateConfig(params map[string]interface{}) error {
	for k, v := range params {
		s.config[k] = v
	}
	return nil
}

// Helper functions

func (s *PositionStrategy) getConfigFloat(key string) float64 {
	if val, ok := s.config[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

// calculateLongTermTrend calculates long-term trend using linear regression
func (s *PositionStrategy) calculateLongTermTrend(data *MarketData, period int) float64 {
	if len(data.PriceHistory) < period {
		return 0.0
	}

	prices := data.PriceHistory[len(data.PriceHistory)-period:]

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
	avgPrice := sumY / n

	if avgPrice == 0 {
		return 0.0
	}

	// Normalize slope and bound between -1 and 1
	trendStrength := slope / avgPrice * 10 // Scale up for visibility
	if trendStrength > 1 {
		trendStrength = 1
	} else if trendStrength < -1 {
		trendStrength = -1
	}

	return trendStrength
}

// calculateMomentumScore calculates momentum score using ROC and RSI
func (s *PositionStrategy) calculateMomentumScore(data *MarketData, period int) float64 {
	if len(data.PriceHistory) < period+14 {
		return 0.0
	}

	// Rate of Change (ROC)
	oldPrice := data.PriceHistory[len(data.PriceHistory)-period-1]
	currentPrice := data.CurrentPrice
	roc := (currentPrice - oldPrice) / oldPrice

	// RSI
	rsi := s.calculateRSI(data.PriceHistory, 14)
	currentRSI := rsi[len(rsi)-1]

	// Normalize RSI to -1 to 1 scale
	rsiNormalized := (currentRSI - 50) / 50
	if rsiNormalized > 1 {
		rsiNormalized = 1
	} else if rsiNormalized < -1 {
		rsiNormalized = -1
	}

	// Combine ROC and RSI
	momentum := (roc * 0.6) + (rsiNormalized * 0.4)

	// Bound between -1 and 1
	if momentum > 1 {
		momentum = 1
	} else if momentum < -1 {
		momentum = -1
	}

	return momentum
}

// calculateVolatility calculates price volatility using standard deviation
func (s *PositionStrategy) calculateVolatility(data *MarketData, period int) float64 {
	if len(data.PriceHistory) < period {
		return 0.0
	}

	prices := data.PriceHistory[len(data.PriceHistory)-period:]
	avgPrice := s.calculateAverageFloat(prices)

	sumSquares := 0.0
	for _, price := range prices {
		sumSquares += math.Pow(price-avgPrice, 2)
	}

	variance := sumSquares / float64(len(prices))
	volatility := math.Sqrt(variance)

	// Normalize by average price
	if avgPrice == 0 {
		return 0.0
	}

	return volatility / avgPrice
}

// getAIForecast returns AI-based price forecast (mock implementation)
func (s *PositionStrategy) getAIForecast(data *MarketData) float64 {
	// Mock AI forecast - in reality this would use ML models
	// For now, use trend and momentum as forecast proxy

	if len(data.PriceHistory) < 50 {
		return 0.0
	}

	trend := s.calculateLongTermTrend(data, 50)
	momentum := s.calculateMomentumScore(data, 20)

	forecast := (trend * 0.7) + (momentum * 0.3)

	// Add some noise to simulate AI uncertainty
	forecast += (math.Sin(float64(time.Now().Unix())) * 0.1)

	if forecast > 1 {
		forecast = 1
	} else if forecast < -1 {
		forecast = -1
	}

	return forecast
}

// getFundamentalScore returns fundamental analysis score (mock implementation)
func (s *PositionStrategy) getFundamentalScore(data *MarketData) float64 {
	// Mock fundamental analysis - in reality this would analyze:
	// P/E ratio, earnings growth, debt levels, market position, etc.
	// For now, use market cap and volume as proxies

	if len(data.VolumeHistory) < 20 {
		return 0.0
	}

	// Volume trend as proxy for market interest
	volumeTrend := s.calculateLongTermTrend(&MarketData{
		PriceHistory:  []float64{},
		VolumeHistory: data.VolumeHistory,
		CurrentPrice:  data.VolumeHistory[len(data.VolumeHistory)-1],
	}, 20)

	// Price stability as proxy for fundamental strength
	volatility := s.calculateVolatility(data, 50)
	stability := 1.0 - volatility

	fundamental := (volumeTrend * 0.4) + (stability * 0.6)

	if fundamental > 1 {
		fundamental = 1
	} else if fundamental < -1 {
		fundamental = -1
	}

	return fundamental
}

// calculateRSI calculates Relative Strength Index
func (s *PositionStrategy) calculateRSI(prices []float64, period int) []float64 {
	if len(prices) < period+1 {
		return nil
	}

	result := make([]float64, 0)

	gains := make([]float64, 0)
	losses := make([]float64, 0)

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

	avgGain := s.calculateAverageFloat(gains[:period])
	avgLoss := s.calculateAverageFloat(losses[:period])

	rsi := 100 - (100 / (1 + avgGain/avgLoss))
	result = append(result, rsi)

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

// calculateAverageFloat calculates simple moving average for float64
func (s *PositionStrategy) calculateAverageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}
