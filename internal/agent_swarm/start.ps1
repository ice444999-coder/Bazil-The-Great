# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# ====================================================================
# 🚀 ARES Coordinator Startup Script
# ====================================================================
# This script validates the environment and starts the WebSocket server.
# ====================================================================

Write-Host ""
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "🚀 Starting ARES Coordinator" -ForegroundColor Cyan
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Step 1: Activate virtual environment
Write-Host "📦 Activating virtual environment..." -ForegroundColor Yellow
$venvPath = Join-Path $ScriptDir "venv\Scripts\Activate.ps1"
if (Test-Path $venvPath) {
    & $venvPath
    Write-Host "   ✅ Virtual environment activated" -ForegroundColor Green
} else {
    Write-Host "   ❌ Virtual environment not found!" -ForegroundColor Red
    Write-Host "   Path checked: $venvPath" -ForegroundColor Gray
    Write-Host "   Run .\setup.ps1 first to create the virtual environment" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

# Step 2: Validate environment variables
Write-Host "🔍 Validating environment configuration..." -ForegroundColor Yellow
$pythonExe = Join-Path $ScriptDir "venv\Scripts\python.exe"
& $pythonExe validate_env.py

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "❌ Environment validation failed!" -ForegroundColor Red
    Write-Host "   Please fix the issues above and try again." -ForegroundColor Yellow
    Write-Host ""
    exit 1
}
Write-Host ""

# Step 3: Start the WebSocket server
Write-Host "🌐 Starting WebSocket server..." -ForegroundColor Yellow
Write-Host "   Server will start on ws://localhost:8765" -ForegroundColor Gray
Write-Host "   Press Ctrl+C to stop" -ForegroundColor Gray
Write-Host ""
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""

& $pythonExe coordinator.py --websocket
