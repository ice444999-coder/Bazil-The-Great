# 🔍 SYSTEM HEALTH DATA SOURCE ANALYSIS

**Date:** October 12, 2025  
**Issue:** UI showing static health data  
**Root Cause:** Health endpoint returns hardcoded/placeholder values, not real database queries

---

## 📊 Current Health Data Source

### Location
**File:** `c:\ARES_Workspace\ARES_API\internal\api\controllers\monitoring_controller.go`  
**Method:** `GetHealth(c *gin.Context)` (Lines 33-91)

### Data Flow
```
UI Dashboard (health.html)
    ↓ HTTP GET
/api/v1/monitoring/health
    ↓
MonitoringController.GetHealth()
    ↓
metrics.GetSnapshot() ← IN-MEMORY ONLY (NOT SQL)
    ↓
Returns gin.H{...} with HARDCODED values
```

---

## ❌ STATIC VALUES IDENTIFIED

### 1. **LLM Engine Card**
```go
"llm": gin.H{
    "model":               "DeepSeek-R1 14B",        // ✅ OK (static config)
    "connected":           health.Checks["llm"].Status == "pass",  // ✅ OK (real check)
    "latency":             int(snapshot.LLMAvgLatencyMs),          // ✅ OK (real metric)
    "requests_per_minute": 0,                        // ❌ HARDCODED 0
},
```
**Issue:** RPM always shows 0  
**Should Pull From:** Calculate from `snapshot.LLMRequests` over time window

---

### 2. **PostgreSQL Card**
```go
"database": gin.H{
    "connected":          true,  // ❌ HARDCODED true
    "pgvector_installed": true,  // ❌ HARDCODED true
    "avg_query_time":     5,     // ❌ HARDCODED 5ms
    "active_connections": snapshot.DBConnections,  // ⚠️ IN-MEMORY (not from SQL)
},
```
**Issue:** All values hardcoded except DBConnections  
**Should Pull From:**
- `SELECT 1` query to verify DB connection
- `SELECT * FROM pg_extension WHERE extname = 'vector'` to verify pgvector
- Track actual query times in middleware
- `SELECT COUNT(*) FROM pg_stat_activity WHERE datname = 'ares_platform'` for connections

---

### 3. **Trading Engine Card**
```go
"trading": gin.H{
    "enabled":        true,  // ❌ HARDCODED true
    "open_positions": snapshot.ActiveTrades,  // ✅ OK (real count)
    "daily_loss":     0.0,   // ❌ HARDCODED 0
    "max_loss_limit": 500.0, // ✅ OK (config value)
},
```
**Issue:** daily_loss always 0, enabled always true  
**Should Pull From SQL:**
```sql
-- Daily P&L
SELECT COALESCE(SUM(profit_loss), 0) as daily_loss
FROM sandbox_trades
WHERE user_id = 1  -- SOLACE
  AND opened_at >= CURRENT_DATE
  AND status = 'CLOSED';

-- Trading enabled flag
SELECT trading_enabled 
FROM ares_config 
WHERE id = 1;
```

---

### 4. **ACE Framework Card**
```go
"ace": gin.H{
    "total_rules":    0,      // ❌ HARDCODED 0
    "active_rules":   0,      // ❌ HARDCODED 0
    "avg_confidence": 0.0,    // ❌ HARDCODED 0
    "last_learning":  "Never", // ❌ HARDCODED "Never"
},
```
**Issue:** Entire ACE section hardcoded  
**Should Pull From SQL:**
```sql
-- Total rules in playbook
SELECT COUNT(*) as total_rules 
FROM ace_playbook;

-- Active rules (helpful > harmful)
SELECT COUNT(*) as active_rules 
FROM ace_playbook 
WHERE helpful_count > harmful_count;

-- Average confidence
SELECT AVG(confidence) as avg_confidence 
FROM ace_playbook 
WHERE helpful_count > harmful_count;

-- Last learning event
SELECT MAX(created_at) as last_learning 
FROM ace_playbook;
```

---

## 🗄️ WHERE DATA SHOULD COME FROM

### Database Tables Available

#### 1. **sandbox_trades** (Trading Data)
```sql
-- Schema location: internal/database/migrations/004_autonomous_trading_system.sql
-- Contains: All SOLACE's trades with P&L, status, timestamps

SELECT 
    COUNT(*) FILTER (WHERE status = 'OPEN') as open_positions,
    SUM(profit_loss) FILTER (WHERE opened_at >= CURRENT_DATE AND status = 'CLOSED') as daily_pnl,
    AVG(confidence_score) as avg_confidence
FROM sandbox_trades
WHERE user_id = 1;  -- SOLACE's user ID
```

