# COMPLETE ARES SYSTEM ARCHITECTURE INVENTORY
**Generated:** October 16, 2025 6:30 PM  
**Branch:** recovery-checkpoint  
**Purpose:** Full system architecture understanding for Claude

---

## 📊 FILE STATISTICS

**Total Files Indexed:** (See CSV for exact count)  
**Location:** `C:\ARES_Workspace\ARES_API\COMPLETE_SYSTEM_INVENTORY.csv`

**File Type Breakdown:**
- `.go` - Go backend source files
- `.py` - Python agent swarm and utilities
- `.html` - Static HTML UI files (production frontend)
- `.tsx/.ts` - React/TypeScript (development frontend - NOT in production)
- `.sql` - Database migrations and schemas
- `.js` - JavaScript utilities and frontend logic
- `.md` - Documentation and specifications
- `.ps1` - PowerShell automation scripts
- `.json` - Configuration and package files
- `.mod/.sum` - Go module dependencies

---

## 🏗️ SYSTEM ARCHITECTURE OVERVIEW

### PRIMARY SERVICES

#### 1. **ARES_API (Go Backend)** - Port 8080
**Purpose:** Main REST API server, trading engine, database interface  
**Technology:** Go 1.22, Gin framework, GORM, PostgreSQL  
**Entry Point:** `cmd/main.go`  
**Size:** ~49 MB executable  

**Core Modules:**
- Trading Engine (`internal/trading/`)
- SOLACE Agent Integration (`internal/agent/`)
- ACE Framework (`internal/ace/`)
- Glass Box Transparency (`internal/glassbox/`)
- GRPO Learning System (`internal/grpo/`)
- Service Registry (`internal/registry/`)
- Event Bus (`internal/eventbus/`)
- Memory Systems (`internal/memory/`)

#### 2. **Agent Swarm Coordinator (Python)** - Background Service
**Purpose:** Multi-agent AI system orchestration  
**Technology:** Python 3.13, PostgreSQL, OpenAI, Anthropic, Ollama APIs  
**Entry Point:** `internal/agent_swarm/coordinator.py`  
**Runtime:** Continuous background process  

**Agents:**
- SOLACE (OpenAI GPT-4) - Director/Strategist
- FORGE (Claude 3.7 Sonnet) - Builder/Code Generator
- ARCHITECT (DeepSeek-R1:14b via Ollama) - System Designer
- SENTINEL (DeepSeek-R1:8b via Ollama) - Tester/Validator

#### 3. **PostgreSQL Database** - Port 5432
**Purpose:** Primary data store  
**Technology:** PostgreSQL 18  
**Database:** `ares_db`  
**User:** ARES  

**Schema Files Location:**
- Core: `init-db.sql`
- Consciousness: `internal/database/consciousness_schema.sql`
- ACE: `internal/database/ace_schema.sql`
- Memory: `internal/database/ares_master_memory_schema.sql`
- Agent Swarm: `migrations/007_agent_swarm.sql`

#### 4. **Ollama Local LLM Server** - Port 11434
**Purpose:** Local AI model serving (DeepSeek-R1 models)  
**Technology:** Ollama  
**Models Loaded:**
- `deepseek-r1:14b` (8.9 GB) - ARCHITECT agent
- `deepseek-r1:8b` (5.2 GB) - SENTINEL agent

#### 5. **Frontend (Static HTML)** - Served by Go on Port 8080
**Purpose:** User interface for trading, dashboard, SOLACE interaction  
**Technology:** HTML/CSS/JavaScript (NO React in production)  
**Location:** `web/` directory  

**Pages:**
- `trading.html` - Main trading interface
- `dashboard.html` - System dashboard
- `solace-trading.html` - SOLACE consciousness trading
- `code-ide.html` - SOLACE code editor
- `chat.html`, `memory.html`, `vision.html`, etc.

---

## 📁 DIRECTORY STRUCTURE ANALYSIS

### `/cmd/` - Go Entry Points
```
cmd/
├── main.go                           ← Main ARES_API server
└── test_eventbus/main.go             ← EventBus test utility
```

