/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MissionProgress represents Phase 1 mission completion status
type MissionProgress struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Phase             int       `gorm:"not null;default:1" json:"phase"`
	Percentage        int       `gorm:"not null;default:0" json:"percentage"`
	Status            string    `gorm:"not null;default:'initializing'" json:"status"`
	SubtasksCompleted int       `gorm:"not null;default:0" json:"subtasks_completed"`
	SubtasksTotal     int       `gorm:"not null;default:12" json:"subtasks_total"`
	LastUpdated       time.Time `gorm:"not null" json:"last_updated"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// MissionController handles mission progress tracking
type MissionController struct {
	db *gorm.DB
}

// NewMissionController creates a new mission controller
func NewMissionController(db *gorm.DB) *MissionController {
	mc := &MissionController{db: db}

	// Auto-migrate the mission_progress table
	if err := db.AutoMigrate(&MissionProgress{}); err != nil {
		panic("Failed to migrate mission_progress table: " + err.Error())
	}

	// Initialize Phase 1 progress if not exists
	var progress MissionProgress
	if err := db.Where("phase = ?", 1).First(&progress).Error; err == gorm.ErrRecordNotFound {
		initialProgress := MissionProgress{
			Phase:             1,
			Percentage:        0,
			Status:            "initializing",
			SubtasksCompleted: 0,
			SubtasksTotal:     12,
			LastUpdated:       time.Now(),
		}
		db.Create(&initialProgress)
	}

	return mc
}

// GetProgress returns current mission progress
// GET /api/mission/progress
func (mc *MissionController) GetProgress(c *gin.Context) {
	var progress MissionProgress

	// Get Phase 1 progress
	if err := mc.db.Where("phase = ?", 1).First(&progress).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Mission progress not found",
		})
		return
	}

	// Calculate dynamic percentage based on completed subtasks
	completionPercentage := (progress.SubtasksCompleted * 100) / progress.SubtasksTotal

	// Update status message based on percentage
	var statusMessage string
	switch {
	case completionPercentage < 25:
		statusMessage = "🔧 Initializing core systems..."
	case completionPercentage < 50:
		statusMessage = "🚀 Trading strategies loading..."
	case completionPercentage < 75:
		statusMessage = "🧠 SOLACE consciousness online..."
	case completionPercentage < 100:
		statusMessage = "⚡ Advanced features activating..."
	default:
		statusMessage = "✅ Phase 1 operational - Ready for action!"
	}

	c.JSON(http.StatusOK, gin.H{
		"phase":              progress.Phase,
		"percentage":         completionPercentage,
		"status":             statusMessage,
		"subtasks_completed": progress.SubtasksCompleted,
		"subtasks_total":     progress.SubtasksTotal,
		"last_updated":       progress.LastUpdated,
	})
}

// UpdateProgress updates mission completion percentage
// POST /api/mission/progress
func (mc *MissionController) UpdateProgress(c *gin.Context) {
	var req struct {
		SubtasksCompleted int    `json:"subtasks_completed" binding:"required,min=0,max=12"`
		Notes             string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	var progress MissionProgress
	if err := mc.db.Where("phase = ?", 1).First(&progress).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Mission progress not found",
		})
		return
	}

	// Update progress
	progress.SubtasksCompleted = req.SubtasksCompleted
	progress.Percentage = (req.SubtasksCompleted * 100) / progress.SubtasksTotal
	progress.LastUpdated = time.Now()

	// Auto-update status based on new percentage
	switch {
	case progress.Percentage < 25:
		progress.Status = "initializing"
	case progress.Percentage < 50:
		progress.Status = "loading_strategies"
	case progress.Percentage < 75:
		progress.Status = "solace_online"
	case progress.Percentage < 100:
		progress.Status = "advanced_features"
	default:
		progress.Status = "operational"
	}

	if err := mc.db.Save(&progress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update progress: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Mission progress updated",
		"phase":              progress.Phase,
		"percentage":         progress.Percentage,
		"status":             progress.Status,
		"subtasks_completed": progress.SubtasksCompleted,
		"subtasks_total":     progress.SubtasksTotal,
		"notes":              req.Notes,
	})
}

// IncrementProgress increments subtask completion by 1
// POST /api/mission/progress/increment
func (mc *MissionController) IncrementProgress(c *gin.Context) {
	var req struct {
		SubtaskName string `json:"subtask_name"`
		Notes       string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	var progress MissionProgress
	if err := mc.db.Where("phase = ?", 1).First(&progress).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Mission progress not found",
		})
		return
	}

	// Increment if not already at max
	if progress.SubtasksCompleted < progress.SubtasksTotal {
		progress.SubtasksCompleted++
		progress.Percentage = (progress.SubtasksCompleted * 100) / progress.SubtasksTotal
		progress.LastUpdated = time.Now()

		// Auto-update status
		switch {
		case progress.Percentage < 25:
			progress.Status = "initializing"
		case progress.Percentage < 50:
			progress.Status = "loading_strategies"
		case progress.Percentage < 75:
			progress.Status = "solace_online"
		case progress.Percentage < 100:
			progress.Status = "advanced_features"
		default:
			progress.Status = "operational"
		}

		if err := mc.db.Save(&progress).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to increment progress: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Subtask completed! 🎯",
		"subtask_name":       req.SubtaskName,
		"phase":              progress.Phase,
		"percentage":         progress.Percentage,
		"status":             progress.Status,
		"subtasks_completed": progress.SubtasksCompleted,
		"subtasks_total":     progress.SubtasksTotal,
		"notes":              req.Notes,
	})
}
