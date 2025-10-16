# ARES Service Interface Contracts
**Version:** 1.0.0  
**Last Updated:** October 16, 2025  
**Status:** Phase 1 - Modular Architecture Implementation

---

## OVERVIEW

This document defines the interface contracts for all ARES services. It specifies:
- REST API endpoints
- Database table ownership
- Future event schemas (Phase 2)
- Backward compatibility rules

---

## SERVICE: ARES_API

**Type:** Monolithic API Server  
**Port:** 8080  
**Language:** Go  
**Health Endpoint:** `GET /health`  
**Detailed Health:** `GET /health/detailed`  
**Service Registry:** `GET /health/services`

### REST API Endpoints

#### Trading Endpoints
- `POST /api/v1/trading/execute`
  - **Request:** `{symbol: string, direction: "buy"|"sell", amount: float, price?: float}`
  - **Response:** `{trade_id: string, status: string, execution_price: float}`
  - **Description:** Execute a new trade (autonomous or manual)

- `GET /api/v1/trading/open`
  - **Response:** `[{id: int, symbol: string, direction: string, entry_price: float, current_pnl: float}]`
  - **Description:** Get all open positions

- `POST /api/v1/trading/close/{trade_id}`
  - **Response:** `{closed: bool, exit_price: float, pnl: float}`
  - **Description:** Close specific trade

- `GET /api/v1/trading/status`
  - **Response:** `{trading_enabled: bool, is_running: bool, today_trades: int, today_loss: float}`
  - **Description:** Get SOLACE trading agent status

#### Health Endpoints (Phase 1 - Standardized)
- `GET /health`
  - **Response:** `{service: string, status: "healthy"|"unhealthy"}`
  - **Description:** Quick health check

- `GET /health/detailed`
  - **Response:** `{service: string, version: string, uptime_seconds: int, dependencies: object, last_check: string}`
  - **Description:** Full system status with dependency health

- `GET /health/services`
  - **Response:** `{services: [{name: string, version: string, status: string, port: int, last_heartbeat: string}], count: int}`
  - **Description:** List all registered services from service_registry table

#### SOLACE Agent Endpoints
- `POST /api/v1/solace-agent/chat`
  - **Request:** `{message: string}`
  - **Response:** `{agent: string, response: string, features: array, status: object}`
  - **Description:** Chat with SOLACE autonomous agent

- `GET /api/v1/solace-agent/memory`
  - **Response:** `{working_memory: array, persistent_memory: object, thought_journal: array}`
  - **Description:** Access SOLACE's memory systems

#### Glass Box Decision Tracing
- `GET /api/v1/glass-box/traces`
  - **Response:** `[{trace_id: int, trade_id: int, spans: array, hash_chain: string}]`
  - **Description:** Get decision traces with blockchain-style verification

- `GET /api/v1/glass-box/traces/{trace_id}`
  - **Response:** `{trace_id: int, spans: [{name: string, duration_ms: int, hash: string}]}`
  - **Description:** Get detailed decision span breakdown

#### ACE Framework Endpoints
- `POST /api/v1/ace/perception`
  - **Request:** `{layer: string, data: object}`
  - **Response:** `{processed: bool, patterns_triggered: array}`
  - **Description:** Submit data to ACE perception layer

### Database Tables (Service Contracts)

#### trades
- **Purpose:** Persistent trade storage
- **Write Access:** Trading Engine (ARES_API)
- **Read Access:** All services, Frontend
- **Schema:** See migrations/001_initial_schema.sql
- **Critical Fields:** `id, user_id, symbol, direction, entry_price, exit_price, status, pnl`

#### decision_traces
- **Purpose:** Glass Box decision logging
- **Write Access:** Glass Box System (ARES_API)
- **Read Access:** Frontend, Analytics, Audit
- **Schema:** See migrations/003_glass_box.sql
- **Critical Fields:** `trace_id, trade_id, root_hash, created_at`

#### decision_spans
- **Purpose:** Individual decision steps in trace
- **Write Access:** Glass Box System (ARES_API)
- **Read Access:** Frontend, Analytics
- **Schema:** See migrations/003_glass_box.sql
- **Critical Fields:** `span_id, trace_id, span_name, hash, previous_hash, position`

#### service_registry (NEW - Phase 1)
- **Purpose:** Service health tracking and discovery
- **Write Access:** All services (self-register on startup)
- **Read Access:** All services, Frontend, Monitoring
- **Schema:** See migrations/004_service_registry.sql
- **Critical Fields:** `name, version, status, port, health_url, last_heartbeat`
- **Update Frequency:** Heartbeat every 30 seconds

