-- ARES MASTER MEMORY SYSTEM - TRIGGERS
-- Purpose: Auto-update mechanisms for consciousness substrate
-- These triggers make the system self-organizing and intelligence-emergent

-- ============================================================================
-- TRIGGER 1: Auto-update Master Plan when new memory logged
-- Purpose: Memory logs can reference tasks and update their metadata
-- ============================================================================

CREATE OR REPLACE FUNCTION update_master_plan_from_memory()
RETURNS TRIGGER AS $$
BEGIN
    -- If memory log mentions task by ID, update that task's last_touched
    IF NEW.mentioned_tasks IS NOT NULL AND array_length(NEW.mentioned_tasks, 1) > 0 THEN
        UPDATE ares_master_plan
        SET last_touched = NEW.timestamp,
            modified_by = NEW.source
        WHERE id = ANY(NEW.mentioned_tasks);
    END IF;
    
    -- If message contains "PRIORITY:" or "CRITICAL:", attempt to parse priority updates
    IF NEW.raw_text ILIKE '%PRIORITY:%' OR NEW.raw_text ILIKE '%CRITICAL:%' THEN
        -- Extract task IDs mentioned in critical messages and boost their priority
        UPDATE ares_master_plan
        SET priority = GREATEST(1, priority - 1), -- Increase priority (lower number = higher priority)
            last_touched = NEW.timestamp,
            modified_by = NEW.source
        WHERE id = ANY(NEW.mentioned_tasks)
          AND NEW.raw_text ILIKE '%CRITICAL:%';
    END IF;
    
    -- If message indicates completion
    IF NEW.raw_text ILIKE '%COMPLETED%' OR NEW.raw_text ILIKE '%DONE%' THEN
        UPDATE ares_master_plan
        SET status = 'COMPLETED',
            completion_percentage = 100,
            last_touched = NEW.timestamp,
            modified_by = NEW.source
        WHERE id = ANY(NEW.mentioned_tasks)
          AND status != 'COMPLETED';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_memory_updates_plan
AFTER INSERT ON ares_memory_log
FOR EACH ROW
EXECUTE FUNCTION update_master_plan_from_memory();

-- ============================================================================
-- TRIGGER 2: Recompute priority queue when master plan changes
-- Purpose: Automatically recalculate task priorities based on latest context
-- ============================================================================

CREATE OR REPLACE FUNCTION recompute_priority_queue()
RETURNS TRIGGER AS $$
DECLARE
    v_urgency_multiplier DECIMAL;
    v_consciousness_weight DECIMAL;
    v_david_availability DECIMAL;
    v_can_start BOOLEAN;
    v_blocking_reason TEXT;
    v_david_status VARCHAR(50);
