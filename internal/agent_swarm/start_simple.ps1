# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# Simple ARES Coordinator Start Script
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

Write-Host "🚀 Starting ARES Coordinator" -ForegroundColor Cyan
Write-Host ""

# Activate venv
& .\venv\Scripts\Activate.ps1

# Validate environment
python validate_env.py
if ($LASTEXITCODE -ne 0) { exit 1 }

# Start server
Write-Host ""
Write-Host "Starting WebSocket server..." -ForegroundColor Green
python coordinator.py --websocket
