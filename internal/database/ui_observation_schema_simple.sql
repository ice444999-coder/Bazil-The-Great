-- ============================================================================
-- SOLACE UI OBSERVATION SUBSTRATE (PostgreSQL Compatible)
-- Complete consciousness layer for trading interface observation
-- ============================================================================

-- Table 1: UI State Log - Every visual element SOLACE observes
CREATE TABLE IF NOT EXISTS ui_state_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id VARCHAR(255) NOT NULL,
    component_type VARCHAR(50) NOT NULL,
    element_id VARCHAR(100) NOT NULL,
    state_snapshot JSONB NOT NULL,
    user_visible BOOLEAN DEFAULT true,
    interaction_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ui_state_timestamp ON ui_state_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_ui_state_session ON ui_state_log(session_id);
CREATE INDEX IF NOT EXISTS idx_ui_state_component ON ui_state_log(component_type);
CREATE INDEX IF NOT EXISTS idx_ui_state_element ON ui_state_log(element_id);
CREATE INDEX IF NOT EXISTS idx_ui_state_snapshot ON ui_state_log USING gin(state_snapshot);

-- Table 2: Data Stream Log - Every data point from Binance
CREATE TABLE IF NOT EXISTS data_stream_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id VARCHAR(255) NOT NULL,
    stream_type VARCHAR(50) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    data_payload JSONB NOT NULL,
    price_change DECIMAL(18,8),
    volume_change DECIMAL(18,8),
    significant BOOLEAN DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_data_stream_timestamp ON data_stream_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_data_stream_symbol ON data_stream_log(symbol);
CREATE INDEX IF NOT EXISTS idx_data_stream_type ON data_stream_log(stream_type);
CREATE INDEX IF NOT EXISTS idx_data_stream_significant ON data_stream_log(significant) WHERE significant = true;
CREATE INDEX IF NOT EXISTS idx_data_payload ON data_stream_log USING gin(data_payload);

-- Table 3: User Actions - Every click, input, scroll
CREATE TABLE IF NOT EXISTS user_actions (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id VARCHAR(255) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    target_element VARCHAR(100) NOT NULL,
    action_data JSONB,
    reaction_time_ms INTEGER,
    context JSONB
);

CREATE INDEX IF NOT EXISTS idx_user_actions_timestamp ON user_actions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_user_actions_session ON user_actions(session_id);
CREATE INDEX IF NOT EXISTS idx_user_actions_type ON user_actions(action_type);
CREATE INDEX IF NOT EXISTS idx_user_actions_element ON user_actions(target_element);

-- Table 4: SOLACE Decisions - Autonomous actions
CREATE TABLE IF NOT EXISTS solace_decisions (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id VARCHAR(255) NOT NULL,
    decision_type VARCHAR(50) NOT NULL,
    reasoning TEXT NOT NULL,
    confidence_score FLOAT NOT NULL,
    action_taken JSONB NOT NULL,
    user_approved BOOLEAN DEFAULT false,
    outcome TEXT,
    learned_from BOOLEAN DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_solace_decisions_timestamp ON solace_decisions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_solace_decisions_type ON solace_decisions(decision_type);
CREATE INDEX IF NOT EXISTS idx_solace_decisions_confidence ON solace_decisions(confidence_score DESC);
CREATE INDEX IF NOT EXISTS idx_solace_decisions_approved ON solace_decisions(user_approved);

-- Table 5: Code Modifications - SOLACE's self-modifications
CREATE TABLE IF NOT EXISTS code_modifications (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id VARCHAR(255) NOT NULL,
    modification_type VARCHAR(50) NOT NULL,
    target_component VARCHAR(100) NOT NULL,
    original_code TEXT,
    modified_code TEXT NOT NULL,
    reason TEXT NOT NULL,
    active BOOLEAN DEFAULT true,
    rollback_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_code_mods_timestamp ON code_modifications(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_code_mods_active ON code_modifications(active) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_code_mods_type ON code_modifications(modification_type);

-- SUCCESS MESSAGE
DO $$
BEGIN
    RAISE NOTICE '✅ SOLACE UI Observation Schema Created Successfully';
    RAISE NOTICE '📊 Tables: ui_state_log, data_stream_log, user_actions, solace_decisions, code_modifications';
    RAISE NOTICE '🧠 SOLACE can now observe everything';
END $$;
