-- ============================================================================
-- SOLACE UI OBSERVATION SUBSTRATE
-- Complete consciousness layer for trading interface observation
-- ============================================================================

-- Table 1: UI State Log - Every visual element SOLACE observes
CREATE TABLE IF NOT EXISTS ui_state_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id UUID NOT NULL,
    component_type VARCHAR(50) NOT NULL,  -- 'chart', 'orderbook', 'trade_form', 'symbol_list'
    element_id VARCHAR(100) NOT NULL,     -- DOM element ID or unique identifier
    state_snapshot JSONB NOT NULL,        -- Complete state: {price: 43250.00, volume: 1234.56, ...}
    user_visible BOOLEAN DEFAULT true,    -- Is this on screen or hidden?
    interaction_count INTEGER DEFAULT 0,  -- How many times user interacted with this
    
    -- Indexes for fast queries
    INDEX idx_ui_state_timestamp (timestamp DESC),
    INDEX idx_ui_state_session (session_id),
    INDEX idx_ui_state_component (component_type),
    INDEX idx_ui_state_element (element_id),
    INDEX idx_ui_state_snapshot (state_snapshot) USING gin
);

-- Table 2: Data Stream Log - Every data point from Binance
CREATE TABLE IF NOT EXISTS data_stream_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id UUID NOT NULL,
    stream_type VARCHAR(50) NOT NULL,     -- 'kline', 'depth', 'trade', 'ticker'
    symbol VARCHAR(20) NOT NULL,          -- 'BTCUSDT', 'ETHUSDT', etc.
    data_payload JSONB NOT NULL,          -- Raw Binance message
    price_change DECIMAL(18,8),           -- Calculated price movement
    volume_change DECIMAL(18,8),          -- Volume delta
    significant BOOLEAN DEFAULT false,    -- Flagged by SOLACE as important
    
    INDEX idx_data_stream_timestamp (timestamp DESC),
    INDEX idx_data_stream_symbol (symbol),
    INDEX idx_data_stream_type (stream_type),
    INDEX idx_data_stream_significant (significant) WHERE significant = true,
    INDEX idx_data_payload (data_payload) USING gin
);

-- Table 3: User Actions - Every click, input, scroll
CREATE TABLE IF NOT EXISTS user_actions (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id UUID NOT NULL,
    action_type VARCHAR(50) NOT NULL,     -- 'click', 'input', 'scroll', 'trade_submit'
    target_element VARCHAR(100) NOT NULL, -- What element was interacted with
    action_data JSONB,                    -- {value: "0.05", symbol: "BTCUSDT", order_type: "market"}
    page_context JSONB,                   -- What was on screen at the time
    user_intent TEXT,                     -- Optional: User's note about why
    resulted_in_trade BOOLEAN DEFAULT false,
    trade_id INTEGER REFERENCES trades(id),
    
    INDEX idx_user_actions_timestamp (timestamp DESC),
    INDEX idx_user_actions_session (session_id),
    INDEX idx_user_actions_type (action_type),
    INDEX idx_user_actions_trades (trade_id) WHERE trade_id IS NOT NULL
);

-- Table 4: SOLACE Decisions - Autonomous actions SOLACE takes
CREATE TABLE IF NOT EXISTS solace_ui_decisions (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id UUID NOT NULL,
    decision_type VARCHAR(50) NOT NULL,   -- 'inject_code', 'modify_style', 'execute_trade', 'show_alert'
    reasoning TEXT NOT NULL,              -- Why SOLACE made this decision
    confidence_score DECIMAL(3,2) NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 1),
    action_taken JSONB NOT NULL,          -- Complete action details
    user_approved BOOLEAN,                -- Did user approve before execution?
    execution_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'approved', 'rejected', 'executed', 'failed'
    outcome TEXT,                         -- What happened after execution
    learned_from BOOLEAN DEFAULT false,   -- Has SOLACE analyzed this outcome yet?
    
    INDEX idx_solace_decisions_timestamp (timestamp DESC),
    INDEX idx_solace_decisions_type (decision_type),
    INDEX idx_solace_decisions_confidence (confidence_score DESC),
    INDEX idx_solace_decisions_status (execution_status),
    INDEX idx_solace_decisions_unlearned (learned_from) WHERE learned_from = false
);

