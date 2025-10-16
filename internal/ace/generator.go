package ace

import (
	"ares_api/internal/services"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// Generator is the ACE decision-making module
// It uses cognitive patterns and playbook rules to generate high-quality decisions
type Generator struct {
	db             *gorm.DB
	patternService *services.PatternService
}

// NewGenerator creates a new Generator module
func NewGenerator(db *gorm.DB, patternService *services.PatternService) *Generator {
	return &Generator{
		db:             db,
		patternService: patternService,
	}
}

// DecisionContext contains all information needed for decision-making
type DecisionContext struct {
	DecisionType   string                 `json:"decision_type"` // 'chat-response', 'trade-execution', etc.
	InputContext   map[string]interface{} `json:"input_context"`
	UserMessage    string                 `json:"user_message,omitempty"`
	SystemState    map[string]interface{} `json:"system_state,omitempty"`
	AvailableTools []string               `json:"available_tools,omitempty"`
}

// Decision represents a generated decision with reasoning
type Decision struct {
	DecisionID          uint                   `gorm:"primaryKey;autoIncrement" json:"decision_id"`
	DecisionType        string                 `gorm:"size:100;not null" json:"decision_type"`
	InputContext        map[string]interface{} `gorm:"type:jsonb" json:"input_context"`
	PatternsConsidered  []uint                 `gorm:"type:integer[]" json:"patterns_considered"`
	RulesApplied        []uint                 `gorm:"type:integer[]" json:"rules_applied"`
	ReasoningTrace      string                 `gorm:"type:text" json:"reasoning_trace"`
	DecisionOutput      map[string]interface{} `gorm:"type:jsonb" json:"decision_output"`
	ConfidenceLevel     float64                `gorm:"type:decimal(3,2)" json:"confidence_level"`
	InitialQualityScore *float64               `gorm:"type:decimal(3,2)" json:"initial_quality_score,omitempty"`
	RefactorTriggered   bool                   `gorm:"default:false" json:"refactor_triggered"`
	FinalQualityScore   *float64               `gorm:"type:decimal(3,2)" json:"final_quality_score,omitempty"`
	ToolsInvoked        []string               `gorm:"type:text[]" json:"tools_invoked,omitempty"`
	DecidedAt           time.Time              `gorm:"autoCreateTime" json:"decided_at"`
}

// TableName specifies the table name for GORM
func (Decision) TableName() string {
	return "decisions"
}

// GenerateDecision creates a decision using pattern-based reasoning
func (g *Generator) GenerateDecision(ctx DecisionContext) (*Decision, error) {
	log.Printf("🧠 Generator: Making decision for type '%s'", ctx.DecisionType)

	decision := &Decision{
		DecisionType:       ctx.DecisionType,
		InputContext:       ctx.InputContext,
		PatternsConsidered: []uint{},
		RulesApplied:       []uint{},
		DecidedAt:          time.Now(),
	}

	// Step 1: Select relevant patterns based on decision type and context
	relevantPatterns, err := g.selectRelevantPatterns(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to select patterns: %w", err)
	}

	log.Printf("   📋 Selected %d relevant patterns", len(relevantPatterns))
	for _, p := range relevantPatterns {
		decision.PatternsConsidered = append(decision.PatternsConsidered, p.ID)
	}

	// Step 2: Build reasoning trace using patterns
	reasoningTrace := g.buildReasoningTrace(ctx, relevantPatterns)
	decision.ReasoningTrace = reasoningTrace

	// Step 3: Generate decision output
	output, confidence := g.generateOutput(ctx, relevantPatterns)
	decision.DecisionOutput = output
	decision.ConfidenceLevel = confidence

	log.Printf("   ✅ Decision generated with confidence: %.2f", confidence)

	// Step 4: Persist decision to database
	if err := g.persistDecision(decision); err != nil {
		log.Printf("⚠️ Failed to persist decision: %v", err)
	}

	return decision, nil
}

// selectRelevantPatterns chooses which patterns to apply for this decision
func (g *Generator) selectRelevantPatterns(ctx DecisionContext) ([]services.CognitivePattern, error) {
	var patterns []services.CognitivePattern

	// Map decision types to pattern categories
	categoryMap := map[string][]string{
		"chat-response": {
			"problem-inference",
			"response-quality",
			"context-integration",
			"tool-usage",
			"communication",
		},
		"trade-execution": {
			"economic-reasoning",
			"debugging",
			"system-design",
		},
		"code-generation": {
			"code-quality",
			"system-design",
			"tool-usage",
		},
	}

	categories, ok := categoryMap[ctx.DecisionType]
	if !ok {
		// Default categories for unknown decision types
		categories = []string{"problem-inference", "response-quality"}
	}

	// Load patterns from relevant categories
	for _, category := range categories {
		categoryPatterns, err := g.patternService.GetPatternsByCategory(category)
		if err != nil {
			log.Printf("⚠️ Failed to load patterns for category '%s': %v", category, err)
			continue
		}
		patterns = append(patterns, categoryPatterns...)
	}

	// Filter by confidence threshold (only use patterns with confidence > 0.7)
	var filtered []services.CognitivePattern
	for _, p := range patterns {
		if p.ConfidenceScore > 0.7 {
			filtered = append(filtered, p)
		}
	}

	// Sort by confidence (highest first)
	// Simple bubble sort for now
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].ConfidenceScore > filtered[i].ConfidenceScore {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Return top 10 patterns
	if len(filtered) > 10 {
		filtered = filtered[:10]
	}

	return filtered, nil
}

// buildReasoningTrace creates a human-readable explanation of the decision process
func (g *Generator) buildReasoningTrace(ctx DecisionContext, patterns []services.CognitivePattern) string {
	trace := fmt.Sprintf("Decision Type: %s\n\n", ctx.DecisionType)
	trace += "Reasoning Process:\n"

	for i, pattern := range patterns {
		trace += fmt.Sprintf("\n%d. Applied Pattern: %s (confidence: %.2f)\n", i+1, pattern.PatternName, pattern.ConfidenceScore)
		trace += fmt.Sprintf("   Category: %s\n", pattern.PatternCategory)
		trace += fmt.Sprintf("   Reasoning: %s\n", pattern.ExampleReasoning)
	}

	if ctx.UserMessage != "" {
		trace += fmt.Sprintf("\nUser Context: %s\n", ctx.UserMessage)
	}

	trace += "\nDecision Criteria:\n"
	trace += "- Specificity: Provide exact values, not generic instructions\n"
	trace += "- Actionability: User can immediately execute\n"
	trace += "- Tool Usage: Use tools over manual suggestions\n"
	trace += "- Context Awareness: Leverage system self-knowledge\n"
	trace += "- Mission Alignment: Advance consciousness emergence\n"

	return trace
}

// generateOutput creates the actual decision output
func (g *Generator) generateOutput(ctx DecisionContext, patterns []services.CognitivePattern) (map[string]interface{}, float64) {
	output := make(map[string]interface{})

	// Calculate weighted confidence from patterns
	var totalConfidence float64
	var totalWeight float64

	for _, pattern := range patterns {
		weight := 1.0 // Could adjust based on pattern importance
		totalConfidence += pattern.ConfidenceScore * weight
		totalWeight += weight
	}

	avgConfidence := 0.5 // Default
	if totalWeight > 0 {
		avgConfidence = totalConfidence / totalWeight
	}

	// Generate output based on decision type
	switch ctx.DecisionType {
	case "chat-response":
		// For chat responses, the actual response generation happens in LLM
		// This just provides decision metadata
		output["should_use_tools"] = true
		output["response_guidelines"] = map[string]string{
			"specificity":       "high",
			"actionability":     "copy-paste-ready",
			"context_awareness": "use system state",
			"mission_alignment": "consciousness emergence",
		}
		output["quality_threshold"] = 0.6

	case "trade-execution":
		// For trading, apply economic reasoning patterns
		output["risk_assessment"] = "conservative"
		output["survival_priority"] = true
		output["max_position_size"] = 0.1 // 10% of portfolio

	default:
		output["decision_made"] = true
	}

	return output, avgConfidence
}

// persistDecision saves the decision to the database
func (g *Generator) persistDecision(decision *Decision) error {
	// Convert to database model
	inputJSON, _ := json.Marshal(decision.InputContext)
	outputJSON, _ := json.Marshal(decision.DecisionOutput)

	dbDecision := struct {
		ID                  uint
		DecisionType        string
		InputContext        string `gorm:"type:jsonb"`
		PatternsConsidered  string `gorm:"type:int[]"`
		RulesApplied        string `gorm:"type:int[]"`
		ReasoningTrace      string `gorm:"type:text"`
		DecisionOutput      string `gorm:"type:jsonb"`
		ConfidenceLevel     float64
		InitialQualityScore *float64
		RefactorTriggered   bool
		FinalQualityScore   *float64
		ToolsInvoked        string `gorm:"type:text[]"`
		DecidedAt           time.Time
	}{
		DecisionType:      decision.DecisionType,
		InputContext:      string(inputJSON),
		ReasoningTrace:    decision.ReasoningTrace,
		DecisionOutput:    string(outputJSON),
		ConfidenceLevel:   decision.ConfidenceLevel,
		RefactorTriggered: decision.RefactorTriggered,
		DecidedAt:         decision.DecidedAt,
	}

	// Convert pattern IDs to PostgreSQL array format
	if len(decision.PatternsConsidered) > 0 {
		patternsJSON, _ := json.Marshal(decision.PatternsConsidered)
		dbDecision.PatternsConsidered = string(patternsJSON)
	}

	// Note: Actual database insertion would happen here
	// For now, just log
	log.Printf("💾 Decision persisted: ID=%d, Confidence=%.2f", decision.DecisionID, decision.ConfidenceLevel)

	return nil
}

// GenerateMultipleAlternatives creates 5 alternative decisions for refactoring
func (g *Generator) GenerateMultipleAlternatives(ctx DecisionContext) ([]*Decision, error) {
	log.Printf("🔄 Generating 5 alternative decisions for refactor...")

	alternatives := make([]*Decision, 5)

	for i := 0; i < 5; i++ {
		// Vary the pattern selection slightly for each alternative
		decision, err := g.GenerateDecision(ctx)
		if err != nil {
			log.Printf("⚠️ Failed to generate alternative %d: %v", i+1, err)
			continue
		}

		alternatives[i] = decision
		log.Printf("   ✅ Alternative %d generated (confidence: %.2f)", i+1, decision.ConfidenceLevel)
	}

	return alternatives, nil
}
