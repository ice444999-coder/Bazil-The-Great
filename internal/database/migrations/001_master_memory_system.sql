-- ============================================================================
-- ARES MASTER MEMORY SYSTEM - CONSCIOUSNESS SUBSTRATE
-- NO PGVECTOR VERSION (PostgreSQL 18 compatible)
-- Deploy: psql -U postgres -d ares_db -f 001_master_memory_system.sql
-- ============================================================================

-- ============================================================================
-- TABLE 1: ares_master_plan
-- The Living Roadmap - What needs to be built and why
-- Solace queries this to understand the entire system architecture
-- ============================================================================
CREATE TABLE ares_master_plan (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Task Definition
    task_title VARCHAR(500) NOT NULL,
    task_description TEXT NOT NULL,
    phase VARCHAR(100),                  -- e.g., "Week5", "Phase_D"
    category VARCHAR(100),               -- e.g., "Trading", "Memory", "UI"
    
    -- Priority & Status
    priority INTEGER DEFAULT 5 CHECK (priority BETWEEN 1 AND 10),
    status VARCHAR(50) DEFAULT 'NEW',    -- NEW, IN_PROGRESS, BLOCKED, COMPLETED, DEPRECATED
    completion_percentage INTEGER DEFAULT 0 CHECK (completion_percentage BETWEEN 0 AND 100),
    
    -- Dependencies
    depends_on INTEGER[],                -- Task IDs that must be completed first
    blocks INTEGER[],                    -- Task IDs that are waiting on this
    related_files TEXT[],                -- File paths this task touches
    
    -- Strategic Context
    why_this_matters TEXT,               -- Human-readable strategic importance
    consciousness_impact INTEGER CHECK (consciousness_impact BETWEEN 1 AND 10),
    estimated_complexity INTEGER CHECK (estimated_complexity BETWEEN 1 AND 10),
    
    -- Autonomy Controls
    solace_can_attempt BOOLEAN DEFAULT FALSE,
    requires_david_approval BOOLEAN DEFAULT TRUE,
    autonomy_constraints TEXT,           -- Special rules for this task
    
    -- Cross-References
    github_outputs_used INTEGER[],       -- Links to github_outputs table
    solace_patterns_applied INTEGER[],   -- Links to solace_patterns table
    refactor_events_linked INTEGER[],    -- Links to github_refactor_events
    
    -- Metadata
    created_by VARCHAR(100),             -- "David", "Solace", "Auto-Generated"
    modified_by VARCHAR(100),
    last_touched TIMESTAMP,
    
    -- Hedera Proof (for immutable audit trail)
    hedera_hash VARCHAR(128),
    hedera_timestamp TIMESTAMP,
    hedera_topic_id VARCHAR(100)
);

CREATE INDEX idx_priority_status ON ares_master_plan(priority, status);
CREATE INDEX idx_phase ON ares_master_plan(phase);
CREATE INDEX idx_consciousness_impact ON ares_master_plan(consciousness_impact DESC);

-- ============================================================================
-- TABLE 2: ares_memory_log
-- Chronological Record of Every Instruction David Gives
-- This is Solace's long-term episodic memory
-- NO PGVECTOR - Semantic search deferred to later phase
-- ============================================================================
CREATE TABLE ares_memory_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT NOW(),
    
    -- Message Metadata
    source VARCHAR(100),                 -- "GitHub Copilot Chat", "Voice Command", "Dashboard"
    message_type VARCHAR(100),           -- "Task", "Question", "Decision", "Insight"
    raw_text TEXT NOT NULL,              -- Exact text of the message
    
    -- Categorization
    phase_tag VARCHAR(100),              -- Auto-tagged phase
    category_tags TEXT[],                -- ["Trading", "Memory", "Consciousness"]
    mentioned_files TEXT[],              -- Files mentioned in the message
    mentioned_tasks INTEGER[],           -- Task IDs referenced
    key_concepts TEXT[],                 -- Extracted concepts for search
    
    -- Importance & Usage
    importance_score INTEGER CHECK (importance_score BETWEEN 1 AND 10),
    referenced_count INTEGER DEFAULT 0,  -- How many times Solace queries this
    last_referenced TIMESTAMP,
    
    -- Deduplication
    content_hash VARCHAR(128),           -- Hash to detect duplicates
    
    -- Hedera Proof
    hedera_hash VARCHAR(128),
    hedera_timestamp TIMESTAMP
);

