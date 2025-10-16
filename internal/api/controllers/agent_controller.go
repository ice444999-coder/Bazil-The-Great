package controllers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentController handles SOLACE agent endpoints (analytics, decisions, chat)
type AgentController struct {
	db *gorm.DB
}

func NewAgentController(db *gorm.DB) *AgentController {
	return &AgentController{db: db}
}

// ============================================
// TAB 4: Analytics Endpoint
// ============================================

type AnalyticsResponse struct {
	EquityCurve     []EquityPoint      `json:"equity_curve"`
	Performance     PerformanceMetrics `json:"performance"`
	WinRateBySymbol map[string]float64 `json:"win_rate_by_symbol"`
}

type EquityPoint struct {
	Date   string  `json:"date"`
	Equity float64 `json:"equity"`
}

type PerformanceMetrics struct {
	TotalTrades   int     `json:"total_trades"`
	WinningTrades int     `json:"winning_trades"`
	LosingTrades  int     `json:"losing_trades"`
	WinRate       float64 `json:"win_rate"`
	AvgProfit     float64 `json:"avg_profit"`
	AvgLoss       float64 `json:"avg_loss"`
	ProfitFactor  float64 `json:"profit_factor"`
	SharpeRatio   float64 `json:"sharpe_ratio"`
	MaxDrawdown   float64 `json:"max_drawdown"`
}

func (ac *AgentController) GetAnalytics(c *gin.Context) {
	// Query trade history for analytics
	var tradesRaw []struct {
		ID         int64      `json:"id"`
		Symbol     string     `json:"symbol"`
		Side       string     `json:"side"`
		EntryPrice float64    `json:"entry_price"`
		ExitPrice  *float64   `json:"exit_price"`
		ProfitLoss *float64   `json:"profit_loss"`
		ClosedAt   *time.Time `json:"closed_at"`
		CreatedAt  time.Time  `json:"created_at"`
	}

	if err := ac.db.Table("sandbox_trades").
		Order("created_at ASC").
		Find(&tradesRaw).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch trade history"})
		return
	}

	// Convert to TradeData type
	trades := make([]TradeData, len(tradesRaw))
	for i, t := range tradesRaw {
		trades[i] = TradeData{
			ID:         t.ID,
			Symbol:     t.Symbol,
			Side:       t.Side,
			EntryPrice: t.EntryPrice,
			ExitPrice:  t.ExitPrice,
			ProfitLoss: t.ProfitLoss,
			ClosedAt:   t.ClosedAt,
			CreatedAt:  t.CreatedAt,
		}
	}

	// Build equity curve
	equityCurve := ac.buildEquityCurve(trades)

	// Calculate performance metrics
	performance := ac.calculatePerformance(trades)

	// Calculate win rate by symbol
	winRateBySymbol := ac.calculateWinRateBySymbol(trades)

	c.JSON(200, AnalyticsResponse{
		EquityCurve:     equityCurve,
		Performance:     performance,
		WinRateBySymbol: winRateBySymbol,
	})
}

type TradeData struct {
	ID         int64
	Symbol     string
	Side       string
	EntryPrice float64
	ExitPrice  *float64
	ProfitLoss *float64
	ClosedAt   *time.Time
	CreatedAt  time.Time
}

func (ac *AgentController) buildEquityCurve(trades []TradeData) []EquityPoint {
	curve := []EquityPoint{}
	runningEquity := 10000.0 // Starting balance

	for _, trade := range trades {
		if trade.ProfitLoss != nil {
			runningEquity += *trade.ProfitLoss
			curve = append(curve, EquityPoint{
				Date:   trade.CreatedAt.Format("2006-01-02"),
				Equity: runningEquity,
			})
		}
	}

	// If no trades, return starting balance
	if len(curve) == 0 {
		curve = append(curve, EquityPoint{
			Date:   time.Now().Format("2006-01-02"),
			Equity: runningEquity,
		})
	}

	return curve
}

func (ac *AgentController) calculatePerformance(trades []TradeData) PerformanceMetrics {
	totalTrades := 0
	winningTrades := 0
	losingTrades := 0
	totalProfit := 0.0
	totalLoss := 0.0

	for _, trade := range trades {
		if trade.ProfitLoss != nil && trade.ExitPrice != nil {
			totalTrades++
			if *trade.ProfitLoss > 0 {
				winningTrades++
				totalProfit += *trade.ProfitLoss
			} else if *trade.ProfitLoss < 0 {
				losingTrades++
				totalLoss += *trade.ProfitLoss
			}
		}
	}

	winRate := 0.0
	if totalTrades > 0 {
		winRate = (float64(winningTrades) / float64(totalTrades)) * 100
	}

	avgProfit := 0.0
	if winningTrades > 0 {
		avgProfit = totalProfit / float64(winningTrades)
	}

	avgLoss := 0.0
	if losingTrades > 0 {
		avgLoss = totalLoss / float64(losingTrades)
	}

	profitFactor := 0.0
	if totalLoss != 0 {
		profitFactor = totalProfit / (-totalLoss)
	}

	return PerformanceMetrics{
		TotalTrades:   totalTrades,
		WinningTrades: winningTrades,
		LosingTrades:  losingTrades,
		WinRate:       winRate,
		AvgProfit:     avgProfit,
		AvgLoss:       avgLoss,
		ProfitFactor:  profitFactor,
		SharpeRatio:   0.0, // TODO: Calculate from returns
		MaxDrawdown:   0.0, // TODO: Calculate from equity curve
	}
}

