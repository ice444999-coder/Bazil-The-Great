package agent

import (
	"ares_api/internal/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MissionProgress tracks ARES/SOLACE mission objectives and progress
// Integrates with solace_mission_progress table from docs
type MissionProgress struct {
	db *gorm.DB
}

// MissionStatus represents the current state of the takeover mission
type MissionStatus struct {
	Phase1          int       `json:"phase1"` // UI/Trading System (0-100%)
	Phase2          int       `json:"phase2"` // Agent Coordination (0-100%)
	Phase3          int       `json:"phase3"` // Self-Evolution (0-100%)
	CurrentFocus    string    `json:"current_focus"`
	LastUpdated     time.Time `json:"last_updated"`
	Blockers        []string  `json:"blockers"`
	CompletedTasks  []string  `json:"completed_tasks"`
	NextMilestone   string    `json:"next_milestone"`
	ConfidenceScore float64   `json:"confidence_score"` // 0-1.0
}

// NewMissionProgress initializes the mission progress tracker
func NewMissionProgress(db *gorm.DB) *MissionProgress {
	return &MissionProgress{db: db}
}

// GetStatus fetches current mission status from database
func (mp *MissionProgress) GetStatus() MissionStatus {
	var status models.MissionProgress
	mp.db.Order("updated_at DESC").First(&status)

	return MissionStatus{
		Phase1:          status.Phase1,
		Phase2:          status.Phase2,
		Phase3:          status.Phase3,
		CurrentFocus:    status.CurrentFocus,
		LastUpdated:     status.UpdatedAt,
		Blockers:        status.Blockers,
		CompletedTasks:  status.CompletedTasks,
		NextMilestone:   status.NextMilestone,
		ConfidenceScore: status.ConfidenceScore,
	}
}

// UpdatePhase updates a specific phase progress
func (mp *MissionProgress) UpdatePhase(phase int, progress int) error {
	var status models.MissionProgress
	mp.db.FirstOrCreate(&status)

	switch phase {
	case 1:
		status.Phase1 = progress
	case 2:
		status.Phase2 = progress
	case 3:
		status.Phase3 = progress
	}

	status.UpdatedAt = time.Now()
	return mp.db.Save(&status).Error
}

// AddBlocker logs a new blocker to mission progress
func (mp *MissionProgress) AddBlocker(blocker string) error {
	var status models.MissionProgress
	mp.db.FirstOrCreate(&status)

	status.Blockers = append(status.Blockers, blocker)
	status.UpdatedAt = time.Now()
	return mp.db.Save(&status).Error
}

// CompleteTask marks a task as completed
func (mp *MissionProgress) CompleteTask(task string) error {
	var status models.MissionProgress
	mp.db.FirstOrCreate(&status)

	status.CompletedTasks = append(status.CompletedTasks, task)
	status.UpdatedAt = time.Now()
	return mp.db.Save(&status).Error
}

// SetFocus updates the current focus area
func (mp *MissionProgress) SetFocus(focus string) error {
	var status models.MissionProgress
	mp.db.FirstOrCreate(&status)

	status.CurrentFocus = focus
	status.UpdatedAt = time.Now()
	return mp.db.Save(&status).Error
}

// GetProgressSummary returns a human-readable summary
func (mp *MissionProgress) GetProgressSummary() string {
	status := mp.GetStatus()
	return fmt.Sprintf("Phase1: %d%% | Phase2: %d%% | Phase3: %d%% | Focus: %s | Confidence: %.2f",
		status.Phase1, status.Phase2, status.Phase3, status.CurrentFocus, status.ConfidenceScore)
}
