# ARES Agent Coordinator - Universal Python Starter
# Finds and uses ANY working Python installation

$ErrorActionPreference = "Stop"

Write-Host "🤖 Starting ARES Agent Swarm Coordinator..." -ForegroundColor Cyan

# Try to find working Python
$pythonCommands = @("python", "python3", "py")
$workingPython = $null

foreach ($cmd in $pythonCommands) {
    if (Get-Command $cmd -ErrorAction SilentlyContinue) {
        try {
            $version = & $cmd --version 2>&1
            if ($LASTEXITCODE -eq 0 -or $version -match "Python") {
                Write-Host "✅ Found Python: $cmd ($version)" -ForegroundColor Green
                $workingPython = $cmd
                break
            }
        } catch {
            continue
        }
    }
}

# If still not found, try direct paths
if (-not $workingPython) {
    Write-Host "⚠️  Python not in PATH, searching common locations..." -ForegroundColor Yellow
    
    $pythonPaths = @(
        "C:\Users\$env:USERNAME\AppData\Local\Programs\Python\Python312\python.exe",
        "C:\Users\$env:USERNAME\AppData\Local\Programs\Python\Python311\python.exe",
        "C:\Program Files\Python312\python.exe",
        "C:\Program Files\Python311\python.exe",
        "C:\Python312\python.exe",
        "C:\Python311\python.exe"
    )
    
    foreach ($path in $pythonPaths) {
        if (Test-Path $path) {
            try {
                $version = & $path --version 2>&1
                Write-Host "✅ Found Python: $path ($version)" -ForegroundColor Green
                $workingPython = $path
                break
            } catch {
                continue
            }
        }
    }
}

if (-not $workingPython) {
    Write-Host "❌ No Python installation found!" -ForegroundColor Red
    Write-Host "   Please install Python 3.10+ from https://www.python.org/downloads/" -ForegroundColor Yellow
    Write-Host "   Make sure to check 'Add Python to PATH' during installation" -ForegroundColor Yellow
    exit 1
}

# Install dependencies
$requirementsPath = "C:\ARES_Workspace\ARES_API\internal\agent_swarm\requirements.txt"
if (Test-Path $requirementsPath) {
    Write-Host "`n📦 Installing dependencies..." -ForegroundColor Yellow
    try {
        & $workingPython -m pip install --quiet --upgrade pip 2>&1 | Out-Null
        & $workingPython -m pip install --quiet -r $requirementsPath 2>&1 | Out-Null
        Write-Host "✅ Dependencies installed" -ForegroundColor Green
    } catch {
        Write-Host "⚠️  Some dependencies may have failed, trying to continue..." -ForegroundColor Yellow
    }
}

# Check if .env exists and load it
$envPath = "C:\ARES_Workspace\ARES_API\.env"
if (Test-Path $envPath) {
    Write-Host "✅ Found .env file" -ForegroundColor Green
} else {
    Write-Host "⚠️  No .env file found - environment variables must be set manually" -ForegroundColor Yellow
}

# Navigate to ARES_API directory
Set-Location C:\ARES_Workspace\ARES_API

# Display info
Write-Host "`n🚀 Launching Agent Swarm Coordinator..." -ForegroundColor Cyan
Write-Host "   Python: $workingPython" -ForegroundColor Gray
Write-Host "   Interval: 10 seconds" -ForegroundColor Gray
Write-Host "   Log file: agent_coordinator.log" -ForegroundColor Gray
Write-Host "   Dashboard: http://localhost:8080/web/agent-dashboard.html" -ForegroundColor Yellow
Write-Host "`nPress Ctrl+C to stop`n" -ForegroundColor Yellow

# Start coordinator
try {
    & $workingPython internal\agent_swarm\coordinator.py --interval 10
} catch {
    Write-Host "`n❌ Coordinator failed to start: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "`nTroubleshooting:" -ForegroundColor Yellow
    Write-Host "  1. Check .env file has all required keys" -ForegroundColor Cyan
    Write-Host "  2. Verify PostgreSQL is running" -ForegroundColor Cyan
    Write-Host "  3. Verify ARES API is running on port 8080" -ForegroundColor Cyan
    Write-Host "  4. Check agent_coordinator.log for details" -ForegroundColor Cyan
    exit 1
}
