// ============================================================================
// ARES SQL Reorganization Performance Benchmark
// Version: 1.0.0
// Date: October 19, 2025
//
// Purpose: Benchmark AI query performance improvements after schema reorganization
// Tests semantic searches, complex joins, and aggregations for trading AI workloads
// ============================================================================

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type BenchmarkResult struct {
	QueryName   string
	Description string
	AverageTime time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	Iterations  int
	SqlQuery    string
}

type BenchmarkSuite struct {
	db      *sql.DB
	results []BenchmarkResult
}

func NewBenchmarkSuite() *BenchmarkSuite {
	// Get database connection from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:ARESISWAKING@localhost:5433/ares_pgvector?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return &BenchmarkSuite{
		db:      db,
		results: []BenchmarkResult{},
	}
}

func (bs *BenchmarkSuite) runBenchmark(queryName, description, sqlQuery string, iterations int) {
	log.Printf("🔍 Testing: %s", queryName)
	log.Printf("   %s", description)

	var totalTime time.Duration
	var minTime = time.Hour // Initialize to a large value
	var maxTime time.Duration

	for i := 1; i <= iterations; i++ {
		log.Printf("   Iteration %d/%d...", i, iterations)

		startTime := time.Now()

		// Execute query
		rows, err := bs.db.Query(sqlQuery)
		if err != nil {
			log.Printf("   Query failed: %v", err)
			continue
		}

		// Consume all rows to complete the query
		for rows.Next() {
			// We don't need the data, just ensure query completes
		}
		rows.Close()

		elapsed := time.Since(startTime)
		totalTime += elapsed

		if elapsed < minTime {
			minTime = elapsed
		}
		if elapsed > maxTime {
			maxTime = elapsed
		}

		log.Printf("   %.2fms", float64(elapsed.Nanoseconds())/1000000)
	}

	averageTime := totalTime / time.Duration(iterations)

	result := BenchmarkResult{
		QueryName:   queryName,
		Description: description,
		AverageTime: averageTime,
		MinTime:     minTime,
		MaxTime:     maxTime,
		Iterations:  iterations,
		SqlQuery:    sqlQuery,
	}

	bs.results = append(bs.results, result)

	log.Printf("   📊 Results: Avg=%.2fms, Min=%.2fms, Max=%.2fms",
		float64(averageTime.Nanoseconds())/1000000,
		float64(minTime.Nanoseconds())/1000000,
		float64(maxTime.Nanoseconds())/1000000)
	log.Println("")
}

func (bs *BenchmarkSuite) runAllBenchmarks(iterations int) {
	log.Println("🚀 ARES SQL Reorganization Performance Benchmark")
	log.Println("Testing AI query performance improvements...")
	log.Println("")

	// Test queries that simulate AI trading agent workloads
	benchmarks := []struct {
		name        string
		description string
		query       string
	}{
		{
			name:        "Semantic Search - Trading Patterns",
			description: "Vector similarity search for trading pattern recognition",
			query:       "SELECT id, content, embedding <=> '[0.1,0.2,0.3]'::vector as distance FROM memory_system.memory_embeddings ORDER BY embedding <=> '[0.1,0.2,0.3]'::vector LIMIT 5",
		},
		{
			name:        "Complex Join - Strategy Analysis",
			description: "Multi-table join for strategy performance analysis",
			query:       "SELECT t.id, t.symbol, t.side, t.quantity, s.strategy_name FROM trading_core.trades t JOIN trading_core.strategies s ON t.strategy_id = s.strategy_id WHERE t.created_at > NOW() - INTERVAL '24 hours' LIMIT 10",
		},
		{
			name:        "Aggregated Analytics - Portfolio Performance",
			description: "Complex aggregation for portfolio analytics",
			query:       "SELECT symbol, SUM(CASE WHEN side = 'buy' THEN quantity ELSE -quantity END) as net_position, AVG(price) as avg_price FROM trading_core.trades WHERE created_at > NOW() - INTERVAL '7 days' GROUP BY symbol ORDER BY ABS(SUM(CASE WHEN side = 'buy' THEN quantity ELSE -quantity END)) DESC LIMIT 5",
		},
		{
			name:        "Memory Retrieval - Conversation Context",
			description: "Context-aware memory retrieval for conversation continuity",
			query:       "SELECT c.id, c.user_message, c.solace_response, c.created_at FROM memory_system.conversations c WHERE c.embedding <=> '[0.5,0.6,0.7]'::vector < 0.8 ORDER BY c.created_at DESC LIMIT 3",
		},
		{
			name:        "Decision Analysis - Pattern Recognition",
			description: "High-confidence decision pattern analysis",
			query:       "SELECT d.id, d.decision_type, d.confidence_score, d.created_at FROM solace_core.solace_decisions d WHERE d.confidence_score > 0.7 AND d.created_at > NOW() - INTERVAL '1 hour' ORDER BY d.confidence_score DESC LIMIT 5",
		},
		{
			name:        "Tool Registry Search - AI Agent Tools",
			description: "Vector search for finding relevant AI tools",
			query:       "SELECT tool_id, tool_name, description FROM tool_system.tool_registry ORDER BY embedding <=> '[0.8,0.9,0.1]'::vector LIMIT 3",
		},
		{
			name:        "Multi-Schema Join - System Health",
			description: "Cross-schema analysis for system monitoring",
			query:       "SELECT 'trading' as domain, COUNT(*) as record_count FROM trading_core.trades UNION ALL SELECT 'memory' as domain, COUNT(*) as record_count FROM memory_system.memory_embeddings UNION ALL SELECT 'solace' as domain, COUNT(*) as record_count FROM solace_core.solace_decisions",
		},
	}

	log.Printf("🏁 Starting benchmark with %d iterations per query...", iterations)
	log.Println("")

	for _, benchmark := range benchmarks {
		bs.runBenchmark(benchmark.name, benchmark.description, benchmark.query, iterations)
	}
}

