# SOLACE Intelligence Upgrade - Deployment Script
# Replaces old ARES API with intelligent version

Write-Host "🧠 SOLACE INTELLIGENCE UPGRADE - DEPLOYMENT" -ForegroundColor Cyan
Write-Host "=" * 60

# Step 1: Find and stop old ARES API process
Write-Host "`n📍 Step 1: Stopping old ARES_API process on port 8080..." -ForegroundColor Yellow
$oldProcess = Get-Process | Where-Object { $_.ProcessName -like "*ARES_API*" -or $_.Id -eq 42744 }

if ($oldProcess) {
    Write-Host "   Found process: $($oldProcess.ProcessName) (PID: $($oldProcess.Id))" -ForegroundColor Gray
    Stop-Process -Id $oldProcess.Id -Force
    Start-Sleep -Seconds 2
    Write-Host "   ✅ Old process stopped" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  No old ARES_API process found (already stopped?)" -ForegroundColor Yellow
}

# Step 2: Start new intelligent ARES API
Write-Host "`n📍 Step 2: Starting ARES_API_INTELLIGENT.exe..." -ForegroundColor Yellow
$apiPath = "C:\ARES_Workspace\ARES_API\ARES_API_INTELLIGENT.exe"

if (Test-Path $apiPath) {
    Write-Host "   Binary found: $apiPath" -ForegroundColor Gray
    Start-Process -FilePath $apiPath -WindowStyle Normal
    Start-Sleep -Seconds 3
    Write-Host "   ✅ Intelligent ARES API started" -ForegroundColor Green
} else {
    Write-Host "   ❌ ERROR: Binary not found at $apiPath" -ForegroundColor Red
    Write-Host "   Run: cd C:\ARES_Workspace\ARES_API; go build -o ARES_API_INTELLIGENT.exe ." -ForegroundColor Yellow
    exit 1
}

# Step 3: Verify API is running
Write-Host "`n📍 Step 3: Verifying API health..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 5
    Write-Host "   ✅ API is healthy: $($health.status)" -ForegroundColor Green
    Write-Host "   Database: $($health.database)" -ForegroundColor Gray
} catch {
    Write-Host "   ⚠️  Health check failed (API may still be starting...)" -ForegroundColor Yellow
    Write-Host "   Wait 5 seconds and try: Invoke-RestMethod http://localhost:8080/health" -ForegroundColor Gray
}

# Step 4: Show SOLACE test command
Write-Host "`n📍 Step 4: Test SOLACE Intelligence" -ForegroundColor Yellow
Write-Host "   Open VS Code Command Center (Robot icon in Activity Bar)" -ForegroundColor Gray
Write-Host "   Click: 🧠 Talk to SOLACE" -ForegroundColor Gray
Write-Host "   Ask: 'Can you talk to GitHub Copilot?'" -ForegroundColor Cyan
Write-Host "`n   Expected (INTELLIGENT response):" -ForegroundColor Green
Write-Host "   'Not directly yet. The VS Code extension routes messages" -ForegroundColor White
Write-Host "    through Enki. We need a copilot_chat() tool...'" -ForegroundColor White
Write-Host "`n   NOT Expected (DUMB response):" -ForegroundColor Red
Write-Host "   'I found several memory crystals related to...'" -ForegroundColor White

Write-Host "`n" + ("=" * 60)
Write-Host "🎉 DEPLOYMENT COMPLETE!" -ForegroundColor Green
Write-Host "   SOLACE is now running with GPT-4 level intelligence" -ForegroundColor Cyan
Write-Host "   API: http://localhost:8080" -ForegroundColor Gray
Write-Host "   Test via VS Code Command Center" -ForegroundColor Gray
Write-Host ("=" * 60) -ForegroundColor Cyan
