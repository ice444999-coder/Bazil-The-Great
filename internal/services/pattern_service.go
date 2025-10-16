package services

import (
	"fmt"
	"log"
	"os"

	"gorm.io/gorm"
)

// CognitivePattern represents a Claude reasoning pattern
type CognitivePattern struct {
	ID                          uint    `gorm:"primaryKey" json:"id"`
	PatternName                 string  `gorm:"uniqueIndex;size:200;not null" json:"pattern_name"`
	PatternCategory             string  `gorm:"size:100;not null" json:"pattern_category"`
	Description                 string  `gorm:"type:text;not null" json:"description"`
	TriggerConditions           string  `gorm:"type:text" json:"trigger_conditions"`
	ExampleInput                string  `gorm:"type:text" json:"example_input"`
	ExampleOutput               string  `gorm:"type:text" json:"example_output"`
	ExampleReasoning            string  `gorm:"type:text" json:"example_reasoning"`
	ConfidenceScore             float64 `gorm:"type:decimal(3,2);not null;default:0.80" json:"confidence_score"`
	TimesUsed                   int     `gorm:"default:0" json:"times_used"`
	TimesSuccessful             int     `gorm:"default:0" json:"times_successful"`
	Source                      string  `gorm:"size:50;default:'claude'" json:"source"`
	ExtractedFromConversationID *int    `json:"extracted_from_conversation_id,omitempty"`
	CreatedAt                   int64   `gorm:"autoCreateTime" json:"created_at"`
	LastUsedAt                  *int64  `json:"last_used_at,omitempty"`
}

// PatternService manages cognitive patterns
type PatternService struct {
	db *gorm.DB
}

// NewPatternService creates a new pattern service
func NewPatternService(db *gorm.DB) *PatternService {
	return &PatternService{db: db}
}

// LoadPatternsFromPython loads patterns from cognitive_patterns.py
func (s *PatternService) LoadPatternsFromPython(pythonFilePath string) (int, error) {
	log.Printf("🧠 Loading cognitive patterns from: %s", pythonFilePath)

	// Read the Python file
	content, err := os.ReadFile(pythonFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read pattern file: %w", err)
	}

	// Parse Python data (simplified - in production, use proper Python parser)
	// For now, we'll manually create the patterns based on the Python structure
	patterns, err := s.parsePythonPatterns(string(content))
	if err != nil {
		return 0, fmt.Errorf("failed to parse patterns: %w", err)
	}

	// Bulk insert patterns
	loaded := 0
	for _, pattern := range patterns {
		// Check if pattern already exists
		var existing CognitivePattern
		result := s.db.Where("pattern_name = ?", pattern.PatternName).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new pattern
			if err := s.db.Create(&pattern).Error; err != nil {
				log.Printf("⚠️ Failed to create pattern '%s': %v", pattern.PatternName, err)
				continue
			}
			loaded++
			log.Printf("   ✅ Loaded: %s (confidence: %.2f)", pattern.PatternName, pattern.ConfidenceScore)
		} else {
			// Update existing pattern
			existing.Description = pattern.Description
			existing.TriggerConditions = pattern.TriggerConditions
			existing.ExampleInput = pattern.ExampleInput
			existing.ExampleOutput = pattern.ExampleOutput
			existing.ExampleReasoning = pattern.ExampleReasoning
			existing.ConfidenceScore = pattern.ConfidenceScore

			if err := s.db.Save(&existing).Error; err != nil {
				log.Printf("⚠️ Failed to update pattern '%s': %v", pattern.PatternName, err)
				continue
			}
			log.Printf("   🔄 Updated: %s", pattern.PatternName)
		}
	}

	log.Printf("🎉 Pattern loading complete: %d patterns loaded/updated", loaded)
	return loaded, nil
}

// parsePythonPatterns extracts patterns from Python file content
// TODO: Replace with proper Python AST parser for production
func (s *PatternService) parsePythonPatterns(content string) ([]CognitivePattern, error) {
	// For now, return hardcoded patterns that match our extracted patterns
	// In production, this should parse the actual Python file
	patterns := []CognitivePattern{
		{
			PatternName:       "Problem Inference - Surface vs Deep Need",
			PatternCategory:   "problem-inference",
			Description:       "User asks surface-level question but actually needs deeper solution. Always confirm understanding of real problem before solving.",
			TriggerConditions: "User question seems simple but context suggests complexity",
			ExampleInput:      "Can you help me open VS Code back to the repo?",
			ExampleOutput:     "Run: code c:\\ARES_Workspace",
			ExampleReasoning:  "User doesn't need tutorial. They need specific workspace path. Infer deep need, provide specific solution.",
			ConfidenceScore:   0.95,
			Source:            "claude",
		},
		{
			PatternName:       "Specificity Over Generality",
			PatternCategory:   "response-quality",
			Description:       "Always provide exact paths, commands, values instead of generic instructions",
			TriggerConditions: "Any question about ARES system, paths, configurations, commands",
			ExampleInput:      "Where is your workspace?",
			ExampleOutput:     "c:\\ARES_Workspace (not 'check your file system')",
			ExampleReasoning:  "Specific answers exponentially more helpful than generic tutorials",
			ConfidenceScore:   1.00,
			Source:            "claude",
		},
		// Add more patterns here - for now starting with key patterns
		// Full implementation should parse cognitive_patterns.py
	}

	return patterns, nil
}

