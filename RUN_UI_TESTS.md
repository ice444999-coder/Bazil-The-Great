# 🧪 How to Run SENTINEL UI Tests

**Complete Playwright testing system that clicks every button, tests scroll, and produces fix logs.**

---

## 🚀 Quick Start (3 Steps)

### Step 1: Make sure ARES is running on port 8080

```powershell
# Check if ARES_API.exe is running
Get-Process | Where-Object { $_.ProcessName -like "*ARES*" }

# Or start it
cd C:\ARES_Workspace\ARES_API
.\ARES_API.exe
```

**Verify:** Open http://localhost:8080 in browser - should see ARES UI

---

### Step 2: Install Playwright browsers (ONE TIME ONLY)

```powershell
cd C:\ARES_Workspace\ARES_API
python -m playwright install chromium
```

This downloads Chromium browser for testing (~100MB, takes 1-2 minutes).

---

### Step 3: Run the tests

```powershell
cd C:\ARES_Workspace\ARES_API
python sentinel_ui_test.py
```

**You'll see:**
- 🌐 Browser window opens (Chromium)
- 📄 Each page loads automatically
- 🔘 Every button clicks
- 📸 Screenshots taken
- 📊 Real-time progress in terminal

**Test Duration:** ~5-10 minutes for all 18 pages

---

## 📊 What Gets Tested

For **each of 18 HTML pages:**

1. ✅ **Page Load** - HTTP status, navigation success
2. ✅ **Window Dimensions** - Viewport vs content size
3. ✅ **Scroll Behavior** - Tests `scrollTo()`, measures max scroll
4. ✅ **Form Elements** - Counts inputs, textareas, selects, forms
5. ✅ **Button Discovery** - Finds all `<button>`, `input[type=button]`, `[role=button]`
6. ✅ **Button Clicks** - Clicks every visible, enabled button
7. ✅ **Screenshots** - Before/after every action + full page
8. ✅ **Error Detection** - JS errors, broken links, timeout issues

**Responsive Design Tests** (3 key pages):
- 🖥️ Desktop (1920x1080)
- 💻 Laptop (1366x768)
- 📱 Tablet (768x1024)
- 📱 Mobile (375x667)

---

## 📁 Output Files

After tests complete, check:

### 1. `tests/UI_TEST_REPORT.md` (Human-Readable)
```markdown
# SENTINEL UI TEST REPORT

## 📊 Executive Summary
- Pages Tested: 18
- Pages Passed: 17 (94.4%)
- Buttons Clicked: 156/163 (95.7% success)
- Errors: 3
- Warnings: 8

### ❌ VERDICT: CRITICAL ISSUES FOUND

---

## 📄 Detailed Results

### ✅ trading.html
**Metrics:**
- Buttons: 12/12 clicked
- Scroll Works: ✅ Yes
- Content: 1920x3450px
```

### 2. `tests/UI_TEST_RESULTS.json` (Machine-Readable)
```json
{
  "generated_at": "2025-10-19T15:30:00",
  "summary": {
    "total_pages": 18,
    "pages_passed": 17,
    "total_buttons": 163,
    "buttons_clicked": 156
  },
  "results": [...]
}
```

### 3. `tests/UI_FIX_LOG.md` (Developer Action Items)
```markdown
# UI FIX LOG - Developer Action Items

## ❌ CRITICAL ERRORS (3)

### 1. [button_click] trading.html
**Button:** Close Position
**Problem:** Endpoint /api/trades/close returned 404

**How to Fix:**
- Verify button click handler is attached
- Check for JavaScript errors in console
- Ensure API endpoint exists in backend
```

### 4. `tests/ui_screenshots/` (All Screenshots)
```
trading.html_initial.png
trading.html_scrolled.png
trading.html_btn_0_before.png
trading.html_btn_0_after.png
trading.html_desktop.png
trading.html_mobile.png
... (100+ screenshots)
```

### 5. PostgreSQL Database
```sql
SELECT * FROM test_activity_logs 
WHERE test_type = 'ui_automation_playwright'
ORDER BY tested_at DESC;
```

---

## 🔧 Advanced Usage

### Test Specific Pages Only

Edit `sentinel_ui_test.py` line 18:

```python
PAGES_TO_TEST = [
    {"name": "trading.html", "description": "Main Trading Dashboard"},
    {"name": "solace-chat.html", "description": "SOLACE Chat Interface"}
    # Comment out pages you don't want to test
]
```

### Run Headless (No Browser Window)

Edit line 357:

```python
browser = p.chromium.launch(
    headless=True,  # Change False to True
    slow_mo=0  # Remove delay
)
```

### Speed Up Tests

Edit line 358:

```python
slow_mo=0  # Change from 100 to 0 (no delay)
```

### Take Fewer Screenshots

Edit line 184 (comment out):

```python
# button_screenshot = SCREENSHOT_DIR / f"{page_name.replace('.html', '')}_btn_{i}_before.png"
# button.screenshot(path=str(button_screenshot), timeout=2000)
```

