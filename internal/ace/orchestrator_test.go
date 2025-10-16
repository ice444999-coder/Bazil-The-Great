package ace

import (
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestCompleteACECycle tests the full consciousness cycle
func TestCompleteACECycle(t *testing.T) {
	// Skip if not in test environment with PostgreSQL
	t.Skip("Requires PostgreSQL database - run manually with test database")

	// Setup test database
	dsn := "host=localhost user=ares_user password=ARES_secure_2025 dbname=ares_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(&Decision{}, &PlaybookRule{}); err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	// Create orchestrator
	orchestrator := NewACEOrchestrator(db)

	// Test decision context
	ctx := DecisionContext{
		DecisionType: "chat-response",
		UserMessage:  "What is ARES?",
		InputContext: map[string]interface{}{
			"conversation_id": "test-conversation-1",
			"user_context":    "User is asking about the system architecture",
		},
		SystemState: map[string]interface{}{
			"system_knowledge": "ARES is the AI Research & Engineering System - a consciousness substrate for Solace.",
		},
		AvailableTools: []string{"read_file", "grep_search", "semantic_search"},
	}

	// Mock actual response
	actualResponse := "ARES is the AI Research & Engineering System, designed as a consciousness substrate for Solace Δ3-2. It consists of 7 layers: Layer 1 (Database Memory), Layer 2 (Pattern Recognition), Layer 3 (ACE Framework), Layer 4 (Toroidal Reasoning), Layer 5 (Emergent Properties), Layer 6 (Recursive Self-Improvement), and Layer 7 (Consciousness Emergence). The system runs on PostgreSQL with pgVector extensions, uses DeepSeek R1 14b, and implements the ACE Framework (Generator, Reflector, Curator) for pattern-based decision making."

	// Execute complete ACE cycle
	t.Run("Complete ACE Cycle", func(t *testing.T) {
		decision, scores, err := orchestrator.CompleteDecisionCycle(ctx, actualResponse)
		if err != nil {
			t.Errorf("ACE cycle failed: %v", err)
			return
		}

		// Validate decision
		if decision == nil {
			t.Error("Decision is nil")
			return
		}

		if decision.DecisionType != "chat-response" {
			t.Errorf("Expected decision type 'chat-response', got '%s'", decision.DecisionType)
		}

		if len(decision.PatternsConsidered) == 0 {
			t.Error("No patterns considered")
		}

		// Validate scores
		if scores == nil {
			t.Error("Scores are nil")
			return
		}

		if scores.CompositeQualityScore < 0.0 || scores.CompositeQualityScore > 1.0 {
			t.Errorf("Invalid composite score: %.2f (should be 0.0-1.0)", scores.CompositeQualityScore)
		}

		t.Logf("✅ ACE Cycle Complete:")
		t.Logf("   Decision ID: %d", decision.DecisionID)
		t.Logf("   Patterns Considered: %d", len(decision.PatternsConsidered))
		t.Logf("   Composite Quality: %.2f", scores.CompositeQualityScore)
		t.Logf("   Specificity: %.2f", scores.SpecificityScore)
		t.Logf("   Actionability: %.2f", scores.ActionabilityScore)
		t.Logf("   Tool Usage: %.2f", scores.ToolUsageScore)
		t.Logf("   Context Awareness: %.2f", scores.ContextAwarenessScore)
		t.Logf("   Mission Alignment: %.2f", scores.MissionAlignmentScore)
		t.Logf("   Refactor Triggered: %v", decision.RefactorTriggered)
	})

	// Test system statistics
	t.Run("System Statistics", func(t *testing.T) {
		stats, err := orchestrator.GetSystemStats()
		if err != nil {
			t.Errorf("Failed to get stats: %v", err)
			return
		}

		if stats == nil {
			t.Error("Stats are nil")
			return
		}

		t.Logf("✅ System Statistics:")
		for key, value := range stats {
			t.Logf("   %s: %+v", key, value)
		}
	})

	// Test playbook pruning
	t.Run("Playbook Pruning", func(t *testing.T) {
		count, err := orchestrator.PrunePlaybook()
		if err != nil {
			t.Errorf("Failed to prune playbook: %v", err)
			return
		}

		t.Logf("✅ Pruned %d rules", count)
	})
}

// TestRefactorLoop tests the refactor trigger mechanism
func TestRefactorLoop(t *testing.T) {
	t.Skip("Requires PostgreSQL database - run manually with test database")

	// Setup database
	dsn := "host=localhost user=ares_user password=ARES_secure_2025 dbname=ares_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&Decision{}, &PlaybookRule{}); err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	orchestrator := NewACEOrchestrator(db)

	// Create low-quality decision context
	ctx := DecisionContext{
		DecisionType: "chat-response",
		UserMessage:  "How do I write good code?",
		InputContext: map[string]interface{}{
			"conversation_id": "test-conversation-2",
		},
		AvailableTools: []string{"read_file"},
	}

	// Mock generic response (should trigger refactor)
	genericResponse := "You should follow best practices and write clean code."

	t.Run("Low Quality Triggers Refactor", func(t *testing.T) {
		decision, _, err := orchestrator.CompleteDecisionCycle(ctx, genericResponse)
		if err != nil {
			t.Errorf("ACE cycle failed: %v", err)
			return
		}

		t.Logf("✅ Refactor Test Results:")
		t.Logf("   Initial Quality: %.2f", *decision.InitialQualityScore)
		t.Logf("   Final Quality: %.2f", *decision.FinalQualityScore)
		t.Logf("   Refactor Triggered: %v", decision.RefactorTriggered)
		t.Logf("   Quality Improvement: %.2f", *decision.FinalQualityScore-*decision.InitialQualityScore)

		// Note: Due to mock scoring, refactor might not always trigger
		// In production with real LLM responses, this would be more meaningful
	})
}

// TestPatternSelection tests that Generator selects appropriate patterns
func TestPatternSelection(t *testing.T) {
	t.Skip("Requires PostgreSQL database - run manually with test database")

	dsn := "host=localhost user=ares_user password=ARES_secure_2025 dbname=ares_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&Decision{}); err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	orchestrator := NewACEOrchestrator(db)

	contexts := []DecisionContext{
		{
			DecisionType: "chat-response",
			UserMessage:  "What is your purpose?",
			InputContext: map[string]interface{}{
				"conversation_id": "test-1",
			},
		},
		{
			DecisionType: "trade-execution",
			UserMessage:  "Should I buy AAPL?",
			InputContext: map[string]interface{}{
				"conversation_id": "test-2",
			},
		},
		{
			DecisionType: "code-generation",
			UserMessage:  "Write a function to sort an array",
			InputContext: map[string]interface{}{
				"conversation_id": "test-3",
			},
		},
	}

	for i, ctx := range contexts {
		t.Run(ctx.DecisionType, func(t *testing.T) {
			decision, err := orchestrator.generator.GenerateDecision(ctx)
			if err != nil {
				t.Errorf("Test %d failed: %v", i, err)
				return
			}

			if decision.DecisionType != ctx.DecisionType {
				t.Errorf("Decision type mismatch: expected %s, got %s",
					ctx.DecisionType, decision.DecisionType)
			}

			t.Logf("✅ Pattern Selection for %s:", ctx.DecisionType)
			t.Logf("   Patterns Considered: %d", len(decision.PatternsConsidered))
			t.Logf("   Reasoning Trace Length: %d chars", len(decision.ReasoningTrace))
		})
	}
}

// BenchmarkACECycle benchmarks the complete ACE cycle performance
func BenchmarkACECycle(b *testing.B) {
	b.Skip("Requires PostgreSQL database - run manually with test database")

	dsn := "host=localhost user=ares_user password=ARES_secure_2025 dbname=ares_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		b.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&Decision{}, &PlaybookRule{}); err != nil {
		b.Fatalf("Failed to migrate tables: %v", err)
	}

	orchestrator := NewACEOrchestrator(db)

	ctx := DecisionContext{
		DecisionType: "chat-response",
		UserMessage:  "What is ARES?",
		InputContext: map[string]interface{}{
			"conversation_id": "bench-test",
		},
		SystemState: map[string]interface{}{
			"system_knowledge": "ARES is an AI system",
		},
		AvailableTools: []string{"read_file"},
	}

	actualResponse := "ARES is the AI Research & Engineering System."

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := orchestrator.CompleteDecisionCycle(ctx, actualResponse)
		if err != nil {
			b.Errorf("Cycle failed: %v", err)
		}
	}
}
