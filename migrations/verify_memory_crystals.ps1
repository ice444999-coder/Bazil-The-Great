# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# SOLACE Memory Crystal System - Verification Script
# Created: October 17, 2025
# Purpose: Verify the memory crystal migration was successful

Write-Host "`n🔮 SOLACE Memory Crystal System - Verification`n" -ForegroundColor Cyan

# Set PostgreSQL credentials
$env:PGPASSWORD = "ARESISWAKING"
$psql = "C:\Program Files\PostgreSQL\18\bin\psql.exe"

Write-Host "1️⃣ Checking table existence..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c "\dt solace_memory_crystals" -t | Out-String

Write-Host "`n2️⃣ Counting crystals..." -ForegroundColor Yellow
$count = & $psql -U ARES -d ares_db -c "SELECT COUNT(*) FROM solace_memory_crystals;" -t
Write-Host "   Total Crystals: $($count.Trim())" -ForegroundColor Green

Write-Host "`n3️⃣ Checking criticality distribution..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c "SELECT criticality, COUNT(*) FROM solace_memory_crystals GROUP BY criticality ORDER BY CASE criticality WHEN 'CRITICAL' THEN 1 WHEN 'HIGH' THEN 2 WHEN 'MEDIUM' THEN 3 ELSE 4 END;"

Write-Host "`n4️⃣ Verifying hash chain integrity..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c "SELECT COUNT(*) as valid_chains FROM crystal_chain_integrity WHERE chain_status = 'VALID' OR chain_status = 'GENESIS';" -t

Write-Host "`n5️⃣ Listing all crystals..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c "SELECT id, LEFT(title, 50) as title, category, criticality FROM solace_memory_crystals ORDER BY id;"

Write-Host "`n6️⃣ Testing query function..." -ForegroundColor Yellow
Write-Host "   Querying for 'async' keyword..." -ForegroundColor Gray
& $psql -U ARES -d ares_db -c "SELECT id, LEFT(title, 60) as title FROM query_crystals(NULL, 'async', 'CRITICAL', 5);"

Write-Host "`n7️⃣ Testing critical crystals view..." -ForegroundColor Yellow
& $psql -U ARES -d ares_db -c "SELECT COUNT(*) FROM critical_crystals;" -t

Write-Host "`n✅ Memory Crystal System Verification Complete!`n" -ForegroundColor Green

Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "  1. Add query_memory_crystals() tool to SOLACE (solace.go)" -ForegroundColor White
Write-Host "  2. Add create_memory_crystal() tool for auto-generation" -ForegroundColor White
Write-Host "  3. Test from SOLACE: 'Query memory crystals for async logging'" -ForegroundColor White
Write-Host ""
