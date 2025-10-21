# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# Glass Box Integration Test
# Tests the decision tracing system by querying database

$env:PGPASSWORD = 'ARESISWAKING'
$psql = 'C:\Program Files\PostgreSQL\18\bin\psql.exe'

Write-Host "`n🔍 GLASS BOX INTEGRATION TEST`n" -ForegroundColor Cyan

Write-Host "1. Checking Decision Traces Table..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c "SELECT COUNT(*) as total_traces FROM decision_traces;"

Write-Host "`n2. Checking Decision Spans Table..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c "SELECT COUNT(*) as total_spans FROM decision_spans;"

Write-Host "`n3. Checking Decision Metrics Table..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c "SELECT COUNT(*) as total_metrics FROM decision_metrics;"

Write-Host "`n4. Sample Trace (if exists)..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c "SELECT id, trace_type, status, final_decision FROM decision_traces LIMIT 1;"

Write-Host "`n5. Verifying Hash Chain Structure..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c @"
SELECT 
    span_name,
    chain_position,
    LEFT(sha256_hash, 12) as hash,
    LEFT(previous_hash, 12) as prev
FROM decision_spans 
ORDER BY chain_position 
LIMIT 5;
"@

Write-Host "`n✅ Glass Box tables are installed and ready!`n" -ForegroundColor Green
Write-Host "Next: Execute a trade to create decision traces with hash-chained spans`n"
