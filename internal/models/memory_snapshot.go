package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONB custom type for storing JSON data in PostgreSQL
type JSONB map[string]interface{}

// Value converts JSONB to database value
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan converts database value to JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

type MemorySnapshot struct {
	ID               uint           `gorm:"primaryKey"`
	Timestamp        time.Time      `gorm:"autoCreateTime;not null;index"`
	EventType        string         `gorm:"type:varchar(100);not null;index"`
	Payload          JSONB          `gorm:"type:jsonb"`
	UserID           uint           `gorm:"index;not null"`
	SessionID        *uuid.UUID     `gorm:"type:uuid;index"`
	ImportanceScore  float64        `gorm:"default:0.5;index"`
	AccessCount      int            `gorm:"default:0;index"`
	LastAccessed     *time.Time     `gorm:"index"`
	MemoryType       string         `gorm:"type:varchar(50);default:'general';index"`
	Tags             []string       `gorm:"type:text[]"`
	CompressionLevel string         `gorm:"type:varchar(20);default:'none'"`
	Archived         bool           `gorm:"default:false;index"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// MemoryEmbedding stores vector embeddings for semantic search
// Note: Requires pgvector extension - if not available, will store as text
type MemoryEmbedding struct {
	ID         uint      `gorm:"primaryKey"`
	SnapshotID uint      `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	Embedding  string    `gorm:"type:text"` // Will be vector(384) if pgvector is installed
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

// EmbeddingQueueItem represents a pending embedding generation task
type EmbeddingQueueItem struct {
	ID           uint       `gorm:"primaryKey"`
	SnapshotID   uint       `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	Status       string     `gorm:"type:varchar(20);default:'pending';index"` // pending, processing, completed, failed
	RetryCount   int        `gorm:"default:0"`
	ErrorMessage string     `gorm:"type:text"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;index"`
	ProcessedAt  *time.Time
}

// MemoryRelationship tracks connections between memories
type MemoryRelationship struct {
	ID               uint      `gorm:"primaryKey"`
	SourceSnapshotID uint      `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	TargetSnapshotID uint      `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	RelationshipType string    `gorm:"type:varchar(50);index"` // follows, related_to, causes, references
	Strength         float64   `gorm:"default:1.0"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
}

// MemoryCacheStats tracks cache temperature for hot/warm/cold hierarchy
type MemoryCacheStats struct {
	SnapshotID  uint       `gorm:"primaryKey;constraint:OnDelete:CASCADE"`
	Temperature string     `gorm:"type:varchar(10);default:'cold';index"` // hot, warm, cold
	CacheHits   int        `gorm:"default:0"`
	CacheMisses int        `gorm:"default:0"`
	LastHit     *time.Time `gorm:"index"`
	PromotedAt  *time.Time
	DemotedAt   *time.Time
	SizeBytes   int
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
