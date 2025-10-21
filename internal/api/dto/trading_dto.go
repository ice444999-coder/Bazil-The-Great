/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package dto

import (
	"time"
)

// ExecuteTradeRequest - SOLACE sends this to execute a sandbox trade
type ExecuteTradeRequest struct {
	SessionID   string  `json:"session_id" binding:"required"`   // UUID of conversation session
	TradingPair string  `json:"trading_pair" binding:"required"` // BTC/USDC, ETH/USDC, SOL/USDC
	Direction   string  `json:"direction" binding:"required"`    // BUY or SELL
	SizeUSD     float64 `json:"size_usd" binding:"required"`     // Amount in USD
	Reasoning   string  `json:"reasoning" binding:"required"`    // Why SOLACE made this trade
}

// ExecuteLeveragedTradeRequest - Execute trade with leverage (1x-20x)
type ExecuteLeveragedTradeRequest struct {
	SessionID   string  `json:"session_id" binding:"required"`   // UUID of conversation session
	TradingPair string  `json:"trading_pair" binding:"required"` // BTC/USDC, ETH/USDC, SOL/USDC
	Direction   string  `json:"direction" binding:"required"`    // BUY or SELL
	SizeUSD     float64 `json:"size_usd" binding:"required"`     // Full position size (not collateral)
	Leverage    float64 `json:"leverage" binding:"required"`     // 1.0 to 20.0
	Reasoning   string  `json:"reasoning" binding:"required"`    // Why SOLACE is using leverage
}

// CloseTradeRequest - Request to close a specific trade
type CloseTradeRequest struct {
	TradeID uint `json:"trade_id" binding:"required"`
}

// CloseAllTradesResponse - Response from kill-switch
type CloseAllTradesResponse struct {
	Message      string `json:"message"`
	TradesClosed int    `json:"trades_closed"`
}

// SandboxTradeResponse - Response for trade operations
type SandboxTradeResponse struct {
	ID                uint                   `json:"id"`
	TradingPair       string                 `json:"trading_pair"`
	Direction         string                 `json:"direction"`
	Size              float64                `json:"size"`
	EntryPrice        float64                `json:"entry_price"`
	ExitPrice         *float64               `json:"exit_price,omitempty"`
	ProfitLoss        *float64               `json:"profit_loss,omitempty"`
	ProfitLossPercent *float64               `json:"profit_loss_percent,omitempty"`
	Fees              float64                `json:"fees"`
	Status            string                 `json:"status"`
	OpenedAt          time.Time              `json:"opened_at"`
	ClosedAt          *time.Time             `json:"closed_at,omitempty"`
	Reasoning         string                 `json:"reasoning"`
	MarketConditions  map[string]interface{} `json:"market_conditions,omitempty" swaggertype:"object"`
	TradeHash         string                 `json:"trade_hash"`
}

// TradingPerformanceResponse - Performance metrics
type TradingPerformanceResponse struct {
	TotalTrades     int       `json:"total_trades"`
	WinningTrades   int       `json:"winning_trades"`
	LosingTrades    int       `json:"losing_trades"`
	WinRate         *float64  `json:"win_rate,omitempty"`
	TotalProfitLoss *float64  `json:"total_profit_loss,omitempty"`
	AvgProfit       *float64  `json:"avg_profit,omitempty"`
	AvgLoss         *float64  `json:"avg_loss,omitempty"`
	LargestWin      *float64  `json:"largest_win,omitempty"`
	LargestLoss     *float64  `json:"largest_loss,omitempty"`
	SharpeRatio     *float64  `json:"sharpe_ratio,omitempty"`
	SortinoRatio    *float64  `json:"sortino_ratio,omitempty"`
	KellyCriterion  *float64  `json:"kelly_criterion,omitempty"`
	Var5Percent     *float64  `json:"var_5_percent,omitempty"`
	RiskOfRuin      *float64  `json:"risk_of_ruin,omitempty"`
	StrategyVersion int       `json:"strategy_version"`
	CalculatedAt    time.Time `json:"calculated_at"`
}

// ============================================================================
// MULTI-STRATEGY DTOs
// ============================================================================

// StrategyInfo - Basic strategy information
type StrategyInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RiskLevel   string `json:"risk_level"`
}

// StrategyListResponse - Response for GET /api/v1/strategy/list
type StrategyListResponse struct {
	Strategies []StrategyInfo `json:"strategies"`
	Total      int            `json:"total"`
}

// StrategyMetricsResponse - Response for GET /api/v1/strategy/:name/metrics
type StrategyMetricsResponse struct {
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

// StrategySandboxTradesResponse - Response for GET /api/v1/strategy/:name/sandbox-trades
type StrategySandboxTradesResponse struct {
	StrategyName string                 `json:"strategy_name"`
	Trades       []SandboxTradeResponse `json:"trades"`
	Total        int                    `json:"total"`
}

// StrategyToggleResponse - Response for POST /api/v1/strategy/:name/toggle
type StrategyToggleResponse struct {
	StrategyName string `json:"strategy_name"`
	Enabled      bool   `json:"enabled"`
	Message      string `json:"message"`
}

// StrategyPromoteResponse - Response for POST /api/v1/strategy/:name/promote-to-live
type StrategyPromoteResponse struct {
	StrategyName string    `json:"strategy_name"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	PromotedAt   time.Time `json:"promoted_at"`
}

// MasterMetricsResponse - Response for GET /api/v1/strategy/master-metrics
type MasterMetricsResponse struct {
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
