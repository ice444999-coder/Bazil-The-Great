package handlers

import (
	"fmt"
	"net/http"
	"time"

	"ares_api/internal/agent"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ForgeHandler handles FORGE apprenticeship learning endpoints
type ForgeHandler struct {
	DB            *gorm.DB
	GitHubCopilot *agent.GitHubCopilot
}

// NewForgeHandler creates a new FORGE handler
func NewForgeHandler(db *gorm.DB) *ForgeHandler {
	return &ForgeHandler{
		DB:            db,
		GitHubCopilot: agent.NewGitHubCopilot(),
	}
}

// ObservationRequest represents a FORGE observation request
type ObservationRequest struct {
	PatternType      string  `json:"pattern_type"`
	TaskDescription  string  `json:"task_description"`
	UserPrompt       string  `json:"user_prompt"`
	ForgeConfidence  float64 `json:"forge_confidence_before"`
	PatternCategory  string  `json:"pattern_category,omitempty"`
	RelatedCrystalID *int    `json:"related_crystal_id,omitempty"`
}

// ObservationResponse represents the response after recording an observation
type ObservationResponse struct {
	ObservationID    int      `json:"observation_id"`
	GitHubCode       string   `json:"github_generated_code"`
	Principles       []string `json:"forge_extracted_principles"`
	ConfidenceBefore float64  `json:"forge_confidence_before"`
	ConfidenceAfter  float64  `json:"forge_confidence_after"`
	Success          bool     `json:"success"`
	GenerationTime   int64    `json:"generation_time_ms"`
	Message          string   `json:"message"`
}

// GraduationDashboard represents graduation status for patterns
type GraduationDashboard struct {
	PatternType     string     `json:"pattern_type"`
	ExamplesCount   int        `json:"examples_count"`
	AvgConfidence   float64    `json:"avg_confidence"`
	SuccessRate     float64    `json:"success_rate"`
	GraduationReady bool       `json:"graduation_ready"`
	GraduatedAt     *time.Time `json:"graduated_at,omitempty"`
}

// LearningProgress represents FORGE learning progress
type LearningProgress struct {
	PatternType         string    `json:"pattern_type"`
	TotalExamples       int       `json:"total_examples"`
	SuccessfulExamples  int       `json:"successful_examples"`
	AvgConfidenceBefore float64   `json:"avg_confidence_before"`
	AvgConfidenceAfter  float64   `json:"avg_confidence_after"`
	ConfidenceGrowth    float64   `json:"confidence_growth"`
	LastObservation     time.Time `json:"last_observation"`
}

// GetConfidenceDashboard returns graduation dashboard
// GET /api/v1/forge/confidence
func (h *ForgeHandler) GetConfidenceDashboard(c *gin.Context) {
	var dashboard []GraduationDashboard

	err := h.DB.Raw(`
		SELECT 
			pattern_type,
			examples_count,
			avg_confidence,
			success_rate,
			graduation_ready,
			graduated_at
		FROM forge_graduation_dashboard
		ORDER BY examples_count DESC, avg_confidence DESC
	`).Scan(&dashboard).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dashboard": dashboard,
		"timestamp": time.Now(),
	})
}

// RecordObservation records FORGE observing GitHub Copilot
// POST /api/v1/forge/observe
func (h *ForgeHandler) RecordObservation(c *gin.Context) {
	var req ObservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate required fields
	if req.PatternType == "" || req.TaskDescription == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pattern_type and task_description are required"})
		return
	}

	// Ask GitHub Copilot to generate code
	copilotResp, err := h.GitHubCopilot.GenerateCode(req.TaskDescription, "go")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("GitHub Copilot error: %v", err)})
		return
	}

	// Extract learning principles from generated code
	principles, _ := h.GitHubCopilot.ExtractPrinciples(copilotResp.Code, req.PatternType)

	// Calculate confidence after observation (Bayesian update)
	confidenceAfter := req.ForgeConfidence
	if copilotResp.Success {
		confidenceAfter = req.ForgeConfidence + (1.0-req.ForgeConfidence)*0.15 // 15% learning rate
	}

	// Insert observation into database
	var observationID int
	err = h.DB.Raw(`
		INSERT INTO forge_confidence_tracker (
			pattern_type,
			pattern_category,
			task_description,
			user_prompt,
			github_generated_code,
			github_model_used,
			generation_timestamp,
			forge_observation,
			forge_extracted_principles,
			forge_confidence_before,
			forge_confidence_after,
			success,
			execution_time_ms,
			related_crystal_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`,
		req.PatternType,
		req.PatternCategory,
		req.TaskDescription,
		req.UserPrompt,
		copilotResp.Code,
		copilotResp.Model,
		copilotResp.GeneratedAt,
		fmt.Sprintf("FORGE observed GitHub Copilot generate %s pattern", req.PatternType),
		principles,
		req.ForgeConfidence,
		confidenceAfter,
		copilotResp.Success,
		copilotResp.ExecutionTime,
		req.RelatedCrystalID,
	).Scan(&observationID).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, gin.H{
		"observation_id":             observationID,
		"github_generated_code":      copilotResp.Code,
		"forge_extracted_principles": principles,
		"forge_confidence_before":    req.ForgeConfidence,
		"forge_confidence_after":     confidenceAfter,
		"success":                    copilotResp.Success,
		"generation_time_ms":         copilotResp.ExecutionTime,
		"message":                    fmt.Sprintf("FORGE observation recorded. Confidence: %.1f%% → %.1f%%", req.ForgeConfidence*100, confidenceAfter*100),
	})
}

// GetLearningProgress returns FORGE learning progress by pattern
// GET /api/v1/forge/learning-progress
func (h *ForgeHandler) GetLearningProgress(c *gin.Context) {
	var progress []LearningProgress

	err := h.DB.Raw(`
		SELECT 
			pattern_type,
			total_examples,
			successful_examples,
			avg_confidence_before,
			avg_confidence_after,
			confidence_growth,
			last_observation
		FROM forge_learning_progress
		ORDER BY confidence_growth DESC, total_examples DESC
	`).Scan(&progress).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"learning_progress": progress,
		"timestamp":         time.Now(),
	})
}

// GraduatePattern manually graduates a pattern
// POST /api/v1/forge/graduate
func (h *ForgeHandler) GraduatePattern(c *gin.Context) {
	var req struct {
		PatternType string `json:"pattern_type"`
		GraduatedBy string `json:"graduated_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	if req.PatternType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pattern_type is required"})
		return
	}

	// Check if pattern is ready for graduation
	var ready bool
	err := h.DB.Raw(`SELECT check_graduation_ready(?)`, req.PatternType).Scan(&ready).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", err)})
		return
	}

	if !ready {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pattern does not meet graduation criteria (>=95% confidence, >=20 examples)"})
		return
	}

	// Graduate the pattern
	result := h.DB.Exec(`
		UPDATE forge_confidence_tracker
		SET graduation_ready = true,
		    graduated_at = NOW(),
		    graduated_by = ?
		WHERE pattern_type = ?
		  AND graduation_ready = false
	`, req.GraduatedBy, req.PatternType)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", result.Error)})
		return
	}

	rowsAffected := result.RowsAffected

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"pattern_type":   req.PatternType,
		"rows_graduated": rowsAffected,
		"message":        fmt.Sprintf("Pattern '%s' graduated! FORGE can now handle this autonomously.", req.PatternType),
		"timestamp":      time.Now(),
	})
}
