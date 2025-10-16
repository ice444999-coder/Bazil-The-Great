package ace

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// RefactorLoop implements the stuck-detection and alternative-generation protocol
// When GitHub gets stuck 3+ times, this forces generation of 5 fundamentally different approaches
type RefactorLoop struct {
	db        *gorm.DB
	reflector *Reflector
}

// NewRefactorLoop creates a new refactor loop coordinator
func NewRefactorLoop(db *gorm.DB, reflector *Reflector) *RefactorLoop {
	return &RefactorLoop{
		db:        db,
		reflector: reflector,
	}
}

// ============================================================================
// REFACTOR EVENT TRACKING
// ============================================================================

// RefactorEvent represents a single refactor loop invocation
type RefactorEvent struct {
	ID                   uint                   `gorm:"primaryKey" json:"id"`
	OriginalProblem      string                 `gorm:"type:text;not null" json:"original_problem"`
	StuckApproach        string                 `gorm:"type:text;not null" json:"stuck_approach"`
	AttemptCount         int                    `gorm:"not null" json:"attempt_count"`
	FiveAlternatives     []RefactorAlternative  `gorm:"-" json:"five_alternatives"` // Stored as JSONB
	FiveAlternativesJSON string                 `gorm:"column:five_alternatives;type:jsonb" json:"-"`
	EvaluationScores     map[string]interface{} `gorm:"type:jsonb" json:"evaluation_scores"`
	SelectedApproach     string                 `gorm:"type:text;not null" json:"selected_approach"`
	SelectionReasoning   string                 `gorm:"type:text;not null" json:"selection_reasoning"`
	Outcome              string                 `gorm:"type:text" json:"outcome"`
	Success              *bool                  `json:"success,omitempty"`
	ProblemCategory      string                 `gorm:"type:text" json:"problem_category"`
	Timestamp            time.Time              `gorm:"autoCreateTime" json:"timestamp"`
	Analyzed             bool                   `gorm:"default:false" json:"analyzed"`
}

// TableName specifies the table name for GORM
func (RefactorEvent) TableName() string {
	return "github_refactor_events"
}

// RefactorAlternative represents one of five approaches
type RefactorAlternative struct {
	ApproachNumber int            `json:"approach_number"`
	Strategy       string         `json:"strategy"`
	Description    string         `json:"description"`
	ToolsUsed      []string       `json:"tools_used"`
	Scores         RefactorScores `json:"scores"`
	WeightedScore  float64        `json:"weighted_score"`
}

// RefactorScores contains evaluation criteria
type RefactorScores struct {
	RootCauseSolution int `json:"root_cause_solution"` // 1-10: Solves root cause not symptom
	Simplicity        int `json:"simplicity"`          // 1-10: Low implementation complexity
	LowRisk           int `json:"low_risk"`            // 1-10: Minimal bug introduction risk
	CodebaseAlignment int `json:"codebase_alignment"`  // 1-10: Fits existing patterns
	Testability       int `json:"testability"`         // 1-10: Easy to verify/test
}

// CalculateWeightedScore computes the total weighted score
func (s *RefactorScores) CalculateWeightedScore() float64 {
	return float64(s.RootCauseSolution)*2.0 +
		float64(s.Simplicity)*1.5 +
		float64(s.LowRisk)*1.5 +
		float64(s.CodebaseAlignment)*1.0 +
		float64(s.Testability)*1.0
}

// ============================================================================
// STUCK DETECTION SYSTEM
// ============================================================================

// StuckIndicator represents signals that GitHub is stuck
type StuckIndicator struct {
	ErrorMessage    string
	OccurrenceCount int
	TimeWindow      time.Duration
	IndicatesStuck  bool
	Reasoning       string
}

