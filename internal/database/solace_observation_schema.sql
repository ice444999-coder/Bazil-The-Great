-- ==========================================
-- SOLACE OBSERVATION TABLES
-- ==========================================

-- UI State Log (what SOLACE "sees")
CREATE TABLE IF NOT EXISTS ui_state_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    component_type VARCHAR(50) NOT NULL,
    element_id VARCHAR(100),
    state_snapshot JSONB NOT NULL,
    visibility BOOLEAN DEFAULT TRUE,
    user_focused BOOLEAN DEFAULT FALSE,
    session_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Data Stream Log (market data observations)
CREATE TABLE IF NOT EXISTS data_stream_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    stream_type VARCHAR(50) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    data_payload JSONB NOT NULL,
    processed_data JSONB,
    displayed BOOLEAN DEFAULT TRUE,
    session_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- User Actions (clicks, inputs, trades)
CREATE TABLE IF NOT EXISTS user_actions (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    action_type VARCHAR(50) NOT NULL,
    target_element VARCHAR(100),
    input_value TEXT,
    context JSONB,
    solace_observation TEXT,
    session_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- SOLACE Decisions (autonomous actions)
CREATE TABLE IF NOT EXISTS solace_decisions (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    decision_type VARCHAR(50) NOT NULL,
    reasoning TEXT NOT NULL,
    confidence_score FLOAT,
    action_taken JSONB NOT NULL,
    outcome JSONB,
    user_approved BOOLEAN,
    session_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Code Modifications (SOLACE self-modification tracking)
CREATE TABLE IF NOT EXISTS code_modifications (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    modification_type VARCHAR(50) NOT NULL,
    target_component VARCHAR(100),
    original_code TEXT,
    modified_code TEXT,
    reason TEXT,
    active BOOLEAN DEFAULT TRUE,
    session_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ==========================================
-- INDEXES FOR PERFORMANCE
-- ==========================================

CREATE INDEX IF NOT EXISTS idx_ui_state_timestamp ON ui_state_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_ui_state_session ON ui_state_log(session_id);
CREATE INDEX IF NOT EXISTS idx_ui_state_component ON ui_state_log(component_type);

CREATE INDEX IF NOT EXISTS idx_data_stream_timestamp ON data_stream_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_data_stream_symbol ON data_stream_log(symbol);
CREATE INDEX IF NOT EXISTS idx_data_stream_session ON data_stream_log(session_id);
CREATE INDEX IF NOT EXISTS idx_data_stream_type ON data_stream_log(stream_type);

CREATE INDEX IF NOT EXISTS idx_user_actions_timestamp ON user_actions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_user_actions_session ON user_actions(session_id);
CREATE INDEX IF NOT EXISTS idx_user_actions_type ON user_actions(action_type);

CREATE INDEX IF NOT EXISTS idx_solace_decisions_timestamp ON solace_decisions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_solace_decisions_session ON solace_decisions(session_id);
CREATE INDEX IF NOT EXISTS idx_solace_decisions_type ON solace_decisions(decision_type);

-- ==========================================
-- ANALYTICS VIEWS
-- ==========================================

-- View: Recent observations summary
CREATE OR REPLACE VIEW recent_observations AS
SELECT 
    'ui_state' as source,
    component_type as type,
    timestamp,
    session_id
FROM ui_state_log
UNION ALL
SELECT 
    'data_stream' as source,
    stream_type as type,
    timestamp,
    session_id
FROM data_stream_log
UNION ALL
SELECT 
    'user_action' as source,
    action_type as type,
    timestamp,
    session_id
FROM user_actions
ORDER BY timestamp DESC
LIMIT 100;

-- View: Session activity summary
CREATE OR REPLACE VIEW session_activity AS
SELECT 
    session_id,
    MIN(timestamp) as session_start,
    MAX(timestamp) as session_end,
    COUNT(*) as total_events,
    COUNT(DISTINCT CASE WHEN component_type IS NOT NULL THEN component_type END) as ui_events,
    COUNT(DISTINCT CASE WHEN action_type IS NOT NULL THEN action_type END) as user_actions
FROM (
    SELECT session_id, timestamp, component_type, NULL as action_type FROM ui_state_log
    UNION ALL
    SELECT session_id, timestamp, NULL, action_type FROM user_actions
) combined
GROUP BY session_id
ORDER BY session_start DESC;

-- ==========================================
-- SAMPLE DATA (Optional - for testing)
-- ==========================================

-- Insert a test observation
INSERT INTO ui_state_log (component_type, state_snapshot, session_id)
VALUES (
    'system_initialized',
    '{"version": "1.0.0", "upgrade": "SOLACE integration", "timestamp": "2025-01-14T00:00:00Z"}',
    'SYSTEM-INIT-001'
);

-- ==========================================
-- VERIFY INSTALLATION
-- ==========================================

-- Check that all tables exist
DO $$
DECLARE
    table_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO table_count
    FROM information_schema.tables
    WHERE table_schema = 'public'
    AND table_name IN ('ui_state_log', 'data_stream_log', 'user_actions', 'solace_decisions', 'code_modifications');
    
    IF table_count = 5 THEN
        RAISE NOTICE '✅ All SOLACE tables created successfully';
    ELSE
        RAISE NOTICE '⚠️ Only % out of 5 tables created', table_count;
    END IF;
END $$;

-- Show table sizes
SELECT 
    table_name,
    pg_size_pretty(pg_total_relation_size(quote_ident(table_name))) as size
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN ('ui_state_log', 'data_stream_log', 'user_actions', 'solace_decisions', 'code_modifications', 'sandbox_trades')
ORDER BY pg_total_relation_size(quote_ident(table_name)) DESC;

-- ==========================================
-- MAINTENANCE FUNCTIONS
-- ==========================================

-- Function to clean old observations (keep last 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_observations()
RETURNS void AS $$
BEGIN
    DELETE FROM ui_state_log WHERE timestamp < NOW() - INTERVAL '30 days';
    DELETE FROM data_stream_log WHERE timestamp < NOW() - INTERVAL '30 days';
    DELETE FROM user_actions WHERE timestamp < NOW() - INTERVAL '30 days';
    
    RAISE NOTICE 'Cleaned observations older than 30 days';
END;
$$ LANGUAGE plpgsql;

-- ==========================================
-- POST-MIGRATION CHECKS
-- ==========================================

-- Verify indexes
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
AND tablename IN ('ui_state_log', 'data_stream_log', 'user_actions')
ORDER BY tablename, indexname;

-- Show row counts
SELECT 
    'ui_state_log' as table_name, 
    COUNT(*) as row_count 
FROM ui_state_log
UNION ALL
SELECT 'data_stream_log', COUNT(*) FROM data_stream_log
UNION ALL
SELECT 'user_actions', COUNT(*) FROM user_actions
UNION ALL
SELECT 'solace_decisions', COUNT(*) FROM solace_decisions
UNION ALL
SELECT 'sandbox_trades', COUNT(*) FROM sandbox_trades;
