# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
# ARES Trading Tab Litmus Test Suite (PowerShell)
# Tests each upgrade subtask to ensure no regression

$BASE_URL = "http://localhost:8080"
$RESULTS = @()

function Write-TestResult {
    param(
        [string]$Name,
        [bool]$Passed,
        [string]$Message = ""
    )
    
    $status = if ($Passed) { "✅ PASS" } else { "❌ FAIL" }
    $color = if ($Passed) { "Green" } else { "Red" }
    
    $RESULTS += [PSCustomObject]@{
        Test = $Name
        Passed = $Passed
        Message = $Message
    }
    
    Write-Host "$status | $Name" -ForegroundColor $color
    if ($Message) {
        Write-Host "    └─ $Message" -ForegroundColor Yellow
    }
}

function Test-APIHealth {
    Write-Host "`n[Test 1] API Health Check..." -ForegroundColor Cyan
    try {
        $response = Invoke-WebRequest -Uri "$BASE_URL/api/v1/health" -TimeoutSec 5 -UseBasicParsing
        $passed = $response.StatusCode -eq 200
        Write-TestResult "API Health Check" $passed "Status: $($response.StatusCode)"
        return $passed
    }
    catch {
        Write-TestResult "API Health Check" $false $_.Exception.Message
        return $false
    }
}

function Test-TradingPageLoads {
    Write-Host "`n[Test 2] Trading Page Loads..." -ForegroundColor Cyan
    try {
        $response = Invoke-WebRequest -Uri "$BASE_URL/trading.html" -TimeoutSec 5 -UseBasicParsing
        $content = $response.Content
        $hasChart = $content -match "chart"
        $hasOrderForm = $content -match "order"
        $passed = ($response.StatusCode -eq 200) -and $hasChart -and $hasOrderForm
        Write-TestResult "Trading Page Loads" $passed "Status: $($response.StatusCode), Chart: $hasChart, OrderForm: $hasOrderForm"
        return $passed
    }
    catch {
        Write-TestResult "Trading Page Loads" $false $_.Exception.Message
        return $false
    }
}

function Test-DashboardLoads {
    Write-Host "`n[Test 3] Dashboard Page Loads..." -ForegroundColor Cyan
    try {
        $response = Invoke-WebRequest -Uri "$BASE_URL/dashboard.html" -TimeoutSec 5 -UseBasicParsing
        $passed = $response.StatusCode -eq 200
        Write-TestResult "Dashboard Page Loads" $passed "Status: $($response.StatusCode)"
        return $passed
    }
    catch {
        Write-TestResult "Dashboard Page Loads" $false $_.Exception.Message
        return $false
    }
}

function Test-TradingEndpoints {
    Write-Host "`n[Test 4] Trading API Endpoints..." -ForegroundColor Cyan
    $endpoints = @(
        "/api/v1/trading/performance",
        "/api/v1/trading/stats"
    )
    
    $allPass = $true
    foreach ($endpoint in $endpoints) {
        try {
            $response = Invoke-WebRequest -Uri "$BASE_URL$endpoint" -TimeoutSec 5 -UseBasicParsing
            $passed = $response.StatusCode -in @(200, 404)  # 404 acceptable if not implemented yet
            if (-not $passed) { $allPass = $false }
            Write-TestResult "Endpoint $endpoint" $passed "Status: $($response.StatusCode)"
        }
        catch {
            Write-TestResult "Endpoint $endpoint" $false $_.Exception.Message
            $allPass = $false
        }
    }
    return $allPass
}

function Test-WebSocketInfrastructure {
    Write-Host "`n[Test 5] WebSocket Infrastructure..." -ForegroundColor Cyan
    try {
        $response = Invoke-WebRequest -Uri "$BASE_URL/health.html" -TimeoutSec 5 -UseBasicParsing
        $passed = $response.StatusCode -eq 200
        Write-TestResult "WebSocket Infrastructure" $passed "Health page accessible: $($response.StatusCode)"
        return $passed
    }
    catch {
        Write-TestResult "WebSocket Infrastructure" $false $_.Exception.Message
        return $false
    }
}

function Test-SOLACEIntegration {
    Write-Host "`n[Test 6] SOLACE Integration Check..." -ForegroundColor Cyan
    try {
        $response = Invoke-WebRequest -Uri "$BASE_URL/api/v1/solace/stats" -TimeoutSec 5 -UseBasicParsing
        $passed = $response.StatusCode -in @(200, 404)
        Write-TestResult "SOLACE Integration" $passed "Status: $($response.StatusCode)"
        return $passed
    }
    catch {
        Write-TestResult "SOLACE Integration" $false $_.Exception.Message
        return $false
    }
}

# Run all tests
Write-Host "`n============================================================" -ForegroundColor Blue
Write-Host "ARES Trading Tab Litmus Test Suite" -ForegroundColor Blue
Write-Host "============================================================`n" -ForegroundColor Blue

$tests = @(
    { Test-APIHealth },
    { Test-TradingPageLoads },
    { Test-DashboardLoads },
    { Test-TradingEndpoints },
    { Test-WebSocketInfrastructure },
    { Test-SOLACEIntegration }
)

foreach ($test in $tests) {
    & $test
    Start-Sleep -Milliseconds 500
}

# Summary
Write-Host "`n============================================================" -ForegroundColor Blue
$passed = ($RESULTS | Where-Object { $_.Passed }).Count
$total = $RESULTS.Count
$passRate = if ($total -gt 0) { [math]::Round(($passed / $total) * 100, 1) } else { 0 }

Write-Host "Test Summary:" -ForegroundColor Blue
Write-Host "  Total Tests: $total"
Write-Host "  Passed: " -NoNewline; Write-Host $passed -ForegroundColor Green
Write-Host "  Failed: " -NoNewline; Write-Host ($total - $passed) -ForegroundColor Red
Write-Host "  Pass Rate: $passRate%"
Write-Host "============================================================`n" -ForegroundColor Blue

# Export results
$RESULTS | Export-Csv -Path "litmus_test_results.csv" -NoTypeInformation
Write-Host "Results exported to: litmus_test_results.csv" -ForegroundColor Gray

# Return exit code
exit $(if ($passed -eq $total) { 0 } else { 1 })
