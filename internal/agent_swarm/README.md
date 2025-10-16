# ARES Agent Swarm System

Multi-agent AI coordination system for ARES Platform.

## 🤖 The Team

### **SOLACE** (OpenAI GPT-4)
- **Role:** Strategic Coordinator & Commander
- **Capabilities:** Strategy, coordination, trading, decision-making
- **API:** OpenAI GPT-4
- **Task:** Analyzes requests, delegates work, makes strategic decisions

### **FORGE** (Anthropic Claude)
- **Role:** UI Builder & Frontend Developer
- **Capabilities:** UI building, coding, React, HTML, CSS
- **API:** Anthropic Claude 3.5 Sonnet
- **Task:** Creates user interfaces and frontend components

### **ARCHITECT** (DeepSeek-R1 via Ollama)
- **Role:** System Designer & Planner
- **Capabilities:** Planning, design patterns, architecture
- **API:** Ollama (DeepSeek-R1 14B local)
- **Task:** Designs system structure and architectural patterns

### **SENTINEL** (DeepSeek-R1 via Ollama)
- **Role:** Quality Assurance & Debugger
- **Capabilities:** Debugging, testing, error detection, validation
- **API:** Ollama (DeepSeek-R1 14B local)
- **Task:** Tests code, finds bugs, validates quality

---

## 📋 Quick Start

### 1. **Prerequisites**
- PostgreSQL database running (ares_db)
- Go backend (ARES_API) running on port 8080
- Python 3.9+ installed
- (Optional) Ollama installed with DeepSeek-R1:14b model
- (Optional) OpenAI API key for SOLACE
- (Optional) Anthropic API key for FORGE

### 2. **Install Python Dependencies**
```bash
cd c:\ARES_Workspace\ARES_API\internal\agent_swarm
pip install -r requirements.txt
```

### 3. **Set Environment Variables**
```powershell
# Required for database
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_NAME = "ares_db"
$env:DB_USER = "ARES"
$env:DB_PASSWORD = "ARESISWAKING"

# Optional - for AI agents
$env:OPENAI_API_KEY = "sk-..."          # For SOLACE
$env:ANTHROPIC_API_KEY = "sk-ant-..."   # For FORGE
```

### 4. **Start the Coordinator**
```powershell
# Option 1: Using the startup script
.\start-coordinator.ps1

# Option 2: Direct Python
python coordinator.py --interval 10
```

### 5. **Open the Dashboard**
Navigate to: `http://localhost:8080/web/agent-dashboard.html`

---

## 🔌 API Endpoints

### Agent Management
```bash
# List all agents
GET /api/v1/agents

# Get specific agent
GET /api/v1/agents/:name

# Get agent performance
GET /api/v1/agents/:name/performance
```

### Task Management
```bash
# Create task (auto-assigns to SOLACE)
POST /api/v1/agents/tasks
{
  "task_type": "ui_build",
  "description": "Create agent dashboard",
  "priority": 8,
  "file_paths": ["web/agent-dashboard.html"]
}

# Get pending tasks
GET /api/v1/agents/tasks/pending

# Get task details
GET /api/v1/agents/tasks/:id

# Assign task manually
POST /api/v1/agents/tasks/:id/assign
{
  "agent_name": "FORGE"
}

# Complete task
POST /api/v1/agents/tasks/:id/complete
{
  "result": { "status": "success" }
}

# Fail task
POST /api/v1/agents/tasks/:id/fail
{
  "error": "Reason for failure"
}
```

### File Registry
```bash
# List all tracked files
GET /api/v1/agents/files

# Get file by path
GET /api/v1/agents/files/by-path?path=/web/dashboard.html
```

### Build History
```bash
# Get recent builds
GET /api/v1/agents/builds
```

---

## 🎯 How It Works

### Workflow Example

1. **User creates a task:**
   ```bash
   POST /api/v1/agents/tasks
   {
     "task_type": "ui_build",
     "description": "Build trading analytics dashboard",
     "priority": 8
   }
   ```

2. **Task auto-assigned to SOLACE** (status: `assigned`)

3. **Coordinator detects assigned task**
   - Polls database every 10 seconds
   - Finds task assigned to SOLACE

4. **SOLACE analyzes the task**
   - Calls OpenAI GPT-4 API
   - Decides: "This is a UI task → Delegate to FORGE"
   - Creates new task assigned to FORGE

5. **FORGE executes the task**
   - Coordinator detects FORGE task
   - Calls Anthropic Claude API
   - FORGE generates HTML/CSS/JS code
   - Saves files to workspace
   - Updates file_registry

6. **Task marked complete**
   - Status: `completed`
   - Agent status: `idle`
   - Logged to `agent_task_history`

7. **Optional: SENTINEL validates**
   - SOLACE can create validation task
   - SENTINEL runs tests
   - Reports results

---

## 📊 Database Schema

### agent_registry
Tracks all AI agents and their status.

### task_queue
Coordinates work between agents.

### file_registry
Tracks all workspace files.

### agent_task_history
Performance tracking and learning.

### build_history
Track build attempts and results.

### file_dependencies
Dependency graph between files.

See `AGENT_SWARM_DEPLOYMENT_SUMMARY.md` for full schema details.

---

## 🔧 Configuration

### Coordinator Settings
Edit `coordinator.py` or use command-line args:

```bash
python coordinator.py --interval 5 --debug
```

- `--interval SECONDS`: How often to check for new tasks (default: 10)
- `--debug`: Enable debug logging

### Agent Availability
The coordinator will work with whatever agents you have API keys for:
- ✅ **SOLACE only**: Basic coordination without delegation
- ✅ **SOLACE + FORGE**: Full UI building capability
- ✅ **All agents**: Complete multi-agent system

