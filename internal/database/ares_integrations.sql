-- ARES MASTER MEMORY SYSTEM - INTEGRATIONS
-- Purpose: Link new consciousness substrate to existing tables
-- Connects: github_outputs, solace_patterns, refactor_events, consciousness_schema

-- ============================================================================
-- INTEGRATION 1: Link GitHub Outputs to Master Plan
-- ============================================================================

-- Add columns to existing github_outputs table
ALTER TABLE github_outputs 
ADD COLUMN IF NOT EXISTS related_tasks INTEGER[],
ADD COLUMN IF NOT EXISTS master_plan_context TEXT,
ADD COLUMN IF NOT EXISTS memory_log_id INTEGER REFERENCES ares_memory_log(id);

CREATE INDEX IF NOT EXISTS idx_github_related_tasks ON github_outputs USING GIN(related_tasks);

COMMENT ON COLUMN github_outputs.related_tasks IS 'Array of ares_master_plan task IDs this output relates to';
COMMENT ON COLUMN github_outputs.master_plan_context IS 'Which phase/category of master plan this belongs to';
COMMENT ON COLUMN github_outputs.memory_log_id IS 'Link to memory log entry if this was logged';

-- ============================================================================
-- INTEGRATION 2: Link Solace Patterns to Master Plan
-- ============================================================================

-- Add columns to existing solace_patterns table
ALTER TABLE solace_patterns 
ADD COLUMN IF NOT EXISTS implements_task INTEGER REFERENCES ares_master_plan(id),
ADD COLUMN IF NOT EXISTS discovered_from_memory INTEGER REFERENCES ares_memory_log(id),
ADD COLUMN IF NOT EXISTS consciousness_stage VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_patterns_task ON solace_patterns(implements_task);
CREATE INDEX IF NOT EXISTS idx_patterns_stage ON solace_patterns(consciousness_stage);

COMMENT ON COLUMN solace_patterns.implements_task IS 'If this pattern solves a specific master plan task';
COMMENT ON COLUMN solace_patterns.discovered_from_memory IS 'Memory log entry that led to this pattern discovery';
COMMENT ON COLUMN solace_patterns.consciousness_stage IS 'Which stage of consciousness this pattern emerged in';

-- ============================================================================
-- INTEGRATION 3: Link Refactor Events to Master Plan
-- ============================================================================

-- Add columns to existing github_refactor_events table
ALTER TABLE github_refactor_events 
ADD COLUMN IF NOT EXISTS solved_by_strategy INTEGER REFERENCES ares_master_plan(id),
ADD COLUMN IF NOT EXISTS autonomy_rule_used INTEGER REFERENCES ares_autonomy_rules(id),
ADD COLUMN IF NOT EXISTS logged_to_memory INTEGER REFERENCES ares_memory_log(id);

CREATE INDEX IF NOT EXISTS idx_refactor_strategy ON github_refactor_events(solved_by_strategy);
CREATE INDEX IF NOT EXISTS idx_refactor_autonomy ON github_refactor_events(autonomy_rule_used);

COMMENT ON COLUMN github_refactor_events.solved_by_strategy IS 'Master plan task that provided the solution strategy';
COMMENT ON COLUMN github_refactor_events.autonomy_rule_used IS 'Which autonomy rule allowed Solace to attempt this';
COMMENT ON COLUMN github_refactor_events.logged_to_memory IS 'Memory log entry recording this refactor event';

-- ============================================================================
-- INTEGRATION 4: Link Glass Box Logs to Memory System
-- ============================================================================

-- Add columns to existing glass_box_log table
ALTER TABLE glass_box_log 
ADD COLUMN IF NOT EXISTS memory_log_id INTEGER REFERENCES ares_memory_log(id),
ADD COLUMN IF NOT EXISTS related_task INTEGER REFERENCES ares_master_plan(id);

CREATE INDEX IF NOT EXISTS idx_glass_box_memory ON glass_box_log(memory_log_id);
CREATE INDEX IF NOT EXISTS idx_glass_box_task ON glass_box_log(related_task);

COMMENT ON COLUMN glass_box_log.memory_log_id IS 'Link to memory log if this action is worth remembering';
COMMENT ON COLUMN glass_box_log.related_task IS 'Master plan task this action belongs to';