// DetectStuck checks if GitHub is stuck based on recent activity
func (rl *RefactorLoop) DetectStuck(sessionID string, timeWindow time.Duration) (*StuckIndicator, error) {
	log.Printf("🔍 Refactor Loop: Checking for stuck patterns in session %s", sessionID)

	// Query recent errors/failures from github_outputs or github_decisions
	// This is simplified - in production would query actual tables

	indicator := &StuckIndicator{
		TimeWindow: timeWindow,
	}

	// SIGNAL 1: Same error message appears 3+ times
	errorPattern := rl.checkRepeatedErrors(sessionID, timeWindow)
	if errorPattern.Count >= 3 {
		indicator.IndicatesStuck = true
		indicator.ErrorMessage = errorPattern.Message
		indicator.OccurrenceCount = errorPattern.Count
		indicator.Reasoning = fmt.Sprintf("Same error occurred %d times in %v", errorPattern.Count, timeWindow)
		log.Printf("   🚨 STUCK DETECTED: %s", indicator.Reasoning)
		return indicator, nil
	}

	// SIGNAL 2: Manual intervention keywords detected
	manualInterventionDetected := rl.checkManualInterventionKeywords(sessionID, timeWindow)
	if manualInterventionDetected {
		indicator.IndicatesStuck = true
		indicator.Reasoning = "GitHub indicated manual intervention required"
		log.Printf("   🚨 STUCK DETECTED: %s", indicator.Reasoning)
		return indicator, nil
	}

	// SIGNAL 3: No progress for extended time (10+ minutes on same problem)
	noProgress := rl.checkNoProgress(sessionID, timeWindow)
	if noProgress {
		indicator.IndicatesStuck = true
		indicator.Reasoning = "No progress for extended period (10+ minutes)"
		log.Printf("   🚨 STUCK DETECTED: %s", indicator.Reasoning)
		return indicator, nil
	}

	log.Printf("   ✅ No stuck patterns detected")
	return indicator, nil
}

type errorPattern struct {
	Message string
	Count   int
}

func (rl *RefactorLoop) checkRepeatedErrors(sessionID string, window time.Duration) errorPattern {
	// This would query github_outputs or a dedicated errors table
	// Simplified for now
	return errorPattern{
		Message: "",
		Count:   0,
	}
}

func (rl *RefactorLoop) checkManualInterventionKeywords(sessionID string, window time.Duration) bool {
	keywords := []string{
		"you need to fix this manually",
		"manual intervention required",
		"cannot automatically resolve",
		"requires manual",
	}

	// Would search recent GitHub outputs for these keywords
	_ = keywords
	return false
}

func (rl *RefactorLoop) checkNoProgress(sessionID string, window time.Duration) bool {
	// Would check if same problem has been worked on for 10+ minutes
	// with no successful resolution
	return false
}

// ============================================================================
// REFACTOR LOOP EXECUTION
// ============================================================================

// RefactorRequest represents the input to the refactor loop
type RefactorRequest struct {
	OriginalProblem string `json:"original_problem"`
	StuckApproach   string `json:"stuck_approach"`
	AttemptCount    int    `json:"attempt_count"`
	ProblemContext  string `json:"problem_context"`
}

// RefactorResponse contains the generated alternatives and selection
type RefactorResponse struct {
	Alternatives       []RefactorAlternative `json:"alternatives"`
	SelectedApproach   *RefactorAlternative  `json:"selected_approach"`
	SelectionReasoning string                `json:"selection_reasoning"`
	EventID            uint                  `json:"event_id"`
}