#### working_memory
- **Purpose:** SOLACE's short-term contextual memory
- **Write Access:** SOLACE Agent (ARES_API)
- **Read Access:** SOLACE Agent, Frontend
- **Schema:** See migrations/001_master_memory_system.sql
- **Critical Fields:** `id, layer, content, created_at`

#### persistent_memory
- **Purpose:** SOLACE's long-term learned knowledge
- **Write Access:** SOLACE Agent (ARES_API)
- **Read Access:** SOLACE Agent, Analytics
- **Schema:** See migrations/001_master_memory_system.sql
- **Critical Fields:** `id, content, embedding, importance, created_at`

### Dependencies

| Dependency | Type | Purpose | Fallback Behavior |
|------------|------|---------|-------------------|
| PostgreSQL (5432) | Database | Persistent storage | **CRITICAL** - No fallback (Phase 3 will add queuing) |
| ollama (11434) | HTTP | LLM inference | Cached responses (partial) |
| Hedera (testnet) | HTTP | Blockchain anchoring | Queue for retry ✅ (already implemented) |
| Redis (6379) | In-Memory | Event bus (Phase 2) | Not yet implemented |

### Health Check Response Format

**Quick Health (`GET /health`):**
```json
{
  "service": "ares-api",
  "status": "healthy"
}
```

**Detailed Health (`GET /health/detailed`):**
```json
{
  "service": "ares-api",
  "version": "1.0.0",
  "status": "healthy",
  "uptime_seconds": 3600,
  "dependencies": {
    "database": "healthy",
    "ollama": "healthy",
    "hedera": "not_configured",
    "redis": "not_configured"
  },
  "last_check": "2025-10-16T10:30:00Z"
}
```

**Service Registry (`GET /health/services`):**
```json
{
  "services": [
    {
      "id": 1,
      "name": "ares-api",
      "version": "1.0.0",
      "status": "online",
      "port": 8080,
      "health_url": "http://localhost:8080/health",
      "last_heartbeat": "2025-10-16T10:29:45Z",
      "created_at": "2025-10-16T09:00:00Z",
      "updated_at": "2025-10-16T10:29:45Z"
    },
    {
      "id": 2,
      "name": "ollama",
      "version": "0.1.0",
      "status": "offline",
      "port": 11434,
      "health_url": "http://localhost:11434/api/health",
      "last_heartbeat": null,
      "created_at": "2025-10-16T09:00:00Z",
      "updated_at": "2025-10-16T09:00:00Z"
    }
  ],
  "count": 2
}
```

---

## SERVICE: ollama

**Type:** LLM Inference Worker  
**Port:** 11434  
**Language:** Go  
**Health Endpoint:** `GET /api/health`

### Endpoints

| Endpoint | Type | Payload | Description |
|----------|------|---------|-------------|
| `POST /api/generate` | HTTP | `{model: string, prompt: string, stream: bool}` | Generate text from LLM |
| `POST /api/chat` | HTTP | `{model: string, messages: array}` | Chat-style inference |
| `GET /api/health` | HTTP | - | Health check |

### Dependencies

| Dependency | Type | Purpose | Fallback |
|------------|------|---------|----------|
| Local models | File | Model weights (DeepSeek-R1 14B) | **CRITICAL** - No fallback |
| GPU/CPU | Hardware | Inference acceleration | CPU fallback (slower) |

---

## FUTURE EVENT CONTRACTS (Phase 2 - Redis Integration)

### Events to Emit (After Redis Event Bus is Added)

#### trade_proposed_v1
**Emitted By:** Decision Engine (SOLACE)  
**Consumed By:** Trading Engine, Frontend, Glass Box  
**When:** SOLACE proposes new trade for approval

**Schema:**
```json
{
  "version": "1.0",
  "event_type": "trade_proposed",
  "timestamp": "2025-10-16T09:00:00Z",
  "data": {
    "trade_id": "uuid",
    "symbol": "HBAR/USD",
    "direction": "buy",
    "quantity": 100.0,
    "confidence": 0.85,
    "reasoning": "Technical indicator signal - RSI oversold + volume spike"
  }
}
```

#### trade_executed_v1
**Emitted By:** Trading Engine  
**Consumed By:** Glass Box, Analytics, Frontend, SOLACE  
**When:** Trade completes successfully

**Schema:**
```json
{
  "version": "1.0",
  "event_type": "trade_executed",
  "timestamp": "2025-10-16T09:00:05Z",
  "data": {
    "trade_id": "uuid",
    "symbol": "HBAR/USD",
    "status": "success",
    "executed_price": 0.0542,
    "executed_quantity": 100.0,
    "fee": 0.05,
    "glass_box_trace_id": 123
  }
}
```

