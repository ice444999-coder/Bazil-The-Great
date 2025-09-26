package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	service "ares_api/internal/interfaces/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	Service service.ChatService
}

func NewChatController(s service.ChatService) *ChatController {
	return &ChatController{Service: s}
}

// @Summary Send a chat message
// @Description Sends a message to Ollama AI and stores the response
// @Tags Chat
// @Accept  json
// @Produce  json
// @Param   chat body dto.ChatRequest true "Chat Message"
// @Success 200 {object} dto.ChatResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /chat/send [post]
func (cc *ChatController) SendMessage(c *gin.Context) {
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get userID from JWT middleware context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	resp, err := cc.Service.SendMessage(userID, req)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	common.JSON(c, http.StatusOK, resp)
}

// @Summary Get chat history
// @Description Retrieves last N chat messages for a user
// @Tags Chat
// @Accept  json
// @Produce  json
// @Param   limit query int false "Number of messages" default(20)
// @Success 200 {array} dto.ChatHistoryResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /chat/history [get]
func (cc *ChatController) GetHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	// Get userID from JWT middleware context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	history, err := cc.Service.GetHistory(userID, limit)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	common.JSON(c, http.StatusOK, history)
}
