# 🚀 QUICK START - UI Testing

## Run Tests (3 Commands)

```powershell
# 1. One-time setup (install Playwright browser)
cd C:\ARES_Workspace\ARES_API
python -m playwright install chromium

# 2. Start ARES (if not running)
.\ARES_API.exe

# 3. Run tests
python sentinel_ui_test.py
```

**Duration:** 5-10 minutes  
**Output:** Reports + Screenshots in `tests/` folder

---

## Check Results

```powershell
# View human-readable report
notepad tests\UI_TEST_REPORT.md

# View developer fix log
notepad tests\UI_FIX_LOG.md

# View screenshots
explorer tests\ui_screenshots
```

---

## What Gets Tested

- ✅ All 18 HTML pages
- ✅ Every button clicked (~160 buttons)
- ✅ Scroll behavior (max height, overflow)
- ✅ Window dimensions (viewport vs content)
- ✅ Screenshots (before/after every action)
- ✅ Responsive design (4 viewport sizes)

---

## Pass/Fail Criteria

| Pass Rate | Status |
|-----------|--------|
| 100% | ✅ Perfect - Deploy! |
| 90-99% | ⚠️ Minor issues - OK to deploy |
| 70-89% | ⚠️ Fix before deploy |
| <70% | ❌ Critical issues - DO NOT DEPLOY |

---

## Files Created

1. **`sentinel_ui_test.py`** (620 lines) - Main script
2. **`RUN_UI_TESTS.md`** (350 lines) - Full guide
3. **`UI_TESTING_SYSTEM_COMPLETE.md`** - Deployment summary

---

## Help

**Tests too slow?**  
Edit `sentinel_ui_test.py` line 357: `headless=True`

**Server not running?**  
`.\ARES_API.exe` in another terminal

**Need details?**  
Read `RUN_UI_TESTS.md` for complete guide

---

**That's it! Start testing now:** `python sentinel_ui_test.py` 🎉
