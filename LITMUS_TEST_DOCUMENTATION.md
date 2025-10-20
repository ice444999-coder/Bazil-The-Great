# 🧪 ARES Enhanced Litmus Test - Documentation

**Location:** `c:\ARES_Workspace\ARES_API\litmus_test_enhanced.py`  
**Quick Run:** `c:\ARES_Workspace\ARES_API\run_litmus_test.ps1`

---

## 🎯 What This Test Does

This comprehensive litmus test validates **ARES system wiring and functionality** through 5 critical stages:

### Stage 1: Code Wiring Check ⚙️
- Scans all Go files for wiring issues
- Patterns checked:
  - `undefined:` - Missing definitions
  - `missing import` - Import errors
  - `wiring` - Explicit wiring issues
  - `deviation` - System deviations
  - `panic` - Runtime panics
  - `error` - General errors
- Runs `go vet` for static analysis
- Runs `golangci-lint` (if installed)

### Stage 2: Build Validation 🔨
- **Go Mod Tidy:** Ensures dependencies are correct
- **Build Test:** Compiles `ares_api.exe` from `cmd/main.go`
- **Unit Tests:** Runs `go test ./...` for all packages

### Stage 3: Runtime Endpoint Testing 🌐
Tests these critical endpoints:
- `/` - Root endpoint
- `/dashboard.html` - Dashboard page
- `/trading.html` - Trading interface
- `/api/bazil/rewards` - Bazil rewards API
- `/api/bazil/findings` - Bazil findings API
- `/solace/ws` - SOLACE WebSocket

**Checks:**
- HTTP status code (expects 200)
- Response content (non-empty, no errors)
- Connection availability

### Stage 4: Log Monitoring 📝
Monitors server logs for 60 seconds looking for:

**Expected Success Patterns:**
- "System healthy"
- "Heal triggered"
- "Patch applied"
- "Verified—system healed"

**Error Patterns:**
- `error` / `err`
- `failed`
- `deviation`
- `mismatch`
- `undefined`
- `missing import`

### Stage 5: Fault Injection & Self-Heal 💉
- **Injects fault:** Adds unused import to `solace_agent.go`
- **Waits 5 seconds** for self-healing loop to detect
- **Verifies healing:** Checks logs for "healed" message
- **Cleanup:** Restores original file

---

## 🚀 How to Run

### Option 1: PowerShell Script (Easiest)
```powershell
cd c:\ARES_Workspace\ARES_API
.\run_litmus_test.ps1
```

### Option 2: Direct Python
```powershell
cd c:\ARES_Workspace\ARES_API
python litmus_test_enhanced.py
```

### Option 3: From VS Code Extension
Add this button to the extension (future enhancement):
```typescript
case 'litmusTest':
    terminal.sendText('cd c:\\ARES_Workspace\\ARES_API; python litmus_test_enhanced.py');
    break;
```

---

## 📋 Prerequisites

### Required:
- ✅ Python 3.x installed (`py --version`)
- ✅ `requests` module (`pip install requests`)
- ✅ Go toolchain (`go version`)
- ✅ ARES_API project at `C:\ARES_Workspace\ARES_API`

### Optional:
- `golangci-lint` (for enhanced static analysis)
- `grep` (test uses PowerShell Select-String as fallback)

---

## 📊 Expected Output

### ✅ Success (100% Wired):
```
Starting Enhanced Litmus Test for ARES Wiring...
Step 1: Scanning for wiring errors...
  (golangci-lint not installed, skipping)
Step 2: Tidy dependencies...
Building app...
Running unit/integration tests...
Step 3: Starting server...
Step 4: Monitoring logs...
Logs excerpt:
2025/10/20 19:45:22 🚀 ARES API Server starting...
2025/10/20 19:45:22 📊 Server listening on :8080
Step 5: Injecting fault and testing heal...
Fault injected: Unused import in solace_agent.go

ALL TESTS PASS - System wired 100% perfectly!
```

### ❌ Failure (Wiring Issues):
```
Starting Enhanced Litmus Test for ARES Wiring...
Step 1: Scanning for wiring errors...
...
ERRORS FOUND (Wiring/Function Issues):
- Endpoint /api/bazil/rewards failed: status 404
- Missing expected logs: Heal triggered, Patch applied
- Go test failures:
  --- FAIL: TestSolaceAgent (0.00s)
- Healing did not trigger or succeed

Fix these and rerun. Paste full output back for patches.
```

---

## 🔧 Configuration

