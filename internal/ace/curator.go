package ace

import (
	"ares_api/internal/services"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// Curator manages the playbook of self-discovered rules
// It synthesizes patterns from experience and prunes ineffective ones
type Curator struct {
	db             *gorm.DB
	patternService *services.PatternService
}

// NewCurator creates a new Curator module
func NewCurator(db *gorm.DB, patternService *services.PatternService) *Curator {
	return &Curator{
		db:             db,
		patternService: patternService,
	}
}

// PlaybookRule represents a self-discovered knowledge rule
type PlaybookRule struct {
	ID                     uint                   `gorm:"primaryKey" json:"id"`
	RuleName               string                 `gorm:"size:200;not null" json:"rule_name"`
	RuleCategory           string                 `gorm:"size:100;not null" json:"rule_category"`
	RuleDescription        string                 `gorm:"type:text;not null" json:"rule_description"`
	Conditions             map[string]interface{} `gorm:"type:jsonb" json:"conditions"`
	Actions                map[string]interface{} `gorm:"type:jsonb" json:"actions"`
	Reasoning              string                 `gorm:"type:text" json:"reasoning"`
	TimesApplied           int                    `gorm:"default:0" json:"times_applied"`
	SuccessfulApplications int                    `gorm:"default:0" json:"successful_applications"`
	FailedApplications     int                    `gorm:"default:0" json:"failed_applications"`
	ConfidenceScore        float64                `gorm:"type:decimal(3,2);default:0.50" json:"confidence_score"`
	DiscoveredAt           time.Time              `gorm:"autoCreateTime" json:"discovered_at"`
	LastAppliedAt          *time.Time             `json:"last_applied_at,omitempty"`
	LastUpdatedAt          time.Time              `gorm:"autoUpdateTime" json:"last_updated_at"`
	BelowThresholdCount    int                    `gorm:"default:0" json:"below_threshold_count"`
	DerivedFromPatternID   *uint                  `json:"derived_from_pattern_id,omitempty"`
	ParentRuleID           *uint                  `json:"parent_rule_id,omitempty"`
}

// SynthesizePatternFromExperience creates a new playbook rule from successful outcomes
func (c *Curator) SynthesizePatternFromExperience(decision *Decision, scores *QualityScores, learning string) (*PlaybookRule, error) {
	log.Printf("🧪 Curator: Synthesizing new pattern from experience...")

	// Only synthesize if quality was high
	if scores.CompositeQualityScore < 0.7 {
		log.Printf("   ⏭️ Quality too low (%.2f < 0.7) - skipping synthesis", scores.CompositeQualityScore)
		return nil, nil
	}

	// Extract pattern name from decision context
	ruleName := fmt.Sprintf("Learned Rule: %s Success Pattern", decision.DecisionType)

	rule := &PlaybookRule{
		RuleName:        ruleName,
		RuleCategory:    decision.DecisionType,
		RuleDescription: fmt.Sprintf("Synthesized from successful decision with quality %.2f", scores.CompositeQualityScore),
		Conditions: map[string]interface{}{
			"decision_type":     decision.DecisionType,
			"min_quality_score": 0.7,
		},
		Actions: map[string]interface{}{
			"apply_patterns":    decision.PatternsConsidered,
			"use_tools":         decision.ToolsInvoked,
			"confidence_target": decision.ConfidenceLevel,
		},
		Reasoning:       learning,
		ConfidenceScore: 0.5, // New rules start at 0.5
		DiscoveredAt:    time.Now(),
	}

	// Link to patterns if applicable
	if len(decision.PatternsConsidered) > 0 {
		rule.DerivedFromPatternID = &decision.PatternsConsidered[0]
	}

	log.Printf("   ✅ Synthesized rule: %s (confidence: %.2f)", rule.RuleName, rule.ConfidenceScore)

	// TODO: Persist to database
	// c.db.Create(&rule)

	return rule, nil
}

// UpdateRuleConfidence adjusts rule confidence based on application outcomes
func (c *Curator) UpdateRuleConfidence(ruleID uint, successful bool) error {
	log.Printf("📊 Curator: Updating rule confidence for rule %d", ruleID)

	var rule PlaybookRule
	if err := c.db.First(&rule, ruleID).Error; err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	// Update application counters
	rule.TimesApplied++
	if successful {
		rule.SuccessfulApplications++
	} else {
		rule.FailedApplications++
	}

	// Calculate new confidence
	if rule.TimesApplied > 0 {
		successRate := float64(rule.SuccessfulApplications) / float64(rule.TimesApplied)

		// Exponential moving average: 80% old confidence + 20% new success rate
		rule.ConfidenceScore = 0.8*rule.ConfidenceScore + 0.2*successRate

		// Clamp to [0, 1]
		if rule.ConfidenceScore > 1.0 {
			rule.ConfidenceScore = 1.0
		} else if rule.ConfidenceScore < 0.0 {
			rule.ConfidenceScore = 0.0
		}
	}

	// Track if below pruning threshold
	const pruningThreshold = 0.3
	if rule.ConfidenceScore < pruningThreshold {
		rule.BelowThresholdCount++
		log.Printf("   ⚠️ Rule confidence low (%.2f < %.2f) - below threshold count: %d",
			rule.ConfidenceScore, pruningThreshold, rule.BelowThresholdCount)
	} else {
		rule.BelowThresholdCount = 0 // Reset counter if back above threshold
	}

	now := time.Now()
	rule.LastAppliedAt = &now

	log.Printf("   📈 Updated: Times=%d, Success=%d, Failed=%d, Confidence=%.2f",
		rule.TimesApplied, rule.SuccessfulApplications, rule.FailedApplications, rule.ConfidenceScore)

	// TODO: Save to database
	// c.db.Save(&rule)

	return nil
}

// PruneIneffectiveRules removes rules that consistently fail
func (c *Curator) PruneIneffectiveRules() (int, error) {
	log.Printf("✂️ Curator: Pruning ineffective rules...")

	// Find rules that have been below threshold for 5+ consecutive checks
	var rulesToPrune []PlaybookRule

	// TODO: Query database
	// c.db.Where("below_threshold_count >= ?", 5).Find(&rulesToPrune)

	prunedCount := 0
	for _, rule := range rulesToPrune {
		log.Printf("   🗑️ Pruning rule: %s (confidence: %.2f, failures: %d/%d)",
			rule.RuleName, rule.ConfidenceScore, rule.FailedApplications, rule.TimesApplied)

		// Archive instead of delete (for learning from failures)
		// TODO: Move to archived_playbook_rules table
		// c.db.Delete(&rule)

		prunedCount++
	}

	if prunedCount > 0 {
		log.Printf("   ✅ Pruned %d ineffective rules", prunedCount)
	} else {
		log.Printf("   ✅ No rules need pruning")
	}

	return prunedCount, nil
}

// EvolveRule creates a new rule by modifying an existing successful one
func (c *Curator) EvolveRule(parentRuleID uint, modifications map[string]interface{}) (*PlaybookRule, error) {
	log.Printf("🧬 Curator: Evolving rule %d...", parentRuleID)

	var parentRule PlaybookRule
	if err := c.db.First(&parentRule, parentRuleID).Error; err != nil {
		return nil, fmt.Errorf("parent rule not found: %w", err)
	}

	// Create evolved rule
	evolvedRule := PlaybookRule{
		RuleName:        fmt.Sprintf("%s (Evolved)", parentRule.RuleName),
		RuleCategory:    parentRule.RuleCategory,
		RuleDescription: fmt.Sprintf("Evolved from rule %d with modifications", parentRuleID),
		Conditions:      parentRule.Conditions,
		Actions:         parentRule.Actions,
		Reasoning:       parentRule.Reasoning + "\n\nEvolved with new learnings.",
		ConfidenceScore: parentRule.ConfidenceScore * 0.8, // Start at 80% of parent confidence
		ParentRuleID:    &parentRuleID,
		DiscoveredAt:    time.Now(),
	}

	// Apply modifications
	for key, value := range modifications {
		switch key {
		case "conditions":
			if v, ok := value.(map[string]interface{}); ok {
				evolvedRule.Conditions = v
			}
		case "actions":
			if v, ok := value.(map[string]interface{}); ok {
				evolvedRule.Actions = v
			}
		case "reasoning":
			if v, ok := value.(string); ok {
				evolvedRule.Reasoning += "\n" + v
			}
		}
	}

	log.Printf("   ✅ Evolved rule created: %s (confidence: %.2f)", evolvedRule.RuleName, evolvedRule.ConfidenceScore)

	// TODO: Persist to database
	// c.db.Create(&evolvedRule)

	return &evolvedRule, nil
}

// GetTopRules retrieves the highest confidence playbook rules
func (c *Curator) GetTopRules(category string, limit int) ([]PlaybookRule, error) {
	var rules []PlaybookRule

	query := c.db.Order("confidence_score DESC").Limit(limit)

	if category != "" {
		query = query.Where("rule_category = ?", category)
	}

	if err := query.Find(&rules).Error; err != nil {
		return nil, err
	}

	return rules, nil
}

// GetPlaybookStats returns statistics about the playbook
func (c *Curator) GetPlaybookStats() (map[string]interface{}, error) {
	var totalRules int64
	var avgConfidence float64
	var totalApplications int64
	var successfulApplications int64

	// Count total rules
	if err := c.db.Model(&PlaybookRule{}).Count(&totalRules).Error; err != nil {
		return nil, err
	}

	// Calculate average confidence
	if err := c.db.Model(&PlaybookRule{}).
		Select("AVG(confidence_score)").
		Scan(&avgConfidence).Error; err != nil {
		return nil, err
	}

	// Sum applications
	if err := c.db.Model(&PlaybookRule{}).
		Select("SUM(times_applied)").
		Scan(&totalApplications).Error; err != nil {
		return nil, err
	}

	// Sum successful applications
	if err := c.db.Model(&PlaybookRule{}).
		Select("SUM(successful_applications)").
		Scan(&successfulApplications).Error; err != nil {
		return nil, err
	}

	successRate := 0.0
	if totalApplications > 0 {
		successRate = float64(successfulApplications) / float64(totalApplications)
	}

	stats := map[string]interface{}{
		"total_rules":             totalRules,
		"average_confidence":      avgConfidence,
		"total_applications":      totalApplications,
		"successful_applications": successfulApplications,
		"success_rate":            successRate,
	}

	return stats, nil
}

// CombinePatterns creates a higher-order rule by combining multiple successful patterns
func (c *Curator) CombinePatterns(patternIDs []uint, context string) (*PlaybookRule, error) {
	log.Printf("🔗 Curator: Combining %d patterns into meta-pattern...", len(patternIDs))

	// Load the patterns
	var patterns []services.CognitivePattern
	for _, id := range patternIDs {
		var pattern services.CognitivePattern
		if err := c.db.First(&pattern, id).Error; err != nil {
			log.Printf("⚠️ Pattern %d not found: %v", id, err)
			continue
		}
		patterns = append(patterns, pattern)
	}

	if len(patterns) < 2 {
		return nil, fmt.Errorf("need at least 2 patterns to combine")
	}

	// Create combined rule
	combinedRule := PlaybookRule{
		RuleName:        fmt.Sprintf("Meta-Pattern: Combined %d patterns", len(patterns)),
		RuleCategory:    "meta-pattern",
		RuleDescription: fmt.Sprintf("Combination of patterns: %s", context),
		Conditions: map[string]interface{}{
			"pattern_ids": patternIDs,
			"context":     context,
		},
		Actions: map[string]interface{}{
			"apply_all_patterns": true,
			"pattern_sequence":   patternIDs,
		},
		Reasoning:       "Patterns work synergistically when applied together",
		ConfidenceScore: c.calculateCombinedConfidence(patterns),
		DiscoveredAt:    time.Now(),
	}

	log.Printf("   ✅ Meta-pattern created with confidence: %.2f", combinedRule.ConfidenceScore)

	return &combinedRule, nil
}

// calculateCombinedConfidence computes confidence for combined patterns
func (c *Curator) calculateCombinedConfidence(patterns []services.CognitivePattern) float64 {
	// Average confidence with a slight penalty for complexity
	total := 0.0
	for _, p := range patterns {
		total += p.ConfidenceScore
	}

	avg := total / float64(len(patterns))

	// Complexity penalty: -0.05 for each pattern beyond 3
	penalty := 0.0
	if len(patterns) > 3 {
		penalty = float64(len(patterns)-3) * 0.05
	}

	confidence := avg - penalty

	// Clamp to [0, 1]
	if confidence < 0.0 {
		confidence = 0.0
	} else if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}
