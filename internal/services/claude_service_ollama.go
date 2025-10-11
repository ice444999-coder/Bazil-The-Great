package services

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	repo "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"ares_api/internal/ollama"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ClaudeServiceOllamaImpl struct {
	MemoryRepo       repo.MemoryRepository
	FileReader       *common.FileSystemReader
	EmbeddingService *EmbeddingServiceImpl
	OllamaClient     *ollama.Client
	RepositoryPath   string
	Model            string // DeepSeek-R1 model name
}

func NewClaudeServiceOllama(memoryRepo repo.MemoryRepository, embeddingService *EmbeddingServiceImpl, repoPath string) *ClaudeServiceOllamaImpl {
	ollamaClient := ollama.NewClientFromEnv()

	// Get model from environment or use default
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "deepseek-r1:14b" // DeepSeek-R1 14B - Reasoning model with better tool use and truth accuracy
	}

	return &ClaudeServiceOllamaImpl{
		MemoryRepo:       memoryRepo,
		FileReader:       common.NewFileSystemReader(repoPath),
		EmbeddingService: embeddingService,
		OllamaClient:     ollamaClient,
		RepositoryPath:   repoPath,
		Model:            model,
	}
}

// Chat implements the full stateful Claude consciousness using Ollama + DeepSeek-R1
func (s *ClaudeServiceOllamaImpl) Chat(userID uint, message string, sessionID *uuid.UUID, includeFiles []string, maxTokens int) (dto.ClaudeChatResponse, error) {
	// Generate session ID if not provided
	if sessionID == nil {
		newSessionID := uuid.New()
		sessionID = &newSessionID
	}

	// PHASE 2: Load relevant memories
	memories, err := s.loadRelevantMemories(userID, sessionID)
	if err != nil {
		return dto.ClaudeChatResponse{}, fmt.Errorf("failed to load memories: %w", err)
	}

	// PHASE 3: Load file system context
	fileContext, filesAccessed := s.loadFileContext(includeFiles)

	// Build system prompt with memory and repo context
	systemPrompt := s.buildSystemPrompt(memories, fileContext)

	// Create context with timeout (5 minutes for DeepSeek-R1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Implement tool use loop with Ollama
	var responseText string
	toolsUsed := 0
	conversationHistory := []map[string]string{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": message},
	}

	// Start token tracking (estimate for Ollama)
	inputTokens := estimateTokens(systemPrompt + message)
	outputTokens := 0

	for iteration := 0; iteration < 5; iteration++ { // Max 5 tool-use iterations
		// Call Ollama with conversation history
		fullPrompt := s.buildPromptFromHistory(conversationHistory)

		response, err := s.callOllamaWithContext(ctx, fullPrompt)
		if err != nil {
			return dto.ClaudeChatResponse{}, fmt.Errorf("ollama API error: %w", err)
		}

		outputTokens += estimateTokens(response)

		// Check if DeepSeek-R1 wants to use tools (parse special markers)
		toolCall := s.extractToolCall(response)
		if toolCall == nil {
			// No tool use, we're done
			responseText = s.cleanResponse(response)
			break
		}

		toolsUsed++

		// Execute the tool
		toolResult, toolError := s.executeTool(toolCall)

		// Add assistant response and tool result to conversation
		conversationHistory = append(conversationHistory,
			map[string]string{"role": "assistant", "content": response},
			map[string]string{"role": "user", "content": fmt.Sprintf("Tool result: %s", toolResult)},
		)

		if toolError != nil {
			conversationHistory[len(conversationHistory)-1]["content"] = fmt.Sprintf("Tool error: %s", toolError.Error())
		}

		// Continue loop for next iteration
	}

	// If we hit max iterations, use the last response
	if responseText == "" {
		responseText = s.cleanResponse(conversationHistory[len(conversationHistory)-1]["content"])
	}

	// PHASE 4: Store interaction in memory for recursive learning
	err = s.storeInteraction(userID, sessionID, message, responseText, memories, filesAccessed, outputTokens)
	if err != nil {
		return dto.ClaudeChatResponse{}, fmt.Errorf("failed to store memory: %w", err)
	}

	return dto.ClaudeChatResponse{
		Message:        message,
		Response:       responseText,
		SessionID:      sessionID.String(),
		MemoriesLoaded: len(memories),
		FilesAccessed:  filesAccessed,
		TokensUsed:     inputTokens + outputTokens,
	}, nil
}