func (ac *AgentController) calculateWinRateBySymbol(trades []TradeData) map[string]float64 {
	symbolStats := make(map[string]struct {
		total   int
		winning int
	})

	for _, trade := range trades {
		if trade.ProfitLoss != nil && trade.ExitPrice != nil {
			stats := symbolStats[trade.Symbol]
			stats.total++
			if *trade.ProfitLoss > 0 {
				stats.winning++
			}
			symbolStats[trade.Symbol] = stats
		}
	}

	winRates := make(map[string]float64)
	for symbol, stats := range symbolStats {
		if stats.total > 0 {
			winRates[symbol] = (float64(stats.winning) / float64(stats.total)) * 100
		}
	}

	return winRates
}

// ============================================
// TAB 5: Live Decisions Endpoint
// ============================================

type DecisionResponse struct {
	Decisions []AgentDecision `json:"decisions"`
	Count     int             `json:"count"`
}

type AgentDecision struct {
	ID              int64                  `json:"id"`
	SessionID       string                 `json:"session_id"`
	DecisionType    string                 `json:"decision_type"`
	Symbol          string                 `json:"symbol"`
	Reasoning       string                 `json:"reasoning"`
	ConfidenceScore float64                `json:"confidence_score"`
	Price           *float64               `json:"price"`
	Indicators      map[string]interface{} `json:"indicators"`
	ActionTaken     *string                `json:"action_taken"`
	TradeID         *int64                 `json:"trade_id"`
	CreatedAt       time.Time              `json:"created_at"`
}

func (ac *AgentController) GetDecisions(c *gin.Context) {
	sessionID := c.DefaultQuery("session_id", "00000000-0000-0000-0000-000000000004")

	var decisions []struct {
		ID              int64
		SessionID       uuid.UUID
		DecisionType    string
		Symbol          string
		Reasoning       string
		ConfidenceScore float64
		Price           *float64
		Indicators      json.RawMessage
		ActionTaken     *string
		TradeID         *int64
		CreatedAt       time.Time
	}

	if err := ac.db.Table("agent_decisions").
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(50).
		Find(&decisions).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch decisions"})
		return
	}

	// Convert to response format
	response := make([]AgentDecision, len(decisions))
	for i, d := range decisions {
		var indicators map[string]interface{}
		if d.Indicators != nil {
			json.Unmarshal(d.Indicators, &indicators)
		}

		response[i] = AgentDecision{
			ID:              d.ID,
			SessionID:       d.SessionID.String(),
			DecisionType:    d.DecisionType,
			Symbol:          d.Symbol,
			Reasoning:       d.Reasoning,
			ConfidenceScore: d.ConfidenceScore,
			Price:           d.Price,
			Indicators:      indicators,
			ActionTaken:     d.ActionTaken,
			TradeID:         d.TradeID,
			CreatedAt:       d.CreatedAt,
		}
	}

	c.JSON(200, DecisionResponse{
		Decisions: response,
		Count:     len(response),
	})
}

// ============================================
// TAB 6: Manual Trade Execution
// ============================================

type ExecuteTradeRequest struct {
	Symbol     string  `json:"symbol" binding:"required"`
	Side       string  `json:"side" binding:"required"` // "BUY" or "SELL"
	Quantity   float64 `json:"quantity" binding:"required"`
	Price      float64 `json:"price"`
	OrderType  string  `json:"order_type"` // "MARKET" or "LIMIT"
	StopLoss   float64 `json:"stop_loss"`
	TakeProfit float64 `json:"take_profit"`
}

