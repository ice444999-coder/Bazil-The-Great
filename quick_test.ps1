# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# Quick Agent Swarm Readiness Test
Write-Host "🧪 ARES Agent Swarm - Quick Readiness Test" -ForegroundColor Cyan
Write-Host "=========================================`n" -ForegroundColor Cyan

# Test 1: Check .env
Write-Host "[1/5] Checking .env configuration..." -ForegroundColor Yellow
$envPath = "C:\ARES_Workspace\ARES_API\.env"
if (Test-Path $envPath) {
    $env = Get-Content $envPath -Raw
    $checks = @{
        "OPENAI_API_KEY" = $env -match "OPENAI_API_KEY=sk-"
        "CLAUDE_API_KEY" = $env -match "CLAUDE_API_KEY=sk-ant-"
        "DB_HOST" = $env -match "DB_HOST="
        "DB_USER" = $env -match "DB_USER="
    }
    
    $allGood = $true
    foreach ($key in $checks.Keys) {
        if ($checks[$key]) {
            Write-Host "  ✅ $key configured" -ForegroundColor Green
        } else {
            Write-Host "  ❌ $key missing" -ForegroundColor Red
            $allGood = $false
        }
    }
    
    if ($allGood) {
        Write-Host "  ✅ All API keys configured`n" -ForegroundColor Green
    }
} else {
    Write-Host "  ❌ .env file not found`n" -ForegroundColor Red
}

# Test 2: Check ARES API
Write-Host "[2/5] Checking ARES API..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -TimeoutSec 3 -ErrorAction Stop
    Write-Host "  ✅ ARES API running on port 8080`n" -ForegroundColor Green
} catch {
    Write-Host "  ❌ ARES API not responding`n" -ForegroundColor Red
}

# Test 3: Check Ollama
Write-Host "[3/5] Checking Ollama..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -TimeoutSec 3 -ErrorAction Stop
    $deepseekModels = $response.models | Where-Object { $_.name -like "deepseek-r1*" }
    
    if ($deepseekModels.Count -gt 0) {
        Write-Host "  ✅ Ollama running with DeepSeek models:" -ForegroundColor Green
        foreach ($model in $deepseekModels) {
            $sizeMB = [math]::Round($model.size / 1MB, 0)
            Write-Host "     - $($model.name) ($sizeMB MB)" -ForegroundColor Cyan
        }
        Write-Host ""
    } else {
        Write-Host "  ⚠️  Ollama running but no DeepSeek models found" -ForegroundColor Yellow
        Write-Host "     Run: ollama pull deepseek-r1:14b`n" -ForegroundColor Cyan
    }
} catch {
    Write-Host "  ❌ Ollama not responding on port 11434`n" -ForegroundColor Red
}

# Test 4: Check PostgreSQL
Write-Host "[4/5] Checking PostgreSQL..." -ForegroundColor Yellow
try {
    $env:PGPASSWORD = 'ARESISWAKING'
    $result = & 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -h localhost -U ARES -d ares_db -c "SELECT COUNT(*) FROM agent_registry;" 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ PostgreSQL connected, agent_registry accessible`n" -ForegroundColor Green
    } else {
        Write-Host "  ❌ PostgreSQL connection failed`n" -ForegroundColor Red
    }
} catch {
    Write-Host "  ❌ PostgreSQL error: $($_.Exception.Message)`n" -ForegroundColor Red
}

# Test 5: Check Python
Write-Host "[5/5] Checking Python..." -ForegroundColor Yellow
try {
    # Try multiple Python commands
    $pythonCmd = $null
    
    if (Get-Command python -ErrorAction SilentlyContinue) {
        $pythonCmd = "python"
    } elseif (Get-Command python3 -ErrorAction SilentlyContinue) {
        $pythonCmd = "python3"
    } elseif (Get-Command py -ErrorAction SilentlyContinue) {
        $pythonCmd = "py"
    }
    
    if ($pythonCmd) {
        $version = & $pythonCmd --version 2>&1
        Write-Host "  ✅ Python found: $version" -ForegroundColor Green
        Write-Host "     Command: $pythonCmd`n" -ForegroundColor Cyan
    } else {
        Write-Host "  ❌ Python not found in PATH`n" -ForegroundColor Red
    }
} catch {
    Write-Host "  ❌ Python check failed: $($_.Exception.Message)`n" -ForegroundColor Red
}

# Summary
Write-Host "`n=========================================`n" -ForegroundColor Cyan
Write-Host "📊 SUMMARY" -ForegroundColor Cyan
Write-Host "=========================================`n" -ForegroundColor Cyan
Write-Host "Ready to start coordinator if all checks passed ✅" -ForegroundColor Green
Write-Host "To start: .\internal\agent_swarm\start-coordinator.ps1`n" -ForegroundColor Yellow
