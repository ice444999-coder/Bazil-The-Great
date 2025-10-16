package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
	log.Println("📚 COGNITIVE PATTERNS LOADER")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	log.Printf("✅ Connected to database: %s@%s/%s\n", user, host, dbname)

	// Parse patterns from Python file
	patterns, err := parsePatternsPython("internal/agent/cognitive_patterns.py")
	if err != nil {
		log.Fatalf("❌ Failed to parse patterns: %v", err)
	}

	log.Printf("📖 Parsed %d patterns from Python file\n", len(patterns))

	// Insert patterns
	log.Println("\n💾 Inserting patterns into database...")

	inserted := 0
	skipped := 0

	for _, pattern := range patterns {
		// Check if pattern already exists
		var existing CognitivePattern
		result := db.Where("pattern_name = ?", pattern.PatternName).First(&existing)

		if result.Error == nil {
			// Pattern exists, skip
			skipped++
			continue
		}

		// Insert new pattern
		if err := db.Create(&pattern).Error; err != nil {
			log.Printf("⚠️  Failed to insert '%s': %v", pattern.PatternName, err)
			continue
		}

		inserted++
		if inserted%10 == 0 {
			log.Printf("   Progress: %d/%d patterns inserted...", inserted, len(patterns))
		}
	}

	log.Printf("\n✅ Insertion complete:")
	log.Printf("   Inserted: %d new patterns", inserted)
	log.Printf("   Skipped:  %d existing patterns", skipped)

	// Verify by category
	log.Println("\n📊 Patterns by category:")

	type CategoryCount struct {
		Category string
		Count    int64
	}

	var categories []CategoryCount
	db.Raw(`
		SELECT pattern_category as category, COUNT(*) as count 
		FROM cognitive_patterns 
		GROUP BY pattern_category 
		ORDER BY count DESC
	`).Scan(&categories)

	totalPatterns := int64(0)
	for _, cat := range categories {
		log.Printf("   %-25s %3d patterns", cat.Category, cat.Count)
		totalPatterns += cat.Count
	}

	log.Printf("\n   %-25s %3d patterns\n", "TOTAL", totalPatterns)

	// Calculate average confidence
	var avgConfidence float64
	db.Raw("SELECT AVG(confidence_score) FROM cognitive_patterns WHERE confidence_score > 0").Scan(&avgConfidence)
	log.Printf("📊 Average confidence: %.3f\n", avgConfidence)

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅ PATTERN LOADING COMPLETE")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func parsePatternsPython(filePath string) ([]CognitivePattern, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var patterns []CognitivePattern
	scanner := bufio.NewScanner(file)

	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var currentPattern *CognitivePattern
	var inPattern bool
	var currentField string
	var fieldValue strings.Builder

	// Regex patterns
	patternStart := regexp.MustCompile(`^\s*{\s*$`)
	patternEnd := regexp.MustCompile(`^\s*},?\s*$`)
	fieldPattern := regexp.MustCompile(`^\s*"(pattern_name|pattern_category|description|trigger_conditions|example_input|example_output|example_reasoning|confidence_score)":\s*(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Start of a new pattern
		if patternStart.MatchString(line) {
			if currentPattern != nil {
				// Save previous pattern
				patterns = append(patterns, *currentPattern)
			}
			currentPattern = &CognitivePattern{}
			inPattern = true
			continue
		}

		// End of pattern
		if patternEnd.MatchString(line) && inPattern {
			// Save field value
			if currentField != "" && currentPattern != nil {
				setField(currentPattern, currentField, fieldValue.String())
				fieldValue.Reset()
				currentField = ""
			}
			// Don't save yet, wait for next pattern or EOF
			continue
		}

		if !inPattern {
			continue
		}

		// Check for field start
		if matches := fieldPattern.FindStringSubmatch(line); len(matches) > 0 {
			// Save previous field
			if currentField != "" && currentPattern != nil {
				setField(currentPattern, currentField, fieldValue.String())
				fieldValue.Reset()
			}

			currentField = matches[1]
			fieldValue.WriteString(matches[2])
		} else if currentField != "" {
			// Continuation of multiline field
			fieldValue.WriteString("\n")
			fieldValue.WriteString(line)
		}
	}

	// Save last pattern
	if currentPattern != nil {
		if currentField != "" {
			setField(currentPattern, currentField, fieldValue.String())
		}
		patterns = append(patterns, *currentPattern)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return patterns, nil
}

func setField(pattern *CognitivePattern, field, value string) {
	// Clean value - remove quotes and extra whitespace
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.TrimSuffix(value, ",")
	value = strings.TrimSpace(value)

	switch field {
	case "pattern_name":
		pattern.PatternName = value
	case "pattern_category":
		pattern.PatternCategory = value
	case "description":
		pattern.Description = value
	case "trigger_conditions":
		pattern.TriggerConditions = value
	case "example_input":
		pattern.ExampleInput = value
	case "example_output":
		pattern.ExampleOutput = value
	case "example_reasoning":
		pattern.ExampleReasoning = value
	case "confidence_score":
		fmt.Sscanf(value, "%f", &pattern.ConfidenceScore)
	}
}
