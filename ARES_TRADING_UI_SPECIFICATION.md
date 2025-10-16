# ARES TRADING UI - FUNCTIONAL SPECIFICATION & TEST MANIFEST
**Version:** 2.0  
**Last Updated:** October 16, 2025  
**Purpose:** Blueprint for agent swarm to validate trading UI functionality

---

## 📋 OVERVIEW

This document defines EXACTLY what the ARES Trading UI should do, how each component should behave, and how to test if it works correctly.

**Agents:** Use this as the SOURCE OF TRUTH when designing, implementing, or testing UI components.

---

## 🎨 UI LAYOUT SPECIFICATION

### Main Trading Dashboard (`/web/trading.html`, `/`)

**Layout Structure:**
```
┌─────────────────────────────────────────────────────────────┐
│ HEADER: "ARES Trading Dashboard v2.9"                      │
│ - Logo (left)                                               │
│ - SOLACE Status Indicator (green dot = connected)          │
│ - Active Positions Count                                    │
│ - Total P&L (color: green if >0, red if <0)               │
├─────────────────────────────────────────────────────────────┤
│ LEFT PANEL (40% width)                                      │
│ ┌─────────────────────────────────────────────────────┐    │
│ │ MARKET SELECTOR                                      │    │
│ │ - Dropdown: BTC/USDT, ETH/USDT, SOL/USDT, etc.      │    │
│ │ - Current Price (updates every 2s)                   │    │
│ │ - 24h Change % (green/red)                           │    │
│ └─────────────────────────────────────────────────────┘    │
│ ┌─────────────────────────────────────────────────────┐    │
│ │ ORDER ENTRY FORM                                     │    │
│ │ - Order Type: Market / Limit (toggle buttons)        │    │
│ │ - Direction: BUY / SELL (green/red buttons)          │    │
│ │ - Amount (USDT): Input field                         │    │
│ │ - Limit Price: Input (only if Limit selected)        │    │
│ │ - Stop Loss %: Optional slider (0-10%)               │    │
│ │ - Take Profit %: Optional slider (0-50%)             │    │
│ │ - Risk Summary: Shows max loss/profit in $           │    │
│ │ - PLACE ORDER Button (disabled if invalid)           │    │
│ └─────────────────────────────────────────────────────┘    │
│ ┌─────────────────────────────────────────────────────┐    │
│ │ SOLACE ANALYSIS (if enabled)                         │    │
│ │ - Last Decision: "BUY" / "SELL" / "HOLD"             │    │
│ │ - Confidence: 0-100%                                  │    │
│ │ - Reasoning: Text explanation                        │    │
│ │ - Auto-Trade Toggle: ON/OFF                          │    │
│ └─────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│ RIGHT PANEL (60% width)                                     │
│ ┌─────────────────────────────────────────────────────┐    │
│ │ PRICE CHART (TradingView or Lightweight Charts)     │    │
│ │ - Timeframe selector: 1m, 5m, 15m, 1h, 4h, 1d       │    │
│ │ - Candlestick chart with volume                      │    │
│ │ - Entry markers (where you opened positions)         │    │
│ │ - SL/TP lines (if set)                               │    │
│ └─────────────────────────────────────────────────────┘    │
│ ┌─────────────────────────────────────────────────────┐    │
│ │ OPEN POSITIONS TABLE                                 │    │
│ │ Columns:                                             │    │
│ │ - ID | Pair | Direction | Size | Entry | Current    │    │
│ │ - P&L | P&L% | Opened At | Actions [Close]          │    │
│ │                                                       │    │
│ │ Empty State: "No open positions"                     │    │
│ └─────────────────────────────────────────────────────┘    │
│ ┌─────────────────────────────────────────────────────┐    │
│ │ TRADE HISTORY (Last 10)                              │    │
│ │ - Closed trades with final P&L                       │    │
│ │ - Color coded: Green (profit) / Red (loss)           │    │
│ │ - Click to view details                              │    │
│ └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔘 COMPONENT FUNCTIONAL SPECIFICATIONS

### 1. Market Selector Dropdown

**File:** `frontend/src/components/MarketSelector.tsx`

**Behavior:**
- ✅ MUST fetch available pairs from `/api/markets/available`
- ✅ MUST display current price that updates every 2 seconds
- ✅ MUST show 24h change % with color coding (green >0, red <0)
- ✅ MUST update chart when pair changes
- ✅ MUST clear order form when switching pairs

**Test Cases:**
```javascript
// TEST 1: Dropdown loads markets
EXPECT: API call to /api/markets/available
EXPECT: Dropdown populated with markets
VERIFY: At least BTC/USDT, ETH/USDT visible

