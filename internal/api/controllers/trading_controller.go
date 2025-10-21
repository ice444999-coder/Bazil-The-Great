/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	service "ares_api/internal/interfaces/service"
	"ares_api/internal/services"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TradingController handles autonomous sandbox trading for SOLACE
type TradingController struct {
	TradingService *services.TradingService
	LedgerService  service.LedgerService
}

func NewTradingController(tradingService *services.TradingService, ledgerService service.LedgerService) *TradingController {
	return &TradingController{
		TradingService: tradingService,
		LedgerService:  ledgerService,
	}
}

// ExecuteTrade - SOLACE calls this to execute sandbox trades
// @Summary Execute Sandbox Trade
// @Description SOLACE executes a sandbox trade with reasoning
// @Tags Sandbox Trading
// @Accept json
// @Produce json
// @Param request body dto.ExecuteTradeRequest true "Trade Execution"
// @Success 200 {object} dto.SandboxTradeResponse
// @Security BearerAuth
// @Router /trading/execute [post]
func (c *TradingController) ExecuteTrade(ctx *gin.Context) {
	var req dto.ExecuteTradeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.JSON(ctx, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("userID")

	// Parse session ID
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		common.JSON(ctx, http.StatusBadRequest, gin.H{"error": "invalid session_id format"})
		return
	}

	// Execute trade
	trade, err := c.TradingService.ExecuteTrade(
		userID,
		sessionID,
		req.TradingPair,
		req.Direction,
		req.SizeUSD,
		req.Reasoning,
	)
	if err != nil {
		// LOG THE ACTUAL ERROR
		fmt.Printf("❌ TRADE EXECUTION FAILED: %v\n", err)
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log to ledger
	_ = c.LedgerService.Append(
		userID,
		"ExecuteTrade",
		"SOLACE executed "+req.Direction+" trade on "+req.TradingPair+" for $"+strconv.FormatFloat(req.SizeUSD, 'f', 2, 64),
	)

	// Return trade response
	response := dto.SandboxTradeResponse{
		ID:               trade.ID,
		TradingPair:      trade.TradingPair,
		Direction:        trade.Direction,
		Size:             trade.Size,
		EntryPrice:       trade.EntryPrice,
		Fees:             trade.Fees,
		Status:           trade.Status,
		OpenedAt:         trade.OpenedAt,
		Reasoning:        trade.Reasoning,
		MarketConditions: trade.MarketConditions,
		TradeHash:        trade.TradeHash,
	}

	common.JSON(ctx, http.StatusOK, response)
}

// ExecuteLeveragedTrade - Execute trade with leverage (1x-20x)
// @Summary Execute Leveraged Trade
// @Description SOLACE executes a leveraged trade with up to 20x leverage
// @Tags Sandbox Trading
// @Accept json
// @Produce json
// @Param request body dto.ExecuteLeveragedTradeRequest true "Leveraged Trade Execution"
// @Success 200 {object} dto.SandboxTradeResponse
// @Security BearerAuth
// @Router /trading/leverage [post]
func (c *TradingController) ExecuteLeveragedTrade(ctx *gin.Context) {
	var req dto.ExecuteLeveragedTradeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.JSON(ctx, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("userID")

	// Parse session ID
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		common.JSON(ctx, http.StatusBadRequest, gin.H{"error": "invalid session_id format"})
		return
	}

	// Validate leverage
	if req.Leverage < 1.0 || req.Leverage > 20.0 {
		common.JSON(ctx, http.StatusBadRequest, gin.H{"error": "leverage must be between 1x and 20x"})
		return
	}

	// Execute leveraged trade
	trade, err := c.TradingService.ExecuteLeveragedTrade(
		userID,
		sessionID,
		req.TradingPair,
		req.Direction,
		req.SizeUSD,
		req.Leverage,
		req.Reasoning,
	)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log to ledger
	_ = c.LedgerService.Append(
		userID,
		"ExecuteLeveragedTrade",
		fmt.Sprintf("SOLACE executed %.2fx leveraged %s trade on %s for $%.2f (collateral: $%.2f)",
			req.Leverage, req.Direction, req.TradingPair, req.SizeUSD, req.SizeUSD/req.Leverage),
	)

	// Return trade response
	response := dto.SandboxTradeResponse{
		ID:               trade.ID,
		TradingPair:      trade.TradingPair,
		Direction:        trade.Direction,
		Size:             trade.Size,
		EntryPrice:       trade.EntryPrice,
		Fees:             trade.Fees,
		Status:           trade.Status,
		OpenedAt:         trade.OpenedAt,
		Reasoning:        trade.Reasoning,
		MarketConditions: trade.MarketConditions,
		TradeHash:        trade.TradeHash,
	}

	common.JSON(ctx, http.StatusOK, response)
}

// CloseTrade - Close an open sandbox trade
// @Summary Close Sandbox Trade
// @Description Close an open trade and calculate P&L
// @Tags Sandbox Trading
// @Accept json
// @Produce json
// @Param request body dto.CloseTradeRequest true "Trade ID"
// @Success 200 {object} dto.SandboxTradeResponse
// @Security BearerAuth
// @Router /trading/close [post]
func (c *TradingController) CloseTrade(ctx *gin.Context) {
	var req dto.CloseTradeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.JSON(ctx, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("userID")

	// Close trade
	trade, err := c.TradingService.CloseTrade(userID, req.TradeID)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log to ledger
	pnl := "N/A"
	if trade.ProfitLoss != nil {
		pnl = strconv.FormatFloat(*trade.ProfitLoss, 'f', 2, 64)
	}
	_ = c.LedgerService.Append(
		userID,
		"CloseTrade",
		"SOLACE closed trade "+strconv.Itoa(int(req.TradeID))+" with P&L: $"+pnl,
	)

	// Return trade response
	response := dto.SandboxTradeResponse{
		ID:                trade.ID,
		TradingPair:       trade.TradingPair,
		Direction:         trade.Direction,
		Size:              trade.Size,
		EntryPrice:        trade.EntryPrice,
		ExitPrice:         trade.ExitPrice,
		ProfitLoss:        trade.ProfitLoss,
		ProfitLossPercent: trade.ProfitLossPercent,
		Fees:              trade.Fees,
		Status:            trade.Status,
		OpenedAt:          trade.OpenedAt,
		ClosedAt:          trade.ClosedAt,
		Reasoning:         trade.Reasoning,
		MarketConditions:  trade.MarketConditions,
		TradeHash:         trade.TradeHash,
	}

	common.JSON(ctx, http.StatusOK, response)
}

// CloseAllTrades - Kill-switch: close all open trades
// @Summary Close All Trades (Kill-Switch)
// @Description Emergency close all open positions
// @Tags Sandbox Trading
// @Produce json
// @Success 200 {object} dto.CloseAllTradesResponse
// @Security BearerAuth
// @Router /trading/close-all [post]
func (c *TradingController) CloseAllTrades(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	// Close all trades
	closed, err := c.TradingService.CloseAllTrades(userID)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log to ledger
	_ = c.LedgerService.Append(
		userID,
		"CloseAllTrades",
		"Kill-switch activated: closed "+strconv.Itoa(closed)+" trades",
	)

	response := dto.CloseAllTradesResponse{
		Message:      "Kill-switch executed successfully",
		TradesClosed: closed,
	}

	common.JSON(ctx, http.StatusOK, response)
}

// GetTradeHistory - Get paginated trade history
// @Summary Get Trade History
// @Description Get sandbox trade history with pagination
// @Tags Sandbox Trading
// @Produce json
// @Param limit query int false "Number of trades (default 50)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {array} dto.SandboxTradeResponse
// @Security BearerAuth
// @Router /trading/history [get]
func (c *TradingController) GetTradeHistory(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	// Parse pagination
	limitQuery := ctx.DefaultQuery("limit", "50")
	offsetQuery := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitQuery)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetQuery)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get trade history
	trades, err := c.TradingService.GetTradeHistory(userID, limit, offset)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response DTOs
	var responses []dto.SandboxTradeResponse
	for _, trade := range trades {
		responses = append(responses, dto.SandboxTradeResponse{
			ID:                trade.ID,
			TradingPair:       trade.TradingPair,
			Direction:         trade.Direction,
			Size:              trade.Size,
			EntryPrice:        trade.EntryPrice,
			ExitPrice:         trade.ExitPrice,
			ProfitLoss:        trade.ProfitLoss,
			ProfitLossPercent: trade.ProfitLossPercent,
			Fees:              trade.Fees,
			Status:            trade.Status,
			OpenedAt:          trade.OpenedAt,
			ClosedAt:          trade.ClosedAt,
			Reasoning:         trade.Reasoning,
			MarketConditions:  trade.MarketConditions,
			TradeHash:         trade.TradeHash,
		})
	}

	common.JSON(ctx, http.StatusOK, responses)
}

