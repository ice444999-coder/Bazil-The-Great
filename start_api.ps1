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
