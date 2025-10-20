// Approach 3: Go-based Database Query Tool
// Build with: go run db_helper.go
// Usage: go run db_helper.go "SELECT * FROM decision_traces"

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run db_helper.go \"SQL_QUERY\"")
		fmt.Println("Example: go run db_helper.go \"SELECT COUNT(*) FROM decision_traces\"")
		os.Exit(1)
	}

	query := os.Args[1]

	// Connection string - matches ARES_API .env settings
	connStr := "host=localhost port=5433 user=postgres password=ARESISWAKING dbname=ares_pgvector sslmode=disable"

	// Connect
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Execute query
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	// Get column names
	cols, _ := rows.Columns()
	fmt.Println("Columns:", cols)

	// Print results
	for rows.Next() {
		// Create a slice of interface{} to hold each column
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Fatal(err)
		}

		for i, col := range cols {
			val := values[i]
			fmt.Printf("%s: %v | ", col, val)
		}
		fmt.Println()
	}
}
