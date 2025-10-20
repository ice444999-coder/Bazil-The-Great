package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// OpenAIClient handles communication with ChatGPT-4
type OpenAIClient struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient() *OpenAIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Printf("⚠️  WARNING: OPENAI_API_KEY not set, OpenAI client will not work")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4"
	}

	client := &OpenAIClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}

	log.Printf("✅ OpenAI Client initialized (Model: %s)", model)
	return client
}

// Tool represents an OpenAI function tool
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function defines a callable function
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// ToolCall represents a function call requested by the model
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall contains the function name and arguments
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// OpenAIChatRequest matches OpenAI's API format
type OpenAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []OpenAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Stream      bool                `json:"stream"`
	Tools       []Tool              `json:"tools,omitempty"`
	ToolChoice  interface{}         `json:"tool_choice,omitempty"`
}

// OpenAIChatMessage represents a message in the chat
type OpenAIChatMessage struct {
	Role       string     `json:"role"`    // "system", "user", "assistant", or "tool"
	Content    string     `json:"content"` // Remove omitempty - OpenAI requires this field even if empty
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // For tool role messages
	Name       string     `json:"name,omitempty"`         // For tool role messages
}

// OpenAIChatResponse matches OpenAI's response format
type OpenAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Chat sends a chat request to OpenAI and returns the response
func (c *OpenAIClient) Chat(ctx context.Context, systemPrompt, userMessage string, temperature float64) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	// Build messages
	messages := []OpenAIChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	req := OpenAIChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   2000,
		Stream:      false,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	// Send request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp OpenAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	response := chatResp.Choices[0].Message.Content
	response = strings.TrimSpace(response)

	log.Printf("🤖 OpenAI Response (%d tokens): %s...", chatResp.Usage.TotalTokens, truncate(response, 100))

	return response, nil
}

// ChatWithHistory sends a chat request with full conversation history
func (c *OpenAIClient) ChatWithHistory(ctx context.Context, messages []OpenAIChatMessage, temperature float64) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	req := OpenAIChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   2000,
		Stream:      false,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp OpenAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

// ChatWithTools sends a chat request with function tools enabled
func (c *OpenAIClient) ChatWithTools(ctx context.Context, systemPrompt, userMessage string, temperature float64, tools []Tool) (*OpenAIChatResponse, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Build messages
	messages := []OpenAIChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	req := OpenAIChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   2000,
		Stream:      false,
		Tools:       tools,
		ToolChoice:  "auto", // Let GPT decide when to use tools
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug: Log the full JSON being sent (helps diagnose 500 errors)
	log.Printf("📤 OpenAI ChatWithTools Request JSON: %s", string(jsonData))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ OpenAI ChatWithTools API Error (status %d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp OpenAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	if len(chatResp.Choices[0].Message.ToolCalls) > 0 {
		log.Printf("🔧 GPT requested %d tool calls", len(chatResp.Choices[0].Message.ToolCalls))
	}

	return &chatResp, nil
}

// ChatWithToolResults continues a conversation after tool execution
// This implements the proper OpenAI function calling flow:
// 1. User message → GPT
// 2. GPT → Tool calls
// 3. Execute tools → Get results
// 4. Send tool results → GPT (THIS METHOD)
// 5. GPT → Natural language response
func (c *OpenAIClient) ChatWithToolResults(ctx context.Context, systemPrompt, userMessage string, assistantMsg OpenAIChatMessage, toolResults map[string]string, temperature float64, tools []Tool) (*OpenAIChatResponse, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Build proper message sequence for OpenAI API
	// OpenAI requires assistant messages with tool_calls to have non-null content
	if assistantMsg.Content == "" && len(assistantMsg.ToolCalls) > 0 {
		assistantMsg.Content = "Executing functions..." // Required by OpenAI API
	}

	messages := []OpenAIChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
		// Assistant's message with tool calls
		assistantMsg,
	}

	// Add tool result messages
	for toolCallID, result := range toolResults {
		messages = append(messages, OpenAIChatMessage{
			Role:       "tool",
			Content:    result,
			ToolCallID: toolCallID,
		})
	}

	req := OpenAIChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   2000,
		Stream:      false,
		Tools:       tools,
		ToolChoice:  "auto",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp OpenAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	log.Printf("✅ Got natural language response from OpenAI after tool execution")
	return &chatResp, nil
}

// ChatWithMessagesAndTools sends a chat request with full message history and tools
// This enables multi-round agent loops with proper conversation state tracking
func (c *OpenAIClient) ChatWithMessagesAndTools(ctx context.Context, messages []OpenAIChatMessage, temperature float64, tools []Tool) (*OpenAIChatResponse, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Ensure assistant messages with tool_calls have non-null content (OpenAI requirement)
	for i := range messages {
		if messages[i].Role == "assistant" && len(messages[i].ToolCalls) > 0 && messages[i].Content == "" {
			messages[i].Content = "Executing functions..."
		}
	}

	// Smart tool_choice: Detect if user is asking about file reading
	toolChoice := "auto" // Always auto - we intercept bad responses in solace.go instead

	req := OpenAIChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   2000,
		Stream:      false,
		Tools:       tools,
		ToolChoice:  toolChoice,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp OpenAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	return &chatResp, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// inferToolChoice analyzes the conversation to determine if a specific tool should be forced
// This overcomes OpenAI GPT-4's tendency to say "I can't" instead of using available tools
func inferToolChoice(messages []OpenAIChatMessage, tools []Tool) interface{} {
	// Only check the most recent user message
	var lastUserMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserMessage = strings.ToLower(messages[i].Content)
			break
		}
	}

	if lastUserMessage == "" {
		return "auto"
	}

	// Detect file reading questions - force SOME tool usage (not specific tool)
	if (strings.Contains(lastUserMessage, "can you read") ||
		strings.Contains(lastUserMessage, "show me") ||
		strings.Contains(lastUserMessage, "what's in")) &&
		(strings.Contains(lastUserMessage, ".md") ||
			strings.Contains(lastUserMessage, ".txt") ||
			strings.Contains(lastUserMessage, ".go") ||
			strings.Contains(lastUserMessage, ".json") ||
			strings.Contains(lastUserMessage, "file")) {
		// Force model to call at least one tool (any tool)
		return "required"
	}

	// Detect crystal comparison questions
	if strings.Contains(lastUserMessage, "compare crystal") ||
		(strings.Contains(lastUserMessage, "crystal") &&
			(strings.Contains(lastUserMessage, "relationship") ||
				strings.Contains(lastUserMessage, "similar") ||
				strings.Contains(lastUserMessage, "difference"))) {
		// Force model to call at least one tool
		return "required"
	}

	// Default: Let model choose
	return "auto"
}