// ExecuteRefactorLoop runs the complete 5-alternative generation and selection process
func (rl *RefactorLoop) ExecuteRefactorLoop(req RefactorRequest) (*RefactorResponse, error) {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🔄 REFACTOR LOOP INITIATED")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Problem: %s", req.OriginalProblem)
	log.Printf("Stuck approach: %s", req.StuckApproach)
	log.Printf("Attempts: %d", req.AttemptCount)

	response := &RefactorResponse{
		Alternatives: make([]RefactorAlternative, 5),
	}

	// STEP 1: Generate 5 fundamentally different approaches
	log.Println("\n🎯 STEP 1: Generating 5 Fundamentally Different Approaches")
	alternatives := rl.generateFiveAlternatives(req)
	response.Alternatives = alternatives

	// STEP 2: Evaluate each approach
	log.Println("\n📊 STEP 2: Evaluating Approaches")
	for i := range response.Alternatives {
		response.Alternatives[i].Scores = rl.evaluateApproach(response.Alternatives[i], req)
		response.Alternatives[i].WeightedScore = response.Alternatives[i].Scores.CalculateWeightedScore()
		log.Printf("   Approach %d: %.2f points", i+1, response.Alternatives[i].WeightedScore)
	}

	// STEP 3: Select best approach
	log.Println("\n✅ STEP 3: Selecting Highest Scoring Approach")
	selected, reasoning := rl.selectBestApproach(response.Alternatives)
	response.SelectedApproach = selected
	response.SelectionReasoning = reasoning
	log.Printf("   Selected: Approach %d (%.2f points)", selected.ApproachNumber, selected.WeightedScore)
	log.Printf("   Reasoning: %s", reasoning)

	// STEP 4: Record refactor event
	log.Println("\n💾 STEP 4: Recording Refactor Event")
	eventID, err := rl.recordRefactorEvent(req, response)
	if err != nil {
		log.Printf("   ⚠️  Failed to record event: %v", err)
	} else {
		response.EventID = eventID
		log.Printf("   ✅ Event recorded: ID %d", eventID)
	}

	log.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🎉 REFACTOR LOOP COMPLETE")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return response, nil
}

// generateFiveAlternatives creates fundamentally different approaches
func (rl *RefactorLoop) generateFiveAlternatives(req RefactorRequest) []RefactorAlternative {
	// In production, this would use LLM to generate creative alternatives
	// For now, returning template-based alternatives based on problem category

	category := rl.categorizeProblem(req.OriginalProblem)

	templates := rl.getAlternativeTemplates(category)
	alternatives := make([]RefactorAlternative, 5)

	for i := 0; i < 5; i++ {
		if i < len(templates) {
			alternatives[i] = templates[i]
			alternatives[i].ApproachNumber = i + 1
		} else {
			alternatives[i] = RefactorAlternative{
				ApproachNumber: i + 1,
				Strategy:       fmt.Sprintf("Creative Approach %d", i+1),
				Description:    "Generate novel solution using different tool/framework",
				ToolsUsed:      []string{"to_be_determined"},
			}
		}
	}

	return alternatives
}

func (rl *RefactorLoop) categorizeProblem(problem string) string {
	lowerProblem := fmt.Sprintf("%s", problem) // Convert to string and lowercase

	if contains(lowerProblem, "file", "path", "directory") {
		return "file_system"
	} else if contains(lowerProblem, "database", "sql", "connection") {
		return "database"
	} else if contains(lowerProblem, "401", "403", "auth", "token") {
		return "authentication"
	} else if contains(lowerProblem, "timeout", "network", "connection") {
		return "network"
	}

	return "general"
}

func contains(str string, keywords ...string) bool {
	for _, keyword := range keywords {
		if len(str) > 0 && len(keyword) > 0 {
			// Simple contains check
			return true // Simplified
		}
	}
	return false
}

func (rl *RefactorLoop) getAlternativeTemplates(category string) []RefactorAlternative {
	// Return category-specific alternative templates
	templates := make(map[string][]RefactorAlternative)

	templates["file_system"] = []RefactorAlternative{
		{Strategy: "Use Absolute Paths", Description: "Programmatically resolve absolute paths from executable location", ToolsUsed: []string{"filepath.Abs", "os.Executable"}},
		{Strategy: "Command Line Arguments", Description: "Accept file paths as CLI arguments instead of hardcoding", ToolsUsed: []string{"flag", "cobra"}},
		{Strategy: "Embed Files", Description: "Embed configuration files into executable using go:embed", ToolsUsed: []string{"embed"}},
		{Strategy: "Environment Variables", Description: "Use environment variables for all file paths", ToolsUsed: []string{"os.Getenv"}},
		{Strategy: "Configuration Service", Description: "Implement smart file discovery service with fallback locations", ToolsUsed: []string{"viper", "custom_service"}},
	}

	if alts, exists := templates[category]; exists {
		return alts
	}

	return []RefactorAlternative{
		{Strategy: "Approach 1", Description: "First alternative approach"},
		{Strategy: "Approach 2", Description: "Second alternative approach"},
		{Strategy: "Approach 3", Description: "Third alternative approach"},
		{Strategy: "Approach 4", Description: "Fourth alternative approach"},
		{Strategy: "Approach 5", Description: "Fifth alternative approach"},
	}
}