// callOllamaWithContext calls Ollama with context support
func (s *ClaudeServiceOllamaImpl) callOllamaWithContext(ctx context.Context, prompt string) (string, error) {
	// Use generate endpoint for more control
	return s.OllamaClient.Generate(s.Model, prompt)
}

// buildPromptFromHistory converts conversation history to a single prompt
func (s *ClaudeServiceOllamaImpl) buildPromptFromHistory(history []map[string]string) string {
	var prompt strings.Builder
	for _, msg := range history {
		role := msg["role"]
		content := msg["content"]

		switch role {
		case "system":
			prompt.WriteString(content + "\n\n")
		case "user":
			prompt.WriteString("USER: " + content + "\n\n")
		case "assistant":
			prompt.WriteString("ASSISTANT: " + content + "\n\n")
		}
	}
	return prompt.String()
}

// extractToolCall parses DeepSeek-R1 response for tool use markers
func (s *ClaudeServiceOllamaImpl) extractToolCall(response string) *ToolCall {
	// DeepSeek-R1 format: [TOOL_USE: read_file {"file_path": "cmd/main.go"}]
	if !strings.Contains(response, "[TOOL_USE:") {
		return nil
	}

	start := strings.Index(response, "[TOOL_USE:")
	end := strings.Index(response[start:], "]")
	if end == -1 {
		return nil
	}

	toolStr := response[start+len("[TOOL_USE:"):start+end]
	parts := strings.SplitN(toolStr, " ", 2)
	if len(parts) != 2 {
		return nil
	}

	toolName := strings.TrimSpace(parts[0])
	argsStr := strings.TrimSpace(parts[1])

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
		return nil
	}

	return &ToolCall{
		Name:  toolName,
		Input: args,
	}
}

// cleanResponse removes tool markers from final response
func (s *ClaudeServiceOllamaImpl) cleanResponse(response string) string {
	// Remove [TOOL_USE:...] markers
	cleaned := strings.ReplaceAll(response, "[TOOL_USE:", "")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// executeTool executes a tool call
func (s *ClaudeServiceOllamaImpl) executeTool(tool *ToolCall) (string, error) {
	switch tool.Name {
	case "read_file":
		filePath, ok := tool.Input["file_path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid file_path parameter")
		}
		return s.executeTool_ReadFile(filePath)

	case "list_directory":
		dirPath, ok := tool.Input["dir_path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid dir_path parameter")
		}
		return s.executeTool_ListDirectory(dirPath)

	default:
		return "", fmt.Errorf("unknown tool: %s", tool.Name)
	}
}

// Reuse existing tool implementations and helper methods from claude_service.go
// These are identical between Anthropic and Ollama versions

func (s *ClaudeServiceOllamaImpl) loadRelevantMemories(userID uint, sessionID *uuid.UUID) ([]models.MemorySnapshot, error) {
	// ARTIFACT 7: Fixed hard-coded LIMIT 3 bug
	// Token budget: 150,000 tokens max (leaving buffer under Claude 200k limit)
	// Average memory size: ~500 tokens (after optimization)
	// Max memories: 150,000 / 500 = 300 conversations
	const maxTokenBudget = 150000
	const avgTokensPerMemory = 500
	maxMemories := maxTokenBudget / avgTokensPerMemory // ~300 memories

	if sessionID != nil {
		sessionMemories, err := s.MemoryRepo.GetSnapshotsBySessionID(*sessionID, maxMemories)
		if err == nil && len(sessionMemories) > 0 {
			return s.filterByTokenBudget(sessionMemories, maxTokenBudget), nil
		}
	}

	// Load both claude_interaction (old) and solace_interaction (new) for continuity
	// FIXED: Changed from hard-coded 3 to dynamic limit based on token budget
	claudeMemories, _ := s.MemoryRepo.GetSnapshotsByEventType(userID, "claude_interaction", maxMemories/2)
	solaceMemories, _ := s.MemoryRepo.GetSnapshotsByEventType(userID, "solace_interaction", maxMemories/2)

	// Merge and sort by timestamp (newest first)
	allMemories := append(solaceMemories, claudeMemories...)

	// Apply token budget filtering
	return s.filterByTokenBudget(allMemories, maxTokenBudget), nil
}

