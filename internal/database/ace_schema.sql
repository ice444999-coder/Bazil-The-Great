-- ACE Framework Tables
-- These tables support the pattern-based decision making system

-- 1. Cognitive Patterns Table
CREATE TABLE IF NOT EXISTS cognitive_patterns (
    pattern_id SERIAL PRIMARY KEY,
    pattern_name VARCHAR(255) NOT NULL UNIQUE,
    pattern_category VARCHAR(100) NOT NULL,
    description TEXT,
    trigger_conditions TEXT,
    example_input TEXT,
    example_output TEXT,
    example_reasoning TEXT,
    confidence_score DECIMAL(5,4) DEFAULT 0.5000,
    usage_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP
);

-- Index for fast category lookups
CREATE INDEX IF NOT EXISTS idx_patterns_category ON cognitive_patterns(pattern_category);
CREATE INDEX IF NOT EXISTS idx_patterns_confidence ON cognitive_patterns(confidence_score DESC);

-- 2. Playbook Rules Table
CREATE TABLE IF NOT EXISTS playbook_rules (
    rule_id SERIAL PRIMARY KEY,
    rule_name VARCHAR(255) NOT NULL,
    rule_category VARCHAR(100) NOT NULL,
    trigger_conditions TEXT,
    application_example TEXT,
    confidence_score DECIMAL(5,4) DEFAULT 0.5000,
    usage_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    source_pattern_ids INTEGER[],
    parent_rule_id INTEGER REFERENCES playbook_rules(rule_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP,
    last_success_rate DECIMAL(5,4),
    consecutive_low_checks INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_rules_category ON playbook_rules(rule_category);
CREATE INDEX IF NOT EXISTS idx_rules_confidence ON playbook_rules(confidence_score DESC);

-- 3. Decisions Table (log of all decisions)
CREATE TABLE IF NOT EXISTS decisions (
    decision_id SERIAL PRIMARY KEY,
    decision_type VARCHAR(100) NOT NULL,
    input_context JSONB,
    patterns_considered INTEGER[],
    rules_applied INTEGER[],
    reasoning_trace TEXT,
    decision_output JSONB,
    confidence_level DECIMAL(5,4),
    initial_quality_score DECIMAL(5,4),
    refactor_triggered BOOLEAN DEFAULT FALSE,
    final_quality_score DECIMAL(5,4),
    tools_invoked TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_decisions_type ON decisions(decision_type);
CREATE INDEX IF NOT EXISTS idx_decisions_created ON decisions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_decisions_quality ON decisions(final_quality_score DESC);

-- 4. Quality Scores Table (detailed quality assessments)
CREATE TABLE IF NOT EXISTS quality_scores (
    score_id SERIAL PRIMARY KEY,
    decision_id INTEGER REFERENCES decisions(decision_id),
    specificity_score DECIMAL(5,4),
    actionability_score DECIMAL(5,4),
    tool_usage_score DECIMAL(5,4),
    context_awareness_score DECIMAL(5,4),
    mission_alignment_score DECIMAL(5,4),
    composite_quality_score DECIMAL(5,4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_quality_decision ON quality_scores(decision_id);

-- Verify tables created
SELECT 
    'cognitive_patterns' AS table_name, 
    COUNT(*) AS row_count 
FROM cognitive_patterns
UNION ALL
SELECT 
    'playbook_rules' AS table_name, 
    COUNT(*) AS row_count 
FROM playbook_rules
UNION ALL
SELECT 
    'decisions' AS table_name, 
    COUNT(*) AS row_count 
FROM decisions
UNION ALL
SELECT 
    'quality_scores' AS table_name, 
    COUNT(*) AS row_count 
FROM quality_scores;