// GetOpenTrades - Get all open trades
// @Summary Get Open Trades
// @Description Get all currently open sandbox trades
// @Tags Sandbox Trading
// @Produce json
// @Success 200 {array} dto.SandboxTradeResponse
// @Security BearerAuth
// @Router /trading/open [get]
func (c *TradingController) GetOpenTrades(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	// Get open trades
	trades, err := c.TradingService.GetOpenTrades(userID)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response DTOs
	var responses []dto.SandboxTradeResponse
	for _, trade := range trades {
		responses = append(responses, dto.SandboxTradeResponse{
			ID:               trade.ID,
			TradingPair:      trade.TradingPair,
			Direction:        trade.Direction,
			Size:             trade.Size,
			EntryPrice:       trade.EntryPrice,
			ExitPrice:        trade.ExitPrice,
			ProfitLoss:       trade.ProfitLoss,
			Fees:             trade.Fees,
			Status:           trade.Status,
			OpenedAt:         trade.OpenedAt,
			ClosedAt:         trade.ClosedAt,
			Reasoning:        trade.Reasoning,
			MarketConditions: trade.MarketConditions,
			TradeHash:        trade.TradeHash,
		})
	}

	common.JSON(ctx, http.StatusOK, responses)
}

// GetPerformance - Get trading performance metrics
// @Summary Get Trading Performance
// @Description Get comprehensive trading performance metrics
// @Tags Sandbox Trading
// @Produce json
// @Success 200 {object} dto.TradingPerformanceResponse
// @Security BearerAuth
// @Router /trading/performance [get]
func (c *TradingController) GetPerformance(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	// Get performance metrics
	perf, err := c.TradingService.GetPerformance(userID)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response DTO
	response := dto.TradingPerformanceResponse{
		TotalTrades:     perf.TotalTrades,
		WinningTrades:   perf.WinningTrades,
		LosingTrades:    perf.LosingTrades,
		WinRate:         perf.WinRate,
		TotalProfitLoss: perf.TotalProfitLoss,
		AvgProfit:       perf.AvgProfit,
		AvgLoss:         perf.AvgLoss,
		LargestWin:      perf.LargestWin,
		LargestLoss:     perf.LargestLoss,
		SharpeRatio:     perf.SharpeRatio,
		SortinoRatio:    perf.SortinoRatio,
		KellyCriterion:  perf.KellyCriterion,
		Var5Percent:     perf.Var5Percent,
		RiskOfRuin:      perf.RiskOfRuin,
		StrategyVersion: perf.StrategyVersion,
		CalculatedAt:    perf.CalculatedAt,
	}

	common.JSON(ctx, http.StatusOK, response)
}