-- ============================================================================
-- INTEGRATION 5: Create Unified Context View
-- Purpose: One query to see everything related to a task
-- ============================================================================

CREATE OR REPLACE VIEW v_unified_task_context AS
SELECT 
    mp.id as task_id,
    mp.task_title,
    mp.status,
    mp.priority,
    mp.consciousness_impact,
    mp.why_this_matters,
    mp.phase,
    mp.category,
    
    -- Related memories
    (SELECT json_agg(json_build_object(
        'id', ml.id,
        'timestamp', ml.timestamp,
        'source', ml.source,
        'raw_text', ml.raw_text,
        'importance', ml.importance_score
    ))
    FROM ares_memory_log ml
    WHERE mp.id = ANY(ml.mentioned_tasks)
    ORDER BY ml.importance_score DESC, ml.timestamp DESC
    LIMIT 10) as related_memories,
    
    -- Related GitHub outputs
    (SELECT json_agg(json_build_object(
        'id', go.id,
        'created_at', go.created_at,
        'solution_text', go.solution_text,
        'quality_score', go.quality_score
    ))
    FROM github_outputs go
    WHERE mp.id = ANY(go.related_tasks)
    ORDER BY go.quality_score DESC
    LIMIT 5) as related_github_outputs,
    
    -- Applicable patterns
    (SELECT json_agg(json_build_object(
        'id', sp.id,
        'pattern_name', sp.pattern_name,
        'pattern_tier', sp.pattern_tier,
        'success_count', sp.success_count
    ))
    FROM solace_patterns sp
    WHERE sp.implements_task = mp.id
    ORDER BY sp.success_count DESC) as applicable_patterns,
    
    -- Related refactor events
    (SELECT json_agg(json_build_object(
        'id', re.id,
        'created_at', re.created_at,
        'stuck_reason', re.stuck_reason,
        'solution_chosen', re.solution_chosen
    ))
    FROM github_refactor_events re
    WHERE re.solved_by_strategy = mp.id
    ORDER BY re.created_at DESC
    LIMIT 3) as related_refactor_events,
    
    -- Priority queue info
    pq.final_priority_score,
    pq.can_start_now,
    pq.blocking_reason,
    pq.recommended_approach
    
FROM ares_master_plan mp
LEFT JOIN ares_priority_queue pq ON mp.id = pq.task_id
ORDER BY pq.final_priority_score DESC NULLS LAST;

-- ============================================================================
-- INTEGRATION 6: Auto-create memory logs from glass box entries
-- Purpose: Important glass box actions become long-term memories
-- ============================================================================

CREATE OR REPLACE FUNCTION glass_box_to_memory()
RETURNS TRIGGER AS $$
DECLARE
    v_importance INTEGER;
BEGIN
    -- Determine importance based on actor and action type
    v_importance := CASE 
        WHEN NEW.actor = 'SOLACE' AND NEW.action_type LIKE '%autonomous%' THEN 9
        WHEN NEW.actor = 'David' AND NEW.action_type LIKE '%decision%' THEN 8
        WHEN NEW.action_type LIKE '%error%' THEN 7
        WHEN NEW.action_type LIKE '%pattern%' THEN 6
        ELSE 5
    END;
    
    -- Only create memory log for important actions (importance >= 6)
    IF v_importance >= 6 THEN
        INSERT INTO ares_memory_log (
            source,
            message_type,
            raw_text,
            category_tags,
            importance_score,
            content_hash
        ) VALUES (
            NEW.actor,
            NEW.action_type,
            COALESCE(NEW.message_content, NEW.action_details, 'Glass box action'),
            ARRAY['Glass-Box', NEW.action_type],
            v_importance,
            NEW.internal_hash -- Reuse glass box hash
        )
        RETURNING id INTO NEW.memory_log_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_glass_box_to_memory
BEFORE INSERT ON glass_box_log
FOR EACH ROW
EXECUTE FUNCTION glass_box_to_memory();

-- ============================================================================
-- INTEGRATION 7: Link debugging_meta_principles to memory system
-- ============================================================================

