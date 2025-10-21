/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"ares_api/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GlassBoxController handles glass box transparency queries
type GlassBoxController struct {
	db *gorm.DB
}

// NewGlassBoxController creates a new glass box controller
func NewGlassBoxController(db *gorm.DB) *GlassBoxController {
	return &GlassBoxController{db: db}
}

// GetRecentLogs retrieves recent glass box logs with full transparency
// GET /api/v1/glass-box/logs?actor=SOLACE&limit=20
func (ctrl *GlassBoxController) GetRecentLogs(c *gin.Context) {
	actor := c.Query("actor")
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 100 {
		limit = 10
	}

	query := ctrl.db.Order("timestamp DESC").Limit(limit)
	if actor != "" {
		query = query.Where("actor = ?", actor)
	}

	var logs []models.GlassBoxLog
	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	// Return glass box formatted response with SHA-256 hashes visible
	var glassBoxLogs []map[string]interface{}
	for _, log := range logs {
		glassBoxLog := map[string]interface{}{
			"log_id":        log.LogID,
			"actor":         log.Actor,
			"action_type":   log.ActionType,
			"internal_hash": log.InternalHash, // SHA-256 hash exposed for transparency
			"timestamp":     log.Timestamp.Format("2006-01-02 15:04:05"),
			"in_batch":      log.MerkleBatchID != nil,
		}

		// Add message preview (first 100 chars)
		if len(log.MessageContent) > 100 {
			glassBoxLog["message_preview"] = log.MessageContent[:100] + "..."
		} else {
			glassBoxLog["message_preview"] = log.MessageContent
		}

		// Add Merkle batch info if available
		if log.MerkleBatchID != nil {
			glassBoxLog["merkle_batch_id"] = *log.MerkleBatchID
			if log.MerkleProof != "" {
				glassBoxLog["merkle_proof_available"] = true
			}
		}

		glassBoxLogs = append(glassBoxLogs, glassBoxLog)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(glassBoxLogs),
		"logs":   glassBoxLogs,
	})
}

// GetLogByID retrieves a specific glass box log with full details
// GET /api/v1/glass-box/logs/:log_id
func (ctrl *GlassBoxController) GetLogByID(c *gin.Context) {
	logIDStr := c.Param("log_id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	var log models.GlassBoxLog
	if err := ctrl.db.First(&log, logID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Return full glass box log with all transparency data
	glassBoxLog := map[string]interface{}{
		"log_id":         log.LogID,
		"actor":          log.Actor,
		"action_type":    log.ActionType,
		"feature_tested": log.FeatureTested,
		"message":        log.MessageContent,
		"action_details": log.ActionDetails,
		"internal_hash":  log.InternalHash, // SHA-256 hash for immutability proof
		"timestamp":      log.Timestamp.Format("2006-01-02 15:04:05"),
		"result":         log.Result,
		"in_batch":       log.MerkleBatchID != nil,
	}

	// Add Merkle tree proof data if available
	if log.MerkleBatchID != nil {
		glassBoxLog["merkle_batch_id"] = *log.MerkleBatchID
		glassBoxLog["merkle_leaf_index"] = log.MerkleLeafIndex

		if log.MerkleProof != "" {
			glassBoxLog["merkle_proof"] = log.MerkleProof
		}

		// Get batch info
		var batch models.MerkleBatch
		if err := ctrl.db.First(&batch, *log.MerkleBatchID).Error; err == nil {
			glassBoxLog["merkle_root_hash"] = batch.RootHash
			if batch.HederaTxID != "" {
				glassBoxLog["hedera_tx_id"] = batch.HederaTxID
				glassBoxLog["hedera_timestamp"] = batch.HederaTimestamp
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"log":    glassBoxLog,
	})
}

// GetLatestByActor retrieves the most recent log for a specific actor
// GET /api/v1/glass-box/latest/:actor
func (ctrl *GlassBoxController) GetLatestByActor(c *gin.Context) {
	actor := c.Param("actor")

	var log models.GlassBoxLog
	if err := ctrl.db.Where("actor = ?", actor).Order("timestamp DESC").First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No logs found for actor"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Return glass box formatted with SHA-256 and timestamp
	glassBoxLog := map[string]interface{}{
		"log_id":        log.LogID,
		"actor":         log.Actor,
		"action_type":   log.ActionType,
		"message":       log.MessageContent,
		"internal_hash": log.InternalHash, // Immutability proof
		"timestamp":     log.Timestamp.Format("2006-01-02 15:04:05"),
		"in_batch":      log.MerkleBatchID != nil,
	}

	if log.MerkleBatchID != nil {
		glassBoxLog["merkle_batch_id"] = *log.MerkleBatchID
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"log":    glassBoxLog,
	})
}