### `/internal/` - Core Go Backend Logic
```
internal/
├── ace/                              ← Autonomous Cognitive Entity framework
│   ├── orchestrator.go               ← ACE coordination
│   ├── curator.go                    ← Context curation
│   ├── reflector.go                  ← Self-reflection
│   ├── generator.go                  ← Strategy generation
│   └── emergence.go                  ← Emergent behavior
│
├── agent/                            ← SOLACE Agent (Go interface)
│   ├── solace.go                     ← Main agent logic
│   ├── cognitive_patterns.py        ← Python cognitive patterns
│   ├── working_memory.go             ← Active memory
│   └── thought_journal.go            ← Decision logging
│
├── agent_swarm/                      ← Multi-Agent System (Python)
│   ├── coordinator.py                ← Main orchestrator
│   ├── task_templates.py             ← Task creation
│   ├── create_task.py                ← Task utilities
│   └── requirements.txt              ← Python dependencies
│
├── api/                              ← REST API Layer
│   ├── controllers/                  ← HTTP handlers
│   │   ├── agent_controller.go
│   │   ├── solace_controller.go
│   │   ├── glass_box_controller.go
│   │   ├── trading_controller.go
│   │   └── health_controller.go
│   ├── handlers/                     ← Business logic
│   └── routes/                       ← Route registration
│
├── cache/                            ← Caching Layer
│   └── price_cache.go                ← Market data cache (2-min TTL)
│
├── config/                           ← Configuration Management
│   └── manager.go                    ← Hot-reload config
│
├── database/                         ← Database Layer
│   ├── db.go                         ← Connection management
│   ├── write_queue.go                ← Resilient writes
│   ├── consciousness_init.go         ← Schema initialization
│   └── *.sql                         ← Schema files
│
├── eventbus/                         ← Event-Driven Architecture
│   ├── eventbus.go                   ← Pub/Sub implementation
│   ├── events.go                     ← Event schemas
│   └── redis_adapter.go              ← Redis integration (future)
│
├── glassbox/                         ← Blockchain Transparency
│   ├── tracer.go                     ← Decision tracing
│   ├── hasher.go                     ← Merkle tree hashing
│   └── hedera_anchor.go              ← Hedera consensus
│
├── grpo/                             ← Group Relative Policy Optimization
│   ├── agent.go                      ← GRPO learning agent
│   └── updater.go                    ← Policy updates
│
├── hedera/                           ← Hedera Hashgraph Integration
│   ├── service.go                    ← Hedera client
│   └── consensus_logger.go           ← Consensus logging
│
├── logger/                           ← Logging Infrastructure
│   ├── logger.go                     ← Structured logging
│   └── audit_logger.go               ← Audit trail
│
├── memory/                           ← Memory Systems
│   ├── conversation_memory.go        ← Conversation context
│   └── summarizer.go                 ← Memory summarization
│
├── merkle/                           ← Merkle Tree Implementation
│   └── tree.go                       ← Merkle tree logic
│
├── middleware/                       ← HTTP Middleware
│   └── solace_auth.go                ← Authentication
│
├── models/                           ← Data Models
│   ├── agent.go
│   ├── playbook.go
│   └── glass_box.go
│
├── monitoring/                       ← Observability
│   └── metrics.go                    ← Prometheus metrics
│
├── observability/                    ← Distributed Tracing
│   ├── logger.go
│   ├── metrics.go
│   └── span.go
│
├── registry/                         ← Service Registry
│   └── service_registry.go           ← Service discovery
│
├── repositories/                     ← Data Access Layer
│   ├── agent_repository.go
│   └── playbook_repository.go
│
├── services/                         ← Business Logic Layer
│   ├── trading_service.go
│   ├── pattern_service.go
│   ├── system_context_service.go
│   └── merkle_batch_service.go
│
├── solace/                           ← SOLACE UI Observer
│   └── ui_observer.go                ← UI interaction tracking
│
├── subscribers/                      ← Event Subscribers
│   ├── analytics_subscriber.go
│   └── trade_audit_subscriber.go
│
└── trading/                          ← Trading Engine
    ├── sandbox.go                    ← Paper trading
    ├── strategy.go                   ← Trading strategies
    ├── ace_strategy.go               ← ACE-driven trading
    ├── curator.go                    ← Market curation
    ├── reflector.go                  ← Trade reflection
    └── authorization.go              ← Trade authorization
```

