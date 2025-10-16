package memory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ConversationSummaryBufferMemory - LangChain pattern adapted for SOLACE
// Keeps recent messages + summarizes older ones when token limit exceeded
type ConversationSummaryBufferMemory struct {
	DB            *gorm.DB
	SessionID     string
	MaxTokenLimit int    // When to trigger summarization
	MovingSummary string // Progressive summary
	KeepLastN     int    // Keep last N messages verbatim
	llmSummarizer func(context.Context, []Message) (string, error)
}

type Message struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func NewConversationMemory(db *gorm.DB, sessionID string, summarizer func(context.Context, []Message) (string, error)) *ConversationSummaryBufferMemory {
	return &ConversationSummaryBufferMemory{
		DB:            db,
		SessionID:     sessionID,
		MaxTokenLimit: 2000, // ~2000 tokens before summarization
		KeepLastN:     5,    // Keep last 5 messages verbatim
		llmSummarizer: summarizer,
	}
}

// LoadMemoryVariables - Get formatted conversation history for LLM
func (m *ConversationSummaryBufferMemory) LoadMemoryVariables(ctx context.Context) (string, error) {
	// Get all messages for this session
	var dbMessages []struct {
		Sender    string
		Message   string
		CreatedAt time.Time
	}

	err := m.DB.Raw(`
		SELECT sender, message, created_at 
		FROM chat_history 
		WHERE session_id = ? 
		ORDER BY created_at ASC
	`, m.SessionID).Scan(&dbMessages).Error

	if err != nil {
		return "", err
	}

	// Convert to Message format
	messages := make([]Message, len(dbMessages))
	for i, msg := range dbMessages {
		messages[i] = Message{
			Role:      msg.Sender,
			Content:   msg.Message,
			Timestamp: msg.CreatedAt,
		}
	}

	// If under token limit, return all messages
	tokenCount := m.estimateTokens(messages)
	if tokenCount <= m.MaxTokenLimit {
		return m.formatMessages(messages), nil
	}

	// Need to summarize - keep last N, summarize the rest
	if len(messages) <= m.KeepLastN {
		return m.formatMessages(messages), nil
	}

	// Split: older messages (to summarize) + recent messages (keep verbatim)
	splitPoint := len(messages) - m.KeepLastN
	olderMessages := messages[:splitPoint]
	recentMessages := messages[splitPoint:]

	// Generate summary of older messages
	summary, err := m.llmSummarizer(ctx, olderMessages)
	if err != nil {
		// Fallback: just use recent messages
		return m.formatMessages(recentMessages), nil
	}

	// Build final context: summary + recent messages
	var builder strings.Builder
	builder.WriteString("CONVERSATION SUMMARY (earlier messages):\n")
	builder.WriteString(summary)
	builder.WriteString("\n\nRECENT MESSAGES:\n")
	builder.WriteString(m.formatMessages(recentMessages))

	m.MovingSummary = summary // Update moving summary
	return builder.String(), nil
}

// SaveContext - Store new message
func (m *ConversationSummaryBufferMemory) SaveContext(ctx context.Context, role, content string) error {
	return m.DB.Exec(`
		INSERT INTO chat_history (session_id, sender, message, created_at)
		VALUES (?, ?, ?, NOW())
	`, m.SessionID, role, content).Error
}

// formatMessages - Format messages for LLM context
func (m *ConversationSummaryBufferMemory) formatMessages(messages []Message) string {
	var builder strings.Builder
	for _, msg := range messages {
		timeStr := msg.Timestamp.Format("15:04:05")
		builder.WriteString(fmt.Sprintf("[%s] %s: %s\n", timeStr, strings.ToUpper(msg.Role), msg.Content))
	}
	return builder.String()
}

// estimateTokens - Rough token estimation (4 chars = 1 token)
func (m *ConversationSummaryBufferMemory) estimateTokens(messages []Message) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
	}
	return totalChars / 4
}

// Clear - Reset conversation memory
func (m *ConversationSummaryBufferMemory) Clear() error {
	m.MovingSummary = ""
	return m.DB.Exec(`DELETE FROM chat_history WHERE session_id = ?`, m.SessionID).Error
}
