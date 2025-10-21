/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package monitoring

import (
	"sync"
	"time"
)

// Metrics tracks system health and performance
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests     int64
	FailedRequests    int64
	AvgResponseTimeMs float64

	// Trading metrics
	ActiveTrades      int
	TotalTradesOpened int64
	TotalTradesClosed int64

	// LLM metrics
	LLMRequests         int64
	LLMFailures         int64
	LLMAvgLatencyMs     float64
	CircuitBreakerState string

	// Database metrics
	DBConnections int
	DBQueryCount  int64
	DBSlowQueries int64

	// System metrics
	StartTime       time.Time
	LastHealthCheck time.Time
	MemoryUsageMB   float64
	GoroutineCount  int

	// Extended system metrics (gopsutil)
	CPUPercent      float64
	RAMTotalGB      float64
	RAMUsedGB       float64
	RAMUsedPercent  float64
	DiskTotalGB     float64
	DiskUsedGB      float64
	DiskUsedPercent float64
	CPUTemperatureC float64

	// Error tracking
	Errors    []ErrorEntry
	MaxErrors int
}

// ErrorEntry represents a logged error
type ErrorEntry struct {
	Timestamp time.Time
	Component string
	Error     string
	UserID    uint
	TraceID   string
}

// PerformanceEntry tracks individual request performance
type PerformanceEntry struct {
	Timestamp  time.Time
	Endpoint   string
	DurationMs float64
	StatusCode int
	UserID     uint
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime:           time.Now(),
		MaxErrors:           1000, // Keep last 1000 errors
		Errors:              make([]ErrorEntry, 0, 1000),
		CircuitBreakerState: "closed",
	}
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(durationMs float64, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	if !success {
		m.FailedRequests++
	}

	// Update average response time (moving average)
	if m.TotalRequests == 1 {
		m.AvgResponseTimeMs = durationMs
	} else {
		m.AvgResponseTimeMs = (m.AvgResponseTimeMs*float64(m.TotalRequests-1) + durationMs) / float64(m.TotalRequests)
	}
}

// RecordTrade records trade activity
func (m *Metrics) RecordTrade(opened bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if opened {
		m.TotalTradesOpened++
		m.ActiveTrades++
	} else {
		m.TotalTradesClosed++
		m.ActiveTrades--
		if m.ActiveTrades < 0 {
			m.ActiveTrades = 0 // Safety check
		}
	}
}

// RecordLLMRequest records LLM activity
func (m *Metrics) RecordLLMRequest(latencyMs float64, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.LLMRequests++
	if !success {
		m.LLMFailures++
	}

	// Update average latency
	if m.LLMRequests == 1 {
		m.LLMAvgLatencyMs = latencyMs
	} else {
		m.LLMAvgLatencyMs = (m.LLMAvgLatencyMs*float64(m.LLMRequests-1) + latencyMs) / float64(m.LLMRequests)
	}
}

// UpdateCircuitBreaker updates circuit breaker state
func (m *Metrics) UpdateCircuitBreaker(state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CircuitBreakerState = state
}

// RecordError logs an error
func (m *Metrics) RecordError(component, errorMsg string, userID uint, traceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := ErrorEntry{
		Timestamp: time.Now(),
		Component: component,
		Error:     errorMsg,
		UserID:    userID,
		TraceID:   traceID,
	}

	m.Errors = append(m.Errors, entry)

	// Trim if too many errors
	if len(m.Errors) > m.MaxErrors {
		m.Errors = m.Errors[len(m.Errors)-m.MaxErrors:]
	}
}

// UpdateSystemMetrics updates system-level metrics
func (m *Metrics) UpdateSystemMetrics(memoryMB float64, goroutines int, dbConns int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.MemoryUsageMB = memoryMB
	m.GoroutineCount = goroutines
	m.DBConnections = dbConns
	m.LastHealthCheck = time.Now()
}

