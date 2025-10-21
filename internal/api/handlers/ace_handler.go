/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package handlers

import (
	"ares_api/internal/ace"
	"ares_api/internal/services"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ACEHandler handles ACE Framework API endpoints
type ACEHandler struct {
	orchestrator    *ace.ACEOrchestrator
	db              *gorm.DB
	assessmentQueue chan *ACEAssessmentRequest
	mu              sync.RWMutex
}

// ACEAssessmentRequest represents a quality assessment request
type ACEAssessmentRequest struct {
	UserID   uint
	Message  string
	Response string
	Context  map[string]interface{}
}

// NewACEHandler creates a new ACE handler
func NewACEHandler(orchestrator *ace.ACEOrchestrator, db *gorm.DB) *ACEHandler {
	handler := &ACEHandler{
		orchestrator:    orchestrator,
		db:              db,
		assessmentQueue: make(chan *ACEAssessmentRequest, 100),
	}

	// Start background worker for quality assessments
	go handler.processAssessments()

	return handler
}

// QueueAssessment queues a response for ACE quality assessment (non-blocking)
func (h *ACEHandler) QueueAssessment(req *ACEAssessmentRequest) {
	select {
	case h.assessmentQueue <- req:
		log.Printf("🧠 ACE: Queued assessment for user %d", req.UserID)
	default:
		log.Printf("⚠️ ACE: Assessment queue full, dropping request for user %d", req.UserID)
	}
}

// processAssessments processes quality assessments in background
func (h *ACEHandler) processAssessments() {
	for req := range h.assessmentQueue {
		// Build decision context for ACE
		decisionCtx := ace.DecisionContext{
			DecisionType: "chat-response",
			UserMessage:  req.Message,
			InputContext: req.Context,
		}

		// Run complete ACE consciousness cycle
		decision, scores, err := h.orchestrator.CompleteDecisionCycle(decisionCtx, req.Response)
		if err != nil {
			log.Printf("⚠️ ACE quality assessment failed for user %d: %v", req.UserID, err)
			continue
		}

		// Log quality assessment results
		if scores != nil {
			log.Printf("🧠 ACE Quality [User %d]: Composite=%.3f, Specificity=%.3f, Actionability=%.3f",
				req.UserID, scores.CompositeQualityScore, scores.SpecificityScore, scores.ActionabilityScore)
		}

		if decision != nil {
			log.Printf("🧠 ACE Decision [User %d]: Type=%s, Confidence=%.3f, Patterns=%d",
				req.UserID, decision.DecisionType, decision.ConfidenceLevel, len(decision.PatternsConsidered))
		}
	}
}

// GetSystemStatistics returns ACE system statistics
// GET /api/v1/ace/stats
func (h *ACEHandler) GetSystemStatistics(c *gin.Context) {
	stats, err := h.orchestrator.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to retrieve ACE statistics: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetRecentDecisions returns recent ACE decisions with quality scores
// GET /api/v1/ace/decisions?limit=20
func (h *ACEHandler) GetRecentDecisions(c *gin.Context) {
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100 // Cap at 100
			}
		}
	}

	var decisions []ace.Decision
	err := h.db.Order("created_at DESC").Limit(limit).Find(&decisions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to retrieve decisions: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"decisions": decisions,
			"count":     len(decisions),
			"limit":     limit,
		},
	})
}

// GetQualityScores returns quality score history
// GET /api/v1/ace/quality?limit=50
func (h *ACEHandler) GetQualityScores(c *gin.Context) {
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 200 {
				limit = 200
			}
		}
	}

	var scores []ace.QualityScores
	err := h.db.Order("created_at DESC").Limit(limit).Find(&scores).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to retrieve quality scores: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"scores": scores,
			"count":  len(scores),
			"limit":  limit,
		},
	})
}

// GetCognitivePatterns returns cognitive pattern library
// GET /api/v1/ace/patterns?category=learning&limit=20
func (h *ACEHandler) GetCognitivePatterns(c *gin.Context) {
	category := c.Query("category")
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100
			}
		}
	}

	query := h.db.Order("confidence_score DESC").Limit(limit)
	if category != "" {
		query = query.Where("pattern_category = ?", category)
	}

	var patterns []services.CognitivePattern
	err := query.Find(&patterns).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to retrieve patterns: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"patterns": patterns,
			"count":    len(patterns),
			"category": category,
			"limit":    limit,
		},
	})
}

// GetPlaybookRules returns curator playbook rules
// GET /api/v1/ace/playbook?limit=20
func (h *ACEHandler) GetPlaybookRules(c *gin.Context) {
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100
			}
		}
	}

	var rules []ace.PlaybookRule
	err := h.db.Order("confidence_score DESC").Limit(limit).Find(&rules).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to retrieve playbook rules: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"rules": rules,
			"count": len(rules),
			"limit": limit,
		},
	})
}

// TriggerPlaybookPruning manually triggers playbook pruning
// POST /api/v1/ace/prune
func (h *ACEHandler) TriggerPlaybookPruning(c *gin.Context) {
	pruned, err := h.orchestrator.PrunePlaybook()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Playbook pruning failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"pruned_rules": pruned,
			"message":      fmt.Sprintf("Successfully pruned %d low-confidence rules", pruned),
		},
	})
}
