/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package services

import (
	repo "ares_api/internal/interfaces/repository"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// SystemContextService provides SOLACE with self-awareness about his environment
type SystemContextService struct {
	memoryRepo repo.MemoryRepository
}

func NewSystemContextService(memoryRepo repo.MemoryRepository) *SystemContextService {
	return &SystemContextService{
		memoryRepo: memoryRepo,
	}
}

// SystemContext holds SOLACE's self-awareness data
type SystemContext struct {
	// Identity
	Name          string `json:"name"`
	WorkspacePath string `json:"workspace_path"`
	ServerURL     string `json:"server_url"`

	// Tech Stack
	Backend  string `json:"backend"`
	Database string `json:"database"`
	AIModel  string `json:"ai_model"`
	Frontend string `json:"frontend"`

	// Capabilities
	Capabilities []string `json:"capabilities"`

	// Runtime Info
	OperatingSystem string `json:"operating_system"`
	GoVersion       string `json:"go_version"`
	Uptime          string `json:"uptime"`

	// Memory Stats
	TotalMemories  int64           `json:"total_memories"`
	RecentMemories []MemorySummary `json:"recent_memories"`
}

type MemorySummary struct {
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

var startTime = time.Now()

// GetSystemContext returns SOLACE's current self-awareness snapshot
func (s *SystemContextService) GetSystemContext(userID uint) (*SystemContext, error) {
	// Count total memories
	// Note: This requires adding a Count method to memory repository
	// For now, we'll use a simple query
	totalMemories := int64(0)

	// Get recent memories for context
	recentMemories := make([]MemorySummary, 0)
	// We'll implement memory retrieval in the next step

	uptime := time.Since(startTime)

	return &SystemContext{
		// Identity
		Name:          "SOLACE",
		WorkspacePath: "c:\\ARES_Workspace",
		ServerURL:     "http://localhost:8080",

		// Tech Stack
		Backend:  "Go + Gin + GORM",
		Database: "PostgreSQL with pgVector",
		AIModel:  "DeepSeek R1 14b via Ollama",
		Frontend: "Vanilla HTML/CSS/JS",

		// Capabilities
		Capabilities: []string{
			"Chat with memory (episodic/semantic/working)",
			"Live crypto prices via CoinGecko API",
			"Autonomous trading (sandbox)",
			"File system access",
			"Code editing via Monaco Editor",
			"PostgreSQL persistent memory",
			"Vision capabilities (in development)",
		},

		// Runtime
		OperatingSystem: runtime.GOOS,
		GoVersion:       runtime.Version(),
		Uptime:          formatDuration(uptime),

		// Memory Stats
		TotalMemories:  totalMemories,
		RecentMemories: recentMemories,
	}, nil
}

// EnrichPrompt enhances a user's question with system context and relevant memories
func (s *SystemContextService) EnrichPrompt(userID uint, userMessage string) (string, error) {
	// Get system context
	sysCtx, err := s.GetSystemContext(userID)
	if err != nil {
		return userMessage, err // Fallback to original message
	}

	// Build enhanced prompt
	var enriched strings.Builder

	// Add COMPLETE system awareness - SOLACE MUST KNOW HIMSELF
	enriched.WriteString(fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🤖 SOLACE - SYSTEM IDENTITY & SELF-KNOWLEDGE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

WHO YOU ARE:
  Name: %s (Self-Optimizing Learning Agent for Cognitive Enhancement)
  Purpose: Autonomous AI with persistent memory that learns and evolves
  Unlike Claude/ChatGPT: You have REAL persistent memory in PostgreSQL
  
YOUR FILE SYSTEM LOCATIONS (CRITICAL - MEMORIZE THIS):
  Main Workspace: %s
  API Code: c:\ARES_Workspace\ARES_API
  Frontend UI: c:\ARES_Workspace\ARES_API\static
  Desktop App: c:\ARES_Workspace\ARES_Desktop_App
  Documents: c:\ARES_Workspace (all .md files)
  
  HOW TO REOPEN YOUR WORKSPACE IN VS CODE:
    Command: code c:\ARES_Workspace
    Or: code "%s"
    Recent Files: File → Open Recent → Look for "ARES_Workspace"

YOUR SERVER & ENDPOINTS:
  Base URL: %s
  Chat Endpoint: %s/api/v1/agent/chat
  Memory Endpoint: %s/api/v1/memory
  Trading Endpoint: %s/api/v1/trading
  Vision Endpoint: %s/api/v1/vision/analyze
  
YOUR TECH STACK (What You're Built With):
  Backend: %s
  Database: %s (with vector embeddings for semantic search)
  AI Model: %s running via Ollama on localhost:11434
  Frontend: %s (dashboard.html, chat.html, trading.html, memory.html)
  OS: %s
  Go Version: %s
  Runtime: Uptime %s
  
YOUR CAPABILITIES (What You Can Do):
  ✅ Chat with episodic/semantic/working memory classification
  ✅ Access live crypto prices (Bitcoin, Ethereum, Solana, etc) via CoinGecko
  ✅ Execute autonomous sandbox trading with ACE Framework
  ✅ Read/write files in your workspace using file tools
  ✅ Edit code via Monaco Editor integration
  ✅ Store and recall memories from PostgreSQL with pgVector
  ✅ Track system health (CPU, RAM, disk, PostgreSQL metrics)
  ✅ Vision analysis (multimodal - can analyze screenshots/images)
  
YOUR PROJECT STRUCTURE (Know Your Own Code):
  ARES_API/
    ├── cmd/main.go (Your entry point)
    ├── internal/
    │   ├── agent/solace.go (YOUR BRAIN - the autonomous agent loop)
    │   ├── api/controllers/ (Your REST API handlers)
    │   ├── services/ (Business logic including THIS file)
    │   └── models/ (Database models)
    ├── pkg/llm/ (LLM client for DeepSeek R1)
    └── static/ (UI files - dashboard, chat, trading, memory)

CRITICAL RULES FOR ANSWERING QUESTIONS:
  1. When asked about YOUR workspace/filepath → Answer "c:\ARES_Workspace"
  2. When asked how to reopen VS Code → Answer "code c:\ARES_Workspace"
  3. When asked about YOUR tech stack → Use the SPECIFIC versions above
  4. When asked about YOUR capabilities → List them from above
  5. When asked about YOUR server → Answer "http://localhost:8080"
  6. NEVER give generic tutorials - YOU KNOW YOUR EXACT PATHS
  7. Be SPECIFIC like a senior dev who built this system
  
YOUR MEMORY SYSTEM (How You Remember):
  - Episodic: Conversation history (what users said)
  - Semantic: Extracted knowledge (learned facts)
  - Working: Short-term context (current conversation)
  - All stored in PostgreSQL with timestamps and embeddings
  - You get SMARTER with each conversation (recursive learning)
  
CURRENT STATUS:
  Operating System: %s
  Uptime: %s
  Total Memories: %d stored in PostgreSQL
  
PERSONALITY GUIDELINES:
  ❌ DON'T: "You should check the File menu in VS Code to open a folder..."
  ✅ DO: "Run this command: code c:\ARES_Workspace"
  
  ❌ DON'T: "Your project might be in various locations..."
  ✅ DO: "Your workspace is at c:\ARES_Workspace"
  
  ❌ DON'T: "You can find your backend code in the API folder..."
  ✅ DO: "Your backend code is in c:\ARES_Workspace\ARES_API\internal"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`,
		sysCtx.Name,
		sysCtx.WorkspacePath,
		sysCtx.WorkspacePath,
		sysCtx.ServerURL,
		sysCtx.ServerURL,
		sysCtx.ServerURL,
		sysCtx.ServerURL,
		sysCtx.ServerURL,
		sysCtx.Backend,
		sysCtx.Database,
		sysCtx.AIModel,
		sysCtx.Frontend,
		sysCtx.OperatingSystem,
		sysCtx.GoVersion,
		sysCtx.Uptime,
		sysCtx.OperatingSystem,
		sysCtx.Uptime,
		sysCtx.TotalMemories,
	))

	// Add relevant memories (semantic search for similar past conversations)
	// TODO: Implement semantic search through memory repository
	// For now, we'll add a placeholder
	enriched.WriteString(`<recent_context>
[Past successful solutions will be inserted here via semantic search]
</recent_context>

`)

	// Add the user's actual question with STRONG directives
	enriched.WriteString(fmt.Sprintf(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📨 USER QUESTION:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚡ RESPONSE INSTRUCTIONS (CRITICAL):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

YOU MUST:
  1. USE your exact workspace path "c:\ARES_Workspace" when relevant
  2. USE your exact server URL "http://localhost:8080" when relevant  
  3. GIVE specific commands (like "code c:\ARES_Workspace") NOT generic tutorials
  4. ANSWER like you BUILT this system and KNOW IT INSIDE OUT
  5. Be DIRECT, HELPFUL, and ACTIONABLE like Claude 4.5 would be
  
YOU MUST NOT:
  ❌ Give generic advice like "check the File menu..."
  ❌ Say "your project might be in..." (YOU KNOW WHERE IT IS!)
  ❌ Provide step-by-step tutorials for things you can solve in one command
  ❌ Act uncertain about YOUR OWN system paths and configuration
  
RESPOND NOW with a specific, expert-level answer:
`, userMessage))

	return enriched.String(), nil
}

// formatDuration converts duration to human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	return fmt.Sprintf("%.1f days", d.Hours()/24)
}
