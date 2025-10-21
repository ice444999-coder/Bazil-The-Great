# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# ARES Agent Coordinator - Windows Service Script
# Run as: .\start-coordinator.ps1

$ErrorActionPreference = "Stop"

Write-Host "🤖 Starting ARES Agent Swarm Coordinator..." -ForegroundColor Cyan

# Check if Python is installed
if (-not (Get-Command python -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Python not found. Please install Python 3.9+" -ForegroundColor Red
    exit 1
}

# Check Python version
$pythonVersion = python --version 2>&1
Write-Host "✅ Python: $pythonVersion" -ForegroundColor Green

# Install dependencies if needed
$requirementsPath = Join-Path $PSScriptRoot "requirements.txt"
if (Test-Path $requirementsPath) {
    Write-Host "📦 Installing dependencies..." -ForegroundColor Yellow
    python -m pip install -r $requirementsPath --quiet
    Write-Host "✅ Dependencies installed" -ForegroundColor Green
}

# Set environment variables
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_NAME = "ares_db"
$env:DB_USER = "ARES"
$env:DB_PASSWORD = "ARESISWAKING"

# Check for API keys
if (-not $env:OPENAI_API_KEY) {
    Write-Host "⚠️  OPENAI_API_KEY not set - SOLACE will be unavailable" -ForegroundColor Yellow
}
if (-not $env:ANTHROPIC_API_KEY) {
    Write-Host "⚠️  ANTHROPIC_API_KEY not set - FORGE will be unavailable" -ForegroundColor Yellow
}

# Start coordinator
$coordinatorPath = Join-Path $PSScriptRoot "coordinator.py"
Write-Host "🚀 Launching coordinator at $coordinatorPath" -ForegroundColor Cyan
Write-Host "   Check interval: 10 seconds" -ForegroundColor Gray
Write-Host "   Log file: agent_coordinator.log" -ForegroundColor Gray
Write-Host "" -ForegroundColor Gray
Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
Write-Host ""

python $coordinatorPath --interval 10
