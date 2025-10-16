# Simple Phase 1 Build Script (No hanging)
Set-Location C:\ARES_Workspace\ARES_API

Write-Host "🔨 Building ARES_API with Phase 1 changes..." -ForegroundColor Cyan

# Build in foreground with timeout protection
$process = Start-Process -FilePath "go" -ArgumentList "build","-o","ares-api.exe",".\cmd\main.go" -NoNewWindow -PassThru -Wait -WorkingDirectory $PWD

if ($process.ExitCode -eq 0 -and (Test-Path ares-api.exe)) {
    $exe = Get-Item ares-api.exe
    Write-Host "✅ BUILD SUCCESS - $([math]::Round($exe.Length/1MB,2)) MB" -ForegroundColor Green
    exit 0
} else {
    Write-Host "❌ BUILD FAILED" -ForegroundColor Red
    exit 1
}
