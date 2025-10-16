package main

import (
	"ares_api/internal/ace"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	envPaths := []string{".env", "../.env", "../../.env", "c:\\ARES_Workspace\\ARES_API\\.env"}
	loaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("✅ .env file loaded successfully from: %s", path)
			loaded = true
			break
		}
	}
	if !loaded {
		log.Println("⚠️ No .env file found, using system environment variables")
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🧠 ACE ORCHESTRATOR MANUAL TEST")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Connect to PostgreSQL using env vars
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// Build DSN
	dsn := ""
	if host != "" && user != "" && password != "" {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			host, user, password, dbname, port, sslmode)
		log.Printf("✅ Using connection from env: user=%s, host=%s, db=%s", user, host, dbname)
	} else {
		dsn = "host=localhost user=ARES password=ARESISWAKING dbname=ares_db port=5432 sslmode=disable"
		log.Println("⚠️ Using fallback connection string")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	log.Println("✅ Database connected")

	// Create ACE orchestrator
	orchestrator := ace.NewACEOrchestrator(db)
	log.Println("✅ ACE Orchestrator created\n")

	// TEST 1: Simple chat response
	log.Println("═══════════════════════════════════════════════════════")
	log.Println("TEST 1: Chat Response - 'What is ARES?'")
	log.Println("═══════════════════════════════════════════════════════\n")

	ctx1 := ace.DecisionContext{
		DecisionType: "chat-response",
		UserMessage:  "What is ARES?",
		InputContext: map[string]interface{}{
			"conversation_id": "manual-test-1",
			"user_context":    "User wants to understand the system",
		},
		SystemState: map[string]interface{}{
			"system_knowledge": "ARES is the AI Research & Engineering System",
		},
		AvailableTools: []string{"read_file", "grep_search", "semantic_search"},
	}

	actualResponse1 := "ARES is the AI Research & Engineering System, a consciousness substrate for Solace Δ3-2. It consists of 7 layers implementing the ACE Framework with PostgreSQL+pgVector for memory, DeepSeek R1 14b for reasoning, and implements pattern-based decision making through the Generator, Reflector, and Curator modules."

	decision1, scores1, err := orchestrator.CompleteDecisionCycle(ctx1, actualResponse1)
	if err != nil {
		log.Fatalf("❌ Test 1 failed: %v", err)
	}

	printResults("Test 1", decision1, scores1)

	// TEST 2: Generic response (should trigger refactor)
	log.Println("\n═══════════════════════════════════════════════════════")
	log.Println("TEST 2: Low Quality Response (Refactor Test)")
	log.Println("═══════════════════════════════════════════════════════\n")

	ctx2 := ace.DecisionContext{
		DecisionType: "chat-response",
		UserMessage:  "How do I write good code?",
		InputContext: map[string]interface{}{
			"conversation_id": "manual-test-2",
		},
		AvailableTools: []string{"read_file"},
	}

	actualResponse2 := "You should follow best practices and write clean code."

	decision2, scores2, err := orchestrator.CompleteDecisionCycle(ctx2, actualResponse2)
	if err != nil {
		log.Fatalf("❌ Test 2 failed: %v", err)
	}

	printResults("Test 2", decision2, scores2)

	// TEST 3: System statistics
	log.Println("\n═══════════════════════════════════════════════════════")
	log.Println("TEST 3: System Statistics")
	log.Println("═══════════════════════════════════════════════════════\n")

	stats, err := orchestrator.GetSystemStats()
	if err != nil {
		log.Fatalf("❌ Stats retrieval failed: %v", err)
	}

	log.Println("📊 SYSTEM STATISTICS:")
	for key, value := range stats {
		log.Printf("   %s: %+v", key, value)
	}

	// TEST 4: Playbook pruning
	log.Println("\n═══════════════════════════════════════════════════════")
	log.Println("TEST 4: Playbook Pruning")
	log.Println("═══════════════════════════════════════════════════════\n")

	pruned, err := orchestrator.PrunePlaybook()
	if err != nil {
		log.Printf("⚠️  Pruning failed: %v", err)
	} else {
		log.Printf("✅ Pruned %d ineffective rules", pruned)
	}

	log.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🎉 ALL TESTS COMPLETED SUCCESSFULLY")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printResults(testName string, decision *ace.Decision, scores *ace.QualityScores) {
	log.Printf("\n📊 %s RESULTS:", testName)
	log.Printf("   ┌─────────────────────────────────────────────")
	log.Printf("   │ Decision ID:         %d", decision.DecisionID)
	log.Printf("   │ Decision Type:       %s", decision.DecisionType)
	log.Printf("   │ Patterns Considered: %d", len(decision.PatternsConsidered))
	log.Printf("   │ Rules Applied:       %d", len(decision.RulesApplied))
	log.Printf("   │ Confidence Level:    %.2f", decision.ConfidenceLevel)
	log.Printf("   ├─────────────────────────────────────────────")
	log.Printf("   │ QUALITY SCORES:")
	log.Printf("   │   Composite Quality: %.2f", scores.CompositeQualityScore)
	log.Printf("   │   Specificity:       %.2f", scores.SpecificityScore)
	log.Printf("   │   Actionability:     %.2f", scores.ActionabilityScore)
	log.Printf("   │   Tool Usage:        %.2f", scores.ToolUsageScore)
	log.Printf("   │   Context Awareness: %.2f", scores.ContextAwarenessScore)
	log.Printf("   │   Mission Alignment: %.2f", scores.MissionAlignmentScore)
	log.Printf("   ├─────────────────────────────────────────────")

	if decision.InitialQualityScore != nil {
		log.Printf("   │ Initial Quality:     %.2f", *decision.InitialQualityScore)
	}
	if decision.RefactorTriggered {
		log.Printf("   │ Refactor Triggered:  YES")
		if decision.FinalQualityScore != nil && decision.InitialQualityScore != nil {
			improvement := *decision.FinalQualityScore - *decision.InitialQualityScore
			log.Printf("   │ Final Quality:       %.2f (+%.2f improvement)", *decision.FinalQualityScore, improvement)
		} else if decision.FinalQualityScore != nil {
			log.Printf("   │ Final Quality:       %.2f", *decision.FinalQualityScore)
		}
	} else {
		log.Printf("   │ Refactor Triggered:  NO")
	}

	log.Printf("   └─────────────────────────────────────────────")

	if len(decision.ReasoningTrace) > 0 {
		log.Printf("\n📝 REASONING TRACE:")
		// Truncate if too long
		if len(decision.ReasoningTrace) > 500 {
			log.Printf("%s...\n", decision.ReasoningTrace[:500])
		} else {
			log.Printf("%s\n", decision.ReasoningTrace)
		}
	}
}