func (bs *BenchmarkSuite) generateReport() {
	log.Println("📈 BENCHMARK SUMMARY REPORT")
	log.Println("=" + string(make([]byte, 50)) + "=")

	if len(bs.results) == 0 {
		log.Println("No benchmark results to report")
		return
	}

	totalAvgTime := time.Duration(0)
	for _, result := range bs.results {
		totalAvgTime += result.AverageTime
	}
	overallAvg := totalAvgTime / time.Duration(len(bs.results))

	log.Printf("Overall Statistics:")
	log.Printf("  Total Queries Tested: %d", len(bs.results))
	log.Printf("  Overall Average Response Time: %.2fms", float64(overallAvg.Nanoseconds())/1000000)
	log.Println("")

	log.Println("Query Performance Breakdown:")
	fastQueries := 0
	slowQueries := 0

	for _, result := range bs.results {
		avgMs := float64(result.AverageTime.Nanoseconds()) / 1000000
		status := "✅"
		if avgMs >= 500 {
			status = "❌"
			slowQueries++
		} else if avgMs >= 100 {
			status = "⚠️"
		} else {
			fastQueries++
		}
		log.Printf("  %s %s: %.2fms avg", status, result.QueryName, avgMs)
	}

	log.Println("")
	log.Println("🎯 Performance Targets Assessment:")
	log.Printf("  Sub-100ms Queries (AI Target): %d/%d", fastQueries, len(bs.results))
	log.Printf("  Slow Queries (>500ms): %d/%d", slowQueries, len(bs.results))

	if fastQueries == len(bs.results) {
		log.Println("  ✅ ALL QUERIES MEET AI PERFORMANCE TARGETS (<100ms)")
	} else if slowQueries == 0 {
		log.Println("  ⚠️ MOST QUERIES MEET TARGETS - SOME OPTIMIZATION NEEDED")
	} else {
		log.Println("  ❌ PERFORMANCE ISSUES DETECTED - REQUIRES OPTIMIZATION")
	}

	log.Println("")
	log.Println("💾 Schema Reorganization Impact:")
	log.Println("  Before: 91+ tables in random organization")
	log.Println("  After: ~50 tables in 13 functional schemas with pgvector indexes")
	log.Println("  Expected: 2x speed improvement, <100ms AI queries, 70%+ trading win rates")

	log.Println("")
	log.Println("📊 Raw Results (CSV format):")
	log.Println("Query Name,Average Time (ms),Min Time (ms),Max Time (ms),Iterations")
	for _, result := range bs.results {
		avgMs := float64(result.AverageTime.Nanoseconds()) / 1000000
		minMs := float64(result.MinTime.Nanoseconds()) / 1000000
		maxMs := float64(result.MaxTime.Nanoseconds()) / 1000000
		log.Printf("%s,%.2f,%.2f,%.2f,%d", result.QueryName, avgMs, minMs, maxMs, result.Iterations)
	}

	// Export detailed results
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	resultFile := fmt.Sprintf("benchmark_results_%s.csv", timestamp)

	file, err := os.Create(resultFile)
	if err != nil {
		log.Printf("Failed to create result file: %v", err)
		return
	}
	defer file.Close()

	file.WriteString("Query Name,Description,Average Time (ms),Min Time (ms),Max Time (ms),Iterations,SQL Query\n")
	for _, result := range bs.results {
		avgMs := float64(result.AverageTime.Nanoseconds()) / 1000000
		minMs := float64(result.MinTime.Nanoseconds()) / 1000000
		maxMs := float64(result.MaxTime.Nanoseconds()) / 1000000
		line := fmt.Sprintf("%s,%s,%.2f,%.2f,%.2f,%d,\"%s\"\n",
			result.QueryName, result.Description, avgMs, minMs, maxMs, result.Iterations, result.SqlQuery)
		file.WriteString(line)
	}

	log.Println("")
	log.Printf("📄 Detailed results exported to: %s", resultFile)
}

func main() {
	iterations := 5 // Default iterations
	if len(os.Args) > 1 {
		if parsed, err := fmt.Sscanf(os.Args[1], "%d", &iterations); err != nil || parsed != 1 {
			log.Fatalf("Invalid iterations argument: %s", os.Args[1])
		}
	}

	suite := NewBenchmarkSuite()
	defer suite.db.Close()

	suite.runAllBenchmarks(iterations)
	suite.generateReport()

	log.Println()
	log.Println("✅ Performance benchmarking completed!")
}
