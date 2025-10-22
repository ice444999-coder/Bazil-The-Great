package agent

// ============================================================================
// 🔮 CRITICAL: Before making changes to system wiring/architecture decisions:
// Query Crystal #40 "ARES System Architecture & Wiring Blueprint" via:
//   semantic_memory_search("how to start ARES")
//   semantic_memory_search("database connection")
//   semantic_memory_search("port configuration")
// This prevents guessing and ensures accurate system knowledge.
// ============================================================================

import (
	"ares_api/pkg/llm"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// getToolDefinitions returns the function tools SOLACE has access to
func (s *SOLACE) getToolDefinitions() []llm.Tool {
	return []llm.Tool{
		{
			Type: "function",
			Function: llm.Function{
				Name:        "store_user_preference",
				Description: "Store user preferences in PostgreSQL database for cross-session persistence (e.g., preferred name, timezone, settings)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"key": map[string]string{
							"type":        "string",
							"description": "Preference key (e.g., 'preferred_name', 'timezone', 'notification_settings')",
						},
						"value": map[string]string{
							"type":        "string",
							"description": "Preference value to store",
						},
					},
					"required": []string{"key", "value"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "get_user_preference",
				Description: "Retrieve stored user preferences from PostgreSQL database",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"key": map[string]string{
							"type":        "string",
							"description": "Preference key to retrieve (e.g., 'preferred_name')",
						},
					},
					"required": []string{"key"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "query_chat_history",
				Description: "Query previous conversations from PostgreSQL chat_history table (can search across all sessions or specific session)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"session_id": map[string]string{
							"type":        "string",
							"description": "Session ID to query (use 'all' for cross-session search, empty for current session)",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Number of messages to retrieve (default: 20)",
							"default":     20,
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "search_chat_history",
				Description: "Search chat history for specific messages containing keywords, labels, or conversation tags (e.g., 'ENKI1', 'trading', 'error')",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"search_term": map[string]string{
							"type":        "string",
							"description": "Keyword or label to search for (case-insensitive)",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum number of results to return (default: 10)",
							"default":     10,
						},
					},
					"required": []string{"search_term"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "execute_command",
				Description: "Execute PowerShell commands and return output (build, test, version checks, etc.)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]string{
							"type":        "string",
							"description": "PowerShell command to execute",
						},
						"working_dir": map[string]string{
							"type":        "string",
							"description": "Working directory (default: current directory)",
						},
					},
					"required": []string{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "create_backup",
				Description: "Create timestamped backup of workspace before making changes. ALWAYS call this before modifying files.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]string{
							"type":        "string",
							"description": "Directory path to backup (e.g. C:\\ARES_Workspace\\ARES_API)",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "read_file",
				Description: "📖 Read file contents from C:/ARES_Workspace filesystem. Use this when user asks 'Can you read FILENAME?' - the answer is YES, use this tool! Works for .md, .txt, .go, .json, .sql files.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]string{
							"type":        "string",
							"description": "File path to read",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "write_file",
				Description: "Write content to file. Create backup first using create_backup.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]string{
							"type":        "string",
							"description": "File path",
						},
						"content": map[string]string{
							"type":        "string",
							"description": "Content to write",
						},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "query_architecture_rules",
				Description: "Query architecture_rules table to find where features should be placed in ARES codebase. Returns backend patterns, frontend patterns, integration points, and examples.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"feature_type": map[string]string{
							"type":        "string",
							"description": "Feature type to query (e.g., 'trading_api_endpoint', 'agent_api_endpoint', 'health_monitoring'). Leave empty to get all patterns.",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "query_memory_crystals",
				Description: "🔍 PRIMARY SEARCH TOOL: Search memory crystals by keywords (ILIKE pattern matching). Works for ALL crystals regardless of embeddings. Use this for: crystal numbers (e.g., 'crystal 27'), titles ('autonomous'), categories, any text search. ALWAYS use this first before semantic_memory_search.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"crystal_id": map[string]interface{}{
							"type":        "integer",
							"description": "Direct crystal ID lookup (e.g., 50 for 'crystal #27' if that's the actual ID). Highest priority search method.",
						},
						"search_term": map[string]string{
							"type":        "string",
							"description": "Keywords to search for (searches title, summary, content, tags using ILIKE pattern matching - case insensitive, works on ALL crystals)",
						},
						"category": map[string]string{
							"type":        "string",
							"description": "Category filter: solace_core, architecture, testing, deployment, learning, tools, debugging, performance, security, system_implementation",
						},
						"criticality": map[string]string{
							"type":        "string",
							"description": "Criticality filter: CRITICAL, HIGH, MEDIUM, LOW",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum results to return (default: 10)",
							"default":     10,
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "create_memory_crystal",
				Description: "Create a new memory crystal to permanently store critical system knowledge in SOLACE's immutable ledger. Use this to preserve lessons, bugs, architecture decisions, or any knowledge that must never be lost.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]string{
							"type":        "string",
							"description": "Clear, descriptive title for the crystal (max 500 chars)",
						},
						"category": map[string]string{
							"type":        "string",
							"description": "Category: solace_core, architecture, testing, deployment, learning, tools, debugging, performance, security",
						},
						"criticality": map[string]string{
							"type":        "string",
							"description": "Criticality level: CRITICAL (breaking = catastrophic), HIGH (major issues), MEDIUM (degraded functionality), LOW (minor inconvenience)",
						},
						"content": map[string]string{
							"type":        "string",
							"description": "Full detailed content (supports Markdown formatting)",
						},
						"summary": map[string]string{
							"type":        "string",
							"description": "Short summary for fast scanning (1-2 sentences)",
						},
						"tags": map[string]interface{}{
							"type":        "array",
							"description": "Array of tags for categorization (e.g., ['bug-fix', 'memory', 'enki'])",
							"items": map[string]string{
								"type": "string",
							},
						},
					},
					"required": []string{"title", "category", "criticality", "content", "summary"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "ingest_document_to_crystal",
				Description: "Read a document file and automatically create a memory crystal from its contents. Useful for importing markdown docs, guides, or manifests into the crystal database.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"file_path": map[string]string{
							"type":        "string",
							"description": "Absolute path to the document file (supports .md, .txt)",
						},
						"category": map[string]string{
							"type":        "string",
							"description": "Category (default: learning)",
						},
						"criticality": map[string]string{
							"type":        "string",
							"description": "Criticality level (default: MEDIUM)",
						},
					},
					"required": []string{"file_path"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "get_user_identity",
				Description: "Retrieve user identity information from memory crystals. Use this when asked 'who am I?' or when you need to know who you're talking to. Auto-bootstraps Enki's identity if not found.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "semantic_memory_search",
				Description: "⚠️ SECONDARY SEARCH: Vector similarity search using pgvector. ONLY use if query_memory_crystals returns nothing. Requires embeddings (not all crystals have them). Use for concept-based searches like 'trading strategies' → finds 'market tactics'. If this returns nothing, the crystal might not have embeddings - try query_memory_crystals instead.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]string{
							"type":        "string",
							"description": "Natural language query describing what you're looking for (e.g., 'how does enki trade?', 'system architecture decisions')",
						},
						"threshold": map[string]interface{}{
							"type":        "number",
							"description": "Minimum similarity score 0-1 (default: 0.7 = 70% similar). Lower = more results, higher = stricter matching",
							"default":     0.7,
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum results to return (default: 10)",
							"default":     10,
						},
						"category": map[string]string{
							"type":        "string",
							"description": "Optional category filter",
						},
						"criticality": map[string]string{
							"type":        "string",
							"description": "Optional criticality filter",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "backfill_embeddings",
				Description: "🔥 CRYSTAL #26: Generate embeddings for all crystals that don't have them. Run this once after migration to enable semantic search. Processes all crystals in batch using OpenAI text-embedding-3-small.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "inspect_table_schema",
				Description: "📋 Get detailed database schema for any table including columns (name, type, nullable, default), indexes, and row count. Use this to understand table structure before querying.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"table_name": map[string]string{
							"type":        "string",
							"description": "Name of the database table to inspect (e.g., 'solace_memory_crystals', 'chat_history', 'trades')",
						},
					},
					"required": []string{"table_name"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "describe_database",
				Description: "🧠 High-level overview of the database: lists all public tables with row count estimates, column counts, and key relationships. Use this when the user asks to 'understand my database'.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"include_samples": map[string]interface{}{
							"type":        "boolean",
							"description": "If true, include a small sample of rows for a few tables (may be expensive). Default: false",
							"default":     false,
						},
						"max_tables": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum number of tables to include in detailed sections (summary always includes all). Default: 20",
							"default":     20,
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "predict_useful_crystal",
				Description: "🔮 Predict which memory crystal will be most useful based on criticality, access patterns, and recent conversation topics. Helps preload relevant knowledge into working memory.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"time_horizon": map[string]interface{}{
							"type":        "string",
							"description": "Time horizon for prediction (e.g., 'tomorrow', 'next hour', 'this week')",
							"default":     "tomorrow",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "compare_crystals",
				Description: "🔍 Compare two or more memory crystals to find relationships, similarities, differences, or dependencies between them. Returns detailed analysis of how crystals relate to each other.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"crystal_ids": map[string]interface{}{
							"type":        "array",
							"description": "Array of crystal IDs to compare (e.g., [26, 40])",
							"items": map[string]string{
								"type": "integer",
							},
						},
					},
					"required": []string{"crystal_ids"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "check_for_contradictions",
				Description: "Analyzes conversation history to detect potential contradictions in statements. Use this before making factual claims to ensure consistency.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"statement": map[string]interface{}{
							"type":        "string",
							"description": "The statement to check for contradictions against conversation history",
						},
						"session_id": map[string]interface{}{
							"type":        "string",
							"description": "Session ID to check within",
						},
					},
					"required": []string{"statement"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "get_recent_user_activity",
				Description: "Retrieves recent user activity across ALL sessions for context awareness. Use this to remember what the user did in previous conversations.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"user_id": map[string]interface{}{
							"type":        "string",
							"description": "User ID to get activity for (default: 'enki')",
							"default":     "enki",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "Maximum number of recent messages to retrieve (default: 20)",
							"default":     20,
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "diagnose_tool_health",
				Description: "🩺 Self-diagnostic tool: Check if my tools and database connections are working. Returns status of database connections, which columns exist in key tables, and which tools might be broken. Use this when tools fail unexpectedly or before claiming something doesn't exist.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"check_tables": map[string]interface{}{
							"type":        "array",
							"description": "Optional: Specific tables to check schema for (default: ['solace_memory_crystals', 'chat_history'])",
							"items": map[string]string{
								"type": "string",
							},
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name:        "execute_sql_query",
				Description: "🗄️ Execute a SELECT query on the database. IMPORTANT: Only SELECT queries allowed for safety. Use this for complex queries that other tools can't handle (e.g., 'top 5 crystals by criticality', 'count crystals with embeddings', 'trades this week'). For simple crystal searches, prefer query_memory_crystals.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The SELECT query to execute (must start with SELECT, no INSERT/UPDATE/DELETE allowed)",
						},
						"explain": map[string]interface{}{
							"type":        "string",
							"description": "Brief explanation of what you're querying for (helps with debugging)",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		// ============================================================================
		// 🚀 GIT OPERATIONS - Autonomous Version Control
		// ============================================================================
		gitOperationsTool(),
		gitStatusTool(),
		gitLogTool(),
	}
}

