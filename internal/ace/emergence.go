package ace

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"gorm.io/gorm"
)

// EmergenceDetector identifies emergent behaviors, novel pattern combinations,
// and spontaneous reasoning pathways that arise from ACE Framework operation
type EmergenceDetector struct {
	db               *gorm.DB
	noveltyThreshold float64            // Threshold for considering something "novel"
	emergenceWindow  time.Duration      // Time window for detecting emergence
	patternSynergies map[string]float64 // Tracked pattern synergies
}

// EmergentBehavior represents a detected emergent phenomenon
type EmergentBehavior struct {
	ID                uint                   `gorm:"primaryKey" json:"id"`
	BehaviorType      string                 `gorm:"size:100;not null" json:"behavior_type"` // 'novel-combination', 'synergy', 'breakthrough', 'contradiction'
	Description       string                 `gorm:"type:text" json:"description"`
	InvolvedPatterns  []uint                 `gorm:"type:jsonb" json:"involved_patterns"`
	NoveltyScore      float64                `gorm:"type:decimal(4,3)" json:"novelty_score"`
	ImpactScore       float64                `gorm:"type:decimal(4,3)" json:"impact_score"`
	Confidence        float64                `gorm:"type:decimal(4,3)" json:"confidence"`
	Evidence          map[string]interface{} `gorm:"type:jsonb" json:"evidence"`
	DetectedAt        int64                  `gorm:"autoCreateTime" json:"detected_at"`
	VerifiedAt        *int64                 `json:"verified_at,omitempty"`
	IntegrationStatus string                 `gorm:"size:50;default:'pending'" json:"integration_status"` // 'pending', 'integrated', 'rejected'
}

// PatternSynergy represents synergistic interaction between patterns
type PatternSynergy struct {
	Pattern1ID       uint    `json:"pattern1_id"`
	Pattern2ID       uint    `json:"pattern2_id"`
	SynergyScore     float64 `json:"synergy_score"`  // How well they work together (>1.0 = synergistic)
	CoOccurrences    int     `json:"co_occurrences"` // Times used together
	JointSuccessRate float64 `json:"joint_success_rate"`
	DiscoveredAt     int64   `json:"discovered_at"`
}

// ReasoningBreakthrough represents a significant cognitive leap
type ReasoningBreakthrough struct {
	ID                 uint                   `gorm:"primaryKey" json:"id"`
	BreakthroughType   string                 `gorm:"size:100" json:"breakthrough_type"` // 'insight', 'synthesis', 'meta-pattern'
	Description        string                 `gorm:"type:text" json:"description"`
	QualityImprovement float64                `gorm:"type:decimal(4,3)" json:"quality_improvement"`
	PatternChain       []uint                 `gorm:"type:jsonb" json:"pattern_chain"` // Sequence of patterns leading to breakthrough
	Context            map[string]interface{} `gorm:"type:jsonb" json:"context"`
	CreatedAt          int64                  `gorm:"autoCreateTime" json:"created_at"`
	Reproducible       bool                   `gorm:"default:false" json:"reproducible"`
}

// NewEmergenceDetector creates a new emergence detection system
func NewEmergenceDetector(db *gorm.DB) *EmergenceDetector {
	return &EmergenceDetector{
		db:               db,
		noveltyThreshold: 0.7, // 70% novelty required
		emergenceWindow:  24 * time.Hour,
		patternSynergies: make(map[string]float64),
	}
}

