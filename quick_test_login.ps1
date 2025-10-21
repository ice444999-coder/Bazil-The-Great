# Quick Test - Create demo account and test login
$baseUrl = "http://localhost:8080"

Write-Host "🔐 Creating demo account..." -ForegroundColor Cyan

$username = "demo"
$email = "demo@ares.ai"
$password = "demo123"

$body = @{
    username = $username
    email = $email
    password = $password
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/signup" `
                                  -Method POST `
                                  -Body $body `
                                  -ContentType "application/json" `
                                  -ErrorAction Stop
    Write-Host "✅ Account created!" -ForegroundColor Green
}
catch {
    $error = $_.ErrorDetails.Message | ConvertFrom-Json
    if ($error.error -like "*already exists*") {
        Write-Host "⚠️ Account already exists, testing login..." -ForegroundColor Yellow
    } else {
        Write-Host "❌ Error: $($error.error)" -ForegroundColor Red
        exit 1
    }
}

# Test login
$loginBody = @{
    username = $username
    password = $password
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/login" `
                                       -Method POST `
                                       -Body $loginBody `
                                       -ContentType "application/json" `
                                       -ErrorAction Stop
    
    Write-Host "✅ Login successful!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📋 Demo Account Credentials:" -ForegroundColor Cyan
    Write-Host "  Username: $username" -ForegroundColor White
    Write-Host "  Password: $password" -ForegroundColor White
    Write-Host ""
    Write-Host "🌐 Login at: http://localhost:8080/login.html" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Access Token (first 50 chars):" -ForegroundColor Gray
    Write-Host "  $($loginResponse.access_token.Substring(0, 50))..." -ForegroundColor DarkGray
}
catch {
    $error = $_.ErrorDetails.Message | ConvertFrom-Json
    Write-Host "❌ Login failed: $($error.error)" -ForegroundColor Red
}