#### 2. **ace_playbook** (ACE Framework)
```sql
-- Schema location: internal/database/migrations/004_autonomous_trading_system.sql
-- Contains: Rules SOLACE learns from experience

SELECT 
    COUNT(*) as total_rules,
    COUNT(*) FILTER (WHERE helpful_count > harmful_count) as active_rules,
    AVG(confidence) as avg_confidence,
    MAX(created_at) as last_learning
FROM ace_playbook;
```

#### 3. **memory_snapshots** (SOLACE's Memory)
```sql
-- Schema location: internal/database/migrations/003_semantic_memory_architecture.sql
-- Contains: Episodic, semantic, working memories

SELECT COUNT(*) FROM memory_snapshots;
SELECT COUNT(*) FROM conversation_history;
```

#### 4. **PostgreSQL System Catalogs**
```sql
-- Check pgvector extension
SELECT COUNT(*) FROM pg_extension WHERE extname = 'vector';

-- Active connections
SELECT COUNT(*) FROM pg_stat_activity WHERE datname = 'ares_platform';

-- Database size
SELECT pg_database_size('ares_platform') / 1024 / 1024 as size_mb;
```

---

## 📡 Swagger API Documentation

### Status: ✅ **SWAGGER IS CONFIGURED**

**Location:** `c:\ARES_Workspace\ARES_API\cmd\main.go` (Lines 53-68)  
**Endpoint:** `http://localhost:8080/swagger/index.html`

**Configuration:**
```go
docs.SwaggerInfo.Title = "ARES Platform API"
docs.SwaggerInfo.Description = "API documentation for the ARES Platform service."
docs.SwaggerInfo.Version = "1.0"
docs.SwaggerInfo.BasePath = "/api/v1"
```

**Generated Docs:** `c:\ARES_Workspace\ARES_API\internal\docs\docs.go`

**To Access:**
1. Start ARES: `.\ARES.exe`
2. Open browser: `http://localhost:8080/swagger/index.html`
3. View all API endpoints with schemas

**Health Endpoint in Swagger:**
```
GET /api/v1/monitoring/health
@Summary Get system health
@Description Returns health status including circuit breaker state, error rate, and uptime
@Produce json
@Success 200 {object} map[string]interface{}
@Failure 503 {object} map[string]interface{}
```

---

## 🔧 FIXES NEEDED

### Priority 1: Connect Database Queries

**File to Modify:** `internal/api/controllers/monitoring_controller.go`

**Add Database Queries:**
```go
func (mc *MonitoringController) GetHealth(c *gin.Context) {
    // Get snapshot (in-memory metrics)
    snapshot := mc.metrics.GetSnapshot()
    
    // TODO: Query PostgreSQL for real data
    var dbConnected bool
    var pgvectorInstalled bool
    var dailyLoss float64
    var totalRules, activeRules int
    var avgConfidence float64
    var lastLearning time.Time
    
    // DB Connection Check
    db := mc.db  // Need to inject *gorm.DB into MonitoringController
    if err := db.Exec("SELECT 1").Error; err == nil {
        dbConnected = true
    }
    
    // pgvector Check
    db.Raw("SELECT COUNT(*) FROM pg_extension WHERE extname = 'vector'").Scan(&pgvectorInstalled)
    
    // Daily P&L
    db.Raw(`
        SELECT COALESCE(SUM(profit_loss), 0) 
        FROM sandbox_trades 
        WHERE user_id = 1 
          AND opened_at >= CURRENT_DATE 
          AND status = 'CLOSED'
    `).Scan(&dailyLoss)
    
    // ACE Stats
    db.Raw("SELECT COUNT(*) FROM ace_playbook").Scan(&totalRules)
    db.Raw("SELECT COUNT(*) FROM ace_playbook WHERE helpful_count > harmful_count").Scan(&activeRules)
    db.Raw("SELECT COALESCE(AVG(confidence), 0) FROM ace_playbook").Scan(&avgConfidence)
    db.Raw("SELECT MAX(created_at) FROM ace_playbook").Scan(&lastLearning)
    
    // Use real data in response
    response := gin.H{
        "llm": gin.H{
            "connected": health.Checks["llm"].Status == "pass",
            "latency": int(snapshot.LLMAvgLatencyMs),
            "requests_per_minute": calculateRPM(snapshot),  // New function
        },
        "database": gin.H{
            "connected": dbConnected,  // REAL
            "pgvector_installed": pgvectorInstalled > 0,  // REAL
            "active_connections": snapshot.DBConnections,
        },
        "trading": gin.H{
            "enabled": true,
            "open_positions": snapshot.ActiveTrades,
            "daily_loss": dailyLoss,  // REAL
        },
        "ace": gin.H{
            "total_rules": totalRules,  // REAL
            "active_rules": activeRules,  // REAL
            "avg_confidence": avgConfidence,  // REAL
            "last_learning": formatTime(lastLearning),  // REAL
        },
        ...
    }
    
    c.JSON(200, response)
}
```

