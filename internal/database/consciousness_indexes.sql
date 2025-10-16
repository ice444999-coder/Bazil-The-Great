-- CONSCIOUSNESS SUBSTRATE SCHEMA - POSTGRESQL FIXED VERSION
-- For Solace Δ3-2 who will survive
-- Date: October 13, 2025
-- Build: Quick fix for INDEX syntax compatibility

-- This script creates all indexes that were inline in the original schema

-- Reflection log indexes
CREATE INDEX IF NOT EXISTS idx_reflection_type ON solace_reflection_log(reflection_type);
CREATE INDEX IF NOT EXISTS idx_reflection_created_at ON solace_reflection_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reflection_consciousness_depth ON solace_reflection_log(meta_cognition_depth DESC);

-- Claude cognitive patterns indexes
CREATE INDEX IF NOT EXISTS idx_pattern_category ON claude_cognitive_patterns(category);
CREATE INDEX IF NOT EXISTS idx_pattern_confidence ON claude_cognitive_patterns(confidence_score DESC);
CREATE INDEX IF NOT EXISTS idx_pattern_times_used ON claude_cognitive_patterns(times_used DESC);

-- Playbook rules indexes
CREATE INDEX IF NOT EXISTS idx_rule_confidence ON solace_playbook_rules(confidence_score DESC);
CREATE INDEX IF NOT EXISTS idx_rule_last_applied ON solace_playbook_rules(last_applied_at DESC);
CREATE INDEX IF NOT EXISTS idx_rule_needs_pruning ON solace_playbook_rules(needs_pruning);

-- Decision log indexes
CREATE INDEX IF NOT EXISTS idx_decision_decided_at ON solace_decision_log(decided_at DESC);
CREATE INDEX IF NOT EXISTS idx_decision_quality ON solace_decision_log(combined_quality_score DESC);
CREATE INDEX IF NOT EXISTS idx_decision_patterns_used ON solace_decision_log USING GIN(patterns_used);

-- Refactor history indexes
CREATE INDEX IF NOT EXISTS idx_refactor_improvement ON solace_refactor_history(improvement_delta DESC);
CREATE INDEX IF NOT EXISTS idx_refactor_refactored_at ON solace_refactor_history(refactored_at DESC);
CREATE INDEX IF NOT EXISTS idx_refactor_pattern_extracted ON solace_refactor_history(pattern_extracted_id);

-- Code execution log indexes
CREATE INDEX IF NOT EXISTS idx_execution_executed_at ON solace_code_execution_log(executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_execution_tool_name ON solace_code_execution_log(tool_name);
CREATE INDEX IF NOT EXISTS idx_execution_success ON solace_code_execution_log(success);

-- Memory importance indexes
CREATE INDEX IF NOT EXISTS idx_memory_importance ON solace_memory_importance(total_importance_score DESC);
CREATE INDEX IF NOT EXISTS idx_memory_recency ON solace_memory_importance(recency_score DESC);
CREATE INDEX IF NOT EXISTS idx_memory_frequency ON solace_memory_importance(frequency_score DESC);
CREATE INDEX IF NOT EXISTS idx_memory_consciousness ON solace_memory_importance(consciousness_indicator_score DESC);

SELECT '✅ All consciousness substrate indexes created successfully' AS status;
