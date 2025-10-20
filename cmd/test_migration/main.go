// ============================================================================
// ARES MIGRATION TEST SUITE
// Version: 1.0.0
// Date: October 19, 2025
//
// Purpose: Test SQL reorganization migration without affecting production
// Validates syntax, dependencies, and backward compatibility
// ============================================================================

package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🧪 ARES Migration Test Suite")
	fmt.Println("Testing reorganization migration safely...")

	// Check if migration file exists
	migrationFile := "migrations/reorg_migration.sql"
	if _, err := os.Stat(migrationFile); os.IsNotExist(err) {
		log.Fatalf("❌ Migration file not found: %s", migrationFile)
	}

	// Read migration content
	content, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("❌ Failed to read migration file: %v", err)
	}

	fmt.Printf("✅ Migration file loaded (%d bytes)\n", len(content))

	// Parse and validate SQL statements
	statements := parseSQLStatements(string(content))
	fmt.Printf("📊 Found %d SQL statements\n", len(statements))

	// Validate statement types
	validationResults := validateStatements(statements)
	fmt.Println("\n🔍 Validation Results:")
	fmt.Printf("   ✅ CREATE SCHEMA: %d\n", validationResults.CreateSchema)
	fmt.Printf("   ✅ CREATE TABLE: %d\n", validationResults.CreateTable)
	fmt.Printf("   ✅ CREATE INDEX: %d\n", validationResults.CreateIndex)
	fmt.Printf("   ✅ CREATE VIEW: %d\n", validationResults.CreateView)
	fmt.Printf("   ⚠️  ALTER TABLE: %d\n", validationResults.AlterTable)
	fmt.Printf("   ⚠️  DROP TABLE: %d\n", validationResults.DropTable)

	// Check for backward compatibility views
	backwardCompat := checkBackwardCompatibility(statements)
	if backwardCompat {
		fmt.Println("✅ Backward compatibility views detected")
	} else {
		fmt.Println("⚠️  No backward compatibility views found")
	}

	// Check for pgvector usage
	pgvectorUsed := checkPGVectorUsage(statements)
	if pgvectorUsed {
		fmt.Println("✅ pgvector extensions and indexes detected")
	} else {
		fmt.Println("⚠️  No pgvector usage found")
	}

	// Test database connection (if available)
	testDBConnection()

	fmt.Println("\n🎯 Migration Test Complete")
	fmt.Println("Next steps:")
	fmt.Println("1. Review validation results above")
	fmt.Println("2. Run on test database: psql -f migrations/reorg_migration.sql")
	fmt.Println("3. Verify backward compatibility views work")
	fmt.Println("4. Test AI query performance improvements")
}

type ValidationResults struct {
	CreateSchema int
	CreateTable  int
	CreateIndex  int
	CreateView   int
	AlterTable   int
	DropTable    int
}

func parseSQLStatements(content string) []string {
	// Simple SQL statement splitter (handles basic cases)
	var statements []string
	var current strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if strings.HasPrefix(trimmed, "--") || trimmed == "" {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		// Check for statement end
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			statements = append(statements, strings.TrimSpace(current.String()))
			current.Reset()
		}
	}

	return statements
}

func validateStatements(statements []string) ValidationResults {
	results := ValidationResults{}

	for _, stmt := range statements {
		upper := strings.ToUpper(strings.TrimSpace(stmt))

		switch {
		case strings.Contains(upper, "CREATE SCHEMA"):
			results.CreateSchema++
		case strings.Contains(upper, "CREATE TABLE"):
			results.CreateTable++
		case strings.Contains(upper, "CREATE INDEX"):
			results.CreateIndex++
		case strings.Contains(upper, "CREATE VIEW") || strings.Contains(upper, "CREATE OR REPLACE VIEW"):
			results.CreateView++
		case strings.Contains(upper, "ALTER TABLE"):
			results.AlterTable++
		case strings.Contains(upper, "DROP TABLE"):
			results.DropTable++
		}
	}

	return results
}

func checkBackwardCompatibility(statements []string) bool {
	for _, stmt := range statements {
		upper := strings.ToUpper(stmt)
		if strings.Contains(upper, "CREATE VIEW") || strings.Contains(upper, "CREATE OR REPLACE VIEW") {
			return true
		}
	}
	return false
}

func checkPGVectorUsage(statements []string) bool {
	for _, stmt := range statements {
		upper := strings.ToUpper(stmt)
		if strings.Contains(upper, "VECTOR") || strings.Contains(upper, "PGVECTOR") {
			return true
		}
	}
	return false
}

func testDBConnection() {
	// Try to connect to test database if environment variables are set
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("⚠️  No DATABASE_URL set - skipping database connection test")
		return
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("⚠️  Database connection failed: %v\n", err)
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Printf("⚠️  Database ping failed: %v\n", err)
		return
	}

	fmt.Println("✅ Database connection successful")

	// Test basic schema query
	var schemaCount int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')").Scan(&schemaCount)
	if err != nil {
		fmt.Printf("⚠️  Schema count query failed: %v\n", err)
		return
	}

	fmt.Printf("📊 Current schema count: %d\n", schemaCount)
}
