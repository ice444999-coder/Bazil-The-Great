package database

import (
	"ares_api/internal/models"

	"gorm.io/gorm"
)

// Function to auto-migrate everything
func AutoMigrateAll(db *gorm.DB) error {

	// Note: pgvector extension must be installed manually if semantic search is needed
	// Run: CREATE EXTENSION IF NOT EXISTS vector;

	return db.AutoMigrate(
	// Add all your models here
	 &models.User{},
	 &models.Chat{},
	 &models.Trade{},
	 &models.Setting{},
	 &models.Ledger{},
	 &models.Balance{},
	 &models.MemorySnapshot{},
	 // Memory embeddings and semantic search
	 &models.MemoryEmbedding{},
	 &models.EmbeddingQueueItem{},
	 &models.MemoryRelationship{},
	 &models.MemoryCacheStats{},
	 // Chat persistence
	 &models.ChatMessage{},
	 // Fault Vault System
	 &models.FaultVaultSession{},
	 &models.FaultVaultAction{},
	 &models.FaultVaultContext{},
	 &models.FaultVaultLearning{},
	 // ARES Foundation
	 &models.AresConfig{},
	 &models.ConversationImport{},
	 &models.FileScanResult{},
	)
}
