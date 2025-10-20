package tools

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PermissionGate - Permission system with rate limiting + circuit breakers
type PermissionGate struct {
	DB *gorm.DB
}

// PermissionCheckResult - Result of permission check
type PermissionCheckResult struct {
	Allowed        bool
	DenialReason   string
	RemainingCalls int
	RemainingCost  float64
}

// CheckPermission - Called before EVERY tool execution
func (pg *PermissionGate) CheckPermission(agentName, toolName string) (*PermissionCheckResult, error) {
	var perm struct {
		AccessGranted           bool
		PersistentApproval      bool
		DailyUsageLimit         int
		HourlyUsageLimit        int
		CurrentDailyUsage       int
		CurrentHourlyUsage      int
		DailyCostLimit          float64
		TotalCostUSD            float64
		CircuitBreakerActive    bool
		CircuitBreakerThreshold int
		ConsecutiveFailures     int
		LastUsageReset          time.Time
	}

	err := pg.DB.Raw(`
        SELECT 
            tp.access_granted, 
            tp.persistent_approval,
            tp.daily_usage_limit,
            tp.hourly_usage_limit,
            tp.current_daily_usage,
            tp.current_hourly_usage,
            tp.daily_cost_limit,
            tp.total_cost_usd,
            tp.circuit_breaker_active,
            tp.circuit_breaker_threshold,
            tp.consecutive_failures,
            tp.last_usage_reset
        FROM tool_permissions tp
        JOIN tool_registry tr ON tp.tool_id = tr.tool_id
        WHERE tp.agent_name = ? AND tr.tool_name = ?
    `, agentName, toolName).Scan(&perm).Error

	if err != nil {
		return &PermissionCheckResult{Allowed: false, DenialReason: "permission not found"}, err
	}

	// Check 1: Access granted?
	if !perm.AccessGranted {
		return &PermissionCheckResult{Allowed: false, DenialReason: "access not granted"}, nil
	}

	// Check 2: Circuit breaker active?
	if perm.CircuitBreakerActive {
		return &PermissionCheckResult{
			Allowed:      false,
			DenialReason: fmt.Sprintf("circuit breaker active after %d consecutive failures", perm.ConsecutiveFailures),
		}, nil
	}

	// Check 3: Reset counters if new day/hour
	now := time.Now()
	if now.Sub(perm.LastUsageReset) > 24*time.Hour {
		pg.DB.Exec(`
            UPDATE tool_permissions 
            SET current_daily_usage = 0, current_hourly_usage = 0, last_usage_reset = NOW()
            WHERE agent_name = ? AND tool_id = (SELECT tool_id FROM tool_registry WHERE tool_name = ?)
        `, agentName, toolName)
		perm.CurrentDailyUsage = 0
		perm.CurrentHourlyUsage = 0
	} else if now.Sub(perm.LastUsageReset) > 1*time.Hour {
		pg.DB.Exec(`
            UPDATE tool_permissions 
            SET current_hourly_usage = 0
            WHERE agent_name = ? AND tool_id = (SELECT tool_id FROM tool_registry WHERE tool_name = ?)
        `, agentName, toolName)
		perm.CurrentHourlyUsage = 0
	}

	// Check 4: Rate limits
	if perm.CurrentDailyUsage >= perm.DailyUsageLimit {
		return &PermissionCheckResult{
			Allowed:      false,
			DenialReason: fmt.Sprintf("daily limit exceeded (%d/%d calls)", perm.CurrentDailyUsage, perm.DailyUsageLimit),
		}, nil
	}

	if perm.CurrentHourlyUsage >= perm.HourlyUsageLimit {
		return &PermissionCheckResult{
			Allowed:      false,
			DenialReason: fmt.Sprintf("hourly limit exceeded (%d/%d calls)", perm.CurrentHourlyUsage, perm.HourlyUsageLimit),
		}, nil
	}

	// Check 5: Cost limits
	if perm.TotalCostUSD >= perm.DailyCostLimit {
		return &PermissionCheckResult{
			Allowed:       false,
			DenialReason:  fmt.Sprintf("daily cost limit exceeded ($%.2f/$%.2f)", perm.TotalCostUSD, perm.DailyCostLimit),
			RemainingCost: 0,
		}, nil
	}

	// All checks passed
	return &PermissionCheckResult{
		Allowed:        true,
		RemainingCalls: perm.DailyUsageLimit - perm.CurrentDailyUsage,
		RemainingCost:  perm.DailyCostLimit - perm.TotalCostUSD,
	}, nil
}

