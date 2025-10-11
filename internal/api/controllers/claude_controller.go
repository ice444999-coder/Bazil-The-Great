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

type ClaudeController struct {
	Service       service.ClaudeService
	LedgerService service.LedgerService
}

func NewClaudeController(s service.ClaudeService, l service.LedgerService) *ClaudeController {
	return &ClaudeController{Service: s, LedgerService: l}
}

// @Summary Chat with stateful Claude AI
// @Description Chat with Claude AI with full memory context, file system access, and recursive learning
// @Tags Claude
// @Accept  json
// @Produce  json
// @Param   chat body dto.ClaudeChatRequest true "Chat Message"
// @Success 200 {object} dto.ClaudeChatResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /claude/chat [post]
func (cc *ClaudeController) Chat(c *gin.Context) {
	var req dto.ClaudeChatRequest
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

	// Call Claude service
	resp, err := cc.Service.Chat(userID, req.Message, sessionID, req.IncludeFiles, req.MaxTokens)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---- Ledger logging ----
	if cc.LedgerService != nil {
		details := fmt.Sprintf(`{"message_length":%d,"session_id":"%s","memories_loaded":%d,"tokens_used":%d}`,
			len(req.Message), resp.SessionID, resp.MemoriesLoaded, resp.TokensUsed)
		_ = cc.LedgerService.Append(userID, "claude_chat", details)
	}

	common.JSON(c, http.StatusOK, resp)
}

// @Summary Retrieve Claude's memories
// @Description Get Claude's past interactions and memories with filtering options
// @Tags Claude
// @Accept  json
// @Produce  json
// @Param   session_id query string false "Filter by session ID"
// @Param   limit query int false "Number of memories to retrieve" default(20)
// @Param   event_type query string false "Filter by event type"
// @Success 200 {object} dto.ClaudeMemoryResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /claude/memory [get]
func (cc *ClaudeController) GetMemory(c *gin.Context) {
	// Get userID from JWT middleware context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	eventType := c.Query("event_type")
	sessionIDStr := c.Query("session_id")

	var sessionID *uuid.UUID
	if sessionIDStr != "" {
		parsedUUID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			common.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid session_id format"})
			return
		}
		sessionID = &parsedUUID
	}

	// Get memories
	memories, err := cc.Service.GetMemories(userID, sessionID, limit, eventType)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	common.JSON(c, http.StatusOK, memories)
}

// @Summary Read a file from the repository
// @Description Read any file from the ARES repository (Claude's file system access)
// @Tags Claude
// @Accept  json
// @Produce  json
// @Param   file body dto.ClaudeFileRequest true "File Path"
// @Success 200 {object} dto.ClaudeFileResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /claude/file [post]
func (cc *ClaudeController) ReadFile(c *gin.Context) {
	var req dto.ClaudeFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get userID for logging (still require auth)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Read file
	fileResp, err := cc.Service.ReadFile(req.FilePath)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---- Ledger logging ----
	if cc.LedgerService != nil {
		details := fmt.Sprintf(`{"file_path":"%s","file_size":%d}`, req.FilePath, fileResp.Size)
		_ = cc.LedgerService.Append(userID, "claude_file_read", details)
	}

	common.JSON(c, http.StatusOK, fileResp)
}

// @Summary Get repository context
// @Description Get an overview of the ARES repository structure
// @Tags Claude
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.ClaudeRepositoryContextResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /claude/repository [get]
func (cc *ClaudeController) GetRepositoryContext(c *gin.Context) {
	// Get userID for auth
	_, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get repository context
	repoContext, err := cc.Service.GetRepositoryContext()
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	common.JSON(c, http.StatusOK, repoContext)
}

// @Summary Semantic memory search
// @Description Search memories using semantic similarity (intelligent retrieval)
// @Tags Claude
// @Accept  json
// @Produce  json
// @Param   search body dto.SemanticSearchRequest true "Search Query"
// @Success 200 {object} dto.SemanticSearchResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /claude/semantic-search [post]
func (cc *ClaudeController) SemanticSearch(c *gin.Context) {
	var req dto.SemanticSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get userID for auth
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Perform semantic search
	resp, err := cc.Service.SemanticMemorySearch(req.Query, req.Limit, req.Threshold)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---- Ledger logging ----
	if cc.LedgerService != nil {
		details := fmt.Sprintf(`{"query":"%s","results_found":%d,"execution_time_ms":%d}`,
			req.Query, resp.ResultsFound, resp.ExecutionTime)
		_ = cc.LedgerService.Append(userID, "claude_semantic_search", details)
	}

	common.JSON(c, http.StatusOK, resp)
}

// @Summary Process embedding queue
// @Description Generate embeddings for pending memory snapshots
// @Tags Claude
// @Accept  json
// @Produce  json
// @Param   process body dto.ProcessEmbeddingsRequest true "Batch Size"
// @Success 200 {object} dto.ProcessEmbeddingsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /claude/process-embeddings [post]
func (cc *ClaudeController) ProcessEmbeddings(c *gin.Context) {
	var req dto.ProcessEmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.BatchSize = 50 // default
	}

	// Get userID for auth
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Process embeddings
	resp, err := cc.Service.ProcessEmbeddingQueue(req.BatchSize)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---- Ledger logging ----
	if cc.LedgerService != nil {
		details := fmt.Sprintf(`{"processed":%d,"pending":%d}`, resp.Processed, resp.Pending)
		_ = cc.LedgerService.Append(userID, "claude_process_embeddings", details)
	}

	common.JSON(c, http.StatusOK, resp)
}
