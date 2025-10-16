package controllers

import (
	"ares_api/internal/agent"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		Context   map[string]interface{} `json:"context"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default session ID
	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	// Call THE REAL SOLACE - the autonomous agent with memory, personality, and state
	response, err := sac.solaceAgent.RespondToUser(c.Request.Context(), req.Message, req.SessionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("SOLACE encountered an error: %v", err),
		})
		return
	}

	// Return response IMMEDIATELY (THE FIX - respond before database writes)
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

	// Log to database AFTER response sent (non-blocking async pattern)
	go func(sessionID, userMsg, solaceResp string, ctx map[string]interface{}) {
		// Log user message
		contextJSON, _ := json.Marshal(ctx)
		if err := sac.db.Exec(`
			INSERT INTO chat_history (session_id, sender, message, context, created_at)
			VALUES (?, 'user', ?, ?, NOW())
		`, sessionID, userMsg, contextJSON).Error; err != nil {
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
		`, sessionID, solaceResp, responseContextJSON).Error; err != nil {
			fmt.Printf("Failed to log SOLACE response: %v\n", err)
		}
	}(req.SessionID, req.Message, response, req.Context)
}
