package dto

type MemoryLearnRequest struct {
	EventType string                 `json:"event_type" binding:"required"`
	Payload   map[string]interface{} `json:"payload" binding:"required"`
	SessionID *string                `json:"session_id,omitempty"`
}

type MemoryRecallRequest struct {
	Limit     int    `json:"limit,omitempty"`
	EventType string `json:"event_type,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

type MemoryRecallResponse struct {
	ID        uint                   `json:"id"`
	Timestamp string                 `json:"timestamp"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	UserID    uint                   `json:"user_id"`
	SessionID *string                `json:"session_id,omitempty"`
}

type MemoryLearnResponse struct {
	Message string `json:"message"`
	ID      uint   `json:"id"`
}

type ConversationImportRequest struct {
	Content string   `json:"content" binding:"required"`
	Source  string   `json:"source"`
	Tags    []string `json:"tags"`
}

type ConversationImportResponse struct {
	Message      string `json:"message"`
	MessageCount int    `json:"message_count"`
	ImportID     uint   `json:"import_id"`
}
