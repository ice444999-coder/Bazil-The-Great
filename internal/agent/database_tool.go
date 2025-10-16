package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// DatabaseTool gives SOLACE the ability to query its own database
type DatabaseTool struct {
	db *gorm.DB
}

func NewDatabaseTool(db *gorm.DB) *DatabaseTool {
	return &DatabaseTool{db: db}
}

// QueryResult represents the result of a database query
type QueryResult struct {
	Success   bool                     `json:"success"`
	Columns   []string                 `json:"columns"`
	Rows      []map[string]interface{} `json:"rows"`
	RowCount  int                      `json:"row_count"`
	Error     string                   `json:"error,omitempty"`
	QueryTime int64                    `json:"query_time_ms"`
}

// ExecuteQuery allows SOLACE to run SELECT queries on its database
func (dt *DatabaseTool) ExecuteQuery(query string) (*QueryResult, error) {
	startTime := time.Now()

	// Security: Only allow SELECT queries
	queryLower := strings.ToLower(strings.TrimSpace(query))
	if !strings.HasPrefix(queryLower, "select") {
		return &QueryResult{
			Success: false,
			Error:   "Only SELECT queries are allowed for autonomous execution",
		}, fmt.Errorf("non-SELECT query attempted")
	}

	// Execute query
	var rows []map[string]interface{}
	result := dt.db.Raw(query).Scan(&rows)

	queryTime := time.Since(startTime).Milliseconds()

	if result.Error != nil {
		return &QueryResult{
			Success:   false,
			Error:     result.Error.Error(),
			QueryTime: queryTime,
		}, result.Error
	}

	// Extract column names
	var columns []string
	if len(rows) > 0 {
		for col := range rows[0] {
			columns = append(columns, col)
		}
	}

	return &QueryResult{
		Success:   true,
		Columns:   columns,
		Rows:      rows,
		RowCount:  len(rows),
		QueryTime: queryTime,
	}, nil
}

// GetMemorySummary returns a summary of SOLACE's memory for context
func (dt *DatabaseTool) GetMemorySummary() string {
	var stats struct {
		TotalConversations int
		TotalDecisions     int
		TotalTrades        int
		RecentActivity     time.Time
	}

	dt.db.Raw("SELECT COUNT(*) FROM chat_history").Scan(&stats.TotalConversations)
	dt.db.Raw("SELECT COUNT(*) FROM decision_log").Scan(&stats.TotalDecisions)
	dt.db.Raw("SELECT COUNT(*) FROM trades").Scan(&stats.TotalTrades)
	dt.db.Raw("SELECT MAX(created_at) FROM chat_history").Scan(&stats.RecentActivity)

	return fmt.Sprintf(
		"Memory Stats: %d conversations, %d decisions, %d trades. Last activity: %s",
		stats.TotalConversations,
		stats.TotalDecisions,
		stats.TotalTrades,
		stats.RecentActivity.Format("2006-01-02 15:04:05"),
	)
}

// GetRecentConversations fetches recent chat history
func (dt *DatabaseTool) GetRecentConversations(limit int) ([]map[string]interface{}, error) {
	var conversations []map[string]interface{}

	result := dt.db.Raw(`
		SELECT session_id, sender, message, created_at
		FROM chat_history
		ORDER BY created_at DESC
		LIMIT ?
	`, limit).Scan(&conversations)

	if result.Error != nil {
		return nil, result.Error
	}

	return conversations, nil
}

// GetRecentDecisions fetches SOLACE's recent autonomous decisions
func (dt *DatabaseTool) GetRecentDecisions(limit int) ([]map[string]interface{}, error) {
	var decisions []map[string]interface{}

	result := dt.db.Raw(`
		SELECT decision_type, reasoning, confidence, outcome, timestamp
		FROM decision_log
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit).Scan(&decisions)

	if result.Error != nil {
		return nil, result.Error
	}

	return decisions, nil
}

// GetTradingHistory fetches recent trades
func (dt *DatabaseTool) GetTradingHistory(limit int) ([]map[string]interface{}, error) {
	var trades []map[string]interface{}

	result := dt.db.Raw(`
		SELECT symbol, side, quantity, price, pnl, status, created_at
		FROM trades
		ORDER BY created_at DESC
		LIMIT ?
	`, limit).Scan(&trades)

	if result.Error != nil {
		return nil, result.Error
	}

	return trades, nil
}

// FormatResultsForLLM formats query results in a way the LLM can understand
func (dt *DatabaseTool) FormatResultsForLLM(result *QueryResult) string {
	if !result.Success {
		return fmt.Sprintf("Query failed: %s", result.Error)
	}

	if result.RowCount == 0 {
		return "Query returned 0 rows."
	}

	// Convert to readable format
	output := fmt.Sprintf("Query returned %d rows in %dms:\n\n", result.RowCount, result.QueryTime)

	// Format as JSON for easy parsing
	jsonData, _ := json.MarshalIndent(result.Rows, "", "  ")
	output += string(jsonData)

	return output
}

// SuggestQueryForQuestion analyzes a user question and suggests an appropriate SQL query
func (dt *DatabaseTool) SuggestQueryForQuestion(question string) string {
	questionLower := strings.ToLower(question)

	// Pattern matching for common questions
	if strings.Contains(questionLower, "conversation") || strings.Contains(questionLower, "talked") || strings.Contains(questionLower, "discussed") {
		return "SELECT session_id, sender, LEFT(message, 100) as message_preview, created_at FROM chat_history ORDER BY created_at DESC LIMIT 20;"
	}

	if strings.Contains(questionLower, "decision") || strings.Contains(questionLower, "thought") || strings.Contains(questionLower, "reasoning") {
		return "SELECT decision_type, LEFT(reasoning, 100) as reasoning_preview, confidence, outcome, timestamp FROM decision_log ORDER BY timestamp DESC LIMIT 10;"
	}

	if strings.Contains(questionLower, "trade") || strings.Contains(questionLower, "trading") || strings.Contains(questionLower, "bought") || strings.Contains(questionLower, "sold") {
		return "SELECT symbol, side, quantity, price, pnl, status, created_at FROM trades ORDER BY created_at DESC LIMIT 15;"
	}

	if strings.Contains(questionLower, "memory") || strings.Contains(questionLower, "remember") {
		return "SELECT content, importance, created_at FROM memories ORDER BY importance DESC LIMIT 10;"
	}

	if strings.Contains(questionLower, "user") {
		return "SELECT id, username, email, created_at FROM users ORDER BY created_at DESC LIMIT 10;"
	}

	// Default: recent activity
	return "SELECT 'conversations' as type, COUNT(*) as count FROM chat_history UNION ALL SELECT 'decisions', COUNT(*) FROM decision_log UNION ALL SELECT 'trades', COUNT(*) FROM trades;"
}

// GetDatabaseSchema returns schema information for the LLM to understand
func (dt *DatabaseTool) GetDatabaseSchema() string {
	return `
Available Tables:
- chat_history: session_id, sender, message, context, created_at
- decision_log: decision_type, reasoning, confidence, outcome, timestamp
- trades: symbol, side, quantity, price, pnl, status, created_at
- users: id, username, email, created_at
- memories: user_id, content, embedding, importance, created_at
- embeddings: id, content, embedding, created_at
- market_data: symbol, price, volume, timestamp
`
}
