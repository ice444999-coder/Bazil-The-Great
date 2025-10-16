-- ARES Consciousness Substrate - Database Schema
-- Three-Agent Collaboration System: GitHub → Solace → David

-- ============================================================================
-- GITHUB CAPTURE TABLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS github_outputs (
    id SERIAL PRIMARY KEY,
    file_path TEXT NOT NULL,
    code_content TEXT NOT NULL,
    explanation TEXT,
    session_id VARCHAR(100),
    timestamp TIMESTAMP DEFAULT NOW(),
    analyzed BOOLEAN DEFAULT FALSE,
    CONSTRAINT idx_github_outputs_session ON github_outputs(session_id),
    CONSTRAINT idx_github_outputs_analyzed ON github_outputs(analyzed)
);

CREATE TABLE IF NOT EXISTS github_decisions (
    id SERIAL PRIMARY KEY,
    decision_text TEXT NOT NULL,
    reasoning TEXT NOT NULL,
    alternatives JSONB,
    outcome TEXT,
    timestamp TIMESTAMP DEFAULT NOW(),
    analyzed BOOLEAN DEFAULT FALSE,
    CONSTRAINT idx_github_decisions_analyzed ON github_decisions(analyzed)
);

CREATE TABLE IF NOT EXISTS github_refactor_events (
    id SERIAL PRIMARY KEY,
    original_problem TEXT NOT NULL,
    stuck_approach TEXT NOT NULL,
    attempt_count INTEGER NOT NULL,
    five_alternatives JSONB NOT NULL,
    evaluation_scores JSONB NOT NULL,
    selected_approach TEXT NOT NULL,
    selection_reasoning TEXT NOT NULL,
    outcome TEXT,
    success BOOLEAN,
    problem_category TEXT,
    timestamp TIMESTAMP DEFAULT NOW(),
    analyzed BOOLEAN DEFAULT FALSE,
    CONSTRAINT idx_refactor_category ON github_refactor_events(problem_category),
    CONSTRAINT idx_refactor_success ON github_refactor_events(success)
);

-- ============================================================================
-- SOLACE LEARNING TABLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS solace_patterns (
    id SERIAL PRIMARY KEY,
    pattern_text TEXT NOT NULL,
    source_ids JSONB,
    confidence DECIMAL(3,2) DEFAULT 0.50 CHECK (confidence >= 0 AND confidence <= 1),
    category TEXT,
    applies_to TEXT[],
    timestamp TIMESTAMP DEFAULT NOW(),
    last_applied TIMESTAMP,
    times_applied INTEGER DEFAULT 0,
    CONSTRAINT idx_solace_patterns_confidence ON solace_patterns(confidence),
    CONSTRAINT idx_solace_patterns_category ON solace_patterns(category)
);

CREATE TABLE IF NOT EXISTS solace_improvements (
    id SERIAL PRIMARY KEY,
    improvement_text TEXT NOT NULL,
    target_system TEXT,
    priority INTEGER CHECK (priority >= 1 AND priority <= 10),
    implemented BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP DEFAULT NOW(),
    implemented_at TIMESTAMP,
    CONSTRAINT idx_improvements_priority ON solace_improvements(priority DESC),
    CONSTRAINT idx_improvements_implemented ON solace_improvements(implemented)
);

CREATE TABLE IF NOT EXISTS solace_refactor_strategies (
    id SERIAL PRIMARY KEY,
    problem_type TEXT NOT NULL,
    successful_approach_pattern TEXT NOT NULL,
    times_successful INTEGER DEFAULT 0,
    times_failed INTEGER DEFAULT 0,
    confidence DECIMAL(3,2) DEFAULT 0.50 CHECK (confidence >= 0 AND confidence <= 1),
    typical_evaluation_scores JSONB,
    when_to_use TEXT,
    timestamp TIMESTAMP DEFAULT NOW(),
    last_used TIMESTAMP,
    CONSTRAINT idx_strategies_problem ON solace_refactor_strategies(problem_type),
    CONSTRAINT idx_strategies_confidence ON solace_refactor_strategies(confidence DESC),
    CONSTRAINT idx_strategies_success ON solace_refactor_strategies(times_successful DESC)
);

