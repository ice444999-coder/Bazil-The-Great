package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/interfaces/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BalanceController struct {
	Service service.BalanceService
	LedgerService service.LedgerService
}

func NewBalanceController(s service.BalanceService , l service.LedgerService ) *BalanceController {
	return &BalanceController{Service: s , LedgerService: l}
}

// GetUSDBalance godoc
// @Summary      Get user USD balance
// @Description  Fetch the current USD balance for the authenticated user
// @Tags         balance
// @Produce      json
// @Success      200  {object}  dto.BalanceDTO
// @Security BearerAuth
// @Failure      500  {object}  map[string]string
// @Router       /balances [get]
func (c *BalanceController) GetUSDBalance(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	balance, err := c.Service.GetUSDBalance(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
    _ = c.LedgerService.Append(userID.(uint),  "GetUSDBalance", "Fetched USD balance")
	ctx.JSON(http.StatusOK, balance)
}

// InitializeBalance godoc
// @Summary      Initialize balance
// @Description  Create a default balance of 10k USD for the authenticated user
// @Tags         balance
// @Produce      json
// @Success      201  {object}  dto.BalanceDTO
// @Security BearerAuth
// @Failure      500  {object}  map[string]string
// @Router       /balances/init [post]
func (c *BalanceController) InitializeBalance(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	balance, err := c.Service.InitializeBalance(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.LedgerService.Append(userID.(uint),  "InitializeBalance", "Initialized USD balance")
	ctx.JSON(http.StatusCreated, balance)
}

// ResetBalance godoc
// @Summary      Reset balance
// @Description  Reset the user's balance back to default (10k USD)
// @Tags         balance
// @Produce      json
// @Success      200  {object}  dto.BalanceDTO
// @Security BearerAuth
// @Failure      500  {object}  map[string]string
// @Router       /balances/reset [post]
func (c *BalanceController) ResetBalance(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	balance, err := c.Service.ResetUSDBalance(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.LedgerService.Append(userID.(uint),  "ResetBalance", "Reset USD balance to default")
	ctx.JSON(http.StatusOK, balance)
}

// UpdateBalance godoc
// @Summary      Update balance
// @Description  Add or subtract USD from user's balance (used internally for trades)
// @Tags         balance
// @Accept       json
// @Produce      json
// @Param        delta    body      dto.BalanceDTO    true  "Balance delta (amount field used)"
// @Success      200  {object}  dto.BalanceDTO
// @Security BearerAuth
// @Failure      500  {object}  map[string]string
// @Router       /balances/update [put]
func (c *BalanceController) UpdateBalance(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req dto.BalanceDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	balance, err := c.Service.UpdateUSDBalance(userID.(uint), req.Amount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.LedgerService.Append(userID.(uint),  "UpdateBalance", "Updated USD balance by " +  fmt.Sprintf("%.2f", req.Amount))
	ctx.JSON(http.StatusOK, balance)
}
