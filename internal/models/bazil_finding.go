package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BazilFinding stores code analysis results with confidence scores
type BazilFinding struct {
	gorm.Model
	FaultType   string          `json:"fault_type"`
	Description string          `json:"description"`
	FilePath    string          `json:"file_path"`
	LineNumber  int             `json:"line_number"`
	Confidence  float64         `json:"confidence"` // 0-1.0 score
	UUID        uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid()" json:"uuid"`
	Status      string          `json:"status"` // pending, approved, rejected, fixed
	ReviewedBy  uint            `json:"reviewed_by,omitempty"`
	ReviewedAt  *gorm.DeletedAt `json:"reviewed_at,omitempty"`
}

// BazilPatchApproval stores human approval decisions
type BazilPatchApproval struct {
	gorm.Model
	PatchID      string          `json:"patch_id" gorm:"uniqueIndex"`
	FindingIDs   string          `json:"finding_ids"` // JSON array of UUIDs
	PatchContent string          `json:"patch_content" gorm:"type:text"`
	Status       string          `json:"status"` // pending, approved, rejected, applied
	ApprovedBy   uint            `json:"approved_by,omitempty"`
	AppliedAt    *gorm.DeletedAt `json:"applied_at,omitempty"`
	BranchName   string          `json:"branch_name"`
	TestResult   string          `json:"test_result" gorm:"type:text"`
}
