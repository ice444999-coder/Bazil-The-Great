package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"ares_api/internal/models"
	"ares_api/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TestLogController handles test activity logging with Merkle tree batching
type TestLogController struct {
	db                 *gorm.DB
	merkleBatchService *services.MerkleBatchService
}

// NewTestLogController creates a new test log controller
func NewTestLogController(db *gorm.DB, merkleBatchService *services.MerkleBatchService) *TestLogController {
	return &TestLogController{
		db:                 db,
		merkleBatchService: merkleBatchService,
	}
}

// LogTestAction logs a test action with maximum security (Merkle tree batching)
// POST /api/v1/test-log
func (ctrl *TestLogController) LogTestAction(c *gin.Context) {
	var req struct {
		Actor         string `json:"actor" binding:"required"`
		ActionType    string `json:"action_type" binding:"required"`
		FeatureTested string `json:"feature_tested"`
		ActionDetails string `json:"action_details"`
		Result        string `json:"result"`
		ResponseData  string `json:"response_data"`
		ErrorMessage  string `json:"error_message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate internal hash (double-hashed for security)
	messageContent := fmt.Sprintf("%s|%s|%s|%s|%d",
		req.Actor,
		req.ActionType,
		req.FeatureTested,
		req.ActionDetails,
		time.Now().Unix(),
	)

	firstHash := sha256.Sum256([]byte(messageContent))
	secondHash := sha256.Sum256(firstHash[:])
	internalHash := hex.EncodeToString(secondHash[:])

	// Create Glass Box log entry
	log := models.GlassBoxLog{
		Actor:          req.Actor,
		ActionType:     req.ActionType,
		FeatureTested:  req.FeatureTested,
		MessageContent: messageContent,
		ActionDetails:  req.ActionDetails,
		InternalHash:   internalHash,
		Result:         req.Result,
		ErrorMessage:   req.ErrorMessage,
		Timestamp:      time.Now(),
	}

	// Save to database
	if err := ctrl.db.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log action"})
		return
	}

	// Add to Merkle batch (async processing)
	if err := ctrl.merkleBatchService.AddToBatch(log); err != nil {
		// Log error but don't fail the request
		fmt.Printf("⚠️ Failed to add to Merkle batch: %v\n", err)
	}

	// Get pending batch info
	batchInfo := ctrl.merkleBatchService.GetPendingBatchInfo()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"log_id": log.LogID,
		"message": fmt.Sprintf("Action logged securely. Pending batch: %d/%d entries",
			batchInfo["pending_logs"], batchInfo["batch_size"]),
		"timestamp": log.Timestamp.Format("2006-01-02 15:04:05"),
	})
}

// GetTestLogs retrieves test activity logs
// GET /api/v1/test-log?actor=SOLACE&limit=20
func (ctrl *TestLogController) GetTestLogs(c *gin.Context) {
	actor := c.Query("actor")
	limit := 20

	query := ctrl.db.Order("timestamp DESC").Limit(limit)
	if actor != "" {
		query = query.Where("actor = ?", actor)
	}

	var logs []models.GlassBoxLog
	err := query.Find(&logs).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	// Convert to safe response (no internal hashes exposed)
	var safeLogs []map[string]interface{}
	for _, log := range logs {
		safeLog := map[string]interface{}{
			"log_id":         log.LogID,
			"actor":          log.Actor,
			"action_type":    log.ActionType,
			"feature_tested": log.FeatureTested,
			"result":         log.Result,
			"timestamp":      log.Timestamp,
			"in_batch":       log.MerkleBatchID != nil,
		}
		if log.MerkleBatchID != nil {
			safeLog["batch_id"] = *log.MerkleBatchID
		}
		safeLogs = append(safeLogs, safeLog)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(safeLogs),
		"logs":   safeLogs,
	})
}

// VerifyLog verifies a log entry with full Merkle proof chain
// GET /api/v1/test-log/verify/:log_id
func (ctrl *TestLogController) VerifyLog(c *gin.Context) {
	var req struct {
		LogID uint `uri:"log_id" binding:"required"`
	}

	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	verification, err := ctrl.merkleBatchService.VerifyLog(req.LogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"verification": verification,
	})
}

// GetBatchInfo returns information about Merkle batches
// GET /api/v1/test-log/batches
func (ctrl *TestLogController) GetBatchInfo(c *gin.Context) {
	var batches []models.MerkleBatch
	err := ctrl.db.Order("created_at DESC").Limit(10).Find(&batches).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve batches"})
		return
	}

	pendingInfo := ctrl.merkleBatchService.GetPendingBatchInfo()

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"recent_batches": batches,
		"pending_batch":  pendingInfo,
	})
}
