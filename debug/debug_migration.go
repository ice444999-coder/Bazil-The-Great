package main

import (
	"flag"
	"fmt"
	"log"

	"ares_api/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := flag.String("dsn", "host=localhost user=postgres password=ARESISWAKING dbname=ares_pgvector port=5433 sslmode=disable", "PostgreSQL DSN")
	mode := flag.String("mode", "cleanup", "Mode: cleanup (delete invalid rows), drop (drop tables), nuclear (drop and recreate DB), migrate (run GORM AutoMigrate), create-missing (create missing tables)")
	flag.Parse()

	db, err := gorm.Open(postgres.Open(*dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	switch *mode {
	case "cleanup":
		cleanupInvalidRows(db)
	case "drop":
		dropProblematicTables(db)
	case "nuclear":
		nuclearReset(*dsn)
	case "migrate":
		runMigration(db)
	case "create-missing":
		createMissingTables(db)
	default:
		log.Fatal("Invalid mode. Use cleanup, drop, nuclear, migrate, or create-missing.")
	}
	log.Println("Operation completed successfully.")
}

func cleanupInvalidRows(db *gorm.DB) {
	// Delete rows in tool_permissions with invalid tool_id
	if err := db.Exec(`
		DELETE FROM tool_permissions
		WHERE tool_id NOT IN (SELECT tool_id FROM tools)
	`).Error; err != nil {
		log.Fatal("Cleanup failed: ", err)
	}
	log.Println("Invalid rows deleted from tool_permissions.")
}

func dropProblematicTables(db *gorm.DB) {
	tables := []string{"tool_permissions", "tool_permission_requests", "tool_execution_log", "tools"} // Add more if needed
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Printf("Failed to drop %s: %v", table, err)
		} else {
			log.Printf("Dropped table %s", table)
		}
	}
}

func nuclearReset(dsn string) {
	// First connect to postgres database to drop/create ares_pgvector
	postgresDSN := "host=localhost user=postgres password=ARESISWAKING dbname=postgres port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to postgres db: ", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Terminate all connections to the database
	if err := db.Exec(`
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = 'ares_pgvector' AND pid <> pg_backend_pid();
	`).Error; err != nil {
		log.Println("Terminate connections failed: ", err)
	}

	// Drop the database
	if err := db.Exec("DROP DATABASE IF EXISTS ares_pgvector;").Error; err != nil {
		log.Println("Drop DB failed: ", err)
	}
	// Create the database
	if err := db.Exec("CREATE DATABASE ares_pgvector;").Error; err != nil {
		log.Fatal("Create DB failed: ", err)
	}
	log.Println("Database dropped and recreated.")

	// Now connect to the new database and install pgvector
	newDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Reconnect failed: ", err)
	}
	defer func() {
		sqlDB, _ := newDB.DB()
		sqlDB.Close()
	}()
	if err := newDB.Exec("CREATE EXTENSION IF NOT EXISTS vector;").Error; err != nil {
		log.Println("pgvector install failed: ", err)
	}
}

func runMigration(db *gorm.DB) {
	// Call your AutoMigrateAll from migration.go
	if err := AutoMigrateAll(db); err != nil {
		log.Fatal("Migration failed: ", err)
	}
	log.Println("Migration completed.")
}

func createMissingTables(db *gorm.DB) {
	// Create missing tables that aren't handled by AutoMigrate
	tables := []string{
		`CREATE TABLE IF NOT EXISTS service_config (
			id SERIAL PRIMARY KEY,
			service_name TEXT NOT NULL,
			config_key TEXT NOT NULL,
			config_value JSONB,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(service_name, config_key)
		)`,
		`CREATE TABLE IF NOT EXISTS service_registry (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			version TEXT,
			status TEXT DEFAULT 'offline',
			port INTEGER,
			health_url TEXT,
			last_heartbeat TIMESTAMP DEFAULT NOW(),
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS service_metrics (
			id SERIAL PRIMARY KEY,
			service_name TEXT NOT NULL,
			metric_name TEXT NOT NULL,
			metric_type TEXT NOT NULL,
			metric_value FLOAT,
			labels JSONB,
			timestamp TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS trades (
			id SERIAL PRIMARY KEY,
			user_id INTEGER,
			symbol TEXT NOT NULL,
			side TEXT NOT NULL,
			type TEXT NOT NULL,
			quantity DECIMAL(20,8),
			price DECIMAL(20,8),
			status TEXT DEFAULT 'pending',
			order_id TEXT,
			exchange_order_id TEXT,
			executed_quantity DECIMAL(20,8) DEFAULT 0,
			executed_price DECIMAL(20,8),
			fees DECIMAL(20,8) DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			deleted_at TIMESTAMP
		)`,
	}

	for _, sql := range tables {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Failed to create table: %v", err)
		} else {
			log.Println("Created missing table successfully")
		}
	}
}

// Add your AutoMigrateAll function here or import from migration.go
func AutoMigrateAll(db *gorm.DB) error {
	// Migrate all models, including tool tables
	return db.AutoMigrate(
		&models.User{},
		&models.Chat{},
		&models.Setting{},
		&models.Ledger{},
		&models.Balance{},
		&models.MemorySnapshot{},
		&models.MemoryEmbedding{},
		&models.EmbeddingQueueItem{},
		&models.MemoryRelationship{},
		&models.MemoryCacheStats{},
		&models.ChatMessage{},
		&models.FaultVaultSession{},
		&models.FaultVaultAction{},
		&models.FaultVaultContext{},
		&models.FaultVaultLearning{},
		&models.AresConfig{},
		&models.ConversationImport{},
		&models.FileScanResult{},
		&models.RepoFileCache{},
		&models.SandboxTrade{},
		&models.TradingPerformance{},
		&models.MarketDataCache{},
		&models.StrategyMutation{},
		&models.RiskEvent{},
		&models.PlaybookRule{},
		&models.Strategy{},
		&models.StrategyVersion{},
		&models.SystemLog{},
		&models.ToolRegistry{},
		&models.ToolPermission{},
		&models.ToolExecutionLog{},
		&models.SolaceDecision{},
		&models.CognitivePattern{},
		&models.ForgeConfidenceTracker{},
		&models.AgentRegistry{},
		&models.GRPOBias{},
		&models.GRPOMetric{},
		&models.GlassBoxLog{},
	)
}