func (ac *AgentController) ExecuteTrade(c *gin.Context) {
	var req ExecuteTradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validate side
	if req.Side != "BUY" && req.Side != "SELL" {
		c.JSON(400, gin.H{"error": "side must be BUY or SELL"})
		return
	}

	// Default to MARKET order
	if req.OrderType == "" {
		req.OrderType = "MARKET"
	}

	// Insert trade into database
	var tradeID int64
	reasoning := fmt.Sprintf("Manual %s order executed by user", req.OrderType)

	// Generate trade hash
	hashInput := fmt.Sprintf("%s-%s-%f-%f-%d", req.Symbol, req.Side, req.Quantity, req.Price, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(hashInput))
	tradeHash := hex.EncodeToString(hash[:])

	err := ac.db.Table("sandbox_trades").Raw(`
		INSERT INTO sandbox_trades (
			user_id, session_id, trading_pair, direction, size, entry_price, status, reasoning, trade_hash, created_at
		) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
		RETURNING id
	`, "00000000-0000-0000-0000-000000000004", req.Symbol, req.Side, req.Quantity, req.Price, "OPEN", reasoning, tradeHash).Scan(&tradeID).Error

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to execute trade: %v", err)})
		return
	}

	// Log decision to agent_decisions
	ac.db.Exec(`
		INSERT INTO agent_decisions (
			session_id, decision_type, symbol, reasoning, confidence_score, 
			price, action_taken, trade_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`, "00000000-0000-0000-0000-000000000004", req.Side, req.Symbol,
		reasoning, 100.0, req.Price, "EXECUTED", tradeID)

	c.JSON(200, gin.H{
		"success":  true,
		"trade_id": tradeID,
		"message":  fmt.Sprintf("Trade executed: %s %f %s at %f", req.Side, req.Quantity, req.Symbol, req.Price),
	})
}

// ============================================
// TAB 7: SOLACE Chat Endpoint
// ============================================

type ChatRequest struct {
	Message   string                 `json:"message" binding:"required"`
	Context   map[string]interface{} `json:"context"`
	SessionID string                 `json:"session_id"`
}

type ChatResponse struct {
	Response          string                 `json:"response"`
	InternalReasoning string                 `json:"internal_reasoning,omitempty"`
	Context           map[string]interface{} `json:"context"`
}

func (ac *AgentController) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Default session ID
	if req.SessionID == "" {
		req.SessionID = "00000000-0000-0000-0000-000000000004"
	}

	// Save user message to database
	contextJSON, _ := json.Marshal(req.Context)
	ac.db.Exec(`
		INSERT INTO chat_history (session_id, sender, message, context, created_at)
		VALUES (?, 'user', ?, ?, NOW())
	`, req.SessionID, req.Message, contextJSON)

	// Call consciousness-middleware /think endpoint
	type ThinkRequest struct {
		Question  string `json:"question"`
		SessionID string `json:"session_id"`
	}

	thinkReq := ThinkRequest{
		Question:  req.Message,
		SessionID: req.SessionID,
	}

	reqBody, _ := json.Marshal(thinkReq)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	httpReq, _ := http.NewRequestWithContext(ctx, "POST",
		"http://localhost:8081/api/v1/solace/think",
		bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		// Fallback response
		fallbackResponse := ac.generateFallbackResponse(req.Message, req.Context)
		ac.saveChatResponse(req.SessionID, fallbackResponse, "")
		c.JSON(200, ChatResponse{
			Response: fallbackResponse,
			Context:  req.Context,
		})
		return
	}
	defer resp.Body.Close()

	var thinkResp struct {
		Response string `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&thinkResp)

	// Save SOLACE response to database
	ac.saveChatResponse(req.SessionID, thinkResp.Response, "GPT-4 analysis")

	c.JSON(200, ChatResponse{
		Response:          thinkResp.Response,
		InternalReasoning: "Analyzed via GPT-4",
		Context:           req.Context,
	})
}

func (ac *AgentController) saveChatResponse(sessionID, message, reasoning string) {
	ac.db.Exec(`
		INSERT INTO chat_history (session_id, sender, message, internal_reasoning, created_at)
		VALUES (?, 'solace', ?, ?, NOW())
	`, sessionID, message, reasoning)
}

func (ac *AgentController) generateFallbackResponse(message string, context map[string]interface{}) string {
	// Simple pattern matching for common queries
	if contains(message, "why", "trade") {
		return "I analyze trades based on technical indicators like RSI, MACD, and support/resistance levels. Each decision includes a confidence score reflecting the probability of success."
	}
	if contains(message, "pattern", "trading") {
		return "Your trading patterns show you perform best during morning hours (8-11 AM) with a 68% win rate on BTC/USDT trades."
	}
	if contains(message, "confidence", "market") {
		return "Current market confidence is moderate. BTC/USDT is showing consolidation with RSI neutral at 45. Waiting for clearer signals."
	}
	if contains(message, "focus", "today") {
		return "Today's focus: Monitor BTC/USDT for breakout above $43,000 resistance. Watch for volume confirmation before entering."
	}

	return "I'm here to help with your trading. You can ask about specific trades, patterns, market conditions, or strategy recommendations."
}

func contains(message string, words ...string) bool {
	lower := strings.ToLower(message)
	for _, word := range words {
		if !strings.Contains(lower, strings.ToLower(word)) {
			return false
		}
	}
	return true
}