// UpdateExtendedSystemMetrics updates extended system metrics (CPU, RAM, Disk, Temp)
func (m *Metrics) UpdateExtendedSystemMetrics(cpuPercent, ramTotalGB, ramUsedGB, ramUsedPercent, diskTotalGB, diskUsedGB, diskUsedPercent, cpuTemp float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CPUPercent = cpuPercent
	m.RAMTotalGB = ramTotalGB
	m.RAMUsedGB = ramUsedGB
	m.RAMUsedPercent = ramUsedPercent
	m.DiskTotalGB = diskTotalGB
	m.DiskUsedGB = diskUsedGB
	m.DiskUsedPercent = diskUsedPercent
	m.CPUTemperatureC = cpuTemp
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MetricsSnapshot{
		StartTime:           m.StartTime, // Added for uptime calculation
		Uptime:              time.Since(m.StartTime).String(),
		TotalRequests:       m.TotalRequests,
		FailedRequests:      m.FailedRequests,
		SuccessRate:         m.calculateSuccessRate(),
		AvgResponseTimeMs:   m.AvgResponseTimeMs,
		ActiveTrades:        m.ActiveTrades,
		TotalTradesOpened:   m.TotalTradesOpened,
		TotalTradesClosed:   m.TotalTradesClosed,
		LLMRequests:         m.LLMRequests,
		LLMFailures:         m.LLMFailures,
		LLMSuccessRate:      m.calculateLLMSuccessRate(),
		LLMAvgLatencyMs:     m.LLMAvgLatencyMs,
		CircuitBreakerState: m.CircuitBreakerState,
		DBConnections:       m.DBConnections,
		DBQueryCount:        m.DBQueryCount,
		MemoryUsageMB:       m.MemoryUsageMB,
		GoroutineCount:      m.GoroutineCount,
		LastHealthCheck:     m.LastHealthCheck,
		CPUPercent:          m.CPUPercent,
		RAMTotalGB:          m.RAMTotalGB,
		RAMUsedGB:           m.RAMUsedGB,
		RAMUsedPercent:      m.RAMUsedPercent,
		DiskTotalGB:         m.DiskTotalGB,
		DiskUsedGB:          m.DiskUsedGB,
		DiskUsedPercent:     m.DiskUsedPercent,
		CPUTemperatureC:     m.CPUTemperatureC,
		RecentErrors:        m.getRecentErrors(10),
	}
}

// MetricsSnapshot represents metrics at a point in time
type MetricsSnapshot struct {
	StartTime           time.Time // Added for uptime calculation
	Uptime              string
	TotalRequests       int64
	FailedRequests      int64
	SuccessRate         float64
	AvgResponseTimeMs   float64
	ActiveTrades        int
	TotalTradesOpened   int64
	TotalTradesClosed   int64
	LLMRequests         int64
	LLMFailures         int64
	LLMSuccessRate      float64
	LLMAvgLatencyMs     float64
	CircuitBreakerState string
	DBConnections       int
	DBQueryCount        int64
	MemoryUsageMB       float64
	GoroutineCount      int
	LastHealthCheck     time.Time

	// Extended system metrics
	CPUPercent      float64
	RAMTotalGB      float64
	RAMUsedGB       float64
	RAMUsedPercent  float64
	DiskTotalGB     float64
	DiskUsedGB      float64
	DiskUsedPercent float64
	CPUTemperatureC float64

	RecentErrors []ErrorEntry
}

func (m *Metrics) calculateSuccessRate() float64 {
	if m.TotalRequests == 0 {
		return 100.0
	}
	return float64(m.TotalRequests-m.FailedRequests) / float64(m.TotalRequests) * 100.0
}

func (m *Metrics) calculateLLMSuccessRate() float64 {
	if m.LLMRequests == 0 {
		return 100.0
	}
	return float64(m.LLMRequests-m.LLMFailures) / float64(m.LLMRequests) * 100.0
}

func (m *Metrics) getRecentErrors(count int) []ErrorEntry {
	if len(m.Errors) == 0 {
		return []ErrorEntry{}
	}

	start := len(m.Errors) - count
	if start < 0 {
		start = 0
	}

	return m.Errors[start:]
}

// HealthStatus represents system health
type HealthStatus struct {
	Status    string                 `json:"status"` // healthy, degraded, unhealthy
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Status  string `json:"status"` // pass, warn, fail
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// CheckHealth performs health checks
func (m *Metrics) CheckHealth() HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	checks := make(map[string]HealthCheck)
	overallHealthy := true

	// Check LLM circuit breaker
	if m.CircuitBreakerState == "open" {
		checks["llm"] = HealthCheck{
			Status:  "fail",
			Message: "Circuit breaker is OPEN - LLM unavailable",
		}
		overallHealthy = false
	} else if m.CircuitBreakerState == "half-open" {
		checks["llm"] = HealthCheck{
			Status:  "warn",
			Message: "Circuit breaker is HALF-OPEN - testing recovery",
		}
	} else {
		checks["llm"] = HealthCheck{
			Status:  "pass",
			Message: "LLM operational",
		}
	}

	// Check request success rate
	successRate := m.calculateSuccessRate()
	if successRate < 90 {
		checks["requests"] = HealthCheck{
			Status:  "fail",
			Message: "High error rate",
		}
		overallHealthy = false
	} else if successRate < 95 {
		checks["requests"] = HealthCheck{
			Status:  "warn",
			Message: "Elevated error rate",
		}
	} else {
		checks["requests"] = HealthCheck{
			Status:  "pass",
			Message: "Requests healthy",
		}
	}

	// Check memory usage
	if m.MemoryUsageMB > 1000 {
		checks["memory"] = HealthCheck{
			Status:  "warn",
			Message: "High memory usage",
		}
	} else {
		checks["memory"] = HealthCheck{
			Status:  "pass",
			Message: "Memory usage normal",
		}
	}

	status := "healthy"
	if !overallHealthy {
		status = "unhealthy"
	} else {
		// Check for warnings
		for _, check := range checks {
			if check.Status == "warn" {
				status = "degraded"
				break
			}
		}
	}

	return HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Checks:    checks,
	}
}
