# 🚀 ARES Coordinator - Production Setup Guide

## 📋 Overview

This is the production-ready setup for the ARES Agent Swarm Coordinator with isolated virtual environment, comprehensive logging, error handling, and environment validation.

## ✨ Features

- **Isolated Virtual Environment**: No dependency conflicts
- **Environment Validation**: Automatic checks for required API keys and configuration
- **Comprehensive Logging**: Rotating logs (10MB × 5 files = 50MB max)
- **Process Cleanup**: Automatic cleanup of orphaned processes
- **Graceful Shutdown**: Ctrl+C cleanup with signal handler
- **One-Command Setup**: `.\setup.ps1` for fresh install
- **One-Command Start**: `.\start.ps1` for validated startup

---

## 🚀 Quick Start

### First-Time Setup (One-Time)

```powershell
# 1. Navigate to the coordinator directory
cd C:\ARES_Workspace\ARES_API\internal\agent_swarm

# 2. Run the setup script
.\setup.ps1

# 3. Copy the environment template
Copy-Item .env.example .env

# 4. Edit .env and add your API keys
notepad .env
```

### Daily Usage

```powershell
# Start the coordinator (validates environment automatically)
.\start.ps1
```

That's it! 🎉

---

## 📁 File Structure

```
agent_swarm/
├── coordinator.py              # Main coordinator (enhanced with logging)
├── validate_env.py            # Environment validation script
├── file_operations.py         # File operation handlers
├── setup.ps1                  # One-command setup script
├── start.ps1                  # One-command start script
├── requirements.txt           # Python dependencies
├── .env.example              # Environment template
├── .env                       # Your actual config (create this!)
├── venv/                      # Virtual environment (auto-created)
└── logs/                      # Log files (auto-created)
    └── solace_coordinator.log
```

---

## 🔧 Configuration Files

### `.env.example` → `.env`

Template for required environment variables:

```ini
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ares_db
DB_USER=ARES
DB_PASSWORD=ARESISWAKING

# OpenAI API Key (Required)
OPENAI_API_KEY=sk-proj-your-key-here

# Anthropic API Key (Optional)
ANTHROPIC_API_KEY=your-key-here

# DeepSeek API (Optional)
DEEPSEEK_API_KEY=your-deepseek-key-here
DEEPSEEK_API_URL=https://api.deepseek.com/v1

# WebSocket Configuration
WEBSOCKET_HOST=localhost
WEBSOCKET_PORT=8765
```

### `requirements.txt`

All Python dependencies with locked versions:

```
psycopg2-binary==2.9.11
openai==2.4.0
anthropic==0.71.0
websockets==15.0.1
python-dotenv==1.1.1
# ... and more
```

---

## 📝 Scripts

### `setup.ps1` - Initial Setup

**Purpose:** Create fresh virtual environment and install dependencies

**What it does:**
1. ✅ Removes old `venv` if exists
2. ✅ Creates new Python virtual environment
3. ✅ Activates the virtual environment
4. ✅ Upgrades pip to latest version
5. ✅ Installs all dependencies from `requirements.txt`
6. ✅ Checks for `.env` file

**Usage:**
```powershell
.\setup.ps1
```

**When to use:**
- First-time setup
- After pulling new code with dependency changes
- When you need to reset the environment

---

### `start.ps1` - Start Coordinator

**Purpose:** Validate environment and start the WebSocket server

**What it does:**
1. ✅ Activates virtual environment
2. ✅ Runs environment validation (`validate_env.py`)
3. ✅ Starts WebSocket server on `ws://localhost:8765`

**Usage:**
```powershell
.\start.ps1
```

**When to use:**
- Daily startup
- After configuration changes
- Anytime you want to start the coordinator

---

### `validate_env.py` - Environment Validation

**Purpose:** Check all required environment variables before startup

**What it checks:**

✅ **Required Variables:**
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- `OPENAI_API_KEY`

ℹ️ **Optional Variables:**
- `ANTHROPIC_API_KEY` (for Claude agent)
- `DEEPSEEK_API_KEY` (for DeepSeek agents)
- `WEBSOCKET_HOST`, `WEBSOCKET_PORT`

**Output Example:**
```
======================================================================
🔍 Validating Environment Configuration
======================================================================
✅ DB_HOST: localhost
✅ DB_PORT: 5432
✅ DB_NAME: ares_db
✅ DB_USER: ARES
✅ DB_PASSWORD: ARESISW...
✅ OPENAI_API_KEY: sk-proj-...

📋 Optional Configuration:
✅ ANTHROPIC_API_KEY: sk-ant-a...
ℹ️  DEEPSEEK_API_KEY: Not set (DeepSeek agents disabled)
✅ WEBSOCKET_HOST: localhost
✅ WEBSOCKET_PORT: 8765
======================================================================
✅ All required environment variables are set!
======================================================================
```

---

## 🛠️ Manual Commands

### Activate Virtual Environment

```powershell
.\venv\Scripts\Activate.ps1
```

### Install Dependencies Manually

```powershell
pip install -r requirements.txt
```

### Run Coordinator Manually

```powershell
# With validation
python validate_env.py
python coordinator.py --websocket

# Skip validation (not recommended)
python coordinator.py --websocket
```

### View Logs

```powershell
# Real-time log monitoring (tail -f equivalent)
Get-Content logs\solace_coordinator.log -Wait -Tail 50

# View full log
Get-Content logs\solace_coordinator.log

# View last 100 lines
Get-Content logs\solace_coordinator.log -Tail 100
```

