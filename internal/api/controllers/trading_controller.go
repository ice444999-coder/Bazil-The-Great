package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	"ares_api/internal/services"
	service "ares_api/internal/interfaces/service"
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
		ID:               trade.ID,
		TradingPair:      trade.TradingPair,
		Direction:        trade.Direction,
		Size:             trade.Size,
		EntryPrice:       trade.EntryPrice,
		ExitPrice:        trade.ExitPrice,
		ProfitLoss:       trade.ProfitLoss,
		ProfitLossPercent: trade.ProfitLossPercent,
		Fees:             trade.Fees,
		Status:           trade.Status,
		OpenedAt:         trade.OpenedAt,
		ClosedAt:         trade.ClosedAt,
		Reasoning:        trade.Reasoning,
		MarketConditions: trade.MarketConditions,
		TradeHash:        trade.TradeHash,
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
			ID:               trade.ID,
			TradingPair:      trade.TradingPair,
			Direction:        trade.Direction,
			Size:             trade.Size,
			EntryPrice:       trade.EntryPrice,
			ExitPrice:        trade.ExitPrice,
			ProfitLoss:       trade.ProfitLoss,
			ProfitLossPercent: trade.ProfitLossPercent,
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
		TotalTrades:    perf.TotalTrades,
		WinningTrades:  perf.WinningTrades,
		LosingTrades:   perf.LosingTrades,
		WinRate:        perf.WinRate,
		TotalProfitLoss: perf.TotalProfitLoss,
		AvgProfit:      perf.AvgProfit,
		AvgLoss:        perf.AvgLoss,
		LargestWin:     perf.LargestWin,
		LargestLoss:    perf.LargestLoss,
		SharpeRatio:    perf.SharpeRatio,
		SortinoRatio:   perf.SortinoRatio,
		KellyCriterion: perf.KellyCriterion,
		Var5Percent:    perf.Var5Percent,
		RiskOfRuin:     perf.RiskOfRuin,
		StrategyVersion: perf.StrategyVersion,
		CalculatedAt:   perf.CalculatedAt,
	}

	common.JSON(ctx, http.StatusOK, response)
}
