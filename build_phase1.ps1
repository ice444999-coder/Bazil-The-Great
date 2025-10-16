# Phase 1 Build and Verification Script
# Builds ARES_API with modular architecture improvements

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "PHASE 1: MODULAR ARCHITECTURE BUILD" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Change to ARES_API directory
Set-Location C:\ARES_Workspace\ARES_API

# Verify go.mod exists
if (-not (Test-Path go.mod)) {
    Write-Host "❌ ERROR: go.mod not found" -ForegroundColor Red
    exit 1
}

Write-Host "📁 Working Directory: $(Get-Location)" -ForegroundColor Gray
Write-Host "📄 go.mod exists: ✅" -ForegroundColor Green

# Clean previous build
if (Test-Path ares-api.exe) {
    Remove-Item ares-api.exe -Force
    Write-Host "🧹 Removed old executable" -ForegroundColor Yellow
}

# Tidy modules
Write-Host "`n📦 Cleaning Go modules..." -ForegroundColor Cyan
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ go mod tidy failed" -ForegroundColor Red
    exit 1
}

# Build with timeout
Write-Host "`n🔨 Building ARES_API..." -ForegroundColor Cyan
$buildJob = Start-Job -ScriptBlock {
    Set-Location C:\ARES_Workspace\ARES_API
    go build -v -o ares-api.exe .\cmd\main.go 2>&1
}

# Wait up to 60 seconds
$completed = Wait-Job $buildJob -Timeout 60
if ($completed) {
    $output = Receive-Job $buildJob
    $exitCode = $buildJob.State -eq 'Completed' ? 0 : 1
    
    if (Test-Path ares-api.exe) {
        Write-Host "✅ BUILD SUCCESS" -ForegroundColor Green
        $exe = Get-Item ares-api.exe
        Write-Host "   File: $($exe.Name)" -ForegroundColor Gray
        Write-Host "   Size: $([math]::Round($exe.Length/1MB, 2)) MB" -ForegroundColor Gray
        Write-Host "   Time: $($exe.LastWriteTime)" -ForegroundColor Gray
    } else {
        Write-Host "❌ BUILD FAILED - no executable created" -ForegroundColor Red
        Write-Host "`nBuild Output:" -ForegroundColor Yellow
        Write-Host $output
        exit 1
    }
} else {
    Write-Host "❌ BUILD TIMEOUT (>60s)" -ForegroundColor Red
    Stop-Job $buildJob
    Remove-Job $buildJob
    exit 1
}

Remove-Job $buildJob

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "VERIFICATION TESTS" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Test 1: Check migrations file exists
Write-Host "TEST 1: Service Registry Migration" -ForegroundColor Yellow
if (Test-Path migrations\004_service_registry.sql) {
    Write-Host "  ✅ migrations/004_service_registry.sql exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ Migration file missing" -ForegroundColor Red
}

# Test 2: Check registry package exists  
Write-Host "`nTEST 2: Registry Package" -ForegroundColor Yellow
if (Test-Path internal\registry\service_registry.go) {
    Write-Host "  ✅ internal/registry/service_registry.go exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ Registry package missing" -ForegroundColor Red
}

# Test 3: Check health controller exists
Write-Host "`nTEST 3: Health Controller" -ForegroundColor Yellow
if (Test-Path internal\api\controllers\health_controller.go) {
    Write-Host "  ✅ internal/api/controllers/health_controller.go exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ Health controller missing" -ForegroundColor Red
}

# Test 4: Check CONTRACTS.md exists
Write-Host "`nTEST 4: Service Contracts Documentation" -ForegroundColor Yellow
if (Test-Path CONTRACTS.md) {
    $contracts = Get-Content CONTRACTS.md
    Write-Host "  ✅ CONTRACTS.md exists ($($contracts.Count) lines)" -ForegroundColor Green
} else {
    Write-Host "  ❌ CONTRACTS.md missing" -ForegroundColor Red
}

# Test 5: Verify database migration applied
Write-Host "`nTEST 5: Database Service Registry Table" -ForegroundColor Yellow
$env:PGPASSWORD = 'ARESISWAKING'
$tableCheck = & 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -U ARES -d ares_db -t -A -c "SELECT COUNT(*) FROM service_registry;" 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✅ service_registry table exists ($tableCheck rows)" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  service_registry table not verified (run migration)" -ForegroundColor Yellow
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "PHASE 1 BUILD COMPLETE" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "  1. Start ARES_API: .\ares-api.exe" -ForegroundColor White
Write-Host "  2. Test health endpoint: curl http://localhost:8080/health" -ForegroundColor White
Write-Host "  3. Check service registry: curl http://localhost:8080/health/services" -ForegroundColor White
Write-Host "  4. Verify heartbeat after 30s" -ForegroundColor White
