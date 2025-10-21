/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package models

import (
	"encoding/json"
	"time"
)

// Agent represents an AI agent in the swarm
type Agent struct {
	AgentID             string          `json:"agent_id" db:"agent_id"`
	AgentName           string          `json:"agent_name" db:"agent_name"`
	AgentType           string          `json:"agent_type" db:"agent_type"` // openai, claude, deepseek
	Capabilities        json.RawMessage `json:"capabilities" db:"capabilities"`
	Status              string          `json:"status" db:"status"` // idle, busy, offline
	CurrentTaskID       *string         `json:"current_task_id,omitempty" db:"current_task_id"`
	TotalTasksCompleted int             `json:"total_tasks_completed" db:"total_tasks_completed"`
	SuccessRate         float64         `json:"success_rate" db:"success_rate"`
	AvgCompletionTimeMs int             `json:"avg_completion_time_ms" db:"avg_completion_time_ms"`
	LastActiveAt        time.Time       `json:"last_active_at" db:"last_active_at"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
}

// Task represents a task in the queue
type Task struct {
	TaskID           string          `json:"task_id" db:"task_id"`
	TaskType         string          `json:"task_type" db:"task_type"`
	Priority         int             `json:"priority" db:"priority"`
	Status           string          `json:"status" db:"status"` // pending, assigned, in_progress, completed, failed
	CreatedBy        string          `json:"created_by" db:"created_by"`
	AssignedToAgent  *string         `json:"assigned_to_agent,omitempty" db:"assigned_to_agent"`
	FilePaths        json.RawMessage `json:"file_paths" db:"file_paths"`
	DependsOnTaskIDs json.RawMessage `json:"depends_on_task_ids" db:"depends_on_task_ids"`
	Description      string          `json:"description" db:"description"`
	Context          JSONB           `json:"context" db:"context"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	AssignedAt       *time.Time      `json:"assigned_at,omitempty" db:"assigned_at"`
	StartedAt        *time.Time      `json:"started_at,omitempty" db:"started_at"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	Deadline         *time.Time      `json:"deadline,omitempty" db:"deadline"`
	Result           JSONB           `json:"result" db:"result"`
	ErrorLog         *string         `json:"error_log,omitempty" db:"error_log"`
	RetryCount       int             `json:"retry_count" db:"retry_count"`
}

// FileRegistry represents a file tracked in the system
type FileRegistry struct {
	FileID         string          `json:"file_id" db:"file_id"`
	FilePath       string          `json:"file_path" db:"file_path"`
	FileType       *string         `json:"file_type,omitempty" db:"file_type"`
	FileHash       *string         `json:"file_hash,omitempty" db:"file_hash"`
	OwnerAgent     *string         `json:"owner_agent,omitempty" db:"owner_agent"`
	CreatedBy      *string         `json:"created_by,omitempty" db:"created_by"`
	LastModifiedBy *string         `json:"last_modified_by,omitempty" db:"last_modified_by"`
	Status         string          `json:"status" db:"status"` // draft, review, complete, deprecated, broken
	Purpose        *string         `json:"purpose,omitempty" db:"purpose"`
	Dependencies   json.RawMessage `json:"dependencies" db:"dependencies"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	LastTestedAt   *time.Time      `json:"last_tested_at,omitempty" db:"last_tested_at"`
	TestStatus     string          `json:"test_status" db:"test_status"`
	BuildRequired  bool            `json:"build_required" db:"build_required"`
	Deployed       bool            `json:"deployed" db:"deployed"`
	SizeBytes      *int64          `json:"size_bytes,omitempty" db:"size_bytes"`
	LineCount      *int            `json:"line_count,omitempty" db:"line_count"`
	Language       *string         `json:"language,omitempty" db:"language"`
}

// AgentTaskHistory represents task execution history
type AgentTaskHistory struct {
	HistoryID      string    `json:"history_id" db:"history_id"`
	AgentName      string    `json:"agent_name" db:"agent_name"`
	TaskID         *string   `json:"task_id,omitempty" db:"task_id"`
	TaskType       *string   `json:"task_type,omitempty" db:"task_type"`
	FileID         *string   `json:"file_id,omitempty" db:"file_id"`
	ActionType     *string   `json:"action_type,omitempty" db:"action_type"`
	Success        bool      `json:"success" db:"success"`
	DurationMs     *int      `json:"duration_ms,omitempty" db:"duration_ms"`
	ErrorMessage   *string   `json:"error_message,omitempty" db:"error_message"`
	LearnedPattern *string   `json:"learned_pattern,omitempty" db:"learned_pattern"`
	CostTokens     *int      `json:"cost_tokens,omitempty" db:"cost_tokens"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// BuildHistory represents build execution history
type BuildHistory struct {
	BuildID       string    `json:"build_id" db:"build_id"`
	BuildNumber   int       `json:"build_number" db:"build_number"`
	TriggeredBy   *string   `json:"triggered_by,omitempty" db:"triggered_by"`
	FilesChanged  JSONB     `json:"files_changed" db:"files_changed"`
	Success       bool      `json:"success" db:"success"`
	DurationMs    *int      `json:"duration_ms,omitempty" db:"duration_ms"`
	ErrorLog      *string   `json:"error_log,omitempty" db:"error_log"`
	Warnings      *string   `json:"warnings,omitempty" db:"warnings"`
	BinaryHash    *string   `json:"binary_hash,omitempty" db:"binary_hash"`
	Deployed      bool      `json:"deployed" db:"deployed"`
	GitCommitHash *string   `json:"git_commit_hash,omitempty" db:"git_commit_hash"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// CreateTaskRequest is the request body for creating a new task
type CreateTaskRequest struct {
	TaskType         string                 `json:"task_type" binding:"required"`
	Description      string                 `json:"description" binding:"required"`
	Priority         int                    `json:"priority"`
	FilePaths        []string               `json:"file_paths"`
	DependsOnTaskIDs []string               `json:"depends_on_task_ids"`
	Context          map[string]interface{} `json:"context"`
	Deadline         *time.Time             `json:"deadline"`
}

// AssignTaskRequest is the request body for assigning a task to an agent
type AssignTaskRequest struct {
	AgentName string `json:"agent_name" binding:"required"`
}
