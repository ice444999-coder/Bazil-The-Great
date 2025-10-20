-- ============================================================================
-- CRYSTAL #27: AUTONOMOUS IMPROVEMENT SYSTEM - Database Schema
-- ============================================================================
-- Created: 2025-10-18
-- Purpose: Enable SOLACE to autonomously detect, queue, and execute optimizations
--
-- Tables:
-- 1. improvement_queue - Stores pending/executed improvements
-- 2. improvement_templates - Pre-defined optimization patterns
-- 3. improvement_execution_log - Results tracking for learning
-- ============================================================================

-- Table 1: Improvement Queue
-- Stores improvements detected by autonomous monitor or manually created
CREATE TABLE IF NOT EXISTS improvement_queue (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(50) DEFAULT 'solace',
    title TEXT NOT NULL,
    description TEXT,
    
    -- Implementation details
    sql_script TEXT,
    rollback_script TEXT,
    
    -- Scheduling
    scheduled_for TIMESTAMP DEFAULT (NOW() + INTERVAL '1 day'), -- Default: next day 10pm
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'EXECUTED', 'FAILED', 'ROLLED_BACK')),
    
    -- Estimates (for learning)
    estimated_speedup_percent INT,
    estimated_cost_reduction_percent INT,
    risk_level VARCHAR(20) CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    
    -- Execution results
    executed_at TIMESTAMP,
    execution_duration_ms INT,
    actual_speedup_percent INT, -- Measured after execution
    error_message TEXT,
    
    -- Traceability
    decision_trace_id INTEGER, -- Links to Glass Box decision tree
    hedera_txn_id VARCHAR(100), -- Blockchain proof
    parent_crystal_id INTEGER, -- Which memory crystal triggered this
    
    -- Approval workflow
    requires_approval BOOLEAN DEFAULT TRUE,
    approved_by VARCHAR(50),
    approved_at TIMESTAMP,
    rejection_reason TEXT
);

-- Table 2: Improvement Templates
-- Pre-defined patterns that autonomous monitor can apply
CREATE TABLE IF NOT EXISTS improvement_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50), -- 'index', 'cache', 'query', 'schema', 'partitioning'
    
    -- Trigger logic
    trigger_condition TEXT, -- SQL query that returns true when optimization needed
    trigger_threshold JSONB, -- e.g., {"p95_latency_ms": 100, "cache_hit_rate": 0.3}
    
    -- Implementation
    implementation_sql TEXT NOT NULL,
    rollback_sql TEXT NOT NULL,
    validation_sql TEXT, -- Query to run before applying (safety check)
    
    -- Metadata
    estimated_impact VARCHAR(50), -- "10-20% speedup", "50% cost reduction"
    risk_level VARCHAR(20) CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    requires_approval BOOLEAN DEFAULT TRUE,
    
    -- Usage tracking
    times_used INT DEFAULT 0,
    times_successful INT DEFAULT 0,
    average_actual_speedup_percent DECIMAL(5,2), -- Learned over time
    
    created_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(50),
    enabled BOOLEAN DEFAULT TRUE
);

