/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
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

	// Create pgvector extension first
	fmt.Println("🔧 Creating pgvector extension...")
	_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS vector;")
	if err != nil {
		log.Printf("Warning: Could not create vector extension: %v", err)
		log.Println("Continuing with migration anyway...")
	} else {
		fmt.Println("✅ pgvector extension created/enabled")
	}

	// Read migration file
	migrationSQL, err := ioutil.ReadFile("benchmark_migration.sql")
	if err != nil {
		log.Fatal("Failed to read migration file:", err)
	}

	fmt.Println("📄 Read migration file, applying...")

	// Execute migration in transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to start transaction:", err)
	}

	_, err = tx.Exec(string(migrationSQL))
	if err != nil {
		tx.Rollback()
		log.Fatal("Failed to execute migration:", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("✅ Migration applied successfully!")
	fmt.Println("🎯 Database reorganized with 13 functional schemas and pgvector indexes")
}
