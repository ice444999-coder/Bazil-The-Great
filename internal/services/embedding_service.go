package services

import (
	"ares_api/internal/models"
	repo "ares_api/internal/interfaces/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingServiceImpl handles generating and managing memory embeddings
type EmbeddingServiceImpl struct {
	MemoryRepo     repo.MemoryRepository
	EmbeddingURL   string // URL to embedding service (local or API)
	EmbeddingModel string // Model name (e.g., "all-MiniLM-L6-v2")
}

func NewEmbeddingService(memoryRepo repo.MemoryRepository) *EmbeddingServiceImpl {
	return &EmbeddingServiceImpl{
		MemoryRepo:     memoryRepo,
		EmbeddingURL:   "http://localhost:11434/api/embeddings", // Ollama embeddings endpoint
		EmbeddingModel: "nomic-embed-text", // Local embedding model via Ollama
	}
}

// EmbeddingRequest for Ollama API
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse from Ollama API
type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// GenerateEmbedding creates a vector embedding for text
func (s *EmbeddingServiceImpl) GenerateEmbedding(text string) ([]float32, error) {
	reqBody := EmbeddingRequest{
		Model:  s.EmbeddingModel,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(s.EmbeddingURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API error: %d - %s", resp.StatusCode, string(body))
	}

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embResp.Embedding, nil
}

// GenerateEmbeddingForMemory creates embedding for a memory snapshot
func (s *EmbeddingServiceImpl) GenerateEmbeddingForMemory(snapshotID uint) error {
	// Get the memory snapshot
	snapshot, err := s.MemoryRepo.GetSnapshotByID(snapshotID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Extract text content from payload
	text := s.extractTextFromSnapshot(snapshot)
	if text == "" {
		return fmt.Errorf("no text content in snapshot %d", snapshotID)
	}

	// Generate embedding
	embedding, err := s.GenerateEmbedding(text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Store embedding in database
	if err := s.MemoryRepo.SaveEmbedding(snapshotID, embedding); err != nil {
		return fmt.Errorf("failed to save embedding: %w", err)
	}

	return nil
}

// extractTextFromSnapshot extracts meaningful text from a memory snapshot
func (s *EmbeddingServiceImpl) extractTextFromSnapshot(snapshot *models.MemorySnapshot) string {
	text := fmt.Sprintf("Event: %s. Time: %s. ", snapshot.EventType, snapshot.Timestamp.Format(time.RFC3339))

	// Extract text from JSONB payload
	if snapshot.Payload != nil {
		// Try common text fields
		if message, ok := snapshot.Payload["message"].(string); ok {
			text += message + ". "
		}
		if response, ok := snapshot.Payload["response"].(string); ok {
			text += response + ". "
		}
		if content, ok := snapshot.Payload["content"].(string); ok {
			text += content + ". "
		}
		if description, ok := snapshot.Payload["description"].(string); ok {
			text += description + ". "
		}
	}

	return text
}

// ProcessEmbeddingQueue processes pending embeddings
func (s *EmbeddingServiceImpl) ProcessEmbeddingQueue(batchSize int) (int, error) {
	// Get pending items from queue
	queueItems, err := s.MemoryRepo.GetPendingEmbeddings(batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending embeddings: %w", err)
	}

	processed := 0
	for _, item := range queueItems {
		// Mark as processing
		if err := s.MemoryRepo.UpdateEmbeddingQueueStatus(item.ID, "processing"); err != nil {
			continue
		}

		// Generate embedding
		if err := s.GenerateEmbeddingForMemory(item.SnapshotID); err != nil {
			// Mark as failed with error
			s.MemoryRepo.UpdateEmbeddingQueueStatus(item.ID, "failed")
			s.MemoryRepo.SetEmbeddingQueueError(item.ID, err.Error())
			continue
		}

		// Mark as completed
		if err := s.MemoryRepo.UpdateEmbeddingQueueStatus(item.ID, "completed"); err != nil {
			continue
		}

		processed++
	}

	return processed, nil
}

// SemanticSearch finds memories similar to query text
func (s *EmbeddingServiceImpl) SemanticSearch(queryText string, limit int, threshold float64) ([]*models.MemorySnapshot, error) {
	// Generate embedding for query
	queryEmbedding, err := s.GenerateEmbedding(queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search using vector similarity
	snapshots, err := s.MemoryRepo.SemanticSearch(queryEmbedding, limit, threshold)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	// Update access stats for retrieved memories
	for _, snapshot := range snapshots {
		s.MemoryRepo.UpdateAccessStats(snapshot.ID)
	}

	return snapshots, nil
}

// UpdateMemoryImportance recalculates importance score for a memory
func (s *EmbeddingServiceImpl) UpdateMemoryImportance(snapshotID uint) error {
	return s.MemoryRepo.RecalculateImportance(snapshotID)
}

// PromoteHotMemories promotes frequently accessed memories to hot cache
func (s *EmbeddingServiceImpl) PromoteHotMemories() error {
	return s.MemoryRepo.UpdateCacheTemperatures()
}

// ConsolidateOldMemories merges and compresses old similar memories
func (s *EmbeddingServiceImpl) ConsolidateOldMemories(daysOld int, similarityThreshold float64) (int, error) {
	// Get old memories that haven't been accessed recently
	oldSnapshots, err := s.MemoryRepo.GetOldMemories(daysOld)
	if err != nil {
		return 0, fmt.Errorf("failed to get old memories: %w", err)
	}

	consolidated := 0

	// Group similar memories
	for i := 0; i < len(oldSnapshots); i++ {
		snapshot := oldSnapshots[i]

		// Find similar memories
		text := s.extractTextFromSnapshot(snapshot)
		similar, err := s.SemanticSearch(text, 5, similarityThreshold)
		if err != nil || len(similar) < 2 {
			continue
		}

		// Consolidate similar memories into one summary
		summaryText := s.summarizeMemories(similar)

		// Create consolidated memory
		consolidatedSnapshot := &models.MemorySnapshot{
			Timestamp: time.Now(),
			EventType: "memory_consolidation",
			Payload: models.JSONB{
				"original_count": len(similar),
				"summary":        summaryText,
				"original_ids":   s.getSnapshotIDs(similar),
			},
			UserID: snapshot.UserID,
		}

		if err := s.MemoryRepo.SaveSnapshot(consolidatedSnapshot); err != nil {
			continue
		}

		// Archive originals
		for _, orig := range similar {
			s.MemoryRepo.ArchiveMemory(orig.ID)
		}

		consolidated++
	}

	return consolidated, nil
}

// summarizeMemories creates a summary of multiple related memories
func (s *EmbeddingServiceImpl) summarizeMemories(snapshots []*models.MemorySnapshot) string {
	summary := fmt.Sprintf("Summary of %d related memories: ", len(snapshots))

	for _, snapshot := range snapshots {
		text := s.extractTextFromSnapshot(snapshot)
		if len(text) > 200 {
			text = text[:200] + "..."
		}
		summary += text + " | "
	}

	return summary
}

// getSnapshotIDs extracts IDs from snapshot slice
func (s *EmbeddingServiceImpl) getSnapshotIDs(snapshots []*models.MemorySnapshot) []uint {
	ids := make([]uint, len(snapshots))
	for i, s := range snapshots {
		ids[i] = s.ID
	}
	return ids
}
