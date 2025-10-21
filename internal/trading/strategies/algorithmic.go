/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package strategies

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// AlgorithmicStrategy implements algorithmic trading using ensemble methods and reinforcement learning
// Target: 50-100% net returns, various timeframes
// Focus: Ensemble models + RL optimization + adaptive algorithms
type AlgorithmicStrategy struct {
	config  map[string]interface{}
	rlModel *RLModel // Reinforcement learning model
}

// RLModel represents a simple reinforcement learning model
type RLModel struct {
	Weights      map[string]float64
	LearningRate float64
	Epsilon      float64 // Exploration rate
}

// NewAlgorithmicStrategy creates a new algorithmic strategy
func NewAlgorithmicStrategy() *AlgorithmicStrategy {
	return &AlgorithmicStrategy{
		config: map[string]interface{}{
			"minGain":             0.01,  // 1% min gain
			"maxHold":             3600,  // 1 hour max hold
			"stopLoss":            -0.05, // 5% stop loss
			"targetGain":          0.10,  // 10% target gain
			"ensembleSize":        5,     // Number of models in ensemble
			"rlLearningRate":      0.01,  // RL learning rate
			"rlEpsilon":           0.1,   // Exploration rate
			"confidenceThreshold": 0.6,   // Minimum confidence for trade
			"maxDrawdown":         0.1,   // Maximum drawdown before pause
		},
		rlModel: &RLModel{
			Weights: map[string]float64{
				"trend":      0.3,
				"momentum":   0.25,
				"volume":     0.2,
				"volatility": 0.15,
				"sentiment":  0.1,
			},
			LearningRate: 0.01,
			Epsilon:      0.1,
		},
	}
}

// Generate generates trading signals based on algorithmic analysis
func (s *AlgorithmicStrategy) Generate(data *MarketData) (*TradeSignal, error) {
	if len(data.PriceHistory) < 50 {
		return &TradeSignal{Action: "hold"}, nil
	}

	// Get ensemble predictions
	ensemblePredictions := s.getEnsemblePredictions(data)
	rlPrediction := s.getRLPrediction(data)

	// Combine predictions using weighted average
	ensembleWeight := 0.7
	rlWeight := 0.3

	finalPrediction := (ensemblePredictions["signal"] * ensembleWeight) +
		(rlPrediction * rlWeight)

	confidence := math.Abs(finalPrediction)

	// Check confidence threshold
	if confidence < s.getConfigFloat("confidenceThreshold") {
		return &TradeSignal{Action: "hold"}, nil
	}

	currentPrice := data.CurrentPrice
	targetGain := s.getConfigFloat("targetGain")
	stopLoss := s.getConfigFloat("stopLoss")

	action := "hold"
	reasoning := ""

	// BUY signal
	if finalPrediction > s.getConfigFloat("confidenceThreshold") {
		action = "buy"
		reasoning = fmt.Sprintf("Algorithmic BUY: Ensemble(%.2f) + RL(%.2f) = %.2f confidence",
			ensemblePredictions["signal"], rlPrediction, confidence)
	}

	// SELL signal
	if finalPrediction < -s.getConfigFloat("confidenceThreshold") {
		action = "sell"
		reasoning = fmt.Sprintf("Algorithmic SELL: Ensemble(%.2f) + RL(%.2f) = %.2f confidence",
			ensemblePredictions["signal"], rlPrediction, confidence)
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
		Strategy:    "algorithmic",
		Timestamp:   time.Now(),
		TargetGain:  targetGain,
		StopLoss:    stopLoss,
		MaxHoldTime: int(s.getConfigFloat("maxHold")),
		Priority:    7, // High priority for algorithmic signals
	}, nil
}

