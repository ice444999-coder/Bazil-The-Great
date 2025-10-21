# Test Login/Signup Script for ARES API
Write-Host "🔐 ARES Authentication Test Script" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"

# Function to test signup
function Test-Signup {
    param(
        [string]$username,
        [string]$email,
        [string]$password
    )
    
    Write-Host "📝 Testing Signup..." -ForegroundColor Yellow
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
        
        Write-Host "✅ Signup successful!" -ForegroundColor Green
        Write-Host "Response: $($response | ConvertTo-Json)" -ForegroundColor Gray
        return $true
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $errorMessage = $_.ErrorDetails.Message
        
        if ($errorMessage) {
            $errorObj = $errorMessage | ConvertFrom-Json
            Write-Host "❌ Signup failed: $($errorObj.error)" -ForegroundColor Red
        } else {
            Write-Host "❌ Signup failed: $($_.Exception.Message)" -ForegroundColor Red
        }
        return $false
    }
}

# Function to test login
function Test-Login {
    param(
        [string]$username,
        [string]$password
    )
    
    Write-Host "🔑 Testing Login..." -ForegroundColor Yellow
    $body = @{
        username = $username
        password = $password
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/login" `
                                      -Method POST `
                                      -Body $body `
                                      -ContentType "application/json" `
                                      -ErrorAction Stop
        
        Write-Host "✅ Login successful!" -ForegroundColor Green
        Write-Host "Access Token: $($response.access_token.Substring(0, 50))..." -ForegroundColor Gray
        Write-Host "Refresh Token: $($response.refresh_token.Substring(0, 50))..." -ForegroundColor Gray
        
        return @{
            success = $true
            accessToken = $response.access_token
            refreshToken = $response.refresh_token
        }
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $errorMessage = $_.ErrorDetails.Message
        
        if ($errorMessage) {
            $errorObj = $errorMessage | ConvertFrom-Json
            Write-Host "❌ Login failed: $($errorObj.error)" -ForegroundColor Red
        } else {
            Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
        }
        return @{ success = $false }
    }
}

# Function to test API health
function Test-ApiHealth {
    Write-Host "🏥 Checking API Health..." -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "$baseUrl/health" -UseBasicParsing -TimeoutSec 5
        if ($response.StatusCode -eq 200) {
            Write-Host "✅ API is running!" -ForegroundColor Green
            return $true
        }
    }
    catch {
        Write-Host "❌ API is not running. Please start the API first." -ForegroundColor Red
        return $false
    }
}

# Main script
Write-Host "Step 1: Checking API Health" -ForegroundColor Cyan
if (-not (Test-ApiHealth)) {
    Write-Host ""
    Write-Host "Please start the API with: .\start_api.ps1" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "Step 2: Create Test Account" -ForegroundColor Cyan
$testUsername = "solace_ai"
$testEmail = "solace@ares.ai"
$testPassword = "ares2025!"

Write-Host "Creating account for: $testUsername" -ForegroundColor Gray
$signupSuccess = Test-Signup -username $testUsername -email $testEmail -password $testPassword

Write-Host ""
Write-Host "Step 3: Test Login" -ForegroundColor Cyan
$loginResult = Test-Login -username $testUsername -password $testPassword

if ($loginResult.success) {
    Write-Host ""
    Write-Host "✅ All tests passed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📋 Test Account Credentials:" -ForegroundColor Cyan
    Write-Host "  Username: $testUsername" -ForegroundColor White
    Write-Host "  Password: $testPassword" -ForegroundColor White
    Write-Host ""
    Write-Host "🌐 You can now login at:" -ForegroundColor Cyan
    Write-Host "  http://localhost:8080/login.html" -ForegroundColor White
} else {
    Write-Host ""
    Write-Host "⚠️ Some tests failed. Please check the errors above." -ForegroundColor Yellow
}
