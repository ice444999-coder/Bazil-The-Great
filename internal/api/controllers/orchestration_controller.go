package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OrchestrationController handles SOLACE orchestration requests
type OrchestrationController struct {
	db                    *gorm.DB
	orchestrationService  interface{} // Will be *services.OrchestrationService
	repoInspectionService interface{} // Will be *services.RepoInspectionService
}

// NewOrchestrationController creates a new orchestration controller
func NewOrchestrationController(db *gorm.DB, orchestrationService interface{}, repoInspectionService interface{}) *OrchestrationController {
	return &OrchestrationController{
		db:                    db,
		orchestrationService:  orchestrationService,
		repoInspectionService: repoInspectionService,
	}
}

// SolaceUserRequest represents the solace_user_requests table
type SolaceUserRequest struct {
	ID                    uint     `json:"id" gorm:"primaryKey"`
	RequestText           string   `json:"request_text"`
	RequestType           string   `json:"request_type"`
	ComplexityScore       int      `json:"complexity_score"`
	ArchitectureRulesUsed []string `json:"architecture_rules_used" gorm:"type:text[]"`
	FilesAffected         []string `json:"files_affected" gorm:"type:text[]"`
	EstimatedInstructions int      `json:"estimated_instructions"`
	Status                string   `json:"status"`
	AnalysisNotes         string   `json:"analysis_notes"`
	FinalOutcome          string   `json:"final_outcome"`
	CreatedAt             string   `json:"created_at"`
	StartedAt             *string  `json:"started_at"`
	CompletedAt           *string  `json:"completed_at"`
}

func (SolaceUserRequest) TableName() string {
	return "solace_user_requests"
}

// GitHubInstruction represents the github_instruction_queue table
type GitHubInstruction struct {
	ID                  uint    `json:"id" gorm:"primaryKey"`
	ParentTaskID        *uint   `json:"parent_task_id"`
	InstructionSequence int     `json:"instruction_sequence"`
	InstructionText     string  `json:"instruction_text"`
	TargetFilePath      string  `json:"target_file_path"`
	ExpectedOutcome     string  `json:"expected_outcome"`
	Status              string  `json:"status"`
	GitHubResponse      string  `json:"github_response"`
	VerificationNotes   string  `json:"verification_notes"`
	RetryCount          int     `json:"retry_count"`
	CreatedAt           string  `json:"created_at"`
	StartedAt           *string `json:"started_at"`
	CompletedAt         *string `json:"completed_at"`
	VerifiedAt          *string `json:"verified_at"`
}

func (GitHubInstruction) TableName() string {
	return "github_instruction_queue"
}

// CreateUserRequest creates a new user request for SOLACE to orchestrate
// POST /api/v1/solace/orchestrate/request
func (c *OrchestrationController) CreateUserRequest(ctx *gin.Context) {
	var req struct {
		RequestText string `json:"request_text" binding:"required"`
		RequestType string `json:"request_type"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userRequest := SolaceUserRequest{
		RequestText: req.RequestText,
		RequestType: req.RequestType,
		Status:      "analyzing",
	}

	if err := c.db.Create(&userRequest).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	ctx.JSON(http.StatusCreated, userRequest)
}

// GetUserRequest retrieves a specific user request
// GET /api/v1/solace/orchestrate/request/:id
func (c *OrchestrationController) GetUserRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	var request SolaceUserRequest
	if err := c.db.First(&request, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	ctx.JSON(http.StatusOK, request)
}

// GetInstructionsForRequest retrieves all GitHub instructions for a user request
// GET /api/v1/solace/orchestrate/request/:id/instructions
func (c *OrchestrationController) GetInstructionsForRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	var instructions []GitHubInstruction
	if err := c.db.Where("parent_task_id = ?", id).Order("instruction_sequence ASC").Find(&instructions).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch instructions"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"parent_task_id": id,
		"instructions":   instructions,
		"count":          len(instructions),
	})
}

// UpdateInstructionStatus updates the status of a GitHub instruction
// PUT /api/v1/solace/orchestrate/instruction/:id
func (c *OrchestrationController) UpdateInstructionStatus(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		Status            string `json:"status"`
		GitHubResponse    string `json:"github_response"`
		VerificationNotes string `json:"verification_notes"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"status": req.Status,
	}

	if req.GitHubResponse != "" {
		updates["github_response"] = req.GitHubResponse
	}

	if req.VerificationNotes != "" {
		updates["verification_notes"] = req.VerificationNotes
	}

	now := time.Now().Format(time.RFC3339)
	if req.Status == "in_progress" {
		updates["started_at"] = now
	} else if req.Status == "completed" {
		updates["completed_at"] = now
	} else if req.Status == "verified" {
		updates["verified_at"] = now
	}

	if err := c.db.Model(&GitHubInstruction{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instruction"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Instruction updated successfully"})
}