---

## 📈 Monitoring

### Real-Time Dashboard
`http://localhost:8080/web/agent-dashboard.html`

Features:
- ✅ Live agent status (idle/busy/offline)
- ✅ Task queue with priorities
- ✅ File registry tracking
- ✅ Auto-refresh every 5 seconds
- ✅ Agent performance metrics

### Logs
- **Coordinator logs:** `agent_coordinator.log`
- **ARES API logs:** Console output
- **Database logs:** PostgreSQL logs

### Database Queries
```sql
-- Check agent status
SELECT agent_name, status, total_tasks_completed, success_rate 
FROM agent_registry;

-- View task queue
SELECT task_id, task_type, status, assigned_to_agent, created_at 
FROM task_queue 
ORDER BY priority DESC, created_at ASC;

-- Agent performance
SELECT agent_name, COUNT(*) as tasks, 
       AVG(duration_ms) as avg_time,
       SUM(CASE WHEN success THEN 1 ELSE 0 END)::float / COUNT(*) as success_rate
FROM agent_task_history
GROUP BY agent_name;
```

---

## 🚀 Production Deployment

### Windows Service
Convert coordinator to a Windows service:
```powershell
# Install NSSM (Non-Sucking Service Manager)
choco install nssm

# Create service
nssm install ARESAgentCoordinator "python.exe" "C:\ARES_Workspace\ARES_API\internal\agent_swarm\coordinator.py"
nssm set ARESAgentCoordinator AppDirectory "C:\ARES_Workspace\ARES_API\internal\agent_swarm"
nssm set ARESAgentCoordinator AppEnvironmentExtra "DB_PASSWORD=ARESISWAKING"

# Start service
nssm start ARESAgentCoordinator
```

### Linux Systemd
Create `/etc/systemd/system/ares-coordinator.service`:
```ini
[Unit]
Description=ARES Agent Swarm Coordinator
After=network.target postgresql.service

[Service]
Type=simple
User=ares
WorkingDirectory=/opt/ARES_API/internal/agent_swarm
Environment="DB_HOST=localhost"
Environment="DB_PASSWORD=ARESISWAKING"
Environment="OPENAI_API_KEY=sk-..."
Environment="ANTHROPIC_API_KEY=sk-ant-..."
ExecStart=/usr/bin/python3 coordinator.py --interval 10
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable ares-coordinator
sudo systemctl start ares-coordinator
```

---

## 🐛 Troubleshooting

### Coordinator not starting
```bash
# Check Python version
python --version  # Should be 3.9+

# Check dependencies
pip install -r requirements.txt

# Check database connection
$env:PGPASSWORD='ARESISWAKING'
psql -h localhost -U ARES -d ares_db -c "SELECT COUNT(*) FROM agent_registry;"
```

### Agent unavailable
```bash
# OpenAI (SOLACE)
echo $env:OPENAI_API_KEY  # Should be set

# Anthropic (FORGE)
echo $env:ANTHROPIC_API_KEY  # Should be set

# Ollama (ARCHITECT, SENTINEL)
curl http://localhost:11434/api/tags  # Should return model list
```

### Tasks not being executed
```bash
# Check coordinator logs
Get-Content agent_coordinator.log -Tail 50

# Check database
psql -h localhost -U ARES -d ares_db -c "SELECT * FROM task_queue WHERE status='assigned';"

# Verify coordinator is running
Get-Process | Where-Object { $_.Name -like "*python*" }
```

### Dashboard not loading
```bash
# Check ARES API is running
curl http://localhost:8080/health

# Check agent endpoint
curl http://localhost:8080/api/v1/agents

# Check browser console for errors
```

---

## 📝 Task Types

Common task types you can create:

- `ui_build` - Build user interface components
- `debug` - Fix bugs in code
- `plan` - Create architectural plans
- `test` - Write and run tests
- `refactor` - Improve code structure
- `document` - Write documentation
- `analyze` - Analyze code or data
- `deploy` - Deployment tasks

---

## 💡 Best Practices

### Task Descriptions
Be specific:
```json
// ❌ Bad
{"description": "Fix the UI"}

// ✅ Good
{"description": "Fix the purple sidebar navigation hover effect in agent-dashboard.html"}
```

### Priority Guidelines
- **1-3**: Low priority, nice-to-have
- **4-6**: Medium priority, should be done
- **7-9**: High priority, important
- **10**: Critical, urgent

### File Tracking
Always include relevant files in tasks:
```json
{
  "description": "Add dark mode toggle",
  "file_paths": ["web/dashboard.html", "static/css/theme.css"]
}
```

---

## 🔮 Future Enhancements

- [ ] File system watcher (auto-detect changes)
- [ ] Initial workspace scan
- [ ] Task dependencies (wait for Task A before B)
- [ ] Multi-file task coordination
- [ ] Agent learning from past mistakes
- [ ] Cost optimization (track API usage)
- [ ] Agent performance analytics
- [ ] WebSocket real-time updates
- [ ] Task scheduling (cron-like)
- [ ] Agent collaboration (multiple agents on one task)

---

## 📚 Resources

- **Full Deployment Guide:** `AGENT_SWARM_DEPLOYMENT_SUMMARY.md`
- **Database Schema:** `migrations/007_agent_swarm.sql`
- **API Documentation:** Swagger at `http://localhost:8080/swagger/index.html`
- **Dashboard:** `http://localhost:8080/web/agent-dashboard.html`

---

**Built with ❤️ for the ARES Platform**  
*Multi-agent AI coordination made simple*
