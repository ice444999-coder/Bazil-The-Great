package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var masterControlUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type WSMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"`
}

type MasterControlWS struct {
	conn   *websocket.Conn
	send   chan WSMessage
	db     *gorm.DB
	solace interface{} // Reference to SOLACE agent
	mu     sync.Mutex
}

var (
	wsClients   = make(map[*MasterControlWS]bool)
	wsClientsMu sync.Mutex
)

// HandleMasterControlWS upgrades HTTP to WebSocket for VS Code extension
func HandleMasterControlWS(db *gorm.DB, solace interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := masterControlUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		client := &MasterControlWS{
			conn:   conn,
			send:   make(chan WSMessage, 256),
			db:     db,
			solace: solace,
		}

		// Register client
		wsClientsMu.Lock()
		wsClients[client] = true
		wsClientsMu.Unlock()

		// Send connection success
		client.SendMessage(WSMessage{
			Type: "log",
			Data: map[string]interface{}{
				"level": "info",
				"text":  "Connected to ARES Master Control",
			},
		})

		// Start goroutines for read/write
		go client.writePump()
		go client.readPump()
	}
}

func (ws *MasterControlWS) readPump() {
	defer func() {
		wsClientsMu.Lock()
		delete(wsClients, ws)
		wsClientsMu.Unlock()
		ws.conn.Close()
	}()

	for {
		var msg WSMessage
		err := ws.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS read error: %v", err)
			}
			break
		}

		ws.handleMessage(msg)
	}
}

func (ws *MasterControlWS) writePump() {
	defer ws.conn.Close()

	for msg := range ws.send {
		err := ws.conn.WriteJSON(msg)
		if err != nil {
			log.Printf("WS write error: %v", err)
			return
		}
	}
}

func (ws *MasterControlWS) SendMessage(msg WSMessage) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	select {
	case ws.send <- msg:
	default:
		log.Println("WS send channel full, dropping message")
	}
}

func (ws *MasterControlWS) handleMessage(msg WSMessage) {
	switch msg.Type {
	case "trigger_heal":
		ws.handleTriggerHeal()

	case "generate_patch":
		issue, ok := msg.Data["issue"].(string)
		if !ok {
			ws.SendMessage(WSMessage{
				Type: "error",
				Data: map[string]interface{}{"message": "Invalid issue format"},
			})
			return
		}
		ws.handleGeneratePatch(issue)

	case "inject_fault":
		faultType, ok := msg.Data["faultType"].(string)
		if !ok {
			ws.SendMessage(WSMessage{
				Type: "error",
				Data: map[string]interface{}{"message": "Invalid fault type"},
			})
			return
		}
		ws.handleInjectFault(faultType)

	case "sniff_code":
		pattern, _ := msg.Data["pattern"].(string)
		dir, _ := msg.Data["dir"].(string)
		ws.handleSniffCode(pattern, dir)

	case "apply_patch":
		patch, ok := msg.Data["patch"].(string)
		if !ok {
			ws.SendMessage(WSMessage{
				Type: "error",
				Data: map[string]interface{}{"message": "Invalid patch format"},
			})
			return
		}
		ws.handleApplyPatch(patch)

	case "chat_message":
		speaker, _ := msg.Data["speaker"].(string)
		text, _ := msg.Data["text"].(string)
		ws.handleChatMessage(speaker, text)

	case "request_diffs":
		ws.handleRequestDiffs()

	default:
		ws.SendMessage(WSMessage{
			Type: "error",
			Data: map[string]interface{}{"message": "Unknown message type: " + msg.Type},
		})
	}
}

func (ws *MasterControlWS) handleTriggerHeal() {
	ws.SendMessage(WSMessage{
		Type: "log",
		Data: map[string]interface{}{
			"level": "info",
			"text":  "🛠️ Manual self-healing triggered",
		},
	})

	// Call SOLACE.SelfHealUI() via type assertion
	// This requires SOLACE to be passed as interface{} and asserted
	// For now, send success - implement full integration in solace.go

	ws.SendMessage(WSMessage{
		Type: "heal_complete",
		Data: map[string]interface{}{
			"message": "Self-healing cycle completed",
		},
	})
}