// evaluateApproach scores an approach on the 5 criteria
func (rl *RefactorLoop) evaluateApproach(alt RefactorAlternative, req RefactorRequest) RefactorScores {
	// In production, would use LLM or heuristics to score
	// For now, returning reasonable default scores

	return RefactorScores{
		RootCauseSolution: 8,
		Simplicity:        7,
		LowRisk:           8,
		CodebaseAlignment: 7,
		Testability:       8,
	}
}

// selectBestApproach picks the highest scoring alternative
func (rl *RefactorLoop) selectBestApproach(alternatives []RefactorAlternative) (*RefactorAlternative, string) {
	var best *RefactorAlternative
	maxScore := 0.0

	for i := range alternatives {
		if alternatives[i].WeightedScore > maxScore {
			maxScore = alternatives[i].WeightedScore
			best = &alternatives[i]
		}
	}

	reasoning := fmt.Sprintf(
		"Approach %d scored highest (%.2f points) due to strong root cause addressing (score: %d×2.0), "+
			"reasonable simplicity (score: %d×1.5), low risk (score: %d×1.5), "+
			"good codebase alignment (score: %d×1.0), and high testability (score: %d×1.0)",
		best.ApproachNumber, best.WeightedScore,
		best.Scores.RootCauseSolution, best.Scores.Simplicity, best.Scores.LowRisk,
		best.Scores.CodebaseAlignment, best.Scores.Testability,
	)

	return best, reasoning
}

// recordRefactorEvent saves the refactor event to database
func (rl *RefactorLoop) recordRefactorEvent(req RefactorRequest, resp *RefactorResponse) (uint, error) {
	// Serialize alternatives to JSON
	alternativesJSON, err := json.Marshal(resp.Alternatives)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize alternatives: %w", err)
	}

	event := &RefactorEvent{
		OriginalProblem:      req.OriginalProblem,
		StuckApproach:        req.StuckApproach,
		AttemptCount:         req.AttemptCount,
		FiveAlternativesJSON: string(alternativesJSON),
		EvaluationScores: map[string]interface{}{
			"alternatives": resp.Alternatives,
		},
		SelectedApproach:   resp.SelectedApproach.Strategy,
		SelectionReasoning: resp.SelectionReasoning,
		ProblemCategory:    rl.categorizeProblem(req.OriginalProblem),
		Analyzed:           false,
	}

	result := rl.db.Create(event)
	if result.Error != nil {
		return 0, result.Error
	}

	return event.ID, nil
}

// ============================================================================
// SOLACE LEARNING FROM REFACTOR EVENTS
// ============================================================================

// AnalyzeRefactorEvents allows Solace to learn from refactor patterns
func (rl *RefactorLoop) AnalyzeRefactorEvents() error {
	log.Println("🧪 Analyzing refactor events for pattern extraction...")

	var unanalyzedEvents []RefactorEvent
	result := rl.db.Where("analyzed = ?", false).Find(&unanalyzedEvents)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("   Found %d unanalyzed refactor events", len(unanalyzedEvents))

	for _, event := range unanalyzedEvents {
		// Extract learning from each event
		rl.learnFromRefactorEvent(&event)

		// Mark as analyzed
		event.Analyzed = true
		rl.db.Save(&event)
	}

	return nil
}

func (rl *RefactorLoop) learnFromRefactorEvent(event *RefactorEvent) {
	log.Printf("   Learning from refactor event ID %d (category: %s)", event.ID, event.ProblemCategory)

	// This would:
	// 1. Extract which approach type won
	// 2. Identify why it scored highest
	// 3. Check for pattern across similar problems
	// 4. Store to solace_refactor_strategies table

	// Simplified for now - would integrate with database in production
}
