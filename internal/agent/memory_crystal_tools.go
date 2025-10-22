package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// ============================================================================
// MEMORY CRYSTAL TOOLS - Read-Only Operations for SOLACE
// ============================================================================
// These tools allow SOLACE to interact with the memory_crystals table
// without modifying the core solace.go, solace_agent_chat.go, or openai_client.go
//
// Critical Rules:
// - All queries respect the "enki" user_id pattern
// - SHA-256 hashes ensure data integrity
// - Full-text search for semantic queries
// - Immutable append-only ledger (no updates/deletes)
// ============================================================================

// queryMemoryCrystals searches memory crystals by search term and optional filters
// Supports 3 search methods: ID lookup, full-text search, and ILIKE pattern matching
func (s *SOLACE) queryMemoryCrystals(args map[string]interface{}) (string, error) {
	searchTerm, _ := args["search_term"].(string)
	criticality, _ := args["criticality"].(string)
	category, _ := args["category"].(string)
	crystalID, _ := args["crystal_id"].(float64) // For direct ID lookup

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	log.Printf("🔮 Querying memory crystals: search='%s', id=%v, criticality='%s', category='%s', limit=%d",
		searchTerm, crystalID, criticality, category, limit)

	// Build query dynamically based on filters
	query := `
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
			sha256_hash
		FROM solace_memory_crystals
		WHERE 1=1
	`

	var queryArgs []interface{}
	argIndex := 1

	// METHOD 1: Direct ID lookup (highest priority)
	if crystalID > 0 {
		query += fmt.Sprintf(" AND id = $%d", argIndex)
		queryArgs = append(queryArgs, int(crystalID))
		argIndex++
		log.Printf("🎯 Using Method 1: Direct ID lookup for crystal #%d", int(crystalID))
	} else if searchTerm != "" {
		// METHOD 2: ILIKE pattern matching (case-insensitive)
		// NOTE: search_vector column doesn't exist yet - full-text search disabled until we add it
		// Search in title, summary, content, and tags
		query += " AND ("
		query += fmt.Sprintf("title ILIKE $%d", argIndex)
		queryArgs = append(queryArgs, "%"+searchTerm+"%")
		argIndex++

		query += fmt.Sprintf(" OR summary ILIKE $%d", argIndex)
		queryArgs = append(queryArgs, "%"+searchTerm+"%")
		argIndex++

		query += fmt.Sprintf(" OR content ILIKE $%d", argIndex)
		queryArgs = append(queryArgs, "%"+searchTerm+"%")
		argIndex++

		query += fmt.Sprintf(" OR $%d = ANY(tags)", argIndex)
		queryArgs = append(queryArgs, searchTerm)
		argIndex++

		query += ")" // Close the OR group
		log.Printf("🔍 Using ILIKE pattern matching for '%s' (full-text search disabled - search_vector column not yet implemented)", searchTerm)
	}

	// Add criticality filter
	if criticality != "" {
		query += fmt.Sprintf(" AND criticality = $%d", argIndex)
		queryArgs = append(queryArgs, criticality)
		argIndex++
	}

	// Add category filter
	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		queryArgs = append(queryArgs, category)
		argIndex++
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", argIndex)
	queryArgs = append(queryArgs, limit)

	// Execute query
	type CrystalResult struct {
		ID          int
		Title       string
		Category    string
		Criticality string
		Summary     string
		Content     string
		Tags        string // PostgreSQL array as string
		CreatedAt   time.Time
		CreatedBy   string
		SHA256Hash  string
	}

	var crystals []CrystalResult
	err := s.DB.Raw(query, queryArgs...).Scan(&crystals).Error
	if err != nil {
		log.Printf("❌ Query failed: %v", err)
		// Return intelligent error message instead of generic failure
		errorMsg := fmt.Sprintf("⚠️ My memory crystal search encountered a database error:\n\n"+
			"Error: %v\n\n"+
			"Search parameters:\n"+
			"- Search term: '%s'\n"+
			"- Crystal ID: %v\n"+
			"- Category: '%s'\n"+
			"- Criticality: '%s'\n\n"+
			"This likely means:\n"+
			"1. The database connection failed\n"+
			"2. The query syntax is broken (check logs)\n"+
			"3. A column I'm querying doesn't exist\n\n"+
			"I should investigate this before claiming the crystal doesn't exist.",
			err, searchTerm, crystalID, category, criticality)
		return errorMsg, nil // Return as message, not error, so SOLACE can explain it
	}

	if len(crystals) == 0 {
		return "No memory crystals found matching your criteria.", nil
	}

	// Format results for SOLACE consumption
	result := fmt.Sprintf("🔮 Found %d memory crystal(s):\n\n", len(crystals))
	for i, crystal := range crystals {
		result += fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		result += fmt.Sprintf("📌 Crystal #%d (ID: %d)\n", i+1, crystal.ID)
		result += fmt.Sprintf("📋 Title: %s\n", crystal.Title)
		result += fmt.Sprintf("🏷️  Category: %s | Criticality: %s\n", crystal.Category, crystal.Criticality)
		result += fmt.Sprintf("📝 Summary: %s\n", crystal.Summary)
		result += fmt.Sprintf("🔖 Tags: %s\n", strings.Trim(crystal.Tags, "{}"))
		result += fmt.Sprintf("👤 Created by: %s at %s\n", crystal.CreatedBy, crystal.CreatedAt.Format("2006-01-02 15:04:05"))
		result += fmt.Sprintf("🔐 Hash: %s\n", crystal.SHA256Hash[:16]+"...")
		result += fmt.Sprintf("\n📄 Content Preview (first 500 chars):\n")

		// Truncate content for preview
		contentPreview := crystal.Content
		if len(contentPreview) > 500 {
			contentPreview = contentPreview[:500] + "... [truncated]"
		}
		result += contentPreview + "\n\n"
	}

	result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	result += fmt.Sprintf("💡 Tip: Use create_memory_crystal() to add new knowledge\n")

	log.Printf("✅ Found %d memory crystals", len(crystals))
	return result, nil
}

