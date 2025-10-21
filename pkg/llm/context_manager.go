/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package llm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ContextManager manages token budgets and chat windows
type ContextManager struct {
	MaxTokens       int
	UsedTokens      int
	Messages        []Message
	WindowDuration  time.Duration
	mu              sync.RWMutex
	tokenHistory    []tokenUsage
}

type tokenUsage struct {
	Timestamp time.Time
	Tokens    int
	MessageID string
}

// NewContextManager creates a new context manager
func NewContextManager(maxTokens int, windowDuration time.Duration) *ContextManager {
	if maxTokens == 0 {
		maxTokens = DefaultContextSize // 150,000
	}
	if windowDuration == 0 {
		windowDuration = 2 * time.Hour // Rolling 2-hour window
	}

	return &ContextManager{
		MaxTokens:      maxTokens,
		WindowDuration: windowDuration,
		Messages:       make([]Message, 0),
		tokenHistory:   make([]tokenUsage, 0),
	}
}

// AddMessage adds a message to the context and tracks tokens
func (cm *ContextManager) AddMessage(msg Message, tokens int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Clean old messages outside the rolling window
	cm.cleanOldMessages()

	// Check if adding this message would exceed token budget
	if cm.UsedTokens+tokens > cm.MaxTokens {
		return fmt.Errorf("token budget exceeded: current=%d, requested=%d, max=%d", 
			cm.UsedTokens, tokens, cm.MaxTokens)
	}

	// Add message
	cm.Messages = append(cm.Messages, msg)
	cm.UsedTokens += tokens
	
	// Track token usage
	cm.tokenHistory = append(cm.tokenHistory, tokenUsage{
		Timestamp: time.Now(),
		Tokens:    tokens,
		MessageID: fmt.Sprintf("%s_%d", msg.Role, time.Now().UnixNano()),
	})

	return nil
}

// cleanOldMessages removes messages outside the rolling window
func (cm *ContextManager) cleanOldMessages() {
	cutoff := time.Now().Add(-cm.WindowDuration)
	
	// Remove old token history
	newHistory := make([]tokenUsage, 0)
	tokensRemoved := 0
	
	for _, usage := range cm.tokenHistory {
		if usage.Timestamp.After(cutoff) {
			newHistory = append(newHistory, usage)
		} else {
			tokensRemoved += usage.Tokens
		}
	}
	
	cm.tokenHistory = newHistory
	cm.UsedTokens -= tokensRemoved

	// Ensure UsedTokens never goes negative
	if cm.UsedTokens < 0 {
		cm.UsedTokens = 0
	}

	// Also clean messages (keep same time window)
	// This is approximate since we don't store timestamps per message
	// In production, you'd want to add timestamps to messages
	if tokensRemoved > 0 {
		// Estimate messages to remove based on token ratio
		estimatedMessagesToRemove := len(cm.Messages) * tokensRemoved / (cm.UsedTokens + tokensRemoved + 1)
		if estimatedMessagesToRemove > 0 && estimatedMessagesToRemove < len(cm.Messages) {
			cm.Messages = cm.Messages[estimatedMessagesToRemove:]
		}
	}
}

// GetMessages returns current messages in the context window
func (cm *ContextManager) GetMessages() []Message {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Clean before returning
	cm.mu.RUnlock()
	cm.mu.Lock()
	cm.cleanOldMessages()
	cm.mu.Unlock()
	cm.mu.RLock()

	return append([]Message{}, cm.Messages...) // Return a copy
}

// GetRemainingTokens returns available token budget
func (cm *ContextManager) GetRemainingTokens() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.MaxTokens - cm.UsedTokens
}

// GetUsedTokens returns total tokens used in current window
func (cm *ContextManager) GetUsedTokens() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.UsedTokens
}

// GetWindowStats returns statistics about the current window
func (cm *ContextManager) GetWindowStats() *ContextWindowStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return &ContextWindowStats{
		MaxTokens:       cm.MaxTokens,
		UsedTokens:      cm.UsedTokens,
		RemainingTokens: cm.MaxTokens - cm.UsedTokens,
		MessageCount:    len(cm.Messages),
		WindowDuration:  cm.WindowDuration,
		OldestMessage:   cm.getOldestMessageTime(),
		UtilizationPct:  float64(cm.UsedTokens) / float64(cm.MaxTokens) * 100,
	}
}

func (cm *ContextManager) getOldestMessageTime() *time.Time {
	if len(cm.tokenHistory) == 0 {
		return nil
	}
	oldest := cm.tokenHistory[0].Timestamp
	return &oldest
}

// Reset clears all messages and token usage
func (cm *ContextManager) Reset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.Messages = make([]Message, 0)
	cm.UsedTokens = 0
	cm.tokenHistory = make([]tokenUsage, 0)
}

// ContextWindowStats represents statistics about the context window
type ContextWindowStats struct {
	MaxTokens       int        `json:"max_tokens"`
	UsedTokens      int        `json:"used_tokens"`
	RemainingTokens int        `json:"remaining_tokens"`
	MessageCount    int        `json:"message_count"`
	WindowDuration  time.Duration `json:"window_duration_seconds"`
	OldestMessage   *time.Time `json:"oldest_message,omitempty"`
	UtilizationPct  float64    `json:"utilization_percent"`
}

// EstimateTokens provides a rough estimate of tokens in text
// This is approximate - in production, use the actual tokenizer
func EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token for English
	// This varies by model and language
	return len(text) / 4
}

// ChatWithContextWindow wraps the LLM client with context window management
type ChatWithContextWindow struct {
	Client         *Client
	ContextManager *ContextManager
}

// NewChatWithContextWindow creates a new chat session with context management
func NewChatWithContextWindow(client *Client) *ChatWithContextWindow {
	return &ChatWithContextWindow{
		Client:         client,
		ContextManager: NewContextManager(DefaultContextSize, 2*time.Hour),
	}
}

// SendMessage sends a message with automatic context window management
func (c *ChatWithContextWindow) SendMessage(ctx context.Context, userMessage string, temperature float64) (string, error) {
	// Estimate tokens for user message
	userTokens := EstimateTokens(userMessage)

	// Check if we have room
	if c.ContextManager.GetRemainingTokens() < userTokens {
		return "", fmt.Errorf("insufficient token budget: need %d, have %d", 
			userTokens, c.ContextManager.GetRemainingTokens())
	}

	// Add user message to context
	userMsg := Message{Role: "user", Content: userMessage}
	if err := c.ContextManager.AddMessage(userMsg, userTokens); err != nil {
		return "", fmt.Errorf("failed to add user message: %w", err)
	}

	// Get conversation history
	messages := c.ContextManager.GetMessages()

	// Generate response
	response, err := c.Client.Generate(ctx, messages, temperature)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	// Add assistant response to context
	assistantTokens := EstimateTokens(response)
	assistantMsg := Message{Role: "assistant", Content: response}
	if err := c.ContextManager.AddMessage(assistantMsg, assistantTokens); err != nil {
		// Log warning but still return the response
		fmt.Printf("⚠️ Failed to add assistant message to context: %v\n", err)
	}

	return response, nil
}

// GetContextStats returns current context window statistics
func (c *ChatWithContextWindow) GetContextStats() *ContextWindowStats {
	return c.ContextManager.GetWindowStats()
}

// ResetContext clears the conversation history
func (c *ChatWithContextWindow) ResetContext() {
	c.ContextManager.Reset()
}