// ============================================================================
// MULTI-STRATEGY API ENDPOINTS
// ============================================================================

// GetStrategyList - Get all available trading strategies
// @Summary List Trading Strategies
// @Description Get list of all available trading strategies with descriptions
// @Tags Strategy
// @Produce json
// @Success 200 {object} dto.StrategyListResponse
// @Security BearerAuth
// @Router /api/v1/strategy/list [get]
func (c *TradingController) GetStrategyList(ctx *gin.Context) {
	// Return hardcoded strategy list for now
	strategyDTOs := []dto.StrategyInfo{
		{
			Name:        "RSI_Oversold",
			Description: "Buys when RSI indicates oversold conditions",
			RiskLevel:   "Medium",
		},
		{
			Name:        "MACD_Divergence",
			Description: "Trades on MACD signal line crossovers",
			RiskLevel:   "Medium",
		},
		{
			Name:        "Trend_Following",
			Description: "Follows EMA trends for momentum trading",
			RiskLevel:   "Low",
		},
		{
			Name:        "Support_Bounce",
			Description: "Buys at support levels, sells at resistance",
			RiskLevel:   "High",
		},
		{
			Name:        "Volume_Spike",
			Description: "Trades on unusual volume patterns",
			RiskLevel:   "High",
		},
	}

	response := dto.StrategyListResponse{
		Strategies: strategyDTOs,
		Total:      len(strategyDTOs),
	}

	common.JSON(ctx, http.StatusOK, response)
}

