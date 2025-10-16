-- Approach 2: SQL File Method
-- Run with: psql -U ARES -d ares_db -f check_schema.sql

-- Check decision_traces structure
\d decision_traces

-- Check if it has SERIAL/IDENTITY for auto-increment
SELECT column_name, column_default, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'decision_traces' 
AND column_name = 'id';

-- Check current sequence value
SELECT last_value FROM decision_traces_id_seq;

-- Check all traces
SELECT id, trace_type, status, start_time 
FROM decision_traces 
ORDER BY id;
