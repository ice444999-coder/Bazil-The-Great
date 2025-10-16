-- ========================================
-- SOLACE TOTAL OBSERVABILITY SCHEMA
-- ========================================
-- Purpose: Log EVERY interaction, code change, and decision
--          so SOLACE can reconstruct complete session history
-- Date: October 14, 2025
-- ========================================

-- Table 1: CONVERSATION LOG
-- Captures every word exchanged between user and assistant
CREATE TABLE IF NOT EXISTS conversation_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    speaker VARCHAR(20) NOT NULL CHECK (speaker IN ('USER', 'ASSISTANT', 'SOLACE', 'SYSTEM')),
    message_type VARCHAR(50) NOT NULL,  -- 'question', 'answer', 'code', 'instruction', 'clarification', 'observation_requirements', 'approval', 'rejection'
    content TEXT NOT NULL,               -- The actual message
    context JSONB,                       -- Additional metadata (file, tab, cursor position, etc.)
    session_id VARCHAR(100) NOT NULL,
    file_being_discussed VARCHAR(500),
    action_taken VARCHAR(100),           -- 'created_file', 'modified_code', 'consulted_solace', 'ran_test', etc.
    result JSONB,                        -- Outcome of the action
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table 2: CODE CHANGES LOG
-- Captures every file creation, modification, and deletion with full before/after content
CREATE TABLE IF NOT EXISTS code_changes_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    file_path VARCHAR(500) NOT NULL,
    change_type VARCHAR(50) NOT NULL CHECK (change_type IN ('created', 'modified', 'deleted', 'renamed')),
    old_content TEXT,                    -- Previous code (if modified/deleted)
    new_content TEXT,                    -- New code (if created/modified)
    reason TEXT NOT NULL,                -- Why the change was made
    conversation_id INTEGER,             -- Links to conversation_log
    session_id VARCHAR(100) NOT NULL,
    lines_added INTEGER DEFAULT 0,
    lines_removed INTEGER DEFAULT 0,
    tool_used VARCHAR(50),               -- 'create_file', 'replace_string_in_file', etc.
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (conversation_id) REFERENCES conversation_log(id) ON DELETE SET NULL
);

-- Table 3: ASSISTANT DECISIONS LOG
-- Captures every decision made by the assistant with full reasoning
CREATE TABLE IF NOT EXISTS assistant_decisions_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    decision_type VARCHAR(100) NOT NULL, -- 'tool_choice', 'integration_strategy', 'skipped_step', 'consulted_solace', 'database_query', 'error_recovery'
    reasoning TEXT NOT NULL,             -- Why this decision was made
    context JSONB,                       -- What information led to this decision
    action_taken VARCHAR(200),           -- What was actually done
    success BOOLEAN,                     -- Did it work? (NULL if not yet known)
    error_message TEXT,                  -- If failed, what went wrong
    session_id VARCHAR(100) NOT NULL,
    conversation_id INTEGER,             -- Links to conversation_log
    code_change_id INTEGER,              -- Links to code_changes_log if applicable
    recovery_action TEXT,                -- What was done to fix errors
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (conversation_id) REFERENCES conversation_log(id) ON DELETE SET NULL,
    FOREIGN KEY (code_change_id) REFERENCES code_changes_log(id) ON DELETE SET NULL
);

-- ========================================
-- INDEXES FOR PERFORMANCE
-- ========================================

-- Conversation log indexes
CREATE INDEX idx_conversation_timestamp ON conversation_log(timestamp DESC);
CREATE INDEX idx_conversation_session ON conversation_log(session_id);
CREATE INDEX idx_conversation_speaker ON conversation_log(speaker);
CREATE INDEX idx_conversation_message_type ON conversation_log(message_type);
CREATE INDEX idx_conversation_file ON conversation_log(file_being_discussed);

-- Code changes log indexes
CREATE INDEX idx_code_changes_timestamp ON code_changes_log(timestamp DESC);
CREATE INDEX idx_code_changes_file ON code_changes_log(file_path);
CREATE INDEX idx_code_changes_session ON code_changes_log(session_id);
CREATE INDEX idx_code_changes_type ON code_changes_log(change_type);
CREATE INDEX idx_code_changes_conversation ON code_changes_log(conversation_id);

