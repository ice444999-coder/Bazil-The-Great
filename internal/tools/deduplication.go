/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
// ============================================================================
// ARES DEDUPLICATION TOOLS
// Version: 1.0.0
// Date: October 19, 2025
//
// Purpose: Tool implementations for SQL deduplication and schema analysis
// Features:
// - dedup_sql_files: SHA-256 and semantic deduplication
// - build_schema_map: ER diagrams and dependency analysis
// - analyze_query_performance: Query optimization recommendations
// ============================================================================

package tools

import (
	"ares_api/internal/agent"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

// ToolDeduplication provides SQL file deduplication capabilities
type ToolDeduplication struct {
	DB *gorm.DB
}

// DedupSQLFilesParams parameters for deduplication
type DedupSQLFilesParams struct {
	Directory    string `json:"directory"`
	DryRun       bool   `json:"dry_run"`
	OutputFormat string `json:"output_format"` // "json" or "markdown"
}

// DedupSQLFilesResult result of deduplication operation
type DedupSQLFilesResult struct {
	TotalFiles      int              `json:"total_files"`
	DuplicateGroups []DuplicateGroup `json:"duplicate_groups"`
	UniqueFiles     int              `json:"unique_files"`
	SpaceSaved      int64            `json:"space_saved_bytes"`
	ProcessingTime  float64          `json:"processing_time_seconds"`
	Report          string           `json:"report,omitempty"`
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Hash     string   `json:"hash"`
	Method   string   `json:"method"` // "exact" or "semantic"
	Files    []string `json:"files"`
	Size     int64    `json:"size_bytes"`
	CanMerge bool     `json:"can_merge"`
}

// DedupSQLFiles executes SQL file deduplication using the Python script
func (td *ToolDeduplication) DedupSQLFiles(params DedupSQLFilesParams) (*DedupSQLFilesResult, error) {
	// Validate parameters
	if params.Directory == "" {
		return nil, fmt.Errorf("directory parameter is required")
	}

	if params.OutputFormat == "" {
		params.OutputFormat = "json"
	}

	// Check if Python script exists
	scriptPath := filepath.Join("scripts", "dedup_sql_files.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// Try relative path from workspace root
		scriptPath = "dedup_sql_files.py"
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("deduplication script not found at %s or %s", filepath.Join("scripts", "dedup_sql_files.py"), "dedup_sql_files.py")
		}
	}

	// Build command arguments
	args := []string{scriptPath, "--directory", params.Directory}
	if params.DryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, "--output-format", params.OutputFormat)

	// Execute Python script
	cmd := exec.Command("python", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("deduplication script failed: %w, output: %s", err, string(output))
	}

	// Parse output based on format
	if params.OutputFormat == "json" {
		var result DedupSQLFilesResult
		if err := json.Unmarshal(output, &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON output: %w", err)
		}
		return &result, nil
	} else {
		// Markdown format - return as report
		return &DedupSQLFilesResult{
			Report: string(output),
		}, nil
	}
}

// BuildSchemaMapParams parameters for schema map generation
type BuildSchemaMapParams struct {
	IncludeERDiagram    bool `json:"include_er_diagram"`
	IncludeDependencies bool `json:"include_dependencies"`
}

// BuildSchemaMapResult result of schema map generation
type BuildSchemaMapResult struct {
	Schemas         []agent.Schema                     `json:"schemas"`
	ERDiagram       *agent.ERDiagram                   `json:"er_diagram,omitempty"`
	DependencyGraph map[string]interface{}             `json:"dependency_graph,omitempty"`
	Recommendations []agent.OptimizationRecommendation `json:"recommendations"`
	GeneratedAt     string                             `json:"generated_at"`
	Version         string                             `json:"version"`
	Report          string                             `json:"report,omitempty"`
}

// BuildSchemaMap generates comprehensive schema analysis
func (td *ToolDeduplication) BuildSchemaMap(params BuildSchemaMapParams) (*BuildSchemaMapResult, error) {
	// Create schema optimizer
	optimizer := agent.NewSchemaOptimizer(td.DB)

	// Build schema map
	schemaMap, err := optimizer.BuildSchemaMap()
	if err != nil {
		return nil, fmt.Errorf("failed to build schema map: %w", err)
	}

	result := &BuildSchemaMapResult{
		Schemas:         schemaMap.Schemas,
		Recommendations: schemaMap.Recommendations,
		GeneratedAt:     schemaMap.GeneratedAt.Format("2006-01-02 15:04:05"),
		Version:         schemaMap.Version,
	}

	if params.IncludeERDiagram {
		result.ERDiagram = schemaMap.ERDiagram
	}

	if params.IncludeDependencies {
		result.DependencyGraph = schemaMap.DependencyGraph
	}

	// Generate markdown report
	result.Report = schemaMap.ToMarkdown()

	return result, nil
}

// AnalyzeQueryPerformanceResult result of query performance analysis
type AnalyzeQueryPerformanceResult struct {
	SlowQueries     []map[string]interface{} `json:"slow_queries"`
	TotalAnalyzed   int                      `json:"total_analyzed"`
	Recommendations []string                 `json:"recommendations"`
	Report          string                   `json:"report"`
}

// AnalyzeQueryPerformance analyzes query performance patterns
func (td *ToolDeduplication) AnalyzeQueryPerformance() (*AnalyzeQueryPerformanceResult, error) {
	// Create schema optimizer
	optimizer := agent.NewSchemaOptimizer(td.DB)

	// Analyze query performance
	performanceData, err := optimizer.AnalyzeQueryPerformance()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze query performance: %w", err)
	}

	result := &AnalyzeQueryPerformanceResult{
		Recommendations: []string{},
	}

	// Extract slow queries
	if slowQueries, ok := performanceData["slow_queries"]; ok {
		if sq, ok := slowQueries.([]map[string]interface{}); ok {
			result.SlowQueries = sq
			result.TotalAnalyzed = len(sq)
		}
	}

	// Extract recommendations
	if recs, ok := performanceData["recommendations"]; ok {
		if r, ok := recs.([]string); ok {
			result.Recommendations = r
		}
	}

	// Generate report
	var report strings.Builder
	report.WriteString("# Query Performance Analysis Report\n\n")
	report.WriteString(fmt.Sprintf("**Total Slow Queries Analyzed:** %d\n\n", result.TotalAnalyzed))

	if len(result.SlowQueries) > 0 {
		report.WriteString("## Slow Queries\n\n")
		for i, query := range result.SlowQueries {
			report.WriteString(fmt.Sprintf("### Query %d\n", i+1))
			if q, ok := query["query"].(string); ok {
				report.WriteString(fmt.Sprintf("**Query:** %s\n", q))
			}
			if calls, ok := query["calls"].(int64); ok {
				report.WriteString(fmt.Sprintf("**Calls:** %d\n", calls))
			}
			if meanTime, ok := query["mean_time"].(int64); ok {
				report.WriteString(fmt.Sprintf("**Mean Time:** %d ms\n", meanTime))
			}
			report.WriteString("\n")
		}
	}

	if len(result.Recommendations) > 0 {
		report.WriteString("## Recommendations\n\n")
		for _, rec := range result.Recommendations {
			report.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}

	result.Report = report.String()

	return result, nil
}
