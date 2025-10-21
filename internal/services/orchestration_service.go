/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package services

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// OrchestrationService handles breaking down user requests into atomic GitHub instructions
type OrchestrationService struct {
	db *gorm.DB
}

// NewOrchestrationService creates a new orchestration service
func NewOrchestrationService(db *gorm.DB) *OrchestrationService {
	return &OrchestrationService{db: db}
}

// ArchitectureRule represents architecture patterns from database
type ArchitectureRule struct {
	ID                uint
	FeatureType       string
	BackendPattern    string
	FrontendPattern   string
	IntegrationPoints []string
	RulesDescription  string
	Examples          []string
}

func (ArchitectureRule) TableName() string {
	return "architecture_rules"
}

// UserRequest represents a user's feature request
type UserRequest struct {
	ID                    uint
	RequestText           string
	RequestType           string
	ComplexityScore       int
	ArchitectureRulesUsed []string
	FilesAffected         []string
	EstimatedInstructions int
	Status                string
	AnalysisNotes         string
}

func (UserRequest) TableName() string {
	return "solace_user_requests"
}

// GitHubInstruction represents an atomic instruction for GitHub
type GitHubInstruction struct {
	ID                  uint
	ParentTaskID        *uint
	InstructionSequence int
	InstructionText     string
	TargetFilePath      string
	ExpectedOutcome     string
	Status              string
}

func (GitHubInstruction) TableName() string {
	return "github_instruction_queue"
}

// AnalyzeRequest analyzes a user request and determines which architecture rules apply
func (s *OrchestrationService) AnalyzeRequest(requestID uint) error {
	// Fetch the user request
	var request UserRequest
	if err := s.db.First(&request, requestID).Error; err != nil {
		return fmt.Errorf("failed to fetch request: %w", err)
	}

	// Fetch all architecture rules
	var rules []ArchitectureRule
	if err := s.db.Find(&rules).Error; err != nil {
		return fmt.Errorf("failed to fetch architecture rules: %w", err)
	}

	// Simple keyword matching to determine which rules apply
	requestLower := strings.ToLower(request.RequestText)
	var applicableRules []string
	var filesAffected []string
	complexityScore := 1

	for _, rule := range rules {
		// Check if request mentions this feature type
		if strings.Contains(requestLower, strings.ToLower(rule.FeatureType)) ||
			strings.Contains(requestLower, "agent") && strings.Contains(rule.FeatureType, "agent") ||
			strings.Contains(requestLower, "trading") && strings.Contains(rule.FeatureType, "trading") {

			applicableRules = append(applicableRules, rule.FeatureType)

			// Add affected files from patterns
			if rule.BackendPattern != "" {
				filesAffected = append(filesAffected, rule.BackendPattern)
			}
			if rule.FrontendPattern != "" {
				filesAffected = append(filesAffected, rule.FrontendPattern)
			}
			filesAffected = append(filesAffected, rule.IntegrationPoints...)

			complexityScore += 2 // Each rule adds complexity
		}
	}

	// Estimate number of GitHub instructions needed
	estimatedInstructions := len(filesAffected) * 2 // Rough estimate: 2 instructions per file

	// Update the request with analysis
	updates := map[string]interface{}{
		"architecture_rules_used": applicableRules,
		"files_affected":          filesAffected,
		"complexity_score":        complexityScore,
		"estimated_instructions":  estimatedInstructions,
		"status":                  "planned",
		"analysis_notes":          fmt.Sprintf("Identified %d applicable architecture patterns, %d files affected", len(applicableRules), len(filesAffected)),
	}

	if err := s.db.Model(&UserRequest{}).Where("id = ?", requestID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	return nil
}

// GenerateInstructions generates atomic GitHub instructions for a user request
func (s *OrchestrationService) GenerateInstructions(requestID uint) ([]GitHubInstruction, error) {
	// Fetch the analyzed request
	var request UserRequest
	if err := s.db.First(&request, requestID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch request: %w", err)
	}

	if request.Status != "planned" {
		return nil, fmt.Errorf("request must be in 'planned' status, current status: %s", request.Status)
	}

	var instructions []GitHubInstruction
	sequence := 1

	// Generate instructions based on files affected
	for _, filePath := range request.FilesAffected {
		instruction := GitHubInstruction{
			ParentTaskID:        &request.ID,
			InstructionSequence: sequence,
			InstructionText:     fmt.Sprintf("Modify or create file: %s", filePath),
			TargetFilePath:      filePath,
			ExpectedOutcome:     fmt.Sprintf("File %s updated with required changes", filePath),
			Status:              "pending",
		}
		instructions = append(instructions, instruction)
		sequence++
	}

	// Save instructions to database
	for i := range instructions {
		if err := s.db.Create(&instructions[i]).Error; err != nil {
			return nil, fmt.Errorf("failed to create instruction %d: %w", i+1, err)
		}
	}

	// Update request status
	s.db.Model(&UserRequest{}).Where("id = ?", requestID).Update("status", "executing")

	return instructions, nil
}
