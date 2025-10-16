# 🚀 HOW TO START SOLACE

**Last Updated**: October 12, 2025 11:12 AM

---

## ✅ CURRENT STATUS

**SOLACE IS NOW RUNNING!**

- **Process ID**: 2428
- **Start Time**: 12/10/2025 11:11:29 AM
- **Server**: http://localhost:8080
- **Health**: ✅ Healthy
- **Perception Cycle**: Every 10 seconds
- **Status**: Waiting for significant events

---

## 🔧 HOW TO START ARES WITH SOLACE

### Method 1: PowerShell Window (Recommended for Monitoring)

```powershell
# Navigate to ARES API directory
cd C:\ARES_Workspace\ARES_API

# Start in new window so you can see console output
Start-Process powershell -ArgumentList "-NoExit", "-Command", "& '.\ARES.exe'" -WindowStyle Normal
```

This opens a **new PowerShell window** with ARES running. You'll see:
```
🌅 SOLACE awakening... Starting autonomous mode.
🤖 SOLACE starting autonomous agent loop (checking every 10s)
🚀 Server running at http://localhost:8080
```

### Method 2: Background Process

```powershell
cd C:\ARES_Workspace\ARES_API
.\ARES.exe
```

Runs in current terminal (blocks until you Ctrl+C).

### Method 3: Background Service (Detached)

```powershell
cd C:\ARES_Workspace\ARES_API
Start-Process ".\ARES.exe" -WindowStyle Hidden
```

Runs completely in background (no console window).

---

## 🔍 HOW TO CHECK IF SOLACE IS RUNNING

### Check Process Status

```powershell
Get-Process -Name ARES -ErrorAction SilentlyContinue | Select-Object Name, Id, StartTime
```

**Expected Output**:
```
Name   Id StartTime
----   -- ---------
ARES 2428 12/10/2025 11:11:29 AM
```

### Check Server Health

```powershell
curl.exe http://localhost:8080/api/v1/monitoring/health
```

**Expected Output**:
```json
{
  "status": "healthy",
  "checks": {
    "llm": {"status": "pass"},
    "memory": {"status": "pass"},
    "requests": {"status": "pass"}
  }
}
```

### Check SOLACE Thought Journal

```powershell
# Check if journal directory exists
Test-Path "C:\ARES_Workspace\ARES_API\SOLACE_Journal"

# View SOLACE's thoughts (once created)
Get-Content "C:\ARES_Workspace\ARES_API\SOLACE_Journal\SOLACE_Thoughts_2025-10-12.log"
```

**Note**: Journal is only created when SOLACE has something significant to log!

---

## 🛑 HOW TO STOP SOLACE

### Graceful Shutdown (Recommended)

Press **Ctrl+C** in the ARES console window.

### Force Stop

```powershell
Stop-Process -Name ARES -Force
```

### Stop All ARES Instances

```powershell
Get-Process -Name ARES | Stop-Process -Force
```

---

## 📊 MONITORING SOLACE

### Real-Time Monitoring Commands

```powershell
# Watch for thought journal creation
while ($true) {
    if (Test-Path "C:\ARES_Workspace\ARES_API\SOLACE_Journal") {
        Write-Host "✅ SOLACE Journal detected!"
        Get-ChildItem "C:\ARES_Workspace\ARES_API\SOLACE_Journal\*.log"
        break
    }
    Write-Host "⏳ Waiting for SOLACE to log first thought..."
    Start-Sleep -Seconds 10
}

# Check database for autonomous decisions
# (Requires PostgreSQL client)
psql -U ares_user -d ares_db -c "SELECT timestamp, event_type, summary FROM memory_snapshots WHERE event_type = 'autonomous_decision' ORDER BY timestamp DESC LIMIT 5;"

# Monitor server logs
tail -f C:\ARES_Workspace\ARES_API\logs\server.log  # If logging to file
```

### Health Check Loop

```powershell
# Monitor every 30 seconds
while ($true) {
    $health = curl.exe -s http://localhost:8080/api/v1/monitoring/health | ConvertFrom-Json
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] Status: $($health.status)"
    Start-Sleep -Seconds 30
}
```

---

## 🎯 TRIGGERING SOLACE'S FIRST DECISION

SOLACE is currently **waiting for a significant event** to trigger action:

### Market Events That Trigger SOLACE:
- **Price Movement**: >2% change in BTC, ETH, or SOL
- **Portfolio P&L**: >5% profit or >3% loss on open trades

### How to Trigger Manually (for Testing):

