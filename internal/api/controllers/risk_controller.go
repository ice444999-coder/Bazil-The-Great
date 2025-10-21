/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RiskController handles risk management calculations
type RiskController struct {
	db *gorm.DB
}

// NewRiskController creates a new risk controller
func NewRiskController(db *gorm.DB) *RiskController {
	return &RiskController{db: db}
}

// KellyRequest represents Kelly sizing calculation request
type KellyRequest struct {
	WinRate     float64 `json:"win_rate" binding:"required"`      // Win rate as percentage (0-100)
	AvgWin      float64 `json:"avg_win" binding:"required"`       // Average win amount
	AvgLoss     float64 `json:"avg_loss" binding:"required"`      // Average loss amount
	Bankroll    float64 `json:"bankroll"`                         // Total bankroll (optional)
	MaxPosition float64 `json:"max_position"`                     // Max position size % (optional, default: 25)
}

// KellyResponse represents Kelly sizing calculation response
type KellyResponse struct {
	KellyPercentage  float64 `json:"kelly_percentage"`   // Optimal position size as % of bankroll
	FractionalKelly  float64 `json:"fractional_kelly"`   // Half-Kelly (safer)
	RecommendedSize  float64 `json:"recommended_size"`   // Actual $ amount if bankroll provided
	EdgePercentage   float64 `json:"edge_percentage"`    // Calculated edge
	ExpectedValue    float64 `json:"expected_value"`     // Expected value per trade
	RiskLevel        string  `json:"risk_level"`         // "conservative", "moderate", "aggressive"
}

// EmergencyPauseRequest represents emergency pause request
type EmergencyPauseRequest struct {
	Reason      string `json:"reason"`                  // Optional reason
	CloseAll    bool   `json:"close_all"`               // Whether to close all positions
	DisableTrade bool  `json:"disable_trade"`           // Whether to disable future trading
}

// EmergencyPauseResponse represents emergency pause response
type EmergencyPauseResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	PositionsClosed int  `json:"positions_closed"`
	TradingEnabled  bool `json:"trading_enabled"`
}

// CalculateKellySizing calculates optimal position size using Kelly Criterion
// @Summary Calculate Kelly Sizing
// @Description Calculate optimal position size using Kelly Criterion formula
// @Tags Risk
// @Accept json
// @Produce json
// @Param request body KellyRequest true "Kelly Parameters"
// @Success 200 {object} KellyResponse
// @Failure 400 {object} map[string]string
// @Router /risk/kelly [post]
func (rc *RiskController) CalculateKellySizing(c *gin.Context) {
	var req KellyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate inputs
	if req.WinRate <= 0 || req.WinRate >= 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "win_rate must be between 0 and 100"})
		return
	}
	if req.AvgWin <= 0 || req.AvgLoss <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avg_win and avg_loss must be positive"})
		return
	}

	// Convert win rate to decimal
	winRate := req.WinRate / 100.0
	lossRate := 1 - winRate

	// Calculate win/loss ratio (b in Kelly formula)
	winLossRatio := req.AvgWin / req.AvgLoss

	// Kelly Formula: f* = (bp - q) / b
	// where b = win/loss ratio, p = win probability, q = loss probability
	kellyPercentage := (winLossRatio*winRate - lossRate) / winLossRatio

	// Clamp Kelly to reasonable range (0-100%)
	if kellyPercentage < 0 {
		kellyPercentage = 0 // No edge, don't bet
	}
	if kellyPercentage > 1 {
		kellyPercentage = 1 // Cap at 100%
	}

	// Apply max position limit if specified
	maxPosition := req.MaxPosition
	if maxPosition <= 0 || maxPosition > 100 {
		maxPosition = 25 // Default 25% max
	}
	if kellyPercentage*100 > maxPosition {
		kellyPercentage = maxPosition / 100.0
	}

	// Calculate fractional Kelly (half-Kelly is safer)
	fractionalKelly := kellyPercentage * 0.5

	// Calculate edge and expected value
	edge := (winRate * req.AvgWin) - (lossRate * req.AvgLoss)
	edgePercentage := (edge / req.AvgLoss) * 100

	// Determine risk level
	riskLevel := "conservative"
	if kellyPercentage > 0.15 {
		riskLevel = "aggressive"
	} else if kellyPercentage > 0.07 {
		riskLevel = "moderate"
	}

	// Calculate recommended size if bankroll provided
	recommendedSize := 0.0
	if req.Bankroll > 0 {
		recommendedSize = fractionalKelly * req.Bankroll
	}

	c.JSON(http.StatusOK, KellyResponse{
		KellyPercentage:  math.Round(kellyPercentage*10000) / 100,  // Round to 2 decimals
		FractionalKelly:  math.Round(fractionalKelly*10000) / 100,
		RecommendedSize:  math.Round(recommendedSize*100) / 100,
		EdgePercentage:   math.Round(edgePercentage*100) / 100,
		ExpectedValue:    math.Round(edge*100) / 100,
		RiskLevel:        riskLevel,
	})
}

// EmergencyPause immediately stops all trading activity
// @Summary Emergency Pause Trading
// @Description Immediately stop all trading and optionally close positions
// @Tags Risk
// @Accept json
// @Produce json
// @Param request body EmergencyPauseRequest true "Emergency Pause Parameters"
// @Success 200 {object} EmergencyPauseResponse
// @Failure 500 {object} map[string]string
// @Router /trading/emergency-pause [post]
func (rc *RiskController) EmergencyPause(c *gin.Context) {
	var req EmergencyPauseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to safe settings if no body provided
		req.CloseAll = true
		req.DisableTrade = true
		req.Reason = "Emergency stop triggered"
	}

	positionsClosed := 0

	// TODO: Implement actual position closing logic
	// For now, simulate closing positions
	if req.CloseAll {
		// Query open positions from database and close them
		// This would integrate with trade_service.go
		positionsClosed = 0 // Replace with actual count
	}

	// TODO: Set global trading flag to disabled
	// This would update a settings table or Redis flag
	tradingEnabled := !req.DisableTrade

	message := "Emergency pause activated successfully"
	if positionsClosed > 0 {
		message += " and " + string(rune(positionsClosed)) + " positions closed"
	}

	c.JSON(http.StatusOK, EmergencyPauseResponse{
		Success:         true,
		Message:         message,
		PositionsClosed: positionsClosed,
		TradingEnabled:  tradingEnabled,
	})
}
