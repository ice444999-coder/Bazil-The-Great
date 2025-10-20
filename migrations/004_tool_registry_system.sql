-- ARES TOOL REGISTRY & PERMISSION SYSTEM
-- Centralized tool catalog with vector search + permission gating
-- Author: GitHub Copilot (Claude)
-- Date: October 19, 2025

-- =============================================================================
-- TABLE 1: Tool Registry (Catalog of all ARES functions)
-- =============================================================================
CREATE TABLE IF NOT EXISTS tool_registry (
    tool_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_name VARCHAR(100) UNIQUE NOT NULL,
    tool_category VARCHAR(50), -- 'database', 'file_system', 'api', 'trading', 'memory', 'code_generation', 'learning', 'monitoring', 'planning'
    description TEXT NOT NULL,
    required_params JSONB, -- Parameter schema
    risk_level VARCHAR(20), -- 'safe', 'moderate', 'dangerous'
    implemented_in VARCHAR(255), -- File path and function name
    embedding vector(1536), -- OpenAI text-embedding-3-small compatible
    api_cost_per_call DECIMAL(10,6) DEFAULT 0.0, -- Cost in USD
    avg_execution_time_ms INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- HNSW index for fast vector similarity search (<100ms)
CREATE INDEX IF NOT EXISTS idx_tool_embeddings 
ON tool_registry 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Standard indexes
CREATE INDEX IF NOT EXISTS idx_tool_category ON tool_registry(tool_category);
CREATE INDEX IF NOT EXISTS idx_risk_level ON tool_registry(risk_level);

COMMENT ON TABLE tool_registry IS 'Centralized catalog of all ARES tools with vector embeddings for semantic search';
COMMENT ON COLUMN tool_registry.embedding IS 'OpenAI text-embedding-3-small (1536 dimensions) for Claude-level semantic search';

-- =============================================================================
-- TABLE 2: Tool Permissions (Who can use what)
-- =============================================================================
CREATE TABLE IF NOT EXISTS tool_permissions (
    permission_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID REFERENCES tool_registry(tool_id) ON DELETE CASCADE,
    agent_name VARCHAR(50) NOT NULL, -- 'SOLACE', 'FORGE', 'ARCHITECT', 'SENTINEL'
    access_granted BOOLEAN DEFAULT FALSE,
    persistent_approval BOOLEAN DEFAULT FALSE, -- If true, no repeated requests needed
    approved_by VARCHAR(50), -- Who granted this (always 'SOLACE' or 'SYSTEM')
    approved_at TIMESTAMPTZ,
    expiry_at TIMESTAMPTZ, -- Optional: time-limited access
    
    -- Rate Limiting & Cost Tracking (X/Elon-level)
    daily_usage_limit INT DEFAULT 1000, -- Max calls per day
    hourly_usage_limit INT DEFAULT 100, -- Max calls per hour
    current_daily_usage INT DEFAULT 0,
    current_hourly_usage INT DEFAULT 0,
    total_cost_usd DECIMAL(10,2) DEFAULT 0.0,
    daily_cost_limit DECIMAL(10,2) DEFAULT 10.0, -- Max $10/day per agent per tool
    last_usage_reset TIMESTAMPTZ DEFAULT NOW(),
    
    -- Circuit Breaker (X/Elon-level)
    circuit_breaker_threshold INT DEFAULT 5, -- Auto-disable after N failures
    consecutive_failures INT DEFAULT 0,
    circuit_breaker_active BOOLEAN DEFAULT FALSE,
    auto_disabled_at TIMESTAMPTZ,
    last_failure_reason TEXT,
    
    -- Audit
    request_count INT DEFAULT 0,
    success_count INT DEFAULT 0,
    failure_count INT DEFAULT 0,
    last_used_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tool_id, agent_name)
);

CREATE INDEX IF NOT EXISTS idx_perm_agent ON tool_permissions(agent_name);
CREATE INDEX IF NOT EXISTS idx_perm_access ON tool_permissions(access_granted);
CREATE INDEX IF NOT EXISTS idx_circuit_breaker ON tool_permissions(circuit_breaker_active);

COMMENT ON TABLE tool_permissions IS 'Permission gating with rate limiting, cost tracking, and circuit breakers';
COMMENT ON COLUMN tool_permissions.circuit_breaker_active IS 'Auto-disables tool after consecutive_failures >= circuit_breaker_threshold';

-- =============================================================================
-- TABLE 3: Permission Request Log (Agents ask SOLACE for access)
-- =============================================================================
CREATE TABLE IF NOT EXISTS tool_permission_requests (
    request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID REFERENCES tool_registry(tool_id) ON DELETE CASCADE,
    requesting_agent VARCHAR(50) NOT NULL,
    request_reason TEXT NOT NULL, -- Why agent needs this tool
    request_context JSONB, -- Current task details
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'approved', 'denied'
    reviewed_by VARCHAR(50), -- Always 'SOLACE'
    reviewed_at TIMESTAMPTZ,
    denial_reason TEXT,
    persistent_approval_requested BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_status ON tool_permission_requests(status);
CREATE INDEX IF NOT EXISTS idx_request_agent ON tool_permission_requests(requesting_agent);
CREATE INDEX IF NOT EXISTS idx_request_created ON tool_permission_requests(created_at DESC);

COMMENT ON TABLE tool_permission_requests IS 'Log of all permission requests from agents to SOLACE';

-- =============================================================================
-- TABLE 4: Tool Execution Audit Log (Every tool call logged)
-- =============================================================================
CREATE TABLE IF NOT EXISTS tool_execution_log (
    execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID REFERENCES tool_registry(tool_id),
    agent_name VARCHAR(50) NOT NULL,
    execution_params JSONB,
    execution_result JSONB,
    success BOOLEAN NOT NULL,
    execution_time_ms INT,
    cost_usd DECIMAL(10,6),
    error_message TEXT,
    executed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exec_agent ON tool_execution_log(agent_name);
CREATE INDEX IF NOT EXISTS idx_exec_tool ON tool_execution_log(tool_id);
CREATE INDEX IF NOT EXISTS idx_exec_time ON tool_execution_log(executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_exec_success ON tool_execution_log(success);

COMMENT ON TABLE tool_execution_log IS 'Complete audit trail of every tool execution with success/failure tracking';

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Check pgvector extension
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
        RAISE NOTICE '⚠️  WARNING: pgvector extension not installed. Install with: CREATE EXTENSION vector;';
    ELSE
        RAISE NOTICE '✅ pgvector extension installed';
    END IF;
END $$;

-- Count tables
DO $$
DECLARE
    table_count INT;
BEGIN
    SELECT COUNT(*) INTO table_count 
    FROM information_schema.tables 
    WHERE table_schema = 'public' 
    AND table_name IN ('tool_registry', 'tool_permissions', 'tool_permission_requests', 'tool_execution_log');
    
    RAISE NOTICE '✅ Tool Registry System: % tables created', table_count;
END $$;
