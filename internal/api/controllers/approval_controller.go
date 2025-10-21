package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ApprovalRequest represents the database model for approval requests
type ApprovalRequest struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	SubtaskID   string     `gorm:"unique;not null" json:"subtask_id"`
	Status      string     `gorm:"not null;default:pending" json:"status"` // pending, approved, rejected
	Description string     `gorm:"not null" json:"description"`
	RequestedAt time.Time  `gorm:"not null" json:"requested_at"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	ApprovedBy  string     `json:"approved_by,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// ApprovalController handles manual approval gates for Grok protocol
type ApprovalController struct {
	db *gorm.DB
}

// NewApprovalController creates a new approval controller
func NewApprovalController(db *gorm.DB) *ApprovalController {
	// Auto-migrate the approval_requests table
	if err := db.AutoMigrate(&ApprovalRequest{}); err != nil {
		log.Printf("⚠️ Failed to auto-migrate approval_requests: %v", err)
	}
	return &ApprovalController{db: db}
}

// RequestApproval logs a new approval request
// POST /api/approve/request
func (ac *ApprovalController) RequestApproval(c *gin.Context) {
	var req struct {
		SubtaskID   string `json:"subtask_id" binding:"required"`
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create or update approval request
	approval := ApprovalRequest{
		SubtaskID:   req.SubtaskID,
		Status:      "pending",
		Description: req.Description,
		RequestedAt: time.Now(),
	}

	// Use GORM's Upsert pattern (create or update on conflict)
	result := ac.db.Where("subtask_id = ?", req.SubtaskID).Assign(approval).FirstOrCreate(&approval)
	if result.Error != nil {
		log.Printf("❌ Failed to create approval request: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create approval request"})
		return
	}

	log.Printf("⏳ Approval requested for %s: %s", req.SubtaskID, req.Description)
	c.JSON(http.StatusOK, approval)
}

// ApproveSubtask approves a pending subtask
// POST /api/approve/:subtask_id
func (ac *ApprovalController) ApproveSubtask(c *gin.Context) {
	subtaskID := c.Param("subtask_id")
	
	var req struct {
		ApprovedBy string `json:"approved_by"`
		Notes      string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow approval without JSON body
		req.ApprovedBy = "dashboard_user"
	}

	now := time.Now()
	
	// Find pending approval
	var approval ApprovalRequest
	result := ac.db.Where("subtask_id = ? AND status = ?", subtaskID, "pending").First(&approval)
	
	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending approval found for this subtask"})
		return
	} else if result.Error != nil {
		log.Printf("❌ Failed to find approval: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find approval"})
		return
	}

	// Update to approved
	approval.Status = "approved"
	approval.ApprovedAt = &now
	approval.ApprovedBy = req.ApprovedBy
	approval.Notes = req.Notes

	if err := ac.db.Save(&approval).Error; err != nil {
		log.Printf("❌ Failed to approve subtask: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve subtask"})
		return
	}

	log.Printf("✅ Approved %s by %s", subtaskID, req.ApprovedBy)
	c.JSON(http.StatusOK, approval)
}

// RejectSubtask rejects a pending subtask
// POST /api/approve/:subtask_id/reject
func (ac *ApprovalController) RejectSubtask(c *gin.Context) {
	subtaskID := c.Param("subtask_id")
	
	var req struct {
		ApprovedBy string `json:"approved_by"`
		Notes      string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.ApprovedBy = "dashboard_user"
	}

	now := time.Now()
	
	// Find pending approval
	var approval ApprovalRequest
	result := ac.db.Where("subtask_id = ? AND status = ?", subtaskID, "pending").First(&approval)
	
	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending approval found for this subtask"})
		return
	} else if result.Error != nil {
		log.Printf("❌ Failed to find approval: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find approval"})
		return
	}

	// Update to rejected
	approval.Status = "rejected"
	approval.ApprovedAt = &now
	approval.ApprovedBy = req.ApprovedBy
	approval.Notes = req.Notes

	if err := ac.db.Save(&approval).Error; err != nil {
		log.Printf("❌ Failed to reject subtask: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject subtask"})
		return
	}

	log.Printf("🚫 Rejected %s by %s: %s", subtaskID, req.ApprovedBy, req.Notes)
	c.JSON(http.StatusOK, approval)
}

// GetApprovalStatus retrieves the current status of a subtask
// GET /api/approve/:subtask_id
func (ac *ApprovalController) GetApprovalStatus(c *gin.Context) {
	subtaskID := c.Param("subtask_id")

	var approval ApprovalRequest
	result := ac.db.Where("subtask_id = ?", subtaskID).Order("requested_at DESC").First(&approval)

	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{"status": "not_requested"})
		return
	} else if result.Error != nil {
		log.Printf("❌ Failed to get approval status: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get approval status"})
		return
	}

	c.JSON(http.StatusOK, approval)
}

// ListPendingApprovals returns all pending approval requests
// GET /api/approve/pending
func (ac *ApprovalController) ListPendingApprovals(c *gin.Context) {
	var approvals []ApprovalRequest
	result := ac.db.Where("status = ?", "pending").Order("requested_at ASC").Find(&approvals)

	if result.Error != nil {
		log.Printf("❌ Failed to list pending approvals: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list approvals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pending_approvals": approvals, "count": len(approvals)})
}

// ApproveAll approves all pending subtasks (for final merge)
// POST /api/approve/all
func (ac *ApprovalController) ApproveAll(c *gin.Context) {
	var req struct {
		ApprovedBy string `json:"approved_by"`
		Notes      string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.ApprovedBy = "dashboard_user"
		req.Notes = "Batch approval - all tests passed"
	}

	now := time.Now()

	// Get all pending approvals
	var approvals []ApprovalRequest
	ac.db.Where("status = ?", "pending").Find(&approvals)

	// Update all to approved
	updates := map[string]interface{}{
		"status":      "approved",
		"approved_at": now,
		"approved_by": req.ApprovedBy,
		"notes":       req.Notes,
	}

	result := ac.db.Model(&ApprovalRequest{}).Where("status = ?", "pending").Updates(updates)
	if result.Error != nil {
		log.Printf("❌ Failed to approve all: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve all"})
		return
	}

	var subtaskIDs []string
	for _, approval := range approvals {
		subtaskIDs = append(subtaskIDs, approval.SubtaskID)
	}

	log.Printf("✅ Approved all %d pending subtasks by %s", len(approvals), req.ApprovedBy)
	c.JSON(http.StatusOK, gin.H{
		"approved_count":    len(approvals),
		"approved_subtasks": subtaskIDs,
		"message":           "All pending subtasks approved",
	})
}

// WaitForApproval polls for approval status (used in automated flows)
// This blocks until approval is granted or timeout
func (ac *ApprovalController) WaitForApproval(subtaskID string, timeoutMinutes int) (bool, error) {
	timeout := time.After(time.Duration(timeoutMinutes) * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Printf("⏳ Waiting for approval: %s (timeout: %dm)", subtaskID, timeoutMinutes)

	for {
		select {
		case <-timeout:
			log.Printf("⏰ Approval timeout for %s after %dm", subtaskID, timeoutMinutes)
			return false, nil
		case <-ticker.C:
			var approval ApprovalRequest
			result := ac.db.Where("subtask_id = ?", subtaskID).Order("requested_at DESC").First(&approval)
			
			if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
				return false, result.Error
			}

			if approval.Status == "approved" {
				log.Printf("✅ Approval granted for %s", subtaskID)
				return true, nil
			} else if approval.Status == "rejected" {
				log.Printf("🚫 Approval rejected for %s", subtaskID)
				return false, nil
			}
		}
	}
}
