package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceLog represents a structured log entry
type ServiceLog struct {
	ID           int64           `json:"id" gorm:"primaryKey"`
	TraceID      *uuid.UUID      `json:"trace_id,omitempty" gorm:"type:uuid"`
	SpanID       string          `json:"span_id,omitempty"`
	ParentSpanID string          `json:"parent_span_id,omitempty"`
	ServiceName  string          `json:"service_name" gorm:"not null"`
	LogLevel     string          `json:"log_level" gorm:"not null"`
	Message      string          `json:"message" gorm:"not null"`
	Metadata     json.RawMessage `json:"metadata,omitempty" gorm:"type:jsonb"`
	Timestamp    time.Time       `json:"timestamp" gorm:"default:now()"`
	SourceFile   string          `json:"source_file,omitempty"`
	SourceLine   int             `json:"source_line,omitempty"`
}

func (ServiceLog) TableName() string {
	return "service_logs"
}

// Logger provides structured logging with distributed tracing
type Logger struct {
	db          *gorm.DB
	serviceName string
	minLevel    LogLevel
}

// LogLevel represents logging severity
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	return []string{"DEBUG", "INFO", "WARN", "ERROR"}[l]
}

// NewLogger creates a new structured logger
func NewLogger(db *gorm.DB, serviceName string) *Logger {
	return &Logger{
		db:          db,
		serviceName: serviceName,
		minLevel:    INFO,
	}
}

// SetLevel sets the minimum logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.minLevel = level
}

// WithTrace creates a context with a new trace ID
func (l *Logger) WithTrace(ctx context.Context) (context.Context, uuid.UUID) {
	traceID := uuid.New()
	return context.WithValue(ctx, traceIDKey, traceID), traceID
}

// GetTraceID retrieves trace ID from context
func (l *Logger) GetTraceID(ctx context.Context) *uuid.UUID {
	if traceID, ok := ctx.Value(traceIDKey).(uuid.UUID); ok {
		return &traceID
	}
	return nil
}

type contextKey string

const traceIDKey contextKey = "trace_id"

// log writes a log entry to the database
func (l *Logger) log(ctx context.Context, level LogLevel, message string, metadata map[string]interface{}) {
	if level < l.minLevel {
		return
	}

	// Get caller info
	_, file, line, _ := runtime.Caller(2)

	// Get trace ID from context
	traceID := l.GetTraceID(ctx)

	// Serialize metadata
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			log.Printf("[LOGGER] Warning: Failed to marshal metadata: %v", err)
		} else {
			metadataJSON = data
		}
	}

	logEntry := ServiceLog{
		TraceID:     traceID,
		ServiceName: l.serviceName,
		LogLevel:    level.String(),
		Message:     message,
		Metadata:    metadataJSON,
		Timestamp:   time.Now(),
		SourceFile:  file,
		SourceLine:  line,
	}

	// Async write to avoid blocking
	go func() {
		if err := l.db.Create(&logEntry).Error; err != nil {
			log.Printf("[LOGGER] ⚠️  Failed to write log: %v", err)
		}
	}()

	// Also log to stdout for immediate visibility
	emoji := map[LogLevel]string{DEBUG: "🔍", INFO: "ℹ️", WARN: "⚠️", ERROR: "❌"}
	prefix := emoji[level]
	if traceID != nil {
		log.Printf("%s [%s] [%s] %s (trace: %s)", prefix, l.serviceName, level.String(), message, traceID.String()[:8])
	} else {
		log.Printf("%s [%s] [%s] %s", prefix, l.serviceName, level.String(), message)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, message string, metadata map[string]interface{}) {
	l.log(ctx, DEBUG, message, metadata)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, message string, metadata map[string]interface{}) {
	l.log(ctx, INFO, message, metadata)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, message string, metadata map[string]interface{}) {
	l.log(ctx, WARN, message, metadata)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, message string, metadata map[string]interface{}) {
	l.log(ctx, ERROR, message, metadata)
}

// Infof logs a formatted info message
func (l *Logger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.Info(ctx, fmt.Sprintf(format, args...), nil)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Error(ctx, fmt.Sprintf(format, args...), nil)
}

// QueryLogs retrieves logs from the database
func (l *Logger) QueryLogs(serviceName string, level string, traceID *uuid.UUID, limit int) ([]ServiceLog, error) {
	var logs []ServiceLog
	query := l.db.Model(&ServiceLog{})

	if serviceName != "" {
		query = query.Where("service_name = ?", serviceName)
	}
	if level != "" {
		query = query.Where("log_level = ?", level)
	}
	if traceID != nil {
		query = query.Where("trace_id = ?", traceID)
	}

	err := query.Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
