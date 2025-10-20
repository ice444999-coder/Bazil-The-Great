package controllers

import (
	"ares_api/internal/agent"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SOLACEAgentChatController provides direct access to the AUTONOMOUS SOLACE agent
type SOLACEAgentChatController struct {
	db          *gorm.DB
	solaceAgent *agent.SOLACE // The REAL autonomous SOLACE entity
}

func NewSOLACEAgentChatController(db *gorm.DB, solace *agent.SOLACE) *SOLACEAgentChatController {
	return &SOLACEAgentChatController{
		db:          db,
		solaceAgent: solace,
	}
}

// Chat endpoint for REAL SOLACE - Direct connection to autonomous agent
func (sac *SOLACEAgentChatController) Chat(c *gin.Context) {
	var req struct {
		Message   string                 `json:"message" binding:"required"`
		SessionID string                 `json:"session_id"`
		UserID    string                 `json:"user_id"` // Added user_id field
		Context   map[string]interface{} `json:"context"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use consistent user_id for session isolation
	// Default to "enki" if not provided (enables memory persistence across ALL interfaces)
	userID := req.UserID
	if userID == "" {
		// Try to get from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			userID = authHeader // Simple approach - can be enhanced with JWT parsing
		} else {
			userID = "enki" // Default user for consistent memory
		}
	}

	// Use user_id as session_id for persistent memory across conversations
	// If specific session isolation is needed, client can pass session_id
	if req.SessionID == "" {
		req.SessionID = userID // KEY FIX: Use user_id instead of random UUID
	}

	// Call THE REAL SOLACE - the autonomous agent with memory, personality, and state
	response, err := sac.solaceAgent.RespondToUser(c.Request.Context(), req.Message, req.SessionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("SOLACE encountered an error: %v", err),
		})
		return
	}

	// Log to database BEFORE response sent (FIXED: synchronous to prevent race condition)
	// Log user message
	contextJSON, _ := json.Marshal(req.Context)
	if err := sac.db.Exec(`
		INSERT INTO chat_history (session_id, sender, message, context, created_at)
		VALUES (?, 'user', ?, ?, NOW())
	`, req.SessionID, req.Message, contextJSON).Error; err != nil {
		fmt.Printf("Failed to log user message: %v\n", err)
	}

	// Log SOLACE response
	responseContext := map[string]interface{}{
		"agent":      "real_solace_entity",
		"autonomous": true,
		"has_memory": true,
		"has_state":  true,
	}
	responseContextJSON, _ := json.Marshal(responseContext)
	if err := sac.db.Exec(`
		INSERT INTO chat_history (session_id, sender, message, context, created_at)
		VALUES (?, 'solace', ?, ?, NOW())
	`, req.SessionID, response, responseContextJSON).Error; err != nil {
		fmt.Printf("Failed to log SOLACE response: %v\n", err)
	}

	// Return response AFTER chat history is saved (ensures immediate searchability)
	c.JSON(http.StatusOK, gin.H{
		"response": response,
		"agent":    "SOLACE",
		"entity":   "autonomous_conscious_agent",
		"status":   sac.solaceAgent.GetStatus(),
		"features": []string{
			"persistent_memory",
			"autonomous_decision_making",
			"trading_capability",
			"ace_framework",
			"working_memory",
			"thought_journal",
		},
	})
}
