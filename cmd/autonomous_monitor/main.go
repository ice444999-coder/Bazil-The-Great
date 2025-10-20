package main

// ============================================================================
// CRYSTAL #27: AUTONOMOUS MONITOR
// ============================================================================
// Purpose: Hourly monitoring that detects when optimizations are needed
// Triggers: Agent swarm (ARCHITECT → FORGE → SENTINEL → queue)
// Runtime: Windows Task Scheduler (hourly)
// ============================================================================

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type MetricThreshold struct {
	P95LatencyMs float64 `json:"p95_latency_ms"`
	CacheHitRate float64 `json:"cache_hit_rate"`
	ErrorRate    float64 `json:"error_rate"`
}

type CurrentMetrics struct {
	P95LatencyMs float64
	CacheHitRate float64
	ErrorRate    float64
	MeasuredAt   time.Time
}

type ImprovementTemplate struct {
	ID               int
	Name             string
	Category         string
	TriggerCondition string
	TriggerThreshold map[string]interface{}
	EstimatedImpact  string
	RiskLevel        string
}

func main() {
	log.SetPrefix("[AUTONOMOUS_MONITOR] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("🔮 Starting autonomous improvement detection...")

	// Connect to database
	db, err := connectDB()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()

	// Step 1: Get current metrics
	metrics, err := getCurrentMetrics(db)
	if err != nil {
		log.Fatalf("❌ Failed to fetch metrics: %v", err)
	}

	log.Printf("📊 Current metrics: P95=%.0fms, CacheHitRate=%.2f%%, ErrorRate=%.2f%%",
		metrics.P95LatencyMs, metrics.CacheHitRate*100, metrics.ErrorRate*100)

	// Step 2: Check thresholds (from Crystal #26)
	thresholds := MetricThreshold{
		P95LatencyMs: 100.0, // From Crystal #26: target <100ms
		CacheHitRate: 0.30,  // From Crystal #26: should be >30%
		ErrorRate:    0.01,  // Target: <1% error rate
	}

	violations := detectViolations(metrics, thresholds)

	if len(violations) == 0 {
		log.Println("✅ All metrics within thresholds - no action needed")
		return
	}

	log.Printf("⚠️ Detected %d threshold violations:", len(violations))
	for _, v := range violations {
		log.Printf("   - %s", v)
	}

	// Step 3: Query improvement templates
	templates, err := getEnabledTemplates(db)
	if err != nil {
		log.Fatalf("❌ Failed to fetch templates: %v", err)
	}

	log.Printf("📋 Checking %d improvement templates...", len(templates))

	// Step 4: Evaluate each template's trigger condition
	triggeredTemplates := []ImprovementTemplate{}
	for _, template := range templates {
		triggered, err := evaluateTrigger(db, template, metrics)
		if err != nil {
			log.Printf("⚠️ Template '%s' trigger evaluation failed: %v", template.Name, err)
			continue
		}

		if triggered {
			log.Printf("✅ Template '%s' triggered (category: %s, risk: %s)",
				template.Name, template.Category, template.RiskLevel)
			triggeredTemplates = append(triggeredTemplates, template)
		}
	}

	if len(triggeredTemplates) == 0 {
		log.Println("ℹ️ No templates triggered - manual optimization may be needed")
		return
	}

	// Step 5: Create improvement queue entries
	for _, template := range triggeredTemplates {
		err := queueImprovement(db, template, metrics, violations)
		if err != nil {
			log.Printf("❌ Failed to queue improvement from template '%s': %v", template.Name, err)
			continue
		}

		log.Printf("📝 Queued improvement: %s (estimated: %s)", template.Name, template.EstimatedImpact)
	}

	log.Println("🎉 Autonomous monitor complete")
}

func connectDB() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5433"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "ARESISWAKING"),
		getEnv("DB_NAME", "ares_pgvector"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func getCurrentMetrics(db *sql.DB) (*CurrentMetrics, error) {
	query := `
		SELECT 
			COALESCE(p95_latency_ms, 0) as p95_latency_ms,
			COALESCE(cache_hit_rate, 0) as cache_hit_rate,
			COALESCE(error_rate, 0) as error_rate,
			measured_at
		FROM memory_system_metrics
		ORDER BY measured_at DESC
		LIMIT 1
	`

	var metrics CurrentMetrics
	err := db.QueryRow(query).Scan(
		&metrics.P95LatencyMs,
		&metrics.CacheHitRate,
		&metrics.ErrorRate,
		&metrics.MeasuredAt,
	)

	if err == sql.ErrNoRows {
		// No metrics yet - use defaults
		log.Println("⚠️ No metrics in database - using defaults")
		return &CurrentMetrics{
			P95LatencyMs: 50.0,
			CacheHitRate: 0.80,
			ErrorRate:    0.001,
			MeasuredAt:   time.Now(),
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return &metrics, nil
}

func detectViolations(metrics *CurrentMetrics, thresholds MetricThreshold) []string {
	violations := []string{}

	if metrics.P95LatencyMs > thresholds.P95LatencyMs {
		violations = append(violations, fmt.Sprintf(
			"P95 latency too high: %.0fms > %.0fms threshold",
			metrics.P95LatencyMs, thresholds.P95LatencyMs,
		))
	}

	if metrics.CacheHitRate < thresholds.CacheHitRate {
		violations = append(violations, fmt.Sprintf(
			"Cache hit rate too low: %.2f%% < %.2f%% threshold",
			metrics.CacheHitRate*100, thresholds.CacheHitRate*100,
		))
	}

	if metrics.ErrorRate > thresholds.ErrorRate {
		violations = append(violations, fmt.Sprintf(
			"Error rate too high: %.2f%% > %.2f%% threshold",
			metrics.ErrorRate*100, thresholds.ErrorRate*100,
		))
	}

	return violations
}

func getEnabledTemplates(db *sql.DB) ([]ImprovementTemplate, error) {
	query := `
		SELECT 
			id, name, category, trigger_condition, 
			trigger_threshold, estimated_impact, risk_level
		FROM improvement_templates
		WHERE enabled = TRUE
		ORDER BY risk_level ASC, id ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := []ImprovementTemplate{}
	for rows.Next() {
		var t ImprovementTemplate
		var thresholdJSON []byte

		err := rows.Scan(
			&t.ID, &t.Name, &t.Category, &t.TriggerCondition,
			&thresholdJSON, &t.EstimatedImpact, &t.RiskLevel,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON threshold
		if len(thresholdJSON) > 0 {
			if err := json.Unmarshal(thresholdJSON, &t.TriggerThreshold); err != nil {
				log.Printf("⚠️ Failed to parse threshold for template %s: %v", t.Name, err)
			}
		}

		templates = append(templates, t)
	}

	return templates, nil
}

func evaluateTrigger(db *sql.DB, template ImprovementTemplate, metrics *CurrentMetrics) (bool, error) {
	// Special handling for cache hit rate template
	if template.Name == "add_redis_cache_layer" {
		threshold, ok := template.TriggerThreshold["cache_hit_rate"].(float64)
		if !ok {
			threshold = 0.30
		}
		return metrics.CacheHitRate < threshold, nil
	}

	// For SQL-based triggers, execute the query
	if template.TriggerCondition != "" {
		var result bool
		err := db.QueryRow(template.TriggerCondition).Scan(&result)
		if err == sql.ErrNoRows {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return result, nil
	}

	return false, nil
}

func queueImprovement(db *sql.DB, template ImprovementTemplate, metrics *CurrentMetrics, violations []string) error {
	description := fmt.Sprintf(
		"Auto-detected via template '%s'.\n\nViolations:\n%s\n\nCurrent metrics: P95=%.0fms, Cache=%.2f%%",
		template.Name,
		formatViolations(violations),
		metrics.P95LatencyMs,
		metrics.CacheHitRate*100,
	)

	// Schedule for next 10pm Brisbane time (AEST = UTC+10)
	now := time.Now().UTC()
	scheduledTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		12, 0, 0, 0, time.UTC, // 10pm Brisbane = 12:00 UTC (during standard time)
	)
	if scheduledTime.Before(now) {
		scheduledTime = scheduledTime.Add(24 * time.Hour)
	}

	query := `
		INSERT INTO improvement_queue (
			created_by, title, description, 
			scheduled_for, status, risk_level,
			estimated_speedup_percent, requires_approval
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id
	`

	var id int
	err := db.QueryRow(
		query,
		"autonomous_monitor",
		template.Name,
		description,
		scheduledTime,
		"PENDING",
		template.RiskLevel,
		extractEstimatedSpeedup(template.EstimatedImpact),
		template.RiskLevel != "LOW", // Auto-approve LOW risk only
	).Scan(&id)

	if err != nil {
		return err
	}

	log.Printf("✅ Created improvement queue entry ID=%d", id)
	return nil
}

func formatViolations(violations []string) string {
	result := ""
	for _, v := range violations {
		result += "- " + v + "\n"
	}
	return result
}

func extractEstimatedSpeedup(impact string) int {
	// Parse strings like "30-50% speedup" → return 40 (midpoint)
	// Simple parsing for now
	if impact == "" {
		return 0
	}

	// Default estimates
	if impact == "10x faster reads, 70%+ cache hit rate" {
		return 70
	}
	if impact == "30-50% speedup for affected queries" {
		return 40
	}
	if impact == "20-40% speedup" {
		return 30
	}

	return 25 // Default
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
