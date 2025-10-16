package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestActivityLog tracks all testing actions
type TestActivityLog struct {
	LogID         uint   `gorm:"primaryKey;autoIncrement"`
	Actor         string `gorm:"size:50;not null"` // 'claude-sonnet-4.5' or 'SOLACE'
	ActionType    string `gorm:"size:100;not null"`
	FeatureTested string `gorm:"size:100"`
	ActionDetails string `gorm:"type:text"`
	Result        string `gorm:"size:20"`
	ResponseData  string `gorm:"type:text"`
	ErrorMessage  string `gorm:"type:text"`
	SessionHash   string `gorm:"size:64"`
	Timestamp     int64  `gorm:"autoCreateTime"`
}

func main() {
	// Load .env
	godotenv.Load(".env")

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	log.Println("✅ Connected to database")

	// Create test_activity_log table
	err = db.AutoMigrate(&TestActivityLog{})
	if err != nil {
		log.Fatalf("❌ Failed to create test_activity_log table: %v", err)
	}

	log.Println("✅ test_activity_log table created successfully!")

	// Insert initial log entry
	initialLog := TestActivityLog{
		Actor:         "claude-sonnet-4.5",
		ActionType:    "session_start",
		FeatureTested: "test_framework",
		ActionDetails: `{"message": "Interactive testing session initialized", "mode": "visual_testing_with_sql_logging"}`,
		Result:        "success",
		SessionHash:   fmt.Sprintf("%d", int64(1729000000)),
	}

	err = db.Create(&initialLog).Error
	if err != nil {
		log.Fatalf("❌ Failed to create initial log: %v", err)
	}

	log.Printf("✅ Initial log created - Session Hash: %s, Timestamp: %d", initialLog.SessionHash, initialLog.Timestamp)
	log.Println("🎯 Test activity logging system ready!")
}
