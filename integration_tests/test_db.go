package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5433"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "ARESISWAKING"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "ares_pgvector"
	}
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode,
	)

	fmt.Printf("Connecting to: %s\n", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to Postgres:", err)
	}

	fmt.Println("✅ Database connection successful!")

	// Test if repo_file_caches table exists
	var count int64
	err = db.Table("repo_file_caches").Count(&count).Error
	if err != nil {
		fmt.Printf("❌ Table repo_file_caches does not exist: %v\n", err)
	} else {
		fmt.Printf("✅ Table repo_file_caches exists with %d records\n", count)
	}
}