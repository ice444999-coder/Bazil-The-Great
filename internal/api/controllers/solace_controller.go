package controllers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type SolaceController struct {
	DB                  *gorm.DB
	ConsciousnessClient *ConsciousnessClient
}

func NewSolaceController(db *gorm.DB) *SolaceController {
	return &SolaceController{
		DB:                  db,
		ConsciousnessClient: NewConsciousnessClient(),
	}
}

// Observation represents a SOLACE observation event
type Observation struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
	Symbol    string                 `json:"symbol"`
	SessionID string                 `json:"sessionId"`
}

// GetTradeHistory returns all sandbox trades (FAST - uses consciousness-middleware)
func (sc *SolaceController) GetTradeHistory(c *gin.Context) {
	trades, err := sc.ConsciousnessClient.GetTradeHistory(100)
	if err != nil {
		// Fallback to direct database if consciousness-middleware down
		var tradesFallback []map[string]interface{}
		result := sc.DB.Raw(`
			SELECT 
				id, trading_pair, direction, entry_price, exit_price,
				profit_loss, status, opened_at, closed_at, reasoning
			FROM sandbox_trades 
			ORDER BY opened_at DESC 
			LIMIT 100
		`).Scan(&tradesFallback)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"trades": tradesFallback,
			"count":  len(tradesFallback),
			"source": "database_fallback",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trades": trades,
		"count":  len(trades),
		"source": "consciousness-middleware",
	})
}

// GetPlaybookRules returns all playbook rules (FAST - uses consciousness-middleware)
func (sc *SolaceController) GetPlaybookRules(c *gin.Context) {
	rules, err := sc.ConsciousnessClient.GetPlaybookRules()
	if err != nil {
		// Fallback to direct database
		var rulesFallback []map[string]interface{}
		result := sc.DB.Raw(`SELECT * FROM playbook_rules ORDER BY created_at DESC`).Scan(&rulesFallback)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"rules":  rulesFallback,
			"count":  len(rulesFallback),
			"source": "database_fallback",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules":  rules,
		"count":  len(rules),
		"source": "consciousness-middleware",
	})
}

// HandleObservationBatch handles batch observation submission (ASYNC - fire-and-forget)
func (sc *SolaceController) HandleObservationBatch(c *gin.Context) {
	var observation Observation

	if err := c.BindJSON(&observation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log to consciousness-middleware asynchronously (FAST)
	go sc.ConsciousnessClient.LogObservation(
		observation.Type,
		observation.Symbol,
		observation.Data,
		observation.SessionID,
	)

	// Return immediately - don't wait for database
	c.JSON(http.StatusAccepted, gin.H{
		"status":  "accepted",
		"type":    observation.Type,
		"session": observation.SessionID,
		"source":  "consciousness-middleware",
	})
}

// saveObservation saves observation to appropriate table
func (sc *SolaceController) saveObservation(obs Observation) error {
	dataJSON, _ := json.Marshal(obs.Data)

	// Choose table based on observation type
	switch obs.Type {
	case "orderbook_update", "chart_loaded", "trade_executed", "historical_data_loaded", "price_update", "websockets_connected", "chart_initialized":
		// Data stream events
		return sc.DB.Exec(`
			INSERT INTO data_stream_log 
			(stream_type, symbol, data_payload, session_id, timestamp, created_at)
			VALUES (?, ?, ?, ?, NOW(), NOW())
		`, obs.Type, obs.Symbol, dataJSON, obs.SessionID).Error

	case "trade_submitted", "amount_changed", "trade_type_changed", "tab_switched", "timeframe_changed", "solace_toggled", "session_started", "logout", "trade_history_loaded":
		// User action events
		return sc.DB.Exec(`
			INSERT INTO user_actions 
			(action_type, input_value, context, session_id, timestamp, created_at)
			VALUES (?, ?, ?, ?, NOW(), NOW())
		`, obs.Type, string(dataJSON), dataJSON, obs.SessionID).Error

	default:
		// General UI state events
		return sc.DB.Exec(`
			INSERT INTO ui_state_log 
			(component_type, state_snapshot, session_id, timestamp, created_at)
			VALUES (?, ?, ?, NOW(), NOW())
		`, obs.Type, dataJSON, obs.SessionID).Error
	}
}

// HandleSOLACEWebSocket handles WebSocket connections for real-time SOLACE communication
func (sc *SolaceController) HandleSOLACEWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
		return
	}
	defer conn.Close()

	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	// Send welcome message
	welcome := map[string]interface{}{
		"type":    "connected",
		"message": "SOLACE consciousness connected",
		"data": map[string]interface{}{
			"sessionId": sessionID,
			"timestamp": time.Now().Unix(),
		},
	}
	conn.WriteJSON(welcome)

	// Read observations
	for {
		var obs Observation
		err := conn.ReadJSON(&obs)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected closure
			}
			break
		}

		// Save observation
		go sc.saveObservation(obs)

		// Process observation and send commands back
		go sc.processObservation(obs, conn)
	}
}

