package controllers

import (
	"ares_api/internal/websocket"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	gorilla_websocket "github.com/gorilla/websocket"
)

var wsUpgrader = gorilla_websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(c *gin.Context) {
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	client := websocket.NewClient(conn)
	hub := websocket.GetGlobalHub()
	hub.RegisterClient(client)

	// Send welcome message to this client
	welcomeMsg := websocket.Message{
		Type: "connected",
		Data: map[string]interface{}{
			"message": "Connected to ARES WebSocket",
			"time":    time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now(),
	}
	jsonData, _ := json.Marshal(welcomeMsg)
	client.Send <- jsonData

	go client.WritePump()
	go client.ReadPump()
}

// BroadcastTradeExecution sends trade execution event to all connected clients
func BroadcastTradeExecution(tradeID string, symbol, side string, amount, price float64) {
	websocket.BroadcastTradeExecution(tradeID, symbol, side, amount, price)
}

// BroadcastPriceUpdate sends price update to all connected clients
func BroadcastPriceUpdate(symbol string, price float64, change float64) {
	websocket.BroadcastPriceUpdate(symbol, price, change)
}

// BroadcastSOLACEDecision sends SOLACE decision to all connected clients
func BroadcastSOLACEDecision(decision string, confidence float64, symbol string) {
	websocket.BroadcastSOLACEDecision(decision, confidence, symbol)
}