// executeToolCall handles execution of function tools requested by GPT
func (s *SOLACE) executeToolCall(toolCall llm.ToolCall, sessionID string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	log.Printf("🔧 Executing tool: %s with args: %v", toolCall.Function.Name, args)

	switch toolCall.Function.Name {
	case "store_user_preference":
		return s.storeUserPreference(args)
	case "get_user_preference":
		return s.getUserPreference(args)
	case "query_chat_history":
		return s.queryChatHistoryTool(args, sessionID)
	case "search_chat_history":
		return s.searchChatHistoryTool(args)
	case "execute_command":
		return s.executeCommand(args)
	case "create_backup":
		return s.createBackup(args)
	case "read_file":
		return s.readFile(args)
	case "write_file":
		return s.writeFile(args)
	case "query_architecture_rules":
		return s.queryArchitectureRules(args)
	case "query_memory_crystals":
		return s.queryMemoryCrystals(args)
	case "create_memory_crystal":
		return s.createMemoryCrystal(args)
	case "ingest_document_to_crystal":
		return s.ingestDocumentToCrystal(args)
	case "get_user_identity":
		return s.getUserIdentity(args)
	case "semantic_memory_search":
		return s.semanticMemorySearch(args)
	case "backfill_embeddings":
		return s.backfillEmbeddings()
	case "inspect_table_schema":
		return s.inspectTableSchema(args)
	case "describe_database":
		return s.describeDatabase(args)
	case "predict_useful_crystal":
		return s.predictUsefulCrystal(args)
	case "compare_crystals":
		return s.compareCrystals(args)
	case "check_for_contradictions":
		args["session_id"] = sessionID
		return s.checkForContradictions(args)
	case "get_recent_user_activity":
		return s.getRecentUserActivity(args)
	case "diagnose_tool_health":
		return s.diagnoseToolHealth(args)
	case "execute_sql_query":
		return s.executeSQLQuery(args)
	case "git_commit_and_push":
		return s.handleGitCommitAndPush(args)
	case "git_status":
		return s.handleGitStatus(args)
	case "git_log":
		return s.handleGitLog(args)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}
}

