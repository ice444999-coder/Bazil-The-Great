# Find and use working Python installation
Write-Host "🔍 Finding Python installation..." -ForegroundColor Cyan

$pythonPaths = @(
    "C:\Users\$env:USERNAME\AppData\Local\Programs\Python\Python312\python.exe",
    "C:\Users\$env:USERNAME\AppData\Local\Programs\Python\Python311\python.exe",
    "C:\Users\$env:USERNAME\AppData\Local\Programs\Python\Python310\python.exe",
    "C:\Program Files\Python312\python.exe",
    "C:\Program Files\Python311\python.exe",
    "C:\Program Files\Python310\python.exe",
    "C:\Python312\python.exe",
    "C:\Python311\python.exe",
    "C:\Python310\python.exe"
)

$workingPython = $null

foreach ($path in $pythonPaths) {
    if (Test-Path $path) {
        try {
            $version = & $path --version 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Found working Python: $path" -ForegroundColor Green
                Write-Host "   Version: $version" -ForegroundColor Cyan
                $workingPython = $path
                break
            }
        } catch {
            continue
        }
    }
}

if (-not $workingPython) {
    Write-Host "❌ No working Python installation found" -ForegroundColor Red
    Write-Host "`nPlease install Python from: https://www.python.org/downloads/" -ForegroundColor Yellow
    exit 1
}

# Create a simple wrapper script
$wrapperContent = @"
# Auto-generated Python wrapper
`$pythonExe = "$workingPython"

Write-Host "Using Python: `$pythonExe" -ForegroundColor Cyan

# Install dependencies
Write-Host "`nInstalling dependencies..." -ForegroundColor Yellow
& `$pythonExe -m pip install --quiet --upgrade pip
& `$pythonExe -m pip install --quiet psycopg2-binary openai anthropic python-dotenv requests playwright

if (`$LASTEXITCODE -eq 0) {
    Write-Host "✅ Dependencies installed" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to install dependencies" -ForegroundColor Red
    exit 1
}

# Install Playwright browsers
Write-Host "`nInstalling Playwright browsers..." -ForegroundColor Yellow
& `$pythonExe -m playwright install chromium --quiet

if (`$LASTEXITCODE -eq 0) {
    Write-Host "✅ Playwright installed" -ForegroundColor Green
} else {
    Write-Host "⚠️  Playwright installation had issues (may still work)" -ForegroundColor Yellow
}

# Start coordinator
Write-Host "`n🚀 Starting ARES Agent Swarm Coordinator..." -ForegroundColor Cyan
Write-Host "   Dashboard: http://localhost:8080/web/agent-dashboard.html`n" -ForegroundColor Yellow

Set-Location C:\ARES_Workspace\ARES_API
& `$pythonExe internal\agent_swarm\coordinator.py
"@

$wrapperContent | Out-File -FilePath "start_coordinator_auto.ps1" -Encoding UTF8

Write-Host "`n✅ Created wrapper script: start_coordinator_auto.ps1" -ForegroundColor Green
Write-Host "`nTo start coordinator, run:" -ForegroundColor Yellow
Write-Host "   .\start_coordinator_auto.ps1" -ForegroundColor Cyan
