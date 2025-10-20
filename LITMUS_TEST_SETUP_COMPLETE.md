# ✅ Enhanced Litmus Test - Setup Complete

**Date:** October 20, 2025  
**Status:** 🎉 READY TO USE

---

## 📦 What Was Created

### 1. Main Test Script
**File:** `c:\ARES_Workspace\ARES_API\litmus_test_enhanced.py`  
**Size:** ~200 lines  
**Purpose:** Comprehensive ARES system validation

**Features:**
- ✅ Code wiring checks (PowerShell-based, Windows-compatible)
- ✅ Build validation (go mod tidy, build, test)
- ✅ Runtime endpoint testing (6 critical endpoints)
- ✅ Log monitoring (60 seconds, pattern matching)
- ✅ Fault injection & self-heal verification

**Fixed for Windows:**
- Changed `grep` → PowerShell `Select-String`
- Changed `taskkill` → PowerShell `Stop-Process` with port check
- Changed `cmd/ares/main.go` → `cmd/main.go` (correct path)
- Changed hardcoded Unix paths → Windows paths
- Added UTF-8 encoding for file operations

### 2. Quick Launcher
**File:** `c:\ARES_Workspace\ARES_API\run_litmus_test.ps1`  
**Purpose:** One-click test execution

**Does:**
- Checks Python installed
- Installs `requests` module if missing
- Runs litmus test
- Shows results

### 3. Complete Documentation
**File:** `c:\ARES_Workspace\ARES_API\LITMUS_TEST_DOCUMENTATION.md`  
**Contents:**
- What the test does (5 stages explained)
- How to run it (3 methods)
- Prerequisites
- Expected output
- Configuration options
- Troubleshooting guide
- Integration tips

---

## 🚀 How to Use

### Quick Start (Easiest)
```powershell
cd c:\ARES_Workspace\ARES_API
.\run_litmus_test.ps1
```

### Direct Python
```powershell
cd c:\ARES_Workspace\ARES_API
python litmus_test_enhanced.py
```

### From Terminal
Just type the command - the script handles everything!

---

## 🧪 What Gets Tested

### 5-Stage Validation Process

**Stage 1: Code Wiring** ⚙️
- Scans for: undefined, missing import, wiring issues, panics, errors
- Runs: go vet, golangci-lint (optional)

**Stage 2: Build** 🔨
- go mod tidy
- go build cmd/main.go
- go test ./...

**Stage 3: Endpoints** 🌐
- Tests 6 critical URLs
- Verifies HTTP 200 responses
- Checks content validity

**Stage 4: Logs** 📝
- 60-second monitoring
- Checks for success patterns
- Detects error patterns

**Stage 5: Self-Heal** 💉
- Injects fault (unused import)
- Waits for auto-heal
- Verifies healing occurred

---

## 📊 Test Results

### Success Output
```
Starting Enhanced Litmus Test for ARES Wiring...
Step 1: Scanning for wiring errors...
Step 2: Tidy dependencies...
Building app...
Running unit/integration tests...
Step 3: Starting server...
Step 4: Monitoring logs...
Step 5: Injecting fault and testing heal...

ALL TESTS PASS - System wired 100% perfectly!
```

### Failure Output
```
ERRORS FOUND (Wiring/Function Issues):
- Endpoint /api/bazil/rewards failed: status 404
- Missing expected logs: Heal triggered
- Go test failures: TestSolaceAgent
- Healing did not trigger or succeed

Fix these and rerun. Paste full output back for patches.
```

---

## 🎯 When to Run

### Must Run
- ✅ Before every deployment
- ✅ After major code changes
- ✅ After wiring modifications
- ✅ When debugging system issues

### Should Run
- ⚙️ After dependency updates
- ⚙️ Weekly health checks
- ⚙️ Before merging PRs

### Can Run
- 🔄 Daily CI/CD pipeline
- 🔄 On-demand diagnostics
- 🔄 Performance benchmarking

---

## 🔧 Customization

### Add More Endpoints
Edit line ~7 in `litmus_test_enhanced.py`:
```python
ENDPOINTS = [
    "/",
    "/dashboard.html",
    "/api/v1/your-new-endpoint",  # Add here
]
```

### Change Expected Logs
Edit line ~9:
```python
EXPECTED_LOGS = [
    "System healthy",
    "Your custom log pattern",  # Add here
]
```

### Adjust Timeouts
Edit line ~40 (server startup):
```python
time.sleep(15)  # Increase if needed
```

Edit line ~122 (log monitoring):
```python
log_errors, logs = check_logs(proc, duration=120)  # Increase duration
```

---

## ⚠️ Prerequisites

### Required
- [x] Python 3.x (`py --version` to check)
- [x] Go toolchain (`go version` to check)
- [x] ARES_API project at correct path

### Auto-Installed
- [ ] `requests` module (script installs if missing)

### Optional
- [ ] `golangci-lint` (enhanced static analysis)

---

## 🐛 Common Issues

### "Python not found"
```powershell
# Check if installed
py --version

# If not, download from python.org
```

### "Module 'requests' not found"
```powershell
pip install requests
# or
python -m pip install requests
```

### "Port 8080 in use"
```powershell
# Kill process on port 8080
$proc = Get-NetTCPConnection -LocalPort 8080 | Select -ExpandProperty OwningProcess
Stop-Process -Id $proc -Force
```

### Test hangs at "Starting server"
- Check if API is already running
- Increase startup wait time in script
- Verify `cmd/main.go` exists and compiles

---

## 📈 Integration Ideas

### Add to VS Code Extension
```typescript
case 'litmusTest':
    const terminal = vscode.window.createTerminal({
        name: 'Litmus Test',
        cwd: 'c:\\ARES_Workspace\\ARES_API'
    });
    terminal.show();
    terminal.sendText('python litmus_test_enhanced.py');
    break;
```

### Add Button to Extension UI
```html
<button id="btnLitmus">🧪 Litmus Test</button>
```

### Automated Daily Run
Create scheduled task:
```powershell
$action = New-ScheduledTaskAction -Execute 'PowerShell.exe' -Argument '-File c:\ARES_Workspace\ARES_API\run_litmus_test.ps1'
$trigger = New-ScheduledTaskTrigger -Daily -At 9am
Register-ScheduledTask -Action $action -Trigger $trigger -TaskName "ARES Litmus Test"
```

---

## ✅ Checklist

Setup:
- [x] Created `litmus_test_enhanced.py`
- [x] Created `run_litmus_test.ps1`
- [x] Created `LITMUS_TEST_DOCUMENTATION.md`
- [x] Fixed Windows compatibility issues
- [x] Fixed file paths (cmd/main.go)
- [x] Added PowerShell commands (not grep/taskkill)

Ready to Test:
- [ ] Run `.\run_litmus_test.ps1`
- [ ] Verify Python and requests installed
- [ ] Check test passes all 5 stages
- [ ] Review any errors found
- [ ] Fix issues and rerun

---

## 🎉 Summary

You now have a **production-ready litmus test** that:

✅ **Validates** 5 critical system areas  
✅ **Works on Windows** with PowerShell commands  
✅ **Auto-installs** dependencies  
✅ **Provides clear output** (pass/fail with details)  
✅ **Tests self-healing** with fault injection  
✅ **Documents everything** comprehensively  

---

**🧪 Ready to test? Run: `.\run_litmus_test.ps1`**

The URL you asked about (**http://localhost:8080**) is tested automatically by this script across 6 different endpoints!
