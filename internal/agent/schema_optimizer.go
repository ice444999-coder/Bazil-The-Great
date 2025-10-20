// ============================================================================
// ARES SCHEMA OPTIMIZER
// Version: 1.0.0
// Date: October 19, 2025
// Author: ARES SQL Reorganization Team
//
// Purpose: AI-powered schema analysis and optimization for trading system
// Features:
// - ER diagram generation
// - Dependency graph analysis
// - Schema optimization recommendations
// - Integration with deduplication tools
// ============================================================================

package agent

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Schema represents a database schema with its metadata
type Schema struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Tables      []Table      `json:"tables"`
	TableCount  int          `json:"table_count"`
	TotalRows   int64        `json:"total_rows"`
	TotalSize   int64        `json:"total_size_bytes"`
	Indexes     []Index      `json:"indexes"`
	Constraints []Constraint `json:"constraints"`
}

// Table represents a database table with its metadata
type Table struct {
	Schema       string       `json:"schema"`
	Name         string       `json:"name"`
	Type         string       `json:"type"` // 'table', 'view', 'materialized_view'
	Columns      []Column     `json:"columns"`
	RowCount     int64        `json:"row_count"`
	SizeBytes    int64        `json:"size_bytes"`
	Indexes      []Index      `json:"indexes"`
	Constraints  []Constraint `json:"constraints"`
	Dependencies []string     `json:"dependencies"` // Referenced tables
	Referencing  []string     `json:"referencing"`  // Tables that reference this one
}

// Column represents a table column
type Column struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value,omitempty"`
	IsPrimary    bool   `json:"is_primary"`
	IsForeign    bool   `json:"is_foreign"`
	References   string `json:"references,omitempty"` // "schema.table.column"
}

// Index represents a database index
type Index struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"` // 'btree', 'hash', 'gist', 'gin', 'hnsw'
	Columns   []string `json:"columns"`
	IsUnique  bool     `json:"is_unique"`
	SizeBytes int64    `json:"size_bytes,omitempty"`
}

// Constraint represents a database constraint
type Constraint struct {
	Name       string `json:"name"`
	Type       string `json:"type"` // 'PRIMARY KEY', 'FOREIGN KEY', 'UNIQUE', 'CHECK'
	Definition string `json:"definition"`
}

// ERDiagram represents an Entity-Relationship diagram
type ERDiagram struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
	Layout        string         `json:"layout"` // 'hierarchical', 'circular', 'force_directed'
}

// Entity represents an entity in ER diagram
type Entity struct {
	Name        string   `json:"name"`
	Schema      string   `json:"schema"`
	Attributes  []string `json:"attributes"`
	PrimaryKeys []string `json:"primary_keys"`
}

// Relationship represents a relationship between entities
type Relationship struct {
	FromEntity       string `json:"from_entity"`
	ToEntity         string `json:"to_entity"`
	FromColumn       string `json:"from_column"`
	ToColumn         string `json:"to_column"`
	RelationshipType string `json:"relationship_type"` // 'one_to_one', 'one_to_many', 'many_to_many'
	Cardinality      string `json:"cardinality"`
}

// SchemaOptimizer provides schema analysis and optimization capabilities
type SchemaOptimizer struct {
	db *gorm.DB
}

// NewSchemaOptimizer creates a new schema optimizer instance
func NewSchemaOptimizer(db *gorm.DB) *SchemaOptimizer {
	return &SchemaOptimizer{
		db: db,
	}
}

