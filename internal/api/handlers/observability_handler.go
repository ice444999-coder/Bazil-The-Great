/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package handlers

import (
	"ares_api/internal/observability"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ObservabilityHandler handles observability endpoints
type ObservabilityHandler struct {
	db     *gorm.DB
	logger *observability.Logger
}

// NewObservabilityHandler creates a new observability handler
func NewObservabilityHandler(db *gorm.DB, logger *observability.Logger) *ObservabilityHandler {
	return &ObservabilityHandler{
		db:     db,
		logger: logger,
	}
}

// GetLogs retrieves service logs with filtering
// GET /api/v1/observability/logs?service=X&level=Y&trace_id=Z&limit=N
func (h *ObservabilityHandler) GetLogs(c *gin.Context) {
	serviceName := c.Query("service")
	level := c.Query("level")
	traceIDStr := c.Query("trace_id")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 1000 {
		limit = 100
	}

	var traceID *uuid.UUID
	if traceIDStr != "" {
		parsed, err := uuid.Parse(traceIDStr)
		if err == nil {
			traceID = &parsed
		}
	}

	logs, err := h.logger.QueryLogs(serviceName, level, traceID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(logs),
		"logs":  logs,
	})
}

// GetMetrics retrieves service metrics
// GET /api/v1/observability/metrics?service=X&metric=Y&hours=Z
func (h *ObservabilityHandler) GetMetrics(c *gin.Context) {
	serviceName := c.Query("service")
	metricName := c.Query("metric")
	hoursStr := c.DefaultQuery("hours", "1")

	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours > 168 {
		hours = 1
	}

	var metrics []observability.ServiceMetric
	query := h.db.Model(&observability.ServiceMetric{}).
		Where("timestamp > NOW() - INTERVAL '? hours'", hours).
		Order("timestamp DESC").
		Limit(1000)

	if serviceName != "" {
		query = query.Where("service_name = ?", serviceName)
	}
	if metricName != "" {
		query = query.Where("metric_name = ?", metricName)
	}

	err = query.Find(&metrics).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(metrics),
		"metrics": metrics,
	})
}

// GetSystemHealth retrieves system health view
// GET /api/v1/observability/health
func (h *ObservabilityHandler) GetSystemHealth(c *gin.Context) {
	type HealthRow struct {
		ServiceName           string  `json:"service_name"`
		Status                string  `json:"status"`
		Version               string  `json:"version"`
		LastHeartbeat         string  `json:"last_heartbeat"`
		SecondsSinceHeartbeat float64 `json:"seconds_since_heartbeat"`
		ActiveTraces          int     `json:"active_traces"`
		ErrorCount1h          int     `json:"error_count_1h"`
	}

	var health []HealthRow
	err := h.db.Table("v_system_health").Find(&health).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp": "now",
		"services":  health,
	})
}

// GetServicePerformance retrieves performance metrics view
// GET /api/v1/observability/performance?service=X
func (h *ObservabilityHandler) GetServicePerformance(c *gin.Context) {
	serviceName := c.Query("service")

	type PerfRow struct {
		ServiceName   string  `json:"service_name"`
		OperationName string  `json:"operation_name"`
		CallCount     int     `json:"call_count"`
		AvgDurationMs float64 `json:"avg_duration_ms"`
		P50Ms         float64 `json:"p50_ms"`
		P95Ms         float64 `json:"p95_ms"`
		P99Ms         float64 `json:"p99_ms"`
		ErrorCount    int     `json:"error_count"`
	}

	query := h.db.Table("v_service_performance")
	if serviceName != "" {
		query = query.Where("service_name = ?", serviceName)
	}

	var perf []PerfRow
	err := query.Find(&perf).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"performance": perf,
	})
}

// GetTrace retrieves all logs/spans for a trace ID
// GET /api/v1/observability/trace/:trace_id
func (h *ObservabilityHandler) GetTrace(c *gin.Context) {
	traceIDStr := c.Param("trace_id")
	traceID, err := uuid.Parse(traceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trace ID"})
		return
	}

	// Get logs
	logs, _ := h.logger.QueryLogs("", "", &traceID, 1000)

	// Get spans
	type ServiceSpan struct {
		ID            int64  `json:"id"`
		TraceID       string `json:"trace_id"`
		SpanID        string `json:"span_id"`
		ParentSpanID  string `json:"parent_span_id,omitempty"`
		ServiceName   string `json:"service_name"`
		OperationName string `json:"operation_name"`
		StartTime     string `json:"start_time"`
		EndTime       string `json:"end_time,omitempty"`
		DurationMs    *int   `json:"duration_ms,omitempty"`
		Status        string `json:"status,omitempty"`
	}
	var spans []ServiceSpan
	h.db.Table("service_spans").Where("trace_id = ?", traceID).Order("start_time").Find(&spans)

	c.JSON(http.StatusOK, gin.H{
		"trace_id": traceID,
		"logs":     logs,
		"spans":    spans,
	})
}

// RegisterRoutes registers observability routes
func (h *ObservabilityHandler) RegisterRoutes(r *gin.RouterGroup) {
	obs := r.Group("/observability")
	{
		obs.GET("/logs", h.GetLogs)
		obs.GET("/metrics", h.GetMetrics)
		obs.GET("/health", h.GetSystemHealth)
		obs.GET("/performance", h.GetServicePerformance)
		obs.GET("/trace/:trace_id", h.GetTrace)
	}
}
