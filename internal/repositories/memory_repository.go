package repositories

import (
	repository "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MemoryRepositoryImpl struct {
	db *gorm.DB
}

func NewMemoryRepository(db *gorm.DB) repository.MemoryRepository {
	return &MemoryRepositoryImpl{db: db}
}

func (r *MemoryRepositoryImpl) SaveSnapshot(snapshot *models.MemorySnapshot) error {
	return r.db.Create(snapshot).Error
}

func (r *MemoryRepositoryImpl) GetRecentSnapshots(userID uint, limit int) ([]models.MemorySnapshot, error) {
	var snapshots []models.MemorySnapshot
	err := r.db.Where("user_id = ?", userID).
		Order("timestamp desc").
		Limit(limit).
		Find(&snapshots).Error
	return snapshots, err
}

func (r *MemoryRepositoryImpl) GetSnapshotsByEventType(userID uint, eventType string, limit int) ([]models.MemorySnapshot, error) {
	var snapshots []models.MemorySnapshot
	err := r.db.Where("user_id = ? AND event_type = ?", userID, eventType).
		Order("timestamp desc").
		Limit(limit).
		Find(&snapshots).Error
	return snapshots, err
}

func (r *MemoryRepositoryImpl) GetSnapshotsBySessionID(sessionID uuid.UUID, limit int) ([]models.MemorySnapshot, error) {
	var snapshots []models.MemorySnapshot
	err := r.db.Where("session_id = ?", sessionID).
		Order("timestamp desc").
		Limit(limit).
		Find(&snapshots).Error
	return snapshots, err
}