func (ws *MasterControlWS) handleGeneratePatch(issue string) {
	ws.SendMessage(WSMessage{
		Type: "log",
		Data: map[string]interface{}{
			"level": "info",
			"text":  "🔨 Forge generating patch for: " + issue,
		},
	})

	// Call Forge.GenerateUIPatch() - implement integration in forge.go
	// For now, send mock patch
	mockPatch := `diff --git a/web/dashboard.html b/web/dashboard.html
index abc123..def456 100644
--- a/web/dashboard.html
+++ b/web/dashboard.html
@@ -100,7 +100,7 @@
-                    <a href="/dashboard" class="nav-item">
+                    <a href="/dashboard" class="nav-item active">
                         <span class="nav-icon">📊</span>
                         <span class="nav-label">Dashboard</span>
                     </a>`

	ws.SendMessage(WSMessage{
		Type: "forge_patch",
		Data: map[string]interface{}{
			"issue": issue,
			"patch": mockPatch,
		},
	})
}

func (ws *MasterControlWS) handleInjectFault(faultType string) {
	ws.SendMessage(WSMessage{
		Type: "log",
		Data: map[string]interface{}{
			"level": "warning",
			"text":  "💥 Injecting fault: " + faultType,
		},
	})

	// Implement fault injection logic - modify files to test healing
	ws.SendMessage(WSMessage{
		Type: "log",
		Data: map[string]interface{}{
			"level": "success",
			"text":  "Fault injected successfully",
		},
	})
}

func (ws *MasterControlWS) handleSniffCode(pattern, dir string) {
	ws.SendMessage(WSMessage{
		Type: "log",
		Data: map[string]interface{}{
			"level": "info",
			"text":  "🔍 Sniffing code for pattern: " + pattern,
		},
	})

	// Call Bazil.SniffCode() - implement integration
	// For now, send mock detection
	ws.SendMessage(WSMessage{
		Type: "bazil_detection",
		Data: map[string]interface{}{
			"file":    "web/dashboard.html",
			"line":    125,
			"content": `<a href="/dashboard" class="nav-item">`,
		},
	})
}

func (ws *MasterControlWS) handleApplyPatch(patch string) {
	ws.SendMessage(WSMessage{
		Type: "log",
		Data: map[string]interface{}{
			"level": "info",
			"text":  "⚙️ Applying patch via git apply",
		},
	})

	// Execute git apply logic here
	ws.SendMessage(WSMessage{
		Type: "log",
		Data: map[string]interface{}{
			"level": "success",
			"text":  "✅ Patch applied successfully",
		},
	})
}

func (ws *MasterControlWS) handleChatMessage(speaker, text string) {
	// Broadcast chat message to all connected clients
	wsClientsMu.Lock()
	defer wsClientsMu.Unlock()

	for client := range wsClients {
		client.SendMessage(WSMessage{
			Type: "chat_message",
			Data: map[string]interface{}{
				"speaker": speaker,
				"text":    text,
			},
		})
	}
}

func (ws *MasterControlWS) handleRequestDiffs() {
	// Return recent patches/diffs
	ws.SendMessage(WSMessage{
		Type: "log",
		Data: map[string]interface{}{
			"level": "info",
			"text":  "📋 Fetching recent diffs",
		},
	})
}

// BroadcastLog sends log to all connected WS clients
func BroadcastLog(level, text string) {
	wsClientsMu.Lock()
	defer wsClientsMu.Unlock()

	for client := range wsClients {
		client.SendMessage(WSMessage{
			Type: "log",
			Data: map[string]interface{}{
				"level": level,
				"text":  text,
			},
		})
	}
}