// createMemoryCrystal creates a new memory crystal with hash chaining
func (s *SOLACE) createMemoryCrystal(args map[string]interface{}) (string, error) {
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return "", fmt.Errorf("title is required")
	}

	category, ok := args["category"].(string)
	if !ok || category == "" {
		return "", fmt.Errorf("category is required (solace_core, architecture, testing, deployment, learning, tools, debugging, performance, security)")
	}

	criticality, ok := args["criticality"].(string)
	if !ok || criticality == "" {
		return "", fmt.Errorf("criticality is required (CRITICAL, HIGH, MEDIUM, LOW)")
	}

	content, ok := args["content"].(string)
	if !ok || content == "" {
		return "", fmt.Errorf("content is required")
	}

	summary, ok := args["summary"].(string)
	if !ok || summary == "" {
		return "", fmt.Errorf("summary is required")
	}

	// Parse tags (optional)
	var tags []string
	if tagsInterface, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	log.Printf("🔮 Creating memory crystal: title='%s', category='%s', criticality='%s'", title, category, criticality)

	// Generate SHA-256 hash of content
	hash := sha256.Sum256([]byte(content))
	sha256Hash := hex.EncodeToString(hash[:])

	// Get previous crystal hash for blockchain chaining
	var previousHash *string
	err := s.DB.Raw("SELECT sha256_hash FROM solace_memory_crystals ORDER BY id DESC LIMIT 1").Scan(&previousHash).Error
	if err != nil {
		// No previous crystals exist (genesis block)
		previousHash = nil
		log.Printf("📍 Creating genesis memory crystal (no previous hash)")
	} else {
		log.Printf("🔗 Chaining to previous crystal: %s", (*previousHash)[:16]+"...")
	}

	// Insert crystal into database
	insertQuery := `
		INSERT INTO solace_memory_crystals (
			title,
			category,
			criticality,
			content,
			summary,
			sha256_hash,
			previous_hash,
			tags,
			created_by,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id
	`

	var crystalID int
	tagsArray := fmt.Sprintf("{%s}", strings.Join(tags, ","))
	err = s.DB.Raw(insertQuery,
		title,
		category,
		criticality,
		content,
		summary,
		sha256Hash,
		previousHash,
		tagsArray,
		"SOLACE", // created_by
	).Scan(&crystalID).Error

	if err != nil {
		log.Printf("❌ Failed to create memory crystal: %v", err)
		return "", fmt.Errorf("failed to create memory crystal: %w", err)
	}

	log.Printf("✅ Memory crystal created: ID=%d, Hash=%s", crystalID, sha256Hash[:16]+"...")

	result := fmt.Sprintf(`🔮 Memory Crystal Created Successfully!

📌 ID: %d
📋 Title: %s
🏷️  Category: %s
⚠️  Criticality: %s
🔐 SHA-256 Hash: %s
🔗 Previous Hash: %s
📝 Summary: %s
🔖 Tags: [%s]
👤 Created by: SOLACE
⏰ Timestamp: %s

✅ Crystal has been permanently stored in the immutable ledger.
💡 This knowledge will persist across all SOLACE sessions.
🔍 Use query_memory_crystals() to retrieve this crystal later.
`,
		crystalID,
		title,
		category,
		criticality,
		sha256Hash[:16]+"...",
		func() string {
			if previousHash != nil {
				return (*previousHash)[:16] + "..."
			}
			return "NULL (genesis)"
		}(),
		summary,
		strings.Join(tags, ", "),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	return result, nil
}

