-- ========================================
-- SEMANTIC MEMORY ARCHITECTURE UPGRADE
-- ========================================
-- This migration transforms ARES from naive "load all memories" to
-- intelligent semantic retrieval with vector embeddings and hierarchy
-- ========================================

-- Enable pgvector extension for vector embeddings
CREATE EXTENSION IF NOT EXISTS vector;

-- ========================================
-- 1. MEMORY EMBEDDINGS TABLE
-- ========================================
-- Store vector embeddings for semantic search
CREATE TABLE memory_embeddings (
    id SERIAL PRIMARY KEY,
    snapshot_id INTEGER REFERENCES memory_snapshots(id) ON DELETE CASCADE,
    embedding vector(384),  -- Using all-MiniLM-L6-v2 (384 dimensions)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast semantic search (cosine similarity)
CREATE INDEX idx_memory_embeddings_vector ON memory_embeddings
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Index for joining with snapshots
CREATE INDEX idx_memory_embeddings_snapshot_id ON memory_embeddings(snapshot_id);

-- ========================================
-- 2. MEMORY METADATA (Enhanced Indexing)
-- ========================================
-- Add metadata columns to memory_snapshots for intelligent retrieval
ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS importance_score FLOAT DEFAULT 0.5;
ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS access_count INTEGER DEFAULT 0;
ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS last_accessed TIMESTAMP;
ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS memory_type VARCHAR(50) DEFAULT 'general';
ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS tags TEXT[];
ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS compression_level VARCHAR(20) DEFAULT 'none';
ALTER TABLE memory_snapshots ADD COLUMN IF NOT EXISTS archived BOOLEAN DEFAULT FALSE;

-- Indices for smart retrieval
CREATE INDEX IF NOT EXISTS idx_memory_importance ON memory_snapshots(importance_score DESC);
CREATE INDEX IF NOT EXISTS idx_memory_access_count ON memory_snapshots(access_count DESC);
CREATE INDEX IF NOT EXISTS idx_memory_last_accessed ON memory_snapshots(last_accessed DESC);
CREATE INDEX IF NOT EXISTS idx_memory_type ON memory_snapshots(memory_type);
CREATE INDEX IF NOT EXISTS idx_memory_archived ON memory_snapshots(archived);
CREATE INDEX IF NOT EXISTS idx_memory_tags ON memory_snapshots USING GIN(tags);

-- ========================================
-- 3. MEMORY RELATIONSHIPS (Graph)
-- ========================================
-- Track connections between memories for context chains
CREATE TABLE memory_relationships (
    id SERIAL PRIMARY KEY,
    source_snapshot_id INTEGER REFERENCES memory_snapshots(id) ON DELETE CASCADE,
    target_snapshot_id INTEGER REFERENCES memory_snapshots(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50),  -- 'follows', 'related_to', 'causes', 'references'
    strength FLOAT DEFAULT 1.0,  -- 0.0 to 1.0
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_snapshot_id, target_snapshot_id, relationship_type)
);

CREATE INDEX idx_memory_rel_source ON memory_relationships(source_snapshot_id);
CREATE INDEX idx_memory_rel_target ON memory_relationships(target_snapshot_id);
CREATE INDEX idx_memory_rel_type ON memory_relationships(relationship_type);