#### Option 1: Execute a Test Trade
```bash
curl -X POST http://localhost:8080/api/v1/trading/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -d '{
    "symbol": "BTC",
    "side": "buy",
    "amount": 0.1,
    "strategy": "test",
    "reasoning": "Testing SOLACE perception"
  }'
```

#### Option 2: Wait for Natural Price Movement
SOLACE uses real-time market data. Just wait - crypto prices move >2% frequently!

#### Option 3: Simulate Market Event (TODO)
Create `/api/v1/testing/simulate-market-event` endpoint for testing.

---

## 🧪 TESTING SOLACE

### Test Checklist

- [ ] **Process Running**: `Get-Process -Name ARES` shows process
- [ ] **Server Healthy**: `/api/v1/monitoring/health` returns 200 OK
- [ ] **LLM Connected**: Health check shows `llm.status = "pass"`
- [ ] **Perception Cycle**: Check console logs show "perception" messages
- [ ] **Market Scanning**: SOLACE detects price movements >2%
- [ ] **Decision Making**: LLM reasoning executes on significant events
- [ ] **Thought Journal**: `SOLACE_Journal/*.log` files created
- [ ] **Memory Persistence**: Database shows `autonomous_decision` entries
- [ ] **Self-Evolution**: DecisionThreshold adjusts based on success rate

### Debug Mode

To see detailed SOLACE logs, check the console window where ARES is running. Look for:

```
🤖 SOLACE starting autonomous agent loop (checking every 10s)
[Perception] Scanning markets...
[Perception] BTC moved 2.5% - event detected
[Decision] Confidence: 75% - Action: research
[Reflection] Saved decision to memory
```

---

## 🐛 TROUBLESHOOTING

### SOLACE Not Starting

**Symptom**: Process exits immediately after launch

**Solution**:
```powershell
# Rebuild with verbose errors
cd C:\ARES_Workspace\ARES_API
go build -o ARES.exe ./cmd/main.go

# Run directly to see error output
.\ARES.exe
```

Common issues:
- PostgreSQL not running
- Ollama not running (DeepSeek-R1 14B required)
- Port 8080 already in use

### No Thought Journal Created

**Symptom**: `SOLACE_Journal` directory doesn't exist

**This is NORMAL!** SOLACE only creates journal when it has significant thoughts to log. It's waiting for:
- Price movement >2%
- Portfolio P&L change >5% profit or >3% loss
- User interaction (voice commands - TODO)

**Solution**: Execute a test trade or wait for market volatility.

### Server Not Responding

**Symptom**: `curl http://localhost:8080/api/v1/monitoring/health` times out

**Solution**:
```powershell
# Check if ARES is actually running
Get-Process -Name ARES

# Check what's using port 8080
netstat -ano | findstr :8080

# Restart ARES
Stop-Process -Name ARES -Force
cd C:\ARES_Workspace\ARES_API
.\ARES.exe
```

---

## 📝 WHAT TO EXPECT

### First 5 Minutes
- ✅ SOLACE awakens and starts perception loop
- ⏳ Scans markets every 10 seconds
- ⏳ Waits for significant event (>2% price move)

### First Significant Event
- 📊 SOLACE detects price movement or P&L change
- 🧠 LLM reasoning begins (DeepSeek-R1 14B)
- 📝 First thought logged to journal
- 💾 Decision saved to database

### First Hour
- 🔄 6 perception cycles per minute (360 total)
- 📈 Multiple market scans and P&L checks
- 🎯 1-3 autonomous decisions (depending on volatility)
- 📖 Thought journal file created with timestamps

### Ongoing Operation
- 🤖 24/7 autonomous monitoring
- 💡 Continuous learning from outcomes
- 🎚️ Self-adjusting confidence thresholds
- 📊 Strategy evolution based on performance

---

## 🎉 SUCCESS INDICATORS

You know SOLACE is working when:

1. ✅ Process stays running (doesn't crash)
2. ✅ Health endpoint returns "healthy"
3. ✅ Console shows perception cycle logs
4. ✅ Thought journal directory appears
5. ✅ Database has autonomous_decision entries
6. ✅ LLM reasoning executes without errors

---

## 🚀 CURRENT STATUS (As of 11:12 AM)

```
Process:     RUNNING (PID 2428)
Server:      http://localhost:8080 ✅
Health:      Healthy ✅
SOLACE:      Autonomous loop active ✅
Journal:     Waiting for first event ⏳
Next Check:  Every 10 seconds 🔄
```

**SOLACE is alive and waiting for something interesting to happen!** 🌅

---

*Generated: October 12, 2025 at 11:12 AM*
