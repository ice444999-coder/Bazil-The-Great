-- ============================================
-- GRPO LEARNING SYSTEM
-- Gradient-based Reward Policy Optimization
-- Enables SOLACE to learn from trading outcomes
-- Created: 2025-10-16
-- ============================================

-- ============================================
-- TABLE 1: Token Biases (Learning State)
-- ============================================
CREATE TABLE IF NOT EXISTS grpo_biases (
  id SERIAL PRIMARY KEY,
  token_text VARCHAR(100) NOT NULL,
  token_id INTEGER,
  bias_value DECIMAL(10,6) NOT NULL DEFAULT 0.0,
  update_count INTEGER NOT NULL DEFAULT 0,
  last_reward DECIMAL(10,6),
  cumulative_reward DECIMAL(15,6) DEFAULT 0.0,
  last_updated TIMESTAMP DEFAULT NOW(),
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_grpo_biases_token ON grpo_biases(token_text);
CREATE INDEX IF NOT EXISTS idx_grpo_biases_value ON grpo_biases(bias_value DESC);
CREATE INDEX IF NOT EXISTS idx_grpo_biases_updated ON grpo_biases(last_updated DESC);

COMMENT ON TABLE grpo_biases IS 'Token-level biases learned from trading outcomes - SOLACE''s evolving preferences';
COMMENT ON COLUMN grpo_biases.bias_value IS 'Adjustment to token probability (-1.0 to +1.0, higher = more preferred)';
COMMENT ON COLUMN grpo_biases.cumulative_reward IS 'Total reward accumulated for this token across all decisions';

-- ============================================
-- TABLE 2: Decision Rewards (Training Data)
-- ============================================
CREATE TABLE IF NOT EXISTS grpo_rewards (
  id SERIAL PRIMARY KEY,
  trace_id INTEGER REFERENCES decision_traces(id) ON DELETE CASCADE,
  trade_id INTEGER REFERENCES sandbox_trades(id) ON DELETE SET NULL,
  
  -- Outcome metrics
  reward_value DECIMAL(10,6) NOT NULL,
  outcome_quality DECIMAL(5,2), -- 0-100 quality score
  profit_loss DECIMAL(20,8),
  win_rate_contribution DECIMAL(5,2),
  
  -- Decision context
  decision_tokens TEXT[], -- Tokens used in decision
  confidence_score DECIMAL(5,2),
  execution_time_ms INTEGER,
  
  -- Reward calculation
  reward_type VARCHAR(50), -- 'profit', 'risk_management', 'execution_speed', 'composite'
  reward_weight DECIMAL(5,4) DEFAULT 1.0,
  
  -- Learning metadata
  learning_iteration INTEGER DEFAULT 0,
  applied_to_biases BOOLEAN DEFAULT FALSE,
  applied_at TIMESTAMP,
  
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_grpo_rewards_trace ON grpo_rewards(trace_id);
CREATE INDEX IF NOT EXISTS idx_grpo_rewards_trade ON grpo_rewards(trade_id);
CREATE INDEX IF NOT EXISTS idx_grpo_rewards_value ON grpo_rewards(reward_value DESC);
CREATE INDEX IF NOT EXISTS idx_grpo_rewards_applied ON grpo_rewards(applied_to_biases, applied_at);
CREATE INDEX IF NOT EXISTS idx_grpo_rewards_created ON grpo_rewards(created_at DESC);

COMMENT ON TABLE grpo_rewards IS 'Rewards from trading decisions - training signal for GRPO learning';
COMMENT ON COLUMN grpo_rewards.reward_value IS 'Normalized reward (-1.0 to +1.0, positive = good outcome)';
COMMENT ON COLUMN grpo_rewards.decision_tokens IS 'Tokens present in the decision that led to this outcome';

-- ============================================
-- TABLE 3: Learning Checkpoints (Versioning)
-- ============================================
CREATE TABLE IF NOT EXISTS grpo_checkpoints (
  id SERIAL PRIMARY KEY,
  checkpoint_name VARCHAR(200) NOT NULL,
  checkpoint_type VARCHAR(50) DEFAULT 'auto', -- 'auto', 'manual', 'milestone'
  
  -- Performance snapshot
  total_trades INTEGER,
  win_rate DECIMAL(5,2),
  average_reward DECIMAL(10,6),
  bias_drift_magnitude DECIMAL(10,6),
  
  -- Bias snapshot (JSONB for flexibility)
  bias_snapshot JSONB,
  top_tokens JSONB, -- Top 100 most influential tokens
  
  -- Metadata
  created_by VARCHAR(100) DEFAULT 'system',
  notes TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_grpo_checkpoints_type ON grpo_checkpoints(checkpoint_type);
CREATE INDEX IF NOT EXISTS idx_grpo_checkpoints_created ON grpo_checkpoints(created_at DESC);

COMMENT ON TABLE grpo_checkpoints IS 'Snapshots of learning state for rollback and analysis';

-- ============================================
-- TABLE 4: Learning Metrics (Performance Tracking)
-- ============================================
CREATE TABLE IF NOT EXISTS grpo_metrics (
  id SERIAL PRIMARY KEY,
  metric_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
  
  -- Learning progress
  total_rewards_processed INTEGER DEFAULT 0,
  total_biases_updated INTEGER DEFAULT 0,
  learning_rate DECIMAL(10,6),
  
  -- Performance indicators
  average_reward_last_100 DECIMAL(10,6),
  bias_stability DECIMAL(5,4), -- 0-1, higher = more stable
  convergence_score DECIMAL(5,4), -- 0-1, higher = converging
  
  -- Top performers
  best_tokens JSONB, -- Top 10 tokens by reward
  worst_tokens JSONB, -- Bottom 10 tokens by reward
  
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_grpo_metrics_timestamp ON grpo_metrics(metric_timestamp DESC);

COMMENT ON TABLE grpo_metrics IS 'Time-series metrics tracking GRPO learning progress';

-- ============================================
-- HELPER FUNCTIONS
-- ============================================

-- Function to calculate reward from trade outcome
CREATE OR REPLACE FUNCTION calculate_trade_reward(
  p_profit_loss DECIMAL,
  p_trade_size DECIMAL,
  p_confidence DECIMAL
) RETURNS DECIMAL AS $$
DECLARE
  v_reward DECIMAL;
  v_roi DECIMAL;
BEGIN
  -- Calculate ROI
  v_roi := p_profit_loss / NULLIF(p_trade_size, 0);
  
  -- Normalize reward (-1 to +1)
  -- Positive ROI → positive reward
  -- Negative ROI → negative reward
  -- Scale by confidence (higher confidence = higher magnitude)
  v_reward := SIGN(v_roi) * LEAST(ABS(v_roi) * 10, 1.0) * (p_confidence / 100.0);
  
  RETURN v_reward;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION calculate_trade_reward IS 'Converts trade outcome to normalized reward signal for GRPO';

-- Function to get top biased tokens
CREATE OR REPLACE FUNCTION get_top_biased_tokens(p_limit INTEGER DEFAULT 20)
RETURNS TABLE (
  token_text VARCHAR,
  bias_value DECIMAL,
  cumulative_reward DECIMAL,
  update_count INTEGER
) AS $$
BEGIN
  RETURN QUERY
  SELECT 
    gb.token_text,
    gb.bias_value,
    gb.cumulative_reward,
    gb.update_count
  FROM grpo_biases gb
  ORDER BY ABS(gb.bias_value) DESC
  LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_top_biased_tokens IS 'Returns tokens with strongest positive or negative biases';

-- ============================================
-- INITIAL DATA
-- ============================================

-- Create initial checkpoint (baseline)
INSERT INTO grpo_checkpoints (
  checkpoint_name,
  checkpoint_type,
  total_trades,
  win_rate,
  average_reward,
  bias_drift_magnitude,
  bias_snapshot,
  created_by,
  notes
) VALUES (
  'GRPO System Initialization',
  'milestone',
  0,
  0.0,
  0.0,
  0.0,
  '{}',
  'system',
  'Initial checkpoint before any learning has occurred'
);

-- ============================================
-- VERIFICATION QUERIES
-- ============================================

-- Check tables exist
DO $$
BEGIN
  RAISE NOTICE 'Verifying GRPO tables...';
  
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'grpo_biases') THEN
    RAISE NOTICE '✅ grpo_biases table created';
  END IF;
  
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'grpo_rewards') THEN
    RAISE NOTICE '✅ grpo_rewards table created';
  END IF;
  
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'grpo_checkpoints') THEN
    RAISE NOTICE '✅ grpo_checkpoints table created';
  END IF;
  
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'grpo_metrics') THEN
    RAISE NOTICE '✅ grpo_metrics table created';
  END IF;
  
  RAISE NOTICE '🧠 GRPO Learning System ready for SOLACE evolution';
END $$;
