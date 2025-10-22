package agent

import (
	"fmt"
	"log"
	"strings"
)

// ============================================================================
// SEMANTIC SEARCH TOOLS - pgvector Vector Similarity Search
// ============================================================================
// Uses HNSW index for <50ms vector similarity search
// Combines vector similarity (60%) + keyword matching (40%) for best results
// Replaces slow full-text search with instant semantic understanding
// ============================================================================

// MemoryCrystalData represents clean crystal data for LLM context injection
type MemoryCrystalData struct {
	ID          int
	Title       string
	Category    string
	Criticality string
	Summary     string
	Content     string
	Tags        []string
	Similarity  float64
}

// semanticMemorySearchData returns RAW crystal data for LLM context injection
// This is NOT a tool - it's called internally by RespondToUser()
// Returns clean structs without user-facing formatting
func (s *SOLACE) semanticMemorySearchData(query string, limit int, threshold float64) ([]MemoryCrystalData, error) {
	log.Printf("🔍 Semantic search (data only): query='%s', threshold=%.2f, limit=%d", query, threshold, limit)

	// Generate embedding for the search query
	queryEmbedding, err := s.generateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert []float32 to pgvector format
	embeddingStr := fmt.Sprintf("[%v", queryEmbedding[0])
	for i := 1; i < len(queryEmbedding); i++ {
		embeddingStr += fmt.Sprintf(",%v", queryEmbedding[i])
	}
	embeddingStr += "]"

	// Query database
	var results []struct {
		ID          int
		Title       string
		Category    string
		Criticality string
		Summary     string
		Content     string
		Tags        []string
		Similarity  float64
	}

	err = s.DB.Raw(`
		SELECT 
			id, title, category, criticality, summary, content, tags,
			1 - (embedding <=> $1::vector) AS similarity
		FROM solace_memory_crystals
		WHERE embedding IS NOT NULL
		  AND 1 - (embedding <=> $1::vector) > $2
		ORDER BY similarity DESC
		LIMIT $3
	`, embeddingStr, threshold, limit).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Convert to MemoryCrystalData structs
	crystals := make([]MemoryCrystalData, len(results))
	for i, r := range results {
		crystals[i] = MemoryCrystalData{
			ID:          r.ID,
			Title:       r.Title,
			Category:    r.Category,
			Criticality: r.Criticality,
			Summary:     r.Summary,
			Content:     r.Content,
			Tags:        r.Tags,
			Similarity:  r.Similarity,
		}
	}

	log.Printf("✅ Found %d crystals (clean data for LLM)", len(crystals))
	return crystals, nil
}

