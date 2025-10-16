package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	service "ares_api/internal/interfaces/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MemoryController struct {
	Service       service.MemoryService
	LedgerService service.LedgerService
}

func NewMemoryController(s service.MemoryService, l service.LedgerService) *MemoryController {
	return &MemoryController{Service: s, LedgerService: l}
}

// @Summary Store a memory snapshot
// @Description Stores a memory event for the user
// @Tags Memory
// @Accept  json
// @Produce  json
// @Param   memory body dto.MemoryLearnRequest true "Memory Event"
// @Success 201 {object} dto.MemoryLearnResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /memory/learn [post]
func (mc *MemoryController) Learn(c *gin.Context) {
	var req dto.MemoryLearnRequest
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

	// Parse SessionID if provided
	var sessionID *uuid.UUID
	if req.SessionID != nil && *req.SessionID != "" {
		parsedUUID, err := uuid.Parse(*req.SessionID)
		if err != nil {
			common.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid session_id format"})
			return
		}
		sessionID = &parsedUUID
	}

	// Call service to save memory snapshot
	err := mc.Service.Learn(userID, req.EventType, req.Payload, sessionID)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---- Ledger logging ----
	if mc.LedgerService != nil {
		details := fmt.Sprintf(`{"event_type":"%s","has_session":%t}`, req.EventType, sessionID != nil)
		_ = mc.LedgerService.Append(userID, "memory_learn", details)
	}

	common.JSON(c, http.StatusCreated, dto.MemoryLearnResponse{
		Message: "Memory snapshot saved successfully",
	})
}

// @Summary Recall memory snapshots
// @Description Retrieves memory snapshots for the user
// @Tags Memory
// @Accept  json
// @Produce  json
// @Param   limit query int false "Number of snapshots to retrieve" default(20)
// @Param   event_type query string false "Filter by event type"
// @Param   session_id query string false "Filter by session ID"
// @Success 200 {array} dto.MemoryRecallResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /memory/recall [get]
func (mc *MemoryController) Recall(c *gin.Context) {
	// Get userID from JWT middleware context, or use default guest user (per SOLACE's decision)
	userID := uint(1) // Default guest user
	userIDInterface, exists := c.Get("userID")
	if exists {
		userID = userIDInterface.(uint)
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	eventType := c.Query("event_type")
	sessionIDStr := c.Query("session_id")

	var memories []dto.MemoryRecallResponse
	var err error

	// Route to appropriate service method based on filters
	if sessionIDStr != "" {
		// Filter by session ID
		sessionID, parseErr := uuid.Parse(sessionIDStr)
		if parseErr != nil {
			common.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid session_id format"})
			return
		}
		memories, err = mc.Service.RecallBySessionID(sessionID, limit)
	} else if eventType != "" {
		// Filter by event type
		memories, err = mc.Service.RecallByEventType(userID, eventType, limit)
	} else {
		// No filters - get recent memories
		memories, err = mc.Service.Recall(userID, limit)
	}

	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---- Ledger logging ----
	if mc.LedgerService != nil {
		details := fmt.Sprintf(`{"limit":%d,"event_type":"%s","has_session":%t}`, limit, eventType, sessionIDStr != "")
		_ = mc.LedgerService.Append(userID, "memory_recall", details)
	}

	common.JSON(c, http.StatusOK, memories)
}

// @Summary Import a conversation
// @Description Imports a pasted conversation or text into memory system
// @Tags Memory
// @Accept  json
// @Produce  json
// @Param   request body dto.ConversationImportRequest true "Conversation Import"
// @Success 200 {object} dto.ConversationImportResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /memory/import [post]
func (mc *MemoryController) ImportConversation(c *gin.Context) {
	// Get userID from JWT middleware context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	var req dto.ConversationImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Import conversation
	messageCount, importID, err := mc.Service.ImportConversation(userID, req.Content, req.Source, req.Tags)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---- Ledger logging ----
	if mc.LedgerService != nil {
		details := fmt.Sprintf(`{"source":"%s","message_count":%d}`, req.Source, messageCount)
		_ = mc.LedgerService.Append(userID, "conversation_import", details)
	}

	common.JSON(c, http.StatusOK, dto.ConversationImportResponse{
		Message:      "Conversation imported successfully",
		MessageCount: messageCount,
		ImportID:     importID,
	})
}

// @Summary Get all memory snapshots
// @Description Retrieves all memory snapshots for the user with optional limit
// @Tags Memory
// @Accept  json
// @Produce  json
// @Param   limit query int false "Number of snapshots to retrieve" default(100)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /memory/snapshots [get]
func (mc *MemoryController) GetSnapshots(c *gin.Context) {
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

	// Get all memories
	memories, err := mc.Service.Recall(userID, limit)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---- Ledger logging ----
	if mc.LedgerService != nil {
		details := fmt.Sprintf(`{"limit":%d,"count":%d}`, limit, len(memories))
		_ = mc.LedgerService.Append(userID, "memory_snapshots", details)
	}

	common.JSON(c, http.StatusOK, gin.H{
		"snapshots": memories,
		"count":     len(memories),
	})
}