// TEST 2: Price updates
EXPECT: Price changes within 5 seconds
VERIFY: New price !== old price

// TEST 3: Pair switch
ACTION: Select ETH/USDT
EXPECT: Chart shows ETH data
EXPECT: Order form shows "ETH/USDT"
```

---

### 2. Order Entry Form (AdvancedOrderForm)

**File:** `frontend/src/components/AdvancedOrderForm.tsx`

**Behavior - Market Order:**
- ✅ Amount field MUST accept numbers only
- ✅ MUST validate amount > 0 and < account balance
- ✅ BUY button MUST be GREEN, SELL button RED
- ✅ Clicking PLACE ORDER MUST:
  1. Show confirmation modal with order details
  2. POST to `/api/trades/open` with:
     ```json
     {
       "symbol": "BTCUSDT",
       "direction": "BUY",
       "amount": 100,
       "orderType": "market",
       "stopLoss": 0.05,
       "takeProfit": 0.10
     }
     ```
  3. Display loading state (button disabled, spinner)
  4. On success: Show success toast, clear form, update positions table
  5. On error: Show error toast with message
- ✅ MUST show estimated cost in USDT
- ✅ MUST calculate and show max loss/profit if SL/TP set

**Behavior - Limit Order:**
- ✅ MUST show "Limit Price" input field
- ✅ MUST validate limit price > 0
- ✅ For BUY: Warn if limit price > current price (will execute immediately)
- ✅ For SELL: Warn if limit price < current price (will execute immediately)

**Behavior - Stop Loss / Take Profit:**
- ✅ Sliders MUST update % value in real-time
- ✅ MUST calculate price based on entry (current or limit)
- ✅ For LONG: SL < entry, TP > entry
- ✅ For SHORT: SL > entry, TP < entry
- ✅ MUST show risk/reward ratio if both SL and TP set

**Test Cases:**
```javascript
// TEST 1: Market Buy Order
ACTION: Select BUY, Amount: 100, Click PLACE ORDER
EXPECT: POST to /api/trades/open
EXPECT: Response status 200
EXPECT: positions table updates with new row
VERIFY: Position shows "BUY" direction

// TEST 2: Validation - Invalid Amount
ACTION: Enter amount: -50
EXPECT: Error message "Amount must be positive"
EXPECT: PLACE ORDER button disabled

// TEST 3: Validation - Insufficient Balance
ACTION: Enter amount: 999999
EXPECT: Error message "Insufficient balance"
EXPECT: PLACE ORDER button disabled

// TEST 4: Stop Loss Calculation
ACTION: BUY, Entry: $50000, SL: 5%
EXPECT: SL Price = $47500
EXPECT: Max Loss = -$2500 (if 100 USDT position)

// TEST 5: Risk/Reward Display
ACTION: Set SL: 5%, TP: 10%
EXPECT: Risk/Reward Ratio = 1:2
EXPECT: Green highlight if ratio > 1:1
```

---

### 3. Open Positions Table

**File:** `frontend/src/components/OpenPositionsTable.tsx`

**Behavior:**
- ✅ MUST fetch positions from `/api/trades/positions` on load
- ✅ MUST update every 5 seconds with latest prices
- ✅ MUST calculate unrealized P&L based on current price
- ✅ P&L column color: GREEN if profit, RED if loss
- ✅ Close button MUST:
  1. Show confirmation modal "Close position #123?"
  2. POST to `/api/trades/close` with position ID
  3. On success: Remove row, update total P&L
  4. On error: Show error toast

**Calculations:**
```javascript
// For LONG position:
unrealizedPnL = (currentPrice - entryPrice) * size - fees

// For SHORT position:
unrealizedPnL = (entryPrice - currentPrice) * size - fees

// P&L Percentage:
pnlPercent = (unrealizedPnL / (entryPrice * size)) * 100
```

**Test Cases:**
```javascript
// TEST 1: Position Display
EXPECT: All open positions visible
VERIFY: P&L updates within 10 seconds
VERIFY: P&L calculation matches manual calculation

