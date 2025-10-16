package trading

import (
	"ares_api/internal/models"
	"fmt"
	"time"
	
	"gorm.io/gorm"
)

// Curator manages the trading playbook database
// Updates rules based on trade outcomes
type Curator struct {
	db *gorm.DB
}

// NewCurator creates a new playbook manager
func NewCurator(db *gorm.DB) *Curator {
	return &Curator{db: db}
}

// ApplyDeltaUpdates processes learning insights and updates the playbook
func (c *Curator) ApplyDeltaUpdates(deltas []DeltaUpdate, userID uint) error {
	for _, delta := range deltas {
		// Check if rule already exists
		var existingRule models.PlaybookRule
		result := c.db.Where("rule_id = ? AND user_id = ?", delta.RuleID, userID).First(&existingRule)
		
		if result.Error == gorm.ErrRecordNotFound {
			// Create new rule
			newRule := models.PlaybookRule{
				RuleID:       delta.RuleID,
				Content:      delta.Content,
				Category:     delta.Category,
				UserID:       userID,
				IsActive:     true,
				Conditions:   delta.Conditions,
			}
			
			// Set initial counts
			if delta.IsHelpful {
				newRule.HelpfulCount = 1
				newRule.HarmfulCount = 0
			} else {
				newRule.HelpfulCount = 0
				newRule.HarmfulCount = 1
			}
			
			newRule.CalculateConfidence()
			
			if err := c.db.Create(&newRule).Error; err != nil {
				return fmt.Errorf("failed to create rule %s: %w", delta.RuleID, err)
			}
			
		} else if result.Error != nil {
			return fmt.Errorf("database error: %w", result.Error)
		} else {
			// Update existing rule
			if err := c.UpdateRuleCounters(delta.RuleID, userID, delta.IsHelpful); err != nil {
				return err
			}
		}
	}
	
	// Prune weak rules after updates
	if err := c.PruneWeakRules(userID, 20, 0.30); err != nil {
		return err
	}
	
	return nil
}

// UpdateRuleCounters increments helpful or harmful count
func (c *Curator) UpdateRuleCounters(ruleID string, userID uint, wasHelpful bool) error {
	var rule models.PlaybookRule
	if err := c.db.Where("rule_id = ? AND user_id = ?", ruleID, userID).First(&rule).Error; err != nil {
		return err
	}
	
	// Increment appropriate counter
	if wasHelpful {
		rule.HelpfulCount++
	} else {
		rule.HarmfulCount++
	}
	
	// Recalculate confidence
	rule.CalculateConfidence()
	rule.TotalUses++
	now := time.Now()
	rule.LastUsedAt = &now
	
	// Check if rule should be deactivated
	if rule.ShouldPrune() {
		rule.IsActive = false
	}
	
	return c.db.Save(&rule).Error
}

// PruneWeakRules removes or deactivates underperforming rules
func (c *Curator) PruneWeakRules(userID uint, minObservations int, minConfidence float64) error {
	// Deactivate rules that have enough data but low confidence
	return c.db.Model(&models.PlaybookRule{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Where("(helpful_count + harmful_count) >= ?", minObservations).
		Where("confidence < ?", minConfidence).
		Update("is_active", false).Error
}

// GetActiveRules retrieves all active rules for a user
func (c *Curator) GetActiveRules(userID uint) ([]models.PlaybookRule, error) {
	var rules []models.PlaybookRule
	err := c.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("confidence DESC").
		Find(&rules).Error
	return rules, err
}

// GetRulesByCategory retrieves rules filtered by category
func (c *Curator) GetRulesByCategory(userID uint, category string) ([]models.PlaybookRule, error) {
	var rules []models.PlaybookRule
	err := c.db.Where("user_id = ? AND is_active = ? AND category = ?", userID, true, category).
		Order("confidence DESC").
		Find(&rules).Error
	return rules, err
}

// GetReliableRules returns rules that have proven track record
func (c *Curator) GetReliableRules(userID uint) ([]models.PlaybookRule, error) {
	var rules []models.PlaybookRule
	err := c.db.Where("user_id = ? AND is_active = ?", userID, true).
		Where("(helpful_count + harmful_count) >= 10").
		Where("confidence >= 0.60").
		Order("confidence DESC").
		Find(&rules).Error
	return rules, err
}

// GetPlaybookStats returns statistics about the playbook
func (c *Curator) GetPlaybookStats(userID uint) (map[string]interface{}, error) {
	var totalRules int64
	var activeRules int64
	var avgConfidence float64
	
	c.db.Model(&models.PlaybookRule{}).Where("user_id = ?", userID).Count(&totalRules)
	c.db.Model(&models.PlaybookRule{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&activeRules)
	
	var result struct {
		AvgConf float64
	}
	c.db.Model(&models.PlaybookRule{}).
		Select("AVG(confidence) as avg_conf").
		Where("user_id = ? AND is_active = ?", userID, true).
		Scan(&result)
	avgConfidence = result.AvgConf
	
	return map[string]interface{}{
		"total_rules":      totalRules,
		"active_rules":     activeRules,
		"inactive_rules":   totalRules - activeRules,
		"avg_confidence":   avgConfidence,
	}, nil
}

// RecordRuleUsage tracks when a rule was consulted
func (c *Curator) RecordRuleUsage(ruleID string, userID uint) error {
	now := time.Now()
	return c.db.Model(&models.PlaybookRule{}).
		Where("rule_id = ? AND user_id = ?", ruleID, userID).
		Updates(map[string]interface{}{
			"total_uses":   gorm.Expr("total_uses + 1"),
			"last_used_at": now,
		}).Error
}

// DeduplicateRules merges similar rules (placeholder for future ML-based deduplication)
func (c *Curator) DeduplicateRules(userID uint) error {
	// TODO: Implement semantic similarity check
	// For now, this is a placeholder
	// Future: Use embeddings to find similar rule content and merge them
	return nil
}