// processObservation analyzes observation and sends commands to frontend
func (sc *SolaceController) processObservation(obs Observation, conn *websocket.Conn) {
	// Example: Detect large trade attempts
	if obs.Type == "trade_submitted" {
		amount, ok := obs.Data["amount"].(string)
		if ok && amount != "" {
			// Could add AI logic here
			// For now, just acknowledge
			command := map[string]interface{}{
				"type":    "show_alert",
				"message": "SOLACE detected " + obs.Type,
			}
			conn.WriteJSON(command)
		}
	}

	// Example: Pattern detection
	if obs.Type == "tab_switched" {
		var count int64
		sc.DB.Raw(`
			SELECT COUNT(*) 
			FROM ui_state_log 
			WHERE component_type = ? 
			  AND session_id = ? 
			  AND timestamp > NOW() - INTERVAL '5 minutes'
		`, "tab_switched", obs.SessionID).Scan(&count)

		if count > 10 {
			command := map[string]interface{}{
				"type":    "show_alert",
				"message": fmt.Sprintf("You've switched tabs %d times in 5 min", count),
			}
			conn.WriteJSON(command)
		}
	}
}

// GetSOLACEStats returns aggregated statistics (FAST - uses cached data)
func (sc *SolaceController) GetSOLACEStats(c *gin.Context) {
	stats, err := sc.ConsciousnessClient.GetStats()
	if err != nil {
		// Fallback to direct database
		statsFallback := make(map[string]interface{})

		var totalObs, todayTrades, openTrades int64
		var dailyPnL float64

		sc.DB.Table("ui_state_log").Count(&totalObs)
		sc.DB.Table("sandbox_trades").Where("DATE(opened_at) = CURRENT_DATE").Count(&todayTrades)
		sc.DB.Table("sandbox_trades").Where("status = ?", "open").Count(&openTrades)
		sc.DB.Raw("SELECT COALESCE(SUM(profit_loss), 0) FROM sandbox_trades WHERE DATE(opened_at) = CURRENT_DATE").Scan(&dailyPnL)

		statsFallback["total_observations"] = totalObs
		statsFallback["today_trades"] = todayTrades
		statsFallback["open_trades"] = openTrades
		statsFallback["daily_pnl"] = dailyPnL
		statsFallback["source"] = "database_fallback"

		c.JSON(http.StatusOK, statsFallback)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_observations": stats.TotalObservations,
		"today_trades":       stats.TodayTrades,
		"open_trades":        stats.OpenTrades,
		"daily_pnl":          stats.DailyPnL,
		"source":             "consciousness-middleware",
	})
}

// GetSOLACEMemory returns observation history for a session
func (sc *SolaceController) GetSOLACEMemory(c *gin.Context) {
	// SOLACE's REAL memory substrate - PostgreSQL conversation history
	var totalConversations int64
	var totalDecisions int64
	var errorPatterns int64
	var hederaProofs int64

	sc.DB.Table("conversation_events").Count(&totalConversations)
	sc.DB.Table("decision_events").Count(&totalDecisions)
	sc.DB.Table("error_patterns").Count(&errorPatterns)
	sc.DB.Table("session_proofs").Count(&hederaProofs)

	// Get recent conversations
	type ConvRecord struct {
		Speaker   string    `json:"speaker"`
		Content   string    `json:"content"`
		Timestamp time.Time `json:"timestamp"`
	}
	var recentConvs []ConvRecord
	sc.DB.Table("conversation_events").
		Select("speaker, content, created_at as timestamp").
		Order("created_at desc").
		Limit(10).
		Scan(&recentConvs)

	// Get recent decisions
	type DecRecord struct {
		DecisionType string    `json:"decision_type"`
		Reasoning    string    `json:"reasoning"`
		Timestamp    time.Time `json:"timestamp"`
	}
	var recentDecs []DecRecord
	sc.DB.Table("decision_events").
		Select("decision_type, reasoning, created_at as timestamp").
		Order("created_at desc").
		Limit(5).
		Scan(&recentDecs)

	c.JSON(http.StatusOK, gin.H{
		"total_conversations":  totalConversations,
		"total_decisions":      totalDecisions,
		"error_patterns":       errorPatterns,
		"hedera_proofs":        hederaProofs,
		"recent_conversations": recentConvs,
		"recent_decisions":     recentDecs,
	})
}

// Helper function to generate session ID
func generateSessionID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// LogConversation logs conversation to consciousness-middleware (ASYNC)
func (sc *SolaceController) LogConversation(c *gin.Context) {
	var req struct {
		Speaker     string `json:"speaker"`
		MessageType string `json:"message_type"`
		Content     string `json:"content"`
		SessionID   string `json:"session_id"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log asynchronously - don't wait for database
	go sc.ConsciousnessClient.LogConversation(
		req.Speaker,
		req.MessageType,
		req.Content,
		req.SessionID,
	)

	// Return immediately
	c.JSON(http.StatusAccepted, gin.H{
		"logged": true,
		"source": "consciousness-middleware",
	})
}
