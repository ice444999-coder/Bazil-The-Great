package grpo

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// ============================================
// GRPO UPDATER - Background learning loop
// Runs every N minutes to apply rewards to biases
// ============================================

// Updater manages the background GRPO learning loop
type Updater struct {
	agent    *Agent
	interval time.Duration
	stopChan chan bool
}

// NewUpdater creates a new GRPO updater
func NewUpdater(db *gorm.DB, learningRate float64, intervalMinutes int) *Updater {
	if intervalMinutes == 0 {
		intervalMinutes = 10 // Default: update every 10 minutes
	}

	return &Updater{
		agent:    NewAgent(db, learningRate),
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan bool),
	}
}

// Start begins the background learning loop
func (u *Updater) Start() error {
	// Load existing biases from database
	if err := u.agent.LoadBiases(); err != nil {
		return err
	}

	log.Printf("[GRPO][UPDATER] Starting background learning loop (interval: %v)", u.interval)

	go func() {
		ticker := time.NewTicker(u.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				u.performUpdate()
			case <-u.stopChan:
				log.Println("[GRPO][UPDATER] Stopping background learning loop")
				return
			}
		}
	}()

	return nil
}

// Stop halts the background learning loop
func (u *Updater) Stop() {
	close(u.stopChan)
}

// performUpdate executes one learning iteration
func (u *Updater) performUpdate() {
	log.Println("[GRPO][UPDATER] Starting learning iteration...")

	start := time.Now()

	// Apply pending rewards to biases
	updated, err := u.agent.UpdateBiases()
	if err != nil {
		log.Printf("[GRPO][UPDATER][ERROR] Failed to update biases: %v", err)
		return
	}

	duration := time.Since(start)

	if updated > 0 {
		stats := u.agent.GetStats()
		log.Printf("[GRPO][UPDATER] ✅ Learning iteration complete (%.2fs)", duration.Seconds())
		log.Printf("   Rewards applied: %d", updated)
		log.Printf("   Total biases: %v", stats["total_biases"])
		log.Printf("   Pending rewards: %v", stats["pending_rewards"])
		log.Printf("   Average reward: %.4f", stats["average_reward"])

		// Log top biases
		topBiases := u.agent.GetTopBiases(5)
		if len(topBiases) > 0 {
			log.Println("   Top 5 biased tokens:")
			for i, bias := range topBiases {
				log.Printf("     %d. '%s': %.4f (updates: %d, cumulative: %.4f)",
					i+1, bias.TokenText, bias.BiasValue, bias.UpdateCount, bias.CumulativeReward)
			}
		}
	} else {
		log.Printf("[GRPO][UPDATER] No pending rewards to apply (checked in %.2fs)", duration.Seconds())
	}
}

// ForceUpdate triggers an immediate learning iteration
func (u *Updater) ForceUpdate() (int, error) {
	log.Println("[GRPO][UPDATER] Manual learning iteration triggered")
	return u.agent.UpdateBiases()
}

// GetAgent returns the underlying GRPO agent
func (u *Updater) GetAgent() *Agent {
	return u.agent
}
