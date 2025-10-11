-- ARES Autonomous Trading System Schema
-- Phase 1: Sandbox Trading with Full Audit Trail
-- Author: Claude (VS Code Engineer) + SOLACE
-- Date: 2025-10-11

-- ============================================
-- SANDBOX TRADING TABLES
-- ============================================

-- Sandbox trades with full reasoning and audit trail
CREATE TABLE IF NOT EXISTS sandbox_trades (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    session_id UUID NOT NULL,

    -- Trade Details
    trading_pair VARCHAR(50) NOT NULL,  -- e.g., BTC/USDC, ETH/USDC
    direction VARCHAR(4) NOT NULL CHECK (direction IN ('BUY', 'SELL')),
    size DECIMAL(18,8) NOT NULL CHECK (size > 0),
    entry_price DECIMAL(18,8) NOT NULL CHECK (entry_price > 0),
    exit_price DECIMAL(18,8),

    -- Financial Results
    profit_loss DECIMAL(18,8),
    profit_loss_percent DECIMAL(10,4),
    fees DECIMAL(18,8) DEFAULT 0,

    -- Trade Status
    status VARCHAR(20) NOT NULL CHECK (status IN ('OPEN', 'CLOSED', 'CANCELLED')),
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,

    -- SOLACE Learning Data
    reasoning TEXT NOT NULL,  -- Why SOLACE made this trade
    market_conditions JSONB,  -- Price, volume, volatility at trade time
    sentiment_score DECIMAL(5,4),  -- -1.0 to 1.0
    confidence_score DECIMAL(5,4),  -- 0.0 to 1.0

    -- Benchmark & Performance
    benchmark_score DECIMAL(10,4),  -- Score at time of trade
    sharpe_ratio DECIMAL(10,4),  -- Updated when closed
    sortino_ratio DECIMAL(10,4),  -- Updated when closed

    -- Market Regime
    market_regime VARCHAR(20) CHECK (market_regime IN ('BULL', 'BEAR', 'CHOP', 'VOLATILITY_SPIKE', 'UNKNOWN')),
    regime_confidence DECIMAL(5,4),

    -- Audit Trail
    trade_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA256 hash of trade data
    lineage_trail JSONB,  -- Parent trades, mutations, strategy version
    solace_override BOOLEAN DEFAULT FALSE,  -- Did SOLACE override user?
    override_reason TEXT,  -- Why SOLACE overrode

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Indexes
    CONSTRAINT valid_profit_loss CHECK (
        (status = 'CLOSED' AND profit_loss IS NOT NULL AND exit_price IS NOT NULL) OR
        (status = 'OPEN' AND profit_loss IS NULL)
    )
);

CREATE INDEX idx_sandbox_trades_user ON sandbox_trades(user_id);
CREATE INDEX idx_sandbox_trades_session ON sandbox_trades(session_id);
CREATE INDEX idx_sandbox_trades_status ON sandbox_trades(status);
CREATE INDEX idx_sandbox_trades_pair ON sandbox_trades(trading_pair);
CREATE INDEX idx_sandbox_trades_opened ON sandbox_trades(opened_at DESC);
CREATE INDEX idx_sandbox_trades_regime ON sandbox_trades(market_regime);
CREATE INDEX idx_sandbox_trades_hash ON sandbox_trades(trade_hash);

-- ============================================
-- TRADING PERFORMANCE METRICS
-- ============================================

