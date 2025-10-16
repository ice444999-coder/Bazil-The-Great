"""
Create UI validation task for SENTINEL agent using the specification.
"""
import sys
sys.path.insert(0, 'C:\\ARES_Workspace\\ARES_API\\Lib\\site-packages')

import psycopg2
import json
from datetime import datetime

def create_ui_validation_task():
    conn = psycopg2.connect('host=localhost dbname=ares_db user=ARES password=ARESISWAKING')
    cur = conn.cursor()
    
    task_description = """UI FUNCTIONAL VALIDATION - Trading Dashboard Complete Test

OBJECTIVE: Verify the ARES Trading UI actually WORKS for real trading.

REFERENCE DOCUMENT: ARES_TRADING_UI_SPECIFICATION.md
This file defines EXACTLY what should work and how to test it.

YOUR TASK (SENTINEL):
1. Read ARES_TRADING_UI_SPECIFICATION.md completely
2. Test each component against the specification
3. Use Playwright to automate browser testing
4. Verify ALL critical path tests pass
5. Report any failures with screenshots and logs

CRITICAL PATH TESTS (MUST PASS):
✅ TEST 1: End-to-End Buy Order
   - Navigate to http://localhost:3000
   - Select BTC/USDT
   - Enter amount: 100 USDT
   - Click BUY → Confirm
   - Verify: Success toast, new position in table, PostgreSQL row created

✅ TEST 2: Position P&L Calculation
   - Verify open positions show correct unrealized P&L
   - Check formula: (currentPrice - entryPrice) * size - fees
   - Confirm P&L updates every 5 seconds
   - Verify color: GREEN if profit, RED if loss

✅ TEST 3: Close Position
   - Click CLOSE on existing position
   - Confirm closure
   - Verify: Position removed, P&L shown in history, total P&L updated

✅ TEST 4: API Endpoints Working
   - GET /api/markets/available → Returns markets
   - GET /api/trades/positions → Returns positions
   - POST /api/trades/open → Creates position
   - POST /api/trades/close → Closes position
   - GET /api/markets/ohlcv → Returns chart data

✅ TEST 5: Error Handling
   - Test invalid amount (negative, zero)
   - Test insufficient balance
   - Test API error response
   - Verify error messages shown to user

VALIDATION CHECKLIST:
□ All buttons work (BUY, SELL, CLOSE, etc.)
□ Forms validate input correctly
□ API calls go to correct endpoints
□ Success/error messages display
□ Real-time updates working (price, P&L, chart)
□ Layout matches specification
□ No console errors
□ No broken functionality

REPORT FORMAT:
{
  "tests_run": 15,
  "tests_passed": 13,
  "tests_failed": 2,
  "critical_bugs": [
    "Close button doesn't work - HTTP 404 on /api/trades/close"
  ],
  "warnings": [
    "Price update slower than spec (7s vs 5s target)"
  ],
  "screenshots": ["error_close_button.png"],
  "api_logs": [...],
  "verdict": "FAIL - Critical bug prevents position closure"
}

SUCCESS CRITERIA:
- All 5 critical path tests PASS
- No critical bugs
- Performance within thresholds
- All API endpoints return 200

FAILURE = Any critical path test fails OR critical bug found

TOOLS AVAILABLE:
- Playwright (already installed)
- PostgreSQL access to verify data
- Screenshot capture
- Network request logging

START BY:
1. Read ARES_TRADING_UI_SPECIFICATION.md
2. Write Playwright test script
3. Run tests against http://localhost:3000
4. Generate detailed report
"""

    cur.execute("""
        INSERT INTO task_queue (
            task_id, task_type, description, priority, status, 
            context, file_paths, created_at, assigned_to_agent
        ) VALUES (
            gen_random_uuid(), 
            'ui_testing',
            %s,
            10,
            'assigned',
            %s,
            %s,
            %s,
            'SENTINEL'
        ) RETURNING task_id
    """, (
        task_description,
        json.dumps({
            "test_type": "ui_functional_validation",
            "specification": "ARES_TRADING_UI_SPECIFICATION.md",
            "critical_tests": 5,
            "tools": ["playwright", "postgresql"],
            "target_url": "http://localhost:3000",
            "must_verify": [
                "buy_order_works",
                "pnl_calculation_correct",
                "position_closes",
                "apis_respond",
                "errors_handled"
            ]
        }),
        json.dumps([
            "ARES_TRADING_UI_SPECIFICATION.md",
            "frontend/src/components/AdvancedOrderForm.tsx",
            "frontend/src/components/OpenPositionsTable.tsx",
            "web/trading.html"
        ]),
        datetime.now()
    ))
    
    task_id = cur.fetchone()[0]
    conn.commit()
    
    print(f"\n{'='*70}")
    print(f"✓ UI VALIDATION TASK CREATED FOR SENTINEL")
    print(f"{'='*70}")
    print(f"Task ID: {task_id}")
    print(f"Assigned to: SENTINEL (Testing Agent)")
    print(f"\nSENTINEL will:")
    print(f"  1. Read UI specification document")
    print(f"  2. Test all critical paths with Playwright")
    print(f"  3. Verify actual trading functionality")
    print(f"  4. Check if buttons/forms/APIs actually work")
    print(f"  5. Report bugs with screenshots and evidence")
    print(f"\nThis answers: 'Do the agents know what should work?'")
    print(f"Answer: NOW THEY DO! Specification = source of truth.")
    print(f"{'='*70}\n")
    
    conn.close()
    return task_id

if __name__ == '__main__':
    create_ui_validation_task()
