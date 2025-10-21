/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"ares_api/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FaultVaultController handles fault vault operations
// This is CRITICAL - ensures SOLACE never repeats mistakes
type FaultVaultController struct {
	db *gorm.DB
}

// NewFaultVaultController creates a new fault vault controller
func NewFaultVaultController(db *gorm.DB) *FaultVaultController {
	return &FaultVaultController{db: db}
}

// GetSessions returns all fault vault sessions
// @Summary Get fault vault sessions
// @Description Returns all fault logging sessions with filtering
// @Tags Fault Vault
// @Produce json
// @Param active query bool false "Filter by active status"
// @Param context_type query string false "Filter by context type"
// @Param limit query int false "Limit results" default(100)
// @Success 200 {object} map[string]interface{}
// @Router /fault-vault/sessions [get]
func (fvc *FaultVaultController) GetSessions(c *gin.Context) {
	var sessions []models.FaultVaultSession
	
	query := fvc.db.Model(&models.FaultVaultSession{})
	
	// Filter by active status
	if active := c.Query("active"); active != "" {
		query = query.Where("active = ?", active == "true")
	}
	
	// Filter by context type
	if contextType := c.Query("context_type"); contextType != "" {
		query = query.Where("context_type = ?", contextType)
	}
	
	// Limit results
	limit := 100
	if limitParam := c.Query("limit"); limitParam != "" {
		c.ShouldBindQuery(&limit)
	}
	query = query.Limit(limit)
	
	// Order by most recent
	query = query.Order("started_at DESC")
	
	if err := query.Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// GetSession returns a specific fault vault session with all actions
// @Summary Get fault vault session
// @Description Returns detailed session with all actions
// @Tags Fault Vault
// @Produce json
// @Param session_id path string true "Session UUID"
// @Success 200 {object} map[string]interface{}
// @Router /fault-vault/sessions/{session_id} [get]
func (fvc *FaultVaultController) GetSession(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}
	
	var session models.FaultVaultSession
	if err := fvc.db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}
	
	// Get all actions for this session
	var actions []models.FaultVaultAction
	fvc.db.Where("session_id = ?", sessionID).Order("timestamp ASC").Find(&actions)
	
	c.JSON(http.StatusOK, gin.H{
		"session": session,
		"actions": actions,
		"action_count": len(actions),
	})
}

// GetActions returns all fault vault actions with filtering
// @Summary Get fault vault actions
// @Description Returns all logged faults/errors/warnings
// @Tags Fault Vault
// @Produce json
// @Param action_type query string false "Filter by action type"
// @Param component query string false "Filter by component"
// @Param result query string false "Filter by result"
// @Param severity_min query int false "Minimum severity"
// @Param limit query int false "Limit results" default(100)
// @Success 200 {object} map[string]interface{}
// @Router /fault-vault/actions [get]
func (fvc *FaultVaultController) GetActions(c *gin.Context) {
	var actions []models.FaultVaultAction
	
	query := fvc.db.Model(&models.FaultVaultAction{})
	
	// Filter by action type
	if actionType := c.Query("action_type"); actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	
	// Filter by component
	if component := c.Query("component"); component != "" {
		query = query.Where("component = ?", component)
	}
	
	// Filter by result
	if result := c.Query("result"); result != "" {
		query = query.Where("result = ?", result)
	}
	
	// Filter by severity
	if severityMin := c.Query("severity_min"); severityMin != "" {
		var severity int
		c.ShouldBindQuery(&severity)
		query = query.Where("severity >= ?", severity)
	}
	
	// Limit results
	limit := 100
	if limitParam := c.Query("limit"); limitParam != "" {
		c.ShouldBindQuery(&limit)
	}
	query = query.Limit(limit)
	
	// Order by most recent
	query = query.Order("timestamp DESC")
	
	if err := query.Find(&actions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"actions": actions,
		"count":   len(actions),
	})
}

