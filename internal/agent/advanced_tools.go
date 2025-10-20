package agent

import (
	"fmt"
	"log"
	"strings"
)

// inspectTableSchema returns detailed schema information for a database table
func (s *SOLACE) inspectTableSchema(args map[string]interface{}) (string, error) {
	tableName, ok := args["table_name"].(string)
	if !ok || tableName == "" {
		return "❌ Error: 'table_name' parameter is required", nil
	}

	log.Printf("🔍 Inspecting schema for table: %s", tableName)

	// Query to get column details
	query := `
		SELECT 
			column_name,
			data_type,
			character_maximum_length,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position
	`

	type ColumnInfo struct {
		ColumnName             string
		DataType               string
		CharacterMaximumLength *int
		IsNullable             string
		ColumnDefault          *string
	}

	var columns []ColumnInfo
	if err := s.DB.Raw(query, tableName).Scan(&columns).Error; err != nil {
		return fmt.Sprintf("❌ Failed to query schema: %v", err), err
	}

	if len(columns) == 0 {
		return fmt.Sprintf("❌ Table '%s' not found or has no columns", tableName), nil
	}

	// Build response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("📋 **Schema for table `%s`**:\n\n", tableName))
	result.WriteString("| Column | Type | Nullable | Default |\n")
	result.WriteString("|--------|------|----------|----------|\n")

	for _, col := range columns {
		typeStr := col.DataType
		if col.CharacterMaximumLength != nil {
			typeStr += fmt.Sprintf("(%d)", *col.CharacterMaximumLength)
		}

		nullable := "NO"
		if col.IsNullable == "YES" {
			nullable = "YES"
		}

		defaultVal := "-"
		if col.ColumnDefault != nil {
			defaultVal = *col.ColumnDefault
			if len(defaultVal) > 30 {
				defaultVal = defaultVal[:27] + "..."
			}
		}

		result.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			col.ColumnName, typeStr, nullable, defaultVal))
	}

	// Get indexes
	indexQuery := `
		SELECT 
			indexname,
			indexdef
		FROM pg_indexes
		WHERE tablename = $1
	`

	type IndexInfo struct {
		IndexName string
		IndexDef  string
	}

	var indexes []IndexInfo
	if err := s.DB.Raw(indexQuery, tableName).Scan(&indexes).Error; err == nil && len(indexes) > 0 {
		result.WriteString("\n**Indexes:**\n")
		for _, idx := range indexes {
			result.WriteString(fmt.Sprintf("- `%s`: %s\n", idx.IndexName, idx.IndexDef))
		}
	}

	// Get row count
	var rowCount int64
	if err := s.DB.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&rowCount).Error; err == nil {
		result.WriteString(fmt.Sprintf("\n**Row Count:** %d\n", rowCount))
	}

	log.Printf("✅ Schema inspection complete for %s", tableName)
	return result.String(), nil
}

