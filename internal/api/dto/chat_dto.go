/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package dto

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}


type ChatResponse struct {
	Message  string `json:"message"`
	Response string `json:"response"`
}


type ChatHistoryResponse struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Message   string `json:"message"`
	Response  string `json:"response"`
	CreatedAt string `json:"created_at"`
}

// ChatHistoryMessage represents a single message in chat history (UI-friendly format)
type ChatHistoryMessage struct {
	Role      string `json:"role"`      // "user" or "assistant"
	Content   string `json:"content"`   // The actual message text
	CreatedAt string `json:"created_at"` // Timestamp
	Thinking  string `json:"thinking,omitempty"` // Optional thinking process
}

// ChatHistoryListResponse wraps the messages array for UI
type ChatHistoryListResponse struct {
	Messages []ChatHistoryMessage `json:"messages"`
}