// TEST 2: Close Position
ACTION: Click Close on position #5
EXPECT: Confirmation modal appears
ACTION: Confirm
EXPECT: POST to /api/trades/close with id=5
EXPECT: Position removed from table
VERIFY: Total P&L header updates

// TEST 3: Empty State
GIVEN: No open positions
EXPECT: Message "No open positions"
EXPECT: No table visible
```

---

### 4. Price Chart

**File:** `frontend/src/components/TradingChart.tsx`

**Behavior:**
- ✅ MUST fetch OHLCV data from `/api/markets/ohlcv?symbol=BTCUSDT&interval=1h`
- ✅ MUST show candlestick chart with volume bars
- ✅ MUST add entry markers when position opened (green arrow up for BUY, red arrow down for SELL)
- ✅ MUST draw horizontal lines for SL/TP if set
- ✅ MUST update in real-time (append new candle every interval)
- ✅ Timeframe buttons MUST re-fetch data with new interval

**Test Cases:**
```javascript
// TEST 1: Chart Loads
EXPECT: API call to /api/markets/ohlcv
EXPECT: At least 100 candles visible
VERIFY: Candles have correct OHLC structure

// TEST 2: Real-time Update
WAIT: 60 seconds
EXPECT: New candle added
VERIFY: Last candle !== previous last candle

// TEST 3: Entry Markers
GIVEN: Open BUY position at $50000
EXPECT: Green up arrow at $50000 on chart
VERIFY: Marker timestamp matches position opened_at

// TEST 4: SL/TP Lines
GIVEN: Position with SL=$47500, TP=$55000
EXPECT: Red horizontal line at $47500
EXPECT: Green horizontal line at $55000
```

---

### 5. SOLACE AI Integration

**File:** `frontend/src/pages/SOLACEConsciousnessTrading.tsx`

**Behavior:**
- ✅ MUST show SOLACE's last decision (BUY/SELL/HOLD)
- ✅ MUST display confidence score 0-100%
- ✅ MUST show reasoning text (why SOLACE made decision)
- ✅ Auto-Trade toggle MUST:
  - When ON: Automatically execute SOLACE's recommendations
  - When OFF: Only display recommendations
- ✅ MUST log all SOLACE decisions to PostgreSQL `solace_decisions` table

**Test Cases:**
```javascript
// TEST 1: Decision Display
EXPECT: Last decision visible (within 60s)
VERIFY: Decision is BUY, SELL, or HOLD
VERIFY: Confidence is 0-100

// TEST 2: Auto-Trade OFF
GIVEN: Auto-Trade = OFF
WHEN: SOLACE decides BUY
EXPECT: Decision displayed
EXPECT: NO automatic trade executed
VERIFY: positions table unchanged

// TEST 3: Auto-Trade ON
GIVEN: Auto-Trade = ON
WHEN: SOLACE decides BUY with 85% confidence
EXPECT: Automatic market order placed
VERIFY: New position appears in table
VERIFY: Order size based on risk management rules
```

---

## 🧪 COMPREHENSIVE TEST SUITE

### Critical Path Tests (MUST PASS)

```javascript
// CRITICAL TEST 1: End-to-End Buy Order
1. Open trading UI
2. Select BTC/USDT
3. Wait for price to load
4. Enter amount: 100 USDT
5. Click BUY
6. Click CONFIRM in modal
7. VERIFY: Success toast appears
8. VERIFY: New position in table
9. VERIFY: Position has correct: pair, direction, size
10. VERIFY: PostgreSQL has new row in trades table

// CRITICAL TEST 2: Close Position with Profit
1. Open position exists (BUY BTC at $50000)
2. Wait for price to rise to $51000
3. Click CLOSE on position
4. Confirm close
5. VERIFY: Position removed from table
6. VERIFY: P&L shown in trade history
7. VERIFY: P&L is GREEN (profit)
8. VERIFY: Total P&L header updated

// CRITICAL TEST 3: Stop Loss Triggered
1. Open BUY position at $50000 with SL=5%
2. Simulate price drop to $47400
3. VERIFY: Position auto-closed
4. VERIFY: Close reason = "Stop Loss"
5. VERIFY: Loss matches expected 5%

