package llm

import "time"

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"` // Message content
}

// ChatRequest represents a request to the LLM
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse represents a response from the LLM
type ChatResponse struct {
	Model     string    `json:"model"`
	Message   Message   `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Done      bool      `json:"done"`
	DoneReason string   `json:"done_reason,omitempty"`
	
	// Token usage tracking
	PromptTokens     int `json:"prompt_eval_count,omitempty"`
	CompletionTokens int `json:"eval_count,omitempty"`
	TotalTokens      int `json:"-"` // Calculated
}

// StreamCallback is called for each chunk in streaming mode
type StreamCallback func(chunk string, done bool) error

// HealthStatus represents LLM service health
type HealthStatus struct {
	Healthy      bool          `json:"healthy"`
	Latency      time.Duration `json:"latency_ms"`
	ModelLoaded  bool          `json:"model_loaded"`
	ErrorMessage string        `json:"error_message,omitempty"`
	CheckedAt    time.Time     `json:"checked_at"`
}

// ContextWindow manages token budget
type ContextWindow struct {
	MaxTokens       int       `json:"max_tokens"`        // 150000 for DeepSeek-R1 14B
	UsedTokens      int       `json:"used_tokens"`
	AvailableTokens int       `json:"available_tokens"`
	Messages        []Message `json:"messages"`
}

// TraceContext for request tracking
type TraceContext struct {
	TraceID   string    `json:"trace_id"`
	UserID    uint      `json:"user_id"`
	SessionID string    `json:"session_id"`
	StartTime time.Time `json:"start_time"`
	Component string    `json:"component"` // "OLLAMA", "MEMORY", "TRADING", etc.
}
