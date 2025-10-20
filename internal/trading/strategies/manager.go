package strategies

import (
	"fmt"
	"sync"
	"time"
)

// StrategyManager orchestrates multiple trading strategies with consensus voting
type StrategyManager struct {
	strategies         map[string]Strategy
	performance        map[string]*StrategyPerformance
	capitalAlloc       map[string]*CapitalAllocation
	mu                 sync.RWMutex
	consensusThreshold float64 // 0.6 = 60% of strategies must agree
	totalCapital       float64
}

// NewStrategyManager creates a new strategy manager instance
func NewStrategyManager(totalCapital float64) *StrategyManager {
	sm := &StrategyManager{
		strategies:         make(map[string]Strategy),
		performance:        make(map[string]*StrategyPerformance),
		capitalAlloc:       make(map[string]*CapitalAllocation),
		consensusThreshold: 0.6,
		totalCapital:       totalCapital,
	}

	// Register all strategies
	sm.RegisterStrategy(NewScalpingStrategy())
	sm.RegisterStrategy(NewWhaleTrackingStrategy())
	sm.RegisterStrategy(NewMomentumStrategy())
	// TODO: Add remaining 7 strategies (DayTrading, Breakout, News, Swing, Position, Algorithmic, PriceAction)

	// Initialize equal capital allocation
	sm.InitializeCapitalAllocation()

	return sm
}

// RegisterStrategy adds a new strategy to the manager
func (sm *StrategyManager) RegisterStrategy(strategy Strategy) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	config := strategy.GetConfig()
	name, ok := config["name"].(string)
	if !ok {
		return fmt.Errorf("strategy missing name in config")
	}

	sm.strategies[name] = strategy
	sm.performance[name] = &StrategyPerformance{
		StrategyName: name,
		LastUpdated:  time.Now(),
	}

	return nil
}

// GetConsensusSignal generates signals from all strategies and votes for consensus
func (sm *StrategyManager) GetConsensusSignal(marketData *MarketData) (*TradeSignal, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Collect signals from all enabled strategies
	signals := make([]*TradeSignal, 0)
	votes := make(map[string]int) // "buy", "sell", "hold" -> vote count
	totalConfidence := make(map[string]float64)
	reasons := make(map[string][]string)

	for name, strategy := range sm.strategies {
		signal, err := strategy.Generate(marketData)
		if err != nil {
			continue // Skip failed strategies
		}

		if signal != nil {
			signals = append(signals, signal)
			votes[signal.Action]++
			totalConfidence[signal.Action] += signal.Confidence
			reasons[signal.Action] = append(reasons[signal.Action],
				fmt.Sprintf("%s (%.0f%%)", name, signal.Confidence*100))
		}
	}

	if len(signals) == 0 {
		return nil, nil // No signals generated
	}

	// Find action with most votes
	maxVotes := 0
	consensusAction := "hold"
	for action, count := range votes {
		if count > maxVotes {
			maxVotes = count
			consensusAction = action
		}
	}

	// Check if consensus threshold is met
	votingStrategies := len(signals)
	consensusPercent := float64(maxVotes) / float64(votingStrategies)

	if consensusPercent < sm.consensusThreshold {
		// No consensus reached
		return &TradeSignal{
			Action:     "hold",
			Symbol:     marketData.Symbol,
			Confidence: 0.3,
			Reasoning: fmt.Sprintf("No consensus: %d/%d strategies agree (%.1f%% < %.1f%% threshold)",
				maxVotes, votingStrategies, consensusPercent*100, sm.consensusThreshold*100),
			Strategy:  "Consensus",
			Timestamp: time.Now(),
		}, nil
	}

	// Calculate average confidence for consensus action
	avgConfidence := totalConfidence[consensusAction] / float64(votes[consensusAction])

	// Build consensus signal
	reasonList := ""
	for _, reason := range reasons[consensusAction] {
		reasonList += fmt.Sprintf("- %s\n", reason)
	}

	return &TradeSignal{
		Action:     consensusAction,
		Symbol:     marketData.Symbol,
		Confidence: avgConfidence,
		Reasoning: fmt.Sprintf("Consensus %s from %d/%d strategies (%.1f%%):\n%s",
			consensusAction, maxVotes, votingStrategies, consensusPercent*100, reasonList),
		Strategy:    "Consensus",
		Timestamp:   time.Now(),
		TargetGain:  0.05, // Use conservative target for consensus
		StopLoss:    -0.02,
		MaxHoldTime: 3600, // 1 hour default
		Priority:    6,
	}, nil
}

// AnalyzeAllStrategies runs analysis from all strategies
func (sm *StrategyManager) AnalyzeAllStrategies(marketData *MarketData) map[string]*StrategyAnalysis {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	analyses := make(map[string]*StrategyAnalysis)

	for name, strategy := range sm.strategies {
		analysis := strategy.Analyze(marketData)
		if analysis != nil {
			analyses[name] = analysis
		}
	}

	return analyses
}

