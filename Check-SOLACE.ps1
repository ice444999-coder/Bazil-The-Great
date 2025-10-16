# SOLACE Dashboard - Real-time Monitoring
# Run this in PowerShell to check SOLACE's status

# ========================================
# CONFIGURATION
# ========================================
$API_URL = "http://localhost:8080"
$USERNAME = "solace_user"
$PASSWORD = "Solace2025!"

# ========================================
# LOGIN
# ========================================
Write-Host "🔐 Logging in..." -ForegroundColor Cyan
$loginBody = @{username=$USERNAME; password=$PASSWORD} | ConvertTo-Json
try {
    $loginResponse = Invoke-WebRequest -Uri "$API_URL/api/v1/users/login" -Method POST -Body $loginBody -ContentType "application/json" -UseBasicParsing
    $loginData = $loginResponse.Content | ConvertFrom-Json
    $TOKEN = $loginData.access_token
    $headers = @{Authorization="Bearer $TOKEN"}
    Write-Host "✅ Login successful!`n" -ForegroundColor Green
} catch {
    Write-Host "❌ Login failed: $_" -ForegroundColor Red
    exit 1
}

# ========================================
# SYSTEM STATUS
# ========================================
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Blue
Write-Host "🤖 SOLACE AUTONOMOUS AGENT DASHBOARD" -ForegroundColor Blue
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor Blue

# Check if ARES process is running
$aresProcess = Get-Process -Name ARES -ErrorAction SilentlyContinue
if ($aresProcess) {
    $uptime = (Get-Date) - $aresProcess.StartTime
    Write-Host "✅ ARES Process: RUNNING (PID: $($aresProcess.Id))" -ForegroundColor Green
    Write-Host "   Start Time: $($aresProcess.StartTime.ToString('yyyy-MM-dd HH:mm:ss'))" -ForegroundColor Gray
    Write-Host "   Uptime: $($uptime.Hours)h $($uptime.Minutes)m $($uptime.Seconds)s`n" -ForegroundColor Gray
} else {
    Write-Host "❌ ARES Process: NOT RUNNING`n" -ForegroundColor Red
    exit 1
}

# Check server health
Write-Host "🏥 Health Check:" -ForegroundColor Cyan
try {
    $healthResponse = Invoke-WebRequest -Uri "$API_URL/api/v1/monitoring/health" -Method GET -UseBasicParsing
    $health = $healthResponse.Content | ConvertFrom-Json
    Write-Host "   Status: $($health.status)" -ForegroundColor Green
    Write-Host "   LLM: $($health.checks.llm.status) - $($health.checks.llm.message)" -ForegroundColor Gray
    Write-Host "   Memory: $($health.checks.memory.status) - $($health.checks.memory.message)" -ForegroundColor Gray
    Write-Host "`n"
} catch {
    Write-Host "   ❌ Health check failed: $_`n" -ForegroundColor Red
}

# ========================================
# SOLACE THOUGHT JOURNAL
# ========================================
Write-Host "📖 SOLACE Thought Journal:" -ForegroundColor Cyan
$journalPath = "C:\ARES_Workspace\ARES_API\SOLACE_Journal"
if (Test-Path $journalPath) {
    $todayLog = Get-ChildItem "$journalPath\SOLACE_Thoughts_$(Get-Date -Format 'yyyy-MM-dd').log" -ErrorAction SilentlyContinue
    if ($todayLog) {
        Write-Host "   ✅ Journal active: $($todayLog.Name)" -ForegroundColor Green
        Write-Host "   Size: $($todayLog.Length) bytes" -ForegroundColor Gray
        Write-Host "   Last modified: $($todayLog.LastWriteTime.ToString('HH:mm:ss'))" -ForegroundColor Gray
        Write-Host "`n   📝 Recent thoughts:" -ForegroundColor Yellow
        Get-Content $todayLog.FullName -Tail 10 | ForEach-Object { Write-Host "      $_" -ForegroundColor Gray }
        Write-Host ""
    } else {
        Write-Host "   ⏳ No thoughts logged today yet (waiting for significant events)`n" -ForegroundColor Yellow
    }
} else {
    Write-Host "   ⏳ Journal directory not created yet (SOLACE is observing silently)`n" -ForegroundColor Yellow
}

