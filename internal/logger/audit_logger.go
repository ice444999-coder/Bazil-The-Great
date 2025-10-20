package logger

import (
	"ares_api/internal/eventbus"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// AuditLogger subscribes to EventBus and logs all events to database
type AuditLogger struct {
	db       *gorm.DB
	eventBus *eventbus.EventBus
	debug    bool
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *gorm.DB, eb *eventbus.EventBus) *AuditLogger {
	return &AuditLogger{
		db:       db,
		eventBus: eb,
		debug:    true, // Set to false in production
	}
}

// Start subscribes to all event types and begins logging
func (al *AuditLogger) Start() {
	if al.eventBus == nil {
		log.Println("[AUDIT][WARN] EventBus not available, audit logging disabled")
		return
	}

	// Subscribe to trade events
	al.eventBus.Subscribe(eventbus.EventTypeTradeExecuted, al.handleTradeEvent)
	al.eventBus.Subscribe(eventbus.EventTypeTradeProposed, al.handleTradeEvent)
	al.eventBus.Subscribe(eventbus.EventTypeDecisionCompleted, al.handleDecisionEvent)

	log.Println("[AUDIT] ✅ Audit logger started, subscribed to events")
}

// handleTradeEvent logs trade events
func (al *AuditLogger) handleTradeEvent(data []byte) {
	var event eventbus.TradeExecutedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("[AUDIT][ERROR] Failed to unmarshal trade event: %v", err)
		return
	}

	// Log to console
	log.Printf("[AUDIT][TRADE] ID=%d Pair=%s Side=%s Size=%.2f Price=%.4f Status=%s",
		event.Data.TradeID,
		event.Data.Symbol,
		event.Data.Side,
		event.Data.Amount,
		event.Data.Price,
		event.Data.Status,
	)

	// Optionally log to database (system_logs table - to be created in Option B)
	// For now, just console logging
}

// handleDecisionEvent logs decision events
func (al *AuditLogger) handleDecisionEvent(data []byte) {
	var event eventbus.DecisionCompletedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("[AUDIT][ERROR] Failed to unmarshal decision event: %v", err)
		return
	}

	// Log to console
	log.Printf("[AUDIT][DECISION] Type=%s Confidence=%.2f Duration=%dms",
		event.Data.DecisionType,
		event.Data.Confidence,
		event.Data.Duration,
	)
}

// LogInfo logs informational messages with service context
func (al *AuditLogger) LogInfo(service, message string) {
	log.Printf("[%s][INFO] %s", service, message)
}

// LogError logs errors with service context
func (al *AuditLogger) LogError(service, message string, err error) {
	if err != nil {
		log.Printf("[%s][ERROR] %s: %v", service, message, err)
	} else {
		log.Printf("[%s][ERROR] %s", service, message)
	}
}

// LogWarn logs warnings with service context
func (al *AuditLogger) LogWarn(service, message string) {
	log.Printf("[%s][WARN] %s", service, message)
}

// LogDebug logs debug messages with service context (only in debug mode)
func (al *AuditLogger) LogDebug(service, message string) {
	if al.debug {
		log.Printf("[%s][DEBUG] %s", service, message)
	}
}

// SystemLog represents a log entry in the database
type SystemLog struct {
	ID        uint      `gorm:"primaryKey"`
	Service   string    `gorm:"size:50;index"`
	Level     string    `gorm:"size:20;index"` // INFO, WARN, ERROR, DEBUG
	Message   string    `gorm:"type:text"`
	EventType string    `gorm:"size:50"`
	EventData string    `gorm:"type:jsonb"`
	CreatedAt time.Time `gorm:"index"`
}

// TableName specifies the table name for SystemLog
func (SystemLog) TableName() string {
	return "system_logs"
}

// LogToDB logs an entry to the database
func (al *AuditLogger) LogToDB(service, level, message, eventType string, eventData map[string]interface{}) error {
	if al.db == nil {
		return fmt.Errorf("database not available")
	}

	eventJSON := ""
	if eventData != nil {
		bytes, _ := json.Marshal(eventData)
		eventJSON = string(bytes)
	}

	logEntry := SystemLog{
		Service:   service,
		Level:     level,
		Message:   message,
		EventType: eventType,
		EventData: eventJSON,
		CreatedAt: time.Now(),
	}

	return al.db.Create(&logEntry).Error
}