// Analyze performs detailed analysis for the strategy
func (s *AlgorithmicStrategy) Analyze(data *MarketData) *StrategyAnalysis {
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
		ensemblePredictions := s.getEnsemblePredictions(data)
		rlPrediction := s.getRLPrediction(data)

		indicators["ensemble_signal"] = ensemblePredictions["signal"]
		indicators["ensemble_confidence"] = ensemblePredictions["confidence"]
		indicators["rl_prediction"] = rlPrediction
		indicators["final_prediction"] = (ensemblePredictions["signal"]*0.7 + rlPrediction*0.3)
		indicators["model_diversity"] = ensemblePredictions["diversity"]
	}

	return &StrategyAnalysis{
		StrategyName:   "algorithmic",
		Score:          score,
		Indicators:     indicators,
		Recommendation: recommendation,
		Timestamp:      time.Now(),
	}
}

// GetConfig returns the strategy configuration
func (s *AlgorithmicStrategy) GetConfig() map[string]interface{} {
	return s.config
}

// UpdateConfig updates the strategy configuration
func (s *AlgorithmicStrategy) UpdateConfig(params map[string]interface{}) error {
	for k, v := range params {
		s.config[k] = v
	}
	return nil
}

// UpdateRLModel updates the reinforcement learning model based on trade outcome
func (s *AlgorithmicStrategy) UpdateRLModel(outcome float64, features map[string]float64) {
	// Simple Q-learning style update
	for feature, value := range features {
		if weight, exists := s.rlModel.Weights[feature]; exists {
			// Update weight based on outcome
			delta := s.rlModel.LearningRate * outcome * value
			s.rlModel.Weights[feature] = weight + delta
		}
	}

	// Decay epsilon over time (reduce exploration)
	s.rlModel.Epsilon *= 0.999
	if s.rlModel.Epsilon < 0.01 {
		s.rlModel.Epsilon = 0.01
	}
}

// Helper functions

