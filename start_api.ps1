# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# ARES API Startup Script
# This script cleanly starts the ARES API without any formatting issues

$ErrorActionPreference = "Stop"

Write-Host "Starting ARES API..." -NoNewline

# Check if already running
$existing = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue
if ($existing) {
    Write-Host " (killing existing process)" -ForegroundColor Yellow
    $procId = $existing.OwningProcess
    Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
    Start-Sleep -Milliseconds 500
}

Write-Host ""
Set-Location "c:\ARES_Workspace\ARES_API"
go run cmd/main.go