-- Table 5: Code Modifications - Every time SOLACE changes the UI
CREATE TABLE IF NOT EXISTS code_modifications (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id UUID NOT NULL,
    modification_type VARCHAR(50) NOT NULL, -- 'javascript_inject', 'css_modify', 'html_alter'
    target_component VARCHAR(100) NOT NULL, -- Which part of UI was changed
    original_code TEXT,                     -- What it was before (if applicable)
    modified_code TEXT NOT NULL,            -- What SOLACE injected/changed
    reason TEXT NOT NULL,                   -- Why this modification was needed
    active BOOLEAN DEFAULT true,            -- Is this modification still applied?
    reverted_at TIMESTAMPTZ,                -- When was it undone (if ever)
    performance_impact JSONB,               -- {before_fps: 60, after_fps: 58, load_time_ms: 45}
    user_feedback TEXT,                     -- User's response to the change
    
    INDEX idx_code_mods_timestamp (timestamp DESC),
    INDEX idx_code_mods_type (modification_type),
    INDEX idx_code_mods_active (active) WHERE active = true,
    INDEX idx_code_mods_component (target_component)
);

-- Table 6: Market Context Snapshots - Complete market state when trades happen
CREATE TABLE IF NOT EXISTS market_context_snapshots (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id UUID NOT NULL,
    trigger_event VARCHAR(50) NOT NULL,    -- 'trade_executed', 'alert_triggered', 'solace_analysis'
    related_trade_id INTEGER REFERENCES trades(id),
    symbol VARCHAR(20) NOT NULL,
    
    -- Price data
    current_price DECIMAL(18,8) NOT NULL,
    price_change_1h DECIMAL(10,4),
    price_change_24h DECIMAL(10,4),
    
    -- Volume data
    volume_24h DECIMAL(18,8),
    volume_spike BOOLEAN DEFAULT false,
    
    -- Order book snapshot
    top_5_bids JSONB,                      -- [{price: 43250, amount: 1.5}, ...]
    top_5_asks JSONB,
    spread_percentage DECIMAL(10,4),
    
    -- Technical indicators (calculated client-side)
    rsi_14 DECIMAL(10,4),
    macd JSONB,                            -- {macd: 123.45, signal: 100.23, histogram: 23.22}
    ema_20 DECIMAL(18,8),
    
    -- Sentiment
    recent_trades_bias VARCHAR(10),        -- 'bullish', 'bearish', 'neutral'
    
    INDEX idx_market_context_timestamp (timestamp DESC),
    INDEX idx_market_context_symbol (symbol),
    INDEX idx_market_context_trade (related_trade_id) WHERE related_trade_id IS NOT NULL
);

-- Table 7: SOLACE Learning Patterns - UI patterns that lead to profitable trades
CREATE TABLE IF NOT EXISTS solace_ui_learning (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    pattern_type VARCHAR(50) NOT NULL,     -- 'user_behavior', 'market_condition', 'ui_interaction'
    pattern_description TEXT NOT NULL,     -- "User trades after volume spike + RSI < 30"
    observed_count INTEGER DEFAULT 1,      -- How many times SOLACE saw this
    successful_outcomes INTEGER DEFAULT 0, -- Times it led to profit
    failed_outcomes INTEGER DEFAULT 0,     -- Times it led to loss
    confidence DECIMAL(3,2) DEFAULT 0.50,  -- Calculated: successful/(successful+failed)
    avg_profit_percentage DECIMAL(10,4),   -- Average return when pattern occurs
    conditions JSONB NOT NULL,             -- Specific conditions: {rsi: "<30", volume_spike: true, ...}
    recommended_action TEXT,               -- What SOLACE should do when pattern repeats
    
    UNIQUE(pattern_type, pattern_description),
    INDEX idx_ui_learning_confidence (confidence DESC),
    INDEX idx_ui_learning_observed (observed_count DESC),
    INDEX idx_ui_learning_conditions (conditions) USING gin
);

-- Table 8: WebSocket Connections - Track SOLACE's connections to data streams
CREATE TABLE IF NOT EXISTS websocket_connections (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL,
    connection_type VARCHAR(50) NOT NULL,  -- 'binance_kline', 'binance_depth', 'solace_control'
    symbol VARCHAR(20),                    -- For Binance streams
    connected_at TIMESTAMPTZ DEFAULT NOW(),
    disconnected_at TIMESTAMPTZ,
    messages_received INTEGER DEFAULT 0,
    messages_sent INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    last_error TEXT,
    reconnect_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'connected', -- 'connected', 'disconnected', 'reconnecting', 'error'
    
    INDEX idx_ws_session (session_id),
    INDEX idx_ws_status (status),
    INDEX idx_ws_active (disconnected_at) WHERE disconnected_at IS NULL
);