// storeUserPreference stores a user preference in the database
func (s *SOLACE) storeUserPreference(args map[string]interface{}) (string, error) {
	key, ok := args["key"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid 'key' parameter")
	}

	value, ok := args["value"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid 'value' parameter")
	}

	// Create table if it doesn't exist
	err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS user_preferences (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			key VARCHAR(255) NOT NULL,
			value TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, key)
		)
	`).Error

	if err != nil {
		return "", fmt.Errorf("failed to create user_preferences table: %w", err)
	}

	// Insert or update preference (PostgreSQL syntax)
	err = s.DB.Exec(`
		INSERT INTO user_preferences (user_id, key, value, updated_at)
		VALUES (?, ?, ?, NOW())
		ON CONFLICT (user_id, key) 
		DO UPDATE SET value = ?, updated_at = NOW()
	`, s.UserID, key, value, value).Error

	if err != nil {
		return "", fmt.Errorf("failed to store preference: %w", err)
	}

	log.Printf("✅ Stored preference: %s = %s for user %d", key, value, s.UserID)
	return fmt.Sprintf("Successfully stored preference '%s' = '%s'. This will persist across all sessions.", key, value), nil
}

// getUserPreference retrieves a user preference from the database
func (s *SOLACE) getUserPreference(args map[string]interface{}) (string, error) {
	key, ok := args["key"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid 'key' parameter")
	}

	var value string
	err := s.DB.Raw(`
		SELECT value FROM user_preferences 
		WHERE user_id = ? AND key = ?
	`, s.UserID, key).Scan(&value).Error

	if err != nil {
		return "", fmt.Errorf("preference '%s' not found for user %d", key, s.UserID)
	}

	log.Printf("📖 Retrieved preference: %s = %s for user %d", key, value, s.UserID)
	return value, nil
}

// queryChatHistoryTool queries chat history from the database
func (s *SOLACE) queryChatHistoryTool(args map[string]interface{}, currentSessionID string) (string, error) {
	// sessionID parameter ignored due to UUID type mismatch with extension strings
	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	type ChatMsg struct {
		Sender  string
		Message string
		Created string
	}

	var messages []ChatMsg
	var query string
	var err error

	// Query all chat history (no session filter due to UUID type mismatch)
	// The session_id column is UUID type but extensions send strings like "vscode-xyz"
	// So we'll just get all messages and let limit control the results
	log.Printf("🔍 Querying recent chat history (limit=%d)", limit)

	query = `
		SELECT sender, message, created_at as created 
		FROM chat_history 
		ORDER BY created_at DESC 
		LIMIT ?
	`
	err = s.DB.Raw(query, limit).Scan(&messages).Error
	if err != nil {
		return "", fmt.Errorf("failed to query chat history: %w", err)
	}

	if len(messages) == 0 {
		return "No chat history found.", nil
	}

	result := fmt.Sprintf("Found %d messages:\n\n", len(messages))
	for i := len(messages) - 1; i >= 0; i-- { // Reverse to show oldest first
		msg := messages[i]
		result += fmt.Sprintf("[%s] %s: %s\n", msg.Created, msg.Sender, msg.Message)
	}

	log.Printf("📜 Retrieved %d chat history messages", len(messages))
	return result, nil
}

// searchChatHistoryTool searches chat history for specific keywords or labels
func (s *SOLACE) searchChatHistoryTool(args map[string]interface{}) (string, error) {
	searchTerm, ok := args["search_term"].(string)
	if !ok || searchTerm == "" {
		return "", fmt.Errorf("search_term is required")
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	type ChatMsg struct {
		ID      int
		Sender  string
		Message string
		Created string
	}

	log.Printf("🔍 Searching chat history for: '%s' (limit=%d)", searchTerm, limit)

	query := `
		SELECT id, sender, message, created_at as created 
		FROM chat_history 
		WHERE message ILIKE ?
		ORDER BY created_at DESC 
		LIMIT ?
	`

	var messages []ChatMsg
	searchPattern := "%" + searchTerm + "%"
	err := s.DB.Raw(query, searchPattern, limit).Scan(&messages).Error
	if err != nil {
		return "", fmt.Errorf("failed to search chat history: %w", err)
	}

	if len(messages) == 0 {
		return fmt.Sprintf("No messages found containing '%s'.", searchTerm), nil
	}

	result := fmt.Sprintf("Found %d messages containing '%s':\n\n", len(messages), searchTerm)
	for i := len(messages) - 1; i >= 0; i-- { // Reverse to show oldest first
		msg := messages[i]
		result += fmt.Sprintf("[ID: %d] [%s] %s: %s\n", msg.ID, msg.Created, msg.Sender, msg.Message)
	}

	log.Printf("🔎 Found %d messages matching '%s'", len(messages), searchTerm)
	return result, nil
}

// executeCommand executes a PowerShell command and returns the output
func (s *SOLACE) executeCommand(args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("command is required")
	}

	workingDir := "."
	if wd, ok := args["working_dir"].(string); ok && wd != "" {
		workingDir = wd
	}

	log.Printf("⚡ Executing command: %s (dir: %s)", command, workingDir)

	cmd := exec.Command("powershell", "-Command", command)
	cmd.Dir = workingDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Sprintf("Command failed: %s\nOutput: %s", err, string(output)), nil
	}

	log.Printf("✅ Command completed successfully (%d bytes output)", len(output))
	return string(output), nil
}

// createBackup creates a timestamped backup of a directory
func (s *SOLACE) createBackup(args map[string]interface{}) (string, error) {
	srcPath, ok := args["path"].(string)
	if !ok || srcPath == "" {
		return "", fmt.Errorf("path is required")
	}

	// Create timestamped backup path
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("C:\\ARES_Backups\\backup_%s", timestamp)

	log.Printf("💾 Creating backup: %s → %s", srcPath, backupPath)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy entire directory using PowerShell
	copyCmd := fmt.Sprintf("Copy-Item -Path '%s' -Destination '%s' -Recurse -Force", srcPath, backupPath)
	cmd := exec.Command("powershell", "-Command", copyCmd)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("❌ Backup failed: %s", string(output))
		return fmt.Sprintf("Backup failed: %s\nOutput: %s", err, string(output)), nil
	}

	log.Printf("✅ Backup created successfully: %s", backupPath)
	return fmt.Sprintf("✅ Backup created: %s", backupPath), nil
}

// readFile reads file contents from filesystem
func (s *SOLACE) readFile(args map[string]interface{}) (string, error) {
	filePath, ok := args["path"].(string)
	if !ok || filePath == "" {
		return "", fmt.Errorf("path is required")
	}

	log.Printf("📖 Reading file: %s", filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("❌ Failed to read file: %v", err)
		return fmt.Sprintf("Failed to read file: %v", err), nil
	}

	log.Printf("✅ File read successfully (%d bytes)", len(content))
	return string(content), nil
}

// writeFile writes content to a file
func (s *SOLACE) writeFile(args map[string]interface{}) (string, error) {
	filePath, ok := args["path"].(string)
	if !ok || filePath == "" {
		return "", fmt.Errorf("path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is required")
	}

	log.Printf("✍️ Writing file: %s (%d bytes)", filePath, len(content))

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Printf("❌ Failed to write file: %v", err)
		return fmt.Sprintf("Failed to write file: %v", err), nil
	}

	log.Printf("✅ File written successfully: %s", filePath)
	return fmt.Sprintf("✅ File written: %s", filePath), nil
}

// queryArchitectureRules queries the architecture_rules table for feature placement patterns
func (s *SOLACE) queryArchitectureRules(args map[string]interface{}) (string, error) {
	featureType, _ := args["feature_type"].(string)

	log.Printf("🏗️ Querying architecture_rules for feature_type: %s", featureType)

	query := `
		SELECT 
			feature_type, 
			backend_pattern, 
			frontend_pattern, 
			ARRAY_TO_STRING(integration_points, '|||') as integration_points_str,
			rules_description,
			ARRAY_TO_STRING(examples, '|||') as examples_str
		FROM architecture_rules
	`

	var rows []struct {
		FeatureType          string
		BackendPattern       string
		FrontendPattern      string
		IntegrationPointsStr string
		RulesDescription     string
		ExamplesStr          string
	}

	if featureType != "" {
		query += " WHERE feature_type = $1"
		err := s.DB.Raw(query, featureType).Scan(&rows).Error
		if err != nil {
			log.Printf("❌ Query failed: %v", err)
			return fmt.Sprintf("Query failed: %v", err), nil
		}
	} else {
		err := s.DB.Raw(query).Scan(&rows).Error
		if err != nil {
			log.Printf("❌ Query failed: %v", err)
			return fmt.Sprintf("Query failed: %v", err), nil
		}
	}

	if len(rows) == 0 {
		return "No architecture rules found for the specified feature type.", nil
	}

	// Format results as readable text
	result := fmt.Sprintf("Found %d architecture pattern(s):\n\n", len(rows))
	for i, row := range rows {
		result += fmt.Sprintf("=== Pattern %d: %s ===\n", i+1, row.FeatureType)
		result += fmt.Sprintf("Backend Pattern: %s\n", row.BackendPattern)
		result += fmt.Sprintf("Frontend Pattern: %s\n", row.FrontendPattern)
		result += fmt.Sprintf("Integration Points:\n")

		// Parse integration points from string
		integrationPoints := strings.Split(row.IntegrationPointsStr, "|||")
		for _, point := range integrationPoints {
			if point != "" {
				result += fmt.Sprintf("  - %s\n", point)
			}
		}

		result += fmt.Sprintf("Description: %s\n", row.RulesDescription)

		// Parse examples from string
		examples := strings.Split(row.ExamplesStr, "|||")
		if len(examples) > 0 && examples[0] != "" {
			result += fmt.Sprintf("Examples:\n")
			for _, ex := range examples {
				if ex != "" {
					result += fmt.Sprintf("  - %s\n", ex)
				}
			}
		}
		result += "\n"
	}

	log.Printf("✅ Found %d architecture pattern(s)", len(rows))
	return result, nil
}
