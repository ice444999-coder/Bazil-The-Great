-- Migration: 005_enhance_trades_for_sandbox.sql
-- Description: Add fields for sandbox trading with proper indexes
-- Date: 2025-10-12

-- Add new columns for sandbox trading
ALTER TABLE trades ADD COLUMN IF NOT EXISTS trade_id VARCHAR(100);
ALTER TABLE trades ADD COLUMN IF NOT EXISTS amount DOUBLE PRECISION DEFAULT 0;
ALTER TABLE trades ADD COLUMN IF NOT EXISTS exit_price DOUBLE PRECISION;
ALTER TABLE trades ADD COLUMN IF NOT EXISTS strategy VARCHAR(50);
ALTER TABLE trades ADD COLUMN IF NOT EXISTS reasoning TEXT;
ALTER TABLE trades ADD COLUMN IF NOT EXISTS profit_loss DOUBLE PRECISION DEFAULT 0;
ALTER TABLE trades ADD COLUMN IF NOT EXISTS profit_loss_pct DOUBLE PRECISION DEFAULT 0;
ALTER TABLE trades ADD COLUMN IF NOT EXISTS fee DOUBLE PRECISION DEFAULT 0;
ALTER TABLE trades ADD COLUMN IF NOT EXISTS transaction_hash VARCHAR(64);
ALTER TABLE trades ADD COLUMN IF NOT EXISTS executed_at TIMESTAMP;
ALTER TABLE trades ADD COLUMN IF NOT EXISTS exited_at TIMESTAMP;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_trades_user_status ON trades(user_id, status);
CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);
CREATE INDEX IF NOT EXISTS idx_trades_executed_at ON trades(executed_at);
CREATE INDEX IF NOT EXISTS idx_trades_exited_at ON trades(exited_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_trades_trade_id ON trades(trade_id);

-- Create index for finding open trades quickly
CREATE INDEX IF NOT EXISTS idx_trades_open ON trades(user_id, status) WHERE status = 'open';

-- Create index for performance metrics queries
CREATE INDEX IF NOT EXISTS idx_trades_closed_user ON trades(user_id, exited_at) WHERE status = 'closed';

-- Add virtual_balance to users table if not exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS virtual_balance DOUBLE PRECISION DEFAULT 10000.00;

-- Create index on virtual_balance for fast balance lookups
CREATE INDEX IF NOT EXISTS idx_users_virtual_balance ON users(virtual_balance);

-- Add comments for documentation
COMMENT ON COLUMN trades.trade_id IS 'Unique sandbox trade identifier (e.g., SANDBOX_1234567_1)';
COMMENT ON COLUMN trades.amount IS 'Amount of base asset traded';
COMMENT ON COLUMN trades.exit_price IS 'Price at which position was closed';
COMMENT ON COLUMN trades.strategy IS 'Trading strategy used (Momentum, MeanReversion, etc)';
COMMENT ON COLUMN trades.reasoning IS 'AI reasoning for the trade decision';
COMMENT ON COLUMN trades.profit_loss IS 'Profit or loss in quote currency (USDC)';
COMMENT ON COLUMN trades.profit_loss_pct IS 'Profit or loss as percentage';
COMMENT ON COLUMN trades.fee IS 'Total trading fees paid (open + close)';
COMMENT ON COLUMN trades.transaction_hash IS 'SHA256 hash for audit trail';
COMMENT ON COLUMN trades.executed_at IS 'Timestamp when trade was opened';
COMMENT ON COLUMN trades.exited_at IS 'Timestamp when trade was closed';

-- Create trades_archive table for old closed trades
CREATE TABLE IF NOT EXISTS trades_archive (LIKE trades INCLUDING ALL);

-- Add trigger to automatically archive trades older than 90 days
-- (This can be run manually or via cron job)
CREATE OR REPLACE FUNCTION archive_old_trades() RETURNS INTEGER AS $$
DECLARE
    archived_count INTEGER;
BEGIN
    INSERT INTO trades_archive
    SELECT * FROM trades
    WHERE status = 'closed' 
    AND exited_at < NOW() - INTERVAL '90 days';
    
    GET DIAGNOSTICS archived_count = ROW_COUNT;
    
    DELETE FROM trades
    WHERE status = 'closed'
    AND exited_at < NOW() - INTERVAL '90 days';
    
    RETURN archived_count;
END;
$$ LANGUAGE plpgsql;

-- Create performance view for quick metrics
CREATE OR REPLACE VIEW user_trading_metrics AS
SELECT 
    user_id,
    COUNT(*) FILTER (WHERE status = 'closed') as total_trades,
    COUNT(*) FILTER (WHERE status = 'closed' AND profit_loss > 0) as winning_trades,
    COUNT(*) FILTER (WHERE status = 'closed' AND profit_loss < 0) as losing_trades,
    ROUND(
        (COUNT(*) FILTER (WHERE status = 'closed' AND profit_loss > 0)::DECIMAL / 
         NULLIF(COUNT(*) FILTER (WHERE status = 'closed'), 0) * 100)::NUMERIC, 
        2
    ) as win_rate_pct,
    ROUND(SUM(profit_loss)::NUMERIC, 2) as total_profit_loss,
    ROUND(MAX(profit_loss)::NUMERIC, 2) as best_trade,
    ROUND(MIN(profit_loss)::NUMERIC, 2) as worst_trade,
    ROUND(AVG(profit_loss)::NUMERIC, 2) as avg_profit_loss,
    COUNT(*) FILTER (WHERE status = 'open') as open_positions
FROM trades
GROUP BY user_id;

COMMENT ON VIEW user_trading_metrics IS 'Real-time trading performance metrics per user';
