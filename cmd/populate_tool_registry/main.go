package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ares_api/internal/agent"
)

func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5433"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "ARESISWAKING"),
		getEnv("DB_NAME", "ares_pgvector"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🚀 ARES Tool Registry Population")
	fmt.Println("==================================\n")

	// Define 29 ARES tools
	tools := []struct {
		Name        string
		Category    string
		Description string
		Params      string
		Risk        string
		Cost        float64
	}{
		{"execute_sql_query", "database", "Execute SELECT queries on PostgreSQL. Returns results as JSON from 91 tables.", `{"query":"string"}`, "moderate", 0.0},
		{"query_memory_crystals", "database", "Retrieve memory snapshots with semantic filtering. Returns memories with embeddings.", `{"category":"string","limit":"int"}`, "safe", 0.0},
		{"create_memory_crystal", "database", "Store new memory with automatic embedding generation for future retrieval.", `{"category":"string","content":"string"}`, "safe", 0.001},
		{"execute_trade", "trading", "Execute buy/sell trade in sandbox. Supports market and limit orders.", `{"symbol":"string","side":"buy|sell","amount":"float"}`, "dangerous", 0.0},
		{"close_position", "trading", "Close open position. Calculates P&L and updates ACE playbook.", `{"trade_id":"uuid"}`, "moderate", 0.0},
		{"get_portfolio_status", "trading", "Get portfolio status including open positions, P&L, daily performance.", `{}`, "safe", 0.0},
		{"get_live_prices", "trading", "Fetch real-time crypto prices from CoinGecko API for 100+ coins.", `{"symbols":"array"}`, "safe", 0.0},
		{"read_file", "filesystem", "Read file contents from ARES workspace. Supports .md, .go, .html, .sql, .ps1, .json.", `{"path":"string"}`, "safe", 0.0},
		{"write_file", "filesystem", "Write content to workspace file. Creates directories if needed. Requires permission.", `{"path":"string","content":"string"}`, "dangerous", 0.0},
		{"list_files", "filesystem", "List files/directories in workspace path. Supports extension filtering.", `{"path":"string","extension":"string"}`, "safe", 0.0},
		{"delete_file", "filesystem", "Delete file from workspace. Requires explicit permission. Cannot delete critical system files.", `{"path":"string"}`, "dangerous", 0.0},
		{"github_copilot_suggest", "codegen", "Generate code suggestions using GitHub Copilot CLI. FORGE learns from this.", `{"prompt":"string"}`, "safe", 0.0},
		{"github_copilot_explain", "codegen", "Explain code using GitHub Copilot CLI. FORGE uses for understanding patterns.", `{"code":"string"}`, "safe", 0.0},
		{"semantic_memory_search", "memory", "Search memories using pgvector semantic similarity. Finds by meaning, not keywords.", `{"query":"string","limit":"int"}`, "safe", 0.001},
		{"store_conversation", "memory", "Store conversation to conversation_log. Auto-generates summary if long.", `{"user_message":"string","agent_response":"string"}`, "safe", 0.0},
		{"analyze_trade_outcome", "ace", "Analyze completed trade and extract cognitive patterns for ACE playbook.", `{"trade_id":"uuid"}`, "safe", 0.002},
		{"apply_playbook_rules", "ace", "Apply ACE playbook rules to market situation. Returns recommended actions.", `{"market_conditions":"object"}`, "safe", 0.0},
		{"prune_weak_rules", "ace", "Remove ACE playbook rules with low confidence (<40%) or low usage.", `{"confidence_threshold":"float"}`, "moderate", 0.0},
		{"record_forge_observation", "forge", "Record FORGE observation of GitHub Copilot pattern for learning.", `{"pattern_name":"string","code_sample":"string","success":"bool"}`, "safe", 0.0},
		{"check_forge_graduation", "forge", "Check if FORGE met graduation criteria (70% confidence, 30+ observations).", `{"pattern_name":"string"}`, "safe", 0.0},
		{"get_system_health", "monitoring", "Get comprehensive health report: PostgreSQL, Ollama, memory, goroutines.", `{}`, "safe", 0.0},
		{"query_masterplan", "monitoring", "Query ARES masterplan for architecture rules or system state.", `{"query":"string"}`, "safe", 0.0},
		{"ollama_chat_completion", "llm", "Send chat request to Ollama (deepseek-r1:14b) for reasoning/learning/analysis.", `{"messages":"array","temperature":"float"}`, "safe", 0.0},
		{"generate_embedding", "llm", "Generate 1536-dim embedding using OpenAI text-embedding-3-small.", `{"text":"string"}`, "safe", 0.0001},
		{"log_glass_box_action", "transparency", "Log action with SHA-256 hash and Merkle proof for tamper-proof audit trail.", `{"actor":"string","action_type":"string","details":"object"}`, "safe", 0.0},
		{"dedup_sql_files", "database", "Deduplicate SQL files using SHA-256 hashing and semantic similarity detection. Supports dry-run mode.", `{"directory":"string","dry_run":"bool","output_format":"json|markdown"}`, "moderate", 0.001},
		{"build_schema_map", "database", "Generate comprehensive schema map with ER diagrams, dependency graphs, and optimization recommendations.", `{"include_er_diagram":"bool","include_dependencies":"bool"}`, "safe", 0.002},
		{"analyze_query_performance", "database", "Analyze query performance patterns and identify slow queries requiring optimization.", `{}`, "safe", 0.001},
	}

	fmt.Printf("📦 Registering %d tools...\n\n", len(tools))

	successCount := 0
	for i, tool := range tools {
		fmt.Printf("[%d/%d] %s (%s)\n", i+1, len(tools), tool.Name, tool.Category)

		// Generate embedding
		embeddingText := tool.Name + ": " + tool.Description
		embedding, err := agent.GenerateEmbedding(embeddingText)
		if err != nil {
			log.Printf("  ❌ Failed to generate embedding: %v", err)
			continue
		}

		// Convert to pgvector format
		embeddingStr := vectorToString(embedding)

		// Insert tool
		toolID := uuid.New()
		insertSQL := `
			INSERT INTO tool_registry (
				tool_id, tool_name, tool_category, description, 
				required_params, risk_level, embedding, api_cost_per_call,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5::jsonb, $6, $7::vector, $8, NOW(), NOW()
			)
			ON CONFLICT (tool_name) DO UPDATE SET
				description = EXCLUDED.description,
				embedding = EXCLUDED.embedding,
				updated_at = NOW()
		`

		if err := db.Exec(insertSQL, toolID, tool.Name, tool.Category, tool.Description,
			tool.Params, tool.Risk, embeddingStr, tool.Cost).Error; err != nil {
			log.Printf("  ❌ Insert failed: %v", err)
			continue
		}

		// Grant SOLACE unlimited permissions
		permID := uuid.New()
		permSQL := `
			INSERT INTO tool_permissions (
				permission_id, tool_id, agent_name, access_granted,
				persistent_approval, approved_by, approved_at,
				daily_usage_limit, hourly_usage_limit, daily_cost_limit,
				circuit_breaker_threshold, circuit_breaker_active,
				last_usage_reset
			) VALUES (
				$1, $2, 'SOLACE', true,
				true, 'SYSTEM', NOW(),
				999999, 999999, 999999.99,
				999999, false,
				NOW()
			)
			ON CONFLICT (tool_id, agent_name) DO UPDATE SET
				access_granted = true,
				daily_usage_limit = 999999,
				updated_at = NOW()
		`

		if err := db.Exec(permSQL, permID, toolID).Error; err != nil {
			log.Printf("  ⚠️  Permission grant failed: %v", err)
		} else {
			fmt.Printf("  ✅ Registered + SOLACE granted\n")
			successCount++
		}
	}

	fmt.Printf("\n==================================\n")
	fmt.Printf("✅ Success: %d/%d tools\n", successCount, len(tools))
	fmt.Printf("\n🎯 Next Steps:\n")
	fmt.Println("1. Test: curl 'http://localhost:8080/api/v1/tools/search?intent=I+want+to+trade'")
	fmt.Println("2. Build API handlers (internal/api/handlers/tool_handler.go)")
	fmt.Println("3. Register routes (internal/api/routes/v1.go)")
	fmt.Println("4. Deploy to production!")
}

func vectorToString(embedding []float32) string {
	if len(embedding) == 0 {
		return "[]"
	}
	result := "["
	for i, val := range embedding {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%f", val)
	}
	result += "]"
	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