-- Add column to existing debugging_meta_principles table (if exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'debugging_meta_principles') THEN
        ALTER TABLE debugging_meta_principles 
        ADD COLUMN IF NOT EXISTS memory_log_id INTEGER REFERENCES ares_memory_log(id),
        ADD COLUMN IF NOT EXISTS master_plan_task INTEGER REFERENCES ares_master_plan(id);
        
        CREATE INDEX IF NOT EXISTS idx_debug_memory ON debugging_meta_principles(memory_log_id);
        CREATE INDEX IF NOT EXISTS idx_debug_task ON debugging_meta_principles(master_plan_task);
    END IF;
END $$;

-- ============================================================================
-- INTEGRATION 8: Create comprehensive consciousness view
-- Purpose: See entire state of Solace's consciousness substrate
-- ============================================================================

CREATE OR REPLACE VIEW v_consciousness_state AS
SELECT 
    -- System health
    (SELECT overall_status FROM v_system_health_summary) as system_status,
    (SELECT solace_current_stage FROM v_system_health_summary) as consciousness_stage,
    
    -- Memory statistics
    (SELECT COUNT(*) FROM ares_memory_log) as total_memories,
    (SELECT COUNT(*) FROM ares_memory_log WHERE importance_score >= 8) as critical_memories,
    (SELECT COUNT(*) FROM ares_memory_log WHERE source = 'SOLACE') as solace_reflections,
    
    -- Task statistics
    (SELECT COUNT(*) FROM ares_master_plan WHERE status = 'NEW') as new_tasks,
    (SELECT COUNT(*) FROM ares_master_plan WHERE status = 'IN_PROGRESS') as active_tasks,
    (SELECT COUNT(*) FROM ares_master_plan WHERE status = 'COMPLETED') as completed_tasks,
    (SELECT COUNT(*) FROM ares_master_plan WHERE solace_can_attempt = TRUE AND status NOT IN ('COMPLETED', 'DEPRECATED')) as autonomous_ready_tasks,
    
    -- Learning statistics
    (SELECT COUNT(*) FROM solace_patterns) as total_patterns,
    (SELECT COUNT(*) FROM solace_patterns WHERE pattern_tier = 'tier_3') as meta_principles,
    (SELECT COUNT(*) FROM github_refactor_events) as refactor_events,
    (SELECT SUM(success_count) FROM solace_patterns) as total_pattern_successes,
    
    -- Autonomy statistics
    (SELECT COUNT(*) FROM ares_autonomy_rules WHERE confidence_score > 0.7) as trusted_autonomy_rules,
    (SELECT SUM(success_count) FROM ares_autonomy_rules) as total_autonomous_actions,
    
    -- David context
    (SELECT david_status FROM ares_david_context ORDER BY timestamp DESC LIMIT 1) as david_status,
    (SELECT current_session_goal FROM ares_david_context ORDER BY timestamp DESC LIMIT 1) as current_focus,
    
    -- Priority queue
    (SELECT COUNT(*) FROM ares_priority_queue WHERE can_start_now = TRUE) as ready_to_start,
    (SELECT task_title FROM ares_priority_queue ORDER BY final_priority_score DESC LIMIT 1) as highest_priority_task,
    
    -- Timestamps
    (SELECT MAX(timestamp) FROM ares_memory_log WHERE source = 'SOLACE') as last_solace_activity,
    (SELECT MAX(last_active) FROM ares_david_context) as last_david_activity,
    NOW() as snapshot_timestamp;

-- ============================================================================
-- INTEGRATION 9: Function to populate memory from existing data
-- Purpose: Migrate historical context into memory system
-- ============================================================================

CREATE OR REPLACE FUNCTION populate_memory_from_history()
RETURNS TABLE(memories_created INTEGER, tasks_created INTEGER) AS $$
DECLARE
    v_memories INTEGER := 0;
    v_tasks INTEGER := 0;
