# Test Runner for Instruction #6
# Runs WebSocket server and tests

Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host ("=" * 69) -ForegroundColor Cyan
Write-Host "🧪 INSTRUCTION #6 TEST RUNNER" -ForegroundColor Cyan
Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host ("=" * 69) -ForegroundColor Cyan
Write-Host ""

# Change to correct directory
Set-Location "C:\ARES_Workspace\ARES_API\internal\agent_swarm"

Write-Host "📂 Working directory: $PWD" -ForegroundColor Yellow
Write-Host ""

# Start server in background job
Write-Host "🚀 Starting WebSocket server..." -ForegroundColor Green
$serverJob = Start-Job -ScriptBlock {
    Set-Location "C:\ARES_Workspace\ARES_API\internal\agent_swarm"
    C:\Python313\python.exe test_websocket_server.py 2>&1
}

Write-Host "   Server job ID: $($serverJob.Id)" -ForegroundColor Gray
Write-Host "   Waiting 3 seconds for server to start..." -ForegroundColor Gray
Start-Sleep -Seconds 3

# Check if server started
$serverOutput = Receive-Job -Job $serverJob -Keep
if ($serverOutput -like "*server listening on*") {
    Write-Host "   ✅ Server started successfully!" -ForegroundColor Green
    Write-Host ""
} else {
    Write-Host "   ❌ Server failed to start!" -ForegroundColor Red
    Write-Host "   Output:" -ForegroundColor Red
    $serverOutput | Write-Host
    Stop-Job -Job $serverJob
    Remove-Job -Job $serverJob
    exit 1
}

# Run tests
Write-Host "🧪 Running backup/command tests..." -ForegroundColor Green
Write-Host ""

# Remove emojis from test file first
(Get-Content test_backup_command.py -Raw) -replace '[🧪✅❌⚠️📤📥🤖]', '' | Set-Content test_backup_command_safe.py

C:\Python313\python.exe test_backup_command_safe.py 2>&1

# Stop server
Write-Host ""
Write-Host "🛑 Stopping server..." -ForegroundColor Yellow
Stop-Job -Job $serverJob
Remove-Job -Job $serverJob

Write-Host ""
Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host ("=" * 69) -ForegroundColor Cyan
Write-Host "✅ Test run complete!" -ForegroundColor Cyan
Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host ("=" * 69) -ForegroundColor Cyan
