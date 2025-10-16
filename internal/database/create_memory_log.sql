-- ARES Memory Log Table (without pgvector - semantic search ready for later)

CREATE TABLE ares_memory_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT NOW(),
    source VARCHAR(100),
    message_type VARCHAR(100),
    raw_text TEXT NOT NULL,
    phase_tag VARCHAR(100),
    category_tags TEXT[],
    mentioned_files TEXT[],
    mentioned_tasks INTEGER[],
    key_concepts TEXT[],
    importance_score INTEGER CHECK (importance_score BETWEEN 1 AND 10),
    referenced_count INTEGER DEFAULT 0,
    last_referenced TIMESTAMP,
    content_hash VARCHAR(128),
    hedera_hash VARCHAR(128),
    hedera_timestamp TIMESTAMP
);

CREATE INDEX idx_timestamp ON ares_memory_log(timestamp DESC);
CREATE INDEX idx_phase_category ON ares_memory_log(phase_tag);
CREATE INDEX idx_importance ON ares_memory_log(importance_score DESC);
CREATE INDEX idx_source ON ares_memory_log(source);
CREATE INDEX idx_key_concepts ON ares_memory_log USING GIN(key_concepts);

-- Trigger: Memory updates Master Plan
CREATE OR REPLACE FUNCTION update_master_plan_from_memory()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.mentioned_tasks IS NOT NULL THEN
        UPDATE ares_master_plan
        SET last_touched = NEW.timestamp, modified_by = NEW.source
        WHERE id = ANY(NEW.mentioned_tasks);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_memory_updates_plan
AFTER INSERT ON ares_memory_log
FOR EACH ROW EXECUTE FUNCTION update_master_plan_from_memory();

-- Trigger: Recompute Priority Queue
CREATE OR REPLACE FUNCTION recompute_priority_queue()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM ares_priority_queue WHERE task_id = NEW.id;
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
        CASE 
            WHEN array_length(NEW.blocks, 1) > 5 THEN 2.0 
            WHEN array_length(NEW.blocks, 1) > 2 THEN 1.5 
            ELSE 1.0 
        END,
        COALESCE(NEW.consciousness_impact::DECIMAL, 5.0) / 10.0,
        CASE 
            WHEN NEW.requires_david_approval AND 
                 EXISTS(SELECT 1 FROM ares_david_context WHERE david_status = 'SLEEPING' ORDER BY timestamp DESC LIMIT 1) 
            THEN 0.5 
            ELSE 1.0 
        END,
        NEW.priority * 
            CASE 
                WHEN array_length(NEW.blocks, 1) > 5 THEN 2.0 
                WHEN array_length(NEW.blocks, 1) > 2 THEN 1.5 
                ELSE 1.0 
            END * 
            COALESCE(NEW.consciousness_impact::DECIMAL, 5.0) / 10.0,
        CASE 
            WHEN NEW.status = 'BLOCKED' THEN FALSE
            WHEN NEW.depends_on IS NOT NULL AND 
                 EXISTS (SELECT 1 FROM ares_master_plan WHERE id = ANY(NEW.depends_on) AND status != 'COMPLETED') 
            THEN FALSE
            WHEN NOT NEW.solace_can_attempt THEN FALSE
            ELSE TRUE 
        END,
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

-- Trigger: Calculate content hash
CREATE OR REPLACE FUNCTION calculate_memory_content_hash()
RETURNS TRIGGER AS $$
BEGIN
    NEW.content_hash := encode(digest(NEW.raw_text, 'sha384'), 'hex');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_calculate_content_hash
BEFORE INSERT OR UPDATE OF raw_text ON ares_memory_log
FOR EACH ROW EXECUTE FUNCTION calculate_memory_content_hash();

-- Trigger: Increment reference count
CREATE OR REPLACE FUNCTION increment_memory_reference()
RETURNS TRIGGER AS $$
BEGIN
    NEW.referenced_count := OLD.referenced_count + 1;
    NEW.last_referenced := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- View: Recent important memories
CREATE OR REPLACE VIEW v_recent_important_memories AS
SELECT 
    id,
    timestamp,
    source,
    message_type,
    LEFT(raw_text, 200) as preview,
    phase_tag,
    key_concepts,
    importance_score,
    referenced_count
FROM ares_memory_log
WHERE importance_score >= 7
ORDER BY timestamp DESC
LIMIT 100;

-- View: SOLACE's next tasks
CREATE OR REPLACE VIEW v_solace_next_tasks AS
SELECT 
    pq.task_id,
    pq.task_title,
    pq.final_priority_score,
    pq.can_start_now,
    pq.blocking_reason,
    mp.phase,
    mp.status,
    mp.consciousness_impact,
    mp.complexity_estimate,
    mp.description
FROM ares_priority_queue pq
JOIN ares_master_plan mp ON pq.task_id = mp.id
WHERE pq.can_start_now = TRUE
ORDER BY pq.final_priority_score DESC
LIMIT 10;