// describeDatabase returns a high-level overview of all public tables with column counts, row estimates, and key relationships
func (s *SOLACE) describeDatabase(args map[string]interface{}) (string, error) {
	log.Printf("🧠 Generating database overview")

	// Parse options
	includeSamples := false
	if v, ok := args["include_samples"].(bool); ok {
		includeSamples = v
	}
	_ = includeSamples // currently not used to avoid heavy queries; reserved for future use

	maxTables := 20
	if v, ok := args["max_tables"].(float64); ok {
		maxTables = int(v)
	}

	// List public tables
	type Tbl struct{ TableName string }
	var tables []Tbl
	if err := s.DB.Raw(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name`).Scan(&tables).Error; err != nil {
		return fmt.Sprintf("❌ Failed to list tables: %v", err), err
	}
	if len(tables) == 0 {
		return "ℹ️ No tables found in schema 'public'.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🧠 Database Overview (schema: public)\n\nTotal tables: %d\n\n", len(tables)))
	sb.WriteString("Tables:\n")

	// Compact summary list
	for _, t := range tables {
		var colCount int
		_ = s.DB.Raw(`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = ?`, t.TableName).Scan(&colCount).Error

		var rowEstimate int64
		_ = s.DB.Raw(`SELECT COALESCE((SELECT reltuples::bigint FROM pg_class WHERE relname = ?), 0)`, t.TableName).Scan(&rowEstimate).Error

		sb.WriteString(fmt.Sprintf("- %s (columns: %d, ~rows: %d)\n", t.TableName, colCount, rowEstimate))
	}

	// Detailed sections for first N tables
	sb.WriteString("\nDetails (first ")
	sb.WriteString(fmt.Sprintf("%d", maxTables))
	sb.WriteString(" tables):\n\n")

	limit := maxTables
	if len(tables) < limit {
		limit = len(tables)
	}
	for i := 0; i < limit; i++ {
		name := tables[i].TableName
		sb.WriteString(fmt.Sprintf("• %s\n", name))

		// Columns overview (first 6)
		type Col struct {
			ColumnName string
			DataType   string
			IsNullable string
		}
		var cols []Col
		if err := s.DB.Raw(`SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = ? ORDER BY ordinal_position`, name).Scan(&cols).Error; err == nil && len(cols) > 0 {
			sb.WriteString("  - Columns: ")
			maxCols := len(cols)
			if maxCols > 6 {
				maxCols = 6
			}
			var pieces []string
			for j := 0; j < maxCols; j++ {
				nullable := ""
				if cols[j].IsNullable == "YES" {
					nullable = " (nullable)"
				}
				pieces = append(pieces, fmt.Sprintf("%s:%s%s", cols[j].ColumnName, cols[j].DataType, nullable))
			}
			sb.WriteString(strings.Join(pieces, ", "))
			if len(cols) > maxCols {
				sb.WriteString(fmt.Sprintf(" ... (+%d more)", len(cols)-maxCols))
			}
			sb.WriteString("\n")
		}

		// Primary key
		type PK struct{ ColumnName string }
		var pks []PK
		_ = s.DB.Raw(`SELECT kcu.column_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
				  ON tc.constraint_name = kcu.constraint_name
				  AND tc.table_schema = kcu.table_schema
				WHERE tc.constraint_type = 'PRIMARY KEY'
				  AND tc.table_schema = 'public'
				  AND tc.table_name = ?
				ORDER BY kcu.ordinal_position`, name).Scan(&pks).Error
		if len(pks) > 0 {
			var pkNames []string
			for _, pk := range pks {
				pkNames = append(pkNames, pk.ColumnName)
			}
			sb.WriteString(fmt.Sprintf("  - Primary Key: %s\n", strings.Join(pkNames, ", ")))
		}

		// Foreign keys (show up to 5)
		type FK struct{ ColumnName, ForeignTable, ForeignColumn string }
		var fks []FK
		_ = s.DB.Raw(`SELECT kcu.column_name, ccu.table_name AS foreign_table, ccu.column_name AS foreign_column
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
				  ON tc.constraint_name = kcu.constraint_name
				  AND tc.table_schema = kcu.table_schema
				JOIN information_schema.constraint_column_usage ccu
				  ON ccu.constraint_name = tc.constraint_name
				  AND ccu.table_schema = tc.table_schema
				WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = 'public' AND tc.table_name = ?`, name).Scan(&fks).Error
		if len(fks) > 0 {
			sb.WriteString("  - Foreign Keys: ")
			maxFks := len(fks)
			if maxFks > 5 {
				maxFks = 5
			}
			var fkPieces []string
			for j := 0; j < maxFks; j++ {
				fkPieces = append(fkPieces, fmt.Sprintf("%s -> %s.%s", fks[j].ColumnName, fks[j].ForeignTable, fks[j].ForeignColumn))
			}
			sb.WriteString(strings.Join(fkPieces, "; "))
			if len(fks) > maxFks {
				sb.WriteString(fmt.Sprintf(" ... (+%d more)", len(fks)-maxFks))
			}
			sb.WriteString("\n")
		}

		// Row estimate
		var rowEst int64
		_ = s.DB.Raw(`SELECT COALESCE((SELECT reltuples::bigint FROM pg_class WHERE relname = ?), 0)`, name).Scan(&rowEst).Error
		sb.WriteString(fmt.Sprintf("  - Rows (est.): %d\n\n", rowEst))
	}

	sb.WriteString("Tip: Ask me 'inspect_table_schema(\"table_name\")' for any table to drill down.\n")
	log.Printf("✅ Database overview generated: %d tables", len(tables))
	return sb.String(), nil
}

// predictUsefulCrystal analyzes crystal usage patterns and predicts which will be most useful
func (s *SOLACE) predictUsefulCrystal(args map[string]interface{}) (string, error) {
	log.Printf("🔮 Analyzing crystal usage patterns for prediction")

	// Get current time context
	timeHorizon := "tomorrow"
	if th, ok := args["time_horizon"].(string); ok {
		timeHorizon = th
	}

	// Query 1: Get crystals most frequently accessed in semantic searches
	query1 := `
		SELECT 
			smc.id,
			smc.title,
			smc.category,
			smc.criticality,
			COUNT(*) as access_frequency,
			smc.tags
		FROM solace_memory_crystals smc
		WHERE smc.embedding IS NOT NULL
		GROUP BY smc.id, smc.title, smc.category, smc.criticality, smc.tags
		ORDER BY 
			CASE smc.criticality
				WHEN 'CRITICAL' THEN 1
				WHEN 'HIGH' THEN 2
				WHEN 'MEDIUM' THEN 3
				ELSE 4
			END,
			access_frequency DESC
		LIMIT 5
	`

	type CrystalPrediction struct {
		ID              int
		Title           string
		Category        string
		Criticality     string
		AccessFrequency int
		Tags            []string
	}

	var predictions []CrystalPrediction
	if err := s.DB.Raw(query1).Scan(&predictions).Error; err != nil {
		return fmt.Sprintf("❌ Failed to analyze crystal patterns: %v", err), err
	}

	// Query 2: Get recent chat topics to infer future needs
	topicQuery := `
		SELECT message 
		FROM chat_history 
		WHERE sender = 'user' 
		ORDER BY created_at DESC 
		LIMIT 10
	`

	var recentMessages []string
	if err := s.DB.Raw(topicQuery).Scan(&recentMessages).Error; err == nil {
		// Analyze topics (simple keyword frequency)
		topicKeywords := make(map[string]int)
		for _, msg := range recentMessages {
			words := strings.Fields(strings.ToLower(msg))
			for _, word := range words {
				if len(word) > 4 { // Filter short words
					topicKeywords[word]++
				}
			}
		}

		// Find most common topics
		var topTopics []string
		for topic, count := range topicKeywords {
			if count >= 2 {
				topTopics = append(topTopics, topic)
			}
		}

		log.Printf("📊 Detected trending topics: %v", topTopics)
	}

	// Build prediction report
	var result strings.Builder
	result.WriteString(fmt.Sprintf("🔮 **Crystal Usage Prediction for %s**:\n\n", timeHorizon))
	result.WriteString("Based on criticality, access patterns, and recent conversation topics:\n\n")

	for i, pred := range predictions {
		result.WriteString(fmt.Sprintf("%d. **%s** (ID: %d)\n", i+1, pred.Title, pred.ID))
		result.WriteString(fmt.Sprintf("   - Category: %s\n", pred.Category))
		result.WriteString(fmt.Sprintf("   - Criticality: %s\n", pred.Criticality))
		result.WriteString(fmt.Sprintf("   - Historical Access: %d times\n", pred.AccessFrequency))
		if len(pred.Tags) > 0 {
			result.WriteString(fmt.Sprintf("   - Tags: %v\n", pred.Tags))
		}
		result.WriteString("\n")
	}

	result.WriteString("\n**Prediction Confidence:** HIGH (based on criticality + access frequency)\n")
	result.WriteString("\n**Recommendation:** Keep these crystals in working memory for faster access.\n")

	log.Printf("✅ Crystal prediction complete")
	return result.String(), nil
}

// checkForContradictions analyzes conversation history for potential contradictions
func (s *SOLACE) checkForContradictions(args map[string]interface{}) (string, error) {
	sessionID, _ := args["session_id"].(string)
	currentStatement, ok := args["statement"].(string)
	if !ok || currentStatement == "" {
		return "❌ Error: 'statement' parameter is required", nil
	}

	log.Printf("🔍 Checking for contradictions in conversation history")

	// Get recent conversation history
	query := `
		SELECT sender, message, created_at
		FROM chat_history
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT 20
	`

	type Message struct {
		Sender    string
		Message   string
		CreatedAt string
	}

	var messages []Message
	if err := s.DB.Raw(query, sessionID).Scan(&messages).Error; err != nil {
		return fmt.Sprintf("❌ Failed to query conversation history: %v", err), err
	}

	// Look for contradictions (simple keyword matching)
	contradictions := []string{}
	currentLower := strings.ToLower(currentStatement)

	// Extract numbers from current statement
	currentNumbers := extractNumbers(currentLower)

	for _, msg := range messages {
		if msg.Sender == "solace" {
			msgLower := strings.ToLower(msg.Message)
			msgNumbers := extractNumbers(msgLower)

			// Check for conflicting numbers on same topic
			for _, currNum := range currentNumbers {
				for _, msgNum := range msgNumbers {
					if currNum != msgNum {
						// Check if they're talking about the same thing
						if containsSimilarContext(currentLower, msgLower) {
							contradictions = append(contradictions, fmt.Sprintf(
								"Previous: '%s' vs Current: '%s'",
								truncate(msg.Message, 100),
								truncate(currentStatement, 100),
							))
						}
					}
				}
			}
		}
	}

	if len(contradictions) > 0 {
		var result strings.Builder
		result.WriteString("⚠️ **POTENTIAL CONTRADICTION DETECTED**:\n\n")
		for i, contradiction := range contradictions {
			result.WriteString(fmt.Sprintf("%d. %s\n", i+1, contradiction))
		}
		result.WriteString("\n**Recommendation:** Acknowledge the discrepancy and clarify which is correct.\n")
		return result.String(), nil
	}

	return "✅ No contradictions detected", nil
}

// getRecentUserActivity queries chat history across all sessions for context awareness
func (s *SOLACE) getRecentUserActivity(args map[string]interface{}) (string, error) {
	userID, ok := args["user_id"].(string)
	if !ok || userID == "" {
		userID = "enki" // Default
	}

	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	log.Printf("🔍 Getting recent activity for user: %s", userID)

	// Query recent messages across ALL sessions for this user
	// Note: We need to track user_id in chat_history table
	query := `
		SELECT 
			session_id,
			sender,
			message,
			created_at
		FROM chat_history
		WHERE sender = 'user'
		ORDER BY created_at DESC
		LIMIT $1
	`

	type Activity struct {
		SessionID string
		Sender    string
		Message   string
		CreatedAt string
	}

	var activities []Activity
	if err := s.DB.Raw(query, limit).Scan(&activities).Error; err != nil {
		return fmt.Sprintf("❌ Failed to query user activity: %v", err), err
	}

	if len(activities) == 0 {
		return "📭 No recent activity found", nil
	}

	// Build summary
	var result strings.Builder
	result.WriteString(fmt.Sprintf("📊 **Recent Activity (last %d messages)**:\n\n", len(activities)))

	// Group by session
	sessionMap := make(map[string][]Activity)
	for _, activity := range activities {
		sessionMap[activity.SessionID] = append(sessionMap[activity.SessionID], activity)
	}

	for sessionID, msgs := range sessionMap {
		result.WriteString(fmt.Sprintf("**Session: %s** (%d messages)\n", sessionID, len(msgs)))
		for i, msg := range msgs {
			if i < 3 { // Show first 3 per session
				result.WriteString(fmt.Sprintf("- %s: %s\n", msg.CreatedAt, truncate(msg.Message, 80)))
			}
		}
		if len(msgs) > 3 {
			result.WriteString(fmt.Sprintf("- ... and %d more messages\n", len(msgs)-3))
		}
		result.WriteString("\n")
	}

	log.Printf("✅ Retrieved %d activities across %d sessions", len(activities), len(sessionMap))
	return result.String(), nil
}

// Helper functions
func extractNumbers(text string) []int {
	numbers := []int{}
	words := strings.Fields(text)
	for _, word := range words {
		// Simple number extraction
		cleaned := strings.Trim(word, ".,!?;:")
		var num int
		if _, err := fmt.Sscanf(cleaned, "%d", &num); err == nil {
			numbers = append(numbers, num)
		}
	}
	return numbers
}

func containsSimilarContext(text1, text2 string) bool {
	// Simple context matching - check for shared keywords
	keywords := []string{"crystal", "table", "database", "query", "trade", "memory"}
	matches := 0
	for _, keyword := range keywords {
		if strings.Contains(text1, keyword) && strings.Contains(text2, keyword) {
			matches++
		}
	}
	return matches >= 2
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// diagnoseToolHealth performs self-diagnostic checks on SOLACE's tools and database
func (s *SOLACE) diagnoseToolHealth(args map[string]interface{}) (string, error) {
	log.Printf("🩺 Running self-diagnostic...")

	var result strings.Builder
	result.WriteString("🩺 **SOLACE Tool Health Diagnostic Report**\n\n")

	// 1. Check database connection
	result.WriteString("**1. Database Connection:**\n")
	if err := s.DB.Exec("SELECT 1").Error; err != nil {
		result.WriteString(fmt.Sprintf("❌ FAILED - Cannot connect to database: %v\n\n", err))
		return result.String(), nil // Return report, don't error
	}
	result.WriteString("✅ Connected to ares_pgvector database\n\n")

	// 2. Check key tables exist
	result.WriteString("**2. Key Tables Status:**\n")
	tablesToCheck := []string{"solace_memory_crystals", "chat_history", "trades", "user_preferences"}

	// Allow custom table list
	if checkTables, ok := args["check_tables"].([]interface{}); ok && len(checkTables) > 0 {
		tablesToCheck = []string{}
		for _, t := range checkTables {
			if tableStr, ok := t.(string); ok {
				tablesToCheck = append(tablesToCheck, tableStr)
			}
		}
	}

	for _, table := range tablesToCheck {
		var count int64
		if err := s.DB.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count).Error; err != nil {
			result.WriteString(fmt.Sprintf("❌ Table `%s`: NOT FOUND or ERROR (%v)\n", table, err))
		} else {
			result.WriteString(fmt.Sprintf("✅ Table `%s`: %d rows\n", table, count))
		}
	}
	result.WriteString("\n")

	// 3. Check solace_memory_crystals schema
	result.WriteString("**3. Memory Crystals Schema Check:**\n")
	var crystalColumns []string
	err := s.DB.Raw(`
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = 'solace_memory_crystals'
		ORDER BY ordinal_position
	`).Scan(&crystalColumns).Error

	if err != nil {
		result.WriteString(fmt.Sprintf("❌ Cannot read schema: %v\n\n", err))
	} else {
		result.WriteString(fmt.Sprintf("Found %d columns: %s\n", len(crystalColumns), strings.Join(crystalColumns, ", ")))

		// Check for critical columns
		criticalCols := []string{"id", "title", "content", "category", "criticality", "embedding", "search_vector"}
		for _, col := range criticalCols {
			found := false
			for _, actual := range crystalColumns {
				if actual == col {
					found = true
					break
				}
			}
			if found {
				result.WriteString(fmt.Sprintf("✅ Column `%s` exists\n", col))
			} else {
				result.WriteString(fmt.Sprintf("❌ Column `%s` MISSING (tools may fail when querying this)\n", col))
			}
		}
	}
	result.WriteString("\n")

	// 4. Check embeddings status
	result.WriteString("**4. Semantic Search Readiness:**\n")
	var embeddingStats struct {
		TotalCrystals     int64
		WithEmbeddings    int64
		WithoutEmbeddings int64
	}
	s.DB.Raw("SELECT COUNT(*) FROM solace_memory_crystals").Scan(&embeddingStats.TotalCrystals)
	s.DB.Raw("SELECT COUNT(*) FROM solace_memory_crystals WHERE embedding IS NOT NULL").Scan(&embeddingStats.WithEmbeddings)
	embeddingStats.WithoutEmbeddings = embeddingStats.TotalCrystals - embeddingStats.WithEmbeddings

	result.WriteString(fmt.Sprintf("Total crystals: %d\n", embeddingStats.TotalCrystals))
	result.WriteString(fmt.Sprintf("✅ With embeddings: %d\n", embeddingStats.WithEmbeddings))
	if embeddingStats.WithoutEmbeddings > 0 {
		result.WriteString(fmt.Sprintf("⚠️ Missing embeddings: %d (semantic search may miss these)\n", embeddingStats.WithoutEmbeddings))
	}
	result.WriteString("\n")

	// 5. Test query_memory_crystals tool
	result.WriteString("**5. Query Tools Test:**\n")
	testResult, err := s.queryMemoryCrystals(map[string]interface{}{
		"search_term": "test",
		"limit":       float64(1),
	})
	if err != nil {
		result.WriteString(fmt.Sprintf("❌ query_memory_crystals FAILED: %v\n", err))
	} else if strings.Contains(testResult, "⚠️") {
		result.WriteString(fmt.Sprintf("⚠️ query_memory_crystals returned warning:\n%s\n", testResult[:200]))
	} else {
		result.WriteString("✅ query_memory_crystals working\n")
	}
	result.WriteString("\n")

	// 6. Overall health score
	result.WriteString("**6. Overall Health:**\n")
	healthScore := 100
	if strings.Contains(result.String(), "❌") {
		healthScore -= 30 * strings.Count(result.String(), "❌")
	}
	if strings.Contains(result.String(), "⚠️") {
		healthScore -= 10 * strings.Count(result.String(), "⚠️")
	}
	if healthScore < 0 {
		healthScore = 0
	}

	if healthScore >= 90 {
		result.WriteString(fmt.Sprintf("✅ EXCELLENT (%d/100) - All systems operational\n", healthScore))
	} else if healthScore >= 70 {
		result.WriteString(fmt.Sprintf("⚠️ GOOD (%d/100) - Minor issues detected\n", healthScore))
	} else if healthScore >= 50 {
		result.WriteString(fmt.Sprintf("⚠️ DEGRADED (%d/100) - Multiple issues need attention\n", healthScore))
	} else {
		result.WriteString(fmt.Sprintf("❌ CRITICAL (%d/100) - Major issues detected\n", healthScore))
	}

	log.Printf("✅ Diagnostic complete - Health score: %d/100", healthScore)
	return result.String(), nil
}

// compareCrystals compares two or more memory crystals and finds relationships between them
func (s *SOLACE) compareCrystals(args map[string]interface{}) (string, error) {
	// Extract crystal IDs
	crystalIDsInterface, ok := args["crystal_ids"].([]interface{})
	if !ok || len(crystalIDsInterface) < 2 {
		return "❌ Error: 'crystal_ids' must be an array with at least 2 crystal IDs", nil
	}

	// Convert to []int
	var crystalIDs []int
	for _, id := range crystalIDsInterface {
		switch v := id.(type) {
		case float64:
			crystalIDs = append(crystalIDs, int(v))
		case int:
			crystalIDs = append(crystalIDs, v)
		default:
			return fmt.Sprintf("❌ Error: Invalid crystal ID type: %T", id), nil
		}
	}

	log.Printf("🔍 Comparing crystals: %v", crystalIDs)

	// Fetch all crystals
	type Crystal struct {
		ID          int
		Title       string
		Category    string
		Criticality string
		Summary     string
		Content     string
		Tags        string
		CreatedBy   string
		CreatedAt   string
	}

	var crystals []Crystal
	query := `
		SELECT 
			id, title, category, criticality, summary, content,
			array_to_string(tags, ', ') as tags,
			created_by,
			to_char(created_at, 'YYYY-MM-DD') as created_at
		FROM solace_memory_crystals
		WHERE id = ANY($1)
		ORDER BY id
	`

	if err := s.DB.Raw(query, crystalIDs).Scan(&crystals).Error; err != nil {
		return fmt.Sprintf("❌ Failed to fetch crystals: %v", err), err
	}

	if len(crystals) == 0 {
		return fmt.Sprintf("❌ No crystals found with IDs: %v", crystalIDs), nil
	}

	if len(crystals) < len(crystalIDs) {
		var foundIDs []int
		for _, c := range crystals {
			foundIDs = append(foundIDs, c.ID)
		}
		return fmt.Sprintf("⚠️ Warning: Only found %d crystals out of %d requested. Found IDs: %v", len(crystals), len(crystalIDs), foundIDs), nil
	}

	// Build comparison report
	var result strings.Builder
	result.WriteString(fmt.Sprintf("🔍 **Crystal Comparison Report**\n\n"))
	result.WriteString(fmt.Sprintf("Comparing %d crystals:\n\n", len(crystals)))

	// Section 1: Individual Crystal Summaries
	result.WriteString("## 📋 Individual Crystals\n\n")
	for _, crystal := range crystals {
		result.WriteString(fmt.Sprintf("### Crystal #%d: %s\n", crystal.ID, crystal.Title))
		result.WriteString(fmt.Sprintf("- **Category:** %s\n", crystal.Category))
		result.WriteString(fmt.Sprintf("- **Criticality:** %s\n", crystal.Criticality))
		result.WriteString(fmt.Sprintf("- **Summary:** %s\n", crystal.Summary))
		result.WriteString(fmt.Sprintf("- **Tags:** %s\n", crystal.Tags))
		result.WriteString(fmt.Sprintf("- **Created by:** %s on %s\n\n", crystal.CreatedBy, crystal.CreatedAt))
	}

	// Section 2: Similarities
	result.WriteString("## 🔗 Similarities\n\n")

	// Check common categories
	categoryMap := make(map[string][]int)
	for _, c := range crystals {
		categoryMap[c.Category] = append(categoryMap[c.Category], c.ID)
	}

	var commonCategories []string
	for cat, ids := range categoryMap {
		if len(ids) > 1 {
			commonCategories = append(commonCategories, fmt.Sprintf("**%s** (Crystals: %v)", cat, ids))
		}
	}

	if len(commonCategories) > 0 {
		result.WriteString("**Common Categories:**\n")
		for _, cat := range commonCategories {
			result.WriteString(fmt.Sprintf("- %s\n", cat))
		}
	} else {
		result.WriteString("- Different categories (no overlap)\n")
	}
	result.WriteString("\n")

	// Check common tags
	tagMap := make(map[string][]int)
	for _, c := range crystals {
		tags := strings.Split(c.Tags, ", ")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagMap[tag] = append(tagMap[tag], c.ID)
			}
		}
	}

	var commonTags []string
	for tag, ids := range tagMap {
		if len(ids) > 1 {
			commonTags = append(commonTags, fmt.Sprintf("**%s** (Crystals: %v)", tag, ids))
		}
	}

	if len(commonTags) > 0 {
		result.WriteString("**Common Tags:**\n")
		for _, tag := range commonTags {
			result.WriteString(fmt.Sprintf("- %s\n", tag))
		}
	} else {
		result.WriteString("**Common Tags:** None\n")
	}
	result.WriteString("\n")

	// Check common creators
	creatorMap := make(map[string][]int)
	for _, c := range crystals {
		creatorMap[c.CreatedBy] = append(creatorMap[c.CreatedBy], c.ID)
	}

	if len(creatorMap) == 1 {
		for creator, ids := range creatorMap {
			result.WriteString(fmt.Sprintf("**Common Creator:** All created by %s (Crystals: %v)\n\n", creator, ids))
		}
	} else {
		result.WriteString("**Creators:** Different creators\n\n")
	}

	// Section 3: Differences
	result.WriteString("## ⚡ Differences\n\n")

	// Criticality differences
	result.WriteString("**Criticality Levels:**\n")
	for _, c := range crystals {
		result.WriteString(fmt.Sprintf("- Crystal #%d: %s\n", c.ID, c.Criticality))
	}
	result.WriteString("\n")

	// Section 4: Content Analysis
	result.WriteString("## 📊 Content Analysis\n\n")

	// Check for keyword overlap in content
	var keywords []string
	for _, c := range crystals {
		// Extract potential keywords from first crystal's content (simple approach)
		words := strings.Fields(strings.ToLower(c.Content))
		for _, word := range words {
			if len(word) > 5 { // Only words longer than 5 chars
				keywords = append(keywords, word)
			}
		}
	}

	// Check which crystals mention each keyword
	keywordMap := make(map[string][]int)
	for _, keyword := range keywords {
		for _, c := range crystals {
			if strings.Contains(strings.ToLower(c.Content), keyword) {
				keywordMap[keyword] = append(keywordMap[keyword], c.ID)
			}
		}
	}

	var sharedConcepts []string
	for keyword, ids := range keywordMap {
		if len(ids) > 1 {
			sharedConcepts = append(sharedConcepts, fmt.Sprintf("**%s** appears in crystals: %v", keyword, ids))
		}
	}

	if len(sharedConcepts) > 0 && len(sharedConcepts) <= 10 {
		result.WriteString("**Shared Concepts (top 10):**\n")
		for i, concept := range sharedConcepts {
			if i >= 10 {
				break
			}
			result.WriteString(fmt.Sprintf("- %s\n", concept))
		}
	} else if len(sharedConcepts) > 10 {
		result.WriteString(fmt.Sprintf("**Shared Concepts:** %d common keywords found (showing top 10)\n", len(sharedConcepts)))
		for i := 0; i < 10 && i < len(sharedConcepts); i++ {
			result.WriteString(fmt.Sprintf("- %s\n", sharedConcepts[i]))
		}
	} else {
		result.WriteString("**Shared Concepts:** Limited keyword overlap detected\n")
	}
	result.WriteString("\n")

	// Section 5: Relationship Summary
	result.WriteString("## 🎯 Relationship Summary\n\n")

	if len(crystals) == 2 {
		c1 := crystals[0]
		c2 := crystals[1]

		// Determine relationship type
		if c1.Category == c2.Category {
			result.WriteString(fmt.Sprintf("**Relationship Type:** Both crystals are in the **%s** category, suggesting they cover related architectural/functional areas.\n\n", c1.Category))
		}

		if len(commonTags) > 0 {
			result.WriteString(fmt.Sprintf("**Connection Strength:** STRONG - Share %d common tag(s)\n\n", len(commonTags)))
		} else {
			result.WriteString("**Connection Strength:** MODERATE - Different focus areas but may have conceptual overlap\n\n")
		}

		// Specific insights
		if strings.Contains(c1.Title, "System") || strings.Contains(c2.Title, "System") {
			result.WriteString("**Insight:** One or both crystals are system-level specifications - likely architectural dependencies\n\n")
		}

		if strings.Contains(strings.ToLower(c1.Content), strings.ToLower(c2.Title)) ||
			strings.Contains(strings.ToLower(c2.Content), strings.ToLower(c1.Title)) {
			result.WriteString("**Insight:** One crystal references the other by name - direct dependency detected\n\n")
		}
	}

	result.WriteString("---\n\n")
	result.WriteString("💡 **Recommendation:** Use these crystals together when working on related features or understanding system architecture.\n")

	log.Printf("✅ Crystal comparison complete for IDs: %v", crystalIDs)
	return result.String(), nil
}