// UpdateStrategyPerformance updates performance metrics for a strategy after trade completion
func (sm *StrategyManager) UpdateStrategyPerformance(strategyName string, profitLoss float64, success bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	perf, exists := sm.performance[strategyName]
	if !exists {
		return fmt.Errorf("strategy %s not found", strategyName)
	}

	perf.TotalTrades++
	perf.TotalProfitLoss += profitLoss

	if success {
		perf.WinningTrades++
		perf.AvgProfit = (perf.AvgProfit*float64(perf.WinningTrades-1) + profitLoss) / float64(perf.WinningTrades)
	} else {
		perf.LosingTrades++
		perf.AvgLoss = (perf.AvgLoss*float64(perf.LosingTrades-1) + profitLoss) / float64(perf.LosingTrades)
	}

	perf.WinRate = float64(perf.WinningTrades) / float64(perf.TotalTrades)

	// Calculate profit factor
	if perf.LosingTrades > 0 && perf.AvgLoss != 0 {
		grossProfit := perf.AvgProfit * float64(perf.WinningTrades)
		grossLoss := -perf.AvgLoss * float64(perf.LosingTrades) // Make loss positive
		if grossLoss > 0 {
			perf.ProfitFactor = grossProfit / grossLoss
		}
	}

	perf.LastUpdated = time.Now()

	return nil
}

// InitializeCapitalAllocation sets equal allocation for all strategies
func (sm *StrategyManager) InitializeCapitalAllocation() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	numStrategies := float64(len(sm.strategies))
	if numStrategies == 0 {
		return
	}

	equalPercent := 1.0 / numStrategies
	equalCapital := sm.totalCapital * equalPercent

	for name := range sm.strategies {
		sm.capitalAlloc[name] = &CapitalAllocation{
			StrategyName:      name,
			AllocatedCapital:  equalCapital,
			AllocationPercent: equalPercent,
			LastRebalance:     time.Now(),
			PerformanceScore:  0.5, // Neutral start
		}
	}
}

// RebalanceCapital reallocates capital based on performance (game-theoretic)
func (sm *StrategyManager) RebalanceCapital() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Calculate performance scores for all strategies
	totalScore := 0.0
	for name, perf := range sm.performance {
		score := sm.calculatePerformanceScore(perf)
		if alloc, exists := sm.capitalAlloc[name]; exists {
			alloc.PerformanceScore = score
			totalScore += score
		}
	}

	if totalScore == 0 {
		return fmt.Errorf("no performance data to rebalance")
	}

	// Allocate capital proportional to performance scores
	for _, alloc := range sm.capitalAlloc {
		alloc.AllocationPercent = alloc.PerformanceScore / totalScore
		alloc.AllocatedCapital = sm.totalCapital * alloc.AllocationPercent
		alloc.LastRebalance = time.Now()
	}

	return nil
}

// calculatePerformanceScore computes a unified score for capital allocation
func (sm *StrategyManager) calculatePerformanceScore(perf *StrategyPerformance) float64 {
	if perf.TotalTrades < 5 {
		return 0.5 // Neutral score for strategies with insufficient data
	}

	// Win rate component (0.0 - 0.4)
	winRateScore := perf.WinRate * 0.4

	// Profit factor component (0.0 - 0.3)
	profitFactorScore := 0.0
	if perf.ProfitFactor > 2.0 {
		profitFactorScore = 0.3
	} else if perf.ProfitFactor > 1.5 {
		profitFactorScore = 0.2
	} else if perf.ProfitFactor > 1.0 {
		profitFactorScore = 0.1
	}

	// Total P&L component (0.0 - 0.3)
	plScore := 0.0
	if perf.TotalProfitLoss > 1000 {
		plScore = 0.3
	} else if perf.TotalProfitLoss > 500 {
		plScore = 0.2
	} else if perf.TotalProfitLoss > 100 {
		plScore = 0.1
	}

	return winRateScore + profitFactorScore + plScore
}

// GetStrategyPerformance returns performance metrics for a specific strategy
func (sm *StrategyManager) GetStrategyPerformance(strategyName string) (*StrategyPerformance, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	perf, exists := sm.performance[strategyName]
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", strategyName)
	}

	return perf, nil
}

// GetAllPerformance returns performance for all strategies
func (sm *StrategyManager) GetAllPerformance() map[string]*StrategyPerformance {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return copy to prevent external modification
	result := make(map[string]*StrategyPerformance)
	for name, perf := range sm.performance {
		result[name] = perf
	}

	return result
}

// GetCapitalAllocation returns current capital allocation
func (sm *StrategyManager) GetCapitalAllocation() map[string]*CapitalAllocation {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]*CapitalAllocation)
	for name, alloc := range sm.capitalAlloc {
		result[name] = alloc
	}

	return result
}

// EnableStrategy enables a specific strategy
func (sm *StrategyManager) EnableStrategy(strategyName string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	strategy, exists := sm.strategies[strategyName]
	if !exists {
		return fmt.Errorf("strategy %s not found", strategyName)
	}

	return strategy.UpdateConfig(map[string]interface{}{"enabled": true})
}

// DisableStrategy disables a specific strategy
func (sm *StrategyManager) DisableStrategy(strategyName string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	strategy, exists := sm.strategies[strategyName]
	if !exists {
		return fmt.Errorf("strategy %s not found", strategyName)
	}

	return strategy.UpdateConfig(map[string]interface{}{"enabled": false})
}