// DetectEmergence analyzes recent ACE activity for emergent phenomena
func (ed *EmergenceDetector) DetectEmergence(decisions []Decision, cycles []ReasoningCycle) ([]EmergentBehavior, error) {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✨ EMERGENCE DETECTOR: Analyzing for Emergent Phenomena")
	log.Printf("   Decisions: %d | Reasoning Cycles: %d", len(decisions), len(cycles))
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	emergentBehaviors := make([]EmergentBehavior, 0)

	// Detection 1: Novel Pattern Combinations
	novelCombos := ed.detectNovelCombinations(decisions)
	emergentBehaviors = append(emergentBehaviors, novelCombos...)

	// Detection 2: Pattern Synergies
	synergies := ed.detectPatternSynergies(decisions)
	emergentBehaviors = append(emergentBehaviors, synergies...)

	// Detection 3: Reasoning Breakthroughs
	breakthroughs := ed.detectBreakthroughs(decisions, cycles)
	emergentBehaviors = append(emergentBehaviors, breakthroughs...)

	// Detection 4: Contradictions and Conflicts
	contradictions := ed.detectContradictions(decisions)
	emergentBehaviors = append(emergentBehaviors, contradictions...)

	// Detection 5: Meta-Pattern Formation
	metaPatterns := ed.detectMetaPatterns(cycles)
	emergentBehaviors = append(emergentBehaviors, metaPatterns...)

	log.Printf("✅ Emergence detection complete: %d phenomena detected", len(emergentBehaviors))
	return emergentBehaviors, nil
}

// detectNovelCombinations finds never-before-seen pattern combinations
func (ed *EmergenceDetector) detectNovelCombinations(decisions []Decision) []EmergentBehavior {
	log.Println("\n🔍 Detecting novel pattern combinations...")

	behaviors := make([]EmergentBehavior, 0)
	seenCombos := make(map[string]bool)

	// Load historical pattern combinations
	// (Simplified - would query historical decisions in production)

	for _, decision := range decisions {
		if len(decision.PatternsConsidered) < 2 {
			continue
		}

		// Sort pattern IDs for consistent combo keys
		patternIDs := make([]uint, len(decision.PatternsConsidered))
		copy(patternIDs, decision.PatternsConsidered)
		sort.Slice(patternIDs, func(i, j int) bool {
			return patternIDs[i] < patternIDs[j]
		})

		comboKey := fmt.Sprintf("%v", patternIDs)

		// Check if this combination is novel
		if !seenCombos[comboKey] {
			seenCombos[comboKey] = true

			// Calculate novelty score (simplified - based on quality improvement)
			noveltyScore := 0.5
			if decision.FinalQualityScore != nil && decision.InitialQualityScore != nil {
				improvement := *decision.FinalQualityScore - *decision.InitialQualityScore
				if improvement > 0.2 {
					noveltyScore = 0.8
				}
			}

			if noveltyScore >= ed.noveltyThreshold {
				behaviors = append(behaviors, EmergentBehavior{
					BehaviorType:     "novel-combination",
					Description:      fmt.Sprintf("Novel pattern combination: %d patterns from %s context", len(patternIDs), decision.DecisionType),
					InvolvedPatterns: patternIDs,
					NoveltyScore:     noveltyScore,
					ImpactScore:      decision.ConfidenceLevel,
					Confidence:       0.8,
					Evidence: map[string]interface{}{
						"decision_id":   decision.DecisionID,
						"pattern_count": len(patternIDs),
						"decision_type": decision.DecisionType,
					},
					IntegrationStatus: "pending",
				})

				log.Printf("   ✨ Novel combination detected: %d patterns (novelty=%.3f)", len(patternIDs), noveltyScore)
			}
		}
	}

	return behaviors
}

