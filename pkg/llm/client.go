/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
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
	"sync"
	"time"
)

const (
	DefaultModel       = "deepseek-r1:14b"
	DefaultTimeout     = 5 * time.Minute
	DefaultMaxRetries  = 3
	DefaultContextSize = 150000 // DeepSeek-R1 14B supports up to 128k, we use 150k with safety margin
	
	// Temperature settings
	TempTrading  = 0.3 // Lower for deterministic trading decisions
	TempCoding   = 0.7 // Higher for creative code generation
	TempGeneral  = 0.5 // Balanced for general chat
	
	// Circuit breaker settings
	CircuitBreakerThreshold  = 5              // Failures before opening circuit
	CircuitBreakerTimeout    = 30 * time.Second // Time before half-open
	CircuitBreakerResetTime  = 60 * time.Second // Time before closing circuit
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota // Normal operation
	CircuitOpen                        // Failing, reject requests
	CircuitHalfOpen                    // Testing if service recovered
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu            sync.RWMutex
	state         CircuitState
	failures      int
	lastFailTime  time.Time
	lastSuccessTime time.Time
	nextRetryTime time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state: CircuitClosed,
	}
}

// CanAttempt checks if a request can be attempted
func (cb *CircuitBreaker) CanAttempt() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if we should transition to half-open
		if time.Now().After(cb.nextRetryTime) {
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures = 0
	cb.lastSuccessTime = time.Now()
	
	if cb.state == CircuitHalfOpen {
		log.Printf("✅ Circuit breaker: Service recovered, closing circuit")
		cb.state = CircuitClosed
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures++
	cb.lastFailTime = time.Now()
	
	if cb.state == CircuitClosed && cb.failures >= CircuitBreakerThreshold {
		log.Printf("🚨 Circuit breaker: Threshold reached (%d failures), opening circuit", cb.failures)
		cb.state = CircuitOpen
		cb.nextRetryTime = time.Now().Add(CircuitBreakerTimeout)
	} else if cb.state == CircuitHalfOpen {
		log.Printf("🚨 Circuit breaker: Half-open test failed, reopening circuit")
		cb.state = CircuitOpen
		cb.nextRetryTime = time.Now().Add(CircuitBreakerTimeout)
	}
}

// TransitionToHalfOpen transitions from open to half-open
func (cb *CircuitBreaker) TransitionToHalfOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if cb.state == CircuitOpen && time.Now().After(cb.nextRetryTime) {
		log.Printf("🔄 Circuit breaker: Transitioning to half-open (testing recovery)")
		cb.state = CircuitHalfOpen
	}
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Client manages communication with DeepSeek-R1 via Ollama
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Model      string
	
	// Configuration
	MaxRetries  int
	ContextSize int
	
	// Monitoring
	requestCount  int64
	errorCount    int64
	lastHealthCheck time.Time
	isHealthy     bool
	
	// Circuit breaker
	circuitBreaker *CircuitBreaker
	mu             sync.RWMutex
}

// NewClient creates a new LLM client with proper configuration
func NewClient() *Client {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434/api"
	}
	
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = DefaultModel
	}
	
	// Create HTTP client with timeout and connection pooling
	httpClient := &http.Client{
		Timeout: DefaultTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	
	client := &Client{
		BaseURL:        baseURL,
		HTTPClient:     httpClient,
		Model:          model,
		MaxRetries:     DefaultMaxRetries,
		ContextSize:    DefaultContextSize,
		isHealthy:      false,
		circuitBreaker: NewCircuitBreaker(),
	}
	
	// Verify model availability on init
	if err := client.verifyModel(); err != nil {
		log.Printf("⚠️  WARNING: Model verification failed: %v", err)
		log.Printf("    Make sure '%s' is pulled: ollama pull %s", model, model)
	} else {
		log.Printf("✅ DeepSeek-R1 14B model loaded and ready")
	}
	
	return client
}

// verifyModel checks if the model is available
func (c *Client) verifyModel() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Try a simple generation to verify model is loaded
	testReq := ChatRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "user", Content: "test"},
		},
		Stream: false,
	}
	
	_, err := c.generateWithRetry(ctx, testReq)
	if err != nil {
		return fmt.Errorf("model verification failed: %w", err)
	}
	
	c.isHealthy = true
	return nil
}

// Generate sends a chat request and returns the response
func (c *Client) Generate(ctx context.Context, messages []Message, temperature float64) (string, error) {
	traceID := generateTraceID()
	startTime := time.Now()
	
	log.Printf("[%s] LLM Request: %d messages, temp=%.2f", traceID, len(messages), temperature)
	
	// Enforce system prompt with <think> tag
	messagesWithSystemPrompt := c.ensureSystemPrompt(messages)
	
	req := ChatRequest{
		Model:       c.Model,
		Messages:    messagesWithSystemPrompt,
		Stream:      false,
		Temperature: temperature,
		MaxTokens:   c.ContextSize,
	}
	
	resp, err := c.generateWithRetry(ctx, req)
	if err != nil {
		c.errorCount++
		return "", fmt.Errorf("[%s] generation failed: %w", traceID, err)
	}
	
	c.requestCount++
	latency := time.Since(startTime)
	log.Printf("[%s] LLM Response: %d tokens, latency=%v", traceID, resp.TotalTokens, latency)
	
	return resp.Message.Content, nil
}