# ========================================
# MEMORY SNAPSHOTS
# ========================================
Write-Host "🧠 Recent Memory Snapshots:" -ForegroundColor Cyan
try {
    $memoryResponse = Invoke-WebRequest -Uri "$API_URL/api/v1/memory/recall?user_id=1&limit=5" -Method GET -Headers $headers -UseBasicParsing
    $memoryData = $memoryResponse.Content | ConvertFrom-Json
    if ($memoryData.memories -and $memoryData.memories.Count -gt 0) {
        $memoryData.memories | ForEach-Object {
            Write-Host "   [$($_.timestamp)] $($_.event_type)" -ForegroundColor Green
            Write-Host "      $($_.summary)" -ForegroundColor Gray
        }
        Write-Host ""
    } else {
        Write-Host "   No autonomous decisions recorded yet`n" -ForegroundColor Yellow
    }
} catch {
    Write-Host "   ⚠️ Could not fetch memories: $_`n" -ForegroundColor Yellow
}

# ========================================
# TRADING STATUS
# ========================================
Write-Host "💰 Trading Account:" -ForegroundColor Cyan
try {
    $balanceResponse = Invoke-WebRequest -Uri "$API_URL/api/v1/balances/" -Method GET -Headers $headers -UseBasicParsing
    $balance = $balanceResponse.Content | ConvertFrom-Json
    Write-Host "   Balance: `$$($balance.amount) USD" -ForegroundColor Green
} catch {
    Write-Host "   Balance: Not initialized`n" -ForegroundColor Yellow
}

Write-Host "`n📊 Open Positions:" -ForegroundColor Cyan
try {
    $tradesResponse = Invoke-WebRequest -Uri "$API_URL/api/v1/trading/open" -Method GET -Headers $headers -UseBasicParsing
    $trades = $tradesResponse.Content | ConvertFrom-Json
    if ($trades -and $trades.Count -gt 0) {
        $trades | ForEach-Object {
            $pnl = if ($_.unrealized_pnl -ge 0) { "+`$$($_.unrealized_pnl)" } else { "-`$$([Math]::Abs($_.unrealized_pnl))" }
            $color = if ($_.unrealized_pnl -ge 0) { "Green" } else { "Red" }
            Write-Host "   $($_.symbol) $($_.side) $($_.amount) @ `$$($_.entry_price)" -ForegroundColor Gray
            Write-Host "      P&L: $pnl ($($_.unrealized_pnl_percent)%)" -ForegroundColor $color
        }
        Write-Host ""
    } else {
        Write-Host "   No open positions - SOLACE is watching markets`n" -ForegroundColor Yellow
    }
} catch {
    Write-Host "   ⚠️ Could not fetch trades: $_`n" -ForegroundColor Yellow
}

Write-Host "📈 Trading Performance:" -ForegroundColor Cyan
try {
    $perfResponse = Invoke-WebRequest -Uri "$API_URL/api/v1/trading/performance" -Method GET -Headers $headers -UseBasicParsing
    $perf = $perfResponse.Content | ConvertFrom-Json
    Write-Host "   Total Trades: $($perf.total_trades)" -ForegroundColor Gray
    if ($perf.total_trades -gt 0) {
        $winRate = if ($perf.total_trades -gt 0) { ($perf.winning_trades / $perf.total_trades * 100).ToString("F1") } else { "0.0" }
        Write-Host "   Win Rate: $winRate% ($($perf.winning_trades)W / $($perf.losing_trades)L)" -ForegroundColor Gray
        Write-Host "   Strategy Version: $($perf.strategy_version)" -ForegroundColor Gray
    }
    Write-Host ""
} catch {
    Write-Host "   ⚠️ Could not fetch performance: $_`n" -ForegroundColor Yellow
}

# ========================================
# FOOTER
# ========================================
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Blue
Write-Host "💡 TIP: Run this script periodically to monitor SOLACE" -ForegroundColor Yellow
Write-Host "   Perception cycle: Every 10 seconds" -ForegroundColor Gray
Write-Host "   Triggers: >2% price movement, >5% profit, >3% loss" -ForegroundColor Gray
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Blue
