package controllers

import (
	"ares_api/internal/monitoring"
	"ares_api/config"
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
)

// MonitoringController handles health and metrics endpoints
type MonitoringController struct {
	metrics *monitoring.Metrics
	flags   *config.FeatureFlags
}

// NewMonitoringController creates a new monitoring controller
func NewMonitoringController(metrics *monitoring.Metrics, flags *config.FeatureFlags) *MonitoringController {
	return &MonitoringController{
		metrics: metrics,
		flags:   flags,
	}
}

// GetHealth returns system health status
// @Summary Get system health
// @Description Returns health status including circuit breaker state, error rate, and uptime
// @Tags Monitoring
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health [get]
func (mc *MonitoringController) GetHealth(c *gin.Context) {
	if !mc.flags.MonitoringEnabled {
		c.JSON(http.StatusOK, gin.H{
			"status": "monitoring_disabled",
		})
		return
	}

	health := mc.metrics.CheckHealth()
	snapshot := mc.metrics.GetSnapshot()
	
	// Calculate uptime
	uptime := time.Since(snapshot.StartTime)
	uptimeStr := formatDuration(uptime)
	
	// Build UI-friendly response
	response := gin.H{
		"status": health.Status,
		"llm": gin.H{
			"model":               "DeepSeek-R1 14B",
			"connected":           health.Checks["llm"].Status == "pass",
			"latency":             int(snapshot.LLMAvgLatencyMs),
			"requests_per_minute": 0, // TODO: Calculate actual RPM
		},
		"database": gin.H{
			"connected":          true, // TODO: Add actual DB health check
			"pgvector_installed": true, // TODO: Add actual pgvector check
			"avg_query_time":     5,    // TODO: Track actual query times
			"active_connections": snapshot.DBConnections,
		},
		"trading": gin.H{
			"enabled":        true, // TODO: Get from trading service
			"open_positions": snapshot.ActiveTrades,
			"daily_loss":     0.0,  // TODO: Calculate actual daily loss
			"max_loss_limit": 500.0,
		},
		"ace": gin.H{
			"total_rules":    0,      // TODO: Get from ACE framework
			"active_rules":   0,      // TODO: Get from ACE framework
			"avg_confidence": 0.0,    // TODO: Get from ACE framework
			"last_learning":  "Never", // TODO: Track last learning time
		},
		"uptime":            uptimeStr,
		"total_requests":    snapshot.TotalRequests,
		"avg_response_time": int(snapshot.AvgResponseTimeMs),
		"api_performance":   gin.H{}, // TODO: Add endpoint-specific metrics
	}
	
	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	c.JSON(statusCode, response)
}

// formatDuration converts duration to human-readable string
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		return strconv.Itoa(days) + "d " + strconv.Itoa(hours) + "h"
	}
	if hours > 0 {
		return strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m"
	}
	return strconv.Itoa(minutes) + "m"
}

// GetMetrics returns comprehensive system metrics
// @Summary Get system metrics
// @Description Returns detailed metrics including request counts, trading stats, LLM usage, and error logs
// @Tags Monitoring
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics [get]
func (mc *MonitoringController) GetMetrics(c *gin.Context) {
	if !mc.flags.MonitoringEnabled {
		c.JSON(http.StatusOK, gin.H{
			"status": "monitoring_disabled",
		})
		return
	}

	snapshot := mc.metrics.GetSnapshot()
	c.JSON(http.StatusOK, snapshot)
}

// GetLogs returns recent system logs
// @Summary Get recent logs
// @Description Returns recent system logs with optional limit
// @Tags Monitoring
// @Produce json
// @Param limit query int false "Number of logs to retrieve" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/logs [get]
func (mc *MonitoringController) GetLogs(c *gin.Context) {
	if !mc.flags.MonitoringEnabled {
		c.JSON(http.StatusOK, gin.H{
			"logs": []interface{}{},
		})
		return
	}

	// Get limit from query params
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
			limit = parsedLimit
		}
	}

	// Get error logs from metrics
	snapshot := mc.metrics.GetSnapshot()
	logs := snapshot.RecentErrors
	
	// Limit the number of logs returned
	if len(logs) > limit {
		logs = logs[len(logs)-limit:]
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"count": len(logs),
	})
}
