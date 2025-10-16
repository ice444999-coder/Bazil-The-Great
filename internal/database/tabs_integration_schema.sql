-- TABS 3-7 Integration Schema
-- Created: 2025-10-15
-- Purpose: Support analytics, decisions, and chat functionality

-- ============================================
-- TAB 5: Live Decisions Table
-- ============================================
CREATE TABLE IF NOT EXISTS agent_decisions (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL,
    decision_type VARCHAR(50) NOT NULL, -- 'BUY', 'SELL', 'HOLD', 'ANALYZE'
    symbol VARCHAR(20) NOT NULL,
    reasoning TEXT NOT NULL, -- SOLACE's explanation
    confidence_score DECIMAL(5,2) NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 100),
    
    -- Market context at decision time
    price DECIMAL(20,8),
    indicators JSONB, -- RSI, MACD, SMA, etc
    
    -- Decision outcome
    action_taken VARCHAR(50), -- 'EXECUTED', 'SKIPPED', 'MANUAL_OVERRIDE'
    trade_id BIGINT REFERENCES sandbox_trades(id) ON DELETE SET NULL,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_decisions_session ON agent_decisions(session_id);
CREATE INDEX IF NOT EXISTS idx_agent_decisions_symbol ON agent_decisions(symbol);
CREATE INDEX IF NOT EXISTS idx_agent_decisions_created ON agent_decisions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_decisions_type ON agent_decisions(decision_type);

-- ============================================
-- TAB 7: Chat History Table
-- ============================================
CREATE TABLE IF NOT EXISTS chat_history (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL,
    sender VARCHAR(20) NOT NULL CHECK (sender IN ('user', 'solace')),
    message TEXT NOT NULL,
    
    -- Context at time of message
    context JSONB, -- Current trades, observations, market state
    
    -- SOLACE's thought process (only for SOLACE messages)
    internal_reasoning TEXT,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_history_session ON chat_history(session_id);
CREATE INDEX IF NOT EXISTS idx_chat_history_created ON chat_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_history_sender ON chat_history(sender);

-- ============================================
-- Sample Data for Testing
-- ============================================

-- Insert sample decision
INSERT INTO agent_decisions (
    session_id, 
    decision_type, 
    symbol, 
    reasoning, 
    confidence_score,
    price,
    indicators,
    action_taken
) VALUES (
    '00000000-0000-0000-0000-000000000004',
    'BUY',
    'BTC/USDT',
    'RSI at 32 (oversold), MACD showing bullish crossover, price bounced off 50-day SMA support. High probability setup.',
    87.5,
    42350.50,
    '{"rsi": 32, "macd": {"value": 0.0012, "signal": -0.0008}, "sma_50": 42200}',
    'EXECUTED'
) ON CONFLICT DO NOTHING;

-- Insert sample chat
INSERT INTO chat_history (
    session_id,
    sender,
    message,
    context
) VALUES (
    '00000000-0000-0000-0000-000000000004',
    'solace',
    'I analyzed your trading patterns over the last 7 days. Your win rate is 68% on BTC/USDT trades during morning hours (8-11 AM).',
    '{"total_trades": 23, "profitable": 15, "avg_win": 2.3, "avg_loss": -1.1}'
) ON CONFLICT DO NOTHING;

COMMENT ON TABLE agent_decisions IS 'Logs every decision SOLACE makes with reasoning and confidence';
COMMENT ON TABLE chat_history IS 'Stores all conversations between user and SOLACE for learning and compliance';

SELECT 'agent_decisions table created' AS status, COUNT(*) AS sample_rows FROM agent_decisions;
SELECT 'chat_history table created' AS status, COUNT(*) AS sample_rows FROM chat_history;
