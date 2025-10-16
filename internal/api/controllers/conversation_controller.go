package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"ares_api/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConversationController handles conversation memory retrieval
type ConversationController struct {
	DB *gorm.DB
}

// NewConversationController creates a new conversation controller
func NewConversationController(db *gorm.DB) *ConversationController {
	return &ConversationController{DB: db}
}

// MemoryCard represents a formatted conversation memory for the UI
type MemoryCard struct {
	ID          uint      `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	MemoryType  string    `json:"memory_type"`
	Content     string    `json:"content"`
	Importance  float64   `json:"importance"`
	UserMessage string    `json:"user_message,omitempty"`
	Response    string    `json:"response,omitempty"`
}

// GetConversations retrieves chat history formatted as memory cards
// @Summary Get conversation memories
// @Description Retrieve chat history from CHATS table formatted as episodic/semantic memories
// @Tags Memory
// @Produce json
// @Param limit query int false "Maximum number of conversations" default(100)
// @Success 200 {object} map[string]interface{} "memories and count"
// @Router /memory/conversations [get]
func (cc *ConversationController) GetConversations(c *gin.Context) {
	// Default to guest user (userID=1) per SOLACE's decision for memory endpoints
	userID := uint(1)
	
	// Get userID from JWT middleware context if available
	userIDInterface, exists := c.Get("userID")
	if exists {
		userID = userIDInterface.(uint)
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	// Query CHATS table for all non-deleted entries
	var chats []models.Chat
	result := cc.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&chats)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Transform each chat into a memory card (per SOLACE's design)
	memories := make([]MemoryCard, 0, len(chats))
	for _, chat := range chats {
		memoryCard := MemoryCard{
			ID:          chat.ID,
			Timestamp:   chat.CreatedAt,
			MemoryType:  classifyMemory(chat.Message, chat.Response),
			Content:     formatContent(chat.Message, chat.Response),
			Importance:  calculateImportance(chat.Message, chat.Response),
			UserMessage: chat.Message,
			Response:    chat.Response,
		}
		memories = append(memories, memoryCard)
	}

	// Return in same format as snapshots for UI compatibility
	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"count":    len(memories),
	})
}

// classifyMemory determines if a conversation is episodic or semantic
// SOLACE's classification logic: episodic = events/actions, semantic = knowledge/facts
func classifyMemory(message, response string) string {
	// Simple keyword-based classification (SOLACE can enhance this later)
	episodicKeywords := []string{"error", "fix", "debug", "issue", "problem", "crash", "alert", "urgent"}
	semanticKeywords := []string{"explain", "what is", "how does", "define", "teach", "learn", "why", "concept"}
	
	messageLower := strings.ToLower(message)
	responseLower := strings.ToLower(response)
	
	// Check episodic keywords
	for _, keyword := range episodicKeywords {
		if strings.Contains(messageLower, keyword) || strings.Contains(responseLower, keyword) {
			return "episodic"
		}
	}
	
	// Check semantic keywords
	for _, keyword := range semanticKeywords {
		if strings.Contains(messageLower, keyword) || strings.Contains(responseLower, keyword) {
			return "semantic"
		}
	}
	
	// Default: working memory (current context)
	return "working"
}

// formatContent combines user message and system response
func formatContent(message, response string) string {
	// Truncate for UI display
	maxLen := 200
	content := message
	if len(content) > maxLen {
		content = content[:maxLen] + "..."
	}
	return content
}

// calculateImportance assigns importance score (SOLACE's metric)
func calculateImportance(message, response string) float64 {
	// Initial implementation: length-based (SOLACE can enhance with ML later)
	score := 0.5 // default
	
	if len(message) > 100 || len(response) > 500 {
		score = 0.8 // longer conversations are more important
	}
	
	return score
}