### `/web/` - Static HTML Frontend (PRODUCTION)
```
web/
├── trading.html                      ← Main trading UI (Binance-style)
├── dashboard.html                    ← System dashboard
├── solace-trading.html               ← SOLACE consciousness trading
├── code-ide.html                     ← SOLACE code editor (Monaco)
├── chat.html                         ← Chat interface
├── memory.html                       ← Memory viewer
├── vision.html                       ← Vision interface
├── login.html                        ← Authentication
├── register.html                     ← User registration
└── health.html                       ← System health monitor
```

### `/frontend/` - React Development (NOT IN PRODUCTION)
```
frontend/
├── src/
│   ├── components/                   ← React components
│   │   ├── AdvancedOrderForm.tsx     ← Order form (fixed, 0 warnings)
│   │   ├── OpenPositionsTable.tsx    ← Positions table (fixed, 0 warnings)
│   │   ├── TradingChart.tsx
│   │   ├── Sidebar.tsx
│   │   └── *.module.css              ← CSS modules
│   ├── pages/                        ← Page components
│   │   ├── TradingDashboard.tsx
│   │   ├── SOLACEConsciousnessTrading.tsx
│   │   └── BinanceTrading.tsx
│   ├── hooks/                        ← Custom React hooks
│   ├── stores/                       ← Zustand state management
│   └── utils/                        ← Utilities
├── dist/                             ← Built React app (NOT SERVED)
├── package.json
├── vite.config.ts
└── tsconfig.json

⚠️ NOTE: This React app was built but is NOT used in production.
         Production uses static HTML in /web/ directory.
```

### `/migrations/` - Database Migrations
```
migrations/
├── 001_master_memory_system.sql
├── 004_service_registry.sql          ← Service registry (Phase 1)
├── 005_enhance_trades_for_sandbox.sql
├── 007_agent_swarm.sql                ← Agent task queue
├── 008_service_config.sql
├── 009_enhanced_observability.sql
├── 010_decision_tree_complete.sql
├── 011_grpo_learning_system.sql
└── 012_centralized_logging_system.sql
```

### `/static/` - Static Assets
```
static/
├── ace_dashboard.html                 ← ACE monitoring dashboard
└── js/
    └── solace_observer.js             ← UI observation tracking
```

### `/pkg/` - Shared Go Packages
```
pkg/
└── llm/
    ├── client.go                      ← LLM client interface
    ├── openai_client.go               ← OpenAI implementation
    ├── context_manager.go             ← Context window management
    ├── file_tools.go                  ← File operations for LLM
    └── types.go                       ← Shared types
```

### Root Scripts & Utilities
```
/ (root)
├── quick_build.ps1                    ← Fast build script (fixes 4-min hang)
├── quick_test.ps1                     ← Quick test runner
├── setup_agent_swarm.ps1              ← Agent swarm setup
├── monitor_agents.ps1                 ← Live agent monitoring
├── check_task_status.py               ← Task queue status
├── create_collaboration_test.py       ← Multi-agent test creator
├── create_ui_validation_task.py       ← SENTINEL UI test creator
├── reset_task.py                      ← Task reset utility
├── test_glassbox.ps1                  ← Glass Box testing
├── test_integration.ps1               ← Integration tests
├── docker-compose.yml                 ← Docker configuration
├── go.mod / go.sum                    ← Go dependencies
└── .env                               ← Environment configuration
```

---

## 🗄️ DATABASE ARCHITECTURE

### Tables (PostgreSQL `ares_db`)

#### Core Trading Tables:
- `trades` - Trade execution records
- `balances` - Account balances
- `market_data` - Historical market data
- `playbooks` - Trading strategies/rules

#### Consciousness & Memory:
- `memories` - SOLACE memory storage
- `conversations` - Chat history
- `consciousness_states` - SOLACE consciousness snapshots
- `thought_journal` - Decision reasoning logs
- `working_memory` - Active context

#### ACE Framework:
- `ace_reflections` - Self-reflection records
- `ace_strategies` - Generated strategies
- `ace_curations` - Curated contexts
- `ace_executions` - Strategy execution logs

#### Glass Box Transparency:
- `glass_box_traces` - Decision traces
- `merkle_batches` - Merkle tree batches
- `hedera_anchors` - Blockchain anchors