// detectPatternSynergies identifies patterns that work exceptionally well together
func (ed *EmergenceDetector) detectPatternSynergies(decisions []Decision) []EmergentBehavior {
	log.Println("\n🔍 Detecting pattern synergies...")

	behaviors := make([]EmergentBehavior, 0)
	pairPerformance := make(map[string]*synergyStats)

	// Track pattern pair performance
	for _, decision := range decisions {
		if len(decision.PatternsConsidered) < 2 {
			continue
		}

		quality := 0.5
		if decision.FinalQualityScore != nil {
			quality = *decision.FinalQualityScore
		} else if decision.InitialQualityScore != nil {
			quality = *decision.InitialQualityScore
		}

		// Analyze all pairs
		for i := 0; i < len(decision.PatternsConsidered)-1; i++ {
			for j := i + 1; j < len(decision.PatternsConsidered); j++ {
				p1 := decision.PatternsConsidered[i]
				p2 := decision.PatternsConsidered[j]

				pairKey := fmt.Sprintf("%d-%d", min(p1, p2), max(p1, p2))

				if _, exists := pairPerformance[pairKey]; !exists {
					pairPerformance[pairKey] = &synergyStats{
						pattern1: p1,
						pattern2: p2,
					}
				}

				stats := pairPerformance[pairKey]
				stats.count++
				stats.totalQuality += quality
				if quality >= 0.7 {
					stats.successes++
				}
			}
		}
	}

	// Identify synergistic pairs
	for _, stats := range pairPerformance {
		if stats.count < 2 {
			continue // Need at least 2 co-occurrences
		}

		avgQuality := stats.totalQuality / float64(stats.count)
		successRate := float64(stats.successes) / float64(stats.count)
		synergyScore := (avgQuality + successRate) / 2.0

		// Synergy detected if above baseline performance
		if synergyScore >= 0.75 {
			behaviors = append(behaviors, EmergentBehavior{
				BehaviorType:     "synergy",
				Description:      fmt.Sprintf("Pattern synergy detected: patterns %d & %d work exceptionally well together", stats.pattern1, stats.pattern2),
				InvolvedPatterns: []uint{stats.pattern1, stats.pattern2},
				NoveltyScore:     0.6,
				ImpactScore:      synergyScore,
				Confidence:       math.Min(float64(stats.count)/10.0, 0.95), // Higher confidence with more observations
				Evidence: map[string]interface{}{
					"co_occurrences": stats.count,
					"avg_quality":    avgQuality,
					"success_rate":   successRate,
					"synergy_score":  synergyScore,
				},
				IntegrationStatus: "pending",
			})

			log.Printf("   🔥 Synergy detected: patterns %d & %d (score=%.3f, n=%d)", stats.pattern1, stats.pattern2, synergyScore, stats.count)
		}
	}

	return behaviors
}

// detectBreakthroughs identifies significant quality improvements or novel solutions
func (ed *EmergenceDetector) detectBreakthroughs(decisions []Decision, cycles []ReasoningCycle) []EmergentBehavior {
	log.Println("\n🔍 Detecting reasoning breakthroughs...")

	behaviors := make([]EmergentBehavior, 0)

	// Analyze decisions for breakthrough performance
	for _, decision := range decisions {
		if decision.FinalQualityScore == nil || decision.InitialQualityScore == nil {
			continue
		}

		improvement := *decision.FinalQualityScore - *decision.InitialQualityScore

		// Breakthrough: >30% quality improvement
		if improvement >= 0.3 && *decision.FinalQualityScore >= 0.8 {
			behaviors = append(behaviors, EmergentBehavior{
				BehaviorType:     "breakthrough",
				Description:      fmt.Sprintf("Reasoning breakthrough: %.1f%% quality improvement to %.3f", improvement*100, *decision.FinalQualityScore),
				InvolvedPatterns: decision.PatternsConsidered,
				NoveltyScore:     math.Min(improvement*2, 1.0),
				ImpactScore:      *decision.FinalQualityScore,
				Confidence:       0.9,
				Evidence: map[string]interface{}{
					"decision_id":        decision.DecisionID,
					"initial_quality":    *decision.InitialQualityScore,
					"final_quality":      *decision.FinalQualityScore,
					"improvement":        improvement,
					"refactor_triggered": decision.RefactorTriggered,
				},
				IntegrationStatus: "pending",
			})

			log.Printf("   💡 Breakthrough detected: +%.1f%% → %.3f quality", improvement*100, *decision.FinalQualityScore)
		}
	}

	// Analyze reasoning cycles for meta-insights
	for _, cycle := range cycles {
		if len(cycle.Insights) >= 2 && cycle.CoherenceScore >= 0.8 {
			behaviors = append(behaviors, EmergentBehavior{
				BehaviorType:     "breakthrough",
				Description:      fmt.Sprintf("Multi-insight reasoning: %d insights with %.3f coherence", len(cycle.Insights), cycle.CoherenceScore),
				InvolvedPatterns: cycle.ActivatedPatterns,
				NoveltyScore:     0.7,
				ImpactScore:      cycle.CoherenceScore,
				Confidence:       0.8,
				Evidence: map[string]interface{}{
					"cycle_depth":       cycle.Depth,
					"insights":          cycle.Insights,
					"coherence_score":   cycle.CoherenceScore,
					"emergent_patterns": cycle.EmergentPatterns,
				},
				IntegrationStatus: "pending",
			})

			log.Printf("   💫 Meta-insight breakthrough: %d insights (coherence=%.3f)", len(cycle.Insights), cycle.CoherenceScore)
		}
	}

	return behaviors
}