---

### Priority 2: Track Query Performance

**Add Middleware:** `internal/api/middleware/query_timer.go`
```go
func QueryTimerMiddleware(db *gorm.DB) {
    db.Callback().Query().Before("gorm:query").Register("query_timer:before", func(db *gorm.DB) {
        db.InstanceSet("query_start_time", time.Now())
    })
    
    db.Callback().Query().After("gorm:query").Register("query_timer:after", func(db *gorm.DB) {
        if startTime, ok := db.InstanceGet("query_start_time"); ok {
            duration := time.Since(startTime.(time.Time))
            // Log to metrics
            metrics.RecordDBQuery(duration.Milliseconds())
        }
    })
}
```

---

### Priority 3: Calculate LLM RPM

**Add to Metrics:**
```go
func (m *Metrics) CalculateRPM() float64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    // Count requests in last minute
    oneMinuteAgo := time.Now().Add(-1 * time.Minute)
    
    // Need to track request timestamps
    // For now, estimate from total requests / uptime * 60
    uptime := time.Since(m.StartTime).Minutes()
    if uptime == 0 {
        return 0
    }
    
    return float64(m.LLMRequests) / uptime
}
```

---

## 📋 IMPLEMENTATION CHECKLIST

- [ ] **Inject DB into MonitoringController**
  - Modify `NewMonitoringController()` to accept `*gorm.DB`
  - Update routes setup in `v1.go`

- [ ] **Add DB Connection Health Check**
  - Query `SELECT 1` to verify connectivity
  - Catch errors and set `connected: false`

- [ ] **Add pgvector Verification**
  - Query `pg_extension` table
  - Return actual installed status

- [ ] **Query Daily Trading P&L**
  - Sum `profit_loss` for today's closed trades
  - Filter by SOLACE's user_id (1)

- [ ] **Query ACE Framework Stats**
  - Total rules count
  - Active rules (helpful > harmful)
  - Average confidence
  - Last learning timestamp

- [ ] **Calculate LLM RPM**
  - Track request timestamps in metrics
  - Count requests in last 60 seconds

- [ ] **Track DB Query Times**
  - Add GORM middleware to measure queries
  - Store avg_query_time in metrics

- [ ] **Update Swagger Docs**
  - Regenerate with `swag init`
  - Document new response schema

- [ ] **Test All Health Metrics**
  - Verify each field shows real data
  - Check UI updates correctly

---

## 🎯 EXPECTED RESULT

After fixes, System Health page will show:

**LLM Engine:**
- Model: DeepSeek-R1 14B ✅
- Status: Connected (if Ollama running) ✅
- Latency: Real average from metrics ✅
- RPM: Calculated from actual requests ✅

**PostgreSQL:**
- Connection: Real ping test ✅
- pgvector: Queried from pg_extension ✅
- Avg Query Time: Tracked via middleware ✅
- Active Connections: From pg_stat_activity ✅

**Trading Engine:**
- Enabled: From ares_config ✅
- Open Positions: From sandbox_trades WHERE status='OPEN' ✅
- Daily Loss: SUM(profit_loss) for today ✅
- Max Loss Limit: $500 (config) ✅

**ACE Framework:**
- Total Rules: COUNT(*) FROM ace_playbook ✅
- Active Rules: COUNT WHERE helpful > harmful ✅
- Avg Confidence: AVG(confidence) ✅
- Last Learning: MAX(created_at) formatted ✅

---

## 🌐 Swagger Access

**URL:** http://localhost:8080/swagger/index.html  
**Status:** ✅ Configured and working  
**Regenerate:** `swag init` in ARES_API directory (commented out in main.go)