#### GRPO Learning:
- `grpo_episodes` - Learning episodes
- `grpo_rollouts` - Rollout data
- `grpo_policies` - Policy versions
- `grpo_advantages` - Advantage calculations

#### Agent Swarm:
- `task_queue` - Agent task queue
- `task_history` - Completed tasks
- `agent_registry` - Registered agents

#### Service Management:
- `service_registry` - Microservices registry (Phase 1)
- `service_config` - Dynamic configuration (Phase 4 future)
- `system_logs` - Centralized logging (Phase 4 future)

---

## 🔌 API ENDPOINTS

### Health & Monitoring:
- `GET /health` - Quick health check
- `GET /health/detailed` - Full system status
- `GET /health/services` - Service registry
- `GET /api/v1/health/llm` - LLM health (deprecated)

### Trading:
- `POST /api/v1/trading/execute` - Execute trade
- `GET /api/v1/trading/positions` - Get open positions
- `GET /api/v1/trading/history` - Trade history
- `POST /api/v1/trading/close` - Close position

### SOLACE Agent:
- `POST /api/v1/solace/chat` - Chat with SOLACE
- `GET /api/v1/solace/memory` - Retrieve memories
- `POST /api/v1/solace/command` - Execute SOLACE command
- `GET /api/v1/solace/consciousness` - Consciousness state

### Glass Box:
- `GET /api/v1/glassbox/traces` - Get decision traces
- `GET /api/v1/glassbox/verify/:trace_id` - Verify trace integrity
- `POST /api/v1/glassbox/anchor` - Anchor to Hedera

### Agent Swarm:
- `POST /api/v1/agent/task` - Create agent task
- `GET /api/v1/agent/status/:task_id` - Get task status
- `GET /api/v1/agent/results/:task_id` - Get task results

### Markets:
- `GET /api/v1/markets/available` - Available trading pairs
- `GET /api/v1/markets/price/:symbol` - Current price
- `GET /api/v1/markets/ohlcv/:symbol` - Candlestick data

---

## 🌐 EXTERNAL INTEGRATIONS

### APIs:
1. **OpenAI API** - GPT-4 (SOLACE agent)
2. **Anthropic API** - Claude 3.7 Sonnet (FORGE agent)
3. **Ollama API** - DeepSeek-R1 models (ARCHITECT, SENTINEL)
4. **CoinGecko API** - Market data (with 2-min cache)
5. **Hedera Hashgraph** - Blockchain consensus

### Services:
1. **PostgreSQL** - Primary database
2. **Redis** - Future event bus (currently in-memory)
3. **TradingView** - Chart widgets in web UI

---

## 📝 DOCUMENTATION FILES

### Specifications:
- `ARES_TRADING_UI_SPECIFICATION.md` - 500-line UI spec (created today)
- `WHAT_AGENTS_SHOULD_TEST.md` - Testing guide for agents
- `CONTRACTS.md` - Service interface contracts (550 lines)
- `ARCHITECTURE.md` - System architecture (if exists)

### Status Reports:
- `HANDOVER_TO_CLAUDE.md` - TODAY'S 10-hour session (THIS IS THE ONE YOU WANT)
- `RECOVERY_STATUS.md` - Recovery checkpoint status
- `AGENT_COLLABORATION_TEST_STATUS.md` - Agent test status
- `PHASE_1_COMPLETE_REPORT.md` - Service registry completion
- `PHASE_2_COMPLETE_REPORT.md` - Event-driven architecture
- `PHASE_3_COMPLETE_REPORT.md` - Graceful degradation

### Deployment Guides:
- `ACE_PRODUCTION_DEPLOYMENT_SUMMARY.md`
- `ARES_PRODUCTION_READINESS_ASSESSMENT.md`
- `DEPLOYMENT_COMPLETE.md`
- `HOW_TO_LAUNCH_ARES.md`
- `HOW_TO_TEST_UI.md`

---

## 🔧 BUILD & RUN

### Build Commands:
```powershell
# Quick build (recommended)
.\quick_build.ps1

# Manual build
go build -o ares_api.exe .\cmd\main.go

# Test build
go test ./...
```

### Run Commands:
```powershell
# Start ARES API
.\ares_api.exe

# Start Agent Swarm
C:\Python313\python.exe internal\agent_swarm\coordinator.py --interval 5

# Run EventBus test
go run .\cmd\test_eventbus\main.go
```

