/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package handlers

import (
	"ares_api/internal/binance"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// HistoricalDataHandler handles historical data endpoints
type HistoricalDataHandler struct {
	dataManager *binance.HistoricalDataManager
}

// NewHistoricalDataHandler creates a new handler
func NewHistoricalDataHandler(db *sql.DB) *HistoricalDataHandler {
	return &HistoricalDataHandler{
		dataManager: binance.NewHistoricalDataManager(db),
	}
}

// RegisterRoutes registers historical data routes
func (h *HistoricalDataHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/historical/candles", h.GetCandles)
	router.GET("/historical/cache-stats", h.GetCacheStats)
	router.DELETE("/historical/cache/cleanup", h.CleanupCache)
	router.GET("/historical/test-connection", h.TestConnection)
	router.GET("/historical/price/:symbol", h.GetLatestPrice)
}

// GetCandles fetches historical candles with caching
// GET /api/v1/historical/candles?symbol=BTCUSDT&interval=1h&start=2024-01-01T00:00:00Z&end=2024-01-31T23:59:59Z
func (h *HistoricalDataHandler) GetCandles(c *gin.Context) {
	// Parse query parameters
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter required"})
		return
	}

	intervalStr := c.Query("interval")
	if intervalStr == "" {
		intervalStr = "1h" // Default to 1 hour
	}

	interval := binance.KlineInterval(intervalStr)

	// Validate interval
	validIntervals := map[string]bool{
		"1m": true, "5m": true, "15m": true,
		"1h": true, "4h": true, "1d": true,
	}
	if !validIntervals[intervalStr] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid interval, must be one of: 1m, 5m, 15m, 1h, 4h, 1d",
		})
		return
	}

	// Parse start time (default: 7 days ago)
	startStr := c.Query("start")
	var startTime time.Time
	if startStr == "" {
		startTime = time.Now().Add(-7 * 24 * time.Hour)
	} else {
		parsed, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start time format, use RFC3339"})
			return
		}
		startTime = parsed
	}

	// Parse end time (default: now)
	endStr := c.Query("end")
	var endTime time.Time
	if endStr == "" {
		endTime = time.Now()
	} else {
		parsed, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end time format, use RFC3339"})
			return
		}
		endTime = parsed
	}

	// Fetch candles (with caching)
	candles, err := h.dataManager.GetHistoricalCandles(symbol, interval, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch historical candles",
			"details": err.Error(),
		})
		return
	}

	// Return candles
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"symbol":   symbol,
			"interval": intervalStr,
			"start":    startTime.Format(time.RFC3339),
			"end":      endTime.Format(time.RFC3339),
			"count":    len(candles),
			"candles":  candles,
		},
	})
}

// GetCacheStats returns cache statistics
// GET /api/v1/historical/cache-stats?symbol=BTCUSDT&interval=1h
func (h *HistoricalDataHandler) GetCacheStats(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter required"})
		return
	}

	intervalStr := c.Query("interval")
	if intervalStr == "" {
		intervalStr = "1h"
	}

	interval := binance.KlineInterval(intervalStr)

	stats, err := h.dataManager.GetCacheStats(symbol, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get cache stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// CleanupCache removes old cached candles
// DELETE /api/v1/historical/cache/cleanup?days=30
func (h *HistoricalDataHandler) CleanupCache(c *gin.Context) {
	daysStr := c.Query("days")
	if daysStr == "" {
		daysStr = "30" // Default: 30 days
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days parameter"})
		return
	}

	deleted, err := h.dataManager.CleanupOldCandles(time.Duration(days) * 24 * time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup cache",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cache cleanup completed",
		"data": gin.H{
			"deleted_candles": deleted,
			"older_than_days": days,
		},
	})
}

// TestConnection tests Binance API connectivity
// GET /api/v1/historical/test-connection
func (h *HistoricalDataHandler) TestConnection(c *gin.Context) {
	err := h.dataManager.TestConnection()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Binance API connection failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Binance API connection successful",
	})
}

// GetLatestPrice fetches current price
// GET /api/v1/historical/price/:symbol
func (h *HistoricalDataHandler) GetLatestPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter required"})
		return
	}

	price, err := h.dataManager.GetLatestPrice(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch price",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"symbol": symbol,
			"price":  price,
			"time":   time.Now().Format(time.RFC3339),
		},
	})
}