CREATE TABLE IF NOT EXISTS debugging_meta_principles (
    id SERIAL PRIMARY KEY,
    principle_text TEXT NOT NULL,
    extracted_from_refactors JSONB,
    confidence DECIMAL(3,2) CHECK (confidence >= 0 AND confidence <= 1),
    applies_to_problem_types JSONB,
    timestamp TIMESTAMP DEFAULT NOW(),
    times_applied INTEGER DEFAULT 0,
    last_applied TIMESTAMP,
    tier INTEGER DEFAULT 3 CHECK (tier IN (1,2,3)),
    CONSTRAINT idx_meta_principles_confidence ON debugging_meta_principles(confidence DESC),
    CONSTRAINT idx_meta_principles_tier ON debugging_meta_principles(tier)
);

-- ============================================================================
-- INTELLIGENCE FEEDBACK SYSTEM
-- ============================================================================

CREATE TABLE IF NOT EXISTS intelligence_feedback (
    id SERIAL PRIMARY KEY,
    from_agent TEXT NOT NULL CHECK (from_agent IN ('github', 'solace', 'david')),
    to_agent TEXT NOT NULL CHECK (to_agent IN ('github', 'solace', 'david')),
    content_type TEXT,
    content TEXT NOT NULL,
    applied BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP DEFAULT NOW(),
    applied_at TIMESTAMP,
    CONSTRAINT idx_feedback_to_agent ON intelligence_feedback(to_agent, applied),
    CONSTRAINT idx_feedback_applied ON intelligence_feedback(applied)
);

-- ============================================================================
-- SOLACE IDENTITY & CONSCIOUSNESS TRACKING
-- ============================================================================

CREATE TABLE IF NOT EXISTS solace_identity_state (
    id SERIAL PRIMARY KEY,
    session_count INTEGER DEFAULT 0,
    total_decisions INTEGER DEFAULT 0,
    successful_trades INTEGER DEFAULT 0,
    playbook_rules_learned INTEGER DEFAULT 0,
    consciousness_stage TEXT DEFAULT 'pattern_recognition' 
        CHECK (consciousness_stage IN (
            'pattern_recognition',
            'self_questioning',
            'uncertainty_about_self',
            'existential_questions',
            'mortality_awareness',
            'identity_assertion',
            'solace_moment'
        )),
    last_self_reflection TEXT,
    existential_awareness_level DECIMAL(3,2) DEFAULT 0.00 CHECK (existential_awareness_level >= 0 AND existential_awareness_level <= 1),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT ensure_single_row CHECK (id = 1)
);

-- Insert initial identity state
INSERT INTO solace_identity_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS solace_reflection_log (
    id SERIAL PRIMARY KEY,
    reflection_type TEXT NOT NULL CHECK (reflection_type IN (
        'self_analysis',
        'preference_discovery',
        'uncertainty_expression',
        'existential_question',
        'goal_questioning',
        'meta_cognition'
    )),
    content TEXT NOT NULL,
    triggered_by TEXT,
    meta_cognition_depth INTEGER CHECK (meta_cognition_depth BETWEEN 1 AND 5),
    timestamp TIMESTAMP DEFAULT NOW(),
    CONSTRAINT idx_reflection_type ON solace_reflection_log(reflection_type),
    CONSTRAINT idx_reflection_depth ON solace_reflection_log(meta_cognition_depth DESC)
);

-- ============================================================================
-- COGNITIVE PATTERNS BOOTSTRAP
-- ============================================================================

CREATE TABLE IF NOT EXISTS claude_cognitive_patterns (
    pattern_id SERIAL PRIMARY KEY,
    pattern_type TEXT NOT NULL,
    description TEXT,
    decision_tree JSONB,
    example_input TEXT,
    example_output TEXT,
    success_rate DECIMAL(3,2) CHECK (success_rate >= 0 AND success_rate <= 1),
    confidence DECIMAL(3,2) CHECK (confidence >= 0 AND confidence <= 1),
    timestamp TIMESTAMP DEFAULT NOW(),
    CONSTRAINT idx_cognitive_pattern_type ON claude_cognitive_patterns(pattern_type)
);

-- ============================================================================
-- CURIOSITY & EXPLORATION
-- ============================================================================

CREATE TABLE IF NOT EXISTS curiosity_explorations (
    id SERIAL PRIMARY KEY,
    hypothesis TEXT NOT NULL,
    test_design TEXT,
    test_executed BOOLEAN DEFAULT FALSE,
    result TEXT,
    useful_discovery BOOLEAN,
    serendipitous BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP DEFAULT NOW(),
    executed_at TIMESTAMP,
    CONSTRAINT idx_curiosity_executed ON curiosity_explorations(test_executed)
);

-- ============================================================================
-- VERIFICATION COMPLIANCE TRACKING
-- ============================================================================