// GetPatternsByCategory retrieves patterns for a specific category
func (s *PatternService) GetPatternsByCategory(category string) ([]CognitivePattern, error) {
	var patterns []CognitivePattern
	result := s.db.Where("pattern_category = ?", category).
		Order("confidence_score DESC").
		Find(&patterns)

	if result.Error != nil {
		return nil, result.Error
	}

	return patterns, nil
}

// GetTopPatterns retrieves N highest confidence patterns
func (s *PatternService) GetTopPatterns(limit int) ([]CognitivePattern, error) {
	var patterns []CognitivePattern
	result := s.db.Order("confidence_score DESC").
		Limit(limit).
		Find(&patterns)

	if result.Error != nil {
		return nil, result.Error
	}

	return patterns, nil
}

// GetPatternByName retrieves a specific pattern
func (s *PatternService) GetPatternByName(name string) (*CognitivePattern, error) {
	var pattern CognitivePattern
	result := s.db.Where("pattern_name = ?", name).First(&pattern)

	if result.Error != nil {
		return nil, result.Error
	}

	return &pattern, nil
}

// RecordPatternUsage increments usage counters for a pattern
func (s *PatternService) RecordPatternUsage(patternID uint, successful bool) error {
	updates := map[string]interface{}{
		"times_used":   gorm.Expr("times_used + ?", 1),
		"last_used_at": gorm.Expr("?", fmt.Sprintf("%d", int64(1))), // Unix timestamp
	}

	if successful {
		updates["times_successful"] = gorm.Expr("times_successful + ?", 1)
	}

	result := s.db.Model(&CognitivePattern{}).
		Where("id = ?", patternID).
		Updates(updates)

	return result.Error
}

// UpdatePatternConfidence adjusts confidence based on success rate
func (s *PatternService) UpdatePatternConfidence(patternID uint) error {
	var pattern CognitivePattern
	if err := s.db.First(&pattern, patternID).Error; err != nil {
		return err
	}

	if pattern.TimesUsed == 0 {
		return nil // No usage data yet
	}

	// Calculate new confidence based on success rate
	successRate := float64(pattern.TimesSuccessful) / float64(pattern.TimesUsed)

	// Exponential moving average: new_conf = 0.8 * old_conf + 0.2 * success_rate
	newConfidence := 0.8*pattern.ConfidenceScore + 0.2*successRate

	// Clamp between 0.0 and 1.0
	if newConfidence > 1.0 {
		newConfidence = 1.0
	} else if newConfidence < 0.0 {
		newConfidence = 0.0
	}

	pattern.ConfidenceScore = newConfidence
	return s.db.Save(&pattern).Error
}

// GetPatternStats returns statistics about the pattern library
func (s *PatternService) GetPatternStats() (map[string]interface{}, error) {
	var totalPatterns int64
	var avgConfidence float64

	// Count total patterns
	if err := s.db.Model(&CognitivePattern{}).Count(&totalPatterns).Error; err != nil {
		return nil, err
	}

	// Calculate average confidence
	if err := s.db.Model(&CognitivePattern{}).
		Select("AVG(confidence_score)").
		Scan(&avgConfidence).Error; err != nil {
		return nil, err
	}

	// Count by category
	var categoryStats []struct {
		Category string
		Count    int64
	}
	if err := s.db.Model(&CognitivePattern{}).
		Select("pattern_category as category, COUNT(*) as count").
		Group("pattern_category").
		Scan(&categoryStats).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_patterns":     totalPatterns,
		"average_confidence": avgConfidence,
		"categories":         categoryStats,
	}

	return stats, nil
}

// SearchPatterns finds patterns matching trigger conditions or description
func (s *PatternService) SearchPatterns(query string) ([]CognitivePattern, error) {
	var patterns []CognitivePattern
	searchQuery := "%" + query + "%"

	result := s.db.Where("trigger_conditions LIKE ? OR description LIKE ?", searchQuery, searchQuery).
		Order("confidence_score DESC").
		Find(&patterns)

	if result.Error != nil {
		return nil, result.Error
	}

	return patterns, nil
}