#### decision_completed_v1
**Emitted By:** Glass Box System  
**Consumed By:** Learning System, Analytics, Frontend  
**When:** Decision tree completes with full audit trail

**Schema:**
```json
{
  "version": "1.0",
  "event_type": "decision_completed",
  "timestamp": "2025-10-16T09:00:03Z",
  "data": {
    "trace_id": 123,
    "decision_type": "trade_approval",
    "result": "approved",
    "confidence": 0.92,
    "ace_layer": "strategic",
    "glass_box_spans": 6,
    "duration_ms": 245,
    "hash_chain_valid": true
  }
}
```

#### market_data_updated_v1
**Emitted By:** Market Data Collector (future service)  
**Consumed By:** Trading Engine, Frontend, Analytics  
**When:** Real-time price update received

**Schema:**
```json
{
  "version": "1.0",
  "event_type": "market_data_updated",
  "timestamp": "2025-10-16T09:00:00Z",
  "data": {
    "symbol": "HBAR/USD",
    "price": 0.0542,
    "volume_24h": 1500000,
    "change_24h": 2.5,
    "source": "coinmarketcap"
  }
}
```

---

## BACKWARD COMPATIBILITY RULES

1. **Event Versions:** Increment version when schema changes (v1 → v2)
2. **Support Window:** Services MUST support previous version for 90 days minimum
3. **Adding Fields:** New optional fields are allowed in any version
4. **Removing Fields:** Requires new major version (breaking change)
5. **Deprecation Notice:** 30 days advance warning before removing endpoints
6. **Migration Path:** Provide migration guide when breaking changes are introduced

**Example Versioning:**
- v1.0: Initial release
- v1.1: Add optional fields (backward compatible)
- v2.0: Remove fields or change required fields (breaking change)

---

## SERVICE DEPENDENCIES DIAGRAM

```
Frontend (React SPA - Phase 1 not yet built)
  ↓
ARES_API (Port 8080)
  ├── PostgreSQL (Port 5432) [CRITICAL]
  ├── ollama (Port 11434) [OPTIONAL - cached fallback]
  ├── Hedera Testnet [OPTIONAL - queue for retry]
  └── Redis (Port 6379) [PHASE 2 - not yet implemented]
```

**Current Architecture:** Monolithic ARES_API with embedded services  
**Phase 2 Goal:** Event-driven microservices with Redis pub/sub  
**Phase 3 Goal:** Graceful degradation (trades continue if Glass Box fails)

---

## ADDING NEW SERVICES (Checklist)

When adding a new service to ARES:

1. ✅ **Register in `service_registry` table** on startup
2. ✅ **Implement `/health` endpoint** (return `{service: string, status: string}`)
3. ✅ **Send heartbeat every 30s** to update `last_heartbeat` column
4. ✅ **Document in this file** (CONTRACTS.md) - list endpoints, events, dependencies
5. ✅ **Define event schemas** (if emitting/consuming events in Phase 2)
6. ✅ **Add to dependency diagram** above
7. ✅ **Test graceful degradation** (what happens if this service fails?)

**Example Registration Code (Go):**
```go
import "ares_api/internal/registry"

func main() {
    db := database.InitDB()
    
    // Register service
    registry.RegisterService(db, "my-service", "1.0.0", 9000, "http://localhost:9000/health")
    
    // Start heartbeat
    go registry.ServiceHeartbeat(db, "my-service", 30*time.Second)
    
    // ... start service ...
}
```

---

## APPENDIX: GLASS BOX DECISION SPAN NAMES

Standard span names used in Glass Box decision tracing:

1. `authorization_check` - User/agent authorization validation
2. `input_validation` - Request parameter validation
3. `market_pricing` - Real-time price lookup
4. `balance_check` - Sufficient funds verification
5. `trade_execution` - Actual trade execution
6. `database_persistence` - Save trade to PostgreSQL
7. `hedera_anchoring` - Blockchain hash anchoring (Phase 4)

Each span includes: `span_name`, `duration_ms`, `hash` (SHA-256), `previous_hash`, `position` (0-6)

---

## VERSION HISTORY

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0.0 | 2025-10-16 | Initial contracts documentation (Phase 1) | GitHub Copilot |
| 1.0.0 | 2025-10-16 | Added service_registry table contract | GitHub Copilot |
| 1.0.0 | 2025-10-16 | Standardized health endpoints | GitHub Copilot |

---

**Status:** Phase 1 Complete ✅  
**Next Phase:** Redis event bus implementation (Phase 2)  
**Contact:** SOLACE Agent (http://localhost:8080/api/v1/solace-agent/chat)