func (r *MemoryRepositoryImpl) GetSnapshotByID(snapshotID uint) (*models.MemorySnapshot, error) {
	var snapshot models.MemorySnapshot
	err := r.db.First(&snapshot, snapshotID).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// ========== EMBEDDING OPERATIONS ==========

func (r *MemoryRepositoryImpl) SaveEmbedding(snapshotID uint, embedding []float32) error {
	// Convert float32 slice to pgvector string format: [0.1, 0.2, 0.3]
	embeddingStr := vectorToString(embedding)

	memoryEmbedding := models.MemoryEmbedding{
		SnapshotID: snapshotID,
		Embedding:  embeddingStr,
	}

	return r.db.Create(&memoryEmbedding).Error
}

func (r *MemoryRepositoryImpl) GetPendingEmbeddings(batchSize int) ([]*models.EmbeddingQueueItem, error) {
	var items []*models.EmbeddingQueueItem
	err := r.db.Where("status = ?", "pending").
		Order("created_at asc").
		Limit(batchSize).
		Find(&items).Error
	return items, err
}

func (r *MemoryRepositoryImpl) UpdateEmbeddingQueueStatus(queueID uint, status string) error {
	now := gorm.Expr("CURRENT_TIMESTAMP")
	updates := map[string]interface{}{"status": status}

	if status == "completed" || status == "failed" {
		updates["processed_at"] = now
	}

	return r.db.Model(&models.EmbeddingQueueItem{}).
		Where("id = ?", queueID).
		Updates(updates).Error
}

func (r *MemoryRepositoryImpl) SetEmbeddingQueueError(queueID uint, errorMsg string) error {
	return r.db.Model(&models.EmbeddingQueueItem{}).
		Where("id = ?", queueID).
		Updates(map[string]interface{}{
			"error_message": errorMsg,
			"retry_count":   gorm.Expr("retry_count + 1"),
		}).Error
}

// ========== SEMANTIC SEARCH ==========

func (r *MemoryRepositoryImpl) SemanticSearch(queryEmbedding []float32, limit int, threshold float64) ([]*models.MemorySnapshot, error) {
	// For now, without pgvector, we'll do a simpler approach:
	// 1. Get all embeddings
	// 2. Calculate cosine similarity in Go
	// 3. Return top results

	// This is a temporary implementation - once pgvector is installed, we'll use the SQL version

	var embeddings []struct {
		SnapshotID uint
		Embedding  string
	}

	err := r.db.Table("memory_embeddings").
		Select("snapshot_id, embedding").
		Find(&embeddings).Error

	if err != nil {
		return nil, err
	}

	// Calculate similarities
	type ScoredSnapshot struct {
		SnapshotID uint
		Score      float64
	}

	var scored []ScoredSnapshot
	for _, emb := range embeddings {
		// Parse embedding string to []float32
		dbVector := stringToVector(emb.Embedding)
		if len(dbVector) != len(queryEmbedding) {
			continue
		}

		// Calculate cosine similarity
		similarity := cosineSimilarity(queryEmbedding, dbVector)
		if similarity >= float32(threshold) {
			scored = append(scored, ScoredSnapshot{
				SnapshotID: emb.SnapshotID,
				Score:      float64(similarity),
			})
		}
	}

	// Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].Score > scored[i].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Get top N snapshot IDs
	topN := limit
	if topN > len(scored) {
		topN = len(scored)
	}

	snapshotIDs := make([]uint, topN)
	for i := 0; i < topN; i++ {
		snapshotIDs[i] = scored[i].SnapshotID
	}

	// Fetch actual snapshots
	var snapshots []*models.MemorySnapshot
	if len(snapshotIDs) > 0 {
		err = r.db.Where("id IN ? AND archived = ?", snapshotIDs, false).
			Find(&snapshots).Error
	}

	return snapshots, err
}

// ========== MEMORY MANAGEMENT ==========

func (r *MemoryRepositoryImpl) UpdateAccessStats(snapshotID uint) error {
	now := gorm.Expr("CURRENT_TIMESTAMP")

	return r.db.Model(&models.MemorySnapshot{}).
		Where("id = ?", snapshotID).
		Updates(map[string]interface{}{
			"access_count":  gorm.Expr("access_count + 1"),
			"last_accessed": now,
		}).Error
}

func (r *MemoryRepositoryImpl) RecalculateImportance(snapshotID uint) error {
	// Use the calculate_memory_importance() function from migration
	return r.db.Exec(`
		UPDATE memory_snapshots
		SET importance_score = calculate_memory_importance(?)
		WHERE id = ?
	`, snapshotID, snapshotID).Error
}

func (r *MemoryRepositoryImpl) GetOldMemories(daysOld int) ([]*models.MemorySnapshot, error) {
	var snapshots []*models.MemorySnapshot

	err := r.db.Where("archived = ?", false).
		Where("timestamp < NOW() - INTERVAL '? days'", daysOld).
		Where("COALESCE(last_accessed, timestamp) < NOW() - INTERVAL '? days'", daysOld/2).
		Order("timestamp asc").
		Find(&snapshots).Error

	return snapshots, err
}

func (r *MemoryRepositoryImpl) ArchiveMemory(snapshotID uint) error {
	return r.db.Model(&models.MemorySnapshot{}).
		Where("id = ?", snapshotID).
		Update("archived", true).Error
}

func (r *MemoryRepositoryImpl) UpdateCacheTemperatures() error {
	// Call update_memory_temperature() for all active memories
	return r.db.Exec(`
		UPDATE memory_snapshots
		SET importance_score = calculate_memory_importance(id)
		WHERE archived = FALSE
	`).Error
}

// ========== HELPER FUNCTIONS ==========

// vectorToString converts []float32 to string format
func vectorToString(vec []float32) string {
	if len(vec) == 0 {
		return "[]"
	}

	result := "["
	for i, v := range vec {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%f", v)
	}
	result += "]"
	return result
}

// stringToVector converts string format back to []float32
func stringToVector(s string) []float32 {
	// Remove brackets
	s = s[1 : len(s)-1]
	if s == "" {
		return []float32{}
	}

	// Split by comma
	var result []float32
	var current string

	for _, c := range s {
		if c == ',' {
			var val float32
			fmt.Sscanf(current, "%f", &val)
			result = append(result, val)
			current = ""
		} else {
			current += string(c)
		}
	}

	// Last value
	if current != "" {
		var val float32
		fmt.Sscanf(current, "%f", &val)
		result = append(result, val)
	}

	return result
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(sqrt(float64(normA))) * float32(sqrt(float64(normB))))
}

func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ { // Newton's method
		z = z - (z*z-x)/(2*z)
	}
	return z
}
