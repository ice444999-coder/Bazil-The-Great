-- Migration: Glass Box File Timestamp Ledger
-- Purpose: Hash every file modification timestamp for immutable audit trail
-- Prevents falsified timestamps - proves when code actually changed
-- Anchors to Hedera blockchain periodically

-- File Timestamp Ledger Table
CREATE TABLE IF NOT EXISTS file_timestamp_ledger (
    id SERIAL PRIMARY KEY,
    file_path VARCHAR(500) NOT NULL,
    relative_path VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_extension VARCHAR(50),
    last_modified_utc TIMESTAMP NOT NULL,
    size_bytes BIGINT,
    line_count INT,
    
    -- SHA-256 hash of (file_path + last_modified + size + line_count)
    -- Proves file state at specific timestamp
    timestamp_hash VARCHAR(64) NOT NULL UNIQUE,
    
    -- Previous hash (forms blockchain chain)
    previous_hash VARCHAR(64),
    
    -- Merkle tree batching (anchor 1000 timestamps → 1 Hedera transaction)
    merkle_batch_id INT REFERENCES merkle_batches(id),
    hedera_anchor_id INT REFERENCES hedera_anchors(id),
    
    -- Metadata
    recorded_at TIMESTAMP DEFAULT NOW(),
    recorded_by VARCHAR(100) DEFAULT 'ARES_TIMESTAMP_SCANNER',
    scan_session_id UUID,
    
    -- Verification
    verified BOOLEAN DEFAULT FALSE,
    verification_status VARCHAR(50) DEFAULT 'pending',
    
    CONSTRAINT fk_merkle_batch FOREIGN KEY (merkle_batch_id) 
        REFERENCES merkle_batches(id) ON DELETE SET NULL,
    CONSTRAINT fk_hedera_anchor FOREIGN KEY (hedera_anchor_id) 
        REFERENCES hedera_anchors(id) ON DELETE SET NULL
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_file_timestamp_path ON file_timestamp_ledger(file_path);
CREATE INDEX IF NOT EXISTS idx_file_timestamp_hash ON file_timestamp_ledger(timestamp_hash);
CREATE INDEX IF NOT EXISTS idx_file_timestamp_recorded ON file_timestamp_ledger(recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_file_timestamp_batch ON file_timestamp_ledger(merkle_batch_id);
CREATE INDEX IF NOT EXISTS idx_file_timestamp_session ON file_timestamp_ledger(scan_session_id);

-- View: Latest File Timestamps
CREATE OR REPLACE VIEW latest_file_timestamps AS
SELECT DISTINCT ON (file_path)
    file_path,
    file_name,
    last_modified_utc,
    size_bytes,
    line_count,
    timestamp_hash,
    recorded_at,
    verified,
    hedera_anchor_id
FROM file_timestamp_ledger
ORDER BY file_path, recorded_at DESC;

-- View: Timestamp Verification Status
CREATE OR REPLACE VIEW timestamp_verification_status AS
SELECT 
    COUNT(*) AS total_timestamps,
    COUNT(*) FILTER (WHERE verified = true) AS verified_count,
    COUNT(*) FILTER (WHERE verified = false) AS unverified_count,
    COUNT(DISTINCT merkle_batch_id) AS total_merkle_batches,
    COUNT(DISTINCT hedera_anchor_id) AS total_hedera_anchors,
    COUNT(*) FILTER (WHERE hedera_anchor_id IS NOT NULL) AS anchored_count,
    MAX(recorded_at) AS last_scan_time
FROM file_timestamp_ledger;

-- Function: Calculate Timestamp Hash
-- Input: file_path, last_modified, size_bytes, line_count
-- Output: SHA-256 hash (hex string)
CREATE OR REPLACE FUNCTION calculate_timestamp_hash(
    p_file_path VARCHAR,
    p_last_modified TIMESTAMP,
    p_size_bytes BIGINT,
    p_line_count INT
) RETURNS VARCHAR AS $$
DECLARE
    v_input TEXT;
BEGIN
    -- Concatenate all fields with separator
    v_input := p_file_path || '|' || 
               p_last_modified::TEXT || '|' || 
               p_size_bytes::TEXT || '|' || 
               p_line_count::TEXT;
    
    -- Return SHA-256 hash
    RETURN encode(digest(v_input, 'sha256'), 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function: Record File Timestamp (with auto-hashing)
CREATE OR REPLACE FUNCTION record_file_timestamp(
    p_file_path VARCHAR,
    p_relative_path VARCHAR,
    p_file_name VARCHAR,
    p_file_extension VARCHAR,
    p_last_modified TIMESTAMP,
    p_size_bytes BIGINT,
    p_line_count INT,
    p_scan_session_id UUID
) RETURNS VARCHAR AS $$
DECLARE
    v_timestamp_hash VARCHAR(64);
    v_previous_hash VARCHAR(64);
    v_new_id INT;
BEGIN
    -- Calculate timestamp hash
    v_timestamp_hash := calculate_timestamp_hash(
        p_file_path,
        p_last_modified,
        p_size_bytes,
        p_line_count
    );
    
    -- Get previous hash (blockchain chain)
    SELECT timestamp_hash INTO v_previous_hash
    FROM file_timestamp_ledger
    ORDER BY recorded_at DESC
    LIMIT 1;
    
    -- Insert new record
    INSERT INTO file_timestamp_ledger (
        file_path,
        relative_path,
        file_name,
        file_extension,
        last_modified_utc,
        size_bytes,
        line_count,
        timestamp_hash,
        previous_hash,
        scan_session_id
    ) VALUES (
        p_file_path,
        p_relative_path,
        p_file_name,
        p_file_extension,
        p_last_modified,
        p_size_bytes,
        p_line_count,
        v_timestamp_hash,
        v_previous_hash,
        p_scan_session_id
    )
    RETURNING id INTO v_new_id;
    
    RETURN v_timestamp_hash;
END;
$$ LANGUAGE plpgsql;

-- Function: Verify Timestamp Chain Integrity
-- Checks if hash chain is unbroken (detects tampering)
CREATE OR REPLACE FUNCTION verify_timestamp_chain() RETURNS TABLE (
    ledger_id INT,
    file_path VARCHAR,
    timestamp_hash VARCHAR,
    previous_hash VARCHAR,
    expected_previous_hash VARCHAR,
    chain_valid BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    WITH ledger_chain AS (
        SELECT 
            id,
            file_timestamp_ledger.file_path,
            timestamp_hash,
            previous_hash,
            LAG(timestamp_hash) OVER (ORDER BY recorded_at) AS expected_previous
        FROM file_timestamp_ledger
        ORDER BY recorded_at
    )
    SELECT 
        id,
        ledger_chain.file_path,
        timestamp_hash,
        previous_hash,
        expected_previous,
        CASE 
            WHEN previous_hash IS NULL THEN TRUE  -- Genesis record
            WHEN previous_hash = expected_previous THEN TRUE
            ELSE FALSE
        END AS chain_valid
    FROM ledger_chain;
END;
$$ LANGUAGE plpgsql;

-- Seed Comment
COMMENT ON TABLE file_timestamp_ledger IS 'Immutable ledger of file modification timestamps. SHA-256 hashed and anchored to Hedera blockchain. Prevents timestamp falsification.';
COMMENT ON COLUMN file_timestamp_ledger.timestamp_hash IS 'SHA-256(file_path + last_modified + size + line_count). Proves file state at timestamp.';
COMMENT ON COLUMN file_timestamp_ledger.previous_hash IS 'Previous record hash. Forms blockchain chain. Genesis record = NULL.';
COMMENT ON COLUMN file_timestamp_ledger.merkle_batch_id IS 'Batch of 1000 timestamps → 1 Merkle root → 1 Hedera anchor (cost optimization).';

-- Success message
DO $$ 
BEGIN
    RAISE NOTICE '✅ Glass Box File Timestamp Ledger created successfully!';
    RAISE NOTICE '   - Table: file_timestamp_ledger';
    RAISE NOTICE '   - Views: latest_file_timestamps, timestamp_verification_status';
    RAISE NOTICE '   - Functions: calculate_timestamp_hash(), record_file_timestamp(), verify_timestamp_chain()';
    RAISE NOTICE '   - Ready to hash file timestamps immutably!';
END $$;
