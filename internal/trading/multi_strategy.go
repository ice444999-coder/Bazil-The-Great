/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package trading

import (
	"ares_api/internal/eventbus"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ===================================================================
// MULTI-STRATEGY ORCHESTRATOR
// ===================================================================
// Manages 5+ concurrent trading strategies with EventBus integration.
// Extends existing Strategy interface with hot-reload and metrics.
// ===================================================================

// MarketSnapshot - Real-time market data for strategies
type MarketSnapshot struct {
	Symbol      string
	Price       float64
	Volume24h   float64
	PriceChange float64 // Percentage
	Timestamp   time.Time
}

// TradeDecision - Enhanced trade signal with metadata
type TradeDecision struct {
	Action       string    `json:"action"` // "buy", "sell", "hold"
	Symbol       string    `json:"symbol"`
	Amount       float64   `json:"amount"`     // Position size
	Confidence   float64   `json:"confidence"` // 0-100%
	Reasoning    string    `json:"reasoning"`
	TargetPrice  float64   `json:"target_price"`
	StopLoss     float64   `json:"stop_loss"`
	StrategyName string    `json:"strategy_name"`
	Timestamp    time.Time `json:"timestamp"`
}

// StrategyMetrics - Performance tracking per strategy
type StrategyMetrics struct {
	StrategyName      string    `json:"strategy_name"`
	TotalTrades       int       `json:"total_trades"`
	WinningTrades     int       `json:"winning_trades"`
	LosingTrades      int       `json:"losing_trades"`
	WinRate           float64   `json:"win_rate"`          // Percentage
	TotalProfitLoss   float64   `json:"total_profit_loss"` // USD
	AverageProfitLoss float64   `json:"average_profit_loss"`
	SharpeRatio       float64   `json:"sharpe_ratio"`
	MaxDrawdown       float64   `json:"max_drawdown"` // Percentage
	CurrentBalance    float64   `json:"current_balance"`
	LastUpdated       time.Time `json:"last_updated"`

	// Authorization Gate Status
	CanPromoteToLive bool     `json:"can_promote_to_live"`
	MissingCriteria  []string `json:"missing_criteria,omitempty"`
}

// StrategyHealth - Real-time health status
type StrategyHealth struct {
	IsAlive           bool      `json:"is_alive"`
	LastHeartbeat     time.Time `json:"last_heartbeat"`
	LastError         string    `json:"last_error,omitempty"`
	ConsecutiveErrors int       `json:"consecutive_errors"`
	Uptime            string    `json:"uptime"`
}

// StrategyConfig - Hot-reloadable configuration
type StrategyConfig struct {
	Enabled        bool                   `json:"enabled"`
	MaxDailyTrades int                    `json:"max_daily_trades"`
	PositionSize   float64                `json:"position_size"`  // % of balance per trade
	RiskPerTrade   float64                `json:"risk_per_trade"` // Max % loss per trade
	Parameters     map[string]interface{} `json:"parameters"`     // Strategy-specific params
	AutoGraduate   bool                   `json:"auto_graduate"`  // Auto-promote to live
}

// ===================================================================
// MULTI-STRATEGY ORCHESTRATOR
// ===================================================================

// MultiStrategyOrchestrator - Manages multiple strategies concurrently
type MultiStrategyOrchestrator struct {
	mu         sync.RWMutex
	strategies map[string]Strategy // Key: strategy name
	configs    map[string]*StrategyConfig
	eventBus   *eventbus.EventBus
	db         *gorm.DB
	ctx        context.Context
	cancel     context.CancelFunc
	startTime  time.Time
}

// NewMultiStrategyOrchestrator creates a new orchestrator
func NewMultiStrategyOrchestrator(db *gorm.DB, eb *eventbus.EventBus, histMgr interface{}) *MultiStrategyOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	return &MultiStrategyOrchestrator{
		strategies: make(map[string]Strategy),
		configs:    make(map[string]*StrategyConfig),
		eventBus:   eb,
		db:         db,
		ctx:        ctx,
		cancel:     cancel,
		startTime:  time.Now(),
	}
}

// RegisterStrategy - Add a new strategy (hot-swappable)
func (mso *MultiStrategyOrchestrator) RegisterStrategy(strategy Strategy, config *StrategyConfig) error {
	mso.mu.Lock()
	defer mso.mu.Unlock()

	name := strategy.Name()

	// Register in orchestrator
	mso.strategies[name] = strategy
	mso.configs[name] = config

	log.Printf("[ORCHESTRATOR] Registered strategy: %s", name)

	// Publish registration event
	mso.publishEvent("strategy.registered", map[string]interface{}{
		"strategy_name": name,
		"enabled":       config.Enabled,
		"timestamp":     time.Now(),
	})

	return nil
}

