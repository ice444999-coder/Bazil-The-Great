package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := flag.String("dsn", "host=localhost user=postgres password=ARESISWAKING dbname=ares_pgvector port=5433 sslmode=disable", "PostgreSQL DSN")
	mode := flag.String("mode", "cleanup", "Mode: cleanup (delete invalid rows), drop (drop tables), nuclear (drop and recreate DB), migrate (run GORM AutoMigrate)")
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
	default:
		log.Fatal("Invalid mode. Use cleanup, drop, nuclear, or migrate.")
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
	// Extract dbname from DSN
	dbName := "ares_pgvector" // Hardcode or parse
	cmd := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-p", "5433", "-d", "postgres", "-c", fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName))
	if err := cmd.Run(); err != nil {
		log.Println("Drop DB failed: ", err)
	}
	cmd = exec.Command("psql", "-U", "postgres", "-h", "localhost", "-p", "5433", "-d", "postgres", "-c", fmt.Sprintf("CREATE DATABASE %s;", dbName))
	if err := cmd.Run(); err != nil {
		log.Fatal("Create DB failed: ", err)
	}
	log.Println("Database dropped and recreated.")
	// Reconnect and install pgvector if needed
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Reconnect failed: ", err)
	}
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgvector;").Error; err != nil {
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

// Add your AutoMigrateAll function here or import from migration.go
func AutoMigrateAll(db *gorm.DB) error {
	// Migrate all models, including tool tables
	return db.AutoMigrate( /* all models */ )
}