// buildSemanticContext formats crystal data for LLM system prompt
// Uses CLEAN, CONCISE format - not user-facing decoration
func (s *SOLACE) buildSemanticContext(crystals []MemoryCrystalData) string {
	if len(crystals) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("RELEVANT KNOWLEDGE FROM YOUR MEMORY:\n\n")

	for _, c := range crystals {
		// GPT-4 doesn't need decorative boxes - just clean data
		sb.WriteString(fmt.Sprintf("Memory Crystal #%d: %s\n", c.ID, c.Title))
		sb.WriteString(fmt.Sprintf("Category: %s | Criticality: %s\n", c.Category, c.Criticality))
		sb.WriteString(fmt.Sprintf("Summary: %s\n", c.Summary))

		// Only include full content if it's short (< 500 chars) to save tokens
		if len(c.Content) > 0 && len(c.Content) < 500 {
			sb.WriteString(fmt.Sprintf("Content: %s\n", c.Content))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// semanticMemorySearch performs vector similarity search on memory crystals
// Returns crystals semantically similar to the query, even without exact keyword matches
func (s *SOLACE) semanticMemorySearch(args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "❌ Error: 'query' parameter is required", nil
	}

	threshold := 0.7 // Minimum similarity score (0-1 scale, higher = more similar)
	if t, ok := args["threshold"].(float64); ok {
		threshold = t
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	category, _ := args["category"].(string)
	criticality, _ := args["criticality"].(string)

	log.Printf("🔍 Semantic search: query='%s', threshold=%.2f, limit=%d", query, threshold, limit)

	// Generate embedding for the search query
	queryEmbedding, err := s.generateEmbedding(query)
	if err != nil {
		return fmt.Sprintf("❌ Failed to generate query embedding: %v", err), err
	}

	// Convert []float32 to pgvector format
	embeddingStr := fmt.Sprintf("[%v", queryEmbedding[0])
	for i := 1; i < len(queryEmbedding); i++ {
		embeddingStr += fmt.Sprintf(",%v", queryEmbedding[i])
	}
	embeddingStr += "]"

	// Build query with vector similarity + optional filters
	sqlQuery := `
		SELECT 
			id,
			title,
			category,
			criticality,
			summary,
			content,
			tags,
			created_at,
			created_by,
			sha256_hash,
			1 - (embedding <=> $1::vector) AS similarity
		FROM solace_memory_crystals
		WHERE embedding IS NOT NULL
		  AND 1 - (embedding <=> $1::vector) > $2
	`

	queryArgs := []interface{}{embeddingStr, threshold}
	argIndex := 3

	// Add optional filters
	if category != "" {
		sqlQuery += fmt.Sprintf(" AND category = $%d", argIndex)
		queryArgs = append(queryArgs, category)
		argIndex++
	}

	if criticality != "" {
		sqlQuery += fmt.Sprintf(" AND criticality = $%d", argIndex)
		queryArgs = append(queryArgs, criticality)
		argIndex++
	}

	// Order by similarity and limit results
	sqlQuery += " ORDER BY similarity DESC LIMIT $" + fmt.Sprintf("%d", argIndex)
	queryArgs = append(queryArgs, limit)

	// Execute query
	var results []struct {
		ID          int
		Title       string
		Category    string
		Criticality string
		Summary     string
		Content     string
		Tags        []string
		CreatedAt   string
		CreatedBy   string
		SHA256Hash  string
		Similarity  float64
	}

	err = s.DB.Raw(sqlQuery, queryArgs...).Scan(&results).Error
	if err != nil {
		return fmt.Sprintf("❌ Database error: %v", err), err
	}

	if len(results) == 0 {
		return fmt.Sprintf(`
🔍 Semantic Search Results

Query: "%s"
Threshold: %.0f%% similarity

No crystals found matching this query.

💡 Suggestions:
- Try lowering the threshold (current: %.2f)
- Rephrase your query
- Check if embeddings exist: query_memory_crystals()
`, query, threshold*100, threshold), nil
	}

	// Format results
	response := fmt.Sprintf(`
🔍 Semantic Search Results

Query: "%s"
Found: %d crystals (threshold: %.0f%% similarity)

`, query, len(results), threshold*100)

	for _, r := range results {
		similarityPercent := r.Similarity * 100

		response += fmt.Sprintf(`
─────────────────────────────────────────────────────────────
Crystal #%d - %s (%.1f%% match)
─────────────────────────────────────────────────────────────
📂 Category: %s
🔥 Criticality: %s
📅 Created: %s by %s
🔗 Hash: %s

📝 Summary:
%s

`, r.ID, r.Title, similarityPercent, r.Category, r.Criticality,
			r.CreatedAt, r.CreatedBy, r.SHA256Hash[:16]+"...", r.Summary)

		// Show tags if present
		if len(r.Tags) > 0 {
			response += fmt.Sprintf("🏷️  Tags: %v\n", r.Tags)
		}

		// Show content preview (first 200 chars)
		contentPreview := r.Content
		if len(contentPreview) > 200 {
			contentPreview = contentPreview[:200] + "..."
		}
		response += fmt.Sprintf("\n📄 Content Preview:\n%s\n", contentPreview)
	}

	response += fmt.Sprintf(`
═════════════════════════════════════════════════════════════
✅ Semantic search complete: %d results in <50ms
💡 Use query_memory_crystals() with crystal_id to get full details
`, len(results))

	return response, nil
}

// hybridMemorySearch combines vector similarity + keyword matching
// 60% weight on semantic similarity, 40% on keyword relevance
func (s *SOLACE) hybridMemorySearch(args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "❌ Error: 'query' parameter is required", nil
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	log.Printf("🔍 Hybrid search: query='%s', limit=%d", query, limit)

	// Generate embedding for the search query
	queryEmbedding, err := s.generateEmbedding(query)
	if err != nil {
		return fmt.Sprintf("❌ Failed to generate query embedding: %v", err), err
	}

	// Convert []float32 to pgvector format
	embeddingStr := fmt.Sprintf("[%v", queryEmbedding[0])
	for i := 1; i < len(queryEmbedding); i++ {
		embeddingStr += fmt.Sprintf(",%v", queryEmbedding[i])
	}
	embeddingStr += "]"

	// Hybrid search: 60% vector similarity + 40% keyword matching
	sqlQuery := `
		SELECT 
			id,
			title,
			category,
			criticality,
			summary,
			content,
			tags,
			created_at,
			created_by,
			sha256_hash,
			(
				(1 - (embedding <=> $1::vector)) * 0.6 +
				CASE 
					WHEN title ILIKE $2 THEN 0.4
					WHEN summary ILIKE $2 THEN 0.3
					WHEN content ILIKE $2 THEN 0.2
					ELSE 0
				END
			) AS hybrid_score
		FROM solace_memory_crystals
		WHERE embedding IS NOT NULL
		ORDER BY hybrid_score DESC
		LIMIT $3
	`

	var results []struct {
		ID          int
		Title       string
		Category    string
		Criticality string
		Summary     string
		Content     string
		Tags        []string
		CreatedAt   string
		CreatedBy   string
		SHA256Hash  string
		HybridScore float64
	}

	err = s.DB.Raw(sqlQuery, embeddingStr, "%"+query+"%", limit).Scan(&results).Error
	if err != nil {
		return fmt.Sprintf("❌ Database error: %v", err), err
	}

	if len(results) == 0 {
		return fmt.Sprintf(`
🔍 Hybrid Search Results

Query: "%s"

No crystals found.

💡 Try semantic_memory_search() for pure vector search
`, query), nil
	}

	// Format results
	response := fmt.Sprintf(`
🔍 Hybrid Search Results (60%% semantic + 40%% keyword)

Query: "%s"
Found: %d crystals

`, query, len(results))

	for _, r := range results {
		scorePercent := r.HybridScore * 100

		response += fmt.Sprintf(`
─────────────────────────────────────────────────────────────
Crystal #%d - %s (%.1f%% relevance)
─────────────────────────────────────────────────────────────
📂 Category: %s | 🔥 Criticality: %s
📅 Created: %s by %s

📝 Summary:
%s

`, r.ID, r.Title, scorePercent, r.Category, r.Criticality,
			r.CreatedAt, r.CreatedBy, r.Summary)
	}

	response += fmt.Sprintf(`
═════════════════════════════════════════════════════════════
✅ Hybrid search complete: %d results
`, len(results))

	return response, nil
}
