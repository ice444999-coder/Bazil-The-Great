package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ares_api/internal/tools"
)

// ToolHandler handles tool registry and permission endpoints
type ToolHandler struct {
	PermissionGate *tools.PermissionGate
	VectorSearch   *tools.ToolVectorSearch
}

// SearchTools - GET /api/v1/tools/search?intent=I+want+to+trade&limit=5
func (h *ToolHandler) SearchTools(c *gin.Context) {
	intent := c.Query("intent")
	if intent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "intent query parameter required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 5
	}

	minSimilarityStr := c.DefaultQuery("min_similarity", "0.7")
	minSimilarity, err := strconv.ParseFloat(minSimilarityStr, 64)
	if err != nil {
		minSimilarity = 0.7
	}

	results, err := h.VectorSearch.SearchToolsByIntent(intent, minSimilarity, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"intent":         intent,
		"min_similarity": minSimilarity,
		"results":        results,
		"count":          len(results),
	})
}

// GetRegistry - GET /api/v1/tools/registry?category=trading
func (h *ToolHandler) GetRegistry(c *gin.Context) {
	category := c.Query("category")

	var query string
	var args []interface{}

	if category != "" {
		query = "SELECT tool_id, tool_name, tool_category, description, required_params, risk_level, api_cost_per_call FROM tool_registry WHERE tool_category = $1 ORDER BY tool_name"
		args = []interface{}{category}
	} else {
		query = "SELECT tool_id, tool_name, tool_category, description, required_params, risk_level, api_cost_per_call FROM tool_registry ORDER BY tool_category, tool_name"
	}

	type ToolInfo struct {
		ToolID         string  `json:"tool_id"`
		ToolName       string  `json:"tool_name"`
		ToolCategory   string  `json:"tool_category"`
		Description    string  `json:"description"`
		RequiredParams string  `json:"required_params"`
		RiskLevel      string  `json:"risk_level"`
		APICost        float64 `json:"api_cost_per_call"`
	}

	var toolsResult []ToolInfo
	if err := h.PermissionGate.DB.Raw(query, args...).Scan(&toolsResult).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tools": toolsResult,
		"count": len(toolsResult),
	})
}

// GetPendingRequests - GET /api/v1/tools/pending-requests
func (h *ToolHandler) GetPendingRequests(c *gin.Context) {
	type PendingRequest struct {
		RequestID   string `json:"request_id"`
		AgentName   string `json:"agent_name"`
		ToolName    string `json:"tool_name"`
		Reason      string `json:"reason"`
		ContextData string `json:"context_data"`
		RequestedAt string `json:"requested_at"`
	}

	var requests []PendingRequest
	query := `
		SELECT 
			tpr.request_id, tpr.agent_name, tr.tool_name, 
			tpr.reason_for_request, tpr.context_data, tpr.requested_at
		FROM tool_permission_requests tpr
		JOIN tool_registry tr ON tpr.tool_id = tr.tool_id
		WHERE tpr.status = 'pending'
		ORDER BY tpr.requested_at DESC
		LIMIT 100
	`

	if err := h.PermissionGate.DB.Raw(query).Scan(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_requests": requests,
		"count":            len(requests),
	})
}

// ApproveRequest - POST /api/v1/tools/approve/:request_id
func (h *ToolHandler) ApproveRequest(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id required"})
		return
	}

	requestUUID, err := uuid.Parse(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request_id format"})
		return
	}

	if err := h.PermissionGate.ApproveRequest(requestUUID.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Permission granted",
		"request_id": requestID,
		"approved":   true,
	})
}

// DenyRequest - POST /api/v1/tools/deny/:request_id
func (h *ToolHandler) DenyRequest(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id required"})
		return
	}

	requestUUID, err := uuid.Parse(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request_id format"})
		return
	}

	type DenyBody struct {
		Reason string `json:"reason"`
	}
	var body DenyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		body.Reason = "Denied by SOLACE"
	}

	if err := h.PermissionGate.DenyRequest(requestUUID.String(), body.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Permission denied",
		"request_id": requestID,
		"denied":     true,
		"reason":     body.Reason,
	})
}

