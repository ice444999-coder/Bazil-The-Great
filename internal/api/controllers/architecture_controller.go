/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ArchitectureController handles architecture rules queries for SOLACE
type ArchitectureController struct {
	db *gorm.DB
}

// NewArchitectureController creates a new architecture controller
func NewArchitectureController(db *gorm.DB) *ArchitectureController {
	return &ArchitectureController{db: db}
}

// ArchitectureRule represents the architecture_rules table
type ArchitectureRule struct {
	ID                uint     `json:"id" gorm:"primaryKey"`
	FeatureType       string   `json:"feature_type"`
	BackendPattern    string   `json:"backend_pattern"`
	FrontendPattern   string   `json:"frontend_pattern"`
	IntegrationPoints []string `json:"integration_points" gorm:"type:text[]"`
	RulesDescription  string   `json:"rules_description"`
	Examples          []string `json:"examples" gorm:"type:text[]"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

func (ArchitectureRule) TableName() string {
	return "architecture_rules"
}

// GetAllRules returns all architecture patterns
// GET /api/v1/solace/architecture/rules
func (c *ArchitectureController) GetAllRules(ctx *gin.Context) {
	var rules []ArchitectureRule

	if err := c.db.Find(&rules).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch architecture rules"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// GetRuleByType returns architecture pattern for specific feature type
// GET /api/v1/solace/architecture/rules/:type
func (c *ArchitectureController) GetRuleByType(ctx *gin.Context) {
	featureType := ctx.Param("type")

	var rule ArchitectureRule
	if err := c.db.Where("feature_type = ?", featureType).First(&rule).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Architecture rule not found"})
		return
	}

	ctx.JSON(http.StatusOK, rule)
}

// SearchRules searches architecture rules by keyword
// POST /api/v1/solace/architecture/search
func (c *ArchitectureController) SearchRules(ctx *gin.Context) {
	var req struct {
		Keyword string `json:"keyword" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var rules []ArchitectureRule
	query := "%" + req.Keyword + "%"

	if err := c.db.Where("feature_type ILIKE ? OR rules_description ILIKE ?", query, query).Find(&rules).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"rules":   rules,
		"count":   len(rules),
		"keyword": req.Keyword,
	})
}
