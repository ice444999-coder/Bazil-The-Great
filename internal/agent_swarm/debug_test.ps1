# Debug test - shows server errors

Set-Location "C:\ARES_Workspace\ARES_API\internal\agent_swarm"

Write-Host "Starting server in background job..."
$serverJob = Start-Job -ScriptBlock {
    Set-Location "C:\ARES_Workspace\ARES_API\internal\agent_swarm"
    C:\Python313\python.exe test_websocket_server.py 2>&1
}

Start-Sleep -Seconds 2

Write-Host "`nServer output:"
Write-Host "=" * 70
Receive-Job -Job $serverJob -Keep | Write-Host
Write-Host "=" * 70

Write-Host "`nSending test message..."
Start-Sleep -Seconds 1

C:\Python313\python.exe test_simple_backup.py 2>&1

Start-Sleep -Seconds 1

Write-Host "`n`nServer output after test:"
Write-Host "=" * 70
Receive-Job -Job $serverJob -Keep | Write-Host
Write-Host "=" * 70

Stop-Job -Job $serverJob
Remove-Job -Job $serverJob