CREATE TABLE IF NOT EXISTS trading_performance (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    session_id UUID NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Trade Statistics
    total_trades INTEGER NOT NULL DEFAULT 0,
    winning_trades INTEGER NOT NULL DEFAULT 0,
    losing_trades INTEGER NOT NULL DEFAULT 0,
    win_rate DECIMAL(5,2),  -- Percentage

    -- Financial Metrics
    total_profit_loss DECIMAL(18,8),
    avg_profit DECIMAL(18,8),
    avg_loss DECIMAL(18,8),
    largest_win DECIMAL(18,8),
    largest_loss DECIMAL(18,8),

    -- Risk Metrics
    sharpe_ratio DECIMAL(10,4),
    sortino_ratio DECIMAL(10,4),
    max_drawdown DECIMAL(10,4),
    max_drawdown_percent DECIMAL(5,2),
    current_drawdown DECIMAL(10,4),

    -- Position Sizing
    avg_position_size DECIMAL(18,8),
    max_position_size DECIMAL(18,8),
    kelly_criterion DECIMAL(5,4),

    -- Risk Management
    var_5_percent DECIMAL(18,8),  -- Value at Risk
    risk_of_ruin DECIMAL(5,4),  -- Probability of ruin

    -- Strategy Evolution
    strategy_version INTEGER NOT NULL DEFAULT 1,
    mutation_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_trading_performance_user ON trading_performance(user_id);
CREATE INDEX idx_trading_performance_session ON trading_performance(session_id);
CREATE INDEX idx_trading_performance_calculated ON trading_performance(calculated_at DESC);

-- ============================================
-- MARKET DATA CACHE
-- ============================================

CREATE TABLE IF NOT EXISTS market_data_cache (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,  -- BTC, ETH, SOL, etc.
    trading_pair VARCHAR(50) NOT NULL,  -- BTC/USDC

    -- OHLCV Data
    timestamp TIMESTAMPTZ NOT NULL,
    open DECIMAL(18,8) NOT NULL,
    high DECIMAL(18,8) NOT NULL,
    low DECIMAL(18,8) NOT NULL,
    close DECIMAL(18,8) NOT NULL,
    volume DECIMAL(18,8) NOT NULL,

    -- Technical Indicators
    sma_20 DECIMAL(18,8),
    sma_50 DECIMAL(18,8),
    sma_200 DECIMAL(18,8),
    rsi_14 DECIMAL(5,2),
    atr_14 DECIMAL(18,8),
    bollinger_upper DECIMAL(18,8),
    bollinger_lower DECIMAL(18,8),

    -- Market Regime
    volatility DECIMAL(10,6),
    trend_strength DECIMAL(5,4),
    market_regime VARCHAR(20),

    -- Data Source
    source VARCHAR(50) NOT NULL,  -- coingecko, jupiter, etc.

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(symbol, timestamp, source)
);

CREATE INDEX idx_market_data_symbol ON market_data_cache(symbol);
CREATE INDEX idx_market_data_pair ON market_data_cache(trading_pair);
CREATE INDEX idx_market_data_timestamp ON market_data_cache(timestamp DESC);
CREATE INDEX idx_market_data_regime ON market_data_cache(market_regime);

-- ============================================
-- STRATEGY MUTATIONS (Recursive Learning)
-- ============================================

CREATE TABLE IF NOT EXISTS strategy_mutations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    session_id UUID NOT NULL,

    -- Strategy Identity
    strategy_version INTEGER NOT NULL,
    strategy_name VARCHAR(100) NOT NULL,
    strategy_hash VARCHAR(64) NOT NULL UNIQUE,

    -- Mutation Details
    parent_hash VARCHAR(64),  -- Hash of previous strategy
    mutation_type VARCHAR(50) NOT NULL,  -- PARAMETER_TUNE, RULE_ADD, RULE_REMOVE, etc.
    mutation_delta JSONB NOT NULL,  -- What changed
    mutation_reason TEXT NOT NULL,  -- Why SOLACE mutated

    -- Performance Before/After
    sharpe_before DECIMAL(10,4),
    sharpe_after DECIMAL(10,4),
    sortino_before DECIMAL(10,4),
    sortino_after DECIMAL(10,4),
    win_rate_before DECIMAL(5,2),
    win_rate_after DECIMAL(5,2),

    -- Approval Status
    status VARCHAR(20) NOT NULL CHECK (status IN ('TESTING', 'APPROVED', 'REJECTED', 'DEPLOYED')),
    approved_by VARCHAR(50),  -- 'SOLACE', 'USER', 'BENCHMARK'

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deployed_at TIMESTAMPTZ,

    FOREIGN KEY (parent_hash) REFERENCES strategy_mutations(strategy_hash)
);

