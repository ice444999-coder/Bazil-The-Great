-- ============================================
-- GLASS BOX DECISION TREE + HEDERA BLOCKCHAIN SYSTEM
-- Production-grade observability with immutable audit trails
-- Created: 2025-10-16
-- ============================================

-- ============================================
-- TABLE 1: Decision Traces (Root of each tree)
-- ============================================
CREATE TABLE IF NOT EXISTS decision_traces (
  id SERIAL PRIMARY KEY,
  trade_id INTEGER REFERENCES trades(id),
  trace_type VARCHAR(50) NOT NULL, -- 'trade_entry', 'trade_exit', 'risk_adjustment'
  status VARCHAR(20) NOT NULL, -- 'in_progress', 'completed', 'failed'
  start_time TIMESTAMP NOT NULL DEFAULT NOW(),
  end_time TIMESTAMP,
  total_duration_ms INTEGER,
  final_decision TEXT,
  confidence_score DECIMAL(5,2), -- 0-100
  
  -- Hedera anchoring
  is_anchored BOOLEAN DEFAULT FALSE,
  merkle_root VARCHAR(64),
  
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_decision_traces_trade_id ON decision_traces(trade_id);
CREATE INDEX IF NOT EXISTS idx_decision_traces_type ON decision_traces(trace_type);
CREATE INDEX IF NOT EXISTS idx_decision_traces_status ON decision_traces(status);
CREATE INDEX IF NOT EXISTS idx_decision_traces_anchored ON decision_traces(is_anchored);
CREATE INDEX IF NOT EXISTS idx_decision_traces_created ON decision_traces(created_at DESC);

COMMENT ON TABLE decision_traces IS 'Root of each decision tree - represents complete decision flow from start to finish';
COMMENT ON COLUMN decision_traces.merkle_root IS 'Merkle tree root hash of all spans for Hedera anchoring';

-- ============================================
-- TABLE 2: Decision Spans (Individual nodes)
-- ============================================
CREATE TABLE IF NOT EXISTS decision_spans (
  id SERIAL PRIMARY KEY,
  trace_id INTEGER NOT NULL REFERENCES decision_traces(id) ON DELETE CASCADE,
  parent_span_id INTEGER REFERENCES decision_spans(id) ON DELETE SET NULL,
  
  -- Span identification
  span_name VARCHAR(100) NOT NULL, -- 'market_check', 'calculate_rsi', 'execute_order'
  span_type VARCHAR(50) NOT NULL, -- 'data_fetch', 'calculation', 'api_call', 'decision'
  chain_position INTEGER NOT NULL, -- Position in the chain (0, 1, 2...)
  
  -- Timing
  start_time TIMESTAMP NOT NULL DEFAULT NOW(),
  end_time TIMESTAMP,
  duration_ms INTEGER,
  
  -- Input/Output
  input_data JSONB,
  output_data JSONB,
  
  -- Decision Logic
  decision_reasoning TEXT,
  confidence_score DECIMAL(5,2), -- 0-100
  
  -- Status
  status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'success', 'failed', 'skipped'
  error_message TEXT,
  
  -- Context Handoff (learning between nodes)
  context_from_previous JSONB,
  context_to_next JSONB,
  
  -- Glass Box Hashing (blockchain-style chaining)
  sha256_hash VARCHAR(64) NOT NULL,
  previous_hash VARCHAR(64), -- Hash of previous span (blockchain-style chaining)
  data_snapshot TEXT, -- Raw data used for hash calculation (for verification)
  
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_decision_spans_trace_id ON decision_spans(trace_id);
CREATE INDEX IF NOT EXISTS idx_decision_spans_parent ON decision_spans(parent_span_id);
CREATE INDEX IF NOT EXISTS idx_decision_spans_type ON decision_spans(span_type);
CREATE INDEX IF NOT EXISTS idx_decision_spans_status ON decision_spans(status);
CREATE INDEX IF NOT EXISTS idx_decision_spans_chain_position ON decision_spans(trace_id, chain_position);
CREATE INDEX IF NOT EXISTS idx_decision_spans_hash ON decision_spans(sha256_hash);
CREATE INDEX IF NOT EXISTS idx_decision_spans_created ON decision_spans(created_at DESC);

COMMENT ON TABLE decision_spans IS 'Individual nodes in decision tree - each represents one step in reasoning process';
COMMENT ON COLUMN decision_spans.sha256_hash IS 'SHA-256 hash of this span (chained with previous_hash for tamper detection)';
COMMENT ON COLUMN decision_spans.previous_hash IS 'Hash of previous span - creates blockchain-style immutable chain';
COMMENT ON COLUMN decision_spans.data_snapshot IS 'Canonical string used for hash calculation - enables verification';

-- ============================================
-- TABLE 3: Decision Metrics
-- ============================================
CREATE TABLE IF NOT EXISTS decision_metrics (
  id SERIAL PRIMARY KEY,
  trace_id INTEGER NOT NULL REFERENCES decision_traces(id) ON DELETE CASCADE,
  span_id INTEGER REFERENCES decision_spans(id) ON DELETE CASCADE, -- NULL = trace-level metric
  metric_name VARCHAR(100) NOT NULL,
  metric_value DECIMAL(15,4) NOT NULL,
  metric_unit VARCHAR(20), -- 'ms', 'percent', 'count', 'usd'
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_decision_metrics_trace ON decision_metrics(trace_id, metric_name);
CREATE INDEX IF NOT EXISTS idx_decision_metrics_span ON decision_metrics(span_id, metric_name);
CREATE INDEX IF NOT EXISTS idx_decision_metrics_name ON decision_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_decision_metrics_created ON decision_metrics(created_at DESC);

COMMENT ON TABLE decision_metrics IS 'Performance and business metrics for traces/spans (latency, confidence, slippage, etc)';

-- ============================================
-- TABLE 4: Hedera Anchors (Blockchain proof)
-- ============================================
CREATE TABLE IF NOT EXISTS hedera_anchors (
  id SERIAL PRIMARY KEY,
  trace_id INTEGER NOT NULL REFERENCES decision_traces(id) ON DELETE CASCADE,
  
  -- Merkle tree data
  merkle_root VARCHAR(64) NOT NULL,
  span_count INTEGER NOT NULL,
  leaf_hashes TEXT[], -- Array of all span hashes (PostgreSQL array)
  
  -- Hedera transaction details
  hedera_topic_id VARCHAR(100) NOT NULL,
  hedera_txn_id VARCHAR(100) NOT NULL,
  hedera_consensus_timestamp TIMESTAMP,
  hedera_sequence_number BIGINT,
  
  -- Verification
  verification_url TEXT,
  verification_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'confirmed', 'failed'
  verified_at TIMESTAMP,
  
  -- Metadata
  anchored_at TIMESTAMP NOT NULL DEFAULT NOW(),
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hedera_anchors_trace ON hedera_anchors(trace_id);
CREATE INDEX IF NOT EXISTS idx_hedera_anchors_merkle ON hedera_anchors(merkle_root);
CREATE INDEX IF NOT EXISTS idx_hedera_anchors_txn ON hedera_anchors(hedera_txn_id);
CREATE INDEX IF NOT EXISTS idx_hedera_anchors_status ON hedera_anchors(verification_status);
CREATE INDEX IF NOT EXISTS idx_hedera_anchors_created ON hedera_anchors(created_at DESC);

-- Ensure one anchor per trace
CREATE UNIQUE INDEX IF NOT EXISTS idx_hedera_anchors_trace_unique ON hedera_anchors(trace_id);

COMMENT ON TABLE hedera_anchors IS 'Hedera Hashgraph blockchain anchors - immutable proof of decision integrity';
COMMENT ON COLUMN hedera_anchors.merkle_root IS 'Root hash of merkle tree built from all span hashes';
COMMENT ON COLUMN hedera_anchors.verification_url IS 'Public URL to verify transaction on Hedera (e.g., HashScan)';

-- ============================================
-- TABLE 5: Hash Chain Verification Log
-- ============================================
CREATE TABLE IF NOT EXISTS hash_chain_verifications (
  id SERIAL PRIMARY KEY,
  trace_id INTEGER NOT NULL REFERENCES decision_traces(id) ON DELETE CASCADE,
  verification_type VARCHAR(50) NOT NULL, -- 'chain_integrity', 'hedera_match', 'full_audit'
  is_valid BOOLEAN NOT NULL,
  error_message TEXT,
  verified_at TIMESTAMP NOT NULL DEFAULT NOW(),
  verified_by VARCHAR(100), -- 'system', 'auditor', 'regulator', 'user'
  
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hash_verifications_trace ON hash_chain_verifications(trace_id, verification_type);
CREATE INDEX IF NOT EXISTS idx_hash_verifications_status ON hash_chain_verifications(is_valid, verification_type);
CREATE INDEX IF NOT EXISTS idx_hash_verifications_created ON hash_chain_verifications(created_at DESC);

COMMENT ON TABLE hash_chain_verifications IS 'Audit log of all verification attempts (internal and external)';

-- ============================================
-- SAMPLE DATA (for testing)
-- ============================================
INSERT INTO decision_traces (id, trace_type, status, start_time, final_decision, confidence_score, is_anchored, merkle_root)
VALUES (
  1, 
  'trade_entry', 
  'completed', 
  NOW() - INTERVAL '5 minutes',
  'BUY BTC/USDT $1000 at $68234.50 (Stop: $61411.05, Target: $81881.40)',
  82.6,
  true,
  'a7f3e2d4b9c8f1e0a3d5c7b2e4f8a1c9d6e3f7b4a8c2d1e5f9b3a6c4e8d2f7a1'
) ON CONFLICT DO NOTHING;

-- ============================================
-- SUCCESS MESSAGE
-- ============================================
SELECT 
  'Glass Box Decision Tree Schema Created' AS status,
  (SELECT COUNT(*) FROM decision_traces) AS sample_traces,
  (SELECT COUNT(*) FROM decision_spans) AS sample_spans,
  (SELECT COUNT(*) FROM hedera_anchors) AS sample_anchors;
