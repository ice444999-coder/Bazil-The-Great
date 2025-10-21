/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// StringArray custom type for PostgreSQL array handling
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// MissionProgress tracks SOLACE's mission objectives and progress
// This table stores the overall progress of the ARES takeover mission
type MissionProgress struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	Phase1          int         `json:"phase1"`                            // UI/Trading System (0-100%)
	Phase2          int         `json:"phase2"`                            // Agent Coordination (0-100%)
	Phase3          int         `json:"phase3"`                            // Self-Evolution (0-100%)
	CurrentFocus    string      `json:"current_focus"`                     // What SOLACE is currently working on
	Blockers        StringArray `gorm:"type:jsonb" json:"blockers"`        // Current blockers
	CompletedTasks  StringArray `gorm:"type:jsonb" json:"completed_tasks"` // Completed tasks
	NextMilestone   string      `json:"next_milestone"`                    // Next major milestone
	ConfidenceScore float64     `json:"confidence_score"`                  // How confident SOLACE is (0-1.0)
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (MissionProgress) TableName() string {
	return "solace_mission_progress"
}