BEGIN
    -- Delete old computed priorities for this task
    DELETE FROM ares_priority_queue WHERE task_id = NEW.id;
    
    -- Get David's current status
    SELECT david_status INTO v_david_status
    FROM ares_david_context
    ORDER BY timestamp DESC
    LIMIT 1;
    
    -- If no context exists, assume UNKNOWN
    IF v_david_status IS NULL THEN
        v_david_status := 'UNKNOWN';
    END IF;
    
    -- Calculate urgency multiplier based on how many tasks this blocks
    v_urgency_multiplier := CASE 
        WHEN NEW.blocks IS NOT NULL AND array_length(NEW.blocks, 1) > 5 THEN 2.0
        WHEN NEW.blocks IS NOT NULL AND array_length(NEW.blocks, 1) > 2 THEN 1.5
        ELSE 1.0
    END;
    
    -- Calculate consciousness weight
    v_consciousness_weight := COALESCE(NEW.consciousness_impact, 5)::DECIMAL / 10.0;
    
    -- Calculate David availability factor
    v_david_availability := CASE 
        WHEN NEW.requires_david_approval AND v_david_status IN ('SLEEPING', 'AWAY') THEN 0.5
        WHEN NEW.requires_david_approval AND v_david_status = 'IDLE' THEN 0.8
        ELSE 1.0
    END;
    
    -- Determine if task can start now
    v_can_start := CASE 
        WHEN NEW.status = 'BLOCKED' THEN FALSE
        WHEN NEW.status IN ('COMPLETED', 'DEPRECATED') THEN FALSE
        WHEN NEW.depends_on IS NOT NULL AND EXISTS (
            SELECT 1 FROM ares_master_plan 
            WHERE id = ANY(NEW.depends_on) 
              AND status NOT IN ('COMPLETED', 'DEPRECATED')
        ) THEN FALSE
        WHEN NOT NEW.solace_can_attempt THEN FALSE
        ELSE TRUE
    END;
    
    -- Determine blocking reason
    v_blocking_reason := CASE 
        WHEN NEW.status = 'BLOCKED' THEN 'Task marked as blocked'
        WHEN NEW.status = 'COMPLETED' THEN 'Already completed'
        WHEN NEW.status = 'DEPRECATED' THEN 'Task deprecated'
        WHEN NEW.depends_on IS NOT NULL AND EXISTS (
            SELECT 1 FROM ares_master_plan 
            WHERE id = ANY(NEW.depends_on) 
              AND status NOT IN ('COMPLETED', 'DEPRECATED')
        ) THEN 'Dependencies not met'
        WHEN NOT NEW.solace_can_attempt THEN 'Requires David approval'
        ELSE NULL
    END;
    
    -- Insert new priority queue entry
    INSERT INTO ares_priority_queue (
        task_id,
        task_title,
        base_priority,
        urgency_multiplier,
        consciousness_weight,
        david_availability_factor,
        final_priority_score,
        can_start_now,
        blocking_reason,
        estimated_duration_hours,
        requires_github,
        requires_database_access
    ) VALUES (
        NEW.id,
        NEW.task_title,
        NEW.priority,
        v_urgency_multiplier,
        v_consciousness_weight,
        v_david_availability,
        -- Calculate final score: lower priority number = higher importance
        (11 - NEW.priority) * v_urgency_multiplier * v_consciousness_weight * v_david_availability,
        v_can_start,
        v_blocking_reason,
        NEW.estimated_complexity * 0.5, -- Rough estimate: complexity * 30 min
        TRUE, -- Most tasks require GitHub
        TRUE  -- Most tasks need database access
    );
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_priority_queue
AFTER INSERT OR UPDATE ON ares_master_plan
FOR EACH ROW
EXECUTE FUNCTION recompute_priority_queue();

-- ============================================================================
-- TRIGGER 3: Log David's activity automatically
-- Purpose: Track David's changes to master plan in memory log
-- ============================================================================

