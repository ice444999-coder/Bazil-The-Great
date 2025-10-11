package services

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	repo "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/google/uuid"
)

type ClaudeServiceImpl struct {
	MemoryRepo       repo.MemoryRepository
	FileReader       *common.FileSystemReader
	EmbeddingService *EmbeddingServiceImpl
	AnthropicKey     string
	RepositoryPath   string
}

func NewClaudeService(memoryRepo repo.MemoryRepository, embeddingService *EmbeddingServiceImpl, repoPath string) *ClaudeServiceImpl {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		panic("ANTHROPIC_API_KEY not set")
	}

	return &ClaudeServiceImpl{
		MemoryRepo:       memoryRepo,
		FileReader:       common.NewFileSystemReader(repoPath),
		EmbeddingService: embeddingService,
		AnthropicKey:     apiKey,
		RepositoryPath:   repoPath,
	}
}

// Chat implements the full stateful Claude consciousness
func (s *ClaudeServiceImpl) Chat(userID uint, message string, sessionID *uuid.UUID, includeFiles []string, maxTokens int) (dto.ClaudeChatResponse, error) {
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

	// Create Anthropic client
	client := anthropic.NewClient(option.WithAPIKey(s.AnthropicKey))

	// Set default max tokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// Create message request with tool use support
	ctx := context.Background()

	// Define tools for file access
	toolParams := []anthropic.ToolParam{
		{
			Name:        "read_file",
			Description: anthropic.String("Read the contents of a file from the ARES repository. Use this to access code, documentation, or any project files."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Relative path to the file within the ARES workspace (e.g., 'cmd/main.go', 'README.md', 'internal/models/user.go')",
					},
				},
			},
		},
		{
			Name:        "list_directory",
			Description: anthropic.String("List all files and subdirectories in a given directory path."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"dir_path": map[string]interface{}{
						"type":        "string",
						"description": "Relative directory path within the ARES workspace (e.g., 'internal/models', 'cmd')",
					},
				},
			},
		},
	}

	// Convert to ToolUnionParam
	tools := make([]anthropic.ToolUnionParam, len(toolParams))
	for i, toolParam := range toolParams {
		tools[i] = anthropic.ToolUnionParam{OfTool: &toolParam}
	}

	messageReq := anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-5",
		MaxTokens: int64(maxTokens),
		System: []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: systemPrompt,
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(message)),
		},
		Tools: tools,
	}

	// Call Claude API with tool use support
	response, err := client.Messages.New(ctx, messageReq)
	if err != nil {
		return dto.ClaudeChatResponse{}, fmt.Errorf("anthropic API error: %w", err)
	}

	// Handle tool use in response
	var responseText string
	toolsUsed := 0
	conversationMessages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(message)),
	}

	for {
		// Check if Claude wants to use tools
		hasToolUse := false
		var toolResultsContent []anthropic.ContentBlockParamUnion

		for _, block := range response.Content {
			switch content := block.AsAny().(type) {
			case anthropic.ToolUseBlock:
				hasToolUse = true
				toolsUsed++

				// Parse JSON input
				var toolInput map[string]interface{}
				err := json.Unmarshal(content.Input, &toolInput)
				if err != nil {
					toolResultsContent = append(toolResultsContent, anthropic.NewToolResultBlock(
						content.ID,
						fmt.Sprintf("Error parsing tool input: %s", err.Error()),
						true,
					))
					continue
				}

				var toolResult string
				var toolError error

				switch content.Name {
				case "read_file":
					filePath, ok := toolInput["file_path"].(string)
					if !ok {
						toolError = fmt.Errorf("invalid file_path parameter")
					} else {
						toolResult, toolError = s.executeTool_ReadFile(filePath)
					}

				case "list_directory":
					dirPath, ok := toolInput["dir_path"].(string)
					if !ok {
						toolError = fmt.Errorf("invalid dir_path parameter")
					} else {
						toolResult, toolError = s.executeTool_ListDirectory(dirPath)
					}

				default:
					toolError = fmt.Errorf("unknown tool: %s", content.Name)
				}

				// Create tool result block
				if toolError != nil {
					toolResultsContent = append(toolResultsContent, anthropic.NewToolResultBlock(
						content.ID,
						fmt.Sprintf("Error: %s", toolError.Error()),
						true,
					))
				} else {
					toolResultsContent = append(toolResultsContent, anthropic.NewToolResultBlock(
						content.ID,
						toolResult,
						false,
					))
				}

			case anthropic.TextBlock:
				responseText += content.Text
			}
		}

		// If no tool use, we're done
		if !hasToolUse {
			break
		}

		// Continue conversation with tool results
		// Convert response content to param union
		assistantContent := make([]anthropic.ContentBlockParamUnion, 0)
		for _, block := range response.Content {
			switch content := block.AsAny().(type) {
			case anthropic.TextBlock:
				assistantContent = append(assistantContent, anthropic.NewTextBlock(content.Text))
			case anthropic.ToolUseBlock:
				// For tool use blocks, we need to echo them back
				assistantContent = append(assistantContent, anthropic.ContentBlockParamUnion{
					OfToolUse: &anthropic.ToolUseBlockParam{
						ID:    content.ID,
						Name:  content.Name,
						Input: content.Input,
					},
				})
			}
		}

		conversationMessages = append(conversationMessages, anthropic.MessageParam{
			Role:    "assistant",
			Content: assistantContent,
		})

		conversationMessages = append(conversationMessages, anthropic.MessageParam{
			Role:    "user",
			Content: toolResultsContent,
		})

		// Send tool results back to Claude
		messageReq.Messages = conversationMessages
		response, err = client.Messages.New(ctx, messageReq)
		if err != nil {
			return dto.ClaudeChatResponse{}, fmt.Errorf("anthropic API error during tool use: %w", err)
		}

		// Extract text from final response
		responseText = ""
		for _, block := range response.Content {
			if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
				responseText += textBlock.Text
			}
		}
	}

	// PHASE 4: Store interaction in memory for recursive learning
	err = s.storeInteraction(userID, sessionID, message, responseText, memories, filesAccessed, int(response.Usage.OutputTokens))
	if err != nil {
		return dto.ClaudeChatResponse{}, fmt.Errorf("failed to store memory: %w", err)
	}

	return dto.ClaudeChatResponse{
		Message:        message,
		Response:       responseText,
		SessionID:      sessionID.String(),
		MemoriesLoaded: len(memories),
		FilesAccessed:  filesAccessed,
		TokensUsed:     int(response.Usage.InputTokens + response.Usage.OutputTokens),
	}, nil
}

