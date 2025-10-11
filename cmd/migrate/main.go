package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	// Build PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	fmt.Println("🔄 Running semantic memory architecture migration...")

	// Execute migrations step by step
	migrations := []string{
		// 1. Enable pgvector (skip if not installed - use TEXT fallback)
		// "CREATE EXTENSION IF NOT EXISTS vector",

		// 2. Add new columns to memory_snapshots
		"ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS importance_score FLOAT DEFAULT 0.5",
		"ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS access_count INTEGER DEFAULT 0",
		"ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS last_accessed TIMESTAMP",
		"ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS memory_type VARCHAR(50) DEFAULT 'general'",
		"ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS tags TEXT[]",
		"ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS compression_level VARCHAR(20) DEFAULT 'none'",
		"ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS archived BOOLEAN DEFAULT FALSE",

		// 3. Create indices
		"CREATE INDEX IF NOT EXISTS idx_memory_importance ON memory_snapshots(importance_score DESC)",
		"CREATE INDEX IF NOT EXISTS idx_memory_access_count ON memory_snapshots(access_count DESC)",
		"CREATE INDEX IF NOT EXISTS idx_memory_last_accessed ON memory_snapshots(last_accessed DESC)",
		"CREATE INDEX IF NOT EXISTS idx_memory_type ON memory_snapshots(memory_type)",
		"CREATE INDEX IF NOT EXISTS idx_memory_archived ON memory_snapshots(archived)",
		"CREATE INDEX IF NOT EXISTS idx_memory_tags ON memory_snapshots USING GIN(tags)",

		// 4. Create memory_embeddings table (using TEXT for now, will migrate to pgvector later)
		`CREATE TABLE IF NOT EXISTS memory_embeddings (
			id SERIAL PRIMARY KEY,
			snapshot_id INTEGER REFERENCES memory_snapshots(id) ON DELETE CASCADE,
			embedding TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		"CREATE INDEX IF NOT EXISTS idx_memory_embeddings_snapshot_id ON memory_embeddings(snapshot_id)",

		// 5. Create embedding queue table
		`CREATE TABLE IF NOT EXISTS embedding_generation_queue (
			id SERIAL PRIMARY KEY,
			snapshot_id INTEGER REFERENCES memory_snapshots(id) ON DELETE CASCADE,
			status VARCHAR(20) DEFAULT 'pending',
			retry_count INTEGER DEFAULT 0,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			processed_at TIMESTAMP
		)`,

		"CREATE INDEX IF NOT EXISTS idx_embedding_queue_status ON embedding_generation_queue(status)",
		"CREATE INDEX IF NOT EXISTS idx_embedding_queue_created ON embedding_generation_queue(created_at)",

		// 6. Create memory_relationships table
		`CREATE TABLE IF NOT EXISTS memory_relationships (
			id SERIAL PRIMARY KEY,
			source_snapshot_id INTEGER REFERENCES memory_snapshots(id) ON DELETE CASCADE,
			target_snapshot_id INTEGER REFERENCES memory_snapshots(id) ON DELETE CASCADE,
			relationship_type VARCHAR(50),
			strength FLOAT DEFAULT 1.0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source_snapshot_id, target_snapshot_id, relationship_type)
		)`,

		"CREATE INDEX IF NOT EXISTS idx_memory_rel_source ON memory_relationships(source_snapshot_id)",
		"CREATE INDEX IF NOT EXISTS idx_memory_rel_target ON memory_relationships(target_snapshot_id)",
		"CREATE INDEX IF NOT EXISTS idx_memory_rel_type ON memory_relationships(relationship_type)",

		// 7. Create memory_cache_stats table
		`CREATE TABLE IF NOT EXISTS memory_cache_stats (
			snapshot_id INTEGER PRIMARY KEY REFERENCES memory_snapshots(id) ON DELETE CASCADE,
			temperature VARCHAR(10) DEFAULT 'cold',
			cache_hits INTEGER DEFAULT 0,
			cache_misses INTEGER DEFAULT 0,
			last_hit TIMESTAMP,
			promoted_at TIMESTAMP,
			demoted_at TIMESTAMP,
			size_bytes INTEGER,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		"CREATE INDEX IF NOT EXISTS idx_cache_temperature ON memory_cache_stats(temperature)",
		"CREATE INDEX IF NOT EXISTS idx_cache_last_hit ON memory_cache_stats(last_hit DESC)",

		// 8. Set default importance for existing memories
		"UPDATE memory_snapshots SET importance_score = 0.5 WHERE importance_score IS NULL",

		// 9. Queue existing memories for embedding
		`INSERT INTO embedding_generation_queue (snapshot_id, status)
		SELECT id, 'pending'
		FROM memory_snapshots
		WHERE NOT EXISTS (
			SELECT 1 FROM embedding_generation_queue WHERE snapshot_id = memory_snapshots.id
		)`,
	}

	for i, stmt := range migrations {
		fmt.Printf("  [%d/%d] Executing migration step...\n", i+1, len(migrations))
		if err := db.Exec(stmt).Error; err != nil {
			// If error is about existing object, continue
			if contains(err.Error(), "already exists") || contains(err.Error(), "duplicate") {
				fmt.Printf("  ⚠️  Skipping (already exists)\n")
				continue
			}
			log.Fatalf("Migration step %d failed: %v\nStatement: %s", i+1, err, stmt)
		}
	}

	fmt.Println("\n✅ Semantic memory migration completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Pull embedding model: ollama pull nomic-embed-text")
	fmt.Println("2. Start ARES API: go run cmd/main.go")
	fmt.Println("3. Process embeddings: POST /api/v1/claude/process-embeddings")
}
