-- Migration: Add strategy_name column to sandbox_trades table
-- Purpose: Enable multi-strategy tracking for sandbox trades
-- Date: 2025-10-19

-- Add strategy_name column
ALTER TABLE sandbox_trades 
ADD COLUMN IF NOT EXISTS strategy_name VARCHAR(50);

-- Create index for efficient strategy-based queries
CREATE INDEX IF NOT EXISTS idx_sandbox_strategy 
ON sandbox_trades(strategy_name);

-- Create composite index for strategy + timestamp queries
CREATE INDEX IF NOT EXISTS idx_sandbox_strategy_time 
ON sandbox_trades(strategy_name, created_at DESC);

-- Add comment for documentation
COMMENT ON COLUMN sandbox_trades.strategy_name IS 'Name of the strategy that executed this trade (RSI, MACD, Trend, Support, Volume)';
