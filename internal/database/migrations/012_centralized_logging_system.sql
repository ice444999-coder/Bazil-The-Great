-- Migration: 012_centralized_logging_system.sql
-- Creates system_logs table for centralized logging
-- Phase: Option B - Consolidate & Clean Up

-- Create system_logs table
CREATE TABLE IF NOT EXISTS system_logs (
    id SERIAL PRIMARY KEY,
    service VARCHAR(50) NOT NULL,           -- Service name (ARES_API, GLASS_BOX, SOLACE, etc.)
    level VARCHAR(20) NOT NULL,             -- Log level (INFO, WARN, ERROR, DEBUG)
    message TEXT NOT NULL,                  -- Log message
    event_type VARCHAR(50),                 -- Optional event type (trade_executed, etc.)
    event_data JSONB,                       -- Optional event payload
    trace_id VARCHAR(64),                   -- Optional correlation ID for distributed tracing
    user_id INTEGER REFERENCES users(id),  -- Optional user context
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_system_logs_service ON system_logs(service, created_at DESC);
CREATE INDEX idx_system_logs_level ON system_logs(level, created_at DESC);
CREATE INDEX idx_system_logs_event_type ON system_logs(event_type) WHERE event_type IS NOT NULL;
CREATE INDEX idx_system_logs_created_at ON system_logs(created_at DESC);
CREATE INDEX idx_system_logs_trace_id ON system_logs(trace_id) WHERE trace_id IS NOT NULL;

-- Partitioning hint: Consider partitioning by created_at for large volumes
-- ALTER TABLE system_logs PARTITION BY RANGE (created_at);

-- Create view for recent logs (last 24 hours)
CREATE OR REPLACE VIEW recent_logs AS
SELECT 
    id,
    service,
    level,
    message,
    event_type,
    created_at,
    EXTRACT(EPOCH FROM (NOW() - created_at)) AS age_seconds
FROM system_logs
WHERE created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;

-- Create view for error summary
CREATE OR REPLACE VIEW error_summary AS
SELECT 
    service,
    level,
    COUNT(*) as count,
    MAX(created_at) as last_occurrence
FROM system_logs
WHERE level IN ('ERROR', 'WARN')
    AND created_at > NOW() - INTERVAL '24 hours'
GROUP BY service, level
ORDER BY count DESC;

-- Auto-cleanup function: Delete logs older than 30 days
CREATE OR REPLACE FUNCTION cleanup_old_logs() 
RETURNS void AS $$
BEGIN
    DELETE FROM system_logs 
    WHERE created_at < NOW() - INTERVAL '30 days';
    
    RAISE NOTICE 'Old logs cleaned up';
END;
$$ LANGUAGE plpgsql;

-- Schedule cleanup (requires pg_cron extension - optional)
-- SELECT cron.schedule('cleanup-logs', '0 2 * * *', 'SELECT cleanup_old_logs()');

-- Grant permissions
GRANT SELECT, INSERT ON system_logs TO ares_api;
GRANT USAGE, SELECT ON SEQUENCE system_logs_id_seq TO ares_api;

-- Insert migration record
INSERT INTO schema_migrations (version, name, applied_at) 
VALUES (12, '012_centralized_logging_system', NOW())
ON CONFLICT (version) DO NOTHING;

COMMENT ON TABLE system_logs IS 'Centralized logging for all ARES services';
COMMENT ON COLUMN system_logs.service IS 'Service that generated the log (ARES_API, GLASS_BOX, SOLACE, TRADING, etc.)';
COMMENT ON COLUMN system_logs.level IS 'Log level: DEBUG, INFO, WARN, ERROR';
COMMENT ON COLUMN system_logs.event_type IS 'Optional event type for event-driven logs';
COMMENT ON COLUMN system_logs.event_data IS 'Optional JSON payload for events';
COMMENT ON COLUMN system_logs.trace_id IS 'Distributed tracing correlation ID';
