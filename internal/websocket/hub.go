/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	Send chan []byte
}

type Message struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

var globalHub *Hub

func init() {
	globalHub = &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go globalHub.Run()
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastMessage(messageType string, data map[string]interface{}) {
	message := Message{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling WebSocket message: %v", err)
		return
	}
	h.broadcast <- jsonData
}

func (c *Client) ReadPump() {
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
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		log.Printf("Received WebSocket message: %s", message)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetGlobalHub returns the global WebSocket hub
func GetGlobalHub() *Hub {
	return globalHub
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		hub:  globalHub,
		conn: conn,
		Send: make(chan []byte, 256),
	}
}

// BroadcastTradeExecution sends trade execution event to all connected clients
func BroadcastTradeExecution(tradeID string, symbol, side string, amount, price float64) {
	globalHub.BroadcastMessage("trade_executed", map[string]interface{}{
		"trade_id": tradeID,
		"symbol":   symbol,
		"side":     side,
		"amount":   amount,
		"price":    price,
	})
}

// BroadcastPriceUpdate sends price update to all connected clients
func BroadcastPriceUpdate(symbol string, price float64, change float64) {
	globalHub.BroadcastMessage("price_update", map[string]interface{}{
		"symbol": symbol,
		"price":  price,
		"change": change,
	})
}

// BroadcastSOLACEDecision sends SOLACE decision to all connected clients
func BroadcastSOLACEDecision(decision string, confidence float64, symbol string) {
	globalHub.BroadcastMessage("solace_decision", map[string]interface{}{
		"decision":   decision,
		"confidence": confidence,
		"symbol":     symbol,
	})
}
