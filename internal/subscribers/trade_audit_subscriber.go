package subscribers

import (
	"ares_api/internal/eventbus"
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
)

// TradeAuditLog represents an audit entry for trade events
type TradeAuditLog struct {
	ID              uint    `gorm:"primaryKey"`
	TradeID         int64   `gorm:"index"`
	EventType       string  `gorm:"type:varchar(50)"`
	TradingPair     string  `gorm:"type:varchar(20)"`
	Direction       string  `gorm:"type:varchar(4)"`
	Size            float64 `gorm:"type:decimal(20,8)"`
	Price           float64 `gorm:"type:decimal(20,8)"`
	Environment     string  `gorm:"type:varchar(20)"`
	Status          string  `gorm:"type:varchar(20)"`
	ExecutionTimeMS int64
	Timestamp       time.Time `gorm:"index"`
	RawEventData    string    `gorm:"type:jsonb"`
	CreatedAt       time.Time
}

// TradeAuditSubscriber handles trade event auditing
type TradeAuditSubscriber struct {
	db *gorm.DB
}

// NewTradeAuditSubscriber creates a new audit subscriber
func NewTradeAuditSubscriber(db *gorm.DB) *TradeAuditSubscriber {
	// Auto-migrate the audit log table
	if err := db.AutoMigrate(&TradeAuditLog{}); err != nil {
		log.Printf("[AUDIT][ERROR] Failed to migrate trade_audit_logs table: %v", err)
	} else {
		log.Println("[AUDIT][INFO] Trade audit log table ready")
	}

	return &TradeAuditSubscriber{db: db}
}

// HandleTradeExecuted processes trade_executed events
func (s *TradeAuditSubscriber) HandleTradeExecuted(data []byte) {
	var event eventbus.TradeExecutedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("[AUDIT][ERROR] Failed to unmarshal trade_executed event: %v", err)
		return
	}

	// Create audit log entry
	auditLog := TradeAuditLog{
		TradeID:         event.Data.TradeID,
		EventType:       event.Type,
		TradingPair:     event.Data.Symbol,
		Direction:       event.Data.Side,
		Size:            event.Data.Amount,
		Price:           event.Data.Price,
		Environment:     event.Data.ExchangeID,
		Status:          event.Data.Status,
		ExecutionTimeMS: event.Data.ExecutionTime,
		Timestamp:       event.Timestamp,
		RawEventData:    string(data),
	}

	// Save to database
	if err := s.db.Create(&auditLog).Error; err != nil {
		log.Printf("[AUDIT][ERROR] Failed to save audit log for trade %d: %v", event.Data.TradeID, err)
		return
	}

	log.Printf("[AUDIT][SUCCESS] Logged trade_executed event: Trade #%d %s %s @ $%.2f",
		event.Data.TradeID, event.Data.Side, event.Data.Symbol, event.Data.Price)
}

// Subscribe registers this subscriber with the EventBus
func (s *TradeAuditSubscriber) Subscribe(eb *eventbus.EventBus) {
	eb.Subscribe(eventbus.EventTypeTradeExecuted, s.HandleTradeExecuted)
	log.Println("[AUDIT][INFO] Subscribed to trade_executed events")
}
