-- Migration 013: SOLACE Orchestration System
-- Purpose: Give SOLACE architectural knowledge and GitHub management tools

-- Architecture Rules: Defines where different feature types go
CREATE TABLE IF NOT EXISTS architecture_rules (
    id SERIAL PRIMARY KEY,
    feature_type VARCHAR(100) NOT NULL,  -- e.g., "agent_api_endpoint", "ui_view", "trading_feature"
    backend_pattern TEXT,                 -- e.g., "internal/api/controllers/{feature}_controller.go"
    frontend_pattern TEXT,                -- e.g., "AvaloniApp/Views/{Feature}View.axaml"
    integration_points TEXT[],            -- Array of files that need updating to wire feature
    rules_description TEXT,               -- Human-readable rules
    examples TEXT[],                      -- Example implementations
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_architecture_feature_type ON architecture_rules(feature_type);

-- Auto-update timestamp trigger
CREATE OR REPLACE FUNCTION update_architecture_rules_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER architecture_rules_updated
    BEFORE UPDATE ON architecture_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_architecture_rules_timestamp();

-- Seed with known patterns from your system
INSERT INTO architecture_rules (feature_type, backend_pattern, frontend_pattern, integration_points, rules_description, examples) VALUES
('agent_api_endpoint', 
 'internal/api/controllers/agent_controller.go', 
 'AvaloniApp/Views/AgentDashboardView.axaml',
 ARRAY['cmd/main.go', 'internal/api/routes/v1.go'],
 'Agent features: Backend controller handles REST API, Avalonia view displays UI, must register routes in v1.go',
 ARRAY['GetAgentStatus endpoint in agent_controller.go', 'AgentStatusCard component in AgentDashboardView.axaml']),

('trading_api_endpoint',
 'internal/api/controllers/trading_controller.go',
 'AvaloniApp/Views/TradingView.axaml',
 ARRAY['cmd/main.go', 'internal/api/routes/v1.go', 'internal/services/trading_service.go'],
 'Trading features: Backend controller → trading service → database, Avalonia UI calls REST endpoints',
 ARRAY['ExecuteTrade in trading_controller.go', 'OrderForm in TradingView.axaml']),

('health_monitoring',
 'internal/api/controllers/health_controller.go',
 'AvaloniApp/Views/SystemHealthView.axaml',
 ARRAY['cmd/main.go', 'internal/api/routes/v1.go'],
 'Health features: Backend exposes /health endpoints, UI polls and displays status',
 ARRAY['/health/detailed endpoint', 'HealthStatusPanel in SystemHealthView.axaml']);

COMMENT ON TABLE architecture_rules IS 'Master architecture patterns that SOLACE uses to generate correct file locations and integration points';

-- ===========================================
-- GITHUB: After running this migration:
-- You will see "CREATE TABLE" and "CREATE INDEX" messages
-- This output IS your verification - do not SELECT
-- ===========================================

-- File Inspection Cache: SOLACE's view of the repo state
CREATE TABLE IF NOT EXISTS repo_file_cache (
    id SERIAL PRIMARY KEY,
    file_path TEXT NOT NULL UNIQUE,
    file_type VARCHAR(20),
    content_hash VARCHAR(64),
    line_count INTEGER,
    last_inspected TIMESTAMP,
    last_modified TIMESTAMP,
    size_bytes BIGINT,
    is_tracked BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_repo_file_path ON repo_file_cache(file_path);
CREATE INDEX idx_repo_file_type ON repo_file_cache(file_type);
CREATE INDEX idx_repo_last_modified ON repo_file_cache(last_modified);

CREATE TRIGGER repo_file_cache_updated
    BEFORE UPDATE ON repo_file_cache
    FOR EACH ROW
    EXECUTE FUNCTION update_architecture_rules_timestamp();

COMMENT ON TABLE repo_file_cache IS 'SOLACE repo state cache - query via REST API only';

-- GitHub Instruction Queue: SOLACE's atomic commands for GitHub
CREATE TABLE IF NOT EXISTS github_instruction_queue (
    id SERIAL PRIMARY KEY,
    parent_task_id INTEGER,                    -- Links to original user request
    instruction_sequence INTEGER NOT NULL,     -- Order: 1, 2, 3...
    instruction_text TEXT NOT NULL,            -- Exact atomic instruction for GitHub
    target_file_path TEXT,                     -- File to create/modify
    expected_outcome TEXT,                     -- What success looks like
    status VARCHAR(20) DEFAULT 'pending',      -- pending, in_progress, completed, failed, verified
    github_response TEXT,                      -- GitHub's output after executing
    verification_notes TEXT,                   -- SOLACE's verification of GitHub's work
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    verified_at TIMESTAMP
);

CREATE INDEX idx_github_instruction_status ON github_instruction_queue(status);
CREATE INDEX idx_github_instruction_sequence ON github_instruction_queue(instruction_sequence);
CREATE INDEX idx_github_parent_task ON github_instruction_queue(parent_task_id);

COMMENT ON TABLE github_instruction_queue IS 'SOLACE breaks user requests into atomic GitHub instructions, tracks execution and verification';

-- User Requests: High-level tasks that SOLACE breaks into GitHub instructions
CREATE TABLE IF NOT EXISTS solace_user_requests (
    id SERIAL PRIMARY KEY,
    request_text TEXT NOT NULL,                -- Original user request: "integrate agent dashboard into UI"
    request_type VARCHAR(50),                  -- feature_integration, bug_fix, refactor, new_feature
    complexity_score INTEGER,                  -- 1-10: How complex is this task?
    architecture_rules_used TEXT[],            -- Which architecture_rules patterns apply
    files_affected TEXT[],                     -- List of files that will change
    estimated_instructions INTEGER,            -- How many GitHub instructions needed
    status VARCHAR(20) DEFAULT 'analyzing',    -- analyzing, planned, executing, verifying, complete, failed
    analysis_notes TEXT,                       -- SOLACE's breakdown of the request
    final_outcome TEXT,                        -- Summary of what was accomplished
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_solace_request_status ON solace_user_requests(status);
CREATE INDEX idx_solace_request_type ON solace_user_requests(request_type);

COMMENT ON TABLE solace_user_requests IS 'High-level user requests that SOLACE orchestrates by generating atomic GitHub instructions';
