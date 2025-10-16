package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Cognitive Pattern model
type CognitivePattern struct {
	PatternID         uint    `gorm:"primaryKey;column:pattern_id"`
	PatternName       string  `gorm:"unique;not null;size:255"`
	PatternCategory   string  `gorm:"size:100;not null"`
	Description       string  `gorm:"type:text"`
	TriggerConditions string  `gorm:"type:text"`
	ExampleInput      string  `gorm:"type:text"`
	ExampleOutput     string  `gorm:"type:text"`
	ExampleReasoning  string  `gorm:"type:text"`
	ConfidenceScore   float64 `gorm:"type:decimal(5,4);default:0.5000"`
	UsageCount        int     `gorm:"default:0"`
	SuccessCount      int     `gorm:"default:0"`
	CreatedAt         int64   `gorm:"autoCreateTime"`
	LastUsed          *int64
}

// Playbook Rule model
type PlaybookRule struct {
	RuleID               uint    `gorm:"primaryKey;column:rule_id"`
	RuleName             string  `gorm:"size:255;not null"`
	RuleCategory         string  `gorm:"size:100;not null"`
	TriggerConditions    string  `gorm:"type:text"`
	ApplicationExample   string  `gorm:"type:text"`
	ConfidenceScore      float64 `gorm:"type:decimal(5,4);default:0.5000"`
	UsageCount           int     `gorm:"default:0"`
	SuccessCount         int     `gorm:"default:0"`
	SourcePatternIDs     string  `gorm:"type:text"` // JSON array stored as text
	ParentRuleID         *uint
	CreatedAt            int64 `gorm:"autoCreateTime"`
	LastUsed             *int64
	LastSuccessRate      *float64 `gorm:"type:decimal(5,4)"`
	ConsecutiveLowChecks int      `gorm:"default:0"`
}

// Decision model
type Decision struct {
	DecisionID          uint     `gorm:"primaryKey;column:decision_id"`
	DecisionType        string   `gorm:"size:100;not null"`
	InputContext        string   `gorm:"type:jsonb"`
	PatternsConsidered  string   `gorm:"type:text"` // JSON array stored as text
	RulesApplied        string   `gorm:"type:text"` // JSON array stored as text
	ReasoningTrace      string   `gorm:"type:text"`
	DecisionOutput      string   `gorm:"type:jsonb"`
	ConfidenceLevel     float64  `gorm:"type:decimal(5,4)"`
	InitialQualityScore *float64 `gorm:"type:decimal(5,4)"`
	RefactorTriggered   bool     `gorm:"default:false"`
	FinalQualityScore   *float64 `gorm:"type:decimal(5,4)"`
	ToolsInvoked        string   `gorm:"type:text"` // JSON array stored as text
	CreatedAt           int64    `gorm:"autoCreateTime"`
}

// Quality Score model
type QualityScore struct {
	ScoreID               uint    `gorm:"primaryKey;column:score_id"`
	DecisionID            uint    `gorm:"not null"`
	SpecificityScore      float64 `gorm:"type:decimal(5,4)"`
	ActionabilityScore    float64 `gorm:"type:decimal(5,4)"`
	ToolUsageScore        float64 `gorm:"type:decimal(5,4)"`
	ContextAwarenessScore float64 `gorm:"type:decimal(5,4)"`
	MissionAlignmentScore float64 `gorm:"type:decimal(5,4)"`
	CompositeQualityScore float64 `gorm:"type:decimal(5,4)"`
	CreatedAt             int64   `gorm:"autoCreateTime"`
}

func main() {
	// Load .env
	envPaths := []string{".env", "../.env", "../../.env", "c:\\ARES_Workspace\\ARES_API\\.env"}
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("✅ Loaded .env from: %s", path)
			break
		}
	}

	// Connect to database
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🔧 ACE FRAMEWORK DATABASE MIGRATION")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	log.Printf("✅ Connected to database: %s@%s/%s", user, host, dbname)

	// Auto-migrate ACE tables
	log.Println("\n📋 Creating ACE Framework tables...")

	err = db.AutoMigrate(
		&CognitivePattern{},
		&PlaybookRule{},
		&Decision{},
		&QualityScore{},
	)
	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ All tables created successfully!")

	// Verify
	log.Println("\n📊 Verifying table creation...")

	var count int64

	db.Table("cognitive_patterns").Count(&count)
	log.Printf("   ✓ cognitive_patterns: %d rows", count)

	db.Table("playbook_rules").Count(&count)
	log.Printf("   ✓ playbook_rules: %d rows", count)

	db.Table("decisions").Count(&count)
	log.Printf("   ✓ decisions: %d rows", count)

	db.Table("quality_scores").Count(&count)
	log.Printf("   ✓ quality_scores: %d rows", count)

	log.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅ MIGRATION COMPLETE")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
