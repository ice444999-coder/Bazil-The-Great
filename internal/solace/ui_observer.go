/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package solace

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ObservationMessage - Message format from browser
type ObservationMessage struct {
	Type        string          `json:"type"` // 'ui_state', 'data_stream', 'user_action', 'market_context'
	Timestamp   time.Time       `json:"timestamp"`
	SessionID   string          `json:"sessionId"`
	Component   string          `json:"component"`
	ElementID   string          `json:"elementId"`
	Data        json.RawMessage `json:"data"`
	UserVisible bool            `json:"userVisible"`
}

// SolaceCommand - Commands SOLACE sends to browser
type SolaceCommand struct {
	Type      string          `json:"type"` // 'inject_javascript', 'modify_css', 'execute_trade', 'show_alert'
	CommandID string          `json:"commandId"`
	Code      string          `json:"code,omitempty"`
	Selector  string          `json:"selector,omitempty"`
	Styles    json.RawMessage `json:"styles,omitempty"`
	TradeData json.RawMessage `json:"tradeData,omitempty"`
	Alert     json.RawMessage `json:"alert,omitempty"`
	Reason    string          `json:"reason"`
	Timestamp time.Time       `json:"timestamp"`
}

// UIObserver - Main observer system
type UIObserver struct {
	db              *sql.DB
	clients         map[string]*websocket.Conn
	clientsMu       sync.RWMutex
	observationChan chan ObservationMessage
	commandChan     chan SolaceCommand
	batchSize       int
	batchTimeout    time.Duration
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development (restrict in production)
	},
}

// NewUIObserver - Initialize the observation system
func NewUIObserver(db *sql.DB) *UIObserver {
	observer := &UIObserver{
		db:              db,
		clients:         make(map[string]*websocket.Conn),
		observationChan: make(chan ObservationMessage, 1000),
		commandChan:     make(chan SolaceCommand, 100),
		batchSize:       50,
		batchTimeout:    time.Second * 2,
	}

	// Start background processors
	go observer.processBatchObservations()
	go observer.processCommands()

	return observer
}

// HandleWebSocket - WebSocket endpoint handler
func (o *UIObserver) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	sessionID := uuid.New().String()
	o.clientsMu.Lock()
	o.clients[sessionID] = conn
	o.clientsMu.Unlock()

	log.Printf("🔌 SOLACE UI Observer connected - Session: %s", sessionID)

	// Log connection to database
	o.logWebSocketConnection(sessionID, "solace_control", true)

	// Send welcome message
	welcome := SolaceCommand{
		Type:      "system_status",
		CommandID: uuid.New().String(),
		Reason:    "SOLACE observation system active",
		Timestamp: time.Now(),
	}
	conn.WriteJSON(welcome)

	// Read loop
	for {
		var msg ObservationMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("⚠️ WebSocket read error: %v", err)
			break
		}

		msg.SessionID = sessionID
		if msg.Timestamp.IsZero() {
			msg.Timestamp = time.Now()
		}

		// Send to observation channel for batch processing
		o.observationChan <- msg
	}

	// Cleanup
	o.clientsMu.Lock()
	delete(o.clients, sessionID)
	o.clientsMu.Unlock()
	o.logWebSocketConnection(sessionID, "solace_control", false)
	log.Printf("🔌 SOLACE UI Observer disconnected - Session: %s", sessionID)
}

// processBatchObservations - Process observations in batches for performance
func (o *UIObserver) processBatchObservations() {
	batch := make([]ObservationMessage, 0, o.batchSize)
	timer := time.NewTimer(o.batchTimeout)

	for {
		select {
		case obs := <-o.observationChan:
			batch = append(batch, obs)
			if len(batch) >= o.batchSize {
				o.saveBatch(batch)
				batch = make([]ObservationMessage, 0, o.batchSize)
				timer.Reset(o.batchTimeout)
			}

		case <-timer.C:
			if len(batch) > 0 {
				o.saveBatch(batch)
				batch = make([]ObservationMessage, 0, o.batchSize)
			}
			timer.Reset(o.batchTimeout)
		}
	}
}