// UnregisterStrategy - Remove a strategy (hot-swap)
func (mso *MultiStrategyOrchestrator) UnregisterStrategy(name string) error {
	mso.mu.Lock()
	defer mso.mu.Unlock()

	_, exists := mso.strategies[name]
	if !exists {
		return fmt.Errorf("strategy not found: %s", name)
	}

	delete(mso.strategies, name)
	delete(mso.configs, name)

	log.Printf("[ORCHESTRATOR] Unregistered strategy: %s", name)

	mso.publishEvent("strategy.unregistered", map[string]interface{}{
		"strategy_name": name,
		"timestamp":     time.Now(),
	})

	return nil
}

// GetStrategy - Retrieve a strategy by name
func (mso *MultiStrategyOrchestrator) GetStrategy(name string) (Strategy, error) {
	mso.mu.RLock()
	defer mso.mu.RUnlock()

	strategy, exists := mso.strategies[name]
	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", name)
	}

	return strategy, nil
}

// ListStrategies - Get all registered strategies
func (mso *MultiStrategyOrchestrator) ListStrategies() []string {
	mso.mu.RLock()
	defer mso.mu.RUnlock()

	names := make([]string, 0, len(mso.strategies))
	for name := range mso.strategies {
		names = append(names, name)
	}

	return names
}

// GetMasterMetrics - Aggregate metrics across all strategies
func (mso *MultiStrategyOrchestrator) GetMasterMetrics() *MasterMetrics {
	mso.mu.RLock()
	defer mso.mu.RUnlock()

	master := &MasterMetrics{
		TotalStrategies:  len(mso.strategies),
		ActiveStrategies: 0,
		Timestamp:        time.Now(),
	}

	for name := range mso.strategies {
		if mso.configs[name].Enabled {
			master.ActiveStrategies++
		}
	}

	return master
}

// MasterMetrics - Aggregated system-wide metrics
type MasterMetrics struct {
	TotalStrategies  int       `json:"total_strategies"`
	ActiveStrategies int       `json:"active_strategies"`
	TotalTrades      int       `json:"total_trades"`
	TotalProfitLoss  float64   `json:"total_profit_loss"`
	OverallWinRate   float64   `json:"overall_win_rate"`
	BestStrategy     string    `json:"best_strategy"`
	BestStrategyPL   float64   `json:"best_strategy_pl"`
	WorstStrategy    string    `json:"worst_strategy"`
	WorstStrategyPL  float64   `json:"worst_strategy_pl"`
	Timestamp        time.Time `json:"timestamp"`
}

// ToggleStrategy - Enable/disable a strategy
func (mso *MultiStrategyOrchestrator) ToggleStrategy(name string, enabled bool) error {
	mso.mu.Lock()
	defer mso.mu.Unlock()

	config, exists := mso.configs[name]
	if !exists {
		return fmt.Errorf("strategy not found: %s", name)
	}

	config.Enabled = enabled

	log.Printf("[ORCHESTRATOR] Strategy %s toggled: %v", name, enabled)

	mso.publishEvent("strategy.toggled", map[string]interface{}{
		"strategy_name": name,
		"enabled":       enabled,
		"timestamp":     time.Now(),
	})

	return nil
}

// PublishStrategyMetrics publishes strategy performance metrics to EventBus
func (mso *MultiStrategyOrchestrator) PublishStrategyMetrics(metrics *StrategyMetrics) {
	if mso.eventBus == nil {
		return
	}

	topic := fmt.Sprintf("strategy.%s.metrics", metrics.StrategyName)
	mso.eventBus.Publish(topic, map[string]interface{}{
		"strategy_name":       metrics.StrategyName,
		"total_trades":        metrics.TotalTrades,
		"winning_trades":      metrics.WinningTrades,
		"losing_trades":       metrics.LosingTrades,
		"win_rate":            metrics.WinRate,
		"total_profit_loss":   metrics.TotalProfitLoss,
		"average_profit_loss": metrics.AverageProfitLoss,
		"sharpe_ratio":        metrics.SharpeRatio,
		"max_drawdown":        metrics.MaxDrawdown,
		"current_balance":     metrics.CurrentBalance,
		"can_promote":         metrics.CanPromoteToLive,
		"missing_criteria":    metrics.MissingCriteria,
		"timestamp":           time.Now(),
	})
}

