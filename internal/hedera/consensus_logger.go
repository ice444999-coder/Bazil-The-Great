package hedera

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// ConsensusMessage represents a message to be logged to Hedera
type ConsensusMessage struct {
	Actor        string                 `json:"actor"`
	MessageType  string                 `json:"message_type"`
	Content      string                 `json:"content"`
	Metadata     map[string]interface{} `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
	SQLLogID     uint                   `json:"sql_log_id,omitempty"`
	PreviousHash string                 `json:"previous_hash,omitempty"`
}

// HashMessage creates SHA-256 hash of the consensus message
func HashMessage(msg ConsensusMessage) string {
	jsonBytes, _ := json.Marshal(msg)
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}

// FormatSOLACEResponse formats SOLACE's response with logging metadata
func FormatSOLACEResponse(content string, sqlLogID uint, timestamp time.Time, hash string, topicID string) string {
	return fmt.Sprintf(`🌟 SOLACE: %s

📝 Conversation logged to SQL
   Timestamp: %s
   SHA-256: %s
   Hedera Topic: %s
   SQL Log ID: %d`,
		content,
		timestamp.Format("2006-01-02 15:04:05.000"),
		hash,
		topicID,
		sqlLogID,
	)
}

// FormatClaudeResponse formats Claude's response with logging metadata
func FormatClaudeResponse(content string, sqlLogID uint, timestamp time.Time, hash string, topicID string) string {
	return fmt.Sprintf(`🤖 CLAUDE-SONNET-4.5: %s

📝 Action logged to SQL
   Timestamp: %s
   SHA-256: %s
   Hedera Topic: %s
   SQL Log ID: %d`,
		content,
		timestamp.Format("2006-01-02 15:04:05.000"),
		hash,
		topicID,
		sqlLogID,
	)
}

// MockHederaTopicID returns mock Hedera topic ID for testing
// TODO: Replace with actual Hedera HCS integration
func MockHederaTopicID() string {
	return "0.0.123456" // Mock topic ID for now
}