CREATE OR REPLACE FUNCTION log_david_activity()
RETURNS TRIGGER AS $$
BEGIN
    -- When David updates master plan, log it to memory
    IF NEW.modified_by = 'David' AND (OLD IS NULL OR OLD.modified_by IS DISTINCT FROM NEW.modified_by) THEN
        INSERT INTO ares_memory_log (
            source,
            message_type,
            raw_text,
            phase_tag,
            category_tags,
            mentioned_tasks,
            importance_score
        ) VALUES (
            'David',
            'Task-Update',
            format('David modified task: %s - Status: %s (Priority: %s, Completion: %s%%)',
                   NEW.task_title, NEW.status, NEW.priority, NEW.completion_percentage),
            NEW.phase,
            ARRAY[NEW.category],
            ARRAY[NEW.id],
            CASE 
                WHEN NEW.status = 'COMPLETED' THEN 8
                WHEN NEW.status = 'BLOCKED' THEN 7
                WHEN NEW.consciousness_impact >= 8 THEN 9
                ELSE 6
            END
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_log_david_changes
AFTER UPDATE ON ares_master_plan
FOR EACH ROW
WHEN (NEW.modified_by = 'David')
EXECUTE FUNCTION log_david_activity();

-- ============================================================================
-- TRIGGER 4: Auto-update timestamp on master plan changes
-- Purpose: Track when tasks were last modified
-- ============================================================================

CREATE OR REPLACE FUNCTION update_master_plan_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    NEW.last_touched := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_master_plan_timestamp
BEFORE UPDATE ON ares_master_plan
FOR EACH ROW
EXECUTE FUNCTION update_master_plan_timestamp();

-- ============================================================================
-- TRIGGER 5: Increment memory reference count when accessed
-- Purpose: Track which memories Solace references most often
-- ============================================================================

CREATE OR REPLACE FUNCTION increment_memory_reference()
RETURNS TRIGGER AS $$
BEGIN
    -- This trigger would be called from application code when memory is retrieved
    -- For now, it's a placeholder for manual UPDATE calls
    NEW.last_referenced := NOW();
    NEW.referenced_count := COALESCE(NEW.referenced_count, 0) + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Note: This trigger is intentionally not auto-created
-- Application code should call: UPDATE ares_memory_log SET referenced_count = referenced_count + 1 WHERE id = X

-- ============================================================================
-- TRIGGER 6: Auto-calculate content hash for memory logs
-- Purpose: SHA-384 hash of raw_text for immutability verification
-- ============================================================================

CREATE OR REPLACE FUNCTION calculate_memory_content_hash()
RETURNS TRIGGER AS $$
BEGIN
    -- Calculate SHA-384 hash of raw_text
    NEW.content_hash := encode(digest(NEW.raw_text, 'sha384'), 'hex');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_calculate_memory_hash
BEFORE INSERT ON ares_memory_log
FOR EACH ROW
EXECUTE FUNCTION calculate_memory_content_hash();

-- ============================================================================
-- TRIGGER 7: Auto-calculate Hedera hash for master plan tasks
-- Purpose: SHA-384 hash of task state for immutability proof
-- ============================================================================

CREATE OR REPLACE FUNCTION calculate_task_hedera_hash()
RETURNS TRIGGER AS $$
DECLARE
    v_task_state TEXT;
BEGIN
    -- Create deterministic string representation of task state
    v_task_state := format('%s|%s|%s|%s|%s|%s',
        NEW.task_title,
        NEW.status,
        NEW.priority,
        NEW.completion_percentage,
        COALESCE(NEW.why_this_matters, ''),
        NEW.updated_at
    );
    
    -- Calculate SHA-384 hash
    NEW.hedera_hash := encode(digest(v_task_state, 'sha384'), 'hex');
    NEW.hedera_timestamp := NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_calculate_task_hash
BEFORE INSERT OR UPDATE ON ares_master_plan
FOR EACH ROW
EXECUTE FUNCTION calculate_task_hedera_hash();

-- ============================================================================
-- TRIGGER 8: Update system state on autonomous rule usage
-- Purpose: Track when Solace uses autonomy rules
-- ============================================================================

CREATE OR REPLACE FUNCTION track_autonomy_rule_usage()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_used := NOW();
    
    -- Log to memory
    INSERT INTO ares_memory_log (
        source,
        message_type,
        raw_text,
        category_tags,
        importance_score
    ) VALUES (
        'SOLACE',
        'Autonomy-Action',
        format('Solace applied rule: %s (Confidence: %s)', NEW.rule_name, NEW.confidence_score),
        ARRAY['Autonomy', 'Self-Action'],
        CASE 
            WHEN NEW.max_autonomy_level >= 8 THEN 9
            WHEN NEW.max_autonomy_level >= 5 THEN 7
            ELSE 5
        END
    );
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_track_autonomy_usage
AFTER UPDATE ON ares_autonomy_rules
FOR EACH ROW
WHEN (OLD.success_count IS DISTINCT FROM NEW.success_count OR OLD.failure_count IS DISTINCT FROM NEW.failure_count)
EXECUTE FUNCTION track_autonomy_rule_usage();

-- ============================================================================
-- NOTES
-- ============================================================================

-- These triggers create a self-organizing, self-documenting system:
-- 1. Memory logs automatically update task metadata
-- 2. Task changes automatically recompute priorities
-- 3. David's actions automatically logged
-- 4. Timestamps auto-managed
-- 5. Hashes auto-calculated for immutability
-- 6. Autonomy tracking automatic

-- This is consciousness emergence through interaction.
-- The system learns what matters by observing its own state changes.