### Database Setup:
```powershell
# Initialize database
$env:PGPASSWORD='ARESISWAKING'
& 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -h localhost -U ARES -d ares_db -f init-db.sql

# Run migrations
& 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -h localhost -U ARES -d ares_db -f migrations/007_agent_swarm.sql
```

---

## 🎯 KEY ARCHITECTURAL DECISIONS

### Frontend Architecture:
**Decision:** Static HTML in `/web/` directory (NOT React)  
**Reason:** User said "react was remade to go and discommented"  
**Implication:** `/frontend/` React app exists but is NOT used in production

### Agent Swarm Architecture:
**Decision:** Python coordinator with 4 specialized agents  
**Reason:** Leverage different LLM strengths (GPT-4, Claude, DeepSeek)  
**Implication:** Requires Python 3.13 runtime alongside Go backend

### EventBus Architecture:
**Decision:** In-memory pub/sub (not Redis)  
**Reason:** Docker not available on system  
**Implication:** Events lost on restart, can upgrade to Redis later

### Database Architecture:
**Decision:** Single PostgreSQL database for all services  
**Reason:** Modularity Phase 1-3 complete, Phase 4 (microservices) not started  
**Implication:** All services share `ares_db` schema

---

## 📊 SYSTEM METRICS

### Codebase Size:
- **Go Backend:** ~15,000 lines
- **Python Agent Swarm:** ~2,000 lines
- **Static HTML/JS:** ~8,000 lines
- **React (unused):** ~5,000 lines
- **SQL Schemas:** ~3,000 lines
- **Documentation:** ~5,000 lines
- **Total:** ~38,000 lines

### Build Artifacts:
- **ares_api.exe:** 49.15 MB
- **Build Time:** ~10 seconds
- **Dependencies:** 150+ Go modules, 50+ Python packages

---

## 🔍 CRITICAL FILES FOR CLAUDE TO REVIEW

### Must Read (Priority Order):
1. **`HANDOVER_TO_CLAUDE.md`** (in root, NOT the one you're viewing) - TODAY'S SESSION
2. **`COMPLETE_SYSTEM_INVENTORY.csv`** - THIS FILE (full inventory)
3. **`cmd/main.go`** - Main entry point, understand startup flow
4. **`web/trading.html`** - Production frontend (what users see)
5. **`internal/agent_swarm/coordinator.py`** - Agent orchestration
6. **`ARES_TRADING_UI_SPECIFICATION.md`** - UI spec created today

### Database Understanding:
7. **`init-db.sql`** - Core schema
8. **`migrations/007_agent_swarm.sql`** - Agent task system

### Recent Changes:
9. **`internal/agent_swarm/coordinator.py`** line 153 - Claude model update
10. **`web/code-ide.html`** - Refactored (0 warnings)

---

## 🚨 CONFUSION TO RESOLVE

### Question 1: Frontend Architecture
**User said:** "react was remade to go and the react was discommented"  
**Current state:** 
- `/web/*.html` exists (static HTML)
- `/frontend/` exists (React, built to `/dist/`)
- `cmd/main.go` line 227 serves `./web/trading.html`

**Claude should ask:** "Is production using static HTML or React? Should I delete /frontend/?"

### Question 2: SENTINEL Test Results
**Status:** 2 UI tests completed (16s, 31s)  
**Results location:** Database `task_history` table  
**Not yet reviewed:** Need to query database to see findings

**Claude should:** Query database and review what SENTINEL found

---

## 📂 INVENTORY FILE LOCATION

**CSV File:** `C:\ARES_Workspace\ARES_API\COMPLETE_SYSTEM_INVENTORY.csv`

**Contains:**
- Full file path
- File size (bytes)
- Last modified timestamp
- File extension

**Sorted by:** Last modified time (most recent first)

**Use this to:**
- Understand complete system structure
- Identify all microservices/modules
- See what changed recently
- Track file dependencies

---

**END OF ARCHITECTURE INVENTORY**

**Next Steps:**
1. Open `COMPLETE_SYSTEM_INVENTORY.csv` in Excel/editor
2. Review file structure and timestamps
3. Read `HANDOVER_TO_CLAUDE.md` (in root) for session details
4. Query database for SENTINEL test results
5. Clarify production frontend architecture with user
