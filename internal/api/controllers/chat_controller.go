/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	service "ares_api/internal/interfaces/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	Service service.ChatService
	LedgerService service.LedgerService
}

func NewChatController(s service.ChatService , l service.LedgerService) *ChatController {
	return &ChatController{Service: s , LedgerService: l}
}

// @Summary Send a chat message
// @Description Sends a message to Ollama AI and stores the response
// @Tags Chat
// @Accept  json
// @Produce  json
// @Param   chat body dto.ChatRequest true "Chat Message"
// @Success 200 {object} dto.ChatResponse
// @Security BearerAuth
// @Router /chat/send [post]
func (cc *ChatController) SendMessage(c *gin.Context) {
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get userID from JWT middleware context, or use default guest user (1)
	userID := uint(1) // Default guest user
	userIDInterface, exists := c.Get("userID")
	if exists {
		userID = userIDInterface.(uint)
	}

	// Call chat service
	resp, err := cc.Service.SendMessage(userID, req)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---- Ledger logging ----
	if cc.LedgerService != nil {
		details := fmt.Sprintf(`{"message":"%s","response":"%s"}`, req.Message, resp.Response)
		_ = cc.LedgerService.Append(userID, "chat", details)
	}

	common.JSON(c, http.StatusOK, resp)
}


// @Summary Get chat history
// @Description Retrieves last N chat messages for a user
// @Tags Chat
// @Accept  json
// @Produce  json
// @Param   limit query int false "Number of messages" default(20)
// @Success 200 {object} dto.ChatHistoryListResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /chat/history [get]
func (cc *ChatController) GetHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	// Get userID from JWT middleware context, or use default guest user (1)
	userID := uint(1) // Default guest user
	userIDInterface, exists := c.Get("userID")
	if exists {
		userID = userIDInterface.(uint)
	}

	history, err := cc.Service.GetHistory(userID, limit)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to UI-friendly format
	messages := make([]dto.ChatHistoryMessage, 0)
	for _, h := range history {
		// Add user message
		messages = append(messages, dto.ChatHistoryMessage{
			Role:      "user",
			Content:   h.Message,
			CreatedAt: h.CreatedAt,
		})
		// Add assistant response
		messages = append(messages, dto.ChatHistoryMessage{
			Role:      "assistant",
			Content:   h.Response,
			CreatedAt: h.CreatedAt,
			Thinking:  "", // TODO: Extract <think> tags if present
		})
	}

	common.JSON(c, http.StatusOK, dto.ChatHistoryListResponse{
		Messages: messages,
	})
}