func (s *AlgorithmicStrategy) getConfigFloat(key string) float64 {
	if val, ok := s.config[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

// getEnsemblePredictions gets predictions from multiple models
func (s *AlgorithmicStrategy) getEnsemblePredictions(data *MarketData) map[string]float64 {
	ensembleSize := int(s.getConfigFloat("ensembleSize"))
	predictions := make([]float64, ensembleSize)

	// Generate predictions from different models
	for i := 0; i < ensembleSize; i++ {
		predictions[i] = s.getSingleModelPrediction(data, i)
	}

	// Calculate ensemble statistics
	avgPrediction := s.calculateAverageFloat(predictions)

	// Calculate diversity (standard deviation of predictions)
	diversity := 0.0
	for _, pred := range predictions {
		diversity += math.Pow(pred-avgPrediction, 2)
	}
	diversity = math.Sqrt(diversity / float64(len(predictions)))

	// Confidence based on agreement (lower diversity = higher confidence)
	confidence := 1.0 / (1.0 + diversity)

	return map[string]float64{
		"signal":     avgPrediction,
		"confidence": confidence,
		"diversity":  diversity,
	}
}

// getSingleModelPrediction generates prediction from a single model variant
func (s *AlgorithmicStrategy) getSingleModelPrediction(data *MarketData, modelIndex int) float64 {
	// Different models use different combinations of indicators
	trend := s.calculateTrendStrength(data, 20+modelIndex*5)
	momentum := s.calculateMomentumScore(data, 10+modelIndex*2)
	volume := s.calculateVolumeScore(data)
	volatility := s.calculateVolatility(data, 10+modelIndex*3)

	var prediction float64

	switch modelIndex % 5 {
	case 0: // Trend-focused model
		prediction = trend*0.7 + momentum*0.3
	case 1: // Momentum-focused model
		prediction = momentum*0.7 + trend*0.3
	case 2: // Volume-focused model
		prediction = volume*0.6 + momentum*0.4
	case 3: // Volatility-adjusted model
		baseSignal := (trend + momentum) / 2
		prediction = baseSignal * (1 - volatility*0.5)
	case 4: // Mean-reversion model
		meanReversion := -s.calculateZScore(data, 20)
		prediction = meanReversion*0.6 + momentum*0.4
	}

	// Add some noise to simulate model differences
	rand.Seed(time.Now().UnixNano() + int64(modelIndex))
	noise := (rand.Float64() - 0.5) * 0.1
	prediction += noise

	// Bound between -1 and 1
	if prediction > 1 {
		prediction = 1
	} else if prediction < -1 {
		prediction = -1
	}

	return prediction
}

// getRLPrediction gets prediction from reinforcement learning model
func (s *AlgorithmicStrategy) getRLPrediction(data *MarketData) float64 {
	features := map[string]float64{
		"trend":      s.calculateTrendStrength(data, 20),
		"momentum":   s.calculateMomentumScore(data, 10),
		"volume":     s.calculateVolumeScore(data),
		"volatility": s.calculateVolatility(data, 10),
		"sentiment":  0.0, // Placeholder for sentiment analysis
	}

	// Calculate prediction using weighted features
	prediction := 0.0
	for feature, value := range features {
		if weight, exists := s.rlModel.Weights[feature]; exists {
			prediction += weight * value
		}
	}

	// Epsilon-greedy exploration
	if rand.Float64() < s.rlModel.Epsilon {
		prediction += (rand.Float64() - 0.5) * 0.4 // Random exploration
	}

	// Bound between -1 and 1
	if prediction > 1 {
		prediction = 1
	} else if prediction < -1 {
		prediction = -1
	}

	return prediction
}

// calculateTrendStrength calculates trend strength using linear regression slope
func (s *AlgorithmicStrategy) calculateTrendStrength(data *MarketData, period int) float64 {
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

	trendStrength := slope / avgPrice * 5 // Scale for visibility
	if trendStrength > 1 {
		trendStrength = 1
	} else if trendStrength < -1 {
		trendStrength = -1
	}

	return trendStrength
}

// calculateMomentumScore calculates momentum score
func (s *AlgorithmicStrategy) calculateMomentumScore(data *MarketData, period int) float64 {
	if len(data.PriceHistory) < period {
		return 0.0
	}

	oldPrice := data.PriceHistory[len(data.PriceHistory)-period]
	currentPrice := data.CurrentPrice

	momentum := (currentPrice - oldPrice) / oldPrice

	if momentum > 1 {
		momentum = 1
	} else if momentum < -1 {
		momentum = -1
	}

	return momentum
}

// calculateVolumeScore calculates volume-based score
func (s *AlgorithmicStrategy) calculateVolumeScore(data *MarketData) float64 {
	if len(data.VolumeHistory) < 20 {
		return 0.0
	}

	currentVolume := data.VolumeHistory[len(data.VolumeHistory)-1]
	avgVolume := s.calculateAverageFloat(data.VolumeHistory[len(data.VolumeHistory)-20:])

	if avgVolume == 0 {
		return 0.0
	}

	volumeScore := (currentVolume - avgVolume) / avgVolume

	if volumeScore > 2 {
		volumeScore = 2
	} else if volumeScore < -2 {
		volumeScore = -2
	}

	return volumeScore / 2 // Normalize to -1 to 1
}

// calculateVolatility calculates price volatility
func (s *AlgorithmicStrategy) calculateVolatility(data *MarketData, period int) float64 {
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

	if avgPrice == 0 {
		return 0.0
	}

	return volatility / avgPrice
}

// calculateZScore calculates z-score for mean reversion
func (s *AlgorithmicStrategy) calculateZScore(data *MarketData, period int) float64 {
	if len(data.PriceHistory) < period {
		return 0.0
	}

	prices := data.PriceHistory[len(data.PriceHistory)-period:]
	avgPrice := s.calculateAverageFloat(prices)

	sumSquares := 0.0
	for _, price := range prices {
		sumSquares += math.Pow(price-avgPrice, 2)
	}

	stdDev := math.Sqrt(sumSquares / float64(len(prices)))

	if stdDev == 0 {
		return 0.0
	}

	currentPrice := data.CurrentPrice
	return (currentPrice - avgPrice) / stdDev
}

// calculateAverageFloat calculates simple moving average for float64
func (s *AlgorithmicStrategy) calculateAverageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}