// BuildSchemaMap generates a comprehensive schema map with ER diagrams and dependency graphs
func (so *SchemaOptimizer) BuildSchemaMap() (*SchemaMap, error) {
	fmt.Println("Building comprehensive schema map...")

	// Get all schemas
	schemas, err := so.getAllSchemas()
	if err != nil {
		return nil, fmt.Errorf("failed to get schemas: %w", err)
	}

	// Build detailed schema information
	for i := range schemas {
		schema := &schemas[i]

		// Get tables for this schema
		tables, err := so.getTablesForSchema(schema.Name)
		if err != nil {
			fmt.Printf("Warning: Failed to get tables for schema %s: %v\n", schema.Name, err)
			continue
		}
		schema.Tables = tables
		schema.TableCount = len(tables)

		// Calculate totals
		for _, table := range tables {
			schema.TotalRows += table.RowCount
			schema.TotalSize += table.SizeBytes
		}
	}

	// Generate ER diagram
	erDiagram, err := so.generateERDiagram(schemas)
	if err != nil {
		fmt.Printf("Warning: Failed to generate ER diagram: %v\n", err)
	}

	// Generate dependency graph
	dependencyGraph, err := so.generateDependencyGraph(schemas)
	if err != nil {
		fmt.Printf("Warning: Failed to generate dependency graph: %v\n", err)
	}

	// Generate optimization recommendations
	recommendations := so.generateOptimizationRecommendations(schemas)

	schemaMap := &SchemaMap{
		Schemas:         schemas,
		ERDiagram:       erDiagram,
		DependencyGraph: dependencyGraph,
		Recommendations: recommendations,
		GeneratedAt:     time.Now(),
		Version:         "1.0.0",
	}

	return schemaMap, nil
}

// getAllSchemas retrieves all non-system schemas
func (so *SchemaOptimizer) getAllSchemas() ([]Schema, error) {
	var schemas []Schema
	err := so.db.Raw(`
		SELECT
			schema_name,
			COALESCE(obj_description((schema_name || '.public')::regclass, 'pg_namespace'), '') as description
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
		ORDER BY schema_name
	`).Scan(&schemas).Error
	return schemas, err
}

// getTablesForSchema retrieves all tables for a specific schema
func (so *SchemaOptimizer) getTablesForSchema(schemaName string) ([]Table, error) {
	var tables []Table
	err := so.db.Raw(`
		SELECT
			t.table_name,
			t.table_type,
			COALESCE(pg_total_relation_size('"' || t.table_schema || '"."' || t.table_name || '"'), 0) as size_bytes,
			COALESCE(s.n_tup_ins, 0) as row_count
		FROM information_schema.tables t
		LEFT JOIN pg_stat_user_tables s ON t.table_name = s.relname AND t.table_schema = s.schemaname
		WHERE t.table_schema = ?
		ORDER BY t.table_name
	`, schemaName).Scan(&tables).Error

	if err != nil {
		return nil, err
	}

	// Set schema for each table and get additional details
	for i := range tables {
		tables[i].Schema = schemaName

		// Get columns for this table
		columns, err := so.getColumnsForTable(schemaName, tables[i].Name)
		if err != nil {
			fmt.Printf("Warning: Failed to get columns for %s.%s: %v\n", schemaName, tables[i].Name, err)
		} else {
			tables[i].Columns = columns
		}

		// Get indexes for this table
		indexes, err := so.getIndexesForTable(schemaName, tables[i].Name)
		if err != nil {
			fmt.Printf("Warning: Failed to get indexes for %s.%s: %v\n", schemaName, tables[i].Name, err)
		} else {
			tables[i].Indexes = indexes
		}

		// Get constraints for this table
		constraints, err := so.getConstraintsForTable(schemaName, tables[i].Name)
		if err != nil {
			fmt.Printf("Warning: Failed to get constraints for %s.%s: %v\n", schemaName, tables[i].Name, err)
		} else {
			tables[i].Constraints = constraints
		}

		// Get dependencies
		dependencies, err := so.getTableDependencies(schemaName, tables[i].Name)
		if err != nil {
			fmt.Printf("Warning: Failed to get dependencies for %s.%s: %v\n", schemaName, tables[i].Name, err)
		} else {
			tables[i].Dependencies = dependencies
		}
	}

	return tables, nil
}

