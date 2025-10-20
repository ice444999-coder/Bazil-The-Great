package tools

import (
	"fmt"

	"ares_api/internal/agent"

	"gorm.io/gorm"
)

// ToolVectorSearch - Mathematical vector search for tools
type ToolVectorSearch struct {
	DB *gorm.DB
}

// ToolSearchResult - Search result with similarity score
type ToolSearchResult struct {
	ToolID         string                 `json:"tool_id"`
	ToolName       string                 `json:"tool_name"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	RequiredParams map[string]interface{} `json:"required_params"`
	RiskLevel      string                 `json:"risk_level"`
	ImplementedIn  string                 `json:"implemented_in"`
	Similarity     float64                `json:"similarity"` // 0.0 to 1.0
}

// SearchToolsByIntent - Claude-level semantic search using pgvector
func (tvs *ToolVectorSearch) SearchToolsByIntent(intent string, minSimilarity float64, limit int) ([]ToolSearchResult, error) {
	// Generate embedding for user intent using existing OpenAI integration
	embedding, err := agent.GenerateEmbedding(intent)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert []float32 to string format for PostgreSQL
	// Mathematical vector similarity search using pgvector
	// Formula: similarity = 1 - cosine_distance
	var results []ToolSearchResult

	err = tvs.DB.Raw(`
        SELECT 
            tool_id,
            tool_name,
            description,
            tool_category AS category,
            required_params,
            risk_level,
            implemented_in,
            1 - (embedding <=> ?::vector) AS similarity
        FROM tool_registry
        WHERE embedding IS NOT NULL
        AND 1 - (embedding <=> ?::vector) > ?
        ORDER BY similarity DESC
        LIMIT ?
    `, embedding, embedding, minSimilarity, limit).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	return results, nil
}

// VerifyMathematicalAccuracy - Test pgvector is working correctly
func (tvs *ToolVectorSearch) VerifyMathematicalAccuracy() error {
	// Test: Identical vectors should have distance = 0.0, similarity = 1.0
	testVector := make([]float32, 1536)
	for i := range testVector {
		testVector[i] = 0.1
	}

	var similarity float64
	err := tvs.DB.Raw(`
        SELECT 1 - (?::vector <=> ?::vector) AS similarity
    `, testVector, testVector).Scan(&similarity).Error

	if err != nil {
		return fmt.Errorf("vector operations not working: %w", err)
	}

	if similarity < 0.999 || similarity > 1.001 {
		return fmt.Errorf("FAILED: identical vectors should have similarity=1.0, got %.4f", similarity)
	}

	fmt.Printf("✅ pgvector mathematical accuracy verified: identical vectors = 1.0\n")
	return nil
}
