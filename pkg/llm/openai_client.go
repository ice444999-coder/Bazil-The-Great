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

// OpenAIChatRequest matches OpenAI's API format
type OpenAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []OpenAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Stream      bool                `json:"stream"`
}

// OpenAIChatMessage represents a message in the chat
type OpenAIChatMessage struct {
	Role    string `json:"role"` // "system", "user", or "assistant"
	Content string `json:"content"`
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
			Role    string `json:"role"`
			Content string `json:"content"`
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
