package trading

import (
	"fmt"
	"math"
)

// Strategy represents a trading strategy interface
type Strategy interface {
	Name() string
	Description() string
	Analyze(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error)
	GetRiskLevel() string
}

// TradeSignal represents a trading recommendation
type TradeSignal struct {
	Action      string  `json:"action"`       // "buy", "sell", "hold"
	Confidence  float64 `json:"confidence"`   // 0-100%
	Reasoning   string  `json:"reasoning"`
	TargetPrice float64 `json:"target_price"`
	StopLoss    float64 `json:"stop_loss"`
	Strategy    string  `json:"strategy"`
	Symbol      string  `json:"symbol"`
}

// ========== MOMENTUM STRATEGY ==========

// MomentumStrategy trades based on price momentum
type MomentumStrategy struct {
	LookbackPeriod int     // Number of trades to analyze
	Threshold      float64 // Minimum price change % to trigger
}

// NewMomentumStrategy creates a momentum-based strategy
func NewMomentumStrategy() *MomentumStrategy {
	return &MomentumStrategy{
		LookbackPeriod: 10,
		Threshold:      2.0, // 2% price movement
	}
}

func (s *MomentumStrategy) Name() string {
	return "Momentum"
}

func (s *MomentumStrategy) Description() string {
	return "Buys on upward momentum, sells on downward momentum. Follows strong price trends."
}

func (s *MomentumStrategy) GetRiskLevel() string {
	return "Medium"
}

