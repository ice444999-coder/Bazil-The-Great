package eventbus

import "time"

// TradeProposedEvent is published when a trade is proposed but not yet executed
type TradeProposedEvent struct {
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		Symbol     string  `json:"symbol"`
		Side       string  `json:"side"`
		Amount     float64 `json:"amount"`
		Price      float64 `json:"price"`
		Confidence float64 `json:"confidence"`
		ProposedBy string  `json:"proposed_by"`
		Reason     string  `json:"reason"`
	} `json:"data"`
}

// TradeExecutedEvent is published when a trade has been executed
type TradeExecutedEvent struct {
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		TradeID       int64   `json:"trade_id"`
		Symbol        string  `json:"symbol"`
		Side          string  `json:"side"`
		Amount        float64 `json:"amount"`
		Price         float64 `json:"price"`
		ExecutedAt    string  `json:"executed_at"`
		ExchangeID    string  `json:"exchange_id"`
		Status        string  `json:"status"`
		ExecutionTime int64   `json:"execution_time_ms"`
	} `json:"data"`
}

// DecisionCompletedEvent is published when SOLACE completes a decision
type DecisionCompletedEvent struct {
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		DecisionID   int64                  `json:"decision_id"`
		DecisionType string                 `json:"decision_type"`
		Input        map[string]interface{} `json:"input"`
		Output       map[string]interface{} `json:"output"`
		Confidence   float64                `json:"confidence"`
		Duration     int64                  `json:"duration_ms"`
		Model        string                 `json:"model"`
	} `json:"data"`
}

// EventTypes constants
const (
	EventTypeTradeProposed     = "trade_proposed"
	EventTypeTradeExecuted     = "trade_executed"
	EventTypeDecisionCompleted = "decision_completed"
	EventVersion1              = "v1"
)

// NewTradeExecutedEvent creates a new TradeExecutedEvent
func NewTradeExecutedEvent(tradeID int64, symbol, side string, amount, price float64, executedAt, exchangeID, status string, executionTime int64) *TradeExecutedEvent {
	event := &TradeExecutedEvent{
		Type:      EventTypeTradeExecuted,
		Version:   EventVersion1,
		Timestamp: time.Now(),
	}
	event.Data.TradeID = tradeID
	event.Data.Symbol = symbol
	event.Data.Side = side
	event.Data.Amount = amount
	event.Data.Price = price
	event.Data.ExecutedAt = executedAt
	event.Data.ExchangeID = exchangeID
	event.Data.Status = status
	event.Data.ExecutionTime = executionTime
	return event
}

// NewTradeProposedEvent creates a new TradeProposedEvent
func NewTradeProposedEvent(symbol, side string, amount, price, confidence float64, proposedBy, reason string) *TradeProposedEvent {
	event := &TradeProposedEvent{
		Type:      EventTypeTradeProposed,
		Version:   EventVersion1,
		Timestamp: time.Now(),
	}
	event.Data.Symbol = symbol
	event.Data.Side = side
	event.Data.Amount = amount
	event.Data.Price = price
	event.Data.Confidence = confidence
	event.Data.ProposedBy = proposedBy
	event.Data.Reason = reason
	return event
}

// NewDecisionCompletedEvent creates a new DecisionCompletedEvent
func NewDecisionCompletedEvent(decisionID int64, decisionType string, input, output map[string]interface{}, confidence float64, duration int64, model string) *DecisionCompletedEvent {
	event := &DecisionCompletedEvent{
		Type:      EventTypeDecisionCompleted,
		Version:   EventVersion1,
		Timestamp: time.Now(),
	}
	event.Data.DecisionID = decisionID
	event.Data.DecisionType = decisionType
	event.Data.Input = input
	event.Data.Output = output
	event.Data.Confidence = confidence
	event.Data.Duration = duration
	event.Data.Model = model
	return event
}
