/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// IndicatorsController handles technical indicator calculations
type IndicatorsController struct {
	db *gorm.DB
}

// NewIndicatorsController creates a new indicators controller
func NewIndicatorsController(db *gorm.DB) *IndicatorsController {
	return &IndicatorsController{db: db}
}

// RSIResponse represents RSI indicator response
type RSIResponse struct {
	Value     float64 `json:"value"`
	Signal    string  `json:"signal"`    // "oversold" (<30), "neutral" (30-70), "overbought" (>70)
	Timestamp string  `json:"timestamp"`
	Period    int     `json:"period"`
}

// MACDResponse represents MACD indicator response
type MACDResponse struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
	CrossType string  `json:"cross_type"` // "bullish_cross", "bearish_cross", "neutral"
	Timestamp string  `json:"timestamp"`
	Fast      int     `json:"fast"`
	Slow      int     `json:"slow"`
	SignalLen int     `json:"signal_len"`
}

// WhaleAlert represents a large trade alert
type WhaleAlert struct {
	Symbol    string    `json:"symbol"`
	Volume    float64   `json:"volume"`      // USD value
	Price     float64   `json:"price"`
	Side      string    `json:"side"`        // "buy" or "sell"
	Timestamp time.Time `json:"timestamp"`
}

// WhaleAlertsResponse represents whale alerts response
type WhaleAlertsResponse struct {
	Alerts    []WhaleAlert `json:"alerts"`
	Count     int          `json:"count"`
	Threshold float64      `json:"threshold"` // Minimum volume to trigger alert
}

// GetRSI calculates and returns RSI indicator
// @Summary Get RSI Indicator
// @Description Calculate RSI (Relative Strength Index) for current market
// @Tags Indicators
// @Produce json
// @Param period query int false "RSI Period (default: 8)"
// @Success 200 {object} RSIResponse
// @Router /indicators/rsi [get]
func (ic *IndicatorsController) GetRSI(c *gin.Context) {
	// Parse period parameter (default: 8 for fast momentum)
	periodStr := c.DefaultQuery("period", "8")
	period, err := strconv.Atoi(periodStr)
	if err != nil || period < 2 || period > 50 {
		period = 8
	}

	// TODO: Implement real RSI calculation using historical price data
	// For now, generate realistic mock data based on time
	rsiValue := generateMockRSI()

	signal := "neutral"
	if rsiValue < 30 {
		signal = "oversold"
	} else if rsiValue > 70 {
		signal = "overbought"
	}

	c.JSON(http.StatusOK, RSIResponse{
		Value:     rsiValue,
		Signal:    signal,
		Timestamp: time.Now().Format(time.RFC3339),
		Period:    period,
	})
}

// GetMACD calculates and returns MACD indicator
// @Summary Get MACD Indicator
// @Description Calculate MACD (Moving Average Convergence Divergence) for current market
// @Tags Indicators
// @Produce json
// @Param fast query int false "Fast period (default: 5)"
// @Param slow query int false "Slow period (default: 35)"
// @Param signal query int false "Signal period (default: 5)"
// @Success 200 {object} MACDResponse
// @Router /indicators/macd [get]
func (ic *IndicatorsController) GetMACD(c *gin.Context) {
	// Parse parameters (Grok specified 5-35-5)
	fastStr := c.DefaultQuery("fast", "5")
	slowStr := c.DefaultQuery("slow", "35")
	signalStr := c.DefaultQuery("signal", "5")

	fast, _ := strconv.Atoi(fastStr)
	slow, _ := strconv.Atoi(slowStr)
	signalLen, _ := strconv.Atoi(signalStr)

	// Validate parameters
	if fast < 2 || fast > 50 {
		fast = 5
	}
	if slow < 10 || slow > 100 {
		slow = 35
	}
	if signalLen < 2 || signalLen > 20 {
		signalLen = 5
	}

	// TODO: Implement real MACD calculation using historical price data
	// For now, generate realistic mock data
	macdValue, signalValue := generateMockMACD()
	histogram := macdValue - signalValue

	// Determine cross type
	crossType := "neutral"
	if histogram > 0 && math.Abs(histogram) > 0.1 {
		crossType = "bullish_cross"
	} else if histogram < 0 && math.Abs(histogram) > 0.1 {
		crossType = "bearish_cross"
	}

	c.JSON(http.StatusOK, MACDResponse{
		MACD:      macdValue,
		Signal:    signalValue,
		Histogram: histogram,
		CrossType: crossType,
		Timestamp: time.Now().Format(time.RFC3339),
		Fast:      fast,
		Slow:      slow,
		SignalLen: signalLen,
	})
}

// GetWhaleAlerts returns recent large trades (>$1M)
// @Summary Get Whale Alerts
// @Description Retrieve recent large trades that exceed whale alert threshold
// @Tags Indicators
// @Produce json
// @Param threshold query float64 false "Minimum trade volume in USD (default: 1000000)"
// @Param limit query int false "Maximum number of alerts (default: 10)"
// @Success 200 {object} WhaleAlertsResponse
// @Router /alerts/whale [get]
func (ic *IndicatorsController) GetWhaleAlerts(c *gin.Context) {
	// Parse parameters
	thresholdStr := c.DefaultQuery("threshold", "1000000")
	limitStr := c.DefaultQuery("limit", "10")

	threshold, _ := strconv.ParseFloat(thresholdStr, 64)
	limit, _ := strconv.Atoi(limitStr)

	if threshold < 100000 {
		threshold = 1000000
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	// TODO: Query real trade data from database
	// For now, generate mock whale alerts for recent 5 minutes
	alerts := generateMockWhaleAlerts(threshold, limit)

	c.JSON(http.StatusOK, WhaleAlertsResponse{
		Alerts:    alerts,
		Count:     len(alerts),
		Threshold: threshold,
	})
}

// Helper function to generate mock RSI (oscillates between 20-80)
func generateMockRSI() float64 {
	// Use current time to create semi-random but consistent value
	seed := float64(time.Now().Unix() % 3600)
	return 45 + 25*math.Sin(seed/600) // Oscillates between 20 and 70
}

// Helper function to generate mock MACD values
func generateMockMACD() (macd float64, signal float64) {
	seed := float64(time.Now().Unix() % 3600)
	macd = 0.5 * math.Sin(seed/300)         // Oscillates between -0.5 and 0.5
	signal = macd - 0.1*math.Cos(seed/450) // Signal lags slightly
	return
}

// Helper function to generate mock whale alerts
func generateMockWhaleAlerts(threshold float64, limit int) []WhaleAlert {
	alerts := []WhaleAlert{}
	now := time.Now()

	// Simulate 0-3 whale alerts in last 5 minutes
	numAlerts := int(time.Now().Unix()%4) // 0-3 alerts

	symbols := []string{"BTC/USDT", "ETH/USDT", "SOL/USDT"}
	sides := []string{"buy", "sell"}

	for i := 0; i < numAlerts && i < limit; i++ {
		volume := threshold + float64((time.Now().Unix()+int64(i*137))%500000)
		price := 50000 + float64((time.Now().Unix()+int64(i*271))%10000)

		alerts = append(alerts, WhaleAlert{
			Symbol:    symbols[i%len(symbols)],
			Volume:    volume,
			Price:     price,
			Side:      sides[i%len(sides)],
			Timestamp: now.Add(-time.Duration(i*45) * time.Second),
		})
	}

	return alerts
}
