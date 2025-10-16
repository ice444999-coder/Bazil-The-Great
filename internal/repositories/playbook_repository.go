package repositories

import (
	"ares_api/internal/models"
	
	"gorm.io/gorm"
)

// PlaybookRepository handles database operations for trading playbook
type PlaybookRepository struct {
	db *gorm.DB
}

// NewPlaybookRepository creates a new playbook repository
func NewPlaybookRepository(db *gorm.DB) *PlaybookRepository {
	return &PlaybookRepository{db: db}
}

// CreateRule adds a new rule to the playbook
func (r *PlaybookRepository) CreateRule(rule *models.PlaybookRule) error {
	return r.db.Create(rule).Error
}

// GetRuleByID retrieves a specific rule
func (r *PlaybookRepository) GetRuleByID(ruleID string, userID uint) (*models.PlaybookRule, error) {
	var rule models.PlaybookRule
	err := r.db.Where("rule_id = ? AND user_id = ?", ruleID, userID).First(&rule).Error
	return &rule, err
}

// GetActiveRules returns all active rules for a user
func (r *PlaybookRepository) GetActiveRules(userID uint) ([]models.PlaybookRule, error) {
	var rules []models.PlaybookRule
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("confidence DESC").
		Find(&rules).Error
	return rules, err
}

// GetReliableRules returns rules with proven track record
func (r *PlaybookRepository) GetReliableRules(userID uint, minConfidence float64) ([]models.PlaybookRule, error) {
	var rules []models.PlaybookRule
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Where("(helpful_count + harmful_count) >= 10").
		Where("confidence >= ?", minConfidence).
		Order("confidence DESC").
		Find(&rules).Error
	return rules, err
}

// UpdateRule saves changes to a rule
func (r *PlaybookRepository) UpdateRule(rule *models.PlaybookRule) error {
	return r.db.Save(rule).Error
}

// DeleteRule removes a rule from playbook
func (r *PlaybookRepository) DeleteRule(ruleID string, userID uint) error {
	return r.db.Where("rule_id = ? AND user_id = ?", ruleID, userID).
		Delete(&models.PlaybookRule{}).Error
}

// GetRulesByCategory filters rules by category
func (r *PlaybookRepository) GetRulesByCategory(userID uint, category string) ([]models.PlaybookRule, error) {
	var rules []models.PlaybookRule
	err := r.db.Where("user_id = ? AND is_active = ? AND category = ?", userID, true, category).
		Order("confidence DESC").
		Find(&rules).Error
	return rules, err
}
