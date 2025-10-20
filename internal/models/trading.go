package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SandboxTrade represents a simulated trade for SOLACE learning
type SandboxTrade struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"-"`

	SessionID uuid.UUID `gorm:"type:uuid;not null" json:"session_id"`

	// Trade Details
	TradingPair  string   `gorm:"size:50;not null" json:"trading_pair"` // BTC/USDC, ETH/USDC
	Direction    string   `gorm:"size:4;not null" json:"direction"`     // BUY, SELL
	Size         float64  `gorm:"type:decimal(18,8);not null" json:"size"`
	EntryPrice   float64  `gorm:"type:decimal(18,8);not null" json:"entry_price"`
	ExitPrice    *float64 `gorm:"type:decimal(18,8)" json:"exit_price,omitempty"`
	StrategyName *string  `gorm:"size:50" json:"strategy_name,omitempty"` // Multi-strategy support

	// Financial Results
	ProfitLoss        *float64 `gorm:"type:decimal(18,8)" json:"profit_loss,omitempty"`
	ProfitLossPercent *float64 `gorm:"type:decimal(10,4)" json:"profit_loss_percent,omitempty"`
	Fees              float64  `gorm:"type:decimal(18,8);default:0" json:"fees"`

	// Trade Status
	Status   string     `gorm:"size:20;not null" json:"status"` // OPEN, CLOSED, CANCELLED
	OpenedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"opened_at"`
	ClosedAt *time.Time `json:"closed_at,omitempty"`

	// SOLACE Learning Data
	Reasoning        string   `gorm:"type:text;not null" json:"reasoning"`
	MarketConditions JSONB    `gorm:"type:jsonb" json:"market_conditions"`
	SentimentScore   *float64 `gorm:"type:decimal(5,4)" json:"sentiment_score,omitempty"`  // -1.0 to 1.0
	ConfidenceScore  *float64 `gorm:"type:decimal(5,4)" json:"confidence_score,omitempty"` // 0.0 to 1.0

	// Benchmark & Performance
	BenchmarkScore *float64 `gorm:"type:decimal(10,4)" json:"benchmark_score,omitempty"`
	SharpeRatio    *float64 `gorm:"type:decimal(10,4)" json:"sharpe_ratio,omitempty"`
	SortinoRatio   *float64 `gorm:"type:decimal(10,4)" json:"sortino_ratio,omitempty"`

	// Market Regime
	MarketRegime     *string  `gorm:"size:20" json:"market_regime,omitempty"` // BULL, BEAR, CHOP, VOLATILITY_SPIKE, UNKNOWN
	RegimeConfidence *float64 `gorm:"type:decimal(5,4)" json:"regime_confidence,omitempty"`

	// Audit Trail
	TradeHash      string  `gorm:"size:64;not null;unique" json:"trade_hash"`
	LineageTrail   JSONB   `gorm:"type:jsonb" json:"lineage_trail"`
	SolaceOverride bool    `gorm:"default:false" json:"solace_override"`
	OverrideReason *string `gorm:"type:text" json:"override_reason,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TradingPerformance stores aggregated metrics for SOLACE's trading performance
type TradingPerformance struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	SessionID uuid.UUID `gorm:"type:uuid;not null" json:"session_id"`

	CalculatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"calculated_at"`

	// Trade Statistics
	TotalTrades   int      `gorm:"default:0" json:"total_trades"`
	WinningTrades int      `gorm:"default:0" json:"winning_trades"`
	LosingTrades  int      `gorm:"default:0" json:"losing_trades"`
	WinRate       *float64 `gorm:"type:decimal(5,2)" json:"win_rate,omitempty"` // Percentage

	// Financial Metrics
	TotalProfitLoss *float64 `gorm:"type:decimal(18,8)" json:"total_profit_loss,omitempty"`
	AvgProfit       *float64 `gorm:"type:decimal(18,8)" json:"avg_profit,omitempty"`
	AvgLoss         *float64 `gorm:"type:decimal(18,8)" json:"avg_loss,omitempty"`
	LargestWin      *float64 `gorm:"type:decimal(18,8)" json:"largest_win,omitempty"`
	LargestLoss     *float64 `gorm:"type:decimal(18,8)" json:"largest_loss,omitempty"`

	// Risk Metrics
	SharpeRatio        *float64 `gorm:"type:decimal(10,4)" json:"sharpe_ratio,omitempty"`
	SortinoRatio       *float64 `gorm:"type:decimal(10,4)" json:"sortino_ratio,omitempty"`
	MaxDrawdown        *float64 `gorm:"type:decimal(10,4)" json:"max_drawdown,omitempty"`
	MaxDrawdownPercent *float64 `gorm:"type:decimal(5,2)" json:"max_drawdown_percent,omitempty"`
	CurrentDrawdown    *float64 `gorm:"type:decimal(10,4)" json:"current_drawdown,omitempty"`

	// Position Sizing
	AvgPositionSize *float64 `gorm:"type:decimal(18,8)" json:"avg_position_size,omitempty"`
	MaxPositionSize *float64 `gorm:"type:decimal(18,8)" json:"max_position_size,omitempty"`
	KellyCriterion  *float64 `gorm:"type:decimal(5,4)" json:"kelly_criterion,omitempty"`

	// Risk Management
	Var5Percent *float64 `gorm:"type:decimal(18,8)" json:"var_5_percent,omitempty"` // Value at Risk
	RiskOfRuin  *float64 `gorm:"type:decimal(5,4)" json:"risk_of_ruin,omitempty"`   // Probability of ruin

	// Strategy Evolution
	StrategyVersion int `gorm:"default:1" json:"strategy_version"`
	MutationCount   int `gorm:"default:0" json:"mutation_count"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// MarketDataCache stores OHLCV data and technical indicators
type MarketDataCache struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Symbol      string `gorm:"size:20;not null" json:"symbol"`       // BTC, ETH, SOL
	TradingPair string `gorm:"size:50;not null" json:"trading_pair"` // BTC/USDC

	// OHLCV Data
	Timestamp time.Time `gorm:"not null" json:"timestamp"`
	Open      float64   `gorm:"type:decimal(18,8);not null" json:"open"`
	High      float64   `gorm:"type:decimal(18,8);not null" json:"high"`
	Low       float64   `gorm:"type:decimal(18,8);not null" json:"low"`
	Close     float64   `gorm:"type:decimal(18,8);not null" json:"close"`
	Volume    float64   `gorm:"type:decimal(18,8);not null" json:"volume"`

	// Technical Indicators
	SMA20          *float64 `gorm:"type:decimal(18,8)" json:"sma_20,omitempty"`
	SMA50          *float64 `gorm:"type:decimal(18,8)" json:"sma_50,omitempty"`
	SMA200         *float64 `gorm:"type:decimal(18,8)" json:"sma_200,omitempty"`
	RSI14          *float64 `gorm:"type:decimal(5,2)" json:"rsi_14,omitempty"`
	ATR14          *float64 `gorm:"type:decimal(18,8)" json:"atr_14,omitempty"`
	BollingerUpper *float64 `gorm:"type:decimal(18,8)" json:"bollinger_upper,omitempty"`
	BollingerLower *float64 `gorm:"type:decimal(18,8)" json:"bollinger_lower,omitempty"`

	// Market Regime
	Volatility    *float64 `gorm:"type:decimal(10,6)" json:"volatility,omitempty"`
	TrendStrength *float64 `gorm:"type:decimal(5,4)" json:"trend_strength,omitempty"`
	MarketRegime  *string  `gorm:"size:20" json:"market_regime,omitempty"`

	// Data Source
	Source string `gorm:"size:50;not null" json:"source"` // coingecko, jupiter, etc.

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// StrategyMutation tracks SOLACE's recursive learning and strategy evolution
type StrategyMutation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	SessionID uuid.UUID `gorm:"type:uuid;not null" json:"session_id"`

	// Strategy Identity
	StrategyVersion int    `gorm:"not null" json:"strategy_version"`
	StrategyName    string `gorm:"size:100;not null" json:"strategy_name"`
	StrategyHash    string `gorm:"size:64;not null;unique" json:"strategy_hash"`

	// Mutation Details
	ParentHash     *string `gorm:"size:64" json:"parent_hash,omitempty"`
	MutationType   string  `gorm:"size:50;not null" json:"mutation_type"` // PARAMETER_TUNE, RULE_ADD, etc.
	MutationDelta  JSONB   `gorm:"type:jsonb;not null" json:"mutation_delta"`
	MutationReason string  `gorm:"type:text;not null" json:"mutation_reason"`

	// Performance Before/After
	SharpeBefore  *float64 `gorm:"type:decimal(10,4)" json:"sharpe_before,omitempty"`
	SharpeAfter   *float64 `gorm:"type:decimal(10,4)" json:"sharpe_after,omitempty"`
	SortinoBefore *float64 `gorm:"type:decimal(10,4)" json:"sortino_before,omitempty"`
	SortinoAfter  *float64 `gorm:"type:decimal(10,4)" json:"sortino_after,omitempty"`
	WinRateBefore *float64 `gorm:"type:decimal(5,2)" json:"win_rate_before,omitempty"`
	WinRateAfter  *float64 `gorm:"type:decimal(5,2)" json:"win_rate_after,omitempty"`

	// Approval Status
	Status     string  `gorm:"size:20;not null" json:"status"`       // TESTING, APPROVED, REJECTED, DEPLOYED
	ApprovedBy *string `gorm:"size:50" json:"approved_by,omitempty"` // SOLACE, USER, BENCHMARK

	// Timestamps
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	DeployedAt *time.Time `json:"deployed_at,omitempty"`
}

// RiskEvent logs kill-switch activations and risk breaches
type RiskEvent struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"-"`

	EventType string `gorm:"size:50;not null" json:"event_type"` // DRAWDOWN_LIMIT, VAR_BREACH, KILL_SWITCH
	Severity  string `gorm:"size:20;not null" json:"severity"`   // INFO, WARNING, CRITICAL

	// Event Details
	Description    string   `gorm:"type:text;not null" json:"description"`
	TriggerValue   *float64 `gorm:"type:decimal(18,8)" json:"trigger_value,omitempty"`
	ThresholdValue *float64 `gorm:"type:decimal(18,8)" json:"threshold_value,omitempty"`

	// Actions Taken
	ActionTaken     string `gorm:"size:100;not null" json:"action_taken"` // CLOSE_ALL_POSITIONS, HALT_TRADING, etc.
	PositionsClosed int    `gorm:"default:0" json:"positions_closed"`

	// Response Time
	DetectedAt        time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"detected_at"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	ResponseLatencyMs *int       `json:"response_latency_ms,omitempty"` // Must be <250ms for kill-switch

	// Metadata
	SolaceDecision  bool `gorm:"default:true" json:"solace_decision"`
	OverrideAllowed bool `gorm:"default:false" json:"override_allowed"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName overrides for proper naming
func (SandboxTrade) TableName() string {
	return "sandbox_trades"
}

func (TradingPerformance) TableName() string {
	return "trading_performance"
}

func (MarketDataCache) TableName() string {
	return "market_data_cache"
}

func (StrategyMutation) TableName() string {
	return "strategy_mutations"
}

func (RiskEvent) TableName() string {
	return "risk_events"
}

// BeforeCreate hook to generate trade hash
func (st *SandboxTrade) BeforeCreate(tx *gorm.DB) error {
	if st.TradeHash == "" {
		// Generate simple hash from trade details
		st.TradeHash = generateTradeHash(st)
	}
	return nil
}

// Helper function to generate trade hash (placeholder)
func generateTradeHash(trade *SandboxTrade) string {
	// TODO: Implement proper SHA256 hash
	return uuid.New().String()[:32]
}

// ============================================================================
// MULTI-STRATEGY MODELS
// ============================================================================

// StrategyMetrics represents performance metrics for a single strategy
type StrategyMetrics struct {
	StrategyName      string    `json:"strategy_name"`
	TotalTrades       int       `json:"total_trades"`
	WinningTrades     int       `json:"winning_trades"`
	LosingTrades      int       `json:"losing_trades"`
	WinRate           float64   `json:"win_rate"`
	TotalProfitLoss   float64   `json:"total_profit_loss"`
	AverageProfitLoss float64   `json:"average_profit_loss"`
	SharpeRatio       float64   `json:"sharpe_ratio"`
	MaxDrawdown       float64   `json:"max_drawdown"`
	CurrentBalance    float64   `json:"current_balance"`
	LastUpdated       time.Time `json:"last_updated"`
	CanPromoteToLive  bool      `json:"can_promote_to_live"`
	MissingCriteria   []string  `json:"missing_criteria,omitempty"`
}

// MasterMetrics represents aggregated metrics across all strategies
type MasterMetrics struct {
	TotalStrategies  int       `json:"total_strategies"`
	ActiveStrategies int       `json:"active_strategies"`
	TotalSignals     int       `json:"total_signals"`
	BuySignals       int       `json:"buy_signals"`
	SellSignals      int       `json:"sell_signals"`
	HoldSignals      int       `json:"hold_signals"`
	TotalTrades      int       `json:"total_trades"`
	TotalProfitLoss  float64   `json:"total_profit_loss"`
	OverallWinRate   float64   `json:"overall_win_rate"`
	BestStrategy     string    `json:"best_strategy"`
	WorstStrategy    string    `json:"worst_strategy"`
	LastUpdated      time.Time `json:"last_updated"`
}