---

## 🐛 Troubleshooting

### "playwright: command not found"
```powershell
cd C:\ARES_Workspace\ARES_API
python -m playwright install
```

### "Could not connect to localhost:8080"
Make sure ARES_API.exe is running:
```powershell
cd C:\ARES_Workspace\ARES_API
.\ARES_API.exe
```

### "Database connection failed"
Check PostgreSQL is running:
```powershell
Get-Service postgresql-x64-16
```

If not running:
```powershell
Start-Service postgresql-x64-16
```

### Tests are too slow
Edit `sentinel_ui_test.py`:
- Set `headless=True` (line 357)
- Set `slow_mo=0` (line 358)
- Reduce `page.wait_for_timeout()` values

### Browser crashes
Increase timeouts:
- Line 88: `timeout=10000` → `timeout=30000`
- Line 195: `timeout=3000` → `timeout=10000`

---

## 📋 What to Do After Tests

1. **Check Exit Status**
   - `Pages passed: 18/18` = ✅ Perfect!
   - `Pages passed: 15/18` = ⚠️ Review failures
   - `Pages passed: <10/18` = ❌ Major issues

2. **Review Fix Log**
   ```powershell
   notepad C:\ARES_Workspace\ARES_API\tests\UI_FIX_LOG.md
   ```

3. **Compare Screenshots**
   - Open `tests/ui_screenshots/` folder
   - Compare `*_before.png` vs `*_after.png`
   - Check if button clicks had visible effect

4. **Fix Issues**
   - Fix errors in order: CRITICAL → WARNINGS
   - Re-run tests after each fix
   - Compare new screenshots with baseline

5. **Update Baseline**
   If UI changes are intentional:
   ```powershell
   # Replace reference images
   Copy-Item "tests\ui_screenshots\trading.html_initial.png" "tests\ui_reference\trading_reference.png"
   ```

---

## 🔄 Automated Testing (CI/CD)

### Run on Every Commit

Create `.github/workflows/ui-tests.yml`:
```yaml
name: UI Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install Python
        uses: actions/setup-python@v2
      - name: Install Playwright
        run: python -m playwright install chromium
      - name: Run ARES
        run: Start-Process .\ARES_API.exe
      - name: Run UI Tests
        run: python sentinel_ui_test.py
      - name: Upload Screenshots
        uses: actions/upload-artifact@v2
        with:
          name: ui-screenshots
          path: tests/ui_screenshots/
```

### Run Every Hour (Task Scheduler)

```powershell
$action = New-ScheduledTaskAction -Execute "python" -Argument "C:\ARES_Workspace\ARES_API\sentinel_ui_test.py" -WorkingDirectory "C:\ARES_Workspace\ARES_API"
$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date) -RepetitionInterval (New-TimeSpan -Hours 1)
Register-ScheduledTask -Action $action -Trigger $trigger -TaskName "ARES_UI_Tests" -Description "Automated UI testing every hour"
```

---

## 📊 Interpreting Results

### Pass Rate Thresholds

| Pass Rate | Status | Action |
|-----------|--------|--------|
| 100% | ✅ Perfect | Deploy to production |
| 90-99% | ⚠️ Good | Review warnings, safe to deploy |
| 70-89% | ⚠️ Issues | Fix errors before deployment |
| <70% | ❌ Critical | Do NOT deploy, major bugs |

### Button Success Rate

| Success Rate | Status | Meaning |
|--------------|--------|---------|
| 100% | ✅ Perfect | All buttons work |
| 90-99% | ⚠️ Minor | Some disabled/hidden buttons OK |
| 70-89% | ⚠️ Issues | Many buttons broken |
| <70% | ❌ Critical | UI fundamentally broken |

### Common Error Patterns

**"HTTP 404"** → API endpoint missing or wrong URL  
**"Timeout"** → Button click handler not responding  
**"Element not visible"** → CSS display/visibility issue  
**"Click intercepted"** → Modal or overlay blocking button  
**"Navigation failed"** → Page doesn't exist or server down  

---

## 🎯 Next Steps

1. ✅ Run tests for first time: `python sentinel_ui_test.py`
2. ✅ Review `UI_FIX_LOG.md` for issues
3. ✅ Fix critical errors first
4. ✅ Re-run tests to verify fixes
5. ✅ Set up automated testing (hourly/daily)
6. ✅ Compare screenshots before/after changes
7. ✅ Keep reference images up to date

---

## 🆘 Support

If tests fail or you need help:

1. Check `UI_FIX_LOG.md` for specific error messages
2. Review screenshots in `tests/ui_screenshots/`
3. Check PostgreSQL for historical test data
4. Run Go API tests: `curl http://localhost:8080/api/v1/ui-test/all`
5. Ask SOLACE: "Why did button X fail on page Y?"

---

**Congratulations! You now have enterprise-grade UI testing.** 🎉