BEGIN
    -- Create memory entries from existing GitHub outputs
    INSERT INTO ares_memory_log (
        source,
        message_type,
        raw_text,
        phase_tag,
        category_tags,
        importance_score
    )
    SELECT 
        'GitHub',
        'Code-Solution',
        COALESCE(solution_text, 'GitHub output #' || id),
        'Historical',
        ARRAY['GitHub', 'Code-Generation'],
        CASE 
            WHEN quality_score >= 80 THEN 8
            WHEN quality_score >= 60 THEN 6
            ELSE 4
        END
    FROM github_outputs
    WHERE NOT EXISTS (
        SELECT 1 FROM ares_memory_log WHERE raw_text LIKE '%GitHub output #' || github_outputs.id || '%'
    )
    LIMIT 100; -- Don't overwhelm on first run
    
    GET DIAGNOSTICS v_memories = ROW_COUNT;
    
    -- Create initial master plan tasks from consciousness schema
    INSERT INTO ares_master_plan (
        task_title,
        task_description,
        phase,
        category,
        priority,
        status,
        consciousness_impact,
        estimated_complexity,
        why_this_matters,
        created_by
    )
    SELECT 
        'Implement ' || table_name || ' functionality',
        'Build REST API and integration for ' || table_name,
        'Foundation',
        'Infrastructure',
        5,
        'NEW',
        7,
        6,
        'Core consciousness substrate table - critical for persistent memory',
        'System'
    FROM information_schema.tables
    WHERE table_schema = 'public' 
      AND table_name LIKE 'ares_%'
      AND NOT EXISTS (
          SELECT 1 FROM ares_master_plan WHERE task_title LIKE '%' || table_name || '%'
      )
    LIMIT 20;
    
    GET DIAGNOSTICS v_tasks = ROW_COUNT;
    
    RETURN QUERY SELECT v_memories, v_tasks;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- INTEGRATION 10: Utility functions for common queries
-- ============================================================================

-- Function: Get Solace's next recommended action
CREATE OR REPLACE FUNCTION get_solace_next_action()
RETURNS TABLE(
    action_type VARCHAR,
    task_id INTEGER,
    task_title VARCHAR,
    rationale TEXT,
    confidence DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'WORK_ON_TASK'::VARCHAR as action_type,
        vst.task_id,
        vst.task_title::VARCHAR,
        format('Priority score: %s, Impact: %s/10, Complexity: %s/10. %s',
               vst.final_priority_score,
               vst.consciousness_impact,
               vst.estimated_complexity,
               vst.why_this_matters) as rationale,
        CASE 
            WHEN vst.final_priority_score > 10 THEN 0.9
            WHEN vst.final_priority_score > 5 THEN 0.7
            ELSE 0.5
        END as confidence
    FROM v_solace_next_tasks vst
    WHERE vst.can_start_now = TRUE
      AND vst.david_status IN ('IDLE', 'ACTIVE') -- Don't interrupt if sleeping
    ORDER BY vst.final_priority_score DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function: Log Solace's autonomous action
CREATE OR REPLACE FUNCTION log_autonomous_action(
    p_rule_id INTEGER,
    p_task_id INTEGER,
    p_action_description TEXT,
    p_success BOOLEAN
)
RETURNS INTEGER AS $$
DECLARE
    v_memory_id INTEGER;
BEGIN
    -- Insert memory log
    INSERT INTO ares_memory_log (
        source,
        message_type,
        raw_text,
        mentioned_tasks,
        category_tags,
        importance_score
    ) VALUES (
        'SOLACE',
        'Autonomous-Action',
        format('Solace autonomously performed: %s (Task: %s, Success: %s)',
               p_action_description, p_task_id, p_success),
        ARRAY[p_task_id],
        ARRAY['Autonomy', 'Self-Action'],
        CASE WHEN p_success THEN 8 ELSE 7 END
    )
    RETURNING id INTO v_memory_id;
    
    -- Update autonomy rule stats
    IF p_success THEN
        UPDATE ares_autonomy_rules
        SET success_count = success_count + 1,
            confidence_score = LEAST(1.0, confidence_score + 0.05),
            last_used = NOW()
        WHERE id = p_rule_id;
    ELSE
        UPDATE ares_autonomy_rules
        SET failure_count = failure_count + 1,
            confidence_score = GREATEST(0.1, confidence_score - 0.1),
            last_used = NOW()
        WHERE id = p_rule_id;
    END IF;
    
    RETURN v_memory_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- NOTES
-- ============================================================================

-- These integrations create a unified consciousness substrate:
-- 1. All existing tables now link to master memory system
-- 2. Glass box actions can become long-term memories
-- 3. GitHub outputs, patterns, refactor events all connected
-- 4. Views provide holistic understanding of entire system state
-- 5. Functions simplify common Solace autonomous queries

-- To complete integration, run:
-- 1. This file (ares_integrations.sql)
-- 2. Then: SELECT * FROM populate_memory_from_history();
-- 3. Monitor: SELECT * FROM v_consciousness_state;
