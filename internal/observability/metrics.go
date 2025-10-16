package observability

import (
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
)

// ServiceMetric represents a metric data point
type ServiceMetric struct {
	ID          int64           `json:"id" gorm:"primaryKey"`
	ServiceName string          `json:"service_name" gorm:"not null"`
	MetricName  string          `json:"metric_name" gorm:"not null"`
	MetricType  string          `json:"metric_type" gorm:"not null"` // counter, gauge, histogram
	MetricValue float64         `json:"metric_value" gorm:"not null"`
	Labels      json.RawMessage `json:"labels,omitempty" gorm:"type:jsonb"`
	Timestamp   time.Time       `json:"timestamp" gorm:"default:now()"`
}

func (ServiceMetric) TableName() string {
	return "service_metrics"
}

// MetricsCollector collects service metrics
type MetricsCollector struct {
	db          *gorm.DB
	serviceName string
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(db *gorm.DB, serviceName string) *MetricsCollector {
	return &MetricsCollector{
		db:          db,
		serviceName: serviceName,
	}
}

// RecordCounter increments a counter metric
func (m *MetricsCollector) RecordCounter(name string, value float64, labels map[string]string) {
	m.record("counter", name, value, labels)
}

// RecordGauge records a gauge metric (current value)
func (m *MetricsCollector) RecordGauge(name string, value float64, labels map[string]string) {
	m.record("gauge", name, value, labels)
}

// RecordHistogram records a histogram metric (duration/size)
func (m *MetricsCollector) RecordHistogram(name string, value float64, labels map[string]string) {
	m.record("histogram", name, value, labels)
}

// record writes a metric to the database
func (m *MetricsCollector) record(metricType, name string, value float64, labels map[string]string) {
	var labelsJSON json.RawMessage
	if labels != nil {
		data, err := json.Marshal(labels)
		if err != nil {
			log.Printf("[METRICS] Warning: Failed to marshal labels: %v", err)
		} else {
			labelsJSON = data
		}
	}

	metric := ServiceMetric{
		ServiceName: m.serviceName,
		MetricName:  name,
		MetricType:  metricType,
		MetricValue: value,
		Labels:      labelsJSON,
		Timestamp:   time.Now(),
	}

	// Async write to avoid blocking
	go func() {
		if err := m.db.Create(&metric).Error; err != nil {
			log.Printf("[METRICS] ⚠️  Failed to write metric: %v", err)
		}
	}()
}

// StartTimer returns a function that records duration when called
func (m *MetricsCollector) StartTimer(name string, labels map[string]string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Milliseconds()
		m.RecordHistogram(name, float64(duration), labels)
	}
}
