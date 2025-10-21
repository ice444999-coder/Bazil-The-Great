# Create Test User Script for ARES API
Write-Host "🔐 ARES Create Test User" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"

# Get username from user
Write-Host "Enter username (default: testuser): " -NoNewline -ForegroundColor Yellow
$username = Read-Host
if ([string]::IsNullOrWhiteSpace($username)) {
    $username = "testuser"
}

# Get email from user
Write-Host "Enter email (default: test@ares.ai): " -NoNewline -ForegroundColor Yellow
$email = Read-Host
if ([string]::IsNullOrWhiteSpace($email)) {
    $email = "test@ares.ai"
}

# Get password from user
Write-Host "Enter password (default: password123): " -NoNewline -ForegroundColor Yellow
$password = Read-Host
if ([string]::IsNullOrWhiteSpace($password)) {
    $password = "password123"
}

Write-Host ""
Write-Host "Creating account..." -ForegroundColor Yellow
Write-Host "  Username: $username" -ForegroundColor Gray
Write-Host "  Email: $email" -ForegroundColor Gray
Write-Host ""

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
    
    Write-Host "✅ Account created successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Now testing login..." -ForegroundColor Yellow
    
    # Test login
    $loginBody = @{
        username = $username
        password = $password
    } | ConvertTo-Json
    
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/login" `
                                       -Method POST `
                                       -Body $loginBody `
                                       -ContentType "application/json" `
                                       -ErrorAction Stop
    
    Write-Host "✅ Login successful!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📋 Your Login Credentials:" -ForegroundColor Cyan
    Write-Host "  Username: $username" -ForegroundColor White
    Write-Host "  Password: $password" -ForegroundColor White
    Write-Host ""
    Write-Host "🌐 Login at: http://localhost:8080/login.html" -ForegroundColor Cyan
    Write-Host ""
}
catch {
    $errorMessage = $_.ErrorDetails.Message
    if ($errorMessage) {
        $errorObj = $errorMessage | ConvertFrom-Json
        Write-Host "❌ Error: $($errorObj.error)" -ForegroundColor Red
    } else {
        Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
    }
}
