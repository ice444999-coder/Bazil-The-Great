-- Migration 019: Add Embedding Column (Placeholder for pgvector)
-- Date: 2025-10-18
-- Purpose: Add embedding support to solace_memory_crystals
-- Note: Using TEXT placeholder until pgvector binary available for PostgreSQL 18
-- Crystal Reference: #27 (Master Spec), #28 (Steps), #29 (Blocker), #30 (Alternative)

-- Add embedding column as TEXT (stores JSON array temporarily)
ALTER TABLE solace_memory_crystals 
ADD COLUMN IF NOT EXISTS embedding TEXT;

-- Add index for faster queries (GIN index for JSON-like text searching)
CREATE INDEX IF NOT EXISTS idx_crystals_embedding_text 
ON solace_memory_crystals USING gin (to_tsvector('english', COALESCE(embedding, '')));

-- Add metadata columns for embedding management
ALTER TABLE solace_memory_crystals 
ADD COLUMN IF NOT EXISTS embedding_model VARCHAR(100) DEFAULT 'text-embedding-3-small',
ADD COLUMN IF NOT EXISTS embedding_generated_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS embedding_version INT DEFAULT 1;

-- Add comment documenting placeholder status
COMMENT ON COLUMN solace_memory_crystals.embedding IS 
'Placeholder TEXT column storing JSON array of float32 values. Will be migrated to vector(1536) when pgvector extension becomes available for PostgreSQL 18.';

-- Verification query
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'solace_memory_crystals'
AND column_name IN ('embedding', 'embedding_model', 'embedding_generated_at', 'embedding_version')
ORDER BY ordinal_position;