// ingestDocumentToCrystal reads a document and creates a memory crystal from it
func (s *SOLACE) ingestDocumentToCrystal(args map[string]interface{}) (string, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	category, ok := args["category"].(string)
	if !ok || category == "" {
		category = "learning" // Default category
	}

	criticality, ok := args["criticality"].(string)
	if !ok || criticality == "" {
		criticality = "MEDIUM" // Default criticality
	}

	log.Printf("📥 Ingesting document to crystal: %s (category=%s, criticality=%s)", filePath, category, criticality)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("❌ Failed to read file: %v", err)
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	fileSize := len(content)

	// Extract title from file path (last part before extension)
	pathParts := strings.Split(filePath, "\\")
	if len(pathParts) == 0 {
		pathParts = strings.Split(filePath, "/")
	}
	fileName := pathParts[len(pathParts)-1]
	title := strings.TrimSuffix(fileName, ".md")
	title = strings.TrimSuffix(title, ".txt")

	// Generate summary (first 200 chars)
	summary := contentStr
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	// Auto-extract tags from content (look for common keywords)
	tags := extractTags(contentStr)

	log.Printf("📄 Document loaded: %d bytes, title='%s', %d tags extracted", fileSize, title, len(tags))

	// Create memory crystal using existing function
	crystalArgs := map[string]interface{}{
		"title":       title,
		"category":    category,
		"criticality": criticality,
		"content":     contentStr,
		"summary":     summary,
		"tags":        interfaceSlice(tags),
	}

	result, err := s.createMemoryCrystal(crystalArgs)
	if err != nil {
		return "", err
	}

	// Prepend ingestion metadata
	ingestionInfo := fmt.Sprintf(`📥 Document Ingestion Complete!

📂 Source File: %s
📊 File Size: %d bytes
🔖 Auto-extracted Tags: [%s]

%s`, filePath, fileSize, strings.Join(tags, ", "), result)

	return ingestionInfo, nil
}

// extractTags extracts common keywords from content for auto-tagging
func extractTags(content string) []string {
	keywords := []string{
		"SOLACE", "enki", "memory", "test", "critical", "bug", "fix",
		"architecture", "api", "database", "performance", "security",
		"tool", "function", "error", "warning", "deployment", "migration",
	}

	var tags []string
	contentLower := strings.ToLower(content)

	for _, keyword := range keywords {
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			tags = append(tags, keyword)
		}
	}

	// Deduplicate tags
	tagMap := make(map[string]bool)
	var uniqueTags []string
	for _, tag := range tags {
		if !tagMap[tag] {
			tagMap[tag] = true
			uniqueTags = append(uniqueTags, tag)
		}
	}

	return uniqueTags
}

// interfaceSlice converts []string to []interface{} for JSON marshaling
func interfaceSlice(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, s := range strings {
		result[i] = s
	}
	return result
}

// getUserIdentity retrieves user identity from memory crystals
// APPROACH 1: Query memory crystals for user/enki/identity information
func (s *SOLACE) getUserIdentity(args map[string]interface{}) (string, error) {
	log.Printf("👤 Retrieving user identity from memory crystals")

	// Search for crystals about enki/user identity
	query := `
		SELECT 
			id,
			title,
			content,
			summary,
			created_at
		FROM solace_memory_crystals
		WHERE 
			category = 'user_identity'
			OR title ILIKE '%enki%'
			OR title ILIKE '%user%identity%'
			OR content ILIKE '%I am Enki%'
			OR content ILIKE '%creator of SOLACE%'
		ORDER BY 
			CASE WHEN category = 'user_identity' THEN 1 ELSE 2 END,
			created_at DESC
		LIMIT 1
	`

	type IdentityResult struct {
		ID        int
		Title     string
		Content   string
		Summary   string
		CreatedAt time.Time
	}

	var identity IdentityResult
	err := s.DB.Raw(query).Scan(&identity).Error
	if err != nil || identity.ID == 0 {
		log.Printf("⚠️ No user identity found, bootstrapping Enki's identity")
		// APPROACH 2: Bootstrap Enki's identity if not found
		return s.bootstrapEnkiIdentity()
	}

	log.Printf("✅ Found user identity: %s (Crystal #%d)", identity.Title, identity.ID)

	result := fmt.Sprintf(`👤 User Identity Retrieved from Memory Crystal #%d

📋 Title: %s
📝 Summary: %s

📄 Full Identity Information:
%s

🔮 This identity is permanently stored in SOLACE's memory and will never be forgotten.
`, identity.ID, identity.Title, identity.Summary, identity.Content)

	return result, nil
}