// GetAgentPermissions - GET /api/v1/tools/permissions/:agent
func (h *ToolHandler) GetAgentPermissions(c *gin.Context) {
	agentName := c.Param("agent")
	if agentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent name required"})
		return
	}

	type PermissionInfo struct {
		ToolName             string  `json:"tool_name"`
		AccessGranted        bool    `json:"access_granted"`
		PersistentApproval   bool    `json:"persistent_approval"`
		DailyUsageLimit      int     `json:"daily_usage_limit"`
		CurrentDailyUsage    int     `json:"current_daily_usage"`
		DailyCostLimit       float64 `json:"daily_cost_limit"`
		TotalCost            float64 `json:"total_cost"`
		CircuitBreakerActive bool    `json:"circuit_breaker_active"`
	}

	var permissions []PermissionInfo
	query := `
		SELECT 
			tr.tool_name, tp.access_granted, tp.persistent_approval,
			tp.daily_usage_limit, tp.current_daily_usage, 
			tp.daily_cost_limit, tp.total_cost_usd,
			tp.circuit_breaker_active
		FROM tool_permissions tp
		JOIN tool_registry tr ON tp.tool_id = tr.tool_id
		WHERE tp.agent_name = $1
		ORDER BY tr.tool_name
	`

	if err := h.PermissionGate.DB.Raw(query, agentName).Scan(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agent_name":  agentName,
		"permissions": permissions,
		"count":       len(permissions),
	})
}

// RequestPermission - POST /api/v1/tools/request-permission
func (h *ToolHandler) RequestPermission(c *gin.Context) {
	type RequestBody struct {
		AgentName  string                 `json:"agent_name" binding:"required"`
		ToolName   string                 `json:"tool_name" binding:"required"`
		Reason     string                 `json:"reason" binding:"required"`
		Context    map[string]interface{} `json:"context"`
		Persistent bool                   `json:"persistent"`
	}

	var body RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestID, err := h.PermissionGate.RequestPermission(
		body.AgentName,
		body.ToolName,
		body.Reason,
		body.Context,
		body.Persistent,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Permission request submitted to SOLACE",
		"request_id": requestID,
		"agent_name": body.AgentName,
		"tool_name":  body.ToolName,
	})
}

// ResetCircuitBreaker - POST /api/v1/tools/reset-circuit-breaker
func (h *ToolHandler) ResetCircuitBreaker(c *gin.Context) {
	type ResetBody struct {
		AgentName string `json:"agent_name" binding:"required"`
		ToolName  string `json:"tool_name" binding:"required"`
	}

	var body ResetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.PermissionGate.ResetCircuitBreaker(body.AgentName, body.ToolName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Circuit breaker reset successfully",
		"agent_name": body.AgentName,
		"tool_name":  body.ToolName,
	})
}

// VerifyAccuracy - GET /api/v1/tools/verify-accuracy
func (h *ToolHandler) VerifyAccuracy(c *gin.Context) {
	if err := h.VectorSearch.VerifyMathematicalAccuracy(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "FAILED",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "PASSED",
		"message": "pgvector mathematical operations verified - cosine similarity working correctly",
	})
}

// GetUsageStats - GET /api/v1/tools/usage-stats/:agent
func (h *ToolHandler) GetUsageStats(c *gin.Context) {
	agentName := c.Param("agent")
	if agentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent name required"})
		return
	}

	type UsageStat struct {
		ToolName             string  `json:"tool_name"`
		TotalExecutions      int     `json:"total_executions"`
		SuccessfulExecutions int     `json:"successful_executions"`
		FailedExecutions     int     `json:"failed_executions"`
		AvgExecutionTimeMs   int     `json:"avg_execution_time_ms"`
		TotalCostUSD         float64 `json:"total_cost_usd"`
	}

	var stats []UsageStat
	query := `
		SELECT 
			tr.tool_name,
			COUNT(tel.log_id) as total_executions,
			SUM(CASE WHEN tel.success THEN 1 ELSE 0 END) as successful_executions,
			SUM(CASE WHEN NOT tel.success THEN 1 ELSE 0 END) as failed_executions,
			AVG(tel.execution_time_ms)::int as avg_execution_time_ms,
			SUM(tel.cost_usd) as total_cost_usd
		FROM tool_execution_log tel
		JOIN tool_registry tr ON tel.tool_id = tr.tool_id
		WHERE tel.agent_name = $1
		GROUP BY tr.tool_name
		ORDER BY total_executions DESC
	`

	if err := h.PermissionGate.DB.Raw(query, agentName).Scan(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agent_name": agentName,
		"stats":      stats,
		"count":      len(stats),
	})
}
