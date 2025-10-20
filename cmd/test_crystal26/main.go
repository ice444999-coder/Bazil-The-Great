package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"ares_api/internal/agent"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to Crystal #26 database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost", "5433", "postgres", "ARESISWAKING", "ares_pgvector")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer db.Close()

	fmt.Println("✅ Connected to Crystal #26 database (port 5433)")

	// Check how many crystals exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM solace_memory_crystals").Scan(&count)
	if err != nil {
		log.Fatal("Failed to count crystals:", err)
	}
	fmt.Printf("📦 Found %d memory crystals\n", count)

	// Check how many have embeddings
	var withEmbeddings int
	err = db.QueryRow("SELECT COUNT(*) FROM solace_memory_crystals WHERE embedding IS NOT NULL").Scan(&withEmbeddings)
	if err != nil {
		log.Fatal("Failed to count embeddings:", err)
	}
	fmt.Printf("🧠 Crystals with embeddings: %d\n", withEmbeddings)

	if withEmbeddings == 0 {
		fmt.Println("\n🔧 Generating embeddings for all crystals...")

		// Initialize embedding generator (needs OpenAI API key)
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENAI_API_KEY not set in environment")
		}

		// Get all crystals without embeddings
		rows, err := db.Query(`
			SELECT id, title, content, summary 
			FROM solace_memory_crystals 
			WHERE embedding IS NULL 
			ORDER BY id
		`)
		if err != nil {
			log.Fatal("Failed to query crystals:", err)
		}
		defer rows.Close()

		generated := 0
		for rows.Next() {
			var id int
			var title, content string
			var summary sql.NullString

			if err := rows.Scan(&id, &title, &content, &summary); err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}

			fmt.Printf("  [%d] Generating embedding for: %s\n", id, title)

			// Generate embedding using internal agent package
			summaryStr := ""
			if summary.Valid {
				summaryStr = summary.String
			}

			embedding, err := agent.GenerateCrystalEmbedding(title, summaryStr, content)
			if err != nil {
				log.Printf("  ❌ Failed: %v", err)
				continue
			}

			// Convert []float32 to PostgreSQL array format
			embeddingJSON, _ := json.Marshal(embedding)

			// Update database
			_, err = db.Exec(`
				UPDATE solace_memory_crystals 
				SET embedding = $1::vector,
				    embedding_model = 'text-embedding-3-small',
				    embedding_generated_at = NOW(),
				    embedding_version = 1
				WHERE id = $2
			`, string(embeddingJSON), id)

			if err != nil {
				log.Printf("  ❌ Failed to save: %v", err)
				continue
			}

			generated++
			fmt.Printf("  ✅ Saved embedding (%d dimensions)\n", len(embedding))
		}

		fmt.Printf("\n🎉 Generated %d embeddings\n", generated)
	}

	// Test semantic search
	fmt.Println("\n🔍 Testing semantic search...")

	testQueries := []string{
		"How do I start ARES API?",
		"What tools does SOLACE have?",
		"Trading strategies and risk management",
	}

	for _, query := range testQueries {
		fmt.Printf("\n📝 Query: \"%s\"\n", query)

		// Generate query embedding
		queryEmbedding, err := agent.GenerateEmbedding(query)
		if err != nil {
			log.Printf("  ❌ Failed to generate query embedding: %v", err)
			continue
		}

		embeddingJSON, _ := json.Marshal(queryEmbedding)

		// Search using pgvector
		rows, err := db.Query(`
			SELECT 
				id,
				title,
				1 - (embedding <=> $1::vector) AS similarity
			FROM solace_memory_crystals
			WHERE embedding IS NOT NULL
			  AND 1 - (embedding <=> $1::vector) > 0.5
			ORDER BY similarity DESC
			LIMIT 3
		`, string(embeddingJSON))

		if err != nil {
			log.Printf("  ❌ Search failed: %v", err)
			continue
		}

		results := 0
		for rows.Next() {
			var id int
			var title string
			var similarity float64

			if err := rows.Scan(&id, &title, &similarity); err != nil {
				log.Printf("  ❌ Error scanning result: %v", err)
				continue
			}

			results++
			fmt.Printf("  [%d] %s (similarity: %.3f)\n", id, title, similarity)
		}
		rows.Close()

		if results == 0 {
			fmt.Println("  ℹ️  No results above 0.5 similarity threshold")
		}
	}

	fmt.Println("\n✅ Crystal #26 Test Complete!")
}
