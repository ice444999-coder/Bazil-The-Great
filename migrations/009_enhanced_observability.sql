-- Migration 009: Enhanced Observability System
-- Distributed tracing, metrics collection, and structured logging
-- Part of Modular Architecture Section 6

-- Structured service logs with trace ID support
CREATE TABLE IF NOT EXISTS service_logs (
    id BIGSERIAL PRIMARY KEY,
    trace_id UUID,
    span_id VARCHAR(50),
    parent_span_id VARCHAR(50),
    service_name VARCHAR(100) NOT NULL,
    log_level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    source_file VARCHAR(255),
    source_line INTEGER
);

-- Indexes for fast log queries
CREATE INDEX idx_logs_trace ON service_logs(trace_id, timestamp DESC);
CREATE INDEX idx_logs_service ON service_logs(service_name, timestamp DESC);
CREATE INDEX idx_logs_level ON service_logs(log_level, timestamp DESC);
CREATE INDEX idx_logs_timestamp ON service_logs(timestamp DESC);

-- Service metrics collection
CREATE TABLE IF NOT EXISTS service_metrics (
    id BIGSERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_type VARCHAR(20) NOT NULL, -- counter, gauge, histogram, summary
    metric_value DOUBLE PRECISION NOT NULL,
    labels JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for metric aggregation
CREATE INDEX idx_metrics_service_name ON service_metrics(service_name, metric_name, timestamp DESC);
CREATE INDEX idx_metrics_timestamp ON service_metrics(timestamp DESC);

-- Service performance spans (distributed tracing)
CREATE TABLE IF NOT EXISTS service_spans (
    id BIGSERIAL PRIMARY KEY,
    trace_id UUID NOT NULL,
    span_id VARCHAR(50) NOT NULL UNIQUE,
    parent_span_id VARCHAR(50),
    service_name VARCHAR(100) NOT NULL,
    operation_name VARCHAR(200) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    status VARCHAR(20), -- ok, error, timeout
    tags JSONB,
    logs JSONB
);

-- Indexes for distributed tracing
CREATE INDEX idx_spans_trace ON service_spans(trace_id, start_time DESC);
CREATE INDEX idx_spans_service ON service_spans(service_name, start_time DESC);
CREATE INDEX idx_spans_duration ON service_spans(duration_ms DESC);

-- System health metrics view
CREATE OR REPLACE VIEW v_system_health AS
SELECT 
    sr.name as service_name,
    sr.status,
    sr.version,
    sr.last_heartbeat,
    EXTRACT(EPOCH FROM (NOW() - sr.last_heartbeat)) as seconds_since_heartbeat,
    COUNT(DISTINCT sl.trace_id) as active_traces,
    COUNT(CASE WHEN sl.log_level = 'ERROR' THEN 1 END) as error_count_1h
FROM service_registry sr
LEFT JOIN service_logs sl ON sl.service_name = sr.name 
    AND sl.timestamp > NOW() - INTERVAL '1 hour'
GROUP BY sr.id, sr.name, sr.status, sr.version, sr.last_heartbeat;

-- Performance metrics view
CREATE OR REPLACE VIEW v_service_performance AS
SELECT 
    service_name,
    operation_name,
    COUNT(*) as call_count,
    AVG(duration_ms) as avg_duration_ms,
    PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY duration_ms) as p50_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms) as p95_ms,
    PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms) as p99_ms,
    COUNT(CASE WHEN status = 'error' THEN 1 END) as error_count
FROM service_spans
WHERE start_time > NOW() - INTERVAL '1 hour'
GROUP BY service_name, operation_name
ORDER BY call_count DESC;

COMMENT ON TABLE service_logs IS 'Structured logs with distributed tracing support';
COMMENT ON TABLE service_metrics IS 'Time-series metrics for all services';
COMMENT ON TABLE service_spans IS 'Distributed tracing spans for performance monitoring';
COMMENT ON VIEW v_system_health IS 'Real-time health status of all services';
COMMENT ON VIEW v_service_performance IS 'Performance metrics aggregated by service and operation';
