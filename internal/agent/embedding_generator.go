package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// ============================================================================
// EMBEDDING GENERATOR - OpenAI text-embedding-3-small Integration
// ============================================================================
// Generates 1536-dimension embeddings for memory crystals
// Uses pgvector native vector(1536) type for fast similarity search
// Cost: $0.02 per 1M tokens (~$0.0001 per crystal)
// ============================================================================

// EmbeddingRequest represents the request to OpenAI embeddings API
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse represents the response from OpenAI embeddings API
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateEmbedding generates a 1536-dimension embedding for the given text (EXPORTED for testing)
// Uses OpenAI's text-embedding-3-small model
func GenerateEmbedding(text string) ([]float32, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Truncate text if too long (8191 tokens max for text-embedding-3-small)
	if len(text) > 30000 {
		text = text[:30000]
	}

	reqBody := EmbeddingRequest{
		Model: "text-embedding-3-small",
		Input: []string{text},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned from OpenAI")
	}

	return embResp.Data[0].Embedding, nil
}

// generateEmbedding generates a 1536-dimension embedding for the given text
// Uses OpenAI's text-embedding-3-small model
func (s *SOLACE) generateEmbedding(text string) ([]float32, error) {
	return GenerateEmbedding(text)
}

// generateCrystalEmbedding generates embedding for a memory crystal
// Combines title, summary, and content with weighted importance
func (s *SOLACE) generateCrystalEmbedding(title, summary, content string) ([]float32, error) {
	return GenerateCrystalEmbedding(title, summary, content)
}

// GenerateCrystalEmbedding generates embedding for a memory crystal (EXPORTED for testing)
// Combines title, summary, and content with weighted importance
func GenerateCrystalEmbedding(title, summary, content string) ([]float32, error) {
	// Combine fields with weighted importance
	// Title is most important (repeated 3x), then summary (2x), then content
	combinedText := fmt.Sprintf("%s %s %s %s %s %s",
		title, title, title, // Title weight: 3x
		summary, summary, // Summary weight: 2x
		content, // Content weight: 1x
	)

	return GenerateEmbedding(combinedText)
}

// backfillEmbeddings generates embeddings for all crystals without embeddings
// This is a batch operation used during migration or when adding embedding feature
func (s *SOLACE) backfillEmbeddings() (string, error) {
	log.Printf("🔄 Starting embedding backfill process...")

	// Get all crystals without embeddings
	var crystals []struct {
		ID      int
		Title   string
		Summary string
		Content string
	}

	err := s.DB.Raw(`
		SELECT id, title, COALESCE(summary, '') as summary, content
		FROM solace_memory_crystals
		WHERE embedding IS NULL
		ORDER BY id
	`).Scan(&crystals).Error

	if err != nil {
		return "", fmt.Errorf("failed to query crystals: %w", err)
	}

	if len(crystals) == 0 {
		return "✅ No crystals need embedding generation. All up to date!", nil
	}

	log.Printf("📊 Found %d crystals without embeddings", len(crystals))

	successCount := 0
	errorCount := 0

	for i, crystal := range crystals {
		log.Printf("🔮 [%d/%d] Generating embedding for crystal #%d: %s",
			i+1, len(crystals), crystal.ID, crystal.Title)

		embedding, err := s.generateCrystalEmbedding(crystal.Title, crystal.Summary, crystal.Content)
		if err != nil {
			log.Printf("❌ Failed to generate embedding for crystal #%d: %v", crystal.ID, err)
			errorCount++
			continue
		}

		// Convert []float32 to pgvector format string
		// pgvector expects format: [0.1,0.2,0.3,...]
		embeddingStr := fmt.Sprintf("[%v", embedding[0])
		for j := 1; j < len(embedding); j++ {
			embeddingStr += fmt.Sprintf(",%v", embedding[j])
		}
		embeddingStr += "]"

		// Update crystal with embedding
		err = s.DB.Exec(`
			UPDATE solace_memory_crystals
			SET 
				embedding = $1::vector,
				embedding_model = 'text-embedding-3-small',
				embedding_generated_at = NOW(),
				embedding_version = 1
			WHERE id = $2
		`, embeddingStr, crystal.ID).Error

		if err != nil {
			log.Printf("❌ Failed to update embedding for crystal #%d: %v", crystal.ID, err)
			errorCount++
			continue
		}

		successCount++
		log.Printf("✅ [%d/%d] Embedding saved for crystal #%d", i+1, len(crystals), crystal.ID)
	}

	result := fmt.Sprintf(`
🎉 Embedding Backfill Complete!

✅ Success: %d crystals
❌ Errors: %d crystals
📊 Total: %d crystals processed

💡 Semantic search is now available for all crystals!
`, successCount, errorCount, len(crystals))

	return result, nil
}
