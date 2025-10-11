# 🚀 ARES - START HERE

## ✅ Everything is Ready!

Your ARES system now has **Semantic Memory** with intelligent retrieval powered by local Ollama embeddings.

---

## 🎯 HOW TO START ARES

### Option 1: Double-click the Batch File (EASIEST)
```
📁 C:\ARES_Workspace\ARES_API\START_ARES.bat
```
**Just double-click this file** - it will:
1. Kill any old ARES processes
2. Start ARES.exe
3. Keep the window open so you can see logs
4. Show background worker activity

### Option 2: Double-click ARES.exe Directly
```
📁 C:\ARES_Workspace\ARES_API\ARES.exe
```
**Note:** The window will close if there's an error (like port already in use).
Use START_ARES.bat instead for better visibility.

---

## 🌐 Access ARES

Once started, ARES is available at:

- **API**: http://localhost:8080
- **Swagger Docs**: http://localhost:8080/swagger/index.html
- **Code Editor**: http://localhost:8080/static/editor.html

---

## 🧠 What's Running

### Background Workers (Automatic):
1. **Trade Processor** - Executes limit orders every 10 seconds
2. **Embedding Processor** - Generates memory embeddings every 30 seconds

You'll see console output like:
```
📊 Processed 25 memory embeddings
```

### New Semantic Memory Features:
- ✅ Local embeddings (no API costs)
- ✅ Intelligent memory retrieval
- ✅ Hot/warm/cold memory hierarchy
- ✅ Scalable to millions of memories

---

## 🧪 Test Semantic Search

### 1. Login to get your token:
```bash
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"YOUR_USERNAME\",\"password\":\"YOUR_PASSWORD\"}"
```

### 2. Chat with Claude (creates memories):
```bash
curl -X POST http://localhost:8080/api/v1/claude/chat \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"message\":\"I want to trade Bitcoin using market orders\"}"
```

### 3. Wait 30-60 seconds for embeddings to process

### 4. Search semantically:
```bash
curl -X POST http://localhost:8080/api/v1/claude/semantic-search \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"query\":\"What trading strategies did I discuss?\",\"limit\":5}"
```

---

## 📚 Documentation

- **Semantic Memory Guide**: `SEMANTIC_MEMORY_GUIDE.md`
- **API Endpoints**: http://localhost:8080/swagger/index.html

---

## 🔧 Troubleshooting

### ARES won't start - "Port 8080 already in use"
**Solution:** Use `START_ARES.bat` - it automatically kills old processes

### Embeddings not generating?
**Check:**
1. Ollama is running: `ollama list` (should show `nomic-embed-text`)
2. ARES background worker is active (you'll see console output every 30 sec)

### Need to stop ARES?
**Press:** `Ctrl + C` in the terminal window

---

## 🎉 You're All Set!

Just run **START_ARES.bat** and you have a fully functional AI-powered trading system with semantic memory!
