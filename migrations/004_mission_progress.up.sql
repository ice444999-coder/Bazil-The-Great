CREATE TABLE IF NOT EXISTS mission_progress (
    id SERIAL PRIMARY KEY,
    phase INTEGER NOT NULL DEFAULT 1,
    percentage INTEGER NOT NULL DEFAULT 0 CHECK (percentage >= 0 AND percentage <= 100),
    status VARCHAR(50) NOT NULL DEFAULT 'initializing',
    subtasks_completed INTEGER NOT NULL DEFAULT 0 CHECK (subtasks_completed >= 0),
    subtasks_total INTEGER NOT NULL DEFAULT 12,
    last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for fast phase lookup
CREATE INDEX idx_mission_progress_phase ON mission_progress(phase);

-- Insert initial Phase 1 progress
INSERT INTO mission_progress (phase, percentage, status, subtasks_completed, subtasks_total, last_updated)
VALUES (1, 0, 'initializing', 0, 12, NOW())
ON CONFLICT DO NOTHING;