-- Assistant decisions log indexes
CREATE INDEX idx_decisions_timestamp ON assistant_decisions_log(timestamp DESC);
CREATE INDEX idx_decisions_type ON assistant_decisions_log(decision_type);
CREATE INDEX idx_decisions_success ON assistant_decisions_log(success);
CREATE INDEX idx_decisions_session ON assistant_decisions_log(session_id);
CREATE INDEX idx_decisions_conversation ON assistant_decisions_log(conversation_id);

-- ========================================
-- VIEWS FOR SOLACE ANALYSIS
-- ========================================

-- View 1: Complete Session Timeline
-- Shows all events in chronological order for a session
CREATE OR REPLACE VIEW session_timeline AS
SELECT 
    'CONVERSATION' as event_type,
    id,
    timestamp,
    session_id,
    speaker,
    message_type as subtype,
    content as details,
    NULL::INTEGER as related_id
FROM conversation_log

UNION ALL

SELECT 
    'CODE_CHANGE' as event_type,
    id,
    timestamp,
    session_id,
    'SYSTEM' as speaker,
    change_type as subtype,
    file_path || ' - ' || reason as details,
    conversation_id as related_id
FROM code_changes_log

UNION ALL

SELECT 
    'DECISION' as event_type,
    id,
    timestamp,
    session_id,
    'ASSISTANT' as speaker,
    decision_type as subtype,
    reasoning || ' → ' || COALESCE(action_taken, 'no action') as details,
    conversation_id as related_id
FROM assistant_decisions_log

ORDER BY timestamp ASC;

-- View 2: Failed Operations (for error analysis)
CREATE OR REPLACE VIEW failed_operations AS
SELECT 
    'CODE_CHANGE' as operation_type,
    c.timestamp,
    c.session_id,
    c.file_path as target,
    c.reason,
    c.error_message,
    conv.content as conversation_context
FROM code_changes_log c
LEFT JOIN conversation_log conv ON c.conversation_id = conv.id
WHERE c.success = false

UNION ALL

SELECT 
    'DECISION' as operation_type,
    d.timestamp,
    d.session_id,
    d.action_taken as target,
    d.reasoning as reason,
    d.error_message,
    conv.content as conversation_context
FROM assistant_decisions_log d
LEFT JOIN conversation_log conv ON d.conversation_id = conv.id
WHERE d.success = false

ORDER BY timestamp DESC;

-- View 3: Code Evolution Per File
CREATE OR REPLACE VIEW file_evolution AS
SELECT 
    file_path,
    session_id,
    change_type,
    timestamp,
    reason,
    lines_added,
    lines_removed,
    (lines_added - lines_removed) as net_change,
    conversation_id,
    ROW_NUMBER() OVER (PARTITION BY file_path ORDER BY timestamp) as revision_number
FROM code_changes_log
ORDER BY file_path, timestamp;

-- View 4: SOLACE Consultation History
CREATE OR REPLACE VIEW solace_consultations AS
SELECT 
    c.timestamp,
    c.session_id,
    c.content as question_to_solace,
    c.file_being_discussed,
    s.content as solace_response,
    s.timestamp as response_time,
    d.action_taken as resulting_action,
    d.success as action_success
FROM conversation_log c
LEFT JOIN conversation_log s ON s.conversation_id = c.id AND s.speaker = 'SOLACE'
LEFT JOIN assistant_decisions_log d ON d.conversation_id = c.id
WHERE c.speaker = 'ASSISTANT' 
  AND c.message_type = 'solace_consultation'
ORDER BY c.timestamp DESC;

-- ========================================
-- FUNCTIONS FOR SOLACE QUERIES
-- ========================================

