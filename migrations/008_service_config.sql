-- Migration 008: Service Configuration Management System
-- Enables hot-reload config for all microservices
-- Part of Modular Architecture Section 5

CREATE TABLE IF NOT EXISTS service_config (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    config_key VARCHAR(255) NOT NULL,
    config_value JSONB NOT NULL,
    description TEXT,
    is_encrypted BOOLEAN DEFAULT FALSE,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by VARCHAR(100),
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(service_name, config_key)
);

-- Index for fast config lookups
CREATE INDEX idx_service_config_lookup ON service_config(service_name, config_key);
CREATE INDEX idx_service_config_updated ON service_config(last_updated DESC);

-- Config change history for audit
CREATE TABLE IF NOT EXISTS service_config_history (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    config_key VARCHAR(255) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    changed_by VARCHAR(100),
    change_reason TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_config_history_service ON service_config_history(service_name, changed_at DESC);

-- Insert default configs for ARES_API
INSERT INTO service_config (service_name, config_key, config_value, description, updated_by) VALUES
('ares-api', 'eventbus.type', '"memory"', 'EventBus backend type: memory or redis', 'system'),
('ares-api', 'eventbus.redis_url', '"localhost:6379"', 'Redis connection URL for distributed EventBus', 'system'),
('ares-api', 'service.heartbeat_interval', '30', 'Service heartbeat interval in seconds', 'system'),
('ares-api', 'service.health_check_timeout', '5', 'Health check timeout in seconds', 'system'),
('ares-api', 'logging.level', '"info"', 'Logging level: debug, info, warn, error', 'system'),
('ares-api', 'logging.trace_enabled', 'true', 'Enable distributed tracing', 'system'),
('agent-coordinator', 'task.max_concurrent', '5', 'Max concurrent agent tasks', 'system'),
('agent-coordinator', 'task.retry_count', '3', 'Max retry count for failed tasks', 'system'),
('agent-coordinator', 'polling_interval', '5', 'Task polling interval in seconds', 'system')
ON CONFLICT (service_name, config_key) DO NOTHING;

COMMENT ON TABLE service_config IS 'Dynamic configuration for all microservices with hot-reload support';
COMMENT ON TABLE service_config_history IS 'Audit trail of configuration changes';