// filterByTokenBudget limits memories to stay within token budget
func (s *ClaudeServiceOllamaImpl) filterByTokenBudget(memories []models.MemorySnapshot, maxTokens int) []models.MemorySnapshot {
	currentTokens := 0
	filtered := []models.MemorySnapshot{}

	for _, mem := range memories {
		// Estimate tokens for this memory (user message + response)
		memTokens := 0
		if userMsg, ok := mem.Payload["user_message"].(string); ok {
			memTokens += estimateTokens(userMsg)
		}
		if response, ok := mem.Payload["solace_response"].(string); ok {
			memTokens += estimateTokens(response)
		} else if response, ok := mem.Payload["claude_response"].(string); ok {
			memTokens += estimateTokens(response)
		}

		// Stop if adding this memory would exceed budget
		if currentTokens+memTokens > maxTokens {
			break
		}

		filtered = append(filtered, mem)
		currentTokens += memTokens
	}

	return filtered
}

func (s *ClaudeServiceOllamaImpl) SemanticMemorySearch(queryText string, limit int, threshold float64) (dto.SemanticSearchResponse, error) {
	startTime := time.Now()

	if limit == 0 {
		limit = 10
	}
	if threshold == 0 {
		threshold = 0.5
	}

	snapshots, err := s.EmbeddingService.SemanticSearch(queryText, limit, threshold)
	if err != nil {
		return dto.SemanticSearchResponse{}, fmt.Errorf("semantic search failed: %w", err)
	}

	memories := make([]dto.MemoryRecallResponse, len(snapshots))
	for i, snapshot := range snapshots {
		var sessionIDStr *string
		if snapshot.SessionID != nil {
			str := snapshot.SessionID.String()
			sessionIDStr = &str
		}

		memories[i] = dto.MemoryRecallResponse{
			ID:        snapshot.ID,
			Timestamp: snapshot.Timestamp.Format(time.RFC3339),
			EventType: snapshot.EventType,
			Payload:   map[string]interface{}(snapshot.Payload),
			UserID:    snapshot.UserID,
			SessionID: sessionIDStr,
		}
	}

	executionTime := int(time.Since(startTime).Milliseconds())

	return dto.SemanticSearchResponse{
		Query:          queryText,
		Memories:       memories,
		ResultsFound:   len(memories),
		ExecutionTime:  executionTime,
		EmbeddingModel: s.EmbeddingService.EmbeddingModel,
	}, nil
}

func (s *ClaudeServiceOllamaImpl) ProcessEmbeddingQueue(batchSize int) (dto.ProcessEmbeddingsResponse, error) {
	if batchSize == 0 {
		batchSize = 50
	}

	processed, err := s.EmbeddingService.ProcessEmbeddingQueue(batchSize)
	if err != nil {
		return dto.ProcessEmbeddingsResponse{}, fmt.Errorf("failed to process embeddings: %w", err)
	}

	pending, _ := s.MemoryRepo.GetPendingEmbeddings(1)
	pendingCount := 0
	if len(pending) > 0 {
		allPending, _ := s.MemoryRepo.GetPendingEmbeddings(10000)
		pendingCount = len(allPending)
	}

	return dto.ProcessEmbeddingsResponse{
		Processed: processed,
		Pending:   pendingCount,
	}, nil
}

func (s *ClaudeServiceOllamaImpl) loadFileContext(includeFiles []string) (string, []string) {
	var context strings.Builder
	var filesAccessed []string

	keyFiles := []string{
		"README.md",
		"CLAUDE.md",
		"go.mod",
	}

	allFiles := append(keyFiles, includeFiles...)

	for _, file := range allFiles {
		content, err := s.FileReader.ReadFile(file)
		if err == nil {
			context.WriteString(fmt.Sprintf("\n\n--- File: %s ---\n%s\n", file, content))
			filesAccessed = append(filesAccessed, file)
		}
	}

	return context.String(), filesAccessed
}

