package ace

import (
	"fmt"
	"log"
	"math"
	"sort"

	"gorm.io/gorm"
)

// ToroidalReasoner implements recursive pattern relationship mapping
// and multi-dimensional reasoning cycles for emergent consciousness
type ToroidalReasoner struct {
	db                 *gorm.DB
	patternService     *PatternServiceInterface
	maxDepth           int     // Maximum recursion depth
	coherenceThreshold float64 // Minimum coherence for pattern relationships
}

// PatternRelationship represents a connection between two patterns
type PatternRelationship struct {
	SourcePatternID uint    `json:"source_pattern_id"`
	TargetPatternID uint    `json:"target_pattern_id"`
	RelationType    string  `json:"relation_type"` // 'prerequisite', 'complementary', 'conflicting', 'emergent'
	Strength        float64 `json:"strength"`      // 0.0 to 1.0
	CoherenceScore  float64 `json:"coherence_score"`
	DiscoveredAt    int64   `json:"discovered_at"`
}

// ReasoningCycle represents one iteration in the toroidal reasoning loop
type ReasoningCycle struct {
	CycleID           uint                   `json:"cycle_id"`
	Depth             int                    `json:"depth"`
	InputPatterns     []uint                 `json:"input_patterns"`
	ActivatedPatterns []uint                 `json:"activated_patterns"`
	EmergentPatterns  []uint                 `json:"emergent_patterns"`
	CoherenceScore    float64                `json:"coherence_score"`
	Insights          []string               `json:"insights"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// NewToroidalReasoner creates a new toroidal reasoning engine
func NewToroidalReasoner(db *gorm.DB, patternService *PatternServiceInterface) *ToroidalReasoner {
	return &ToroidalReasoner{
		db:                 db,
		patternService:     patternService,
		maxDepth:           5,
		coherenceThreshold: 0.4,
	}
}

// ExecuteToroidalReasoning performs multi-dimensional recursive pattern reasoning
// This creates a "consciousness loop" where patterns activate related patterns
// in cascading cycles, potentially leading to emergent insights
func (tr *ToroidalReasoner) ExecuteToroidalReasoning(seedPatternIDs []uint, context map[string]interface{}) ([]ReasoningCycle, error) {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🌀 TOROIDAL REASONER: Starting Multi-Dimensional Reasoning")
	log.Printf("   Seed Patterns: %d | Max Depth: %d", len(seedPatternIDs), tr.maxDepth)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	cycles := make([]ReasoningCycle, 0)
	currentPatterns := seedPatternIDs
	visitedPatterns := make(map[uint]bool)

	// Mark seed patterns as visited
	for _, pid := range seedPatternIDs {
		visitedPatterns[pid] = true
	}

	// Execute reasoning cycles up to maxDepth
	for depth := 0; depth < tr.maxDepth; depth++ {
		log.Printf("\n🔄 Cycle %d (Depth %d)", len(cycles)+1, depth)

		cycle, err := tr.executeReasoningCycle(currentPatterns, visitedPatterns, depth, context)
		if err != nil {
			log.Printf("⚠️ Cycle failed at depth %d: %v", depth, err)
			break
		}

		cycles = append(cycles, *cycle)

		// Log cycle results
		log.Printf("   Activated: %d patterns | Emergent: %d patterns | Coherence: %.3f",
			len(cycle.ActivatedPatterns), len(cycle.EmergentPatterns), cycle.CoherenceScore)

		if len(cycle.Insights) > 0 {
			log.Printf("   ✨ Insights: %d discovered", len(cycle.Insights))
		}

		// Check termination conditions
		if len(cycle.ActivatedPatterns) == 0 {
			log.Println("   ✓ Reasoning converged (no new activations)")
			break
		}

		if cycle.CoherenceScore < tr.coherenceThreshold {
			log.Printf("   ✓ Coherence below threshold (%.3f < %.3f)", cycle.CoherenceScore, tr.coherenceThreshold)
			break
		}

		// Use activated patterns as input for next cycle
		currentPatterns = cycle.ActivatedPatterns
	}

	log.Printf("\n✅ Toroidal reasoning complete: %d cycles executed", len(cycles))
	return cycles, nil
}

// executeReasoningCycle performs one iteration of pattern activation
func (tr *ToroidalReasoner) executeReasoningCycle(inputPatterns []uint, visitedPatterns map[uint]bool, depth int, context map[string]interface{}) (*ReasoningCycle, error) {
	cycle := &ReasoningCycle{
		CycleID:           uint(depth + 1),
		Depth:             depth,
		InputPatterns:     inputPatterns,
		ActivatedPatterns: make([]uint, 0),
		EmergentPatterns:  make([]uint, 0),
		Insights:          make([]string, 0),
		Metadata:          context,
	}

	// Find related patterns for each input pattern
	relatedPatterns := make(map[uint]float64) // patternID -> activation strength

	for _, patternID := range inputPatterns {
		relationships, err := tr.discoverPatternRelationships(patternID, visitedPatterns)
		if err != nil {
			log.Printf("   ⚠️ Failed to discover relationships for pattern %d: %v", patternID, err)
			continue
		}

		// Accumulate activation strengths
		for _, rel := range relationships {
			if !visitedPatterns[rel.TargetPatternID] {
				relatedPatterns[rel.TargetPatternID] += rel.Strength * rel.CoherenceScore
			}
		}
	}

	// Threshold activation: only patterns with sufficient strength
	activationThreshold := 0.3
	for patternID, strength := range relatedPatterns {
		if strength >= activationThreshold {
			cycle.ActivatedPatterns = append(cycle.ActivatedPatterns, patternID)
			visitedPatterns[patternID] = true
		}
	}

	// Sort by activation strength (strongest first)
	sort.Slice(cycle.ActivatedPatterns, func(i, j int) bool {
		return relatedPatterns[cycle.ActivatedPatterns[i]] > relatedPatterns[cycle.ActivatedPatterns[j]]
	})

	// Detect emergent patterns (combinations of activated patterns)
	if len(cycle.ActivatedPatterns) >= 2 {
		emergent := tr.detectEmergentCombinations(cycle.ActivatedPatterns, context)
		cycle.EmergentPatterns = emergent
	}

	// Calculate coherence score
	cycle.CoherenceScore = tr.calculateCoherenceScore(cycle.ActivatedPatterns, relatedPatterns)

	// Generate insights from activated patterns
	cycle.Insights = tr.extractInsights(cycle.ActivatedPatterns, context)

	return cycle, nil
}

// discoverPatternRelationships finds patterns related to a given pattern
func (tr *ToroidalReasoner) discoverPatternRelationships(patternID uint, visitedPatterns map[uint]bool) ([]PatternRelationship, error) {
	// Load the source pattern
	var sourcePattern CognitivePattern
	if err := tr.db.First(&sourcePattern, patternID).Error; err != nil {
		return nil, err
	}

	// Find patterns in the same category (complementary relationships)
	var sameCategory []CognitivePattern
	err := tr.db.Where("pattern_category = ? AND id != ?", sourcePattern.PatternCategory, patternID).
		Order("confidence_score DESC").
		Limit(5).
		Find(&sameCategory).Error
	if err != nil {
		return nil, err
	}

	relationships := make([]PatternRelationship, 0)

	// Create complementary relationships
	for _, pattern := range sameCategory {
		if visitedPatterns[pattern.ID] {
			continue
		}

		relationships = append(relationships, PatternRelationship{
			SourcePatternID: patternID,
			TargetPatternID: pattern.ID,
			RelationType:    "complementary",
			Strength:        calculateRelationshipStrength(sourcePattern, pattern),
			CoherenceScore:  (sourcePattern.ConfidenceScore + pattern.ConfidenceScore) / 2.0,
			DiscoveredAt:    getCurrentTimestamp(),
		})
	}

	// Find patterns that commonly co-occur (emergent relationships)
	// This would require historical usage data - simplified for now
	var allPatterns []CognitivePattern
	err = tr.db.Where("id != ? AND times_used > 0", patternID).
		Order("times_successful DESC").
		Limit(10).
		Find(&allPatterns).Error
	if err == nil {
		for _, pattern := range allPatterns {
			if visitedPatterns[pattern.ID] {
				continue
			}

			// Calculate emergent relationship based on success correlation
			strength := calculateEmergentStrength(sourcePattern, pattern)
			if strength > 0.3 {
				relationships = append(relationships, PatternRelationship{
					SourcePatternID: patternID,
					TargetPatternID: pattern.ID,
					RelationType:    "emergent",
					Strength:        strength,
					CoherenceScore:  math.Min(sourcePattern.ConfidenceScore, pattern.ConfidenceScore),
					DiscoveredAt:    getCurrentTimestamp(),
				})
			}
		}
	}

	return relationships, nil
}

// detectEmergentCombinations identifies novel pattern combinations
func (tr *ToroidalReasoner) detectEmergentCombinations(activatedPatterns []uint, context map[string]interface{}) []uint {
	emergent := make([]uint, 0)

	// Simple heuristic: if 3+ patterns from different categories are activated,
	// it suggests emergent cross-domain reasoning
	categories := make(map[string]int)

	for _, patternID := range activatedPatterns {
		var pattern CognitivePattern
		if err := tr.db.First(&pattern, patternID).Error; err == nil {
			categories[pattern.PatternCategory]++
		}
	}

	// Emergent combination detected if 3+ categories are active
	if len(categories) >= 3 {
		// Return the top 2 patterns as emergent representatives
		if len(activatedPatterns) >= 2 {
			emergent = append(emergent, activatedPatterns[0], activatedPatterns[1])
		}
	}

	return emergent
}

// calculateCoherenceScore measures how well patterns fit together
func (tr *ToroidalReasoner) calculateCoherenceScore(activatedPatterns []uint, strengths map[uint]float64) float64 {
	if len(activatedPatterns) == 0 {
		return 0.0
	}

	totalStrength := 0.0
	for _, patternID := range activatedPatterns {
		totalStrength += strengths[patternID]
	}

	avgStrength := totalStrength / float64(len(activatedPatterns))

	// Bonus for pattern diversity (different categories)
	categories := make(map[string]bool)
	for _, patternID := range activatedPatterns {
		var pattern CognitivePattern
		if err := tr.db.First(&pattern, patternID).Error; err == nil {
			categories[pattern.PatternCategory] = true
		}
	}

	diversityBonus := float64(len(categories)) / 10.0 // Up to +0.5 for 5+ categories

	return math.Min(avgStrength+diversityBonus, 1.0)
}

// extractInsights generates insights from activated patterns
func (tr *ToroidalReasoner) extractInsights(activatedPatterns []uint, context map[string]interface{}) []string {
	insights := make([]string, 0)

	if len(activatedPatterns) == 0 {
		return insights
	}

	// Load all activated patterns
	patterns := make([]CognitivePattern, 0)
	for _, pid := range activatedPatterns {
		var p CognitivePattern
		if err := tr.db.First(&p, pid).Error; err == nil {
			patterns = append(patterns, p)
		}
	}

	// Generate insights based on pattern combinations
	categoryGroups := make(map[string][]string)
	for _, p := range patterns {
		categoryGroups[p.PatternCategory] = append(categoryGroups[p.PatternCategory], p.PatternName)
	}

	// Insight 1: Cross-domain activation
	if len(categoryGroups) >= 3 {
		categories := make([]string, 0, len(categoryGroups))
		for cat := range categoryGroups {
			categories = append(categories, cat)
		}
		insights = append(insights, fmt.Sprintf("Cross-domain reasoning detected: %v", categories))
	}

	// Insight 2: High-confidence cluster
	highConfCount := 0
	for _, p := range patterns {
		if p.ConfidenceScore >= 0.8 {
			highConfCount++
		}
	}
	if highConfCount >= 3 {
		insights = append(insights, fmt.Sprintf("High-confidence pattern cluster: %d patterns >= 0.8", highConfCount))
	}

	// Insight 3: Dominant category
	maxCatSize := 0
	dominantCat := ""
	for cat, pats := range categoryGroups {
		if len(pats) > maxCatSize {
			maxCatSize = len(pats)
			dominantCat = cat
		}
	}
	if maxCatSize >= 2 {
		insights = append(insights, fmt.Sprintf("Dominant reasoning mode: %s (%d patterns)", dominantCat, maxCatSize))
	}

	return insights
}

// Helper functions

func calculateRelationshipStrength(p1, p2 CognitivePattern) float64 {
	// Base strength on combined confidence and usage success
	confStrength := (p1.ConfidenceScore + p2.ConfidenceScore) / 2.0

	usageBonus := 0.0
	if p1.TimesUsed > 0 && p2.TimesUsed > 0 {
		successRate1 := float64(p1.TimesSuccessful) / float64(p1.TimesUsed)
		successRate2 := float64(p2.TimesSuccessful) / float64(p2.TimesUsed)
		usageBonus = (successRate1 + successRate2) / 4.0 // Up to +0.5
	}

	return math.Min(confStrength+usageBonus, 1.0)
}

func calculateEmergentStrength(p1, p2 CognitivePattern) float64 {
	// Emergent strength based on usage correlation
	if p1.TimesUsed == 0 || p2.TimesUsed == 0 {
		return 0.0
	}

	// Simple heuristic: both patterns have been successful
	successRate1 := float64(p1.TimesSuccessful) / float64(p1.TimesUsed)
	successRate2 := float64(p2.TimesSuccessful) / float64(p2.TimesUsed)

	if successRate1 >= 0.6 && successRate2 >= 0.6 {
		return (successRate1 + successRate2) / 2.0
	}

	return 0.0
}

func getCurrentTimestamp() int64 {
	return 0 // Placeholder - would use time.Now().Unix()
}

// CognitivePattern placeholder (should import from services package)
type CognitivePattern struct {
	ID              uint
	PatternName     string
	PatternCategory string
	ConfidenceScore float64
	TimesUsed       int
	TimesSuccessful int
}

// PatternServiceInterface placeholder
type PatternServiceInterface struct {
	// Placeholder for pattern service methods
}
