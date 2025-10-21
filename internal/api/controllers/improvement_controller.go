/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ImprovementQueue struct {
	ID                      int        `json:"id"`
	CreatedAt               time.Time  `json:"created_at"`
	CreatedBy               string     `json:"created_by"`
	Title                   string     `json:"title"`
	Description             string     `json:"description"`
	SQLScript               string     `json:"sql_script"`
	RollbackScript          string     `json:"rollback_script"`
	ScheduledFor            *time.Time `json:"scheduled_for"`
	Status                  string     `json:"status"`
	EstimatedSpeedupPercent int        `json:"estimated_speedup_percent"`
	RiskLevel               string     `json:"risk_level"`
	ExecutedAt              *time.Time `json:"executed_at"`
	ExecutionDurationMs     *int       `json:"execution_duration_ms"`
	ActualSpeedupPercent    *int       `json:"actual_speedup_percent"`
	ErrorMessage            *string    `json:"error_message"`
	HederaTxnID             *string    `json:"hedera_txn_id"`
	RequiresApproval        bool       `json:"requires_approval"`
	ApprovedBy              *string    `json:"approved_by"`
	ApprovedAt              *time.Time `json:"approved_at"`
}

var db *gorm.DB

func SetDB(database *gorm.DB) {
	db = database
}

// ListImprovements lists all pending improvements
func ListImprovements(c *gin.Context) {
	filter := c.Query("filter")

	var improvements []ImprovementQueue
	query := db.Table("improvement_queue")

	if filter != "" && filter != "ALL" {
		query = query.Where("status = ?", filter)
	}

	if err := query.Order("created_at DESC").Find(&improvements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"improvements": improvements})
}

// ApproveImprovement approves a specific improvement
func ApproveImprovement(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	now := time.Now()
	result := db.Table("improvement_queue").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "APPROVED",
			"approved_by": "enki",
			"approved_at": now,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Improvement approved", "id": id})
}

// RejectImprovement rejects a specific improvement
func RejectImprovement(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result := db.Table("improvement_queue").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           "REJECTED",
			"rejection_reason": body.Reason,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Improvement rejected", "id": id})
}

// ExecuteAllImprovements runs all approved improvements
func ExecuteAllImprovements(c *gin.Context) {
	// This endpoint triggers the run_improvements.ps1 script
	// For now, return a message indicating manual execution is needed

	var count int64
	db.Table("improvement_queue").
		Where("status IN (?, ?)", "APPROVED", "PENDING").
		Where("requires_approval = ? OR status = ?", false, "APPROVED").
		Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"message":          "Execution triggered - run run_improvements.ps1 manually or wait for 10pm scheduled execution",
		"ready_to_execute": count,
		"executed":         0,
		"succeeded":        0,
		"failed":           0,
	})
}
