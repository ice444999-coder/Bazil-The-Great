-- Strategy Versions Table
-- Stores versioned configurations for each strategy (Git-style)
CREATE TABLE IF NOT EXISTS strategy_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    strategy_name TEXT NOT NULL,
    version INTEGER NOT NULL,
    config_json TEXT NOT NULL, -- JSON blob with all strategy parameters
    code_hash TEXT, -- Hash of strategy code (for detecting code changes)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT DEFAULT 'system',
    notes TEXT, -- Version notes (why was this version created)
    backtest_result_id INTEGER, -- Link to backtest results
    is_active BOOLEAN DEFAULT 0, -- Only one version can be active per strategy
    
    UNIQUE(strategy_name, version),
    FOREIGN KEY (backtest_result_id) REFERENCES backtest_results(id)
);

CREATE INDEX idx_strategy_versions_name ON strategy_versions(strategy_name);
CREATE INDEX idx_strategy_versions_active ON strategy_versions(strategy_name, is_active);
CREATE INDEX idx_strategy_versions_created ON strategy_versions(created_at);

-- Backtest Results Table
-- Stores historical backtest results for version comparison
CREATE TABLE IF NOT EXISTS backtest_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    strategy_name TEXT NOT NULL,
    version INTEGER,
    symbol TEXT NOT NULL,
    timeframe TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    total_candles INTEGER NOT NULL,
    
    -- Performance Metrics
    total_trades INTEGER DEFAULT 0,
    winning_trades INTEGER DEFAULT 0,
    losing_trades INTEGER DEFAULT 0,
    win_rate REAL DEFAULT 0.0,
    total_pnl REAL DEFAULT 0.0,
    return_pct REAL DEFAULT 0.0,
    sharpe_ratio REAL DEFAULT 0.0,
    max_drawdown REAL DEFAULT 0.0,
    profit_factor REAL DEFAULT 0.0,
    avg_win REAL DEFAULT 0.0,
    avg_loss REAL DEFAULT 0.0,
    largest_win REAL DEFAULT 0.0,
    largest_loss REAL DEFAULT 0.0,
    avg_trade_duration_minutes INTEGER DEFAULT 0,
    
    -- Promotion Criteria
    passes_criteria BOOLEAN DEFAULT 0,
    
    execution_time_ms INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (strategy_name, version) REFERENCES strategy_versions(strategy_name, version)
);

CREATE INDEX idx_backtest_results_strategy ON backtest_results(strategy_name);
CREATE INDEX idx_backtest_results_version ON backtest_results(strategy_name, version);
CREATE INDEX idx_backtest_results_created ON backtest_results(created_at);
CREATE INDEX idx_backtest_results_performance ON backtest_results(win_rate, sharpe_ratio, total_pnl);

-- Strategy Rollback History
-- Audit trail for all version changes
CREATE TABLE IF NOT EXISTS strategy_rollback_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    strategy_name TEXT NOT NULL,
    from_version INTEGER NOT NULL,
    to_version INTEGER NOT NULL,
    reason TEXT,
    triggered_by TEXT DEFAULT 'manual', -- 'manual', 'auto', 'emergency'
    performance_before TEXT, -- JSON snapshot of metrics before rollback
    performance_after TEXT, -- JSON snapshot of metrics after rollback
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT DEFAULT 'system'
);

CREATE INDEX idx_rollback_history_strategy ON strategy_rollback_history(strategy_name);
CREATE INDEX idx_rollback_history_created ON strategy_rollback_history(created_at);

-- Insert initial versions for all existing strategies
INSERT INTO strategy_versions (strategy_name, version, config_json, notes, is_active)
VALUES 
    ('RSI_Oversold', 1, '{"rsi_period":14,"oversold_threshold":30,"position_size_pct":2.0,"stop_loss_pct":2.0,"target_profit_pct":3.0}', 'Initial production version', 1),
    ('MACD_Crossover', 1, '{"fast_period":12,"slow_period":26,"signal_period":9,"position_size_pct":2.0,"stop_loss_pct":2.0,"target_profit_pct":3.0}', 'Initial production version', 1),
    ('Trend_Following', 1, '{"ma_period":20,"atr_period":14,"atr_multiplier":2.0,"position_size_pct":2.0,"stop_loss_pct":3.0,"target_profit_pct":5.0}', 'Initial production version', 1),
    ('Support_Bounce', 1, '{"lookback_period":50,"support_threshold":0.02,"position_size_pct":2.0,"stop_loss_pct":2.0,"target_profit_pct":4.0}', 'Initial production version', 1),
    ('Volume_Breakout', 1, '{"volume_ma_period":20,"volume_multiplier":2.0,"breakout_threshold":0.01,"position_size_pct":2.0,"stop_loss_pct":2.0,"target_profit_pct":3.0}', 'Initial production version', 1);
