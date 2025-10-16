package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// Logger is the centralized logger for ARES
type Logger struct {
	db          *gorm.DB
	service     string
	enableDB    bool
	enableDebug bool
}

// NewLogger creates a new centralized logger
func NewLogger(service string, db *gorm.DB) *Logger {
	return &Logger{
		db:          db,
		service:     service,
		enableDB:    db != nil,
		enableDebug: os.Getenv("LOG_LEVEL") == "DEBUG",
	}
}

// Debug logs debug messages (only in debug mode)
func (l *Logger) Debug(message string, keyvals ...interface{}) {
	if !l.enableDebug {
		return
	}
	l.log(DEBUG, message, keyvals...)
}

// Info logs informational messages
func (l *Logger) Info(message string, keyvals ...interface{}) {
	l.log(INFO, message, keyvals...)
}

// Warn logs warning messages
func (l *Logger) Warn(message string, keyvals ...interface{}) {
	l.log(WARN, message, keyvals...)
}

// Error logs error messages
func (l *Logger) Error(message string, err error, keyvals ...interface{}) {
	if err != nil {
		keyvals = append(keyvals, "error", err.Error())
	}
	l.log(ERROR, message, keyvals...)
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, message string, keyvals ...interface{}) {
	// Format console output
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	consoleMsg := fmt.Sprintf("[%s][%s][%s] %s", timestamp, l.service, level, message)

	// Append key-value pairs
	if len(keyvals) > 0 {
		kvStr := formatKeyVals(keyvals...)
		consoleMsg = fmt.Sprintf("%s %s", consoleMsg, kvStr)
	}

	// Print to console
	log.Println(consoleMsg)

	// Optionally log to database
	if l.enableDB && level != DEBUG {
		go l.logToDB(level, message, keyvals...)
	}
}

// logToDB logs to the database asynchronously
func (l *Logger) logToDB(level LogLevel, message string, keyvals ...interface{}) {
	if l.db == nil {
		return
	}

	// Convert keyvals to JSON
	eventData := make(map[string]interface{})
	for i := 0; i < len(keyvals)-1; i += 2 {
		key := fmt.Sprintf("%v", keyvals[i])
		eventData[key] = keyvals[i+1]
	}

	eventJSON := ""
	if len(eventData) > 0 {
		bytes, _ := json.Marshal(eventData)
		eventJSON = string(bytes)
	}

	// Insert log entry
	logEntry := SystemLog{
		Service:   l.service,
		Level:     string(level),
		Message:   message,
		EventData: eventJSON,
		CreatedAt: time.Now(),
	}

	// Don't fail if DB insert fails
	if err := l.db.Create(&logEntry).Error; err != nil {
		log.Printf("[LOGGER][ERROR] Failed to write log to database: %v", err)
	}
}

// formatKeyVals formats key-value pairs for console output
func formatKeyVals(keyvals ...interface{}) string {
	if len(keyvals) == 0 {
		return ""
	}

	result := ""
	for i := 0; i < len(keyvals)-1; i += 2 {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%v=%v", keyvals[i], keyvals[i+1])
	}
	return result
}

// LogEvent logs a structured event with type
func (l *Logger) LogEvent(eventType string, data map[string]interface{}) {
	if l.db == nil {
		l.Info(fmt.Sprintf("Event: %s", eventType), mapToKeyVals(data)...)
		return
	}

	eventJSON := ""
	if data != nil {
		bytes, _ := json.Marshal(data)
		eventJSON = string(bytes)
	}

	logEntry := SystemLog{
		Service:   l.service,
		Level:     string(INFO),
		Message:   fmt.Sprintf("Event: %s", eventType),
		EventType: eventType,
		EventData: eventJSON,
		CreatedAt: time.Now(),
	}

	go func() {
		if err := l.db.Create(&logEntry).Error; err != nil {
			log.Printf("[LOGGER][ERROR] Failed to write event to database: %v", err)
		}
	}()

	// Also log to console
	l.Info(fmt.Sprintf("Event: %s", eventType), mapToKeyVals(data)...)
}

// mapToKeyVals converts a map to key-value pairs
func mapToKeyVals(data map[string]interface{}) []interface{} {
	result := make([]interface{}, 0, len(data)*2)
	for k, v := range data {
		result = append(result, k, v)
	}
	return result
}

// Global logger instance (initialized in main.go)
var GlobalLogger *Logger

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	GlobalLogger = logger
}

// Convenience functions using global logger
func Debug(message string, keyvals ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Debug(message, keyvals...)
	}
}

func Info(message string, keyvals ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Info(message, keyvals...)
	} else {
		log.Printf("[INFO] %s", message)
	}
}

func Warn(message string, keyvals ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Warn(message, keyvals...)
	} else {
		log.Printf("[WARN] %s", message)
	}
}

func Error(message string, err error, keyvals ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Error(message, err, keyvals...)
	} else {
		log.Printf("[ERROR] %s: %v", message, err)
	}
}

func LogEvent(eventType string, data map[string]interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.LogEvent(eventType, data)
	}
}
