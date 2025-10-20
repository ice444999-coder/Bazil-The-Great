package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection parameters
	connStr := "host=localhost port=5433 user=postgres password=ARESISWAKING dbname=ares_pgvector sslmode=disable"

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("✅ Connected to database successfully")

	// Read migration file
	migrationSQL, err := ioutil.ReadFile("migrations/reorg_migration.sql")
	if err != nil {
		log.Fatal("Failed to read migration file:", err)
	}

	fmt.Println("📄 Read migration file, applying...")

	// Execute migration
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		log.Fatal("Failed to execute migration:", err)
	}

	fmt.Println("✅ Migration applied successfully!")
	fmt.Println("🎯 Database reorganized with 13 functional schemas and pgvector indexes")
}