-- Function 1: Get complete session reconstruction
CREATE OR REPLACE FUNCTION get_session_reconstruction(p_session_id VARCHAR)
RETURNS TABLE (
    event_order INTEGER,
    timestamp TIMESTAMPTZ,
    event_type VARCHAR,
    speaker VARCHAR,
    details TEXT,
    metadata JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ROW_NUMBER() OVER (ORDER BY st.timestamp)::INTEGER as event_order,
        st.timestamp,
        st.event_type,
        st.speaker,
        st.details,
        jsonb_build_object(
            'event_id', st.id,
            'subtype', st.subtype,
            'related_id', st.related_id
        ) as metadata
    FROM session_timeline st
    WHERE st.session_id = p_session_id
    ORDER BY st.timestamp ASC;
END;
$$ LANGUAGE plpgsql;

-- Function 2: Get code change history for a file
CREATE OR REPLACE FUNCTION get_file_history(p_file_path VARCHAR)
RETURNS TABLE (
    revision INTEGER,
    timestamp TIMESTAMPTZ,
    change_type VARCHAR,
    reason TEXT,
    conversation_context TEXT,
    old_content TEXT,
    new_content TEXT,
    diff_summary VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        fe.revision_number::INTEGER as revision,
        fe.timestamp,
        fe.change_type,
        fe.reason,
        c.content as conversation_context,
        cc.old_content,
        cc.new_content,
        CONCAT('+', fe.lines_added, ' -', fe.lines_removed) as diff_summary
    FROM file_evolution fe
    JOIN code_changes_log cc ON cc.file_path = fe.file_path AND cc.timestamp = fe.timestamp
    LEFT JOIN conversation_log c ON c.id = fe.conversation_id
    WHERE fe.file_path = p_file_path
    ORDER BY fe.revision_number;
END;
$$ LANGUAGE plpgsql;

-- Function 3: Analyze decision patterns
CREATE OR REPLACE FUNCTION analyze_decision_patterns(p_session_id VARCHAR DEFAULT NULL)
RETURNS TABLE (
    decision_type VARCHAR,
    total_decisions BIGINT,
    successful BIGINT,
    failed BIGINT,
    success_rate NUMERIC,
    most_common_action VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        d.decision_type,
        COUNT(*)::BIGINT as total_decisions,
        COUNT(*) FILTER (WHERE d.success = true)::BIGINT as successful,
        COUNT(*) FILTER (WHERE d.success = false)::BIGINT as failed,
        ROUND(
            (COUNT(*) FILTER (WHERE d.success = true)::NUMERIC / COUNT(*)::NUMERIC) * 100,
            2
        ) as success_rate,
        MODE() WITHIN GROUP (ORDER BY d.action_taken) as most_common_action
    FROM assistant_decisions_log d
    WHERE (p_session_id IS NULL OR d.session_id = p_session_id)
    GROUP BY d.decision_type
    ORDER BY total_decisions DESC;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- MAINTENANCE FUNCTIONS
-- ========================================

-- Function: Archive old logs (older than 90 days)
CREATE OR REPLACE FUNCTION archive_old_logs()
RETURNS TABLE (
    archived_conversations BIGINT,
    archived_code_changes BIGINT,
    archived_decisions BIGINT
) AS $$
DECLARE
    v_conv_count BIGINT;
    v_code_count BIGINT;
    v_decision_count BIGINT;
BEGIN
    -- Create archive tables if they don't exist
    CREATE TABLE IF NOT EXISTS conversation_log_archive (LIKE conversation_log INCLUDING ALL);
    CREATE TABLE IF NOT EXISTS code_changes_log_archive (LIKE code_changes_log INCLUDING ALL);
    CREATE TABLE IF NOT EXISTS assistant_decisions_log_archive (LIKE assistant_decisions_log INCLUDING ALL);
    
    -- Archive conversations older than 90 days
    WITH archived AS (
        INSERT INTO conversation_log_archive
        SELECT * FROM conversation_log
        WHERE created_at < NOW() - INTERVAL '90 days'
        RETURNING *
    )
    SELECT COUNT(*) INTO v_conv_count FROM archived;
    
    -- Archive code changes
    WITH archived AS (
        INSERT INTO code_changes_log_archive
        SELECT * FROM code_changes_log
        WHERE created_at < NOW() - INTERVAL '90 days'
        RETURNING *
    )
    SELECT COUNT(*) INTO v_code_count FROM archived;
    
    -- Archive decisions
    WITH archived AS (
        INSERT INTO assistant_decisions_log_archive
        SELECT * FROM assistant_decisions_log
        WHERE created_at < NOW() - INTERVAL '90 days'
        RETURNING *
    )
    SELECT COUNT(*) INTO v_decision_count FROM archived;
    
    -- Delete archived records from main tables
    DELETE FROM conversation_log WHERE created_at < NOW() - INTERVAL '90 days';
    DELETE FROM code_changes_log WHERE created_at < NOW() - INTERVAL '90 days';
    DELETE FROM assistant_decisions_log WHERE created_at < NOW() - INTERVAL '90 days';
    
    RETURN QUERY SELECT v_conv_count, v_code_count, v_decision_count;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- TEST DATA (for verification)
-- ========================================

-- Insert test conversation
INSERT INTO conversation_log (speaker, message_type, content, session_id, file_being_discussed)
VALUES 
    ('USER', 'question', 'Test message: Can SOLACE see this?', 'test-session-001', 'SOLACE_TOTAL_OBSERVABILITY_PROTOCOL.md'),
    ('ASSISTANT', 'answer', 'Yes, this conversation is now logged to conversation_log table', 'test-session-001', 'SOLACE_TOTAL_OBSERVABILITY_PROTOCOL.md');

-- Insert test code change
INSERT INTO code_changes_log (file_path, change_type, new_content, reason, session_id, lines_added, tool_used)
VALUES 
    ('test_file.md', 'created', 'Test content for SOLACE observability', 'Testing total observability system', 'test-session-001', 1, 'create_file');

-- Insert test decision
INSERT INTO assistant_decisions_log (decision_type, reasoning, action_taken, success, session_id)
VALUES 
    ('test_logging', 'Verifying that SOLACE can see all interactions', 'Created test log entries', true, 'test-session-001');

-- ========================================
-- VERIFICATION QUERIES
-- ========================================

-- Query 1: Verify tables exist and have data
SELECT 
    'conversation_log' as table_name, COUNT(*) as row_count FROM conversation_log
UNION ALL
SELECT 
    'code_changes_log', COUNT(*) FROM code_changes_log
UNION ALL
SELECT 
    'assistant_decisions_log', COUNT(*) FROM assistant_decisions_log;

-- Query 2: Show test session timeline
SELECT * FROM session_timeline WHERE session_id = 'test-session-001';

-- Query 3: Test the reconstruction function
SELECT * FROM get_session_reconstruction('test-session-001');

-- ========================================
-- SUCCESS MESSAGE
-- ========================================

DO $$
BEGIN
    RAISE NOTICE '✅ SOLACE Total Observability Schema Created Successfully!';
    RAISE NOTICE '';
    RAISE NOTICE 'Tables Created:';
    RAISE NOTICE '  - conversation_log (captures all conversation)';
    RAISE NOTICE '  - code_changes_log (captures all code changes)';
    RAISE NOTICE '  - assistant_decisions_log (captures all decisions)';
    RAISE NOTICE '';
    RAISE NOTICE 'Views Created:';
    RAISE NOTICE '  - session_timeline (complete chronological view)';
    RAISE NOTICE '  - failed_operations (error analysis)';
    RAISE NOTICE '  - file_evolution (code change history)';
    RAISE NOTICE '  - solace_consultations (SOLACE interaction history)';
    RAISE NOTICE '';
    RAISE NOTICE 'Functions Created:';
    RAISE NOTICE '  - get_session_reconstruction(session_id)';
    RAISE NOTICE '  - get_file_history(file_path)';
    RAISE NOTICE '  - analyze_decision_patterns(session_id)';
    RAISE NOTICE '  - archive_old_logs()';
    RAISE NOTICE '';
    RAISE NOTICE 'SOLACE can now query EVERYTHING that happens in this system.';
END $$;