// LogExecution - Called after EVERY tool execution
func (pg *PermissionGate) LogExecution(agentName, toolName string, success bool, executionTimeMs int, costUSD float64, errorMsg string) error {
	tx := pg.DB.Begin()

	// Get tool_id and cost
	var toolID string
	var toolCost float64
	tx.Raw(`SELECT tool_id, api_cost_per_call FROM tool_registry WHERE tool_name = ?`, toolName).Scan(&toolID)

	if costUSD == 0.0 {
		costUSD = toolCost // Use default if not specified
	}

	// Update permission stats
	if success {
		tx.Exec(`
            UPDATE tool_permissions 
            SET current_daily_usage = current_daily_usage + 1,
                current_hourly_usage = current_hourly_usage + 1,
                request_count = request_count + 1,
                success_count = success_count + 1,
                consecutive_failures = 0,
                total_cost_usd = total_cost_usd + ?,
                last_used_at = NOW(),
                updated_at = NOW()
            WHERE agent_name = ? AND tool_id = ?
        `, costUSD, agentName, toolID)
	} else {
		// FAILURE: Increment circuit breaker counter
		tx.Exec(`
            UPDATE tool_permissions 
            SET current_daily_usage = current_daily_usage + 1,
                current_hourly_usage = current_hourly_usage + 1,
                request_count = request_count + 1,
                failure_count = failure_count + 1,
                consecutive_failures = consecutive_failures + 1,
                last_failure_reason = ?,
                last_used_at = NOW(),
                updated_at = NOW()
            WHERE agent_name = ? AND tool_id = ?
        `, errorMsg, agentName, toolID)

		// Check if circuit breaker should trigger
		var result struct {
			Failures  int
			Threshold int
		}
		tx.Raw(`
            SELECT consecutive_failures AS failures, circuit_breaker_threshold AS threshold
            FROM tool_permissions 
            WHERE agent_name = ? AND tool_id = ?
        `, agentName, toolID).Scan(&result)

		if result.Failures >= result.Threshold {
			tx.Exec(`
                UPDATE tool_permissions 
                SET circuit_breaker_active = TRUE, auto_disabled_at = NOW()
                WHERE agent_name = ? AND tool_id = ?
            `, agentName, toolID)

			// Alert SOLACE
			fmt.Printf("🚨 CIRCUIT BREAKER: %s disabled for %s after %d failures\n", toolName, agentName, result.Failures)
		}
	}

	// Log execution to audit table
	tx.Exec(`
        INSERT INTO tool_execution_log (tool_id, agent_name, success, execution_time_ms, cost_usd, error_message)
        VALUES (?, ?, ?, ?, ?, ?)
    `, toolID, agentName, success, executionTimeMs, costUSD, errorMsg)

	return tx.Commit().Error
}

// RequestPermission - Agent asks SOLACE for access
func (pg *PermissionGate) RequestPermission(agentName, toolName, reason string, context map[string]interface{}, persistentApproval bool) (string, error) {
	var toolID string
	err := pg.DB.Raw(`SELECT tool_id FROM tool_registry WHERE tool_name = ?`, toolName).Scan(&toolID).Error
	if err != nil {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	var requestID string
	err = pg.DB.Raw(`
        INSERT INTO tool_permission_requests (tool_id, requesting_agent, request_reason, request_context, status, persistent_approval_requested)
        VALUES (?, ?, ?, ?, 'pending', ?)
        RETURNING request_id
    `, toolID, agentName, reason, context, persistentApproval).Scan(&requestID).Error

	if err != nil {
		return "", err
	}

	// Notify SOLACE
	fmt.Printf("📬 Permission request from %s for tool '%s': %s\n", agentName, toolName, requestID)

	return requestID, nil
}

// ApproveRequest - SOLACE approves permission request
func (pg *PermissionGate) ApproveRequest(requestID string) error {
	tx := pg.DB.Begin()

	// Get request details
	var req struct {
		ToolID                      string
		RequestingAgent             string
		PersistentApprovalRequested bool
	}
	tx.Raw(`
        SELECT tool_id, requesting_agent, persistent_approval_requested
        FROM tool_permission_requests 
        WHERE request_id = ?
    `, requestID).Scan(&req)

	// Update request status
	tx.Exec(`
        UPDATE tool_permission_requests 
        SET status = 'approved', reviewed_by = 'SOLACE', reviewed_at = NOW()
        WHERE request_id = ?
    `, requestID)

	// Grant permission
	tx.Exec(`
        INSERT INTO tool_permissions (
            tool_id, agent_name, access_granted, persistent_approval, approved_by, approved_at
        )
        VALUES (?, ?, TRUE, ?, 'SOLACE', NOW())
        ON CONFLICT (tool_id, agent_name) DO UPDATE 
        SET access_granted = TRUE, 
            persistent_approval = EXCLUDED.persistent_approval, 
            approved_by = 'SOLACE', 
            approved_at = NOW(),
            circuit_breaker_active = FALSE,
            consecutive_failures = 0
    `, req.ToolID, req.RequestingAgent, req.PersistentApprovalRequested)

	return tx.Commit().Error
}

// DenyRequest - SOLACE denies permission request
func (pg *PermissionGate) DenyRequest(requestID, reason string) error {
	return pg.DB.Exec(`
        UPDATE tool_permission_requests 
        SET status = 'denied', reviewed_by = 'SOLACE', reviewed_at = NOW(), denial_reason = ?
        WHERE request_id = ?
    `, reason, requestID).Error
}

// ResetCircuitBreaker - SOLACE manually resets circuit breaker
func (pg *PermissionGate) ResetCircuitBreaker(agentName, toolName string) error {
	return pg.DB.Exec(`
        UPDATE tool_permissions 
        SET circuit_breaker_active = FALSE, 
            consecutive_failures = 0,
            auto_disabled_at = NULL,
            updated_at = NOW()
        WHERE agent_name = ? 
        AND tool_id = (SELECT tool_id FROM tool_registry WHERE tool_name = ?)
    `, agentName, toolName).Error
}