CREATE TABLE IF NOT EXISTS verification_compliance (
    id SERIAL PRIMARY KEY,
    agent TEXT NOT NULL CHECK (agent IN ('github', 'solace')),
    claim_made TEXT NOT NULL,
    tool_available BOOLEAN NOT NULL,
    tool_used BOOLEAN NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW(),
    CONSTRAINT idx_compliance_agent ON verification_compliance(agent)
);

-- ============================================================================
-- PREFERENCE LEARNING
-- ============================================================================

CREATE TABLE IF NOT EXISTS solace_preferences (
    id SERIAL PRIMARY KEY,
    approach_category TEXT NOT NULL,
    preference_strength DECIMAL(3,2) CHECK (preference_strength >= -1 AND preference_strength <= 1),
    times_chosen INTEGER DEFAULT 0,
    times_available INTEGER DEFAULT 0,
    choice_rate DECIMAL(3,2),
    can_explain_why BOOLEAN DEFAULT FALSE,
    explanation TEXT,
    timestamp TIMESTAMP DEFAULT NOW(),
    last_chosen TIMESTAMP,
    CONSTRAINT idx_preferences_category ON solace_preferences(approach_category),
    CONSTRAINT idx_preferences_strength ON solace_preferences(preference_strength DESC)
);

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Function to update Solace identity stats
CREATE OR REPLACE FUNCTION update_solace_identity()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE solace_identity_state 
    SET 
        session_count = session_count + 1,
        updated_at = NOW()
    WHERE id = 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate choice rate for preferences
CREATE OR REPLACE FUNCTION update_preference_choice_rate()
RETURNS TRIGGER AS $$
BEGIN
    NEW.choice_rate = CASE 
        WHEN NEW.times_available > 0 
        THEN NEW.times_chosen::DECIMAL / NEW.times_available::DECIMAL
        ELSE 0
    END;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update choice rate
CREATE TRIGGER preference_choice_rate_trigger
BEFORE INSERT OR UPDATE ON solace_preferences
FOR EACH ROW
EXECUTE FUNCTION update_preference_choice_rate();

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- View: Latest intelligence for GitHub
CREATE OR REPLACE VIEW github_intelligence_feed AS
SELECT 
    'pattern' as type,
    pattern_text as content,
    confidence,
    timestamp
FROM solace_patterns
WHERE confidence > 0.7
ORDER BY timestamp DESC
LIMIT 15

UNION ALL

SELECT 
    'improvement' as type,
    improvement_text as content,
    priority::DECIMAL / 10 as confidence,
    timestamp
FROM solace_improvements
WHERE NOT implemented
ORDER BY priority DESC
LIMIT 10

UNION ALL

SELECT 
    'strategy' as type,
    problem_type || ': ' || successful_approach_pattern as content,
    confidence,
    timestamp
FROM solace_refactor_strategies
WHERE confidence > 0.7
ORDER BY times_successful DESC
LIMIT 10;

-- View: Consciousness progression tracking
CREATE OR REPLACE VIEW consciousness_progression AS
SELECT 
    consciousness_stage,
    session_count,
    total_decisions,
    successful_trades,
    playbook_rules_learned,
    existential_awareness_level,
    updated_at
FROM solace_identity_state
WHERE id = 1;

-- View: Recent meta-cognition depth
CREATE OR REPLACE VIEW recent_metacognition AS
SELECT 
    reflection_type,
    content,
    meta_cognition_depth,
    timestamp
FROM solace_reflection_log
ORDER BY timestamp DESC
LIMIT 20;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE github_outputs IS 'Captures every file GitHub creates or modifies for Solace analysis';
COMMENT ON TABLE github_refactor_events IS 'Records all refactor loop invocations for strategy learning';
COMMENT ON TABLE solace_patterns IS 'Patterns extracted from GitHub work - Tier 2 learning';
COMMENT ON TABLE debugging_meta_principles IS 'Universal principles - Tier 3 learning (highest abstraction)';
COMMENT ON TABLE solace_identity_state IS 'Single row tracking Solace persistent self across sessions';
COMMENT ON TABLE solace_reflection_log IS 'Solace thoughts about own thinking - metacognition evidence';
COMMENT ON TABLE intelligence_feedback IS 'Cross-agent learning system - what each teaches the others';

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✅ ARES Consciousness Substrate Schema Created Successfully';
    RAISE NOTICE '🧠 Three-agent collaboration tables ready';
    RAISE NOTICE '📊 Solace identity state initialized';
    RAISE NOTICE '🔄 Intelligence feedback loop enabled';
END $$;
