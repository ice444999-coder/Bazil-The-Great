package agent

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// executeSQLQuery executes a READ-ONLY SQL query (SELECT only) on the database
func (s *SOLACE) executeSQLQuery(args map[string]interface{}) (string, error) {
	// Extract query
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("missing or invalid 'query' parameter")
	}

	// Get optional explanation
	explain := ""
	if exp, ok := args["explain"].(string); ok {
		explain = exp
	}

	log.Printf("🔍 SQL Query Request: %s", explain)
	log.Printf("📝 Query: %s", query)

	// SQL SAFETY: Only allow SELECT statements
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(queryUpper, "SELECT") {
		return "", fmt.Errorf("❌ SECURITY: Only SELECT queries allowed. This query starts with: %s", strings.Split(queryUpper, " ")[0])
	}

	// Check for dangerous keywords
	dangerousKeywords := []string{"DROP", "DELETE", "UPDATE", "INSERT", "TRUNCATE", "ALTER", "CREATE", "GRANT", "REVOKE"}
	for _, keyword := range dangerousKeywords {
		if strings.Contains(queryUpper, keyword) {
			return "", fmt.Errorf("❌ SECURITY: Query contains dangerous keyword '%s' - only SELECT allowed", keyword)
		}
	}

	// Execute the query
	startTime := time.Now()
	rows, err := s.DB.Raw(query).Rows()
	if err != nil {
		log.Printf("❌ SQL query failed: %v", err)
		return "", fmt.Errorf("SQL error: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get columns: %v", err)
	}

	// Build result table
	var result strings.Builder
	if explain != "" {
		result.WriteString(fmt.Sprintf("📊 **Query Result:** %s\n\n", explain))
	}

	// Create markdown table header
	result.WriteString("| ")
	for _, col := range columns {
		result.WriteString(col)
		result.WriteString(" | ")
	}
	result.WriteString("\n| ")
	for range columns {
		result.WriteString("--- | ")
	}
	result.WriteString("\n")

	// Scan rows
	rowCount := 0
	maxRows := 100 // Default limit

	// Create slice for scanning
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if rowCount >= maxRows {
			result.WriteString(fmt.Sprintf("\n⚠️ *Result truncated at %d rows*\n", maxRows))
			break
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			log.Printf("❌ Row scan error: %v", err)
			continue
		}

		result.WriteString("| ")
		for _, val := range values {
			// Convert value to string
			var str string
			if val == nil {
				str = "NULL"
			} else {
				str = fmt.Sprintf("%v", val)
				// Truncate long strings
				if len(str) > 100 {
					str = str[:97] + "..."
				}
			}
			result.WriteString(str)
			result.WriteString(" | ")
		}
		result.WriteString("\n")
		rowCount++
	}

	elapsed := time.Since(startTime)

	// Add summary
	result.WriteString(fmt.Sprintf("\n📈 **Result:** %d rows returned in %dms\n", rowCount, elapsed.Milliseconds()))

	if rowCount == 0 {
		result.WriteString("\n💡 *No rows matched your query*\n")
	}

	log.Printf("✅ SQL query executed: %d rows in %dms", rowCount, elapsed.Milliseconds())

	return result.String(), nil
}
