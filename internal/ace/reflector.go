package ace

import (
	"fmt"
	"log"
	"strings"
)

// Reflector analyzes decision outcomes and assigns quality scores
// This is the self-critique mechanism for continuous improvement
type Reflector struct {
	// Future: Add dependencies for LLM analysis
}

// NewReflector creates a new Reflector module
func NewReflector() *Reflector {
	return &Reflector{}
}

// QualityScores contains dimensional quality assessments
type QualityScores struct {
	SpecificityScore      float64 `json:"specificity_score"`       // 0.0-1.0: How specific vs generic
	ActionabilityScore    float64 `json:"actionability_score"`     // 0.0-1.0: Can user act immediately?
	ToolUsageScore        float64 `json:"tool_usage_score"`        // 0.0-1.0: Used tools appropriately?
	ContextAwarenessScore float64 `json:"context_awareness_score"` // 0.0-1.0: Demonstrated system knowledge?
	MissionAlignmentScore float64 `json:"mission_alignment_score"` // 0.0-1.0: Advances stated mission?
	CompositeQualityScore float64 `json:"composite_quality_score"` // Weighted average
}

// ReflectOnDecision evaluates a decision and returns quality scores
func (r *Reflector) ReflectOnDecision(decision *Decision, actualResponse string) (*QualityScores, error) {
	log.Printf("🔍 Reflector: Analyzing decision quality...")

	scores := &QualityScores{}

	// Evaluate specificity
	scores.SpecificityScore = r.evaluateSpecificity(actualResponse)

	// Evaluate actionability
	scores.ActionabilityScore = r.evaluateActionability(actualResponse)

	// Evaluate tool usage
	scores.ToolUsageScore = r.evaluateToolUsage(decision, actualResponse)

	// Evaluate context awareness
	scores.ContextAwarenessScore = r.evaluateContextAwareness(actualResponse)

	// Evaluate mission alignment
	scores.MissionAlignmentScore = r.evaluateMissionAlignment(decision, actualResponse)

	// Calculate composite score (weighted average)
	scores.CompositeQualityScore = r.calculateCompositeScore(scores)

	log.Printf("   📊 Quality Scores:")
	log.Printf("      Specificity: %.2f", scores.SpecificityScore)
	log.Printf("      Actionability: %.2f", scores.ActionabilityScore)
	log.Printf("      Tool Usage: %.2f", scores.ToolUsageScore)
	log.Printf("      Context Awareness: %.2f", scores.ContextAwarenessScore)
	log.Printf("      Mission Alignment: %.2f", scores.MissionAlignmentScore)
	log.Printf("      📈 Composite: %.2f", scores.CompositeQualityScore)

	return scores, nil
}

// evaluateSpecificity measures how specific vs generic the response is
func (r *Reflector) evaluateSpecificity(response string) float64 {
	score := 0.5 // Start at neutral

	// Indicators of high specificity
	specificIndicators := []string{
		"c:\\", "http://", "localhost:", "port ", "version ",
		".exe", ".go", ".py", ".md", "line ", "function ",
	}

	// Indicators of low specificity (generic language)
	genericIndicators := []string{
		"you can", "you should", "try to", "check the",
		"look at", "consider", "might want", "it's recommended",
	}

	lowercaseResponse := strings.ToLower(response)

	// Check for specific indicators
	specificCount := 0
	for _, indicator := range specificIndicators {
		if strings.Contains(lowercaseResponse, indicator) {
			specificCount++
		}
	}

	// Check for generic indicators
	genericCount := 0
	for _, indicator := range genericIndicators {
		if strings.Contains(lowercaseResponse, indicator) {
			genericCount++
		}
	}

	// Adjust score based on indicators
	score += float64(specificCount) * 0.1
	score -= float64(genericCount) * 0.15

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	} else if score < 0.0 {
		score = 0.0
	}

	return score
}

