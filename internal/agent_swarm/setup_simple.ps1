# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# Simple ARES Coordinator Setup Script
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

Write-Host "🚀 ARES Coordinator Setup" -ForegroundColor Cyan
Write-Host ""

# Remove old venv
Remove-Item -Recurse -Force venv -ErrorAction SilentlyContinue

# Create virtual environment
Write-Host "Creating virtual environment..."
C:\Python313\python.exe -m venv venv

# Activate and install
Write-Host "Installing dependencies..."
& .\venv\Scripts\Activate.ps1
pip install --upgrade pip --quiet
pip install -r requirements.txt --quiet

Write-Host ""
Write-Host "✅ Setup complete!" -ForegroundColor Green
Write-Host "Next steps:"
Write-Host "1. Copy .env.example to .env and add your API keys"
Write-Host "2. Run: powershell -File start_simple.ps1"