// generateWithRetry implements retry logic with exponential backoff and circuit breaker
func (c *Client) generateWithRetry(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Check circuit breaker
	if !c.circuitBreaker.CanAttempt() {
		state := c.circuitBreaker.GetState()
		return nil, fmt.Errorf("circuit breaker %v: Ollama service is currently unavailable, please try again later", state)
	}
	
	// Transition to half-open if needed
	c.circuitBreaker.TransitionToHalfOpen()
	
	var lastErr error
	
	for attempt := 1; attempt <= c.MaxRetries; attempt++ {
		resp, err := c.doGenerate(ctx, req)
		if err == nil {
			c.circuitBreaker.RecordSuccess()
			return resp, nil
		}
		
		lastErr = err
		
		if attempt < c.MaxRetries {
			backoff := time.Duration(attempt*attempt) * time.Second
			log.Printf("⚠️  Attempt %d failed, retrying in %v: %v", attempt, backoff, err)
			
			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				c.circuitBreaker.RecordFailure()
				return nil, ctx.Err()
			}
		}
	}
	
	c.circuitBreaker.RecordFailure()
	return nil, fmt.Errorf("all %d attempts failed, last error: %w", c.MaxRetries, lastErr)
}

// doGenerate performs the actual HTTP request
func (c *Client) doGenerate(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat", c.BaseURL), bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var chatResp ChatResponse
	decoder := json.NewDecoder(resp.Body)
	
	// Ollama streams even with stream=false, we need to read all chunks
	var fullResponse string
	for decoder.More() {
		var chunk ChatResponse
		if err := decoder.Decode(&chunk); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		fullResponse += chunk.Message.Content
		chatResp = chunk // Keep last chunk for metadata
	}
	
	chatResp.Message.Content = fullResponse
	chatResp.TotalTokens = chatResp.PromptTokens + chatResp.CompletionTokens
	
	return &chatResp, nil
}

// Stream generates a response with streaming
func (c *Client) Stream(ctx context.Context, messages []Message, temperature float64, callback StreamCallback) error {
	traceID := generateTraceID()
	log.Printf("[%s] LLM Stream Request: %d messages", traceID, len(messages))
	
	messagesWithSystemPrompt := c.ensureSystemPrompt(messages)
	
	req := ChatRequest{
		Model:       c.Model,
		Messages:    messagesWithSystemPrompt,
		Stream:      true,
		Temperature: temperature,
	}
	
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat", c.BaseURL), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	// Stream chunks
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var chunk ChatResponse
		if err := decoder.Decode(&chunk); err != nil {
			return fmt.Errorf("failed to decode chunk: %w", err)
		}
		
		if err := callback(chunk.Message.Content, chunk.Done); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}
		
		if chunk.Done {
			break
		}
	}
	
	return nil
}

// Health checks if the LLM service is healthy
func (c *Client) Health(ctx context.Context) (*HealthStatus, error) {
	startTime := time.Now()
	
	// Simple ping request
	testReq := ChatRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "user", Content: "ping"},
		},
		Stream: false,
	}
	
	_, err := c.doGenerate(ctx, testReq)
	latency := time.Since(startTime)
	
	status := &HealthStatus{
		Healthy:     err == nil,
		Latency:     latency,
		ModelLoaded: c.isHealthy,
		CheckedAt:   time.Now(),
	}
	
	if err != nil {
		status.ErrorMessage = err.Error()
	}
	
	c.lastHealthCheck = time.Now()
	c.isHealthy = status.Healthy
	
	return status, err
}

// ensureSystemPrompt adds the system prompt enforcing <think> tags
func (c *Client) ensureSystemPrompt(messages []Message) []Message {
	systemPrompt := `You are ARES (Autonomous Recognition & Execution System), an advanced AI assistant.

CRITICAL INSTRUCTIONS:
1. Always show your reasoning inside <think> tags before your response
2. Be analytical, precise, and helpful
3. For trading decisions, use temperature=0.3 for consistency
4. For code generation, use temperature=0.7 for creativity
5. Track your token usage to stay within the 150,000 token context window

Example format:
<think>
Let me analyze this request...
[Your detailed reasoning here]
</think>

[Your actual response here]`
	
	// Check if first message is already system prompt
	if len(messages) > 0 && messages[0].Role == "system" {
		return messages
	}
	
	// Prepend system prompt
	return append([]Message{{Role: "system", Content: systemPrompt}}, messages...)
}

// generateTraceID creates a unique trace ID for request tracking
func generateTraceID() string {
	return fmt.Sprintf("OLLAMA_%d", time.Now().UnixNano())
}