### Customize Endpoints
Edit `ENDPOINTS` list in the script:
```python
ENDPOINTS = [
    "/",
    "/dashboard.html",
    "/trading.html",
    "/api/v1/health",          # Add custom endpoints
    "/api/v1/solace-agent/chat",
    # Add more as needed
]
```

### Customize Expected Logs
Edit `EXPECTED_LOGS` for your system:
```python
EXPECTED_LOGS = [
    "System healthy",
    "Heal triggered",
    "Patch applied",
    "Verified—system healed",
    "ARES initialized",        # Add custom patterns
]
```

### Customize Error Patterns
Edit `ERROR_PATTERNS` to catch specific issues:
```python
ERROR_PATTERNS = [
    r"err(or)?",
    "failed",
    "deviation",
    "panic",
    "nil pointer",            # Add custom patterns
    "connection refused",
]
```

---

## 🐛 Troubleshooting

### Test Hangs at "Starting server..."
**Cause:** Port 8080 already in use  
**Fix:** Kill existing process:
```powershell
$proc = Get-NetTCPConnection -LocalPort 8080 | Select -ExpandProperty OwningProcess
Stop-Process -Id $proc -Force
```

### "Module 'requests' not found"
**Fix:** Install requests:
```powershell
pip install requests
# or
python -m pip install requests
```

### "golangci-lint not installed"
**Not critical** - test continues without it.  
**Optional install:**
```powershell
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Fault Injection Fails
**Cause:** File path mismatch  
**Fix:** Check if `internal/solace/solace_agent.go` exists.  
Update line in script if file is elsewhere:
```python
solace_file = os.path.join(PROJECT_DIR, "internal", "solace", "solace_agent.go")
```

### All Endpoints Return 404
**Cause:** Server not fully started or routes not registered  
**Fix:** 
1. Increase startup wait time (line ~40):
   ```python
   time.sleep(15)  # Increase from 10 to 15
   ```
2. Check `cmd/main.go` has route registrations

---

## 📈 Interpreting Results

### Wiring Score
- **100%:** All tests pass - System perfectly wired ✅
- **80-99%:** Minor issues (logs missing, some endpoints 404) ⚠️
- **50-79%:** Moderate issues (build warnings, some tests fail) 🟡
- **0-49%:** Critical issues (build fails, major errors) 🔴

### Action Items by Score

**100%:** 🎉 Ship it!

**80-99%:**
- Review missing log patterns
- Check endpoint routing
- Verify test coverage

**50-79%:**
- Fix build warnings
- Review failed unit tests
- Check wiring patterns found

**0-49%:**
- Fix build errors immediately
- Review all wiring issues
- Do not deploy

---

## 🎯 Integration with ARES System

### Auto-Run on Commit (Future)
Add to `.git/hooks/pre-commit`:
```bash
#!/bin/bash
cd c:\ARES_Workspace\ARES_API
python litmus_test_enhanced.py > litmus_report.txt
if [ $? -ne 0 ]; then
    echo "Litmus test failed. Commit blocked."
    exit 1
fi
```

### CI/CD Integration
GitHub Actions workflow:
```yaml
name: ARES Litmus Test
on: [push, pull_request]
jobs:
  litmus:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run Litmus Test
        run: |
          cd ARES_API
          python litmus_test_enhanced.py
```

### VS Code Extension Button (Suggested)
```typescript
case 'runLitmus':
    const litmusTerminal = vscode.window.createTerminal({
        name: 'Litmus Test',
        cwd: 'c:\\ARES_Workspace\\ARES_API'
    });
    litmusTerminal.show();
    litmusTerminal.sendText('python litmus_test_enhanced.py');
    break;
```

---

## 📁 Files Created

| File | Purpose |
|------|---------|
| `litmus_test_enhanced.py` | Main test script |
| `run_litmus_test.ps1` | Quick launcher script |
| `LITMUS_TEST_DOCUMENTATION.md` | This file |

---

## 🔗 Related Documentation

- `STATUS_BUTTON_AUTO_RESTART.md` - Status button functionality
- `DEPENDENCY_FIX_COMPLETE.md` - Dependency resolution
- `ACE_TESTING_PROTOCOL.md` - Full testing framework

---

## 💡 Tips

1. **Run before every deploy** to catch wiring issues
2. **Run after major changes** to verify system integrity
3. **Save test output** to track system health over time
4. **Customize patterns** for your specific needs
5. **Integrate with CI/CD** for automated quality gates

---

**🧪 Ready to test? Run: `.\run_litmus_test.ps1`**
