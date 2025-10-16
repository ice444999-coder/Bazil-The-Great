# Create task for SOLACE to fix trading dashboard UI

$task = @{
    task_type = "ui_audit_and_fix"
    description = @"
CRITICAL UI FIX - Trading Dashboard (localhost:8080/chat.html)

CURRENT PROBLEMS:
1. Navigation buttons route incorrectly (clicking buttons goes to wrong pages)
2. TradingView chart visible but isolated - no trading panels around it
3. Missing P&L display (Total P&L, Today P&L, Win Rate, Avg Win/Loss)
4. Missing positions table (10 open trades with Entry/SL/TP/Current/P&L)
5. Missing Solace thinking feed (real-time decision log)

REQUIRED LAYOUT:
┌─────────────────────────────────────────────┐
│ Top Nav: Dashboard | Trading | Positions   │
├─────────┬───────────────────┬───────────────┤
│ POS     │   CHART (60%)     │   P&L STATS   │
│ TABLE   │   TradingView     │   Total: $$247 │
│ (20%)   │   (existing)      │   Win: 62%    │
├─────────┴───────────────────┴───────────────┤
│ SOLACE BRAIN FEED (bottom 20% height)      │
│ [14:23] Analyzing BTC... RSI: 47.3, Buy    │
└─────────────────────────────────────────────┘

WORKFLOW:
1. SENTINEL: Audit navigation - click every button, log what breaks
2. ARCHITECT: Design container layout with specs for each panel
3. FORGE: Build components (positions table, P&L stats, Solace feed)
4. FORGE: Fix navigation routing based on SENTINEL findings
5. FORGE: Integrate TradingView chart into new layout (don't break it)
6. SENTINEL: Verify all components work, navigation fixed
7. SOLACE: Deploy and report completion

API ENDPOINTS AVAILABLE:
- GET /api/v1/trading/positions (for positions table)
- GET /api/v1/trading/performance (for P&L stats)
- GET /api/v1/agents/tasks?agent=SOLACE&status=completed (for Solace feed)

WEBSOCKET: ws://localhost:8080/ws/trading (real-time price updates)

THEME: Dark (Binance style) - #1e2329 background, #2b3139 cards, green profits, red losses

SUCCESS CRITERIA:
- All navigation buttons route correctly
- Positions table shows 10 open trades (or "No positions")
- P&L stats display with correct calculations
- Solace feed shows last 10 decisions
- TradingView chart still works
- Layout responsive (doesn't break on resize)
- Dark theme consistent
"@
    priority = 10
    file_paths = @("web/chat.html", "web/css/", "web/js/", "internal/api/trading_handler.go")
    context = @{
        current_url = "http://localhost:8080/chat.html"
        issues = @("broken_navigation", "missing_pnl", "missing_positions", "missing_solace_feed")
        has_tradingview_chart = $true
        backend_ready = $true
    }
} | ConvertTo-Json -Depth 5

Write-Host "🚀 Creating UI fix task for SOLACE..." -ForegroundColor Cyan

try {
    # Send task to SOLACE
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/agents/tasks" `
        -Method Post `
        -ContentType "application/json" `
        -Body $task

    Write-Host "✅ UI fix task created and assigned to SOLACE" -ForegroundColor Green
    Write-Host "   Task ID: $($response.task_id)" -ForegroundColor White
    Write-Host ""
    Write-Host "📊 Watch progress: http://localhost:8080/web/agent-dashboard.html" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Failed to create task: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