// saveBatch - Save batch of observations to database
func (o *UIObserver) saveBatch(batch []ObservationMessage) {
	tx, err := o.db.Begin()
	if err != nil {
		log.Printf("❌ Failed to start transaction: %v", err)
		return
	}
	defer tx.Rollback()

	for _, obs := range batch {
		switch obs.Type {
		case "ui_state":
			o.saveUIState(tx, obs)
		case "data_stream":
			o.saveDataStream(tx, obs)
		case "user_action":
			o.saveUserAction(tx, obs)
		case "market_context":
			o.saveMarketContext(tx, obs)
		default:
			log.Printf("⚠️ Unknown observation type: %s", obs.Type)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("❌ Failed to commit batch: %v", err)
		return
	}

	log.Printf("💾 Saved batch of %d observations", len(batch))
}

// saveUIState - Save UI state observation
func (o *UIObserver) saveUIState(tx *sql.Tx, obs ObservationMessage) {
	query := `
		INSERT INTO ui_state_log (session_id, component_type, element_id, state_snapshot, user_visible, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.Exec(query, obs.SessionID, obs.Component, obs.ElementID, obs.Data, obs.UserVisible, obs.Timestamp)
	if err != nil {
		log.Printf("❌ Failed to save UI state: %v", err)
	}
}

// saveDataStream - Save data stream observation
func (o *UIObserver) saveDataStream(tx *sql.Tx, obs ObservationMessage) {
	var data map[string]interface{}
	json.Unmarshal(obs.Data, &data)

	symbol, _ := data["symbol"].(string)
	streamType, _ := data["streamType"].(string)

	query := `
		INSERT INTO data_stream_log (session_id, stream_type, symbol, data_payload, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.Exec(query, obs.SessionID, streamType, symbol, obs.Data, obs.Timestamp)
	if err != nil {
		log.Printf("❌ Failed to save data stream: %v", err)
	}
}

// saveUserAction - Save user action observation
func (o *UIObserver) saveUserAction(tx *sql.Tx, obs ObservationMessage) {
	var data map[string]interface{}
	json.Unmarshal(obs.Data, &data)

	actionType, _ := data["actionType"].(string)
	targetElement, _ := data["targetElement"].(string)
	userIntent, _ := data["userIntent"].(string)

	query := `
		INSERT INTO user_actions (session_id, action_type, target_element, action_data, user_intent, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.Exec(query, obs.SessionID, actionType, targetElement, obs.Data, userIntent, obs.Timestamp)
	if err != nil {
		log.Printf("❌ Failed to save user action: %v", err)
	}
}

// saveMarketContext - Save market context snapshot
func (o *UIObserver) saveMarketContext(tx *sql.Tx, obs ObservationMessage) {
	var data map[string]interface{}
	json.Unmarshal(obs.Data, &data)

	symbol, _ := data["symbol"].(string)
	triggerEvent, _ := data["triggerEvent"].(string)
	currentPrice, _ := data["currentPrice"].(float64)

	query := `
		INSERT INTO market_context_snapshots 
		(session_id, trigger_event, symbol, current_price, top_5_bids, top_5_asks, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.Exec(query, obs.SessionID, triggerEvent, symbol, currentPrice,
		data["top5Bids"], data["top5Asks"], obs.Timestamp)
	if err != nil {
		log.Printf("❌ Failed to save market context: %v", err)
	}
}

// logWebSocketConnection - Track WebSocket connections
func (o *UIObserver) logWebSocketConnection(sessionID, connType string, connecting bool) {
	if connecting {
		query := `
			INSERT INTO websocket_connections (session_id, connection_type, status)
			VALUES ($1, $2, 'connected')
		`
		o.db.Exec(query, sessionID, connType)
	} else {
		query := `
			UPDATE websocket_connections 
			SET disconnected_at = NOW(), status = 'disconnected'
			WHERE session_id = $1 AND disconnected_at IS NULL
		`
		o.db.Exec(query, sessionID)
	}
}

// processCommands - Send commands from SOLACE to browser
func (o *UIObserver) processCommands() {
	for cmd := range o.commandChan {
		cmd.Timestamp = time.Now()
		cmd.CommandID = uuid.New().String()

		// Log command to database
		o.logSolaceDecision(cmd)

		// Broadcast to all connected clients
		o.clientsMu.RLock()
		for sessionID, conn := range o.clients {
			err := conn.WriteJSON(cmd)
			if err != nil {
				log.Printf("⚠️ Failed to send command to session %s: %v", sessionID, err)
			}
		}
		o.clientsMu.RUnlock()

		log.Printf("📤 SOLACE command sent: %s - %s", cmd.Type, cmd.Reason)
	}
}

// logSolaceDecision - Save SOLACE decision to database
func (o *UIObserver) logSolaceDecision(cmd SolaceCommand) {
	actionData, _ := json.Marshal(cmd)
	query := `
		INSERT INTO solace_ui_decisions 
		(session_id, decision_type, reasoning, confidence_score, action_taken, execution_status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
	`
	// Use empty session for broadcast commands
	o.db.Exec(query, "broadcast", cmd.Type, cmd.Reason, 0.75, actionData)
}

// SendCommand - Public method to send command to browser
func (o *UIObserver) SendCommand(cmd SolaceCommand) {
	o.commandChan <- cmd
}

// GetRecentObservations - Query recent observations
func (o *UIObserver) GetRecentObservations(c *gin.Context) {
	limit := c.DefaultQuery("limit", "100")
	obsType := c.Query("type")

	var query string
	if obsType != "" {
		query = fmt.Sprintf(`
			SELECT id, timestamp, session_id, component_type, element_id, state_snapshot
			FROM ui_state_log
			WHERE component_type = $1
			ORDER BY timestamp DESC
			LIMIT %s
		`, limit)
	} else {
		query = fmt.Sprintf(`
			SELECT id, timestamp, session_id, component_type, element_id, state_snapshot
			FROM ui_state_log
			ORDER BY timestamp DESC
			LIMIT %s
		`, limit)
	}

	var rows *sql.Rows
	var err error
	if obsType != "" {
		rows, err = o.db.Query(query, obsType)
	} else {
		rows, err = o.db.Query(query)
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	observations := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var timestamp time.Time
		var sessionID, componentType, elementID string
		var stateSnapshot json.RawMessage

		rows.Scan(&id, &timestamp, &sessionID, &componentType, &elementID, &stateSnapshot)
		observations = append(observations, map[string]interface{}{
			"id":            id,
			"timestamp":     timestamp,
			"sessionId":     sessionID,
			"componentType": componentType,
			"elementId":     elementID,
			"state":         stateSnapshot,
		})
	}

	c.JSON(200, gin.H{
		"count":        len(observations),
		"observations": observations,
	})
}

// GetActiveSessions - Get active observation sessions
func (o *UIObserver) GetActiveSessions(c *gin.Context) {
	query := `SELECT * FROM active_session_context`
	rows, err := o.db.Query(query)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	sessions := []map[string]interface{}{}
	for rows.Next() {
		var sessionID, connType string
		var connectedAt, lastUIUpdate, lastDataReceived, lastUserAction *time.Time
		var msgsReceived, uiEvents, dataPoints, userActions, solaceDecisions int

		rows.Scan(&sessionID, &connType, &connectedAt, &msgsReceived, &uiEvents,
			&dataPoints, &userActions, &solaceDecisions, &lastUIUpdate,
			&lastDataReceived, &lastUserAction)

		sessions = append(sessions, map[string]interface{}{
			"sessionId":        sessionID,
			"connectionType":   connType,
			"connectedAt":      connectedAt,
			"messagesReceived": msgsReceived,
			"uiEventsCount":    uiEvents,
			"dataPointsCount":  dataPoints,
			"userActionsCount": userActions,
			"solaceDecisions":  solaceDecisions,
			"lastUIUpdate":     lastUIUpdate,
			"lastDataReceived": lastDataReceived,
			"lastUserAction":   lastUserAction,
		})
	}

	c.JSON(200, gin.H{
		"activeCount": len(sessions),
		"sessions":    sessions,
	})
}