// GetStrategyMetrics - Get performance metrics for a specific strategy
// @Summary Get Strategy Metrics
// @Description Get detailed performance metrics for a specific strategy
// @Tags Strategy
// @Produce json
// @Param name path string true "Strategy Name" example:"RSI_Oversold"
// @Success 200 {object} dto.StrategyMetricsResponse
// @Security BearerAuth
// @Router /api/v1/strategy/{name}/metrics [get]
func (c *TradingController) GetStrategyMetrics(ctx *gin.Context) {
	strategyName := ctx.Param("name")
	userID := ctx.GetUint("userID")

	metrics, err := c.TradingService.GetStrategyMetrics(userID, strategyName)
	if err != nil {
		common.JSON(ctx, http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := dto.StrategyMetricsResponse{
		StrategyName:      metrics.StrategyName,
		TotalTrades:       metrics.TotalTrades,
		WinningTrades:     metrics.WinningTrades,
		LosingTrades:      metrics.LosingTrades,
		WinRate:           metrics.WinRate,
		TotalProfitLoss:   metrics.TotalProfitLoss,
		AverageProfitLoss: metrics.AverageProfitLoss,
		SharpeRatio:       metrics.SharpeRatio,
		MaxDrawdown:       metrics.MaxDrawdown,
		CurrentBalance:    metrics.CurrentBalance,
		LastUpdated:       metrics.LastUpdated,
		CanPromoteToLive:  metrics.CanPromoteToLive,
		MissingCriteria:   metrics.MissingCriteria,
	}

	common.JSON(ctx, http.StatusOK, response)
}

// GetStrategySandboxTrades - Get sandbox trades for a specific strategy
// @Summary Get Strategy Sandbox Trades
// @Description Get all sandbox trades executed by a specific strategy
// @Tags Strategy
// @Produce json
// @Param name path string true "Strategy Name" example:"RSI_Oversold"
// @Param limit query int false "Number of trades to return" default:50
// @Success 200 {object} dto.StrategySandboxTradesResponse
// @Security BearerAuth
// @Router /api/v1/strategy/{name}/sandbox-trades [get]
func (c *TradingController) GetStrategySandboxTrades(ctx *gin.Context) {
	strategyName := ctx.Param("name")
	userID := ctx.GetUint("userID")

	// Parse limit query parameter
	limitStr := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	trades, err := c.TradingService.GetStrategySandboxTrades(userID, strategyName, limit)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to DTOs
	tradeDTOs := make([]dto.SandboxTradeResponse, 0, len(trades))
	for _, trade := range trades {
		tradeDTOs = append(tradeDTOs, dto.SandboxTradeResponse{
			ID:                trade.ID,
			TradingPair:       trade.TradingPair,
			Direction:         trade.Direction,
			Size:              trade.Size,
			EntryPrice:        trade.EntryPrice,
			ExitPrice:         trade.ExitPrice,
			ProfitLoss:        trade.ProfitLoss,
			ProfitLossPercent: trade.ProfitLossPercent,
			Fees:              trade.Fees,
			Status:            trade.Status,
			OpenedAt:          trade.OpenedAt,
			ClosedAt:          trade.ClosedAt,
			Reasoning:         trade.Reasoning,
			MarketConditions:  trade.MarketConditions,
			TradeHash:         trade.TradeHash,
		})
	}

	response := dto.StrategySandboxTradesResponse{
		StrategyName: strategyName,
		Trades:       tradeDTOs,
		Total:        len(tradeDTOs),
	}

	common.JSON(ctx, http.StatusOK, response)
}

// ToggleStrategy - Enable/disable a specific strategy
// @Summary Toggle Strategy
// @Description Enable or disable a specific trading strategy
// @Tags Strategy
// @Accept json
// @Produce json
// @Param name path string true "Strategy Name" example:"RSI_Oversold"
// @Success 200 {object} dto.StrategyToggleResponse
// @Security BearerAuth
// @Router /api/v1/strategy/{name}/toggle [post]
func (c *TradingController) ToggleStrategy(ctx *gin.Context) {
	strategyName := ctx.Param("name")
	userID := ctx.GetUint("userID")

	newStatus, err := c.TradingService.ToggleStrategy(userID, strategyName)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log to ledger
	action := "enabled"
	if !newStatus {
		action = "disabled"
	}
	_ = c.LedgerService.Append(
		userID,
		"ToggleStrategy",
		fmt.Sprintf("Strategy '%s' %s", strategyName, action),
	)

	response := dto.StrategyToggleResponse{
		StrategyName: strategyName,
		Enabled:      newStatus,
		Message:      fmt.Sprintf("Strategy '%s' successfully %s", strategyName, action),
	}

	common.JSON(ctx, http.StatusOK, response)
}

// PromoteStrategyToLive - Promote strategy from sandbox to live trading
// @Summary Promote Strategy to Live
// @Description Promote a strategy from sandbox to live trading (requires minimum criteria)
// @Tags Strategy
// @Accept json
// @Produce json
// @Param name path string true "Strategy Name" example:"RSI_Oversold"
// @Success 200 {object} dto.StrategyPromoteResponse
// @Security BearerAuth
// @Router /api/v1/strategy/{name}/promote-to-live [post]
func (c *TradingController) PromoteStrategyToLive(ctx *gin.Context) {
	strategyName := ctx.Param("name")
	userID := ctx.GetUint("userID")

	// Check if strategy meets promotion criteria
	canPromote, missingCriteria, err := c.TradingService.CanPromoteStrategy(userID, strategyName)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !canPromote {
		common.JSON(ctx, http.StatusBadRequest, gin.H{
			"error":            "Strategy does not meet promotion criteria",
			"missing_criteria": missingCriteria,
		})
		return
	}

	// Promote strategy
	err = c.TradingService.PromoteStrategy(userID, strategyName)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log to ledger
	_ = c.LedgerService.Append(
		userID,
		"PromoteStrategy",
		fmt.Sprintf("Strategy '%s' promoted to LIVE trading", strategyName),
	)

	response := dto.StrategyPromoteResponse{
		StrategyName: strategyName,
		Status:       "LIVE",
		Message:      fmt.Sprintf("Strategy '%s' successfully promoted to live trading", strategyName),
	}

	common.JSON(ctx, http.StatusOK, response)
}

// GetMasterMetrics - Get aggregated metrics across all strategies
// @Summary Get Master Metrics
// @Description Get aggregated performance metrics across all trading strategies
// @Tags Strategy
// @Produce json
// @Success 200 {object} dto.MasterMetricsResponse
// @Security BearerAuth
// @Router /api/v1/strategy/master-metrics [get]
func (c *TradingController) GetMasterMetrics(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	metrics, err := c.TradingService.GetMasterMetrics(userID)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.MasterMetricsResponse{
		TotalStrategies:  metrics.TotalStrategies,
		ActiveStrategies: metrics.ActiveStrategies,
		TotalSignals:     metrics.TotalSignals,
		BuySignals:       metrics.BuySignals,
		SellSignals:      metrics.SellSignals,
		HoldSignals:      metrics.HoldSignals,
		TotalTrades:      metrics.TotalTrades,
		TotalProfitLoss:  metrics.TotalProfitLoss,
		OverallWinRate:   metrics.OverallWinRate,
		BestStrategy:     metrics.BestStrategy,
		WorstStrategy:    metrics.WorstStrategy,
		LastUpdated:      metrics.LastUpdated,
	}

	common.JSON(ctx, http.StatusOK, response)
}