// ScanRepository triggers a full repository scan
// POST /api/v1/solace/orchestrate/scan
func (c *OrchestrationController) ScanRepository(ctx *gin.Context) {
	if c.repoInspectionService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Repo inspection service not available"})
		return
	}

	// Type assertion to call the service method
	type RepoScanner interface {
		ScanRepository() error
	}

	scanner, ok := c.repoInspectionService.(RepoScanner)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid repo inspection service"})
		return
	}

	if err := scanner.ScanRepository(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Scan failed: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Repository scan completed successfully",
		"status":  "complete",
	})
}

// AnalyzeRequest triggers analysis of a user request
// POST /api/v1/solace/orchestrate/analyze/:id
func (c *OrchestrationController) AnalyzeRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	if c.orchestrationService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Orchestration service not available"})
		return
	}

	// Type assertion to call the service method
	type RequestAnalyzer interface {
		AnalyzeRequest(requestID uint) error
	}

	analyzer, ok := c.orchestrationService.(RequestAnalyzer)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid orchestration service"})
		return
	}

	// Convert string ID to uint
	var requestID uint
	fmt.Sscanf(id, "%d", &requestID)

	if err := analyzer.AnalyzeRequest(requestID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Analysis failed: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Analysis completed successfully",
		"request_id": id,
		"status":     "planned",
	})
}

// GenerateInstructions generates atomic GitHub instructions for a request
// POST /api/v1/solace/orchestrate/generate/:id
func (c *OrchestrationController) GenerateInstructions(ctx *gin.Context) {
	id := ctx.Param("id")

	if c.orchestrationService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Orchestration service not available"})
		return
	}

	// Type assertion to call the service method
	type InstructionGenerator interface {
		GenerateInstructions(requestID uint) ([]interface{}, error)
	}

	generator, ok := c.orchestrationService.(InstructionGenerator)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid orchestration service"})
		return
	}

	// Convert string ID to uint
	var requestID uint
	fmt.Sscanf(id, "%d", &requestID)

	instructions, err := generator.GenerateInstructions(requestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Generation failed: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Instructions generated successfully",
		"request_id":   id,
		"instructions": instructions,
		"count":        len(instructions),
		"status":       "executing",
	})
}

// GetNextInstruction gets the next pending instruction for GitHub
// GET /api/v1/solace/orchestrate/next
func (c *OrchestrationController) GetNextInstruction(ctx *gin.Context) {
	var instruction GitHubInstruction

	if err := c.db.Where("status = ?", "pending").Order("instruction_sequence ASC").First(&instruction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusOK, gin.H{
				"message":  "No pending instructions",
				"has_next": false,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch instruction"})
		return
	}

	// Update status to in_progress
	now := time.Now().Format(time.RFC3339)
	c.db.Model(&GitHubInstruction{}).Where("id = ?", instruction.ID).Updates(map[string]interface{}{
		"status":     "in_progress",
		"started_at": now,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"has_next":    true,
		"instruction": instruction,
	})
}