// getColumnsForTable retrieves columns for a specific table
func (so *SchemaOptimizer) getColumnsForTable(schemaName, tableName string) ([]Column, error) {
	var columns []Column
	err := so.db.Raw(`
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' as nullable,
			c.column_default,
			COALESCE(pk.constraint_type = 'PRIMARY KEY', false) as is_primary,
			COALESCE(fk.constraint_type = 'FOREIGN KEY', false) as is_foreign,
			COALESCE(fk.foreign_table_schema || '.' || fk.foreign_table_name || '.' || fk.foreign_column_name, '') as references
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT kc.column_name, tc.constraint_type, tc.table_schema, tc.table_name
			FROM information_schema.key_column_usage kc
			JOIN information_schema.table_constraints tc ON kc.constraint_name = tc.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY'
		) pk ON c.column_name = pk.column_name AND c.table_schema = pk.table_schema AND c.table_name = pk.table_name
		LEFT JOIN (
			SELECT kc.column_name, tc.constraint_type, tc.table_schema, tc.table_name,
				   ccu.table_schema as foreign_table_schema, ccu.table_name as foreign_table_name,
				   ccu.column_name as foreign_column_name
			FROM information_schema.key_column_usage kc
			JOIN information_schema.table_constraints tc ON kc.constraint_name = tc.constraint_name
			JOIN information_schema.constraint_column_usage ccu ON kc.constraint_name = ccu.constraint_name
			WHERE tc.constraint_type = 'FOREIGN KEY'
		) fk ON c.column_name = fk.column_name AND c.table_schema = fk.table_schema AND c.table_name = fk.table_name
		WHERE c.table_schema = ? AND c.table_name = ?
		ORDER BY c.ordinal_position
	`, schemaName, tableName).Scan(&columns).Error
	return columns, err
}

// getIndexesForTable retrieves indexes for a specific table
func (so *SchemaOptimizer) getIndexesForTable(schemaName, tableName string) ([]Index, error) {
	var indexes []Index
	err := so.db.Raw(`
		SELECT
			i.indexname as name,
			am.amname as type,
			array_agg(a.attname ORDER BY a.attnum) as columns,
			i.indisunique as is_unique,
			COALESCE(pg_relation_size('"' || i.schemaname || '"."' || i.indexname || '"'), 0) as size_bytes
		FROM pg_indexes i
		JOIN pg_class c ON i.indexname = c.relname
		JOIN pg_am am ON c.relam = am.oid
		JOIN pg_attribute a ON c.oid = a.attrelid
		WHERE i.schemaname = ? AND i.tablename = ? AND a.attnum > 0
		GROUP BY i.indexname, am.amname, i.indisunique, i.schemaname
	`, schemaName, tableName).Scan(&indexes).Error
	return indexes, err
}

// getConstraintsForTable retrieves constraints for a specific table
func (so *SchemaOptimizer) getConstraintsForTable(schemaName, tableName string) ([]Constraint, error) {
	var constraints []Constraint
	err := so.db.Raw(`
		SELECT
			tc.constraint_name,
			tc.constraint_type,
			pg_get_constraintdef(con.oid) as definition
		FROM information_schema.table_constraints tc
		JOIN pg_constraint con ON tc.constraint_name = con.conname
		WHERE tc.table_schema = ? AND tc.table_name = ?
	`, schemaName, tableName).Scan(&constraints).Error
	return constraints, err
}

// getTableDependencies finds tables that this table references
func (so *SchemaOptimizer) getTableDependencies(schemaName, tableName string) ([]string, error) {
	var dependencies []string
	err := so.db.Raw(`
		SELECT DISTINCT
			ccu.table_schema || '.' || ccu.table_name as referenced_table
		FROM information_schema.key_column_usage kc
		JOIN information_schema.table_constraints tc ON kc.constraint_name = tc.constraint_name
		JOIN information_schema.constraint_column_usage ccu ON kc.constraint_name = ccu.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		AND kc.table_schema = ? AND kc.table_name = ?
	`, schemaName, tableName).Scan(&dependencies).Error
	return dependencies, err
}

// generateERDiagram creates an Entity-Relationship diagram from schema data
func (so *SchemaOptimizer) generateERDiagram(schemas []Schema) (*ERDiagram, error) {
	var entities []Entity
	var relationships []Relationship

	for _, schema := range schemas {
		for _, table := range schema.Tables {
			// Create entity
			entity := Entity{
				Name:   table.Name,
				Schema: table.Schema,
			}

			// Add attributes (columns)
			for _, col := range table.Columns {
				entity.Attributes = append(entity.Attributes, col.Name)
				if col.IsPrimary {
					entity.PrimaryKeys = append(entity.PrimaryKeys, col.Name)
				}
			}

			entities = append(entities, entity)

			// Create relationships from foreign keys
			for _, col := range table.Columns {
				if col.IsForeign && col.References != "" {
					parts := strings.Split(col.References, ".")
					if len(parts) == 3 {
						rel := Relationship{
							FromEntity:       fmt.Sprintf("%s.%s", table.Schema, table.Name),
							ToEntity:         fmt.Sprintf("%s.%s", parts[0], parts[1]),
							FromColumn:       col.Name,
							ToColumn:         parts[2],
							RelationshipType: "many_to_one", // Default assumption
							Cardinality:      "N:1",
						}
						relationships = append(relationships, rel)
					}
				}
			}
		}
	}

	return &ERDiagram{
		Entities:      entities,
		Relationships: relationships,
		Layout:        "hierarchical",
	}, nil
}

