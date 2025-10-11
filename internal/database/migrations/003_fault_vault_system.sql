-- Fault Vault System - Complete Database Schema
-- Migration 003: Fault Vault + Chat Persistence

-- Message Persistence for Chat
CREATE TABLE IF NOT EXISTS memory_snapshots (
    message_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_message TEXT NOT NULL,
    assistant_response TEXT,
    context JSONB,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memory_snapshots_session ON memory_snapshots(session_id);
CREATE INDEX idx_memory_snapshots_timestamp ON memory_snapshots(timestamp DESC);
CREATE INDEX idx_memory_snapshots_user ON memory_snapshots(user_id);

-- Fault Vault Sessions
CREATE TABLE IF NOT EXISTS fault_vault_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    context_type TEXT NOT NULL CHECK (context_type IN ('vscode_claude', 'ares_claude', 'ares_autonomous')),
    session_summary TEXT,
    active BOOLEAN DEFAULT TRUE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    metadata JSONB
);

CREATE INDEX idx_fault_vault_sessions_active ON fault_vault_sessions(active) WHERE active = TRUE;
CREATE INDEX idx_fault_vault_sessions_started ON fault_vault_sessions(started_at DESC);
CREATE INDEX idx_fault_vault_sessions_context ON fault_vault_sessions(context_type);

-- Fault Vault Actions
CREATE TABLE IF NOT EXISTS fault_vault_actions (
    action_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES fault_vault_sessions(session_id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor TEXT NOT NULL,
    action_type TEXT NOT NULL CHECK (action_type IN (
        'code_change', 'build', 'test', 'debug', 'crash',
        'decision', 'feature_start', 'feature_complete',
        'checkpoint', 'rollback'
    )),
    file_path TEXT,
    function_name TEXT,
    intent TEXT,
    changes_made TEXT,
    result TEXT CHECK (result IN ('success', 'partial', 'failure', 'crash', 'pending')),
    error_message TEXT,
    stack_trace TEXT,
    next_steps TEXT,
    related_actions UUID[],
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_fault_vault_actions_session ON fault_vault_actions(session_id);
CREATE INDEX idx_fault_vault_actions_timestamp ON fault_vault_actions(timestamp DESC);
CREATE INDEX idx_fault_vault_actions_type ON fault_vault_actions(action_type);
CREATE INDEX idx_fault_vault_actions_result ON fault_vault_actions(result);
CREATE INDEX idx_fault_vault_actions_actor ON fault_vault_actions(actor);

-- Fault Vault Context
CREATE TABLE IF NOT EXISTS fault_vault_context (
    context_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES fault_vault_sessions(session_id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    conversation_snapshot JSONB,
    user_intent TEXT,
    system_state JSONB,
    memory_refs UUID[],
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_fault_vault_context_session ON fault_vault_context(session_id);
CREATE INDEX idx_fault_vault_context_timestamp ON fault_vault_context(timestamp DESC);

-- Fault Vault Learnings
CREATE TABLE IF NOT EXISTS fault_vault_learnings (
    learning_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern TEXT NOT NULL,
    outcome TEXT NOT NULL CHECK (outcome IN ('success', 'failure')),
    reason TEXT,
    confidence FLOAT DEFAULT 0.5 CHECK (confidence >= 0 AND confidence <= 1),
    times_observed INT DEFAULT 1,
    last_seen TIMESTAMPTZ DEFAULT NOW(),
    recommendation TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_fault_vault_learnings_pattern ON fault_vault_learnings(pattern);
CREATE INDEX idx_fault_vault_learnings_confidence ON fault_vault_learnings(confidence DESC);
CREATE INDEX idx_fault_vault_learnings_last_seen ON fault_vault_learnings(last_seen DESC);

-- Add comments
COMMENT ON TABLE memory_snapshots IS 'Stores all chat messages for persistence and history loading';
COMMENT ON TABLE fault_vault_sessions IS 'Tracks development sessions across all three actors';
COMMENT ON TABLE fault_vault_actions IS 'Logs every action taken during development with full context';
COMMENT ON TABLE fault_vault_context IS 'Captures snapshots of system state for crash recovery';
COMMENT ON TABLE fault_vault_learnings IS 'Extracted patterns and learnings from observed actions';
