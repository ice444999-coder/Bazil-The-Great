package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	service "ares_api/internal/interfaces/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TradeController struct {
	Service service.TradeService
	LedgerService service.LedgerService
}

func NewTradeController(s service.TradeService , l service.LedgerService) *TradeController {
	return &TradeController{Service: s , LedgerService: l}
}

// @Summary Execute Market Order
// @Tags Trading
// @Accept json
// @Produce json
// @Param request body dto.MarketOrderRequest true "Market Order"
// @Success 200 {object} dto.TradeResponse
// @Security BearerAuth
// @Router /trades/market [post]
func (c *TradeController) MarketOrder(ctx *gin.Context) {
	var req dto.MarketOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.JSON(ctx, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("userID") // from JWT middleware

	res, err := c.Service.MarketOrder(userID, req)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.LedgerService.Append(userID,  "MarketOrder", "Executed market order for symbol: " + req.Symbol)
	common.JSON(ctx, http.StatusOK, res)
}

// @Summary Execute Limit Order
// @Tags Trading
// @Accept json
// @Produce json
// @Param request body dto.LimitOrderRequest true "Limit Order"
// @Success 200 {object} dto.TradeResponse
// @Security BearerAuth
// @Router /trades/limit [post]
func (c *TradeController) LimitOrder(ctx *gin.Context) {
	var req dto.LimitOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.JSON(ctx, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("userID") // from JWT middleware

	res, err := c.Service.LimitOrder(userID, req)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.LedgerService.Append(userID,  "LimitOrder", "Placed limit order for symbol: " + req.Symbol)
	common.JSON(ctx, http.StatusOK, res)
}

// @Summary Get last N trades for user
// @Tags Trading
// @Produce json
// @Param limit query int true "Number of trades"
// @Success 200 {array} dto.TradeResponse
// @Security BearerAuth
// @Router /trades/history [get]
func (c *TradeController) GetHistory(ctx *gin.Context) {
	limitQuery := ctx.Query("limit")
	if limitQuery == "" {
		limitQuery = "10" // default 10
	}
	limit, err := strconv.Atoi(limitQuery)
	if err != nil || limit <= 0 {
		common.JSON(ctx, http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	userID := ctx.GetUint("userID")

	res, err := c.Service.GetHistory(userID, limit)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.LedgerService.Append(userID,  "GetHistory", "Fetched last " + limitQuery + " trades")
	common.JSON(ctx, http.StatusOK, res)
}

// @Summary Get pending limit orders for user
// @Tags Trading
// @Produce json
// @Success 200 {array} dto.TradeResponse
// @Security BearerAuth
// @Router /trades/pending [get]
func (c *TradeController) GetPendingLimitOrders(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	res, err := c.Service.GetPendingLimitOrders(userID)
	if err != nil {
		common.JSON(ctx, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.LedgerService.Append(userID,  "GetPendingLimitOrders", "Fetched pending limit orders")
	common.JSON(ctx, http.StatusOK, res)
}