-- ========================================
-- 4. MEMORY CONSOLIDATION LOG
-- ========================================
-- Track memory merges and compressions
CREATE TABLE memory_consolidations (
    id SERIAL PRIMARY KEY,
    original_snapshot_ids INTEGER[],  -- IDs that were merged
    consolidated_snapshot_id INTEGER REFERENCES memory_snapshots(id),
    consolidation_type VARCHAR(50),  -- 'merge', 'compress', 'summary'
    details JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_consolidations_snapshot ON memory_consolidations(consolidated_snapshot_id);
CREATE INDEX idx_consolidations_type ON memory_consolidations(consolidation_type);

-- ========================================
-- 5. MEMORY CACHE (Hot/Warm/Cold Hierarchy)
-- ========================================
-- Track memory temperature for caching decisions
CREATE TABLE memory_cache_stats (
    snapshot_id INTEGER PRIMARY KEY REFERENCES memory_snapshots(id) ON DELETE CASCADE,
    temperature VARCHAR(10) DEFAULT 'cold',  -- 'hot', 'warm', 'cold'
    cache_hits INTEGER DEFAULT 0,
    cache_misses INTEGER DEFAULT 0,
    last_hit TIMESTAMP,
    promoted_at TIMESTAMP,
    demoted_at TIMESTAMP,
    size_bytes INTEGER,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cache_temperature ON memory_cache_stats(temperature);
CREATE INDEX idx_cache_last_hit ON memory_cache_stats(last_hit DESC);

-- ========================================
-- 6. SMART RETRIEVAL ANALYTICS
-- ========================================
-- Track query patterns to optimize retrieval
CREATE TABLE memory_query_analytics (
    id SERIAL PRIMARY KEY,
    query_text TEXT,
    query_embedding vector(384),
    retrieved_snapshot_ids INTEGER[],
    relevance_scores FLOAT[],
    execution_time_ms INTEGER,
    user_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_query_analytics_user ON memory_query_analytics(user_id);
CREATE INDEX idx_query_analytics_created ON memory_query_analytics(created_at DESC);
CREATE INDEX idx_query_analytics_embedding ON memory_query_analytics
USING ivfflat (query_embedding vector_cosine_ops) WITH (lists = 50);

-- ========================================
-- 7. EMBEDDING GENERATION QUEUE
-- ========================================
-- Queue for async embedding generation
CREATE TABLE embedding_generation_queue (
    id SERIAL PRIMARY KEY,
    snapshot_id INTEGER REFERENCES memory_snapshots(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'processing', 'completed', 'failed'
    retry_count INTEGER DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP
);

CREATE INDEX idx_embedding_queue_status ON embedding_generation_queue(status);
CREATE INDEX idx_embedding_queue_created ON embedding_generation_queue(created_at);

-- ========================================
-- 8. HELPER FUNCTIONS
-- ========================================

-- Function: Calculate memory importance based on multiple factors
CREATE OR REPLACE FUNCTION calculate_memory_importance(
    p_snapshot_id INTEGER
) RETURNS FLOAT AS $$
DECLARE
    v_importance FLOAT;
    v_access_count INTEGER;
    v_age_days FLOAT;
    v_has_relationships INTEGER;
BEGIN
    SELECT
        access_count,
        EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - timestamp)) / 86400.0,
        (SELECT COUNT(*) FROM memory_relationships
         WHERE source_snapshot_id = p_snapshot_id OR target_snapshot_id = p_snapshot_id)
    INTO v_access_count, v_age_days, v_has_relationships
    FROM memory_snapshots
    WHERE id = p_snapshot_id;

    -- Importance formula:
    -- Base: 0.5
    -- +0.1 per access (max +0.3)
    -- +0.1 if accessed in last 7 days
    -- +0.1 if has relationships
    -- -0.01 per day old (decay over time)

    v_importance := 0.5
        + LEAST(v_access_count * 0.05, 0.3)
        + CASE WHEN v_age_days < 7 THEN 0.1 ELSE 0 END
        + CASE WHEN v_has_relationships > 0 THEN 0.1 ELSE 0 END
        - (v_age_days * 0.001);

    RETURN GREATEST(0.0, LEAST(1.0, v_importance));
END;
$$ LANGUAGE plpgsql;

-- Function: Update cache temperature based on access patterns
CREATE OR REPLACE FUNCTION update_memory_temperature(
    p_snapshot_id INTEGER
) RETURNS VARCHAR(10) AS $$
DECLARE
    v_access_count INTEGER;
    v_last_accessed TIMESTAMP;
    v_age_hours FLOAT;
    v_temperature VARCHAR(10);
BEGIN
    SELECT access_count, last_accessed
    INTO v_access_count, v_last_accessed
    FROM memory_snapshots
    WHERE id = p_snapshot_id;

    v_age_hours := EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - COALESCE(v_last_accessed, CURRENT_TIMESTAMP - INTERVAL '1 year'))) / 3600.0;

    -- Temperature logic:
    -- HOT: accessed 5+ times AND within last 24 hours
    -- WARM: accessed 2+ times AND within last 7 days
    -- COLD: everything else

    IF v_access_count >= 5 AND v_age_hours < 24 THEN
        v_temperature := 'hot';
    ELSIF v_access_count >= 2 AND v_age_hours < 168 THEN
        v_temperature := 'warm';
    ELSE
        v_temperature := 'cold';
    END IF;

    INSERT INTO memory_cache_stats (snapshot_id, temperature, updated_at)
    VALUES (p_snapshot_id, v_temperature, CURRENT_TIMESTAMP)
    ON CONFLICT (snapshot_id)
    DO UPDATE SET temperature = v_temperature, updated_at = CURRENT_TIMESTAMP;

    RETURN v_temperature;
