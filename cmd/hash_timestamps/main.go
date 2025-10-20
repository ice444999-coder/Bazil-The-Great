package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// FileTimestampLedger model
type FileTimestampLedger struct {
	ID               uint      `gorm:"primaryKey"`
	FilePath         string    `gorm:"type:text;not null"`
	LastModified     time.Time `gorm:"not null"`
	FileSize         int64     `gorm:"not null"`
	LineCount        int       `gorm:"not null"`
	SHA256Hash       string    `gorm:"type:varchar(64);not null;index"`
	TimestampHash    string    `gorm:"type:varchar(64);not null;unique"`
	PreviousHash     string    `gorm:"type:varchar(64)"`
	ChainIndex       int       `gorm:"not null"`
	Verified         bool      `gorm:"default:false"`
	AnchoredToLedger bool      `gorm:"default:false"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
}

// FileManifest from JSON
type FileManifest struct {
	Path         string    `json:"Path"`
	LastModified time.Time `json:"LastModified"`
	SizeBytes    int64     `json:"SizeBytes"`
	LineCount    int       `json:"LineCount"`
	SHA256Hash   string    `json:"SHA256Hash"`
}

// ManifestWrapper wraps the Files array
type ManifestWrapper struct {
	Files []FileManifest `json:"Files"`
}

func main() {
	// Load environment
	godotenv.Load()

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create table if not exists
	if err := db.AutoMigrate(&FileTimestampLedger{}); err != nil {
		log.Fatalf("Failed to migrate table: %v", err)
	}

	// Read manifest
	data, err := os.ReadFile("filesystem_manifest.json")
	if err != nil {
		log.Fatalf("Failed to read manifest: %v", err)
	}

	var wrapper ManifestWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		log.Fatalf("Failed to parse manifest: %v", err)
	}

	manifest := wrapper.Files

	// Get last chain index
	var lastEntry FileTimestampLedger
	db.Order("chain_index DESC").First(&lastEntry)
	chainIndex := lastEntry.ChainIndex

	// Hash each file timestamp into ledger
	for _, file := range manifest {
		chainIndex++

		// Create timestamp hash: SHA256(filepath + last_modified + file_size + previous_hash)
		timestampData := fmt.Sprintf("%s|%d|%d|%s",
			file.Path,
			file.LastModified.Unix(),
			file.SizeBytes,
			lastEntry.TimestampHash,
		)

		hash := sha256.Sum256([]byte(timestampData))
		timestampHash := hex.EncodeToString(hash[:])

		entry := FileTimestampLedger{
			FilePath:         file.Path,
			LastModified:     file.LastModified,
			FileSize:         file.SizeBytes,
			LineCount:        file.LineCount,
			SHA256Hash:       file.SHA256Hash,
			TimestampHash:    timestampHash,
			PreviousHash:     lastEntry.TimestampHash,
			ChainIndex:       chainIndex,
			Verified:         false,
			AnchoredToLedger: false,
		}

		if err := db.Create(&entry).Error; err != nil {
			log.Printf("Failed to insert %s: %v", file.Path, err)
			continue
		}

		log.Printf("✅ Hashed: %s (chain index %d)", file.Path, chainIndex)
		lastEntry = entry
	}

	log.Printf("\n🎉 Successfully hashed %d file timestamps into ledger!", len(manifest))
	log.Printf("Latest chain index: %d", chainIndex)
	log.Printf("Latest hash: %s", lastEntry.TimestampHash)
}
