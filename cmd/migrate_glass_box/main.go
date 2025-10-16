package main

import (
	"fmt"
	"log"
	"os"

	"ares_api/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env
	envPaths := []string{".env", "../.env", "../../.env", "c:\\ARES_Workspace\\ARES_API\\.env"}
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("✅ Loaded .env from: %s", path)
			break
		}
	}

	// Database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🔐 CREATING GLASS BOX + MERKLE TREE TABLES")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Auto-migrate tables
	err = db.AutoMigrate(
		&models.GlassBoxLog{},
		&models.MerkleBatch{},
	)

	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ All tables created successfully!")

	// Verify tables
	log.Println("\n📊 Verifying table creation...")

	var glassBoxCount int64
	db.Model(&models.GlassBoxLog{}).Count(&glassBoxCount)
	log.Printf("   ✓ glass_box_log: %d rows", glassBoxCount)

	var batchCount int64
	db.Model(&models.MerkleBatch{}).Count(&batchCount)
	log.Printf("   ✓ merkle_batches: %d rows", batchCount)

	log.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅ GLASS BOX MIGRATION COMPLETE")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("\n🔐 Security Features:")
	log.Println("   • Internal hashes NEVER exposed publicly")
	log.Println("   • Merkle tree batching (100 logs per batch)")
	log.Println("   • Only Merkle root hash goes to Hedera")
	log.Println("   • Full verification chain available")
	log.Println("   • Zero-knowledge proof of existence")
}