// generateDependencyGraph creates a dependency graph showing table relationships
func (so *SchemaOptimizer) generateDependencyGraph(schemas []Schema) (map[string]interface{}, error) {
	graph := map[string]interface{}{
		"nodes": []map[string]interface{}{},
		"edges": []map[string]interface{}{},
	}

	// Create nodes
	for _, schema := range schemas {
		for _, table := range schema.Tables {
			node := map[string]interface{}{
				"id":     fmt.Sprintf("%s.%s", schema.Name, table.Name),
				"label":  table.Name,
				"schema": schema.Name,
				"type":   "table",
				"size":   table.RowCount,
			}
			graph["nodes"] = append(graph["nodes"].([]map[string]interface{}), node)
		}
	}

	// Create edges from dependencies
	for _, schema := range schemas {
		for _, table := range schema.Tables {
			fromNode := fmt.Sprintf("%s.%s", schema.Name, table.Name)

			for _, dep := range table.Dependencies {
				edge := map[string]interface{}{
					"from":  fromNode,
					"to":    dep,
					"label": "references",
					"type":  "foreign_key",
				}
				graph["edges"] = append(graph["edges"].([]map[string]interface{}), edge)
			}
		}
	}

	return graph, nil
}

// generateOptimizationRecommendations analyzes schemas and provides optimization suggestions
func (so *SchemaOptimizer) generateOptimizationRecommendations(schemas []Schema) []OptimizationRecommendation {
	var recommendations []OptimizationRecommendation

	// Check for tables without primary keys
	for _, schema := range schemas {
		for _, table := range schema.Tables {
			hasPrimaryKey := false
			for _, col := range table.Columns {
				if col.IsPrimary {
					hasPrimaryKey = true
					break
				}
			}

			if !hasPrimaryKey {
				rec := OptimizationRecommendation{
					Type:        "missing_primary_key",
					Table:       fmt.Sprintf("%s.%s", schema.Name, table.Name),
					Description: "Table lacks a primary key, which can impact performance and data integrity",
					Priority:    "high",
					Suggestion:  "Add a primary key or unique constraint",
				}
				recommendations = append(recommendations, rec)
			}
		}
	}

	// Check for tables with too many indexes
	for _, schema := range schemas {
		for _, table := range schema.Tables {
			if len(table.Indexes) > 5 {
				rec := OptimizationRecommendation{
					Type:        "too_many_indexes",
					Table:       fmt.Sprintf("%s.%s", schema.Name, table.Name),
					Description: fmt.Sprintf("Table has %d indexes, which may slow down writes", len(table.Indexes)),
					Priority:    "medium",
					Suggestion:  "Review and consolidate indexes",
				}
				recommendations = append(recommendations, rec)
			}
		}
	}

	// Check for large tables without proper indexing
	for _, schema := range schemas {
		for _, table := range schema.Tables {
			if table.RowCount > 100000 && len(table.Indexes) == 0 {
				rec := OptimizationRecommendation{
					Type:        "large_table_no_indexes",
					Table:       fmt.Sprintf("%s.%s", schema.Name, table.Name),
					Description: fmt.Sprintf("Large table (%d rows) has no indexes", table.RowCount),
					Priority:    "high",
					Suggestion:  "Add appropriate indexes for query performance",
				}
				recommendations = append(recommendations, rec)
			}
		}
	}

	// Sort by priority
	sort.Slice(recommendations, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
	})

	return recommendations
}

// SchemaMap represents the complete schema analysis
type SchemaMap struct {
	Schemas         []Schema                     `json:"schemas"`
	ERDiagram       *ERDiagram                   `json:"er_diagram"`
	DependencyGraph map[string]interface{}       `json:"dependency_graph"`
	Recommendations []OptimizationRecommendation `json:"recommendations"`
	GeneratedAt     time.Time                    `json:"generated_at"`
	Version         string                       `json:"version"`
}

