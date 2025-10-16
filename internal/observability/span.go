package observability

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ServiceSpan represents a distributed trace span
type ServiceSpan struct {
	ID            int64           `json:"id" gorm:"primaryKey"`
	TraceID       uuid.UUID       `json:"trace_id" gorm:"type:uuid;not null"`
	SpanID        string          `json:"span_id" gorm:"not null;unique"`
	ParentSpanID  string          `json:"parent_span_id,omitempty"`
	ServiceName   string          `json:"service_name" gorm:"not null"`
	OperationName string          `json:"operation_name" gorm:"not null"`
	StartTime     time.Time       `json:"start_time" gorm:"not null"`
	EndTime       *time.Time      `json:"end_time,omitempty"`
	DurationMs    *int            `json:"duration_ms,omitempty"`
	Status        string          `json:"status,omitempty"` // ok, error, timeout
	Tags          json.RawMessage `json:"tags,omitempty" gorm:"type:jsonb"`
	Logs          json.RawMessage `json:"logs,omitempty" gorm:"type:jsonb"`
}

func (ServiceSpan) TableName() string {
	return "service_spans"
}
