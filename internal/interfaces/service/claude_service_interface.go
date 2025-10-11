package service

import (
	"ares_api/internal/api/dto"

	"github.com/google/uuid"
)

type ClaudeService interface {
	// Chat with Claude with full memory and file system context
	Chat(userID uint, message string, sessionID *uuid.UUID, includeFiles []string, maxTokens int) (dto.ClaudeChatResponse, error)

	// Get Claude's memories for a user/session
	GetMemories(userID uint, sessionID *uuid.UUID, limit int, eventType string) (dto.ClaudeMemoryResponse, error)

	// Read a file from the repository
	ReadFile(filePath string) (dto.ClaudeFileResponse, error)

	// Get repository context overview
	GetRepositoryContext() (dto.ClaudeRepositoryContextResponse, error)

	// Semantic search through memories (INTELLIGENT RETRIEVAL)
	SemanticMemorySearch(queryText string, limit int, threshold float64) (dto.SemanticSearchResponse, error)

	// Process pending embeddings
	ProcessEmbeddingQueue(batchSize int) (dto.ProcessEmbeddingsResponse, error)
}
