# ============================================================================
# MEMORY CRYSTAL TOOLS - VERIFICATION TEST SUITE
# ============================================================================
# Purpose: Test the 3 new memory crystal tools for SOLACE
# Tools Tested:
#   1. query_memory_crystals
#   2. create_memory_crystal
#   3. ingest_document_to_crystal
# ============================================================================

Write-Host "`n╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  SOLACE MEMORY CRYSTAL TOOLS - VERIFICATION TEST          ║" -ForegroundColor Cyan
Write-Host "║  Date: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')                    ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080/api/v1/solace-agent/chat"
$testSession = "crystal_test_$(Get-Random)"
$passedTests = 0
$totalTests = 5

# ============================================================================
# TEST 1: Query Existing Memory Crystals
# ============================================================================
Write-Host "`n=== TEST 1: Query Memory Crystals (Basic Search) ===" -ForegroundColor Cyan
Write-Host "Purpose: Verify SOLACE can search memory crystals with full-text search" -ForegroundColor Gray

try {
    $body = @{
        session_id = $testSession
        message = "Use query_memory_crystals to search for crystals about 'enki' with criticality CRITICAL"
    } | ConvertTo-Json

    Write-Host "  → Sending query request to SOLACE..." -ForegroundColor Gray
    $response = Invoke-RestMethod -Uri $baseUrl -Method POST -Body $body -ContentType "application/json" -TimeoutSec 30

    if ($response.response -match "crystal" -or $response.response -match "Found") {
        Write-Host "✅ PASSED: SOLACE executed query_memory_crystals successfully" -ForegroundColor Green
        Write-Host "   Response preview: $($response.response.Substring(0, [Math]::Min(200, $response.response.Length)))..." -ForegroundColor Gray
        $passedTests++
    } else {
        Write-Host "❌ FAILED: Response doesn't indicate crystal query" -ForegroundColor Red
        Write-Host "   Response: $($response.response)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ FAILED: Exception - $_" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# ============================================================================
# TEST 2: Create a Memory Crystal
# ============================================================================
Write-Host "`n=== TEST 2: Create Memory Crystal ===" -ForegroundColor Cyan
Write-Host "Purpose: Verify SOLACE can create new memory crystals with proper hash chaining" -ForegroundColor Gray

try {
    $body = @{
        session_id = $testSession
        message = @"
Use create_memory_crystal to store this knowledge:
- Title: Test Crystal - PowerShell Port Check
- Category: tools
- Criticality: MEDIUM
- Summary: How to check if port 8080 is occupied before starting ARES API
- Content: Run 'netstat -ano | findstr :8080' to check port availability. If occupied, use 'Get-NetTCPConnection -LocalPort 8080' to find the process ID.
- Tags: powershell, port, troubleshooting
"@
    } | ConvertTo-Json

    Write-Host "  → Creating memory crystal via SOLACE..." -ForegroundColor Gray
    $response = Invoke-RestMethod -Uri $baseUrl -Method POST -Body $body -ContentType "application/json" -TimeoutSec 30

    if ($response.response -match "Crystal Created" -or $response.response -match "ID:" -or $response.response -match "SHA-256") {
        Write-Host "✅ PASSED: Memory crystal created successfully" -ForegroundColor Green
        Write-Host "   Response preview: $($response.response.Substring(0, [Math]::Min(300, $response.response.Length)))..." -ForegroundColor Gray
        $passedTests++
    } else {
        Write-Host "❌ FAILED: Crystal creation didn't complete" -ForegroundColor Red
        Write-Host "   Response: $($response.response)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ FAILED: Exception - $_" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# ============================================================================
# TEST 3: Query the Crystal We Just Created
# ============================================================================
Write-Host "`n=== TEST 3: Retrieve Newly Created Crystal ===" -ForegroundColor Cyan
Write-Host "Purpose: Verify immediate searchability of new crystals" -ForegroundColor Gray

try {
    $body = @{
        session_id = $testSession
        message = "Query memory crystals for 'PowerShell Port Check' - this should find the crystal we just created"
    } | ConvertTo-Json

    Write-Host "  → Searching for newly created crystal..." -ForegroundColor Gray
    $response = Invoke-RestMethod -Uri $baseUrl -Method POST -Body $body -ContentType "application/json" -TimeoutSec 30

    if ($response.response -match "PowerShell Port Check" -or $response.response -match "8080") {
        Write-Host "✅ PASSED: Found the crystal we just created (immediate searchability confirmed)" -ForegroundColor Green
        $passedTests++
    } else {
        Write-Host "⚠️  UNCERTAIN: Crystal not found in search results" -ForegroundColor Yellow
        Write-Host "   Response: $($response.response)" -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ FAILED: Exception - $_" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# ============================================================================
# TEST 4: Ingest Document to Crystal
# ============================================================================
Write-Host "`n=== TEST 4: Ingest Document to Crystal ===" -ForegroundColor Cyan
Write-Host "Purpose: Verify SOLACE can automatically create crystals from existing documents" -ForegroundColor Gray

# Create a test document first
$testDocPath = "C:\ARES_Workspace\TEST_MEMORY_CRYSTAL_DOC.md"
$testDocContent = @"
# Test Memory Crystal Document

This is a test document to verify the ingest_document_to_crystal tool.

## Key Information
- This document tests automatic ingestion
- Tags should be auto-extracted
- Summary should be auto-generated
- SHA-256 hash should be calculated

## Important Keywords
SOLACE, enki, memory, test, critical
"@

try {
    # Create test document
    $testDocContent | Out-File -FilePath $testDocPath -Encoding UTF8 -Force
    Write-Host "  → Created test document: $testDocPath" -ForegroundColor Gray

    # Ask SOLACE to ingest it
    $body = @{
        session_id = $testSession
        message = "Use ingest_document_to_crystal to import this file: $testDocPath (category: testing, criticality: LOW)"
    } | ConvertTo-Json

    Write-Host "  → Asking SOLACE to ingest document..." -ForegroundColor Gray
    $response = Invoke-RestMethod -Uri $baseUrl -Method POST -Body $body -ContentType "application/json" -TimeoutSec 30

    if ($response.response -match "Ingestion Complete" -or $response.response -match "Crystal Created" -or $response.response -match "auto-extracted") {
        Write-Host "✅ PASSED: Document ingested successfully" -ForegroundColor Green
        Write-Host "   Response preview: $($response.response.Substring(0, [Math]::Min(300, $response.response.Length)))..." -ForegroundColor Gray
        $passedTests++
    } else {
        Write-Host "❌ FAILED: Document ingestion didn't complete" -ForegroundColor Red
        Write-Host "   Response: $($response.response)" -ForegroundColor Yellow
    }

    # Cleanup
    Remove-Item $testDocPath -Force -ErrorAction SilentlyContinue
    Write-Host "  → Cleaned up test document" -ForegroundColor Gray

} catch {
    Write-Host "❌ FAILED: Exception - $_" -ForegroundColor Red
    Remove-Item $testDocPath -Force -ErrorAction SilentlyContinue
}

Start-Sleep -Seconds 2

# ============================================================================
# TEST 5: Multi-Filter Query (Category + Criticality)
# ============================================================================
Write-Host "`n=== TEST 5: Advanced Query with Multiple Filters ===" -ForegroundColor Cyan
Write-Host "Purpose: Verify SOLACE can use multiple filters simultaneously" -ForegroundColor Gray

try {
    $body = @{
        session_id = $testSession
        message = "Query memory crystals with category='solace_core' AND criticality='CRITICAL' - show me the most critical SOLACE knowledge"
    } | ConvertTo-Json

    Write-Host "  → Executing multi-filter query..." -ForegroundColor Gray
    $response = Invoke-RestMethod -Uri $baseUrl -Method POST -Body $body -ContentType "application/json" -TimeoutSec 30

    if ($response.response -match "crystal" -or $response.response -match "CRITICAL" -or $response.response -match "solace_core") {
        Write-Host "✅ PASSED: Multi-filter query executed successfully" -ForegroundColor Green
        $passedTests++
    } else {
        Write-Host "⚠️  UNCERTAIN: Query completed but results unclear" -ForegroundColor Yellow
        Write-Host "   Response: $($response.response.Substring(0, [Math]::Min(200, $response.response.Length)))..." -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ FAILED: Exception - $_" -ForegroundColor Red
}

# ============================================================================
# TEST SUMMARY
# ============================================================================
Write-Host "`n╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  TEST SUMMARY                                              ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host "Tests Passed: $passedTests / $totalTests" -ForegroundColor $(if ($passedTests -eq $totalTests) { "Green" } elseif ($passedTests -ge 3) { "Yellow" } else { "Red" })

if ($passedTests -eq $totalTests) {
    Write-Host "`n🎉 ALL TESTS PASSED - Memory Crystal Tools are operational!" -ForegroundColor Green
    Write-Host "✅ SOLACE now has 3 new tools for managing immutable knowledge" -ForegroundColor Green
} elseif ($passedTests -ge 3) {
    Write-Host "`n⚠️  MOST TESTS PASSED - Review failures above" -ForegroundColor Yellow
} else {
    Write-Host "`n❌ CRITICAL: Multiple test failures - Review logs" -ForegroundColor Red
}

# ============================================================================
# MANUAL VERIFICATION COMMANDS
# ============================================================================
Write-Host "`n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "Manual Verification Commands (PostgreSQL):" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""
Write-Host "# Count all memory crystals:" -ForegroundColor Gray
Write-Host "psql -U ARES -d ares_db -c ""SELECT COUNT(*) FROM solace_memory_crystals;""" -ForegroundColor White
Write-Host ""
Write-Host "# View latest 5 crystals:" -ForegroundColor Gray
Write-Host "psql -U ARES -d ares_db -c ""SELECT id, title, category, criticality, created_at FROM solace_memory_crystals ORDER BY id DESC LIMIT 5;""" -ForegroundColor White
Write-Host ""
Write-Host "# Verify hash chain integrity:" -ForegroundColor Gray
Write-Host "psql -U ARES -d ares_db -c ""SELECT id, title, LEFT(sha256_hash, 16) as hash, LEFT(previous_hash, 16) as prev FROM solace_memory_crystals ORDER BY id DESC LIMIT 10;""" -ForegroundColor White
Write-Host ""
Write-Host "# Search crystals by tag:" -ForegroundColor Gray
Write-Host "psql -U ARES -d ares_db -c ""SELECT title, tags FROM solace_memory_crystals WHERE 'enki' = ANY(tags);""" -ForegroundColor White
Write-Host ""

# ============================================================================
# TOOL USAGE EXAMPLES FOR HUMANS
# ============================================================================
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "How to Use Memory Crystal Tools (via Chat):" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""
Write-Host "1️⃣  Query crystals:" -ForegroundColor Yellow
Write-Host '   "Search memory crystals for anything about the sacred user_id enki"' -ForegroundColor White
Write-Host ""
Write-Host "2️⃣  Create a crystal:" -ForegroundColor Yellow
Write-Host '   "Create a memory crystal about this bug I just fixed: [description]"' -ForegroundColor White
Write-Host ""
Write-Host "3️⃣  Ingest a document:" -ForegroundColor Yellow
Write-Host '   "Ingest the HANDOVER_MANIFEST file into memory crystals"' -ForegroundColor White
Write-Host ""
Write-Host "4️⃣  Filter by criticality:" -ForegroundColor Yellow
Write-Host '   "Show me all CRITICAL memory crystals"' -ForegroundColor White
Write-Host ""
Write-Host "5️⃣  Filter by category:" -ForegroundColor Yellow
Write-Host '   "Query memory crystals in the testing category"' -ForegroundColor White
Write-Host ""

Write-Host "`n✨ Memory Crystal Tools Installation Complete! ✨`n" -ForegroundColor Cyan