// LogFault creates a new fault entry
// @Summary Log a fault
// @Description Records a fault/error/warning to prevent repetition
// @Tags Fault Vault
// @Accept json
// @Produce json
// @Param fault body LogFaultRequest true "Fault details"
// @Success 200 {object} map[string]interface{}
// @Router /fault-vault/log [post]
func (fvc *FaultVaultController) LogFault(c *gin.Context) {
	var req LogFaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get or create session
	var session models.FaultVaultSession
	if req.SessionID != uuid.Nil {
		// Use existing session
		fvc.db.Where("session_id = ?", req.SessionID).First(&session)
	}
	
	if session.SessionID == uuid.Nil {
		// Create new session
		session = models.FaultVaultSession{
			SessionID:      uuid.New(),
			UserID:         &req.UserID,
			ContextType:    req.ContextType,
			SessionSummary: req.Summary,
			Active:         true,
			StartedAt:      time.Now(),
		}
		if err := fvc.db.Create(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	
	// Create action
	action := models.FaultVaultAction{
		ActionID:     uuid.New(),
		SessionID:    session.SessionID,
		Timestamp:    time.Now(),
		Actor:        req.Actor,
		ActionType:   req.ActionType,
		Intent:       req.Intent,
		ChangesMade:  req.ChangesMade,
		Result:       req.Result,
		ErrorMessage: req.ErrorMessage,
		StackTrace:   req.StackTrace,
	}
	
	if err := fvc.db.Create(&action).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"session_id": session.SessionID,
		"action_id":  action.ActionID,
		"message":    "Fault logged successfully",
	})
}

// GetStats returns fault vault statistics
// @Summary Get fault vault statistics
// @Description Returns summary statistics about logged faults
// @Tags Fault Vault
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /fault-vault/stats [get]
func (fvc *FaultVaultController) GetStats(c *gin.Context) {
	var stats struct {
		TotalSessions   int64
		ActiveSessions  int64
		TotalActions    int64
		ErrorCount      int64
		WarningCount    int64
		AvgSeverity     float64
		MostCommonFault string
	}
	
	// Count sessions
	fvc.db.Model(&models.FaultVaultSession{}).Count(&stats.TotalSessions)
	fvc.db.Model(&models.FaultVaultSession{}).Where("active = ?", true).Count(&stats.ActiveSessions)
	
	// Count actions
	fvc.db.Model(&models.FaultVaultAction{}).Count(&stats.TotalActions)
	fvc.db.Model(&models.FaultVaultAction{}).Where("action_type = ?", "error").Count(&stats.ErrorCount)
	fvc.db.Model(&models.FaultVaultAction{}).Where("action_type = ?", "warning").Count(&stats.WarningCount)
	
	// Average severity
	fvc.db.Model(&models.FaultVaultAction{}).Select("AVG(severity)").Scan(&stats.AvgSeverity)
	
	// Most common fault
	var mostCommon struct {
		Component string
		Count     int64
	}
	fvc.db.Model(&models.FaultVaultAction{}).
		Select("component, COUNT(*) as count").
		Group("component").
		Order("count DESC").
		Limit(1).
		Scan(&mostCommon)
	stats.MostCommonFault = mostCommon.Component
	
	c.JSON(http.StatusOK, stats)
}

// LogFaultRequest is the request body for logging a fault
type LogFaultRequest struct {
	SessionID    uuid.UUID `json:"session_id,omitempty"`
	UserID       int       `json:"user_id"`
	ContextType  string    `json:"context_type"`
	Summary      string    `json:"summary"`
	Actor        string    `json:"actor"`
	ActionType   string    `json:"action_type"` // code_change, build, test, debug, crash, decision
	Intent       string    `json:"intent,omitempty"`
	ChangesMade  string    `json:"changes_made,omitempty"`
	Result       string    `json:"result"` // success, partial, failure, crash, pending
	ErrorMessage string    `json:"error_message,omitempty"`
	StackTrace   string    `json:"stack_trace,omitempty"`
}