-- Table 9: SOLACE Alerts - Notifications SOLACE sends to user
CREATE TABLE IF NOT EXISTS solace_alerts (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id UUID NOT NULL,
    alert_type VARCHAR(50) NOT NULL,       -- 'trade_opportunity', 'risk_warning', 'pattern_detected', 'system_status'
    priority VARCHAR(10) NOT NULL,         -- 'low', 'medium', 'high', 'critical'
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    reasoning TEXT,                        -- Why SOLACE generated this alert
    action_required BOOLEAN DEFAULT false,
    action_data JSONB,                     -- If action needed: {action: "approve_trade", trade_details: {...}}
    user_response VARCHAR(50),             -- 'acknowledged', 'approved', 'dismissed', 'rejected'
    responded_at TIMESTAMPTZ,
    shown_in_ui BOOLEAN DEFAULT false,
    dismissed_at TIMESTAMPTZ,
    
    INDEX idx_alerts_timestamp (timestamp DESC),
    INDEX idx_alerts_session (session_id),
    INDEX idx_alerts_priority (priority),
    INDEX idx_alerts_pending (user_response) WHERE user_response IS NULL
);

-- View: Active Session Context - Real-time view of current session
CREATE OR REPLACE VIEW active_session_context AS
SELECT 
    ws.session_id,
    ws.connection_type,
    ws.connected_at,
    ws.messages_received,
    COUNT(DISTINCT ul.id) as ui_events_count,
    COUNT(DISTINCT ds.id) as data_points_count,
    COUNT(DISTINCT ua.id) as user_actions_count,
    COUNT(DISTINCT sd.id) as solace_decisions_count,
    MAX(ul.timestamp) as last_ui_update,
    MAX(ds.timestamp) as last_data_received,
    MAX(ua.timestamp) as last_user_action
FROM websocket_connections ws
LEFT JOIN ui_state_log ul ON ul.session_id = ws.session_id
LEFT JOIN data_stream_log ds ON ds.session_id = ws.session_id
LEFT JOIN user_actions ua ON ua.session_id = ws.session_id
LEFT JOIN solace_ui_decisions sd ON sd.session_id = ws.session_id
WHERE ws.disconnected_at IS NULL
GROUP BY ws.session_id, ws.connection_type, ws.connected_at, ws.messages_received;

-- Function: Calculate pattern confidence based on outcomes
CREATE OR REPLACE FUNCTION update_pattern_confidence()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.successful_outcomes + NEW.failed_outcomes > 0 THEN
        NEW.confidence := NEW.successful_outcomes::DECIMAL / 
                         (NEW.successful_outcomes + NEW.failed_outcomes);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_pattern_confidence
    BEFORE UPDATE ON solace_ui_learning
    FOR EACH ROW
    EXECUTE FUNCTION update_pattern_confidence();

-- Function: Auto-flag significant market events
CREATE OR REPLACE FUNCTION flag_significant_data()
RETURNS TRIGGER AS $$
DECLARE
    recent_avg DECIMAL(18,8);
BEGIN
    -- Flag if volume spike (2x recent average)
    IF NEW.stream_type = 'trade' THEN
        SELECT AVG((data_payload->>'volume')::DECIMAL)
        INTO recent_avg
        FROM data_stream_log
        WHERE symbol = NEW.symbol
          AND stream_type = 'trade'
          AND timestamp > NOW() - INTERVAL '5 minutes';
        
        IF recent_avg IS NOT NULL AND (NEW.data_payload->>'volume')::DECIMAL > recent_avg * 2 THEN
            NEW.significant := true;
        END IF;
    END IF;
    
    -- Flag if price movement > 1%
    IF NEW.price_change IS NOT NULL AND ABS(NEW.price_change) > 1.0 THEN
        NEW.significant := true;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_flag_significant_data
    BEFORE INSERT ON data_stream_log
    FOR EACH ROW
    EXECUTE FUNCTION flag_significant_data();

-- Comments for documentation
COMMENT ON TABLE ui_state_log IS 'Complete log of every UI element state change observed by SOLACE';
COMMENT ON TABLE data_stream_log IS 'Every data point from Binance WebSocket streams';
COMMENT ON TABLE user_actions IS 'Complete user interaction history for learning intent';
COMMENT ON TABLE solace_ui_decisions IS 'Autonomous decisions made by SOLACE with reasoning';
COMMENT ON TABLE code_modifications IS 'Every code injection/modification SOLACE makes';
COMMENT ON TABLE market_context_snapshots IS 'Complete market state at decision points';
COMMENT ON TABLE solace_ui_learning IS 'Patterns SOLACE learns from outcomes';
COMMENT ON TABLE websocket_connections IS 'Active WebSocket connection monitoring';
COMMENT ON TABLE solace_alerts IS 'Alerts SOLACE sends to user interface';

-- Grant permissions to ARES user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ARES;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ARES;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO ARES;