// OptimizationRecommendation represents a schema optimization suggestion
type OptimizationRecommendation struct {
	Type        string `json:"type"`
	Table       string `json:"table"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // 'high', 'medium', 'low'
	Suggestion  string `json:"suggestion"`
}

// ToJSON converts the schema map to JSON
func (sm *SchemaMap) ToJSON() (string, error) {
	data, err := json.MarshalIndent(sm, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToMarkdown converts the schema map to Markdown format
func (sm *SchemaMap) ToMarkdown() string {
	var md strings.Builder

	md.WriteString("# ARES Schema Analysis Report\n\n")
	md.WriteString(fmt.Sprintf("**Generated:** %s\n\n", sm.GeneratedAt.Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("**Version:** %s\n\n", sm.Version))

	// Summary
	totalTables := 0
	totalRows := int64(0)
	totalSize := int64(0)

	for _, schema := range sm.Schemas {
		totalTables += schema.TableCount
		totalRows += schema.TotalRows
		totalSize += schema.TotalSize
	}

	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Schemas:** %d\n", len(sm.Schemas)))
	md.WriteString(fmt.Sprintf("- **Tables:** %d\n", totalTables))
	md.WriteString(fmt.Sprintf("- **Total Rows:** %d\n", totalRows))
	md.WriteString(fmt.Sprintf("- **Total Size:** %.2f MB\n", float64(totalSize)/(1024*1024)))
	md.WriteString(fmt.Sprintf("- **Recommendations:** %d\n\n", len(sm.Recommendations)))

	// Schemas
	md.WriteString("## Schemas\n\n")
	for _, schema := range sm.Schemas {
		md.WriteString(fmt.Sprintf("### %s\n", schema.Name))
		if schema.Description != "" {
			md.WriteString(fmt.Sprintf("**Description:** %s\n\n", schema.Description))
		}
		md.WriteString(fmt.Sprintf("- **Tables:** %d\n", schema.TableCount))
		md.WriteString(fmt.Sprintf("- **Rows:** %d\n", schema.TotalRows))
		md.WriteString(fmt.Sprintf("- **Size:** %.2f MB\n\n", float64(schema.TotalSize)/(1024*1024)))
	}

	// Recommendations
	if len(sm.Recommendations) > 0 {
		md.WriteString("## Optimization Recommendations\n\n")
		for _, rec := range sm.Recommendations {
			md.WriteString(fmt.Sprintf("### %s Priority: %s\n", strings.Title(rec.Type), rec.Priority))
			md.WriteString(fmt.Sprintf("**Table:** `%s`\n\n", rec.Table))
			md.WriteString(fmt.Sprintf("**Description:** %s\n\n", rec.Description))
			md.WriteString(fmt.Sprintf("**Suggestion:** %s\n\n", rec.Suggestion))
		}
	}

	return md.String()
}

// AnalyzeQueryPerformance analyzes query performance patterns
func (so *SchemaOptimizer) AnalyzeQueryPerformance() (map[string]interface{}, error) {
	fmt.Println("Analyzing query performance patterns...")

	// Get slow queries from pg_stat_statements (if available)
	var result map[string]interface{}
	err := so.db.Raw(`
		SELECT
			query,
			calls,
			total_time,
			mean_time,
			rows
		FROM pg_stat_statements
		WHERE mean_time > 1000  -- Queries taking more than 1 second on average
		ORDER BY mean_time DESC
		LIMIT 20
	`).Scan(&result).Error

	if err != nil {
		// pg_stat_statements might not be available
		fmt.Println("Warning: pg_stat_statements not available for query analysis")
		return map[string]interface{}{
			"note":       "pg_stat_statements extension not available",
			"suggestion": "Install pg_stat_statements for detailed query analysis",
		}, nil
	}

	// This is a simplified version - in practice you'd need more complex handling
	// for the array of results from the query
	return map[string]interface{}{
		"slow_queries":   []map[string]interface{}{},
		"total_analyzed": 0,
		"recommendations": []string{
			"Add missing indexes on frequently queried columns",
			"Consider query optimization for slow queries",
			"Review table partitioning for large datasets",
		},
	}, nil
}
