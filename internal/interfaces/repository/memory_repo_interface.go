package Repositories

import (
	"ares_api/internal/models"

	"github.com/google/uuid"
)

type MemoryRepository interface {
	// Basic snapshot operations
	SaveSnapshot(snapshot *models.MemorySnapshot) error
	GetRecentSnapshots(userID uint, limit int) ([]models.MemorySnapshot, error)
	GetSnapshotsByEventType(userID uint, eventType string, limit int) ([]models.MemorySnapshot, error)
	GetSnapshotsBySessionID(sessionID uuid.UUID, limit int) ([]models.MemorySnapshot, error)
	GetSnapshotByID(snapshotID uint) (*models.MemorySnapshot, error)

	// Embedding operations
	SaveEmbedding(snapshotID uint, embedding []float32) error
	GetPendingEmbeddings(batchSize int) ([]*models.EmbeddingQueueItem, error)
	UpdateEmbeddingQueueStatus(queueID uint, status string) error
	SetEmbeddingQueueError(queueID uint, errorMsg string) error

	// Semantic search
	SemanticSearch(queryEmbedding []float32, limit int, threshold float64) ([]*models.MemorySnapshot, error)

	// Memory management
	UpdateAccessStats(snapshotID uint) error
	RecalculateImportance(snapshotID uint) error
	GetOldMemories(daysOld int) ([]*models.MemorySnapshot, error)
	ArchiveMemory(snapshotID uint) error
	UpdateCacheTemperatures() error
}
