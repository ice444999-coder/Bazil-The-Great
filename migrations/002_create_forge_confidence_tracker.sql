-- ============================================================================
-- CRYSTAL #27.1: FORGE CONFIDENCE TRACKER - GitHub Copilot Apprenticeship
-- ============================================================================
-- Created: 2025-10-18
-- Purpose: Track FORGE's observation of GitHub Copilot code generation
--          Build confidence per pattern type through apprenticeship learning
--          Graduate patterns at 95% confidence + 20 successful examples
-- ============================================================================

CREATE TABLE IF NOT EXISTS forge_confidence_tracker (
    id SERIAL PRIMARY KEY,
    
    -- Pattern Classification
    pattern_type VARCHAR(100) NOT NULL, -- e.g., 'add_index', 'create_api_endpoint', 'write_sql_query'
    pattern_category VARCHAR(50), -- e.g., 'database', 'backend', 'frontend', 'testing'
    
    -- Task Details
    task_description TEXT NOT NULL, -- What was GitHub Copilot asked to do
    user_prompt TEXT, -- The exact prompt given to gh copilot
    
    -- Code Generation
    github_generated_code TEXT NOT NULL, -- Code from GitHub Copilot
    github_model_used VARCHAR(50), -- e.g., 'gpt-4', 'claude-3', 'codex'
    generation_timestamp TIMESTAMP DEFAULT NOW(),
    
    -- FORGE Observation
    forge_observation TEXT, -- FORGE's analysis: "I observe this pattern uses X to achieve Y"
    forge_extracted_principles TEXT, -- Key principles FORGE learned
    forge_confidence_before DECIMAL(3,2), -- FORGE's confidence before seeing this example
    forge_confidence_after DECIMAL(3,2), -- FORGE's updated confidence after observation
    
    -- Execution Results
    success BOOLEAN NOT NULL, -- Did the generated code work?
    test_passed BOOLEAN, -- Did it pass tests?
    execution_time_ms INT, -- How long did it take to run?
    error_message TEXT, -- If failed, what was the error?
    
    -- Learning Metrics
    similar_examples_count INT DEFAULT 1, -- How many times FORGE has seen this pattern
    graduation_ready BOOLEAN DEFAULT FALSE, -- >= 95% confidence + >= 20 examples
    graduated_at TIMESTAMP, -- When FORGE took over this pattern
    graduated_by VARCHAR(50), -- 'solace' when approved for graduation
    
    -- Traceability
    related_crystal_id INT, -- Which memory crystal documents this pattern
    solace_reviewed BOOLEAN DEFAULT FALSE, -- Has SOLACE reviewed this observation?
    solace_approved BOOLEAN, -- Did SOLACE approve FORGE's analysis?
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for fast queries
CREATE INDEX idx_forge_pattern_type ON forge_confidence_tracker(pattern_type);
CREATE INDEX idx_forge_graduation_ready ON forge_confidence_tracker(graduation_ready, pattern_type);
CREATE INDEX idx_forge_success ON forge_confidence_tracker(success);
CREATE INDEX idx_forge_pattern_category ON forge_confidence_tracker(pattern_category);

-- View: Graduation Dashboard
-- Shows patterns ready for FORGE to take over
CREATE OR REPLACE VIEW forge_graduation_dashboard AS
SELECT 
    pattern_type,
    pattern_category,
    COUNT(*) as total_examples,
    COUNT(*) FILTER (WHERE success = TRUE) as successful_examples,
    ROUND(100.0 * COUNT(*) FILTER (WHERE success = TRUE) / COUNT(*), 1) as success_rate_pct,
    AVG(forge_confidence_after) as avg_confidence,
    MAX(forge_confidence_after) as max_confidence,
    COUNT(*) FILTER (WHERE graduation_ready = TRUE) as ready_for_graduation,
    MAX(generation_timestamp) as last_observed
FROM forge_confidence_tracker
GROUP BY pattern_type, pattern_category
HAVING COUNT(*) >= 5 -- At least 5 examples to show pattern
ORDER BY success_rate_pct DESC, total_examples DESC;

-- View: FORGE Learning Progress
-- Tracks FORGE's confidence growth over time
CREATE OR REPLACE VIEW forge_learning_progress AS
SELECT 
    pattern_type,
    generation_timestamp::DATE as observation_date,
    COUNT(*) as examples_today,
    AVG(forge_confidence_after) as daily_avg_confidence,
    MAX(forge_confidence_after) as peak_confidence,
    COUNT(*) FILTER (WHERE success = TRUE) as successes_today
FROM forge_confidence_tracker
WHERE generation_timestamp >= NOW() - INTERVAL '30 days'
GROUP BY pattern_type, generation_timestamp::DATE
ORDER BY observation_date DESC, pattern_type;

-- View: GitHub Copilot Quality Analysis
-- Which models/prompts produce the best code?
CREATE OR REPLACE VIEW github_copilot_quality_metrics AS
SELECT 
    github_model_used,
    pattern_category,
    COUNT(*) as total_generations,
    COUNT(*) FILTER (WHERE success = TRUE) as successful_generations,
    ROUND(100.0 * COUNT(*) FILTER (WHERE success = TRUE) / COUNT(*), 1) as success_rate_pct,
    AVG(execution_time_ms) as avg_execution_time_ms,
    COUNT(DISTINCT pattern_type) as unique_patterns_covered
FROM forge_confidence_tracker
WHERE github_model_used IS NOT NULL
GROUP BY github_model_used, pattern_category
ORDER BY success_rate_pct DESC;

-- Function: Update FORGE confidence after new observation
CREATE OR REPLACE FUNCTION update_forge_confidence(
    p_pattern_type VARCHAR(100),
    p_success BOOLEAN
) RETURNS DECIMAL(3,2) AS $$
DECLARE
    v_total_examples INT;
    v_successful_examples INT;
    v_new_confidence DECIMAL(3,2);
BEGIN
    -- Count examples for this pattern
    SELECT 
        COUNT(*),
        COUNT(*) FILTER (WHERE success = TRUE)
    INTO v_total_examples, v_successful_examples
    FROM forge_confidence_tracker
    WHERE pattern_type = p_pattern_type;
    
    -- Calculate new confidence (success rate with Bayesian smoothing)
    -- Prior: 50% confidence with 2 virtual examples
    -- Formula: (successful + 1) / (total + 2)
    v_new_confidence := ROUND(
        (v_successful_examples + 1.0) / (v_total_examples + 2.0)::DECIMAL,
        2
    );
    
    -- Cap at 0.95 until graduation review
    IF v_new_confidence > 0.95 THEN
        v_new_confidence := 0.95;
    END IF;
    
    RETURN v_new_confidence;
END;
$$ LANGUAGE plpgsql;

-- Function: Check if pattern is ready for graduation
CREATE OR REPLACE FUNCTION check_graduation_ready(
    p_pattern_type VARCHAR(100)
) RETURNS BOOLEAN AS $$
DECLARE
    v_confidence DECIMAL(3,2);
    v_example_count INT;
    v_success_rate DECIMAL(3,2);
BEGIN
    -- Count examples first
    SELECT COUNT(*) INTO v_example_count
    FROM forge_confidence_tracker
    WHERE pattern_type = p_pattern_type;
    
    -- Return FALSE if no examples yet
    IF v_example_count = 0 THEN
        RETURN FALSE;
    END IF;
    
    -- Get confidence and success rate
    SELECT 
        update_forge_confidence(p_pattern_type, TRUE), -- Get current confidence
        ROUND(100.0 * COUNT(*) FILTER (WHERE success = TRUE) / COUNT(*), 2)
    INTO v_confidence, v_success_rate
    FROM forge_confidence_tracker
    WHERE pattern_type = p_pattern_type;
    
    -- Graduation criteria:
    -- 1. At least 20 examples observed
    -- 2. Success rate >= 95%
    -- 3. Confidence >= 0.95
    RETURN (
        v_example_count >= 20 AND
        v_success_rate >= 95.0 AND
        v_confidence >= 0.95
    );
END;
$$ LANGUAGE plpgsql;

-- Trigger: Auto-update graduation readiness after INSERT/UPDATE
CREATE OR REPLACE FUNCTION trigger_update_graduation_status()
RETURNS TRIGGER AS $$
BEGIN
    NEW.graduation_ready := check_graduation_ready(NEW.pattern_type);
    NEW.updated_at := NOW();
    
    -- Update similar_examples_count
    SELECT COUNT(*) INTO NEW.similar_examples_count
    FROM forge_confidence_tracker
    WHERE pattern_type = NEW.pattern_type;
    
    -- Update forge_confidence_after if not set
    IF NEW.forge_confidence_after IS NULL THEN
        NEW.forge_confidence_after := update_forge_confidence(NEW.pattern_type, NEW.success);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER forge_tracker_auto_update
    BEFORE INSERT OR UPDATE ON forge_confidence_tracker
    FOR EACH ROW
    EXECUTE FUNCTION trigger_update_graduation_status();

-- Seed: Add initial pattern types FORGE will learn
INSERT INTO forge_confidence_tracker (
    pattern_type, pattern_category, task_description, 
    github_generated_code, success, forge_observation,
    similar_examples_count
) VALUES 
(
    'create_database_index',
    'database',
    'Example: Create index on solace_memory_crystals(criticality) for faster ORDER BY queries',
    'CREATE INDEX idx_crystals_criticality ON solace_memory_crystals(criticality);',
    TRUE,
    'Pattern: CREATE INDEX idx_{table}_{column} ON {table}({column}); - Always name indexes descriptively.',
    1
),
(
    'create_rest_api_endpoint',
    'backend',
    'Example: Create GET /api/v1/forge/confidence endpoint that returns graduation dashboard',
    'router.GET("/api/v1/forge/confidence", handlers.GetForgeConfidence)',
    TRUE,
    'Pattern: router.{METHOD}("{path}", handlers.{HandlerFunc}) - RESTful naming, handlers in separate package.',
    1
),
(
    'write_sql_migration',
    'database',
    'Example: Create migration file with proper numbering and rollback',
    'CREATE TABLE example (...); -- Rollback: DROP TABLE example;',
    TRUE,
    'Pattern: Migrations always include CREATE + comment with DROP for rollback. Number sequentially (001_, 002_).',
    1
)
ON CONFLICT DO NOTHING;

-- Grant permissions (if using role-based access)
-- GRANT SELECT, INSERT, UPDATE ON forge_confidence_tracker TO solace_user;
-- GRANT SELECT ON forge_graduation_dashboard TO solace_user;
-- GRANT SELECT ON forge_learning_progress TO solace_user;

COMMENT ON TABLE forge_confidence_tracker IS 'Tracks FORGEs apprenticeship learning from GitHub Copilot code generation. Crystal #27.1';
COMMENT ON COLUMN forge_confidence_tracker.pattern_type IS 'Classification of the coding pattern (e.g., add_index, create_endpoint)';
COMMENT ON COLUMN forge_confidence_tracker.graduation_ready IS 'TRUE when >= 95% confidence + >= 20 successful examples';
COMMENT ON VIEW forge_graduation_dashboard IS 'Shows which patterns FORGE is ready to take over from GitHub Copilot';
COMMENT ON FUNCTION update_forge_confidence IS 'Calculates FORGEs confidence using Bayesian success rate (prevents overconfidence)';
COMMENT ON FUNCTION check_graduation_ready IS 'Returns TRUE if pattern meets graduation criteria (20 examples, 95% success)';
