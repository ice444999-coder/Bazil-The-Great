package dto

type ClaudeChatRequest struct {
	Message   string  `json:"message" binding:"required"`
	SessionID *string `json:"session_id,omitempty"`
	// Optional: request specific files to include in context
	IncludeFiles []string `json:"include_files,omitempty"`
	// Optional: max tokens for response
	MaxTokens int `json:"max_tokens,omitempty"`
}

type ClaudeChatResponse struct {
	Message         string                 `json:"message"`
	Response        string                 `json:"response"`
	SessionID       string                 `json:"session_id"`
	MemoriesLoaded  int                    `json:"memories_loaded"`
	FilesAccessed   []string               `json:"files_accessed,omitempty"`
	ThinkingProcess map[string]interface{} `json:"thinking_process,omitempty"`
	TokensUsed      int                    `json:"tokens_used,omitempty"`
}

type ClaudeMemoryRequest struct {
	SessionID *string `json:"session_id,omitempty"`
	Limit     int     `json:"limit,omitempty"`
	EventType string  `json:"event_type,omitempty"`
}

type ClaudeMemoryResponse struct {
	Memories      []MemoryRecallResponse `json:"memories"`
	TotalCount    int                    `json:"total_count"`
	SessionsFound []string               `json:"sessions_found,omitempty"`
}

type ClaudeFileRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

type ClaudeFileResponse struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
	Size     int64  `json:"size"`
}

type ClaudeRepositoryContextResponse struct {
	Structure      string   `json:"structure"`
	KeyFiles       []string `json:"key_files"`
	RecentCommits  []string `json:"recent_commits,omitempty"`
	TotalFiles     int      `json:"total_files"`
	RepositoryPath string   `json:"repository_path"`
}

// Semantic Search DTOs
type SemanticSearchRequest struct {
	Query     string  `json:"query" binding:"required"`
	Limit     int     `json:"limit,omitempty"`     // Default 10
	Threshold float64 `json:"threshold,omitempty"` // Default 0.5
}

type SemanticSearchResponse struct {
	Query          string                 `json:"query"`
	Memories       []MemoryRecallResponse `json:"memories"`
	ResultsFound   int                    `json:"results_found"`
	ExecutionTime  int                    `json:"execution_time_ms"`
	EmbeddingModel string                 `json:"embedding_model"`
}

// Embedding Queue DTOs
type ProcessEmbeddingsRequest struct {
	BatchSize int `json:"batch_size,omitempty"` // Default 50
}

type ProcessEmbeddingsResponse struct {
	Processed int      `json:"processed"`
	Failed    int      `json:"failed"`
	Pending   int      `json:"pending"`
	Errors    []string `json:"errors,omitempty"`
}
