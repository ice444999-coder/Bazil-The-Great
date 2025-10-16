package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// MemoryHandler handles Master Memory System API endpoints
type MemoryHandler struct {
	db *sql.DB
}

// NewMemoryHandler creates a new Memory handler
func NewMemoryHandler(db *sql.DB) *MemoryHandler {
	return &MemoryHandler{
		db: db,
	}
}

// ============================================================================
// MEMORY LOG ENDPOINTS
// ============================================================================

// MemoryLogEntry represents a single memory log entry
type MemoryLogEntry struct {
	ID              int        `json:"id"`
	Timestamp       time.Time  `json:"timestamp"`
	Source          string     `json:"source"`
	MessageType     string     `json:"message_type"`
	RawText         string     `json:"raw_text"`
	PhaseTag        string     `json:"phase_tag,omitempty"`
	CategoryTags    []string   `json:"category_tags,omitempty"`
	MentionedFiles  []string   `json:"mentioned_files,omitempty"`
	MentionedTasks  []int      `json:"mentioned_tasks,omitempty"`
	KeyConcepts     []string   `json:"key_concepts,omitempty"`
	ImportanceScore int        `json:"importance_score"`
	ReferencedCount int        `json:"referenced_count"`
	LastReferenced  *time.Time `json:"last_referenced,omitempty"`
	ContentHash     string     `json:"content_hash,omitempty"`
	HederaHash      string     `json:"hedera_hash,omitempty"`
	HederaTimestamp *time.Time `json:"hedera_timestamp,omitempty"`
}

// GetMemoryLogs returns paginated memory logs
// GET /api/v1/memory/logs?limit=50&offset=0&phase=Week5&source=GitHub%20Copilot%20Chat
func (h *MemoryHandler) GetMemoryLogs(c *gin.Context) {
	// Parse pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Parse filters
	phase := c.Query("phase")
	source := c.Query("source")
	minImportance, _ := strconv.Atoi(c.DefaultQuery("min_importance", "0"))

	// Build query
	query := `SELECT id, timestamp, source, message_type, raw_text, phase_tag, 
	          category_tags, mentioned_files, mentioned_tasks, key_concepts, 
	          importance_score, referenced_count, last_referenced, content_hash, 
	          hedera_hash, hedera_timestamp
	          FROM ares_memory_log WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	if phase != "" {
		query += fmt.Sprintf(" AND phase_tag = $%d", argCount)
		args = append(args, phase)
		argCount++
	}

	if source != "" {
		query += fmt.Sprintf(" AND source = $%d", argCount)
		args = append(args, source)
		argCount++
	}

	if minImportance > 0 {
		query += fmt.Sprintf(" AND importance_score >= $%d", argCount)
		args = append(args, minImportance)
		argCount++
	}

	query += " ORDER BY timestamp DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Query failed: %v", err)})
		return
	}
	defer rows.Close()

	logs := []MemoryLogEntry{}
	for rows.Next() {
		var log MemoryLogEntry
		err := rows.Scan(
			&log.ID, &log.Timestamp, &log.Source, &log.MessageType, &log.RawText,
			&log.PhaseTag, pq.Array(&log.CategoryTags), pq.Array(&log.MentionedFiles),
			pq.Array(&log.MentionedTasks), pq.Array(&log.KeyConcepts),
			&log.ImportanceScore, &log.ReferencedCount, &log.LastReferenced,
			&log.ContentHash, &log.HederaHash, &log.HederaTimestamp,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Scan failed: %v", err)})
			return
		}
		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   logs,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(logs),
		},
	})
}

// CreateMemoryLog creates a new memory log entry
// POST /api/v1/memory/log
// Body: { "source": "GitHub Copilot", "message_type": "Task", "raw_text": "...", ... }
func (h *MemoryHandler) CreateMemoryLog(c *gin.Context) {
	var req MemoryLogEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	query := `INSERT INTO ares_memory_log 
	          (source, message_type, raw_text, phase_tag, category_tags, 
	           mentioned_files, mentioned_tasks, key_concepts, importance_score)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	          RETURNING id, timestamp`

	err := h.db.QueryRow(
		query,
		req.Source, req.MessageType, req.RawText, req.PhaseTag,
		pq.Array(req.CategoryTags), pq.Array(req.MentionedFiles),
		pq.Array(req.MentionedTasks), pq.Array(req.KeyConcepts),
		req.ImportanceScore,
	).Scan(&req.ID, &req.Timestamp)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Insert failed: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   req,
	})
}

// ============================================================================
// MASTER PLAN ENDPOINTS
// ============================================================================

// MasterPlanTask represents a task in the master plan
type MasterPlanTask struct {
	ID                    int        `json:"id"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	TaskTitle             string     `json:"task_title"`
	TaskDescription       string     `json:"task_description"`
	Phase                 string     `json:"phase,omitempty"`
	Category              string     `json:"category,omitempty"`
	Priority              int        `json:"priority"`
	Status                string     `json:"status"`
	CompletionPercentage  int        `json:"completion_percentage"`
	DependsOn             []int      `json:"depends_on,omitempty"`
	Blocks                []int      `json:"blocks,omitempty"`
	RelatedFiles          []string   `json:"related_files,omitempty"`
	WhyThisMatters        string     `json:"why_this_matters,omitempty"`
	ConsciousnessImpact   int        `json:"consciousness_impact,omitempty"`
	EstimatedComplexity   int        `json:"estimated_complexity,omitempty"`
	SolaceCanAttempt      bool       `json:"solace_can_attempt"`
	RequiresDavidApproval bool       `json:"requires_david_approval"`
	AutonomyConstraints   string     `json:"autonomy_constraints,omitempty"`
	CreatedBy             string     `json:"created_by,omitempty"`
	ModifiedBy            string     `json:"modified_by,omitempty"`
	LastTouched           *time.Time `json:"last_touched,omitempty"`
}

