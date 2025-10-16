# Missing API Endpoints for UI - Fix List

## ✅ Already Working (No Changes Needed)
- `POST /api/v1/users/signup` - Registration ✓
- `POST /api/v1/users/login` - Login ✓
- `GET /api/v1/trading/performance` - Trading performance stats ✓
- `GET /api/v1/trading/playbook/stats` - ACE playbook stats ✓
- `GET /api/v1/trading/history` - Trade history ✓
- `GET /api/v1/trading/open` - Open positions ✓
- `POST /api/v1/trading/execute` - Execute trade ✓
- `POST /api/v1/trading/close` - Close single position ✓
- `POST /api/v1/trading/close-all` - Close all positions ✓
- `GET /api/v1/trading/playbook/` - Get playbook rules ✓
- `GET /api/v1/chat/history` - Chat history ✓
- `POST /api/v1/chat/send` - Send chat message ✓

## ⚠️ Needs Attention

### 1. Agent Chat Endpoint (UI expects different path)
**UI calls:** `POST /api/v1/agent/chat`  
**Backend has:** `POST /api/v1/chat/send`

**Fix:** Add route alias or update UI to use `/api/v1/chat/send`

### 2. Agent Chat History (UI expects different path)
**UI calls:** `GET /api/v1/agent/chat/history`  
**Backend has:** `GET /api/v1/chat/history`

**Fix:** Add route alias or update UI to use `/api/v1/chat/history`

### 3. Memory Snapshots Endpoint (Missing)
**UI calls:** `GET /api/v1/memory/snapshots`  
**Backend has:** `POST /api/v1/memory/learn` and `GET /api/v1/memory/recall`

**Fix:** Need to add `GET /api/v1/memory/snapshots` to return all memory snapshots

### 4. Memory Recall (Wrong method)
**UI calls:** `POST /api/v1/memory/recall` (with JSON body containing query)  
**Backend has:** `GET /api/v1/memory/recall`

**Fix:** Change to POST or update UI to use GET with query params

### 5. Monitoring Logs Endpoint (Missing)
**UI calls:** `GET /api/v1/monitoring/logs?limit=10`  
**Backend has:** No logs endpoint

**Fix:** Need to add logs endpoint to return recent system logs

### 6. User Profile Endpoint (Partially missing)
**UI calls:** `GET /api/v1/users/profile` (for dashboard username)  
**Backend has:** Auth controller created but route not registered

**Fix:** Add route for getting user profile

## 🔧 Quick Fixes Needed

1. Add `/agent/chat` route aliases
2. Add `GET /api/v1/memory/snapshots` endpoint
3. Change memory recall to POST or update UI
4. Add logging endpoint
5. Register user profile route
