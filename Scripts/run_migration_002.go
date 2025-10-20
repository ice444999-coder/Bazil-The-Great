package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Database connection string
	dsn := "host=localhost user=postgres password=ARESISWAKING dbname=ares_pgvector port=5433 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	log.Println("✓ Connected to database")

	// Read migration file
	migrationSQL, err := os.ReadFile("../migrations/002_add_strategy_name_to_sandbox_trades.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	log.Println("✓ Migration file loaded")

	// Execute migration
	if err := db.Exec(string(migrationSQL)).Error; err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	log.Println("✓ Migration executed successfully")

	// Verify the column exists
	var columnExists bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='sandbox_trades' 
			AND column_name='strategy_name'
		);
	`
	if err := db.Raw(query).Scan(&columnExists).Error; err != nil {
		log.Fatalf("Failed to verify migration: %v", err)
	}

	if columnExists {
		log.Println("✓ Column 'strategy_name' verified in sandbox_trades table")
	} else {
		log.Fatal("✗ Column 'strategy_name' not found after migration")
	}

	// Verify indexes
	indexQuery := `
		SELECT indexname 
		FROM pg_indexes 
		WHERE tablename = 'sandbox_trades' 
		AND indexname LIKE 'idx_sandbox_strategy%';
	`
	var indexes []string
	if err := db.Raw(indexQuery).Scan(&indexes).Error; err != nil {
		log.Fatalf("Failed to verify indexes: %v", err)
	}

	log.Printf("✓ Found %d strategy-related indexes", len(indexes))
	for _, idx := range indexes {
		log.Printf("  - %s", idx)
	}

	fmt.Println("\n🎉 Migration 002 completed successfully!")
	fmt.Println("   - Added strategy_name column to sandbox_trades")
	fmt.Println("   - Created performance indexes")
	fmt.Println("   - Ready for multi-strategy system")
}