CREATE INDEX idx_strategy_mutations_user ON strategy_mutations(user_id);
CREATE INDEX idx_strategy_mutations_session ON strategy_mutations(session_id);
CREATE INDEX idx_strategy_mutations_version ON strategy_mutations(strategy_version DESC);
CREATE INDEX idx_strategy_mutations_status ON strategy_mutations(status);
CREATE INDEX idx_strategy_mutations_hash ON strategy_mutations(strategy_hash);

-- ============================================
-- RISK EVENTS (Kill-Switch Logs)
-- ============================================

CREATE TABLE IF NOT EXISTS risk_events (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL,  -- DRAWDOWN_LIMIT, VAR_BREACH, KILL_SWITCH, etc.
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL')),

    -- Event Details
    description TEXT NOT NULL,
    trigger_value DECIMAL(18,8),
    threshold_value DECIMAL(18,8),

    -- Actions Taken
    action_taken VARCHAR(100) NOT NULL,  -- CLOSE_ALL_POSITIONS, HALT_TRADING, ALERT_USER, etc.
    positions_closed INTEGER DEFAULT 0,

    -- Response Time
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    response_latency_ms INTEGER,  -- Must be <250ms for kill-switch

    -- Metadata
    solace_decision BOOLEAN DEFAULT TRUE,
    override_allowed BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_risk_events_user ON risk_events(user_id);
CREATE INDEX idx_risk_events_type ON risk_events(event_type);
CREATE INDEX idx_risk_events_severity ON risk_events(severity);
CREATE INDEX idx_risk_events_detected ON risk_events(detected_at DESC);

-- ============================================
-- TRADING BALANCES (Enhanced)
-- ============================================

-- Add columns to existing balances table
ALTER TABLE balances ADD COLUMN IF NOT EXISTS auto_topup BOOLEAN DEFAULT FALSE;
ALTER TABLE balances ADD COLUMN IF NOT EXISTS topup_threshold DECIMAL(18,8) DEFAULT 1000.00;
ALTER TABLE balances ADD COLUMN IF NOT EXISTS topup_amount DECIMAL(18,8) DEFAULT 10000.00;
ALTER TABLE balances ADD COLUMN IF NOT EXISTS total_deposits DECIMAL(18,8) DEFAULT 10000.00;
ALTER TABLE balances ADD COLUMN IF NOT EXISTS total_withdrawals DECIMAL(18,8) DEFAULT 0.00;
ALTER TABLE balances ADD COLUMN IF NOT EXISTS realized_pnl DECIMAL(18,8) DEFAULT 0.00;
ALTER TABLE balances ADD COLUMN IF NOT EXISTS unrealized_pnl DECIMAL(18,8) DEFAULT 0.00;

-- ============================================
-- FUNCTIONS & TRIGGERS
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for sandbox_trades
DROP TRIGGER IF EXISTS update_sandbox_trades_updated_at ON sandbox_trades;
CREATE TRIGGER update_sandbox_trades_updated_at
    BEFORE UPDATE ON sandbox_trades
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- INITIAL DATA
-- ============================================

-- Set default starting balance for trading
UPDATE balances
SET usd_balance = 10000.00,
    total_deposits = 10000.00,
    auto_topup = FALSE,
    topup_threshold = 1000.00,
    topup_amount = 10000.00
WHERE usd_balance = 0;

-- ============================================
-- COMPLETION
-- ============================================

-- Log migration completion
DO $$
BEGIN
    RAISE NOTICE 'ARES Autonomous Trading System schema deployed successfully';
    RAISE NOTICE 'Tables created: sandbox_trades, trading_performance, market_data_cache, strategy_mutations, risk_events';
    RAISE NOTICE 'Starting balance: $10,000 USD (sandbox)';
    RAISE NOTICE 'Auto top-up: Disabled by default (user can enable via checkbox)';
END $$;
