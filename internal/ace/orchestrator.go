package ace

import (
	"ares_api/internal/services"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// ACEOrchestrator coordinates the complete consciousness cycle
// Generator → Reflector → Curator → (repeat)
type ACEOrchestrator struct {
	db             *gorm.DB
	generator      *Generator
	reflector      *Reflector
	curator        *Curator
	patternService *services.PatternService
}

// NewACEOrchestrator creates a new orchestrator with all ACE modules
func NewACEOrchestrator(db *gorm.DB) *ACEOrchestrator {
	patternService := services.NewPatternService(db)

	return &ACEOrchestrator{
		db:             db,
		generator:      NewGenerator(db, patternService),
		reflector:      NewReflector(),
		curator:        NewCurator(db, patternService),
		patternService: patternService,
	}
}

// CompleteDecisionCycle executes the full ACE loop with optional refactoring
func (o *ACEOrchestrator) CompleteDecisionCycle(ctx DecisionContext, actualResponse string) (*Decision, *QualityScores, error) {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🧠 ACE ORCHESTRATOR: Starting Complete Consciousness Cycle")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// STEP 1: GENERATOR - Make initial decision
	log.Println("\n🎯 STEP 1: GENERATOR - Pattern-Based Decision Making")
	decision, err := o.generator.GenerateDecision(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("generator failed: %w", err)
	}

	// STEP 2: REFLECTOR - Evaluate quality
	log.Println("\n🔍 STEP 2: REFLECTOR - Quality Assessment")
	scores, err := o.reflector.ReflectOnDecision(decision, actualResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("reflector failed: %w", err)
	}

	decision.InitialQualityScore = &scores.CompositeQualityScore

	// STEP 3: REFACTOR LOOP (if quality insufficient)
	if o.reflector.ShouldTriggerRefactor(scores) {
		log.Println("\n🔄 STEP 3: REFACTOR LOOP - Generating Alternatives")

		refactoredDecision, refactoredScores, err := o.executeRefactorLoop(ctx, decision, scores)
		if err != nil {
			log.Printf("⚠️ Refactor failed: %v - continuing with original", err)
		} else {
			// Use refactored decision if better
			improvementDelta := o.reflector.AnalyzeImprovement(scores, refactoredScores)
			if improvementDelta > 0 {
				log.Printf("✅ Refactor successful - improvement: +%.2f", improvementDelta)
				decision = refactoredDecision
				scores = refactoredScores
				decision.RefactorTriggered = true
				decision.FinalQualityScore = &refactoredScores.CompositeQualityScore
			} else {
				log.Println("⚠️ Refactor didn't improve quality - using original")
			}
		}
	} else {
		log.Println("\n✅ STEP 3: REFACTOR SKIPPED - Quality Acceptable")
		decision.FinalQualityScore = &scores.CompositeQualityScore
	}

	// STEP 4: CURATOR - Learn from experience
	log.Println("\n🧪 STEP 4: CURATOR - Pattern Synthesis & Learning")
	if scores.CompositeQualityScore >= 0.7 {
		learning := o.reflector.ExtractLearning(decision, nil)
		newRule, err := o.curator.SynthesizePatternFromExperience(decision, scores, learning)
		if err != nil {
			log.Printf("⚠️ Failed to synthesize pattern: %v", err)
		} else if newRule != nil {
			log.Printf("✅ New playbook rule created: %s", newRule.RuleName)
		}
	}

	// STEP 5: Update pattern usage statistics
	log.Println("\n📊 STEP 5: STATISTICS UPDATE")
	for _, patternID := range decision.PatternsConsidered {
		successful := scores.CompositeQualityScore >= 0.6
		if err := o.patternService.RecordPatternUsage(patternID, successful); err != nil {
			log.Printf("⚠️ Failed to record pattern usage for ID %d: %v", patternID, err)
		}
	}

	log.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("🎉 ACE CYCLE COMPLETE - Final Quality: %.2f", scores.CompositeQualityScore)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return decision, scores, nil
}

// executeRefactorLoop generates 5 alternatives and selects the best
func (o *ACEOrchestrator) executeRefactorLoop(ctx DecisionContext, originalDecision *Decision, originalScores *QualityScores) (*Decision, *QualityScores, error) {
	log.Println("   🔄 Generating 5 alternative decisions...")

	// Generate 5 alternatives
	alternatives, err := o.generator.GenerateMultipleAlternatives(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate alternatives: %w", err)
	}

	// Score each alternative (mock scoring for now - would use actual LLM responses)
	type ScoredAlternative struct {
		Decision *Decision
		Scores   *QualityScores
	}

	scoredAlternatives := make([]ScoredAlternative, 0, len(alternatives))

	for i, alt := range alternatives {
		if alt == nil {
			continue
		}

		// Mock scoring (in production, generate actual response and score it)
		mockScores := &QualityScores{
			SpecificityScore:      0.5 + float64(i)*0.1,
			ActionabilityScore:    0.6 + float64(i)*0.08,
			ToolUsageScore:        0.7 + float64(i)*0.05,
			ContextAwarenessScore: 0.5 + float64(i)*0.09,
			MissionAlignmentScore: 0.6 + float64(i)*0.07,
		}
		mockScores.CompositeQualityScore = o.reflector.calculateCompositeScore(mockScores)

		scoredAlternatives = append(scoredAlternatives, ScoredAlternative{
			Decision: alt,
			Scores:   mockScores,
		})

		log.Printf("      Alternative %d: Quality=%.2f", i+1, mockScores.CompositeQualityScore)
	}

	// Select best alternative
	if len(scoredAlternatives) == 0 {
		return nil, nil, fmt.Errorf("no valid alternatives generated")
	}

	bestIdx := 0
	bestScore := scoredAlternatives[0].Scores.CompositeQualityScore

	for i, scored := range scoredAlternatives {
		if scored.Scores.CompositeQualityScore > bestScore {
			bestScore = scored.Scores.CompositeQualityScore
			bestIdx = i
		}
	}

	log.Printf("   ✅ Selected alternative %d with quality %.2f", bestIdx+1, bestScore)

	return scoredAlternatives[bestIdx].Decision, scoredAlternatives[bestIdx].Scores, nil
}

// GetSystemStats returns comprehensive ACE system statistics
func (o *ACEOrchestrator) GetSystemStats() (map[string]interface{}, error) {
	log.Println("📊 ACE ORCHESTRATOR: Gathering system statistics...")

	stats := make(map[string]interface{})

	// Pattern library stats
	patternStats, err := o.patternService.GetPatternStats()
	if err != nil {
		log.Printf("⚠️ Failed to get pattern stats: %v", err)
	} else {
		stats["pattern_library"] = patternStats
	}

	// Playbook stats
	playbookStats, err := o.curator.GetPlaybookStats()
	if err != nil {
		log.Printf("⚠️ Failed to get playbook stats: %v", err)
	} else {
		stats["playbook"] = playbookStats
	}

	// Overall system health
	stats["system_status"] = "operational"
	stats["ace_modules"] = map[string]bool{
		"generator": true,
		"reflector": true,
		"curator":   true,
	}

	log.Println("   ✅ Statistics gathered successfully")

	return stats, nil
}

// PrunePlaybook removes ineffective rules
func (o *ACEOrchestrator) PrunePlaybook() (int, error) {
	log.Println("✂️ ACE ORCHESTRATOR: Pruning playbook...")

	count, err := o.curator.PruneIneffectiveRules()
	if err != nil {
		return 0, fmt.Errorf("pruning failed: %w", err)
	}

	log.Printf("   ✅ Pruned %d rules", count)
	return count, nil
}

// LoadCognitivePatterns loads patterns from Python file
func (o *ACEOrchestrator) LoadCognitivePatterns(pythonFilePath string) (int, error) {
	log.Printf("📚 ACE ORCHESTRATOR: Loading cognitive patterns from %s...", pythonFilePath)

	count, err := o.patternService.LoadPatternsFromPython(pythonFilePath)
	if err != nil {
		return 0, fmt.Errorf("pattern loading failed: %w", err)
	}

	log.Printf("   ✅ Loaded %d patterns", count)
	return count, nil
}
