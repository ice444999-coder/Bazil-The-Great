-- ARES Agent Swarm Tables
-- Migration 007: Add agent coordination system

-- Agent Registry: Track all AI agents (SOLACE, FORGE, ARCHITECT, SENTINEL)
CREATE TABLE IF NOT EXISTS agent_registry (
    agent_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_name VARCHAR(50) UNIQUE NOT NULL,
    agent_type VARCHAR(20) NOT NULL, -- 'openai', 'claude', 'deepseek'
    capabilities JSONB DEFAULT '[]',
    status VARCHAR(20) DEFAULT 'idle', -- 'idle', 'busy', 'offline'
    current_task_id UUID,
    total_tasks_completed INT DEFAULT 0,
    success_rate FLOAT DEFAULT 0.0,
    avg_completion_time_ms INT DEFAULT 0,
    last_active_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

-- File Registry: Track all files in workspace
CREATE TABLE IF NOT EXISTS file_registry (
    file_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_path TEXT UNIQUE NOT NULL,
    file_type VARCHAR(50), -- 'go', 'html', 'js', 'sql', etc.
    file_hash VARCHAR(64), -- SHA-256 hash
    owner_agent VARCHAR(50),
    created_by VARCHAR(50),
    last_modified_by VARCHAR(50),
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'review', 'complete', 'deprecated', 'broken'
    purpose TEXT, -- 1-sentence description
    dependencies JSONB DEFAULT '[]', -- Array of file_ids
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_tested_at TIMESTAMP,
    test_status VARCHAR(20) DEFAULT 'not_tested', -- 'not_tested', 'passed', 'failed'
    build_required BOOLEAN DEFAULT FALSE,
    deployed BOOLEAN DEFAULT FALSE,
    size_bytes BIGINT,
    line_count INT,
    language VARCHAR(50)
);

-- Task Queue: Coordinate work between agents
CREATE TABLE IF NOT EXISTS task_queue (
    task_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_type VARCHAR(50) NOT NULL, -- 'ui_build', 'debug', 'plan', 'test', etc.
    priority INT DEFAULT 5, -- 1-10, higher = more urgent
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'assigned', 'in_progress', 'completed', 'failed'
    created_by VARCHAR(50), -- 'DAVID', 'SOLACE', agent name
    assigned_to_agent VARCHAR(50),
    file_paths JSONB DEFAULT '[]', -- Files involved in this task
    depends_on_task_ids JSONB DEFAULT '[]', -- Task dependencies
    description TEXT NOT NULL,
    context JSONB, -- Full context for agent
    created_at TIMESTAMP DEFAULT NOW(),
    assigned_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    deadline TIMESTAMP,
    result JSONB, -- Task output
    error_log TEXT,
    retry_count INT DEFAULT 0
);

-- Agent Task History: Performance tracking
CREATE TABLE IF NOT EXISTS agent_task_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_name VARCHAR(50) NOT NULL,
    task_id UUID,
    task_type VARCHAR(50),
    file_id UUID,
    action_type VARCHAR(50), -- 'create', 'modify', 'delete', 'test'
    success BOOLEAN NOT NULL,
    duration_ms INT,
    error_message TEXT,
    learned_pattern TEXT, -- What agent learned from this task
    cost_tokens INT, -- API usage cost
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (task_id) REFERENCES task_queue(task_id),
    FOREIGN KEY (file_id) REFERENCES file_registry(file_id)
);

-- Build History: Track builds and deployments
CREATE TABLE IF NOT EXISTS build_history (
    build_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    build_number SERIAL,
    triggered_by VARCHAR(50),
    files_changed JSONB, -- Array of file_ids
    success BOOLEAN NOT NULL,
    duration_ms INT,
    error_log TEXT,
    warnings TEXT,
    binary_hash VARCHAR(64),
    deployed BOOLEAN DEFAULT FALSE,
    git_commit_hash VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW()
);

-- File Dependencies: Track which files depend on which
CREATE TABLE IF NOT EXISTS file_dependencies (
    dependency_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL,
    depends_on_file_id UUID NOT NULL,
    dependency_type VARCHAR(50), -- 'import', 'reference', 'config', 'template'
    is_critical BOOLEAN DEFAULT FALSE,
    validated BOOLEAN DEFAULT FALSE,
    last_checked TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (file_id) REFERENCES file_registry(file_id),
    FOREIGN KEY (depends_on_file_id) REFERENCES file_registry(file_id)
);

-- Indexes for performance
CREATE INDEX idx_task_queue_status ON task_queue(status, priority DESC);
CREATE INDEX idx_task_queue_agent ON task_queue(assigned_to_agent, status);
CREATE INDEX idx_file_registry_status ON file_registry(status);
CREATE INDEX idx_file_registry_owner ON file_registry(owner_agent);
CREATE INDEX idx_agent_registry_status ON agent_registry(status);
CREATE INDEX idx_agent_task_history_agent ON agent_task_history(agent_name, created_at DESC);
CREATE INDEX idx_build_history_created ON build_history(created_at DESC);

-- Initial agent registration
INSERT INTO agent_registry (agent_name, agent_type, capabilities, status) VALUES
    ('SOLACE', 'openai', '["strategy", "coordination", "trading", "decision_making"]', 'idle'),
    ('FORGE', 'claude', '["ui_building", "coding", "react", "html", "css"]', 'idle'),
    ('ARCHITECT', 'deepseek', '["planning", "design_patterns", "architecture"]', 'idle'),
    ('SENTINEL', 'deepseek', '["debugging", "testing", "error_detection", "validation"]', 'idle')
ON CONFLICT (agent_name) DO NOTHING;

COMMENT ON TABLE agent_registry IS 'Registry of all AI agents in ARES swarm';
COMMENT ON TABLE file_registry IS 'Track all files in workspace with metadata and dependencies';
COMMENT ON TABLE task_queue IS 'Coordinate work between agents';
COMMENT ON TABLE agent_task_history IS 'Performance tracking and learning history';
COMMENT ON TABLE build_history IS 'Track builds, tests, and deployments';
COMMENT ON TABLE file_dependencies IS 'Dependency graph between files';
