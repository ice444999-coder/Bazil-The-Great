package database

import (
	"ares_api/internal/models"
	"log"

	"gorm.io/gorm"
)

// Function to auto-migrate everything
func AutoMigrateAll(db *gorm.DB) error {

	// Note: pgvector extension must be installed manually if semantic search is needed
	// Run: CREATE EXTENSION IF NOT EXISTS vector;

	return db.AutoMigrate(
		// Add all your models here
		&models.User{},
		&models.Chat{},
		// &models.Trade{}, // SKIP - Already migrated, causes conflict
		&models.Setting{},
		&models.Ledger{},
		&models.Balance{},
		&models.MemorySnapshot{},
		// Memory embeddings and semantic search
		&models.MemoryEmbedding{},
		&models.EmbeddingQueueItem{},
		&models.MemoryRelationship{},
		&models.MemoryCacheStats{},
		// Chat persistence
		&models.ChatMessage{},
		// Fault Vault System
		&models.FaultVaultSession{},
		&models.FaultVaultAction{},
		&models.FaultVaultContext{},
		&models.FaultVaultLearning{},
		// ARES Foundation
		&models.AresConfig{},
		&models.ConversationImport{},
		&models.FileScanResult{},
		// Repository Inspection
		&models.RepoFileCache{},
		// Autonomous Trading System
		&models.SandboxTrade{},
		&models.TradingPerformance{},
		&models.MarketDataCache{},
		&models.StrategyMutation{},
		&models.RiskEvent{},
		// ACE Framework (Agentic Context Engineering)
		&models.PlaybookRule{},
		// Strategy Management
		&models.Strategy{},
		&models.StrategyVersion{},
		// System Logging
		&models.SystemLog{},
		// Tool Registry System
		&models.ToolRegistry{},
		&models.ToolPermission{},
		&models.ToolExecutionLog{},
		// SOLACE Consciousness
		&models.SolaceDecision{},
		&models.CognitivePattern{},
		// FORGE Apprenticeship
		&models.ForgeConfidenceTracker{},
		// Agent Swarm
		&models.AgentRegistry{},
		// GRPO Learning
		&models.GRPOBias{},
		&models.GRPOMetric{},
		// Glass Box Transparency
		&models.GlassBoxLog{},
		// Mission Progress Tracking
		&models.MissionProgress{},
		// Self-Healing System
		&models.BazilReward{},
	)
}

// Migrate runs all database migrations
func Migrate(db *gorm.DB) error {
	// Install pgvector extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector;").Error; err != nil {
		return err
	}

	// Drop ALL tables to start fresh (nuclear option for development)
	allTables := []string{
		"tool_permissions",
		"tool_execution_logs",
		"solace_decisions",
		"cognitive_patterns",
		"forge_confidence_trackers",
		"agent_registries",
		"grpo_biases",
		"grpo_metrics",
		"glass_box_logs",
		"tool_registries",
		// Add all other tables that might conflict
		"users",
		"chats",
		"settings",
		"ledgers",
		"balances",
		"memory_snapshots",
		"memory_embeddings",
		"embedding_queue_items",
		"memory_relationships",
		"memory_cache_stats",
		"chat_messages",
		"fault_vault_sessions",
		"fault_vault_actions",
		"fault_vault_contexts",
		"fault_vault_learnings",
		"ares_configs",
		"conversation_imports",
		"file_scan_results",
		"repo_file_caches",
		"sandbox_trades",
		"trading_performances",
		"market_data_caches",
		"strategy_mutations",
		"risk_events",
		"playbook_rules",
		"strategies",
		"strategy_versions",
		"system_logs",
	}

	// Drop all tables
	for _, table := range allTables {
		db.Migrator().DropTable(table)
	}

	// Also try raw SQL drop
	db.Exec(`
		DO $$ DECLARE
		    r RECORD;
		BEGIN
		    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
		        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
		    END LOOP;
		END $$;
	`)

	// Now create all tables fresh with proper constraints
	models := []interface{}{
		// Base models first
		&models.User{},
		&models.Chat{},
		&models.Setting{},
		&models.Ledger{},
		&models.Balance{},
		&models.MemorySnapshot{},
		// Memory system
		&models.MemoryEmbedding{},
		&models.EmbeddingQueueItem{},
		&models.MemoryRelationship{},
		&models.MemoryCacheStats{},
		// Chat system
		&models.ChatMessage{},
		// Fault vault
		&models.FaultVaultSession{},
		&models.FaultVaultAction{},
		&models.FaultVaultContext{},
		&models.FaultVaultLearning{},
		// ARES foundation
		&models.AresConfig{},
		&models.ConversationImport{},
		&models.FileScanResult{},
		// Repository inspection
		&models.RepoFileCache{},
		// Trading system
		&models.SandboxTrade{},
		&models.TradingPerformance{},
		&models.MarketDataCache{},
		&models.StrategyMutation{},
		&models.RiskEvent{},
		// ACE framework
		&models.PlaybookRule{},
		// Strategy management
		&models.Strategy{},
		&models.StrategyVersion{},
		// System logging
		&models.SystemLog{},
		// Tool registry system (in correct order - parent first)
		&models.ToolRegistry{},
		&models.ToolPermission{},
		&models.ToolExecutionLog{},
		// SOLACE consciousness
		&models.SolaceDecision{},
		&models.CognitivePattern{},
		// FORGE apprenticeship
		&models.ForgeConfidenceTracker{},
		// Agent swarm
		&models.AgentRegistry{},
		// GRPO learning
		&models.GRPOBias{},
		&models.GRPOMetric{},
		// Glass box transparency
		&models.GlassBoxLog{},
	}

	// Create each table individually
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("Warning: Failed to migrate %T: %v", model, err)
		}
	}

	log.Println("Migration completed - all tables recreated")
	return nil
}
