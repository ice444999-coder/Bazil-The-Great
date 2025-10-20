package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
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

	// Query sandbox_trades schema
	type ColumnInfo struct {
		ColumnName string
		DataType   string
		IsNullable string
	}

	var columns []ColumnInfo
	query := `
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'sandbox_trades' 
		ORDER BY ordinal_position;
	`
	if err := db.Raw(query).Scan(&columns).Error; err != nil {
		log.Fatalf("Failed to query schema: %v", err)
	}

	fmt.Println("\n📋 sandbox_trades table schema:")
	fmt.Println("=====================================")
	for _, col := range columns {
		nullable := "NOT NULL"
		if col.IsNullable == "YES" {
			nullable = "NULL"
		}
		fmt.Printf("%-25s %-20s %s\n", col.ColumnName, col.DataType, nullable)
	}

	// Query sample data
	var count int64
	db.Raw("SELECT COUNT(*) FROM sandbox_trades").Scan(&count)
	fmt.Printf("\n📊 Total sandbox trades: %d\n", count)

	// Query strategies if any exist
	var strategies []string
	db.Raw("SELECT DISTINCT strategy_name FROM sandbox_trades WHERE strategy_name IS NOT NULL ORDER BY strategy_name").Scan(&strategies)

	if len(strategies) > 0 {
		fmt.Printf("\n🎯 Active strategies: %v\n", strategies)
	} else {
		fmt.Println("\n🎯 No strategies assigned yet (ready for multi-strategy system)")
	}
}