---

## 📊 Logging

### Log Configuration

- **Location:** `logs/solace_coordinator.log`
- **Rotation:** 10MB per file, 5 backup files (50MB total)
- **Format:** `2025-10-17 14:30:00 - INFO - [WEBSOCKET] Message received`
- **Encoding:** UTF-8 (handles special characters)

### Log Levels

```python
# Change in coordinator.py (line 54)
logger.setLevel(logging.INFO)   # Production (default)
logger.setLevel(logging.DEBUG)  # Verbose debugging
```

### Log Output Examples

**Startup:**
```
======================================================================
SOLACE Coordinator Starting...
======================================================================
INFO: Signal handler registered (Ctrl+C for graceful shutdown)
======================================================================
Starting SOLACE WebSocket Server
======================================================================
INFO: ✅ WebSocket server started on ws://localhost:8765
```

**WebSocket Messages:**
```
INFO: [WEBSOCKET] New connection from 127.0.0.1:xxxxx
INFO: [WEBSOCKET] Message type: chat
INFO: [CHAT] Message received (25 chars): "What is the weather?"
INFO: [CHAT] ✓ Final response sent
```

**Errors:**
```
ERROR: [WEBSOCKET] JSON decode error: Expecting value: line 1 column 1
ERROR: [WEBSOCKET] File not found: config.json
```

---

## 🔧 Troubleshooting

### Issue: "Virtual environment not found"

**Solution:** Run setup first
```powershell
.\setup.ps1
```

### Issue: "Missing environment variables"

**Solution:** Create and configure `.env` file
```powershell
Copy-Item .env.example .env
notepad .env
```

### Issue: "Port 8765 already in use"

**Solution:** Kill existing process
```powershell
Get-Process python | Where-Object {$_.Path -like "*agent_swarm*"} | Stop-Process -Force
```

Or use the built-in cleanup (runs automatically on start):
```powershell
.\start.ps1
```

### Issue: "Database connection failed"

**Solution:** Check database configuration in `.env`
```ini
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ares_db
DB_USER=ARES
DB_PASSWORD=ARESISWAKING
```

### Issue: "API key invalid"

**Solution:** Verify API key in `.env`
```powershell
# Test OpenAI key
python -c "import os; from dotenv import load_dotenv; load_dotenv(); print(os.getenv('OPENAI_API_KEY')[:20])"
```

---

## 🧪 Testing

### Test Environment Validation

```powershell
python validate_env.py
```

### Test WebSocket Server

```powershell
# Terminal 1: Start server
.\start.ps1

# Terminal 2: Test connection
python test_websocket_server.py
```

### Run All Instruction #6 Tests

```powershell
.\run_instruction_6_tests.ps1
```

---

## 🔄 Updating

### Pull New Code

```powershell
git pull origin main
```

### Update Dependencies

```powershell
.\setup.ps1  # Re-run setup to update packages
```

### Reset Environment (Nuclear Option)

```powershell
Remove-Item -Recurse -Force venv
.\setup.ps1
```

---

## 📞 Support

### Log Locations

- **Coordinator Log:** `logs/solace_coordinator.log`
- **Backup Logs:** `logs/solace_coordinator.log.1` through `.5`

### Common Issues

| Issue | File to Check |
|-------|---------------|
| Database errors | `.env` (database config) |
| API errors | `.env` (API keys) |
| Import errors | `requirements.txt` (dependencies) |
| Port conflicts | Process cleanup (automatic) |

### Debug Mode

```powershell
# Run coordinator with debug logging
.\venv\Scripts\Activate.ps1
python coordinator.py --websocket --debug
```

---

## 🎯 Architecture

### Components

1. **coordinator.py**: Main orchestration controller
   - Handles WebSocket connections
   - Routes messages to appropriate handlers
   - Manages OpenAI function calling
   - Comprehensive logging and error handling

2. **validate_env.py**: Environment validator
   - Checks required variables
   - Reports optional variables
   - Masks sensitive values in output

3. **file_operations.py**: File system operations
   - Read, write, list files
   - Used by WebSocket handlers

4. **setup.ps1**: Setup automation
   - Virtual environment creation
   - Dependency installation

5. **start.ps1**: Startup automation
   - Environment activation
   - Validation
   - Server start

---

## 🔒 Security Notes

- ✅ API keys are stored in `.env` (not committed to git)
- ✅ Sensitive values are masked in validation output
- ✅ Logs do not contain API keys or passwords
- ⚠️ Add `logs/` to `.gitignore` if not already present
- ⚠️ Never commit `.env` file

---

## 📈 Performance

- **Log Rotation:** Prevents disk space issues (50MB max)
- **Process Cleanup:** Automatic cleanup on startup
- **Connection Pooling:** Database connections managed by coordinator
- **Async WebSocket:** Non-blocking message handling

---

## 🎉 Summary

**Setup:**
```powershell
.\setup.ps1
Copy-Item .env.example .env
notepad .env  # Add your API keys
```

**Daily Use:**
```powershell
.\start.ps1
```

**Stop Server:**
```
Press Ctrl+C
```

**View Logs:**
```powershell
Get-Content logs\solace_coordinator.log -Wait -Tail 50
```

---

*Last Updated: October 17, 2025*
*ARES Coordinator Version: Production v2.0*
