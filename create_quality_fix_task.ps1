# Fix All Code Quality Issues - Zero Tolerance
# Task for ARES Agent Swarm

Write-Host "🧹 Creating comprehensive code quality fix task..." -ForegroundColor Cyan

$env:PGPASSWORD = 'ARESISWAKING'

# Create task in database
$taskQuery = @"
INSERT INTO task_queue (task_type, description, priority, context, file_paths)
VALUES (
    'code_refactoring',
    'CRITICAL: Fix ALL 278 warnings and 8 errors. Zero tolerance policy.

PHASE 1 - Fix React Components (Priority: CRITICAL):
1. AdvancedOrderForm.tsx - Remove ALL inline styles (40+ violations)
   - Create styled-components or CSS module
   - Extract all inline style objects to classes
   - Maintain exact visual appearance

2. OpenPositionsTable.tsx - Remove ALL inline styles (5+ violations)
   - Use CSS modules for table styling
   - Keep current layout and colors

PHASE 2 - Fix HTML Files (Priority: HIGH):
3. code-ide.html - Remove ALL inline styles (7+ violations)
   - Move to <style> block or external CSS
   - Fix unsupported CSS property: field-sizing

PHASE 3 - Standards & Validation (Priority: MEDIUM):
4. Run ESLint with --max-warnings 0
5. Run TypeScript with strict mode
6. Validate all CSS browser compatibility
7. Remove any deprecated properties

SUCCESS CRITERIA:
- ZERO errors
- ZERO warnings
- All functionality preserved
- Visual appearance identical
- Browser compatibility: Chrome 120+, Firefox 120+, Safari 16+

EXECUTION PLAN:
1. SENTINEL: Audit all files, list every violation
2. ARCHITECT: Design refactoring approach (CSS modules vs styled-components)
3. FORGE: Implement fixes, test each component
4. SENTINEL: Validate zero warnings, verify visual regression
5. SOLACE: Final review and approval

Expected Duration: 30-45 minutes
Quality Standard: PRODUCTION READY - ZERO DEFECTS',
    10,
    '{
        "strict_mode": true,
        "max_warnings": 0,
        "max_errors": 0,
        "browser_targets": ["chrome >= 120", "firefox >= 120", "safari >= 16"],
        "code_style": "airbnb",
        "testing_required": true
    }'::jsonb,
    '["frontend/src/components/AdvancedOrderForm.tsx", "frontend/src/components/OpenPositionsTable.tsx", "web/code-ide.html"]'::jsonb
) RETURNING task_id, created_at;
"@

Write-Host "`nExecuting task creation..." -ForegroundColor Yellow
$result = & 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -h localhost -U ARES -d ares_db -c $taskQuery

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Code quality fix task created successfully!" -ForegroundColor Green
    Write-Host "`nTask Details:" -ForegroundColor Cyan
    Write-Host "  Type: code_refactoring" -ForegroundColor White
    Write-Host "  Priority: 10 (CRITICAL)" -ForegroundColor Red
    Write-Host "  Files: 3 files to fix" -ForegroundColor White
    Write-Host "  Warnings to fix: 278" -ForegroundColor Yellow
    Write-Host "  Errors to fix: 8" -ForegroundColor Red
    Write-Host "  Target: ZERO DEFECTS" -ForegroundColor Green
    
    Write-Host "`n📊 Monitor progress:" -ForegroundColor Cyan
    Write-Host "  Dashboard: http://localhost:8080/web/agent-dashboard.html" -ForegroundColor Yellow
    Write-Host "  Log: Get-Content agent_coordinator.log -Tail 50 -Wait" -ForegroundColor Gray
    
    Write-Host "`n🤖 Agent Execution Flow:" -ForegroundColor Cyan
    Write-Host "  1. SOLACE assigns to SENTINEL (audit)" -ForegroundColor White
    Write-Host "  2. SENTINEL lists all violations" -ForegroundColor White
    Write-Host "  3. ARCHITECT designs refactoring strategy" -ForegroundColor White
    Write-Host "  4. FORGE implements fixes" -ForegroundColor White
    Write-Host "  5. SENTINEL validates zero warnings" -ForegroundColor White
    Write-Host "  6. SOLACE approves completion" -ForegroundColor White
    
    Write-Host "`n⏱️  Expected completion: 30-45 minutes" -ForegroundColor Yellow
    Write-Host "✅ Agents will enforce ZERO TOLERANCE for warnings/errors`n" -ForegroundColor Green
    
} else {
    Write-Host "`n❌ Failed to create task" -ForegroundColor Red
    Write-Host "Error: $result" -ForegroundColor Red
    exit 1
}
