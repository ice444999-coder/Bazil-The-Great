package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// FaultVaultSession tracks development sessions across all three actors
type FaultVaultSession struct {
	SessionID      uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"session_id"`
	StartedAt      time.Time      `gorm:"not null;default:NOW()" json:"started_at"`
	EndedAt        *time.Time     `json:"ended_at,omitempty"`
	ContextType    string         `gorm:"not null;check:context_type IN ('vscode_claude', 'ares_claude', 'ares_autonomous')" json:"context_type"`
	SessionSummary string         `json:"session_summary,omitempty"`
	Active         bool           `gorm:"default:true" json:"active"`
	UserID         *int           `gorm:"index" json:"user_id,omitempty"`
	Metadata       map[string]any `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt      time.Time      `gorm:"default:NOW()" json:"created_at"`
}

// FaultVaultAction logs every action taken during development
type FaultVaultAction struct {
	ActionID       uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"action_id"`
	SessionID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"session_id"`
	Timestamp      time.Time      `gorm:"not null;default:NOW();index:idx_actions_timestamp,sort:desc" json:"timestamp"`
	Actor          string         `gorm:"not null;index" json:"actor"`
	ActionType     string         `gorm:"not null;check:action_type IN ('code_change', 'build', 'test', 'debug', 'crash', 'decision', 'feature_start', 'feature_complete', 'checkpoint', 'rollback');index" json:"action_type"`
	FilePath       string         `json:"file_path,omitempty"`
	FunctionName   string         `json:"function_name,omitempty"`
	Intent         string         `json:"intent,omitempty"`
	ChangesMade    string         `json:"changes_made,omitempty"`
	Result         string         `gorm:"check:result IN ('success', 'partial', 'failure', 'crash', 'pending');index" json:"result,omitempty"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	StackTrace     string         `json:"stack_trace,omitempty"`
	NextSteps      string         `json:"next_steps,omitempty"`
	RelatedActions pq.StringArray `gorm:"type:uuid[]" json:"related_actions,omitempty"`
	Metadata       map[string]any `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt      time.Time      `gorm:"default:NOW()" json:"created_at"`
}

// FaultVaultContext captures snapshots of system state
type FaultVaultContext struct {
	ContextID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"context_id"`
	SessionID            uuid.UUID      `gorm:"type:uuid;not null;index" json:"session_id"`
	Timestamp            time.Time      `gorm:"not null;default:NOW();index:idx_context_timestamp,sort:desc" json:"timestamp"`
	ConversationSnapshot map[string]any `gorm:"type:jsonb" json:"conversation_snapshot,omitempty"`
	UserIntent           string         `json:"user_intent,omitempty"`
	SystemState          map[string]any `gorm:"type:jsonb" json:"system_state,omitempty"`
	MemoryRefs           pq.StringArray `gorm:"type:uuid[]" json:"memory_refs,omitempty"`
	CreatedAt            time.Time      `gorm:"default:NOW()" json:"created_at"`
}

// FaultVaultLearning stores extracted patterns and learnings
type FaultVaultLearning struct {
	LearningID     uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"learning_id"`
	Pattern        string         `gorm:"not null;index" json:"pattern"`
	Outcome        string         `gorm:"not null;check:outcome IN ('success', 'failure')" json:"outcome"`
	Reason         string         `json:"reason,omitempty"`
	Confidence     float64        `gorm:"default:0.5;check:confidence >= 0 AND confidence <= 1;index:idx_learnings_confidence,sort:desc" json:"confidence"`
	TimesObserved  int            `gorm:"default:1" json:"times_observed"`
	LastSeen       time.Time      `gorm:"default:NOW();index:idx_learnings_last_seen,sort:desc" json:"last_seen"`
	Recommendation string         `json:"recommendation,omitempty"`
	Metadata       map[string]any `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt      time.Time      `gorm:"default:NOW()" json:"created_at"`
}

// TableName overrides
func (FaultVaultSession) TableName() string {
	return "fault_vault_sessions"
}

func (FaultVaultAction) TableName() string {
	return "fault_vault_actions"
}

func (FaultVaultContext) TableName() string {
	return "fault_vault_context"
}

func (FaultVaultLearning) TableName() string {
	return "fault_vault_learnings"
}