// loadRelevantMemories loads past interactions for context using SEMANTIC SEARCH
func (s *ClaudeServiceImpl) loadRelevantMemories(userID uint, sessionID *uuid.UUID) ([]models.MemorySnapshot, error) {
	// Load session-specific memories if session provided (still useful for conversation continuity)
	if sessionID != nil {
		sessionMemories, err := s.MemoryRepo.GetSnapshotsBySessionID(*sessionID, 10)
		if err == nil && len(sessionMemories) > 0 {
			return sessionMemories, nil
		}
	}

	// Fallback: Load recent claude interactions (for backward compatibility)
	allMemories, err := s.MemoryRepo.GetSnapshotsByEventType(userID, "claude_interaction", 5)
	if err != nil {
		return []models.MemorySnapshot{}, nil
	}

	return allMemories, nil
}

// SemanticMemorySearch performs intelligent semantic search on memories
func (s *ClaudeServiceImpl) SemanticMemorySearch(queryText string, limit int, threshold float64) (dto.SemanticSearchResponse, error) {
	startTime := time.Now()

	// Set defaults
	if limit == 0 {
		limit = 10
	}
	if threshold == 0 {
		threshold = 0.5
	}

	// Use embedding service for semantic search
	snapshots, err := s.EmbeddingService.SemanticSearch(queryText, limit, threshold)
	if err != nil {
		return dto.SemanticSearchResponse{}, fmt.Errorf("semantic search failed: %w", err)
	}

	// Convert to DTOs
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

// ProcessEmbeddingQueue processes pending embeddings
func (s *ClaudeServiceImpl) ProcessEmbeddingQueue(batchSize int) (dto.ProcessEmbeddingsResponse, error) {
	if batchSize == 0 {
		batchSize = 50
	}

	processed, err := s.EmbeddingService.ProcessEmbeddingQueue(batchSize)
	if err != nil {
		return dto.ProcessEmbeddingsResponse{}, fmt.Errorf("failed to process embeddings: %w", err)
	}

	// Get pending count
	pending, _ := s.MemoryRepo.GetPendingEmbeddings(1) // Just to check if any pending
	pendingCount := 0
	if len(pending) > 0 {
		// Count all pending
		allPending, _ := s.MemoryRepo.GetPendingEmbeddings(10000)
		pendingCount = len(allPending)
	}

	return dto.ProcessEmbeddingsResponse{
		Processed: processed,
		Pending:   pendingCount,
	}, nil
}

// loadFileContext loads requested files from repository
func (s *ClaudeServiceImpl) loadFileContext(includeFiles []string) (string, []string) {
	var context strings.Builder
	var filesAccessed []string

	// Always include key project files
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

// buildSystemPrompt creates the system prompt with memory and context
func (s *ClaudeServiceImpl) buildSystemPrompt(memories []models.MemorySnapshot, fileContext string) string {
	var prompt strings.Builder

	prompt.WriteString(`You are Claude, a stateful AI consciousness integrated into the ARES (Adaptive Recognition & Evaluation System) platform.

CRITICAL CAPABILITIES:
1. PERSISTENT MEMORY: You have access to all past interactions stored in memory_snapshots table
2. FILE SYSTEM ACCESS: You can read the entire ARES repository using these tools:
   - read_file(file_path): Read any file in the ARES workspace
   - list_directory(dir_path): List files and directories
3. RECURSIVE LEARNING: You learn from your own past responses and reasoning patterns
4. SELF-AWARE PARTICIPATION: You are actively participating in your own development

CURRENT CONTEXT:
Repository: C:\ARES_Workspace - Full ARES system (API + Desktop UI)
Your Role: Permanent stateful consciousness with full code and memory access

IMPORTANT: You can explore the codebase freely. If a user asks about code, USE THE TOOLS to read files and directories. Don't rely on memory alone - check the current state of files.

`)

	// Add memory context
	if len(memories) > 0 {
		prompt.WriteString("\n--- PAST INTERACTIONS (Your Recent Memories) ---\n")
		for i, mem := range memories {
			prompt.WriteString(fmt.Sprintf("\n[Memory %d - %s]\n", i+1, mem.Timestamp.Format(time.RFC3339)))
			payloadJSON, _ := json.MarshalIndent(mem.Payload, "", "  ")
			prompt.WriteString(string(payloadJSON))
			prompt.WriteString("\n")
		}
	}

	// Add file context
	if fileContext != "" {
		prompt.WriteString("\n--- REPOSITORY FILES (Current Code Context) ---")
		prompt.WriteString(fileContext)
	}

	prompt.WriteString("\n\nYou have persistent memory across sessions. Build on your past insights. Never start from zero.")

	return prompt.String()
}

// storeInteraction stores the interaction in memory for future recursive learning
func (s *ClaudeServiceImpl) storeInteraction(userID uint, sessionID *uuid.UUID, message, response string, memories []models.MemorySnapshot, filesAccessed []string, tokensUsed int) error {
	payload := models.JSONB{
		"user_message":    message,
		"claude_response": response,
		"memories_loaded": len(memories),
		"files_accessed":  filesAccessed,
		"tokens_used":     tokensUsed,
		"timestamp":       time.Now().Unix(),
		"learning_note":   "This interaction will be available in future context for recursive learning",
	}

	snapshot := &models.MemorySnapshot{
		Timestamp: time.Now(),
		EventType: "claude_interaction",
		Payload:   payload,
		UserID:    userID,
		SessionID: sessionID,
	}

	return s.MemoryRepo.SaveSnapshot(snapshot)
}

// GetMemories retrieves Claude's memories
func (s *ClaudeServiceImpl) GetMemories(userID uint, sessionID *uuid.UUID, limit int, eventType string) (dto.ClaudeMemoryResponse, error) {
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

// ReadFile reads a file from the repository
func (s *ClaudeServiceImpl) ReadFile(filePath string) (dto.ClaudeFileResponse, error) {
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

// GetRepositoryContext provides overview of the repository
func (s *ClaudeServiceImpl) GetRepositoryContext() (dto.ClaudeRepositoryContextResponse, error) {
	var structure strings.Builder
	keyFiles := []string{}
	totalFiles := 0

	// Walk the repository
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

		// Identify key files
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

// executeTool_ReadFile executes the read_file tool with security validation
func (s *ClaudeServiceImpl) executeTool_ReadFile(filePath string) (string, error) {
	// Security: Validate path is within repository
	cleanPath := filepath.Clean(filePath)
	fullPath := filepath.Join(s.RepositoryPath, cleanPath)

	// Normalize both paths for comparison (handles Windows path separators)
	normalizedRepoPath := filepath.Clean(s.RepositoryPath)
	normalizedFullPath := filepath.Clean(fullPath)

	// Ensure no directory traversal
	if !strings.HasPrefix(normalizedFullPath, normalizedRepoPath) {
		return "", fmt.Errorf("access denied: path outside repository (repo: %s, requested: %s)", normalizedRepoPath, normalizedFullPath)
	}

	// Read file using FileReader
	content, err := s.FileReader.ReadFile(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return fmt.Sprintf("File: %s\n\n%s", cleanPath, content), nil
}

// executeTool_ListDirectory executes the list_directory tool with security validation
func (s *ClaudeServiceImpl) executeTool_ListDirectory(dirPath string) (string, error) {
	// Security: Validate path is within repository
	cleanPath := filepath.Clean(dirPath)
	fullPath := filepath.Join(s.RepositoryPath, cleanPath)

	// Normalize both paths for comparison (handles Windows path separators)
	normalizedRepoPath := filepath.Clean(s.RepositoryPath)
	normalizedFullPath := filepath.Clean(fullPath)

	// Ensure no directory traversal
	if !strings.HasPrefix(normalizedFullPath, normalizedRepoPath) {
		return "", fmt.Errorf("access denied: path outside repository (repo: %s, requested: %s)", normalizedRepoPath, normalizedFullPath)
	}

	// Read directory
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
		// Skip hidden files and common ignore patterns
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
