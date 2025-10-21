# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# Run ARES Litmus Test
# Quick launcher for the enhanced litmus test

Write-Host "🧪 Starting ARES Enhanced Litmus Test..." -ForegroundColor Cyan
Write-Host ""

# Check Python
if (!(Get-Command python -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Python not found. Install Python 3.x first." -ForegroundColor Red
    exit 1
}

# Check requests module
python -c "import requests" 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️ Installing requests module..." -ForegroundColor Yellow
    python -m pip install requests
}

# Run test
Set-Location C:\ARES_Workspace\ARES_API
python litmus_test_enhanced.py

Write-Host ""
Write-Host "✅ Test complete. Check output above." -ForegroundColor Green
