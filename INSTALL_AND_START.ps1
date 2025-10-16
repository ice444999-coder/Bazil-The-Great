# ARES Agent Swarm - Complete Setup with Python Installation
# This script will download Python if needed and set up everything

$ErrorActionPreference = "Stop"

Write-Host "================================" -ForegroundColor Cyan
Write-Host "ARES AGENT SWARM - COMPLETE SETUP" -ForegroundColor Cyan
Write-Host "================================`n" -ForegroundColor Cyan

# Step 1: Check for Python
Write-Host "[1/6] Checking Python installation..." -ForegroundColor Yellow

$pythonCommands = @("python", "python3", "py")
$workingPython = $null

foreach ($cmd in $pythonCommands) {
    if (Get-Command $cmd -ErrorAction SilentlyContinue) {
        try {
            $version = & $cmd --version 2>&1
            if ($version -match "Python (\d+)\.(\d+)") {
                $major = [int]$Matches[1]
                $minor = [int]$Matches[2]
                if ($major -ge 3 -and $minor -ge 9) {
                    Write-Host "  ✅ Found Python: $cmd ($version)" -ForegroundColor Green
                    $workingPython = $cmd
                    break
                }
            }
        } catch {
            continue
        }
    }
}

if (-not $workingPython) {
    Write-Host "  ❌ Python 3.9+ not found. Downloading..." -ForegroundColor Yellow
    
    # Download Python 3.12 installer
    $pythonUrl = "https://www.python.org/ftp/python/3.12.0/python-3.12.0-amd64.exe"
    $installerPath = "$env:TEMP\python-installer.exe"
    
    Write-Host "  📥 Downloading Python 3.12..." -ForegroundColor Cyan
    try {
        Invoke-WebRequest -Uri $pythonUrl -OutFile $installerPath -UseBasicParsing
        Write-Host "  ✅ Downloaded Python installer" -ForegroundColor Green
        
        # Install Python silently
        Write-Host "  ⚙️  Installing Python (this may take a minute)..." -ForegroundColor Cyan
        Start-Process -FilePath $installerPath -ArgumentList "/quiet", "InstallAllUsers=0", "PrependPath=1", "Include_test=0" -Wait
        
        Write-Host "  ✅ Python installed" -ForegroundColor Green
        
        # Refresh PATH
        $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "User") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "Machine")
        
        # Find newly installed Python
        Start-Sleep -Seconds 2
        foreach ($cmd in $pythonCommands) {
            if (Get-Command $cmd -ErrorAction SilentlyContinue) {
                $workingPython = $cmd
                Write-Host "  ✅ Python command available: $cmd" -ForegroundColor Green
                break
            }
        }
        
        if (-not $workingPython) {
            # Try direct path
            $pythonPath = "$env:LOCALAPPDATA\Programs\Python\Python312\python.exe"
            if (Test-Path $pythonPath) {
                $workingPython = $pythonPath
                Write-Host "  ✅ Python available at: $pythonPath" -ForegroundColor Green
            } else {
                Write-Host "  ⚠️  Python installed but not in PATH. Please restart terminal." -ForegroundColor Yellow
                Write-Host "  Using direct path for now..." -ForegroundColor Yellow
                $workingPython = "python"
            }
        }
        
        # Cleanup installer
        Remove-Item $installerPath -Force -ErrorAction SilentlyContinue
        
    } catch {
        Write-Host "  ❌ Failed to download/install Python: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host "  Please install Python manually from: https://www.python.org/downloads/" -ForegroundColor Yellow
        exit 1
    }
}

# Step 2: Verify .env
Write-Host "`n[2/6] Verifying .env configuration..." -ForegroundColor Yellow
Set-Location C:\ARES_Workspace\ARES_API

if (-not (Test-Path ".env")) {
    Write-Host "  ❌ .env file not found!" -ForegroundColor Red
    exit 1
}

$envContent = Get-Content ".env" -Raw
$requiredKeys = @("OPENAI_API_KEY", "CLAUDE_API_KEY", "DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME")
$allFound = $true