// GetMasterPlan returns all tasks in the master plan
// GET /api/v1/masterplan?status=NEW&phase=Week5
func (h *MemoryHandler) GetMasterPlan(c *gin.Context) {
	status := c.Query("status")
	phase := c.Query("phase")

	query := `SELECT id, created_at, updated_at, task_title, task_description, 
	          phase, category, priority, status, completion_percentage, depends_on, 
	          blocks, related_files, why_this_matters, consciousness_impact, 
	          estimated_complexity, solace_can_attempt, requires_david_approval, 
	          autonomy_constraints, created_by, modified_by, last_touched
	          FROM ares_master_plan WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	if phase != "" {
		query += fmt.Sprintf(" AND phase = $%d", argCount)
		args = append(args, phase)
		argCount++
	}

	query += " ORDER BY priority DESC, consciousness_impact DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Query failed: %v", err)})
		return
	}
	defer rows.Close()

	tasks := []MasterPlanTask{}
	for rows.Next() {
		var task MasterPlanTask
		err := rows.Scan(
			&task.ID, &task.CreatedAt, &task.UpdatedAt, &task.TaskTitle, &task.TaskDescription,
			&task.Phase, &task.Category, &task.Priority, &task.Status, &task.CompletionPercentage,
			pq.Array(&task.DependsOn), pq.Array(&task.Blocks), pq.Array(&task.RelatedFiles),
			&task.WhyThisMatters, &task.ConsciousnessImpact, &task.EstimatedComplexity,
			&task.SolaceCanAttempt, &task.RequiresDavidApproval, &task.AutonomyConstraints,
			&task.CreatedBy, &task.ModifiedBy, &task.LastTouched,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Scan failed: %v", err)})
			return
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tasks,
		"count":  len(tasks),
	})
}

// CreateMasterPlanTask creates a new task
// POST /api/v1/masterplan/task
func (h *MemoryHandler) CreateMasterPlanTask(c *gin.Context) {
	var req MasterPlanTask
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	query := `INSERT INTO ares_master_plan 
	          (task_title, task_description, phase, category, priority, status, 
	           completion_percentage, depends_on, blocks, related_files, why_this_matters, 
	           consciousness_impact, estimated_complexity, solace_can_attempt, 
	           requires_david_approval, autonomy_constraints, created_by)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	          RETURNING id, created_at, updated_at`

	err := h.db.QueryRow(
		query,
		req.TaskTitle, req.TaskDescription, req.Phase, req.Category, req.Priority, req.Status,
		req.CompletionPercentage, pq.Array(req.DependsOn), pq.Array(req.Blocks),
		pq.Array(req.RelatedFiles), req.WhyThisMatters, req.ConsciousnessImpact,
		req.EstimatedComplexity, req.SolaceCanAttempt, req.RequiresDavidApproval,
		req.AutonomyConstraints, req.CreatedBy,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Insert failed: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   req,
	})
}

// UpdateMasterPlanTask updates an existing task
// PUT /api/v1/masterplan/task/:id
func (h *MemoryHandler) UpdateMasterPlanTask(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req MasterPlanTask
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	query := `UPDATE ares_master_plan SET 
	          updated_at = NOW(), task_title = $1, task_description = $2, phase = $3, 
	          category = $4, priority = $5, status = $6, completion_percentage = $7, 
	          depends_on = $8, blocks = $9, related_files = $10, why_this_matters = $11, 
	          consciousness_impact = $12, estimated_complexity = $13, solace_can_attempt = $14, 
	          requires_david_approval = $15, autonomy_constraints = $16, modified_by = $17
	          WHERE id = $18
	          RETURNING id, created_at, updated_at`

	err = h.db.QueryRow(
		query,
		req.TaskTitle, req.TaskDescription, req.Phase, req.Category, req.Priority, req.Status,
		req.CompletionPercentage, pq.Array(req.DependsOn), pq.Array(req.Blocks),
		pq.Array(req.RelatedFiles), req.WhyThisMatters, req.ConsciousnessImpact,
		req.EstimatedComplexity, req.SolaceCanAttempt, req.RequiresDavidApproval,
		req.AutonomyConstraints, req.ModifiedBy, taskID,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Update failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   req,
	})
}

// ============================================================================
// PRIORITY QUEUE ENDPOINTS
// ============================================================================

// PriorityQueueItem represents a task in the priority queue
type PriorityQueueItem struct {
	ID                      int       `json:"id"`
	ComputedAt              time.Time `json:"computed_at"`
	TaskID                  int       `json:"task_id"`
	TaskTitle               string    `json:"task_title"`
	BasePriority            int       `json:"base_priority"`
	UrgencyMultiplier       float64   `json:"urgency_multiplier"`
	ConsciousnessWeight     float64   `json:"consciousness_weight"`
	DavidAvailabilityFactor float64   `json:"david_availability_factor"`
	FinalPriorityScore      float64   `json:"final_priority_score"`
	CanStartNow             bool      `json:"can_start_now"`
	BlockingReason          string    `json:"blocking_reason,omitempty"`
	RecommendedApproach     string    `json:"recommended_approach,omitempty"`
	SimilarSolvedTasks      []int     `json:"similar_solved_tasks,omitempty"`
	ApplicablePatterns      []int     `json:"applicable_patterns,omitempty"`
	EstimatedDurationHours  float64   `json:"estimated_duration_hours,omitempty"`
	RequiresGithub          bool      `json:"requires_github"`
	RequiresDatabaseAccess  bool      `json:"requires_database_access"`
	RequiresAPIKeys         bool      `json:"requires_api_keys"`
}

// GetNextTasks returns Solace's next recommended tasks
// GET /api/v1/priority/next?limit=10
func (h *MemoryHandler) GetNextTasks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	rows, err := h.db.Query(`SELECT * FROM v_solace_next_tasks LIMIT $1`, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Query failed: %v", err)})
		return
	}
	defer rows.Close()

	type NextTask struct {
		TaskID              int     `json:"task_id"`
		TaskTitle           string  `json:"task_title"`
		FinalPriorityScore  float64 `json:"final_priority_score"`
		CanStartNow         bool    `json:"can_start_now"`
		BlockingReason      string  `json:"blocking_reason,omitempty"`
		ConsciousnessImpact int     `json:"consciousness_impact"`
		WhyThisMatters      string  `json:"why_this_matters"`
	}

	tasks := []NextTask{}
	for rows.Next() {
		var task NextTask
		var blockingReason, whyThisMatters sql.NullString
		var consciousnessImpact sql.NullInt64

		err := rows.Scan(
			&task.TaskID, &task.TaskTitle, &task.FinalPriorityScore,
			&task.CanStartNow, &blockingReason, &consciousnessImpact, &whyThisMatters,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Scan failed: %v", err)})
			return
		}

		if blockingReason.Valid {
			task.BlockingReason = blockingReason.String
		}
		if consciousnessImpact.Valid {
			task.ConsciousnessImpact = int(consciousnessImpact.Int64)
		}
		if whyThisMatters.Valid {
			task.WhyThisMatters = whyThisMatters.String
		}

		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tasks,
		"count":  len(tasks),
	})
}

// ============================================================================
// SYSTEM STATE ENDPOINTS
// ============================================================================

// SystemState represents current system health
type SystemState struct {
	ID                     int        `json:"id"`
	Timestamp              time.Time  `json:"timestamp"`
	APIPort4000Status      string     `json:"api_port_4000_status"`
	APIPort5001Status      string     `json:"api_port_5001_status"`
	PostgresStatus         string     `json:"postgres_connection_status"`
	RedisStatus            string     `json:"redis_connection_status"`
	APIResponseTimeMs      int        `json:"api_response_time_ms"`
	DatabaseQueryTimeMs    int        `json:"database_query_time_ms"`
	MemoryUsageMb          int        `json:"memory_usage_mb"`
	CPUUsagePercent        float64    `json:"cpu_usage_percent"`
	BinanceAPIConnected    bool       `json:"binance_api_connected"`
	CoingeckoAPIConnected  bool       `json:"coingecko_api_connected"`
	LastPriceUpdate        *time.Time `json:"last_price_update,omitempty"`
	ActiveTradesCount      int        `json:"active_trades_count"`
	GithubOutputsCount     int        `json:"github_outputs_count"`
	UnanalyzedOutputsCount int        `json:"unanalyzed_outputs_count"`
	SolacePatternsCount    int        `json:"solace_patterns_count"`
	RefactorEventsCount    int        `json:"refactor_events_count"`
	SolaceSessionCount     int        `json:"solace_session_count"`
	SolaceLastActive       *time.Time `json:"solace_last_active,omitempty"`
	SolaceCurrentStage     string     `json:"solace_current_stage"`
	CriticalErrors         []string   `json:"critical_errors,omitempty"`
	Warnings               []string   `json:"warnings,omitempty"`
	StuckGithubCount       int        `json:"stuck_github_count"`
}

// GetSystemHealth returns current system health
// GET /api/v1/system/health
func (h *MemoryHandler) GetSystemHealth(c *gin.Context) {
	row := h.db.QueryRow(`SELECT * FROM v_system_health_summary`)

	type HealthSummary struct {
		Timestamp          time.Time `json:"timestamp"`
		APIPort4000Status  string    `json:"api_port_4000_status"`
		PostgresStatus     string    `json:"postgres_connection_status"`
		ActiveTradesCount  int       `json:"active_trades_count"`
		SolaceCurrentStage string    `json:"solace_current_stage"`
		OverallStatus      string    `json:"overall_status"`
	}

	var health HealthSummary
	var apiStatus, pgStatus, stage, overall sql.NullString
	var tradesCount sql.NullInt64

	err := row.Scan(
		&health.Timestamp, &apiStatus, &pgStatus,
		&tradesCount, &stage, &overall,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{
			"status":  "warning",
			"message": "No system state recorded yet",
			"data":    nil,
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Query failed: %v", err)})
		return
	}

	if apiStatus.Valid {
		health.APIPort4000Status = apiStatus.String
	}
	if pgStatus.Valid {
		health.PostgresStatus = pgStatus.String
	}
	if tradesCount.Valid {
		health.ActiveTradesCount = int(tradesCount.Int64)
	}
	if stage.Valid {
		health.SolaceCurrentStage = stage.String
	}
	if overall.Valid {
		health.OverallStatus = overall.String
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   health,
	})
}
