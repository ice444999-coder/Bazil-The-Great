package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DatabaseController handles direct SQL query execution for SOLACE
type DatabaseController struct {
	db *gorm.DB
}

func NewDatabaseController(db *gorm.DB) *DatabaseController {
	return &DatabaseController{db: db}
}

// QueryRequest represents an SQL query request
type QueryRequest struct {
	Query     string `json:"query" binding:"required"`
	SessionID string `json:"session_id"`
}

// QueryResponse represents the result of a query
type QueryResponse struct {
	Success   bool                     `json:"success"`
	Columns   []string                 `json:"columns,omitempty"`
	Rows      []map[string]interface{} `json:"rows,omitempty"`
	RowCount  int                      `json:"row_count"`
	Error     string                   `json:"error,omitempty"`
	QueryTime int64                    `json:"query_time_ms"`
	Query     string                   `json:"query"`
}

// ExecuteQuery allows SOLACE to execute SQL queries
// @Summary Execute SQL Query
// @Description Execute a read-only SQL query on the ARES database
// @Tags database
// @Accept json
// @Produce json
// @Param request body QueryRequest true "SQL Query"
// @Success 200 {object} QueryResponse
// @Router /api/v1/database/query [post]
func (dc *DatabaseController) ExecuteQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Security: Only allow SELECT queries (read-only)
	queryLower := strings.ToLower(strings.TrimSpace(req.Query))
	if !strings.HasPrefix(queryLower, "select") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Only SELECT queries are allowed. For safety, INSERT/UPDATE/DELETE must use specific endpoints.",
		})
		return
	}

	// Execute query
	var rows []map[string]interface{}
	result := dc.db.Raw(req.Query).Scan(&rows)

	queryTime := time.Since(startTime).Milliseconds()

	if result.Error != nil {
		c.JSON(http.StatusOK, QueryResponse{
			Success:   false,
			Error:     result.Error.Error(),
			Query:     req.Query,
			QueryTime: queryTime,
		})
		return
	}

	// Extract column names
	var columns []string
	if len(rows) > 0 {
		for col := range rows[0] {
			columns = append(columns, col)
		}
	}

	// Log query for SOLACE's memory
	if req.SessionID != "" {
		dc.db.Exec(`
			INSERT INTO chat_history (session_id, sender, message, context, created_at)
			VALUES (?, 'system', ?, '{"type":"sql_query","row_count":?}', NOW())
		`, req.SessionID, fmt.Sprintf("SQL Query: %s", req.Query), len(rows))
	}

	c.JSON(http.StatusOK, QueryResponse{
		Success:   true,
		Columns:   columns,
		Rows:      rows,
		RowCount:  len(rows),
		Query:     req.Query,
		QueryTime: queryTime,
	})
}

// GetTables returns all tables in the database
// @Summary List Database Tables
// @Description Get a list of all tables in the ARES database
// @Tags database
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/database/tables [get]
func (dc *DatabaseController) GetTables(c *gin.Context) {
	var tables []struct {
		TableName string `json:"table_name"`
	}

	dc.db.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`).Scan(&tables)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tables":  tables,
		"count":   len(tables),
	})
}

// GetTableSchema returns the schema for a specific table
// @Summary Get Table Schema
// @Description Get column information for a specific table
// @Tags database
// @Param table path string true "Table Name"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/database/tables/{table}/schema [get]
func (dc *DatabaseController) GetTableSchema(c *gin.Context) {
	tableName := c.Param("table")

	var columns []struct {
		ColumnName string `json:"column_name"`
		DataType   string `json:"data_type"`
		IsNullable string `json:"is_nullable"`
	}

	dc.db.Raw(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = ?
		ORDER BY ordinal_position
	`, tableName).Scan(&columns)

	if len(columns) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Table '%s' not found", tableName),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"table_name": tableName,
		"columns":    columns,
		"count":      len(columns),
	})
}

// GetSOLACEMemoryStats returns statistics about SOLACE's memory
// @Summary Get SOLACE Memory Statistics
// @Description Get statistics about SOLACE's conversation history and memory
// @Tags database
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/database/solace-memory [get]
func (dc *DatabaseController) GetSOLACEMemoryStats(c *gin.Context) {
	var stats struct {
		TotalConversations int       `json:"total_conversations"`
		TotalDecisions     int       `json:"total_decisions"`
		TotalTrades        int       `json:"total_trades"`
		FirstMessage       time.Time `json:"first_message"`
		LastMessage        time.Time `json:"last_message"`
	}

	// Get conversation count
	dc.db.Raw("SELECT COUNT(*) FROM chat_history").Scan(&stats.TotalConversations)

	// Get decision count
	dc.db.Raw("SELECT COUNT(*) FROM decision_log").Scan(&stats.TotalDecisions)

	// Get trade count
	dc.db.Raw("SELECT COUNT(*) FROM trades").Scan(&stats.TotalTrades)

	// Get first and last message times
	dc.db.Raw("SELECT MIN(created_at) FROM chat_history").Scan(&stats.FirstMessage)
	dc.db.Raw("SELECT MAX(created_at) FROM chat_history").Scan(&stats.LastMessage)

	// Get recent conversations
	var recentConversations []struct {
		SessionID string    `json:"session_id"`
		Sender    string    `json:"sender"`
		Message   string    `json:"message"`
		CreatedAt time.Time `json:"created_at"`
	}
	dc.db.Raw(`
		SELECT session_id, sender, message, created_at
		FROM chat_history
		ORDER BY created_at DESC
		LIMIT 10
	`).Scan(&recentConversations)

	c.JSON(http.StatusOK, gin.H{
		"success":              true,
		"stats":                stats,
		"recent_conversations": recentConversations,
	})
}