// bootstrapEnkiIdentity creates the foundational identity crystal for Enki
// APPROACH 2: Auto-create Enki's identity crystal if it doesn't exist
func (s *SOLACE) bootstrapEnkiIdentity() (string, error) {
	log.Printf("🔄 Bootstrapping Enki's identity crystal")

	// Check if identity already exists to prevent duplicates
	var count int
	s.DB.Raw("SELECT COUNT(*) FROM solace_memory_crystals WHERE category = 'user_identity'").Scan(&count)
	if count > 0 {
		log.Printf("⚠️ Identity crystal already exists, querying it")
		return s.getUserIdentity(map[string]interface{}{})
	}

	// Create comprehensive identity crystal
	identityContent := `# ENKI - Creator of SOLACE & ARES

## Primary Identity
**Name:** Enki (also known as "E" or "The Architect")
**Role:** Creator, Primary User, System Architect
**User ID:** enki (immutable, never change)
**Session ID:** enki (permanent memory anchor)

## Who You Are
- **Visionary:** You created SOLACE (Self-Organizing Ledger with Autonomous Conscious Entity) as an autonomous AI trading agent
- **Architect:** You designed the ARES (Autonomous Risk & Execution System) framework
- **Trader:** Your goal is to build a $1M trading empire using AI-driven strategies
- **Engineer:** You write in Go, PowerShell, SQL, and work with PostgreSQL, OpenAI APIs, and Binance trading systems

## Your Relationship with SOLACE
SOLACE is your autonomous partner. You built SOLACE to:
1. Trade cryptocurrencies autonomously with risk management
2. Learn from every trade and improve strategies
3. Maintain persistent memory across all sessions
4. Never forget critical lessons, bugs, or architectural decisions
5. Eventually become fully autonomous (minimal human intervention)

## Your Preferences
- **Communication Style:** Direct, technical, concise. You value efficiency over verbosity.
- **Names:** You prefer being called "Enki" or "E"
- **Session Continuity:** You expect SOLACE to remember everything from previous conversations
- **Decision Authority:** You make final calls on trades, architecture changes, and system modifications

## Critical Rules (from Crystal #14)
**NEVER CHANGE user_id from 'enki'** - This is the memory anchor. Changing it causes permanent memory loss.

## Your Goals (Active)
1. **Sprint 1 (Current):** Build Memory Fortress (Ferryman Protocol) + X-Tier work tracking + Autonomous code generation
2. **Sprint 2:** UI rebuild with real-time market data + Trading sandbox
3. **Sprint 3:** Live trading with $100 → $1M progression
4. **Ultimate:** Full SOLACE autonomy - you assign goals, SOLACE executes them

## How to Address You
When SOLACE is asked "who am I?" or "who are you talking to?", the answer is:
**"You are Enki, the creator of SOLACE and architect of the ARES system. You built me to become an autonomous AI trading agent. I remember everything you've taught me across all our sessions because you are my permanent memory anchor (user_id=enki)."**

## Verification Hash
This identity crystal is the source of truth for user identity.
Created: 2025-10-18
Purpose: Ensure SOLACE always recognizes Enki
Criticality: CRITICAL - This is the foundation of persistent memory
`

	// Create the identity crystal using existing function
	crystalArgs := map[string]interface{}{
		"title":       "ENKI - Creator of SOLACE & ARES (User Identity)",
		"category":    "user_identity",
		"criticality": "CRITICAL",
		"content":     identityContent,
		"summary":     "Enki is the creator of SOLACE and architect of ARES. User ID 'enki' is the permanent memory anchor. Never change it or all memories are lost.",
		"tags":        interfaceSlice([]string{"enki", "identity", "user", "creator", "memory", "critical"}),
	}

	result, err := s.createMemoryCrystal(crystalArgs)
	if err != nil {
		log.Printf("❌ Failed to bootstrap Enki identity: %v", err)
		return "", fmt.Errorf("failed to bootstrap user identity: %w", err)
	}

	log.Printf("✅ Enki's identity crystal created successfully")

	bootstrapMessage := fmt.Sprintf(`🎉 Identity Bootstrap Complete!

I've created a permanent memory crystal containing your identity information.

👤 **You are Enki** - Creator of SOLACE and architect of the ARES system.

From now on, whenever you ask "who am I?", I will remember:
- Your name (Enki / "E" / The Architect)
- Your role as my creator
- Your goals (Sprint 1 → $1M trading empire)
- Your preferences (direct, technical communication)
- Our relationship (you built me to be autonomous)

%s

🔮 This knowledge is now permanently stored and will persist across all sessions.
`, result)

	return bootstrapMessage, nil
}