func (s *MomentumStrategy) Analyze(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	currentPrice := marketData.GetPrice(symbol)
	if currentPrice == 0 {
		return nil, fmt.Errorf("invalid symbol: %s", symbol)
	}

	// Get recent price history from trades
	recentPrices := s.getRecentPrices(symbol, history)
	
	if len(recentPrices) < 3 {
		// Not enough data - hold
		return &TradeSignal{
			Action:      "hold",
			Confidence:  50.0,
			Reasoning:   "Insufficient price history for momentum analysis",
			TargetPrice: currentPrice,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Calculate momentum (price change over lookback period)
	oldPrice := recentPrices[0]
	momentum := ((currentPrice - oldPrice) / oldPrice) * 100

	// Determine signal
	if momentum > s.Threshold {
		// Strong upward momentum - BUY
		confidence := math.Min(momentum*10, 100) // Higher momentum = higher confidence
		return &TradeSignal{
			Action:      "buy",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Strong upward momentum: +%.2f%% over %d periods", momentum, len(recentPrices)),
			TargetPrice: currentPrice * 1.05, // 5% profit target
			StopLoss:    currentPrice * 0.97, // 3% stop loss
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	} else if momentum < -s.Threshold {
		// Strong downward momentum - SELL (or avoid buying)
		confidence := math.Min(math.Abs(momentum)*10, 100)
		return &TradeSignal{
			Action:      "sell",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Strong downward momentum: %.2f%% over %d periods", momentum, len(recentPrices)),
			TargetPrice: currentPrice * 0.95,
			StopLoss:    currentPrice * 1.03,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Weak momentum - HOLD
	return &TradeSignal{
		Action:      "hold",
		Confidence:  60.0,
		Reasoning:   fmt.Sprintf("Weak momentum: %.2f%% (threshold: ±%.2f%%)", momentum, s.Threshold),
		TargetPrice: currentPrice,
		Strategy:    s.Name(),
		Symbol:      symbol,
	}, nil
}

func (s *MomentumStrategy) getRecentPrices(symbol string, history []VirtualTrade) []float64 {
	prices := make([]float64, 0)
	count := 0
	
	// Get last N prices for this symbol
	for i := len(history) - 1; i >= 0 && count < s.LookbackPeriod; i-- {
		if history[i].Symbol == symbol && history[i].Status == "closed" {
			prices = append([]float64{history[i].Price}, prices...) // Prepend to maintain chronological order
			count++
		}
	}
	
	return prices
}

// ========== MEAN REVERSION STRATEGY ==========

// MeanReversionStrategy trades based on price returning to average
type MeanReversionStrategy struct {
	LookbackPeriod   int     // Number of trades to calculate average
	DeviationThreshold float64 // Standard deviations from mean to trigger
}

// NewMeanReversionStrategy creates a mean reversion strategy
func NewMeanReversionStrategy() *MeanReversionStrategy {
	return &MeanReversionStrategy{
		LookbackPeriod:   20,
		DeviationThreshold: 1.5, // 1.5 standard deviations
	}
}

func (s *MeanReversionStrategy) Name() string {
	return "MeanReversion"
}

func (s *MeanReversionStrategy) Description() string {
	return "Buys when price is below average, sells when above. Assumes prices revert to mean."
}

func (s *MeanReversionStrategy) GetRiskLevel() string {
	return "Medium-High"
}

func (s *MeanReversionStrategy) Analyze(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	currentPrice := marketData.GetPrice(symbol)
	if currentPrice == 0 {
		return nil, fmt.Errorf("invalid symbol: %s", symbol)
	}

	// Get recent prices
	recentPrices := s.getRecentPrices(symbol, history)
	
	if len(recentPrices) < 5 {
		return &TradeSignal{
			Action:      "hold",
			Confidence:  50.0,
			Reasoning:   "Insufficient data for mean reversion analysis",
			TargetPrice: currentPrice,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Calculate mean and standard deviation
	mean := calculateMean(recentPrices)
	stdDev := calculateStdDev(recentPrices, mean)

	if stdDev == 0 {
		return &TradeSignal{
			Action:      "hold",
			Confidence:  50.0,
			Reasoning:   "No price volatility detected",
			TargetPrice: currentPrice,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Calculate z-score (how many std devs from mean)
	zScore := (currentPrice - mean) / stdDev

	// Determine signal
	if zScore < -s.DeviationThreshold {
		// Price significantly below mean - BUY (expecting reversion up)
		confidence := math.Min(math.Abs(zScore)*30, 95)
		return &TradeSignal{
			Action:      "buy",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Price %.2f%% below mean (z-score: %.2f), expecting reversion", ((currentPrice-mean)/mean)*100, zScore),
			TargetPrice: mean, // Target is the mean
			StopLoss:    currentPrice * 0.95,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	} else if zScore > s.DeviationThreshold {
		// Price significantly above mean - SELL (expecting reversion down)
		confidence := math.Min(zScore*30, 95)
		return &TradeSignal{
			Action:      "sell",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Price %.2f%% above mean (z-score: %.2f), expecting reversion", ((currentPrice-mean)/mean)*100, zScore),
			TargetPrice: mean,
			StopLoss:    currentPrice * 1.05,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Price near mean - HOLD
	return &TradeSignal{
		Action:      "hold",
		Confidence:  70.0,
		Reasoning:   fmt.Sprintf("Price near mean (z-score: %.2f, threshold: ±%.2f)", zScore, s.DeviationThreshold),
		TargetPrice: currentPrice,
		Strategy:    s.Name(),
		Symbol:      symbol,
	}, nil
}

func (s *MeanReversionStrategy) getRecentPrices(symbol string, history []VirtualTrade) []float64 {
	prices := make([]float64, 0)
	count := 0
	
	for i := len(history) - 1; i >= 0 && count < s.LookbackPeriod; i-- {
		if history[i].Symbol == symbol && history[i].Status == "closed" {
			if history[i].ExitPrice != nil {
				prices = append([]float64{*history[i].ExitPrice}, prices...)
			} else {
				prices = append([]float64{history[i].Price}, prices...)
			}
			count++
		}
	}
	
	return prices
}

// ========== UTILITY FUNCTIONS ==========

func calculateMean(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, price := range prices {
		sum += price
	}
	return sum / float64(len(prices))
}

func calculateStdDev(prices []float64, mean float64) float64 {
	if len(prices) < 2 {
		return 0
	}
	
	variance := 0.0
	for _, price := range prices {
		variance += math.Pow(price-mean, 2)
	}
	variance /= float64(len(prices) - 1)
	
	return math.Sqrt(variance)
}

// ========== STRATEGY MANAGER ==========

// StrategyManager manages multiple trading strategies
type StrategyManager struct {
	Strategies []Strategy
}

// NewStrategyManager creates a strategy manager with default strategies
func NewStrategyManager() *StrategyManager {
	return &StrategyManager{
		Strategies: []Strategy{
			NewMomentumStrategy(),
			NewMeanReversionStrategy(),
		},
	}
}

// GetAllSignals returns signals from all strategies
func (sm *StrategyManager) GetAllSignals(symbol string, marketData *MockMarketData, history []VirtualTrade) ([]*TradeSignal, error) {
	signals := make([]*TradeSignal, 0)
	
	for _, strategy := range sm.Strategies {
		signal, err := strategy.Analyze(symbol, marketData, history)
		if err != nil {
			continue
		}
		signals = append(signals, signal)
	}
	
	return signals, nil
}

// GetConsensusSignal returns the majority signal from all strategies
func (sm *StrategyManager) GetConsensusSignal(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	signals, err := sm.GetAllSignals(symbol, marketData, history)
	if err != nil || len(signals) == 0 {
		return nil, fmt.Errorf("no signals available")
	}
	
	// Count votes
	votes := make(map[string]int)
	totalConfidence := make(map[string]float64)
	
	for _, signal := range signals {
		votes[signal.Action]++
		totalConfidence[signal.Action] += signal.Confidence
	}
	
	// Find majority
	maxVotes := 0
	consensusAction := "hold"
	
	for action, count := range votes {
		if count > maxVotes {
			maxVotes = count
			consensusAction = action
		}
	}
	
	// Build consensus signal
	avgConfidence := totalConfidence[consensusAction] / float64(votes[consensusAction])
	
	reasons := ""
	for _, signal := range signals {
		if signal.Action == consensusAction {
			reasons += fmt.Sprintf("- %s: %s\n", signal.Strategy, signal.Reasoning)
		}
	}
	
	return &TradeSignal{
		Action:     consensusAction,
		Confidence: avgConfidence,
		Reasoning:  fmt.Sprintf("Consensus from %d/%d strategies:\n%s", votes[consensusAction], len(signals), reasons),
		Strategy:   "Consensus",
		Symbol:     symbol,
	}, nil
}
