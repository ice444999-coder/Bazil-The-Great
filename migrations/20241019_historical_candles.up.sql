-- Migration: Create historical_candles table for caching Binance data
-- Purpose: Cache historical candlestick data to reduce API calls and improve performance
-- Created: 2024-10-19

CREATE TABLE IF NOT EXISTS historical_candles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,                    -- Trading pair (e.g., 'BTCUSDT')
    interval TEXT NOT NULL,                  -- Timeframe (e.g., '1m', '5m', '1h')
    timestamp INTEGER NOT NULL,              -- Candle open time (Unix milliseconds)
    open REAL NOT NULL,                      -- Opening price
    high REAL NOT NULL,                      -- Highest price
    low REAL NOT NULL,                       -- Lowest price
    close REAL NOT NULL,                     -- Closing price
    volume REAL NOT NULL,                    -- Trading volume
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, interval, timestamp)      -- Prevent duplicate candles
);

-- Index for fast lookups by symbol, interval, and time range
CREATE INDEX IF NOT EXISTS idx_historical_candles_lookup 
ON historical_candles(symbol, interval, timestamp);

-- Index for efficient time-based queries
CREATE INDEX IF NOT EXISTS idx_historical_candles_timestamp 
ON historical_candles(timestamp);

-- Index for cleanup queries (remove old data)
CREATE INDEX IF NOT EXISTS idx_historical_candles_created_at 
ON historical_candles(created_at);

-- Create view for easy candle aggregation
CREATE VIEW IF NOT EXISTS v_candle_stats AS
SELECT 
    symbol,
    interval,
    COUNT(*) as total_candles,
    MIN(timestamp) as first_candle,
    MAX(timestamp) as last_candle,
    MIN(created_at) as first_cached,
    MAX(created_at) as last_cached
FROM historical_candles
GROUP BY symbol, interval;