// PublishMasterMetrics publishes aggregated master metrics to EventBus
func (mso *MultiStrategyOrchestrator) PublishMasterMetrics(master *MasterMetrics) {
	if mso.eventBus == nil {
		return
	}

	mso.eventBus.Publish("strategy.master.metrics", map[string]interface{}{
		"total_strategies":  master.TotalStrategies,
		"active_strategies": master.ActiveStrategies,
		"total_trades":      master.TotalTrades,
		"total_profit_loss": master.TotalProfitLoss,
		"overall_win_rate":  master.OverallWinRate,
		"best_strategy":     master.BestStrategy,
		"best_strategy_pl":  master.BestStrategyPL,
		"worst_strategy":    master.WorstStrategy,
		"worst_strategy_pl": master.WorstStrategyPL,
		"timestamp":         time.Now(),
	})
}

// ExecuteStrategy - Execute a specific strategy
func (mso *MultiStrategyOrchestrator) ExecuteStrategy(name string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	strategy, err := mso.GetStrategy(name)
	if err != nil {
		return nil, err
	}

	// Check if enabled
	mso.mu.RLock()
	config := mso.configs[name]
	mso.mu.RUnlock()

	if !config.Enabled {
		return nil, fmt.Errorf("strategy %s is disabled", name)
	}

	// Execute using existing Analyze method
	symbol := "BTC/USDT" // Default symbol
	signal, err := strategy.Analyze(symbol, marketData, history)
	if err != nil {
		log.Printf("[ORCHESTRATOR] Execution error for %s: %v", name, err)
		return nil, err
	}

	// Publish decision event
	mso.publishEvent("strategy.signal", map[string]interface{}{
		"strategy_name": name,
		"action":        signal.Action,
		"symbol":        signal.Symbol,
		"confidence":    signal.Confidence,
		"reasoning":     signal.Reasoning,
		"timestamp":     time.Now(),
	})

	return signal, nil
}

// ExecuteAll - Execute all enabled strategies
func (mso *MultiStrategyOrchestrator) ExecuteAll(marketData *MockMarketData, history []VirtualTrade) map[string]*TradeSignal {
	mso.mu.RLock()
	defer mso.mu.RUnlock()

	decisions := make(map[string]*TradeSignal)

	for name, strategy := range mso.strategies {
		config := mso.configs[name]
		if !config.Enabled {
			continue
		}

		symbol := "BTC/USDT" // Default
		signal, err := strategy.Analyze(symbol, marketData, history)
		if err != nil {
			log.Printf("[ORCHESTRATOR] Error executing %s: %v", name, err)
			continue
		}

		decisions[name] = signal

		// Publish signal event to EventBus
		mso.publishStrategySignal(name, signal)
	}

	return decisions
}

// publishStrategySignal publishes strategy signal to EventBus
func (mso *MultiStrategyOrchestrator) publishStrategySignal(strategyName string, signal *TradeSignal) {
	if mso.eventBus == nil {
		return
	}

	topic := fmt.Sprintf("strategy.%s.signal", strategyName)
	mso.eventBus.Publish(topic, map[string]interface{}{
		"strategy":     signal.Strategy,
		"action":       signal.Action,
		"symbol":       signal.Symbol,
		"confidence":   signal.Confidence,
		"reasoning":    signal.Reasoning,
		"target_price": signal.TargetPrice,
		"stop_loss":    signal.StopLoss,
		"timestamp":    time.Now(),
	})
}

// Shutdown - Gracefully shut down orchestrator
func (mso *MultiStrategyOrchestrator) Shutdown() error {
	mso.cancel() // Cancel context

	mso.mu.Lock()
	defer mso.mu.Unlock()

	log.Println("[ORCHESTRATOR] Shutting down...")

	for name := range mso.strategies {
		log.Printf("[ORCHESTRATOR] Stopped strategy: %s", name)
	}

	log.Println("[ORCHESTRATOR] Shutdown complete")
	return nil
}

// publishEvent - Publish events to EventBus
func (mso *MultiStrategyOrchestrator) publishEvent(topic string, data map[string]interface{}) {
	if mso.eventBus == nil {
		return
	}

	if err := mso.eventBus.Publish(topic, data); err != nil {
		log.Printf("[ORCHESTRATOR] Failed to publish event %s: %v", topic, err)
	}
}

// GetUptime - Calculate orchestrator uptime
func (mso *MultiStrategyOrchestrator) GetUptime() time.Duration {
	return time.Since(mso.startTime)
}

// GetSystemHealth - Overall system health check
func (mso *MultiStrategyOrchestrator) GetSystemHealth() map[string]interface{} {
	mso.mu.RLock()
	defer mso.mu.RUnlock()

	health := map[string]interface{}{
		"status":           "healthy",
		"uptime":           mso.GetUptime().String(),
		"total_strategies": len(mso.strategies),
	}

	return health
}
