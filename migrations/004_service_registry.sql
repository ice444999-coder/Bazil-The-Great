-- Service Registry for Persistent Service Tracking
-- Phase 1 of Modular Architecture Implementation
-- Created: 2025-10-16

CREATE TABLE IF NOT EXISTS service_registry (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    version VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'offline',
    port INTEGER,
    health_url VARCHAR(255),
    last_heartbeat TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for fast queries
CREATE INDEX IF NOT EXISTS idx_service_status ON service_registry(status);
CREATE INDEX IF NOT EXISTS idx_service_heartbeat ON service_registry(last_heartbeat);

-- Auto-update updated_at timestamp on changes
CREATE OR REPLACE FUNCTION update_service_registry_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER service_registry_updated
BEFORE UPDATE ON service_registry
FOR EACH ROW
EXECUTE FUNCTION update_service_registry_timestamp();

-- Initial service entries (will be updated by actual services on startup)
INSERT INTO service_registry (name, version, status, port, health_url) VALUES
('ares-api', '1.0.0', 'offline', 8080, 'http://localhost:8080/health'),
('ollama', '0.1.0', 'offline', 11434, 'http://localhost:11434/api/health')
ON CONFLICT (name) DO NOTHING;

-- Success message
SELECT 'Service registry table created successfully' AS status;