foreach ($key in $requiredKeys) {
    if ($envContent -match "$key=.+") {
        Write-Host "  ✅ $key configured" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $key missing" -ForegroundColor Red
        $allFound = $false
    }
}

if (-not $allFound) {
    Write-Host "  ❌ Missing required keys in .env" -ForegroundColor Red
    exit 1
}

# Step 3: Install Python dependencies
Write-Host "`n[3/6] Installing Python dependencies..." -ForegroundColor Yellow
try {
    & $workingPython -m pip install --quiet --upgrade pip 2>&1 | Out-Null
    & $workingPython -m pip install --quiet psycopg2-binary openai anthropic python-dotenv requests playwright
    Write-Host "  ✅ Dependencies installed" -ForegroundColor Green
} catch {
    Write-Host "  ⚠️  Some dependencies may have issues, continuing..." -ForegroundColor Yellow
}

# Step 4: Install Playwright browsers
Write-Host "`n[4/6] Installing Playwright browsers..." -ForegroundColor Yellow
try {
    & $workingPython -m playwright install chromium
    Write-Host "  ✅ Playwright Chromium installed" -ForegroundColor Green
} catch {
    Write-Host "  ⚠️  Playwright installation had issues (may still work)" -ForegroundColor Yellow
}

# Step 5: Verify services
Write-Host "`n[5/6] Verifying services..." -ForegroundColor Yellow

# Check ARES API
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -TimeoutSec 2 -ErrorAction Stop
    Write-Host "  ✅ ARES API running (port 8080)" -ForegroundColor Green
} catch {
    Write-Host "  ❌ ARES API not responding" -ForegroundColor Red
    Write-Host "     Start ARES API first: .\ares_api.exe" -ForegroundColor Yellow
    exit 1
}

# Check Ollama
try {
    $response = Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -TimeoutSec 2 -ErrorAction Stop
    $deepseek = $response.models | Where-Object { $_.name -like "deepseek-r1*" }
    if ($deepseek) {
        Write-Host "  ✅ Ollama running with DeepSeek models" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  Ollama running but no DeepSeek model" -ForegroundColor Yellow
        Write-Host "     Run: ollama pull deepseek-r1:14b" -ForegroundColor Cyan
    }
} catch {
    Write-Host "  ❌ Ollama not responding" -ForegroundColor Red
    Write-Host "     Start Ollama first" -ForegroundColor Yellow
    exit 1
}

# Check PostgreSQL
try {
    $env:PGPASSWORD = 'ARESISWAKING'
    $result = & 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -h localhost -U ARES -d ares_db -c "SELECT COUNT(*) FROM agent_registry;" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ PostgreSQL connected" -ForegroundColor Green
    } else {
        throw "Connection failed"
    }
} catch {
    Write-Host "  ❌ PostgreSQL connection failed" -ForegroundColor Red
    exit 1
}

# Step 6: Start Coordinator
Write-Host "`n[6/6] Starting Agent Swarm Coordinator..." -ForegroundColor Yellow
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "🚀 LAUNCHING ARES AGENT SWARM" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Python: $workingPython" -ForegroundColor Gray
Write-Host "Poll Interval: 10 seconds" -ForegroundColor Gray
Write-Host "Dashboard: http://localhost:8080/web/agent-dashboard.html" -ForegroundColor Yellow
Write-Host ""
Write-Host "Agents:" -ForegroundColor Cyan
Write-Host "  - SOLACE (OpenAI GPT-4): Strategy & Coordination" -ForegroundColor White
Write-Host "  - FORGE (Claude 3.5): UI Building & Coding" -ForegroundColor White
Write-Host "  - ARCHITECT (DeepSeek-R1): Planning & Architecture" -ForegroundColor White
Write-Host "  - SENTINEL (DeepSeek-R1): Testing & Debugging" -ForegroundColor White
Write-Host ""
Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
Write-Host "========================================`n" -ForegroundColor Cyan

# Start coordinator
try {
    & $workingPython internal\agent_swarm\coordinator.py --interval 10
} catch {
    Write-Host "`n❌ Coordinator error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