END;
$$ LANGUAGE plpgsql;

-- Function: Semantic search for memories
CREATE OR REPLACE FUNCTION semantic_memory_search(
    p_query_embedding vector(384),
    p_limit INTEGER DEFAULT 10,
    p_similarity_threshold FLOAT DEFAULT 0.5
) RETURNS TABLE (
    snapshot_id INTEGER,
    similarity_score FLOAT,
    timestamp TIMESTAMP,
    event_type VARCHAR(50),
    importance FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ms.id,
        1 - (me.embedding <=> p_query_embedding) AS similarity,
        ms.timestamp,
        ms.event_type,
        ms.importance_score
    FROM memory_embeddings me
    JOIN memory_snapshots ms ON me.snapshot_id = ms.id
    WHERE
        ms.archived = FALSE
        AND (1 - (me.embedding <=> p_query_embedding)) >= p_similarity_threshold
    ORDER BY me.embedding <=> p_query_embedding
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- 9. TRIGGERS
-- ========================================

-- Trigger: Auto-update importance score when memory accessed
CREATE OR REPLACE FUNCTION update_memory_access() RETURNS TRIGGER AS $$
BEGIN
    UPDATE memory_snapshots
    SET
        access_count = access_count + 1,
        last_accessed = CURRENT_TIMESTAMP,
        importance_score = calculate_memory_importance(NEW.id)
    WHERE id = NEW.id;

    PERFORM update_memory_temperature(NEW.id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Queue new memories for embedding generation
CREATE OR REPLACE FUNCTION queue_embedding_generation() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO embedding_generation_queue (snapshot_id, status)
    VALUES (NEW.id, 'pending');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_queue_embedding
AFTER INSERT ON memory_snapshots
FOR EACH ROW
EXECUTE FUNCTION queue_embedding_generation();

-- ========================================
-- 10. INITIAL DATA MIGRATION
-- ========================================

-- Set default importance for existing memories
UPDATE memory_snapshots
SET importance_score = 0.5
WHERE importance_score IS NULL;

-- Queue existing memories for embedding generation
INSERT INTO embedding_generation_queue (snapshot_id, status)
SELECT id, 'pending'
FROM memory_snapshots
WHERE NOT EXISTS (
    SELECT 1 FROM embedding_generation_queue WHERE snapshot_id = memory_snapshots.id
);

-- ========================================
-- MIGRATION COMPLETE
-- ========================================
-- ARES now has semantic memory with intelligent retrieval
-- Next steps:
-- 1. Run embedding generation service
-- 2. Test semantic search
-- 3. Monitor cache hit rates
-- 4. Tune similarity thresholds
-- ========================================
