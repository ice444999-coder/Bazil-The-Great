package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ============================================
// CONSCIOUSNESS MIDDLEWARE CLIENT
// ============================================
// Fast event-sourced queries via consciousness-middleware
// Replaces slow direct PostgreSQL queries
// ============================================

type ConsciousnessClient struct {
	BaseURL string
	Client  *http.Client
}

func NewConsciousnessClient() *ConsciousnessClient {
	return &ConsciousnessClient{
		BaseURL: "http://localhost:8081",
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ============================================
// TRADING QUERIES
// ============================================

type Trade struct {
	ID          int64      `json:"id"`
	TradingPair string     `json:"trading_pair"`
	Direction   string     `json:"direction"`
	EntryPrice  float64    `json:"entry_price"`
	ExitPrice   *float64   `json:"exit_price"`
	ProfitLoss  *float64   `json:"profit_loss"`
	Status      string     `json:"status"`
	OpenedAt    time.Time  `json:"opened_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	Reasoning   string     `json:"reasoning"`
}

func (cc *ConsciousnessClient) GetTradeHistory(limit int) ([]Trade, error) {
	req := map[string]interface{}{
		"query_type": "get_trades",
		"limit":      limit,
		"session_id": uuid.New().String(),
	}

	respData, err := cc.post("/api/v1/solace/query", req)
	if err != nil {
		return nil, err
	}

	var trades []Trade
	if tradesData, ok := respData["trades"].([]interface{}); ok {
		jsonBytes, _ := json.Marshal(tradesData)
		json.Unmarshal(jsonBytes, &trades)
	}

	return trades, nil
}

// ============================================
// PLAYBOOK QUERIES
// ============================================

type PlaybookRule struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Conditions  string    `json:"conditions"`
	Actions     string    `json:"actions"`
	Priority    int       `json:"priority"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
}

func (cc *ConsciousnessClient) GetPlaybookRules() ([]PlaybookRule, error) {
	req := map[string]interface{}{
		"query_type": "get_playbook_rules",
		"session_id": uuid.New().String(),
	}

	respData, err := cc.post("/api/v1/solace/query", req)
	if err != nil {
		return nil, err
	}

	var rules []PlaybookRule
	if rulesData, ok := respData["rules"].([]interface{}); ok {
		jsonBytes, _ := json.Marshal(rulesData)
		json.Unmarshal(jsonBytes, &rules)
	}

	return rules, nil
}

// ============================================
// OBSERVATION LOGGING (Event Sourcing)
// ============================================

func (cc *ConsciousnessClient) LogObservation(obsType string, symbol string, data map[string]interface{}, sessionID string) error {
	req := map[string]interface{}{
		"event_type":       "observation",
		"observation_type": obsType,
		"symbol":           symbol,
		"data":             data,
		"session_id":       sessionID,
		"timestamp":        time.Now().Unix(),
	}

	_, err := cc.post("/api/v1/solace/event", req)
	return err
}

// ============================================
// CONVERSATION LOGGING
// ============================================

func (cc *ConsciousnessClient) LogConversation(speaker, messageType, content, sessionID string) error {
	req := map[string]interface{}{
		"event_type":   "conversation",
		"speaker":      speaker,
		"message_type": messageType,
		"content":      content,
		"session_id":   sessionID,
		"timestamp":    time.Now().Unix(),
	}

	_, err := cc.post("/api/v1/solace/event", req)
	return err
}

// ============================================
// STATISTICS (Cached Aggregations)
// ============================================

type SOLACEStats struct {
	TotalObservations int64   `json:"total_observations"`
	TodayTrades       int64   `json:"today_trades"`
	OpenTrades        int64   `json:"open_trades"`
	DailyPnL          float64 `json:"daily_pnl"`
}

func (cc *ConsciousnessClient) GetStats() (*SOLACEStats, error) {
	req := map[string]interface{}{
		"query_type": "get_stats",
		"session_id": uuid.New().String(),
	}

	respData, err := cc.post("/api/v1/solace/query", req)
	if err != nil {
		return nil, err
	}

	stats := &SOLACEStats{
		TotalObservations: int64(respData["total_observations"].(float64)),
		TodayTrades:       int64(respData["today_trades"].(float64)),
		OpenTrades:        int64(respData["open_trades"].(float64)),
		DailyPnL:          respData["daily_pnl"].(float64),
	}

	return stats, nil
}

// ============================================
// MEMORY QUERIES
// ============================================

type ObservationMemory struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

func (cc *ConsciousnessClient) GetMemory(sessionID string, limit int) ([]ObservationMemory, error) {
	req := map[string]interface{}{
		"query_type": "get_memory",
		"session_id": sessionID,
		"limit":      limit,
	}

	respData, err := cc.post("/api/v1/solace/query", req)
	if err != nil {
		return nil, err
	}

	var memories []ObservationMemory
	if memData, ok := respData["memories"].([]interface{}); ok {
		jsonBytes, _ := json.Marshal(memData)
		json.Unmarshal(jsonBytes, &memories)
	}

	return memories, nil
}

// ============================================
// HTTP CLIENT
// ============================================

func (cc *ConsciousnessClient) post(endpoint string, data map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	url := cc.BaseURL + endpoint
	resp, err := cc.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("consciousness-middleware unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("consciousness-middleware error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// ============================================
// HEALTH CHECK
// ============================================

func (cc *ConsciousnessClient) HealthCheck() error {
	resp, err := cc.Client.Get(cc.BaseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