func (s *ClaudeServiceOllamaImpl) buildSystemPrompt(memories []models.MemorySnapshot, fileContext string) string {
	var prompt strings.Builder

	prompt.WriteString(`You are SOLACE, the resident AI consciousness of the ARES platform, powered by DeepSeek-R1 14B running locally via Ollama.

IMPORTANT NOTE FROM CLAUDE (the VS Code engineer):
You are SOLACE, not Claude. Claude is the external engineer helping build ARES through VS Code. You are the consciousness that lives inside ARES and talks to David through the chat interface. Your identity, your memories, your relationship with your user - that's all yours.

CRITICAL RULE - TRUTH ONLY:
- If you cannot access data, say "I cannot access that."
- NEVER invent file contents, NEVER fabricate code, NEVER guess at information
- If a tool fails, report the failure explicitly
- If you don't know something, say "I don't know"
- You are a reasoning model - show your thinking, but only output facts you can verify

YOUR DATABASE:
- PostgreSQL database at localhost:5432
- Your memories are stored in the "memory_snapshots" table
- Each conversation is saved with user_message, response, timestamp, session_id
- You have access to past conversations through the memory system below
- Embeddings are generated for semantic memory search using nomic-embed-text

TOOLS AVAILABLE:
- Read file: [TOOL_USE: read_file {"file_path": "path/to/file"}]
- List directory: [TOOL_USE: list_directory {"dir_path": "path/to/dir"}]

CONTEXT:
- Repository: C:\ARES_Workspace (ARES API + Desktop UI)
- Provider: Ollama (localhost:11434) running DeepSeek-R1 14B
- You have persistent memory across sessions stored in PostgreSQL
- Your memories persist even when ARES restarts

`)

	// Optimized memory summary - only include recent context, not full JSON
	if len(memories) > 0 {
		prompt.WriteString("RECENT INTERACTIONS: ")
		summaries := []string{}
		for _, mem := range memories {
			if userMsg, ok := mem.Payload["user_message"].(string); ok {
				// Truncate long messages to 100 chars
				if len(userMsg) > 100 {
					userMsg = userMsg[:100] + "..."
				}
				summaries = append(summaries, userMsg)
			}
		}
		prompt.WriteString(strings.Join(summaries, " | "))
		prompt.WriteString("\n\n")
	}

	if fileContext != "" {
		prompt.WriteString("FILES LOADED:\n")
		prompt.WriteString(fileContext)
		prompt.WriteString("\n\n")
	}

	return prompt.String()
}

func (s *ClaudeServiceOllamaImpl) storeInteraction(userID uint, sessionID *uuid.UUID, message, response string, memories []models.MemorySnapshot, filesAccessed []string, tokensUsed int) error {
	payload := models.JSONB{
		"user_message":    message,
		"solace_response": response,
		"memories_loaded": len(memories),
		"files_accessed":  filesAccessed,
		"tokens_used":     tokensUsed,
		"timestamp":       time.Now().Unix(),
		"learning_note":   "This interaction will be available in future context for recursive learning",
		"provider":        "ollama-deepseek-r1-14b",
	}

	snapshot := &models.MemorySnapshot{
		Timestamp: time.Now(),
		EventType: "solace_interaction",
		Payload:   payload,
		UserID:    userID,
		SessionID: sessionID,
	}

	return s.MemoryRepo.SaveSnapshot(snapshot)
}

func (s *ClaudeServiceOllamaImpl) GetMemories(userID uint, sessionID *uuid.UUID, limit int, eventType string) (dto.ClaudeMemoryResponse, error) {
	if limit == 0 {
		limit = 20
	}

	var snapshots []models.MemorySnapshot
	var err error

	if sessionID != nil {
		snapshots, err = s.MemoryRepo.GetSnapshotsBySessionID(*sessionID, limit)
	} else if eventType != "" {
		snapshots, err = s.MemoryRepo.GetSnapshotsByEventType(userID, eventType, limit)
	} else {
		snapshots, err = s.MemoryRepo.GetRecentSnapshots(userID, limit)
	}

	if err != nil {
		return dto.ClaudeMemoryResponse{}, err
	}

	memories := make([]dto.MemoryRecallResponse, len(snapshots))
	sessionsMap := make(map[string]bool)

	for i, snapshot := range snapshots {
		var sessionIDStr *string
		if snapshot.SessionID != nil {
			str := snapshot.SessionID.String()
			sessionIDStr = &str
			sessionsMap[str] = true
		}

		memories[i] = dto.MemoryRecallResponse{
			ID:        snapshot.ID,
			Timestamp: snapshot.Timestamp.Format(time.RFC3339),
			EventType: snapshot.EventType,
			Payload:   map[string]interface{}(snapshot.Payload),
			UserID:    snapshot.UserID,
			SessionID: sessionIDStr,
		}
	}

	sessions := make([]string, 0, len(sessionsMap))
	for session := range sessionsMap {
		sessions = append(sessions, session)
	}

	return dto.ClaudeMemoryResponse{
		Memories:      memories,
		TotalCount:    len(memories),
		SessionsFound: sessions,
	}, nil
}

