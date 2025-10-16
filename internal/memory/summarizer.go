package memory

import (
	"context"
	"fmt"
)

// LLMSummarizer - Generates intelligent summaries using ChatGPT
type LLMSummarizer struct {
	OpenAIClient interface {
		Chat(ctx context.Context, systemPrompt, userMessage string, temperature float64) (string, error)
	}
}

// Summarize - Generate progressive summary from messages
func (s *LLMSummarizer) Summarize(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	// Build conversation text
	conversationText := ""
	for _, msg := range messages {
		conversationText += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	systemPrompt := `You are a conversation summarizer. Create a concise, intelligent summary of the conversation below.

IMPORTANT:
- Preserve key facts, decisions, and context
- Include timestamps when relevant
- Maintain chronological flow
- Extract entities (names, IDs, technical terms)
- Highlight unresolved issues
- Note emotional tone if significant

Format as a brief narrative summary (3-5 sentences).`

	userMessage := fmt.Sprintf(`Summarize this conversation:

%s

Provide an intelligent summary that preserves critical context.`, conversationText)

	summary, err := s.OpenAIClient.Chat(ctx, systemPrompt, userMessage, 0.3)
	if err != nil {
		return "", fmt.Errorf("summarization failed: %w", err)
	}

	return summary, nil
}