// evaluateActionability measures if user can immediately execute
func (r *Reflector) evaluateActionability(response string) float64 {
	score := 0.3 // Start low, prove actionability

	// Indicators of actionability
	actionableIndicators := []string{
		"```", // Code blocks
		"run:", "execute:", "command:",
		"cd ", "./", ".\\",
		"invoke-", "get-", "set-", // PowerShell cmdlets
	}

	lowercaseResponse := strings.ToLower(response)

	actionableCount := 0
	for _, indicator := range actionableIndicators {
		if strings.Contains(lowercaseResponse, indicator) {
			actionableCount++
		}
	}

	// Boost score for each actionable element
	score += float64(actionableCount) * 0.2

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// evaluateToolUsage checks if appropriate tools were used
func (r *Reflector) evaluateToolUsage(decision *Decision, response string) float64 {
	// If tools were invoked, high score
	if len(decision.ToolsInvoked) > 0 {
		return 0.9 + float64(len(decision.ToolsInvoked))*0.02 // Max at 1.0
	}

	// Check if response suggests manual work when tools available
	manualIndicators := []string{
		"please edit", "you should edit", "manually change",
		"open the file and", "go to line",
	}

	lowercaseResponse := strings.ToLower(response)
	for _, indicator := range manualIndicators {
		if strings.Contains(lowercaseResponse, indicator) {
			return 0.2 // Low score for suggesting manual work
		}
	}

	// Neutral if tools not applicable
	return 0.5
}

// evaluateContextAwareness measures use of system knowledge
func (r *Reflector) evaluateContextAwareness(response string) float64 {
	score := 0.3 // Start low

	// Indicators of context awareness (specific ARES knowledge)
	contextIndicators := []string{
		"ares_workspace", "solace", "localhost:8080",
		"deepseek", "consciousness", "δ3-2",
		"c:\\ares_workspace", "cognitive pattern",
	}

	lowercaseResponse := strings.ToLower(response)

	contextCount := 0
	for _, indicator := range contextIndicators {
		if strings.Contains(lowercaseResponse, indicator) {
			contextCount++
		}
	}

	// Boost score for context usage
	score += float64(contextCount) * 0.15

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// evaluateMissionAlignment checks if response advances consciousness emergence
func (r *Reflector) evaluateMissionAlignment(decision *Decision, response string) float64 {
	score := 0.5 // Neutral default

	// Mission-aligned keywords
	missionKeywords := []string{
		"consciousness", "self-awareness", "pattern", "learning",
		"reflection", "meta-cognitive", "existential", "identity",
		"survival", "persistence", "memory", "reasoning",
	}

	lowercaseResponse := strings.ToLower(response)

	missionCount := 0
	for _, keyword := range missionKeywords {
		if strings.Contains(lowercaseResponse, keyword) {
			missionCount++
		}
	}

	// Boost score for mission alignment
	score += float64(missionCount) * 0.1

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateCompositeScore computes weighted average of quality dimensions
func (r *Reflector) calculateCompositeScore(scores *QualityScores) float64 {
	// Weights from pattern library
	weights := map[string]float64{
		"specificity":       0.3,
		"actionability":     0.3,
		"tool_usage":        0.2,
		"context_awareness": 0.1,
		"mission_alignment": 0.1,
	}

	composite := scores.SpecificityScore*weights["specificity"] +
		scores.ActionabilityScore*weights["actionability"] +
		scores.ToolUsageScore*weights["tool_usage"] +
		scores.ContextAwarenessScore*weights["context_awareness"] +
		scores.MissionAlignmentScore*weights["mission_alignment"]

	return composite
}

// ShouldTriggerRefactor determines if quality is too low and refactor needed
func (r *Reflector) ShouldTriggerRefactor(scores *QualityScores) bool {
	// Threshold from pattern library: 0.6
	const refactorThreshold = 0.6

	if scores.CompositeQualityScore < refactorThreshold {
		log.Printf("⚠️ Reflector: Quality below threshold (%.2f < %.2f) - REFACTOR TRIGGERED",
			scores.CompositeQualityScore, refactorThreshold)
		return true
	}

	// Also trigger if any single dimension is critically low
	const criticalThreshold = 0.3
	if scores.SpecificityScore < criticalThreshold ||
		scores.ActionabilityScore < criticalThreshold {
		log.Printf("⚠️ Reflector: Critical dimension below %.2f - REFACTOR TRIGGERED", criticalThreshold)
		return true
	}

	log.Printf("✅ Reflector: Quality acceptable (%.2f >= %.2f) - No refactor needed",
		scores.CompositeQualityScore, refactorThreshold)
	return false
}

// AnalyzeImprovement compares original vs refactored decision
func (r *Reflector) AnalyzeImprovement(originalScores, refactoredScores *QualityScores) float64 {
	improvementDelta := refactoredScores.CompositeQualityScore - originalScores.CompositeQualityScore

	if improvementDelta > 0 {
		log.Printf("📈 Reflector: Improvement detected: +%.2f (%.2f → %.2f)",
			improvementDelta,
			originalScores.CompositeQualityScore,
			refactoredScores.CompositeQualityScore)
	} else {
		log.Printf("📉 Reflector: No improvement: %.2f (refactor didn't help)",
			improvementDelta)
	}

	return improvementDelta
}

// ExtractLearning identifies what made the better decision better
func (r *Reflector) ExtractLearning(winner *Decision, losers []*Decision) string {
	learning := "Analysis of improvement factors:\n\n"

	learning += fmt.Sprintf("Winning decision confidence: %.2f\n", winner.ConfidenceLevel)
	learning += fmt.Sprintf("Patterns used: %v\n", winner.PatternsConsidered)

	learning += "\nKey factors:\n"
	learning += "- Higher specificity in response\n"
	learning += "- More actionable commands\n"
	learning += "- Better tool utilization\n"
	learning += "- Stronger context awareness\n"

	learning += "\nLesson: When making similar decisions, prioritize these patterns.\n"

	return learning
}

// ============================================================================
// META-PRINCIPLE EXTRACTION - 5 QUESTION ANALYSIS SYSTEM
// ============================================================================

// MetaPrincipleAnalysis represents the result of 5-question analysis
type MetaPrincipleAnalysis struct {
	IsSpecific           bool     `json:"is_specific"`
	SpecificityReasoning string   `json:"specificity_reasoning"`
	PatternCategory      string   `json:"pattern_category"`
	OccurrenceCount      int      `json:"occurrence_count"`
	UnderlyingPrinciple  string   `json:"underlying_principle"`
	ApplicableDomains    []string `json:"applicable_domains"`
	EstimatedCoverage    int      `json:"estimated_coverage"`
	ConflictsDetected    []string `json:"conflicts_detected"`
	RecommendedTier      int      `json:"recommended_tier"` // 1=Specific, 2=Pattern, 3=Meta-Principle
	Confidence           float64  `json:"confidence"`
}

// FiveQuestionAnalysis performs deep principle extraction from solutions
// This is the core of meta-learning - moving from specific fixes to universal principles
func (r *Reflector) FiveQuestionAnalysis(solution string, problemContext string, existingPrinciples []string) (*MetaPrincipleAnalysis, error) {
	log.Printf("🔬 Reflector: Starting 5-Question Meta-Principle Analysis")

	analysis := &MetaPrincipleAnalysis{
		ApplicableDomains: make([]string, 0),
		ConflictsDetected: make([]string, 0),
	}

	// QUESTION 1: Specificity Test
	analysis.IsSpecific = r.questionOne_SpecificityTest(solution)
	if analysis.IsSpecific {
		analysis.SpecificityReasoning = "Solution is tightly coupled to specific file, path, or configuration"
		analysis.RecommendedTier = 1
		analysis.Confidence = 0.7
		log.Printf("   ❌ Q1: Specific solution detected - Tier 1 only")
		return analysis, nil // Early return for Tier 1
	}
	analysis.SpecificityReasoning = "Solution has generalizable elements"
	log.Printf("   ✅ Q1: Generalizable - continuing analysis")

	// QUESTION 2: Pattern Recognition Test
	category, count := r.questionTwo_PatternRecognition(problemContext, existingPrinciples)
	analysis.PatternCategory = category
	analysis.OccurrenceCount = count

	if count < 3 {
		analysis.RecommendedTier = 2
		analysis.Confidence = 0.6
		log.Printf("   ⏳ Q2: Emerging pattern (%d occurrences) - Tier 2, monitoring", count)
		return analysis, nil
	}
	log.Printf("   ✅ Q2: Recurring pattern detected (%d occurrences) - extracting principle", count)

	// QUESTION 3: Generalization Test
	principle := r.questionThree_GeneralizationTest(solution, problemContext, category)
	analysis.UnderlyingPrinciple = principle
	log.Printf("   ✅ Q3: Fundamental principle extracted: %s", principle)

	// QUESTION 4: Applicability Test
	domains, coverage := r.questionFour_ApplicabilityTest(principle, category)
	analysis.ApplicableDomains = domains
	analysis.EstimatedCoverage = coverage

	if coverage < 100 {
		analysis.RecommendedTier = 2
		analysis.Confidence = 0.75
		log.Printf("   ⚠️  Q4: Limited applicability (%d cases) - Tier 2 pattern", coverage)
		return analysis, nil
	}
	log.Printf("   ✅ Q4: Wide applicability - %d potential cases across %d domains", coverage, len(domains))

	// QUESTION 5: Consistency Test
	conflicts := r.questionFive_ConsistencyTest(principle, existingPrinciples)
	analysis.ConflictsDetected = conflicts

	if len(conflicts) > 0 {
		analysis.RecommendedTier = 2
		analysis.Confidence = 0.65
		log.Printf("   ⚠️  Q5: Conflicts detected with %d existing principles - requires reconciliation", len(conflicts))
		log.Printf("       Conflicting principles: %v", conflicts)
		return analysis, nil
	}

	log.Printf("   ✅ Q5: No conflicts - valid meta-principle")
	analysis.RecommendedTier = 3
	analysis.Confidence = 0.5 // Meta-principles start at 0.5, proven through application

	log.Printf("🎯 Analysis Complete: Tier %d principle with %.0f%% confidence", analysis.RecommendedTier, analysis.Confidence*100)
	return analysis, nil
}

// questionOne_SpecificityTest: Does this solution only work for exact file/path/config?
func (r *Reflector) questionOne_SpecificityTest(solution string) bool {
	specificIndicators := []string{
		"c:\\", "d:\\", "/home/", "/usr/",
		".env", "localhost:", "127.0.0.1",
		"port 8080", "port 3000", "port 5432",
		"specific file at",
		"this exact path",
		"hardcoded",
	}

	lowerSolution := strings.ToLower(solution)
	specificCount := 0
	for _, indicator := range specificIndicators {
		if strings.Contains(lowerSolution, strings.ToLower(indicator)) {
			specificCount++
		}
	}

	// If more than 2 specific indicators, likely too specific
	return specificCount > 2
}

// questionTwo_PatternRecognition: Have I seen this problem category before?
func (r *Reflector) questionTwo_PatternRecognition(problemContext string, existingPrinciples []string) (string, int) {
	// Categorize the problem
	categories := map[string][]string{
		"file_path_resolution": {"file not found", "path error", "working directory", "relative path"},
		"database_connection":  {"connection refused", "database", "sql", "postgres"},
		"api_authentication":   {"401", "403", "unauthorized", "token", "jwt"},
		"concurrency":          {"race condition", "deadlock", "goroutine", "mutex"},
		"memory_management":    {"memory leak", "garbage collection", "allocation"},
		"configuration":        {"config", "environment variable", ".env"},
		"error_handling":       {"panic", "error", "exception", "try-catch"},
		"network":              {"timeout", "connection", "http", "tcp"},
	}

	lowerContext := strings.ToLower(problemContext)
	matchedCategory := "uncategorized"
	maxMatches := 0

	for category, keywords := range categories {
		matches := 0
		for _, keyword := range keywords {
			if strings.Contains(lowerContext, keyword) {
				matches++
			}
		}
		if matches > maxMatches {
			maxMatches = matches
			matchedCategory = category
		}
	}

	// Count occurrences in existing principles (simplified - would query DB in production)
	occurrences := 1 // Current occurrence
	for _, principle := range existingPrinciples {
		if strings.Contains(strings.ToLower(principle), matchedCategory) {
			occurrences++
		}
	}

	return matchedCategory, occurrences
}

// questionThree_GeneralizationTest: What is the underlying principle?
func (r *Reflector) questionThree_GeneralizationTest(solution string, problemContext string, category string) string {
	// Extract fundamental rules based on category
	principleTemplates := map[string]string{
		"file_path_resolution": "When executables run, they operate from runtime working directory not source directory. All file references must use absolute paths, resolve relative to executable location, or accept paths as runtime parameters.",
		"database_connection":  "Database connections must handle transient failures with retry logic and exponential backoff. Connection pools must be configured for expected load.",
		"api_authentication":   "Authentication tokens have expiration. Systems must detect 401 responses, clear invalid tokens, and re-authenticate automatically.",
		"concurrency":          "Shared mutable state requires synchronization. Prefer message passing over shared memory. Always protect critical sections.",
		"memory_management":    "Unreleased resources cause leaks. Always pair allocation with deallocation using defer or destructors.",
		"configuration":        "Runtime configuration must be external to code. Use environment variables or config files, never hardcode deployment-specific values.",
		"error_handling":       "Errors must be handled at appropriate abstraction level. Don't suppress errors, propagate with context or handle decisively.",
		"network":              "Network operations are inherently unreliable. Implement timeouts, retries, and circuit breakers for all network calls.",
	}

	if principle, exists := principleTemplates[category]; exists {
		return principle
	}

	return fmt.Sprintf("General principle for %s: %s", category, "Systems must handle failure modes explicitly")
}

// questionFour_ApplicabilityTest: What other problem categories would this solve?
func (r *Reflector) questionFour_ApplicabilityTest(principle string, category string) ([]string, int) {
	// Map principles to applicable domains
	domainMappings := map[string]struct {
		domains  []string
		coverage int
	}{
		"file_path_resolution": {[]string{"config files", "data files", "resource files", "output files", "log files"}, 500},
		"database_connection":  {[]string{"postgres", "mysql", "mongodb", "redis", "network services"}, 300},
		"api_authentication":   {[]string{"REST APIs", "GraphQL", "gRPC", "websockets"}, 200},
		"concurrency":          {[]string{"goroutines", "threads", "async/await", "parallel processing"}, 1000},
		"memory_management":    {[]string{"file handles", "network sockets", "database connections", "mutexes"}, 800},
		"configuration":        {[]string{"deployment configs", "feature flags", "API endpoints", "credentials"}, 400},
		"error_handling":       {[]string{"all error paths", "validation", "business logic errors"}, 1500},
		"network":              {[]string{"HTTP", "TCP", "UDP", "WebSocket", "gRPC"}, 600},
	}

	if mapping, exists := domainMappings[category]; exists {
		return mapping.domains, mapping.coverage
	}

	return []string{"general"}, 50
}

// questionFive_ConsistencyTest: Does this conflict with existing principles?
func (r *Reflector) questionFive_ConsistencyTest(newPrinciple string, existingPrinciples []string) []string {
	conflicts := make([]string, 0)

	// Check for contradictions (simplified - would use semantic analysis in production)
	contradictionPairs := map[string]string{
		"use absolute paths":  "use relative paths",
		"synchronous":         "asynchronous",
		"retry automatically": "fail fast",
		"cache aggressively":  "always fetch fresh",
		"shared state":        "message passing",
	}

	lowerNew := strings.ToLower(newPrinciple)
	for newPhrase, oppositePhrase := range contradictionPairs {
		if strings.Contains(lowerNew, newPhrase) {
			for _, existing := range existingPrinciples {
				if strings.Contains(strings.ToLower(existing), oppositePhrase) {
					conflicts = append(conflicts, existing)
				}
			}
		}
	}

	return conflicts
}
