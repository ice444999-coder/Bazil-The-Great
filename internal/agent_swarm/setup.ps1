# ====================================================================
# 🚀 ARES Coordinator Setup Script
# ====================================================================
# This script creates a fresh virtual environment and installs all
# required dependencies for the ARES Agent Swarm Coordinator.
# ====================================================================

Write-Host ""
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "🚀 ARES Coordinator Setup" -ForegroundColor Cyan
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""

# Get script directory and change to it
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir
Write-Host "Working directory: $ScriptDir" -ForegroundColor Gray
Write-Host ""

# Step 1: Clean up old virtual environment
Write-Host "📦 Step 1: Cleaning up old virtual environment..." -ForegroundColor Yellow
Remove-Item -Recurse -Force venv -ErrorAction SilentlyContinue
if ($?) {
    Write-Host "   ✅ Old venv removed" -ForegroundColor Green
} else {
    Write-Host "   ℹ️  No previous venv found" -ForegroundColor Gray
}
Write-Host ""

# Step 2: Create new virtual environment
Write-Host "📦 Step 2: Creating fresh virtual environment..." -ForegroundColor Yellow
C:\Python313\python.exe -m venv venv
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ Virtual environment created" -ForegroundColor Green
} else {
    Write-Host "   ❌ Failed to create virtual environment" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 3: Activate virtual environment
Write-Host "📦 Step 3: Activating virtual environment..." -ForegroundColor Yellow
$venvActivate = Join-Path $ScriptDir "venv\Scripts\Activate.ps1"
& $venvActivate
Write-Host "   ✅ Virtual environment activated" -ForegroundColor Green
Write-Host ""

# Step 4: Upgrade pip
Write-Host "📦 Step 4: Upgrading pip..." -ForegroundColor Yellow
$pythonExe = Join-Path $ScriptDir "venv\Scripts\python.exe"
& $pythonExe -m pip install --upgrade pip --quiet
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ pip upgraded" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  pip upgrade had warnings (continuing)" -ForegroundColor Yellow
}
Write-Host ""

# Step 5: Install dependencies
Write-Host "📦 Step 5: Installing dependencies..." -ForegroundColor Yellow
$requirementsPath = Join-Path $ScriptDir "requirements.txt"
if (Test-Path $requirementsPath) {
    Write-Host "   Installing from requirements.txt..." -ForegroundColor Gray
    & $pythonExe -m pip install -r $requirementsPath --quiet
} else {
    Write-Host "   Installing packages directly..." -ForegroundColor Gray
    & $pythonExe -m pip install psycopg2-binary openai anthropic websockets python-dotenv requests --quiet
}

if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ Dependencies installed" -ForegroundColor Green
} else {
    Write-Host "   ❌ Failed to install dependencies" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 6: Check for .env file
Write-Host "📦 Step 6: Checking environment configuration..." -ForegroundColor Yellow
$envPath = Join-Path $ScriptDir ".env"
if (Test-Path $envPath) {
    Write-Host "   ✅ .env file found" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  .env file not found" -ForegroundColor Yellow
    Write-Host "   📝 You need to create a .env file" -ForegroundColor Yellow
}
Write-Host ""

# Done!
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "✅ Setup Complete!" -ForegroundColor Green
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "📝 Next Steps:" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path $envPath)) {
    Write-Host "   1. Copy .env.example to .env:" -ForegroundColor White
    Write-Host "      Copy-Item .env.example .env" -ForegroundColor Gray
    Write-Host ""
    Write-Host "   2. Edit .env and add your API keys:" -ForegroundColor White
    Write-Host "      notepad .env" -ForegroundColor Gray
    Write-Host ""
    Write-Host "   3. Start the coordinator:" -ForegroundColor White
    Write-Host "      powershell -File .\start.ps1" -ForegroundColor Gray
} else {
    Write-Host "   1. Start the coordinator:" -ForegroundColor White
    Write-Host "      powershell -File .\start.ps1" -ForegroundColor Gray
}

Write-Host ""
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""