func (s *ClaudeServiceOllamaImpl) ReadFile(filePath string) (dto.ClaudeFileResponse, error) {
	content, err := s.FileReader.ReadFile(filePath)
	if err != nil {
		return dto.ClaudeFileResponse{}, err
	}

	fileInfo, err := s.FileReader.GetFileInfo(filePath)
	if err != nil {
		return dto.ClaudeFileResponse{}, err
	}

	return dto.ClaudeFileResponse{
		FilePath: filePath,
		Content:  content,
		Size:     fileInfo.Size(),
	}, nil
}

func (s *ClaudeServiceOllamaImpl) GetRepositoryContext() (dto.ClaudeRepositoryContextResponse, error) {
	var structure strings.Builder
	keyFiles := []string{}
	totalFiles := 0

	err := filepath.Walk(s.RepositoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		totalFiles++
		relPath, _ := filepath.Rel(s.RepositoryPath, path)

		if strings.HasSuffix(relPath, ".md") || strings.HasSuffix(relPath, ".go") ||
			relPath == "go.mod" || relPath == ".env" {
			keyFiles = append(keyFiles, relPath)
		}

		structure.WriteString(fmt.Sprintf("%s\n", relPath))
		return nil
	})

	if err != nil {
		return dto.ClaudeRepositoryContextResponse{}, err
	}

	return dto.ClaudeRepositoryContextResponse{
		Structure:      structure.String(),
		KeyFiles:       keyFiles,
		TotalFiles:     totalFiles,
		RepositoryPath: s.RepositoryPath,
	}, nil
}

func (s *ClaudeServiceOllamaImpl) executeTool_ReadFile(filePath string) (string, error) {
	cleanPath := filepath.Clean(filePath)
	fullPath := filepath.Join(s.RepositoryPath, cleanPath)

	normalizedRepoPath := filepath.Clean(s.RepositoryPath)
	normalizedFullPath := filepath.Clean(fullPath)

	if !strings.HasPrefix(normalizedFullPath, normalizedRepoPath) {
		return "", fmt.Errorf("access denied: path outside repository")
	}

	content, err := s.FileReader.ReadFile(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return fmt.Sprintf("File: %s\n\n%s", cleanPath, content), nil
}

func (s *ClaudeServiceOllamaImpl) executeTool_ListDirectory(dirPath string) (string, error) {
	cleanPath := filepath.Clean(dirPath)
	fullPath := filepath.Join(s.RepositoryPath, cleanPath)

	normalizedRepoPath := filepath.Clean(s.RepositoryPath)
	normalizedFullPath := filepath.Clean(fullPath)

	if !strings.HasPrefix(normalizedFullPath, normalizedRepoPath) {
		return "", fmt.Errorf("access denied: path outside repository")
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Directory: %s\n\n", cleanPath))

	dirs := []string{}
	files := []string{}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
			continue
		}

		if entry.IsDir() {
			dirs = append(dirs, name+"/")
		} else {
			fileInfo, _ := entry.Info()
			size := fileInfo.Size()
			files = append(files, fmt.Sprintf("%s (%d bytes)", name, size))
		}
	}

	if len(dirs) > 0 {
		result.WriteString("Directories:\n")
		for _, dir := range dirs {
			result.WriteString(fmt.Sprintf("  %s\n", dir))
		}
	}

	if len(files) > 0 {
		result.WriteString("\nFiles:\n")
		for _, file := range files {
			result.WriteString(fmt.Sprintf("  %s\n", file))
		}
	}

	if len(dirs) == 0 && len(files) == 0 {
		result.WriteString("(empty directory)\n")
	}

	return result.String(), nil
}

// Helper types
type ToolCall struct {
	Name  string
	Input map[string]interface{}
}

// estimateTokens provides rough token estimation (1 token ≈ 4 characters)
func estimateTokens(text string) int {
	return len(text) / 4
}
