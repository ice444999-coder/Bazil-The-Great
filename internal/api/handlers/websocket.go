package handlers

import (
	"ares_api/internal/eventbus"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a real-time update message
type WebSocketMessage struct {
	Type      string                 `json:"type"`      // "signal", "trade", "metrics", "alert"
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan WebSocketMessage
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.RWMutex
	eventBus   *eventbus.EventBus
}

// WebSocketClient represents a connected client
type WebSocketClient struct {
	hub  *WebSocketHub
	conn *websocket.Conn
	send chan WebSocketMessage
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// TODO: Restrict origins in production
		return true
	},
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(eventBus *eventbus.EventBus) *WebSocketHub {
	hub := &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan WebSocketMessage, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		eventBus:   eventBus,
	}

	// Subscribe to EventBus topics
	hub.subscribeToEventBus()

	return hub
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	log.Println("[WEBSOCKET] Hub started")

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WEBSOCKET] Client registered (total: %d)", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[WEBSOCKET] Client unregistered (total: %d)", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client buffer full, disconnect
					h.mu.RUnlock()
					h.mu.Lock()
					close(client.send)
					delete(h.clients, client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// subscribeToEventBus subscribes to all relevant EventBus topics
func (h *WebSocketHub) subscribeToEventBus() {
	// Subscribe to strategy signals
	h.eventBus.Subscribe("strategy.RSI_Oversold.signal", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "signal",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	h.eventBus.Subscribe("strategy.MACD_Divergence.signal", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "signal",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	h.eventBus.Subscribe("strategy.Trend_Following.signal", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "signal",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	h.eventBus.Subscribe("strategy.Support_Bounce.signal", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "signal",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	h.eventBus.Subscribe("strategy.Volume_Spike.signal", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "signal",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	// Subscribe to strategy metrics
	h.eventBus.Subscribe("strategy.RSI_Oversold.metrics", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "metrics",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	h.eventBus.Subscribe("strategy.MACD_Divergence.metrics", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "metrics",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	h.eventBus.Subscribe("strategy.Trend_Following.metrics", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "metrics",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	h.eventBus.Subscribe("strategy.Support_Bounce.metrics", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "metrics",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	h.eventBus.Subscribe("strategy.Volume_Spike.metrics", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "metrics",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	// Subscribe to trade executions
	h.eventBus.Subscribe("trade.executed", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "trade",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	// Subscribe to trade closures
	h.eventBus.Subscribe("trade.closed", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "trade_closed",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	// Subscribe to auto-graduate events
	h.eventBus.Subscribe("strategy.promoted", func(jsonData []byte) {
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return
		}

		message := WebSocketMessage{
			Type:      "alert",
			Timestamp: time.Now(),
			Data:      data,
		}

		h.broadcast <- message
	})

	log.Println("[WEBSOCKET] Subscribed to EventBus topics")
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WEBSOCKET][ERROR] Upgrade failed: %v", err)
		return
	}

	client := &WebSocketClient{
		hub:  h,
		conn: conn,
		send: make(chan WebSocketMessage, 256),
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()

	// Send initial connection success message
	welcomeMsg := WebSocketMessage{
		Type:      "connected",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message": "Connected to ARES real-time updates",
			"topics": []string{
				"strategy.*.signal",
				"strategy.*.metrics",
				"trade.executed",
				"trade.closed",
				"alert.*",
			},
		},
	}
	client.send <- welcomeMsg
}

// readPump handles reading from the WebSocket connection
func (c *WebSocketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WEBSOCKET][ERROR] Unexpected close: %v", err)
			}
			break
		}

		// Handle incoming messages (e.g., subscription requests)
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			// Process subscription requests, filters, etc.
			if msgType, ok := msg["type"].(string); ok {
				switch msgType {
				case "ping":
					pongMsg := WebSocketMessage{
						Type:      "pong",
						Timestamp: time.Now(),
						Data:      map[string]interface{}{},
					}
					c.send <- pongMsg
				case "subscribe":
					// Handle subscription to specific topics
					// TODO: Implement topic filtering
				}
			}
		}
	}
}

// writePump handles writing to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send JSON message
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("[WEBSOCKET][ERROR] Write failed: %v", err)
				return
			}

		case <-ticker.C:
			// Send ping
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// BroadcastMessage sends a message to all connected clients
func (h *WebSocketHub) BroadcastMessage(msgType string, data map[string]interface{}) {
	message := WebSocketMessage{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}
	h.broadcast <- message
}

// GetClientCount returns the number of connected clients
func (h *WebSocketHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
