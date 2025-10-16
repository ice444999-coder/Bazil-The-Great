# ARES Agent Swarm - Complete Setup and Test Script
# Executes all Phase 3 steps in order

Write-Host "================================" -ForegroundColor Cyan
Write-Host "ARES AGENT SWARM - PHASE 3 SETUP" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

$ErrorActionPreference = "Stop"

# Navigate to ARES_API directory
cd C:\ARES_Workspace\ARES_API

# STEP 1: Verify .env configuration
Write-Host "`n[STEP 1] Verifying .env configuration..." -ForegroundColor Yellow

if (-not (Test-Path ".env")) {
    Write-Host "  ❌ ERROR: .env file not found!" -ForegroundColor Red
    exit 1
}

$envContent = Get-Content ".env" -Raw
$requiredKeys = @("OPENAI_API_KEY", "CLAUDE_API_KEY", "DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME")
$missingKeys = @()

foreach ($key in $requiredKeys) {
    if ($envContent -notmatch "$key=.+") {
        $missingKeys += $key
    } else {
        Write-Host "  ✅ $key found" -ForegroundColor Green
    }
}

if ($missingKeys.Count -gt 0) {
    Write-Host "  ❌ Missing keys: $($missingKeys -join ', ')" -ForegroundColor Red
    exit 1
}

# STEP 2: Check Python and dependencies
Write-Host "`n[STEP 2] Checking Python environment..." -ForegroundColor Yellow

try {
    $pythonVersion = python --version 2>&1
    Write-Host "  ✅ Python: $pythonVersion" -ForegroundColor Green
} catch {
    Write-Host "  ❌ Python not found" -ForegroundColor Red
    exit 1
}

Write-Host "  Installing Python dependencies..." -ForegroundColor Cyan
pip install -q -r internal/agent_swarm/requirements.txt
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✅ Dependencies installed" -ForegroundColor Green
} else {
    Write-Host "  ❌ Failed to install dependencies" -ForegroundColor Red
    exit 1
}

# STEP 3: Check Ollama
Write-Host "`n[STEP 3] Checking Ollama (DeepSeek)..." -ForegroundColor Yellow

try {
    $ollamaTest = Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  ✅ Ollama running" -ForegroundColor Green
    
    # Check for DeepSeek model
    if ($ollamaTest.models | Where-Object { $_.name -like "*deepseek*" }) {
        Write-Host "  ✅ DeepSeek model found" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  DeepSeek model not found" -ForegroundColor Yellow
        Write-Host "     Run: ollama pull deepseek-r1:14b" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  ❌ Ollama not running" -ForegroundColor Red
    Write-Host "     Start Ollama: ollama serve" -ForegroundColor Yellow
    exit 1
}

# STEP 4: Install Playwright for SENTINEL
Write-Host "`n[STEP 4] Setting up Playwright (for SENTINEL UI testing)..." -ForegroundColor Yellow

try {
    playwright install chromium 2>&1 | Out-Null
    Write-Host "  ✅ Playwright Chromium installed" -ForegroundColor Green
} catch {
    Write-Host "  ⚠️  Playwright install skipped (may already be installed)" -ForegroundColor Yellow
}

# STEP 5: Check ARES backend
Write-Host "`n[STEP 5] Checking ARES backend API..." -ForegroundColor Yellow

try {
    $apiTest = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/agents" -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  ✅ ARES API running (found $($apiTest.Count) agents)" -ForegroundColor Green
} catch {
    Write-Host "  ⚠️  ARES API not running" -ForegroundColor Yellow
    Write-Host "     Start it: .\ares_api.exe" -ForegroundColor Cyan
    $continue = Read-Host "Continue without ARES API? (y/n)"
    if ($continue -ne "y") {
        exit 1
    }
}

# STEP 6: Run connection tests
Write-Host "`n[STEP 6] Testing all API connections..." -ForegroundColor Yellow

python internal/agent_swarm/test_connections.py
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✅ All connection tests passed!" -ForegroundColor Green
} else {
    Write-Host "  ❌ Some connection tests failed" -ForegroundColor Red
    Write-Host "     Check errors above and fix before continuing" -ForegroundColor Yellow
    exit 1
}

# STEP 7: Summary
Write-Host "`n================================" -ForegroundColor Cyan
Write-Host "SETUP COMPLETE - READY TO START" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

Write-Host "`n📋 Next Steps:" -ForegroundColor White
Write-Host "  1. Start coordinator:" -ForegroundColor Cyan
Write-Host "     .\internal\agent_swarm\start-coordinator.ps1" -ForegroundColor White
Write-Host ""
Write-Host "  2. View dashboard:" -ForegroundColor Cyan
Write-Host "     http://localhost:8080/web/agent-dashboard.html" -ForegroundColor White
Write-Host ""
Write-Host "  3. Run end-to-end test:" -ForegroundColor Cyan
Write-Host "     python internal/agent_swarm/test_full_workflow.py" -ForegroundColor White
Write-Host ""
Write-Host "  4. Create UI fix task:" -ForegroundColor Cyan
Write-Host "     .\internal\agent_swarm\create_ui_fix_task.ps1" -ForegroundColor White
Write-Host ""

$startNow = Read-Host "Start coordinator now? (y/n)"
if ($startNow -eq "y") {
    Write-Host "`n🚀 Starting agent coordinator..." -ForegroundColor Green
    .\internal\agent_swarm\start-coordinator.ps1
}
