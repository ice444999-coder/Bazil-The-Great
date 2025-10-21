# ARES Trading Tab Upgrade Control System

## Safety Features Implemented

### 1. Human Control Points
- **DRY_RUN Mode**: Set `$env:DRY_RUN="true"` to preview changes without applying
- **Manual Approval**: Each subtask requires approval via `/api/approve/{subtask_name}`
- **Emergency Stop**: Kill ARES API process to halt all operations
- **Git Branching**: Each subtask on separate branch for safe rollback

### 2. Automated Safety Guards
- **Max Attempts**: 3 retries with exponential backoff (1min, 2min, 4min)
- **Auto Rollback**: `git revert HEAD` on any error
- **SHA256 Verification**: Each file change verified against expected hash
- **Litmus Testing**: Full test suite runs after each subtask

### 3. Testing Protocol
```powershell
# Run litmus tests
python litmus_test.py

# If tests fail, regenerate without specific issue
# Example: "regenerate without WebGL CDN fail"
```

### 4. Approval Workflow
```powershell
# Start upgrade (DRY_RUN mode by default)
$env:DRY_RUN = "true"
.\upgrade_trading_tab.ps1 -Subtask 1

# Review changes
git diff

# Approve and apply
$env:DRY_RUN = "false"
.\upgrade_trading_tab.ps1 -Subtask 1 -Approve

# Or reject and rollback
git checkout main
git branch -D ui_chart_fix
```

### 5. Emergency Procedures
```powershell
# Full stop
Get-Process | Where-Object {$_.ProcessName -like "*ares*"} | Stop-Process -Force

# Rollback last change
git revert HEAD

# Rollback entire subtask
git checkout main
git branch -D {subtask_branch_name}

# Nuclear option: Reset to last known good commit
git reset --hard 77fd60f
```

## Current Status
- ✅ Litmus test suite created
- ✅ Control system documented
- ⏳ Ready to start Subtask 1: Chart Upgrade

## Next Steps
1. Run baseline litmus test: `python litmus_test.py`
2. Start Subtask 1 in DRY_RUN mode
3. Review, approve, apply
4. Test, verify, merge
5. Repeat for remaining 11 subtasks
