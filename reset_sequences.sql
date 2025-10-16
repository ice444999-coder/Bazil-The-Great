-- Reset decision_traces sequence to allow clean test runs
-- Run with: psql -U ARES -d ares_db -f reset_sequences.sql

-- Delete test data (keep the sample trace)
DELETE FROM decision_spans WHERE trace_id > 1;
DELETE FROM decision_metrics WHERE trace_id > 1;
DELETE FROM decision_traces WHERE id > 1;

-- Reset sequence to start from 2
SELECT setval('decision_traces_id_seq', 1, true);

-- Verify
SELECT 'Sequence reset to: ' || last_value as status FROM decision_traces_id_seq;
SELECT 'Remaining traces: ' || COUNT(*) as status FROM decision_traces;