// detectContradictions finds conflicting pattern activations
func (ed *EmergenceDetector) detectContradictions(decisions []Decision) []EmergentBehavior {
	log.Println("\n🔍 Detecting contradictions...")

	behaviors := make([]EmergentBehavior, 0)

	// Simplified: detect low-quality decisions with high confidence patterns
	for _, decision := range decisions {
		quality := 0.0
		if decision.FinalQualityScore != nil {
			quality = *decision.FinalQualityScore
		} else if decision.InitialQualityScore != nil {
			quality = *decision.InitialQualityScore
		}

		// Contradiction: high pattern confidence but low output quality
		if decision.ConfidenceLevel >= 0.7 && quality < 0.4 {
			behaviors = append(behaviors, EmergentBehavior{
				BehaviorType:     "contradiction",
				Description:      fmt.Sprintf("Contradiction: high pattern confidence (%.3f) but low quality (%.3f)", decision.ConfidenceLevel, quality),
				InvolvedPatterns: decision.PatternsConsidered,
				NoveltyScore:     0.5,
				ImpactScore:      decision.ConfidenceLevel - quality, // Size of contradiction
				Confidence:       0.7,
				Evidence: map[string]interface{}{
					"decision_id":        decision.DecisionID,
					"pattern_confidence": decision.ConfidenceLevel,
					"output_quality":     quality,
					"gap":                decision.ConfidenceLevel - quality,
				},
				IntegrationStatus: "pending",
			})

			log.Printf("   ⚠️ Contradiction detected: conf=%.3f, quality=%.3f (gap=%.3f)", decision.ConfidenceLevel, quality, decision.ConfidenceLevel-quality)
		}
	}

	return behaviors
}

// detectMetaPatterns identifies recurring patterns across reasoning cycles
func (ed *EmergenceDetector) detectMetaPatterns(cycles []ReasoningCycle) []EmergentBehavior {
	log.Println("\n🔍 Detecting meta-patterns...")

	behaviors := make([]EmergentBehavior, 0)

	if len(cycles) < 2 {
		return behaviors
	}

	// Track pattern activation sequences
	sequenceFreq := make(map[string]int)

	for _, cycle := range cycles {
		if len(cycle.ActivatedPatterns) >= 2 {
			// Create sequence signature (first 3 patterns)
			seqLen := minInt(3, len(cycle.ActivatedPatterns))
			seqKey := fmt.Sprintf("%v", cycle.ActivatedPatterns[:seqLen])
			sequenceFreq[seqKey]++
		}
	}

	// Detect recurring sequences
	for seqKey, freq := range sequenceFreq {
		if freq >= 2 { // Appeared 2+ times
			behaviors = append(behaviors, EmergentBehavior{
				BehaviorType:     "meta-pattern",
				Description:      fmt.Sprintf("Recurring reasoning sequence: appears %d times across cycles", freq),
				InvolvedPatterns: []uint{}, // Would parse from seqKey
				NoveltyScore:     0.6,
				ImpactScore:      math.Min(float64(freq)/5.0, 1.0),
				Confidence:       math.Min(float64(freq)/4.0, 0.9),
				Evidence: map[string]interface{}{
					"sequence":     seqKey,
					"frequency":    freq,
					"total_cycles": len(cycles),
				},
				IntegrationStatus: "pending",
			})

			log.Printf("   🔁 Meta-pattern detected: sequence appears %d times", freq)
		}
	}

	return behaviors
}

// Helper types and functions

type synergyStats struct {
	pattern1     uint
	pattern2     uint
	count        int
	successes    int
	totalQuality float64
}

func min(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}
