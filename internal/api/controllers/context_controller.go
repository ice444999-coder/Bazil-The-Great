package controllers

import (
	"ares_api/internal/interfaces/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ContextController exposes context window statistics
type ContextController struct {
	ChatService service.ChatService
}

// NewContextController creates a new context controller
func NewContextController(chatService service.ChatService) *ContextController {
	return &ContextController{ChatService: chatService}
}

// GetContextStats godoc
// @Summary Get context window statistics
// @Description Returns current token usage and message count for the user's rolling 2-hour window
// @Tags context
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /context/stats [get]
func (ctrl *ContextController) GetContextStats(c *gin.Context) {
	// Get userID from JWT middleware
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// This requires exposing GetContextStats from ChatService
	// For now, return a placeholder - we'll implement this properly
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "Context stats will be available after chat service refactoring",
		"max_tokens": 150000,
		"window_duration_hours": 2,
	})
}
