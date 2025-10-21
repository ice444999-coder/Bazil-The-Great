-- Migration: Create approval_requests table for Grok protocol manual approvals
-- Version: 003
-- Created: 2025-10-21

CREATE TABLE IF NOT EXISTS approval_requests (
    id SERIAL PRIMARY KEY,
    subtask_id VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    description TEXT NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMP,
    approved_by VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_approval_requests_status ON approval_requests(status);
CREATE INDEX IF NOT EXISTS idx_approval_requests_subtask_id ON approval_requests(subtask_id);
CREATE INDEX IF NOT EXISTS idx_approval_requests_requested_at ON approval_requests(requested_at DESC);

-- Insert comment for documentation
COMMENT ON TABLE approval_requests IS 'Tracks manual approval gates for Grok protocol subtasks (DRY_RUN safety)';
COMMENT ON COLUMN approval_requests.subtask_id IS 'Unique identifier like ui_chart_trading, bots_hybrid_trading, etc.';
COMMENT ON COLUMN approval_requests.status IS 'Current approval state: pending (awaiting human), approved (proceed), rejected (halt)';