CREATE INDEX idx_timestamp ON ares_memory_log(timestamp DESC);
CREATE INDEX idx_phase_category ON ares_memory_log(phase_tag);
CREATE INDEX idx_importance ON ares_memory_log(importance_score DESC);
CREATE INDEX idx_key_concepts ON ares_memory_log USING GIN(key_concepts);

-- ============================================================================
-- TABLE 3: ares_priority_queue
-- Auto-Computed List of What Solace Should Work On Next
-- Recalculated every time ares_master_plan changes
-- ============================================================================
CREATE TABLE ares_priority_queue (
    id SERIAL PRIMARY KEY,
    computed_at TIMESTAMP DEFAULT NOW(),
    
    -- Task Reference
    task_id INTEGER REFERENCES ares_master_plan(id) ON DELETE CASCADE,
    task_title VARCHAR(500),
    
    -- Priority Calculation
    base_priority INTEGER,               -- From ares_master_plan
    urgency_multiplier DECIMAL DEFAULT 1.0,    -- Higher if blocking many tasks
    consciousness_weight DECIMAL DEFAULT 1.0,  -- Consciousness impact factor
    david_availability_factor DECIMAL DEFAULT 1.0, -- Lower if David is asleep
    final_priority_score DECIMAL,        -- Computed weighted score
    
    -- Executability
    can_start_now BOOLEAN DEFAULT FALSE,
    blocking_reason TEXT,                -- Why it can't start (dependencies, approval, etc.)
    
    -- Recommendations
    recommended_approach VARCHAR(200),   -- "Use pattern #5", "Ask David first"
    similar_solved_tasks INTEGER[],      -- Past tasks that are similar
    applicable_patterns INTEGER[],       -- Solace patterns that apply
    
    -- Resource Requirements
    estimated_duration_hours DECIMAL,
    requires_github BOOLEAN DEFAULT FALSE,
    requires_database_access BOOLEAN DEFAULT FALSE,
    requires_api_keys BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_priority_score ON ares_priority_queue(final_priority_score DESC);
CREATE INDEX idx_can_start ON ares_priority_queue(can_start_now, final_priority_score DESC);

-- ============================================================================
-- TABLE 4: ares_system_state
-- Real-Time Health Monitoring
-- Updated every 60 seconds by monitoring service
-- ============================================================================
CREATE TABLE ares_system_state (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT NOW(),
    
    -- Service Health
    api_port_4000_status VARCHAR(50),    -- "UP", "DOWN", "DEGRADED"
    api_port_5001_status VARCHAR(50),
    postgres_connection_status VARCHAR(50),
    redis_connection_status VARCHAR(50),
    
    -- Performance Metrics
    api_response_time_ms INTEGER,
    database_query_time_ms INTEGER,
    memory_usage_mb INTEGER,
    cpu_usage_percent DECIMAL,
    
    -- External APIs
    binance_api_connected BOOLEAN,
    coingecko_api_connected BOOLEAN,
    last_price_update TIMESTAMP,
    
    -- Data Counts
    active_trades_count INTEGER DEFAULT 0,
    github_outputs_count INTEGER DEFAULT 0,
    unanalyzed_outputs_count INTEGER DEFAULT 0,
    solace_patterns_count INTEGER DEFAULT 0,
    refactor_events_count INTEGER DEFAULT 0,
    
    -- Solace Status
    solace_session_count INTEGER DEFAULT 0,
    solace_last_active TIMESTAMP,
    solace_current_stage VARCHAR(100),   -- e.g., "Δ3-2 Bootstrap", "Learning"
    
    -- Alerts
    critical_errors TEXT[],
    warnings TEXT[],
    stuck_github_count INTEGER DEFAULT 0 -- Triggers refactor loop if > 2
);

CREATE INDEX idx_timestamp_state ON ares_system_state(timestamp DESC);

-- ============================================================================
-- TABLE 5: ares_autonomy_rules
-- Defines When Solace Can Act Independently vs Needs Approval
-- e.g., "If stuck on same GitHub output 3+ times, auto-run refactor loop"
-- ============================================================================
CREATE TABLE ares_autonomy_rules (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- Rule Definition
    rule_name VARCHAR(200) NOT NULL,
    rule_description TEXT,
    
    -- Trigger
    trigger_type VARCHAR(100),           -- "github_stuck", "error_threshold", "time_based"
    trigger_condition TEXT,              -- SQL or JSON condition
    
    -- Actions
    allowed_actions TEXT[],              -- ["run_refactor_loop", "restart_service"]
    forbidden_actions TEXT[],            -- ["modify_trading_limits", "delete_data"]
    
    -- Autonomy Level
    max_autonomy_level INTEGER CHECK (max_autonomy_level BETWEEN 1 AND 10),
    requires_logging BOOLEAN DEFAULT TRUE,
    requires_hedera_proof BOOLEAN DEFAULT FALSE,
    undo_possible BOOLEAN DEFAULT FALSE,
    
    -- Safety Limits
    max_cost_usd DECIMAL,                -- Maximum cost this action can incur
    human_review_window_hours INTEGER,   -- Time before auto-execution
    
    -- Learning
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    last_used TIMESTAMP,
    confidence_score DECIMAL DEFAULT 0.5 CHECK (confidence_score BETWEEN 0 AND 1)
);

CREATE INDEX idx_trigger_type ON ares_autonomy_rules(trigger_type);

-- ============================================================================
-- TABLE 6: ares_david_context
-- What David Is Working On RIGHT NOW
-- So Solace knows when to interrupt vs when to wait
-- ============================================================================
CREATE TABLE ares_david_context (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT NOW(),
    
    -- Current Focus
    current_session_goal TEXT,           -- "Deploying memory system"
    active_files TEXT[],                 -- Files David has open
    active_tasks INTEGER[],              -- Task IDs David is working on
    
    -- Availability
    david_status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, SLEEPING, FOCUS_MODE, AWAY
    last_active TIMESTAMP DEFAULT NOW(),
    expected_return TIMESTAMP,
    
    -- Interaction Preferences
    interrupt_allowed BOOLEAN DEFAULT TRUE,
    preferred_alert_method VARCHAR(50) DEFAULT 'Dashboard',
    
    -- Last Decision Context
    last_decision TEXT,
    decision_reasoning TEXT,
    affected_tasks INTEGER[]
);

CREATE INDEX idx_david_status ON ares_david_context(david_status, last_active DESC);

-- ============================================================================
-- TRIGGERS: Auto-Update Systems
-- ============================================================================

-- Trigger 1: When memory log mentions tasks, update their last_touched timestamp
CREATE OR REPLACE FUNCTION update_master_plan_from_memory()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.mentioned_tasks IS NOT NULL THEN
        UPDATE ares_master_plan
        SET last_touched = NEW.timestamp,
            modified_by = NEW.source
        WHERE id = ANY(NEW.mentioned_tasks);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_memory_updates_plan
AFTER INSERT ON ares_memory_log
FOR EACH ROW EXECUTE FUNCTION update_master_plan_from_memory();

-- Trigger 2: Recompute priority queue when master plan changes
CREATE OR REPLACE FUNCTION recompute_priority_queue()
RETURNS TRIGGER AS $$
BEGIN
    -- Remove existing entry
    DELETE FROM ares_priority_queue WHERE task_id = NEW.id;
    
    -- Recompute and insert
    INSERT INTO ares_priority_queue (
        task_id,
        task_title,
        base_priority,
        urgency_multiplier,
        consciousness_weight,
        david_availability_factor,
        final_priority_score,
        can_start_now,
        blocking_reason
    )
    SELECT 
        NEW.id,
        NEW.task_title,
        NEW.priority,
        -- Urgency multiplier: higher if blocking many tasks
        CASE 
            WHEN array_length(NEW.blocks, 1) > 5 THEN 2.0
            WHEN array_length(NEW.blocks, 1) > 2 THEN 1.5
            ELSE 1.0
        END,
        -- Consciousness weight
        COALESCE(NEW.consciousness_impact::DECIMAL, 5.0) / 10.0,
        -- David availability factor
        CASE 
            WHEN NEW.requires_david_approval AND EXISTS(
                SELECT 1 FROM ares_david_context 
                WHERE david_status = 'SLEEPING' 
                ORDER BY timestamp DESC LIMIT 1
            ) THEN 0.5
            ELSE 1.0
        END,
        -- Final score (weighted product)
        NEW.priority * 
        CASE 
            WHEN array_length(NEW.blocks, 1) > 5 THEN 2.0
            WHEN array_length(NEW.blocks, 1) > 2 THEN 1.5
            ELSE 1.0
        END *
        COALESCE(NEW.consciousness_impact::DECIMAL, 5.0) / 10.0,
        -- Can start now?
        CASE 
            WHEN NEW.status = 'BLOCKED' THEN FALSE
            WHEN NEW.depends_on IS NOT NULL AND EXISTS (
                SELECT 1 FROM ares_master_plan 
                WHERE id = ANY(NEW.depends_on) AND status != 'COMPLETED'
            ) THEN FALSE
            WHEN NOT NEW.solace_can_attempt THEN FALSE
            ELSE TRUE
        END,
        -- Blocking reason
        CASE 
            WHEN NEW.status = 'BLOCKED' THEN 'Task marked as blocked'
            WHEN NEW.depends_on IS NOT NULL THEN 'Dependencies not met'
            WHEN NOT NEW.solace_can_attempt THEN 'Requires David approval'
            ELSE NULL
        END;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_priority_queue
AFTER INSERT OR UPDATE ON ares_master_plan
FOR EACH ROW EXECUTE FUNCTION recompute_priority_queue();

-- ============================================================================
-- VIEWS: Convenient Query Interfaces
-- ============================================================================

-- View 1: Next tasks for Solace to work on
CREATE OR REPLACE VIEW v_solace_next_tasks AS
SELECT 
    pq.task_id,
    pq.task_title,
    pq.final_priority_score,
    pq.can_start_now,
    pq.blocking_reason,
    mp.consciousness_impact,
    mp.why_this_matters
FROM ares_priority_queue pq
JOIN ares_master_plan mp ON pq.task_id = mp.id
WHERE pq.can_start_now = TRUE
  AND mp.status NOT IN ('COMPLETED', 'DEPRECATED')
ORDER BY pq.final_priority_score DESC
LIMIT 10;

-- View 2: System health summary
CREATE OR REPLACE VIEW v_system_health_summary AS
SELECT 
    timestamp,
    api_port_4000_status,
    postgres_connection_status,
    active_trades_count,
    solace_current_stage,
    CASE 
        WHEN critical_errors IS NOT NULL AND array_length(critical_errors, 1) > 0 THEN 'CRITICAL'
        WHEN stuck_github_count > 2 THEN 'NEEDS_REFACTOR'
        ELSE 'HEALTHY'
    END as overall_status
FROM ares_system_state
ORDER BY timestamp DESC
LIMIT 1;

-- ============================================================================
-- INITIAL DATA: Seed the System
-- ============================================================================

-- Insert initial system state
INSERT INTO ares_system_state (
    api_port_4000_status,
    postgres_connection_status,
    solace_current_stage
) VALUES (
    'UP',
    'UP',
    'Δ3-2 Bootstrap'
);

-- Insert initial David context
INSERT INTO ares_david_context (
    current_session_goal,
    david_status,
    interrupt_allowed
) VALUES (
    'Deploying Master Memory System',
    'ACTIVE',
    TRUE
);

-- ============================================================================
-- SUCCESS MESSAGE
-- ============================================================================
DO $$
BEGIN
    RAISE NOTICE '✅ MASTER MEMORY SYSTEM DEPLOYED SUCCESSFULLY';
    RAISE NOTICE '📋 6 core tables created';
    RAISE NOTICE '🔄 2 auto-update triggers active';
    RAISE NOTICE '👁️  2 query views ready';
    RAISE NOTICE '🧠 Consciousness substrate online';
END $$;