// CRITICAL TEST 4: Real-time Updates
1. Open position
2. Wait 10 seconds
3. VERIFY: Position P&L changed
4. VERIFY: Chart updated with new candle
5. VERIFY: Header P&L total updated
```

### Edge Case Tests

```javascript
// EDGE TEST 1: API Error Handling
SIMULATE: API returns 500 error on /api/trades/open
EXPECT: Error toast "Failed to place order"
EXPECT: Form NOT cleared
EXPECT: Order button re-enabled

// EDGE TEST 2: Websocket Disconnect
SIMULATE: Kill websocket connection
EXPECT: Warning "Connection lost"
EXPECT: Auto-reconnect attempt
VERIFY: Reconnects within 10 seconds

// EDGE TEST 3: Concurrent Orders
ACTION: Click PLACE ORDER twice rapidly
EXPECT: Only 1 order created
VERIFY: No duplicate positions
```

---

## 📊 PERFORMANCE BENCHMARKS

| Metric | Target | Critical Threshold |
|--------|--------|-------------------|
| Page Load Time | < 2s | < 5s |
| API Response (open trade) | < 500ms | < 2s |
| API Response (close trade) | < 300ms | < 1s |
| Price Update Frequency | Every 2s | Every 5s |
| Chart Render Time | < 1s | < 3s |
| WebSocket Latency | < 100ms | < 500ms |

---

## 🎯 SENTINEL AGENT TEST INSTRUCTIONS

**SENTINEL: When testing UI, you MUST:**

1. **Load this specification file FIRST**
2. **For each component, verify:**
   - ✅ Layout matches specification
   - ✅ All required fields/buttons present
   - ✅ API calls go to correct endpoints
   - ✅ Data structure matches expected format
   - ✅ Error handling works
   - ✅ Success cases work
   - ✅ Edge cases handled

3. **Use Playwright to:**
   ```python
   # Example test script
   page.goto("http://localhost:3000")
   page.fill("#amount", "100")
   page.click("button:has-text('BUY')")
   page.click("button:has-text('Confirm')")
   
   # Verify success
   toast = page.locator(".toast-success")
   assert toast.is_visible()
   assert "Order placed" in toast.text_content()
   
   # Verify position created
   positions = page.locator(".position-row")
   assert positions.count() > 0
   ```

4. **Generate test report with:**
   - ✅/❌ for each test case
   - Screenshots of failures
   - API call logs
   - Console errors
   - Performance metrics

---

## 🚨 FAILURE CRITERIA

**Report CRITICAL BUG if:**
- ❌ Order placed but no position created
- ❌ Close button doesn't close position
- ❌ P&L calculation wrong by >1%
- ❌ SL/TP not triggered when price reached
- ❌ API error not shown to user
- ❌ Money lost due to UI bug

**Report WARNING if:**
- ⚠️ Layout doesn't match specification
- ⚠️ Update slower than 5 seconds
- ⚠️ Missing error message
- ⚠️ Inconsistent styling

---

## 📝 AGENT USAGE EXAMPLES

### ARCHITECT Example:
```
I need to design a new Risk Management Panel.

REFERENCE: This specification, section "Component Functional Specifications"
PATTERN: Follow same structure as Order Entry Form
API ENDPOINT: Define /api/risk/settings
VALIDATION: Must calculate position sizing based on account risk %
TEST CASES: Define at least 5 test cases like above examples
```

### FORGE Example:
```
Implementing TradingPerformanceCard component.

REFERENCE: ARES_TRADING_UI_SPECIFICATION.md
MUST INCLUDE:
- Total P&L with color coding (green/red)
- Win rate calculation
- Best/worst trade display
- API fetch from /api/trades/stats
- Update every 5 seconds
- Match layout specification format
```

### SENTINEL Example:
```
Testing AdvancedOrderForm component.

LOAD: ARES_TRADING_UI_SPECIFICATION.md section 2
RUN: All 5 test cases defined
VERIFY: Each behavior checkbox (✅)
REPORT: Pass/Fail for each test
CAPTURE: Screenshots if failed
LOG: All API calls made
CHECK: Console for errors
```

---

**END OF SPECIFICATION**

**Agents: Bookmark this file. Reference it for EVERY UI task.**