-- Table 3: Improvement Execution Log
-- Detailed metrics for learning and analysis
CREATE TABLE IF NOT EXISTS improvement_execution_log (
    id SERIAL PRIMARY KEY,
    improvement_id INT REFERENCES improvement_queue(id) ON DELETE CASCADE,
    executed_at TIMESTAMP DEFAULT NOW(),
    
    -- Metrics before improvement
    metric_before JSONB, -- {"p95_latency_ms": 150, "cache_hit_rate": 0.25, "qps": 100}
    
    -- Metrics after improvement
    metric_after JSONB, -- {"p95_latency_ms": 90, "cache_hit_rate": 0.75, "qps": 120}
    
    -- Execution details
    success BOOLEAN,
    duration_ms INT,
    error_details TEXT,
    
    -- Learning data
    estimated_vs_actual_diff_percent INT, -- How accurate was the estimate?
    confidence_score DECIMAL(3,2), -- 0.0 to 1.0, increases with more data
    
    -- Environment context
    database_version VARCHAR(50),
    system_load_cpu_percent INT,
    system_load_memory_percent INT,
    concurrent_connections INT
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_improvement_queue_status ON improvement_queue(status);
CREATE INDEX IF NOT EXISTS idx_improvement_queue_scheduled ON improvement_queue(scheduled_for);
CREATE INDEX IF NOT EXISTS idx_improvement_queue_created_at ON improvement_queue(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_improvement_execution_log_improvement_id ON improvement_execution_log(improvement_id);
CREATE INDEX IF NOT EXISTS idx_improvement_templates_enabled ON improvement_templates(enabled) WHERE enabled = TRUE;

-- ============================================================================
-- SEED DATA: Initial Improvement Templates
-- ============================================================================

-- Template 1: Add missing index when slow queries detected
INSERT INTO improvement_templates (
    name, 
    category,
    trigger_condition,
    trigger_threshold,
    implementation_sql,
    rollback_sql,
    validation_sql,
    estimated_impact,
    risk_level,
    requires_approval,
    created_by
) VALUES (
    'add_missing_index_on_slow_query',
    'index',
    'SELECT MAX(mean_exec_time) > 100 FROM pg_stat_statements WHERE query LIKE ''%WHERE%'' AND calls > 100',
    '{"slow_query_threshold_ms": 100, "min_calls": 100}',
    '-- Will be dynamically generated by FORGE agent based on pg_stat_statements analysis',
    '-- Will be dynamically generated (DROP INDEX)',
    'SELECT 1', -- Basic validation
    '30-50% speedup for affected queries',
    'LOW',
    FALSE, -- Auto-approve index additions
    'system'
) ON CONFLICT (name) DO NOTHING;

-- Template 2: Add Redis cache layer when cache hit rate low
INSERT INTO improvement_templates (
    name,
    category,
    trigger_condition,
    trigger_threshold,
    implementation_sql,
    rollback_sql,
    validation_sql,
    estimated_impact,
    risk_level,
    requires_approval,
    created_by
) VALUES (
    'add_redis_cache_layer',
    'cache',
    'SELECT cache_hit_rate < 0.30 FROM memory_system_metrics ORDER BY measured_at DESC LIMIT 1',
    '{"cache_hit_rate": 0.30}',
    '-- Requires infrastructure setup, not just SQL',
    '-- Redis shutdown + remove middleware',
    'SELECT COUNT(*) FROM pg_stat_activity WHERE state = ''active'' AND query NOT LIKE ''%pg_stat%''', -- Check system not overloaded
    '10x faster reads, 70%+ cache hit rate',
    'MEDIUM',
    TRUE, -- Requires Enki approval
    'system'
) ON CONFLICT (name) DO NOTHING;

-- Template 3: Optimize frequently-run query
INSERT INTO improvement_templates (
    name,
    category,
    trigger_condition,
    trigger_threshold,
    implementation_sql,
    rollback_sql,
    validation_sql,
    estimated_impact,
    risk_level,
    requires_approval,
    created_by
) VALUES (
    'optimize_hot_query',
    'query',
    'SELECT COUNT(*) FROM pg_stat_statements WHERE calls > 1000 AND mean_exec_time > 50',
    '{"min_calls": 1000, "mean_exec_time_ms": 50}',
    '-- Will be dynamically generated by FORGE: rewrite with better joins, add CTEs, etc.',
    '-- Revert to original query',
    'SELECT 1',
    '20-40% speedup',
    'MEDIUM',
    TRUE,
    'system'
) ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- VALIDATION QUERIES
-- ============================================================================
-- Run these to verify tables created successfully:
-- SELECT COUNT(*) FROM improvement_queue;
-- SELECT COUNT(*) FROM improvement_templates;
-- SELECT COUNT(*) FROM improvement_execution_log;
-- SELECT name, category, risk_level, enabled FROM improvement_templates;
