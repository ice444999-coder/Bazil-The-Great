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
