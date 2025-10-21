/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package grpo

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"gorm.io/gorm"
)

// ============================================
// GRPO AGENT - Learning from trading outcomes
// ============================================

// Bias represents a learned token bias
type Bias struct {
	ID               uint      `gorm:"primaryKey"`
	TokenText        string    `gorm:"column:token_text;uniqueIndex;size:100"`
	TokenID          *int      `gorm:"column:token_id"`
	BiasValue        float64   `gorm:"column:bias_value;type:decimal(10,6)"`
	UpdateCount      int       `gorm:"column:update_count"`
	LastReward       *float64  `gorm:"column:last_reward;type:decimal(10,6)"`
	CumulativeReward float64   `gorm:"column:cumulative_reward;type:decimal(15,6)"`
	LastUpdated      time.Time `gorm:"column:last_updated"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (Bias) TableName() string {
	return "grpo_biases"
}

// Reward represents a training signal from a trading decision
type Reward struct {
	ID                  uint       `gorm:"primaryKey"`
	TraceID             *int       `gorm:"column:trace_id"`
	TradeID             *int       `gorm:"column:trade_id"`
	RewardValue         float64    `gorm:"column:reward_value;type:decimal(10,6)"`
	OutcomeQuality      *float64   `gorm:"column:outcome_quality;type:decimal(5,2)"`
	ProfitLoss          *float64   `gorm:"column:profit_loss;type:decimal(20,8)"`
	WinRateContribution *float64   `gorm:"column:win_rate_contribution;type:decimal(5,2)"`
	DecisionTokens      string     `gorm:"column:decision_tokens;type:text[]"` // PostgreSQL array
	ConfidenceScore     *float64   `gorm:"column:confidence_score;type:decimal(5,2)"`
	ExecutionTimeMS     *int       `gorm:"column:execution_time_ms"`
	RewardType          string     `gorm:"column:reward_type;size:50"`
	RewardWeight        float64    `gorm:"column:reward_weight;type:decimal(5,4);default:1.0"`
	LearningIteration   int        `gorm:"column:learning_iteration;default:0"`
	AppliedToBiases     bool       `gorm:"column:applied_to_biases;default:false"`
	AppliedAt           *time.Time `gorm:"column:applied_at"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
}

func (Reward) TableName() string {
	return "grpo_rewards"
}

// Agent manages GRPO learning
type Agent struct {
	db           *gorm.DB
	learningRate float64
	biases       map[string]*Bias // In-memory cache of biases
}

// NewAgent creates a new GRPO agent
func NewAgent(db *gorm.DB, learningRate float64) *Agent {
	if learningRate == 0 {
		learningRate = 0.01 // Default learning rate
	}

	return &Agent{
		db:           db,
		learningRate: learningRate,
		biases:       make(map[string]*Bias),
	}
}

// LoadBiases loads all biases from database into memory
func (a *Agent) LoadBiases() error {
	var biases []Bias
	if err := a.db.Find(&biases).Error; err != nil {
		return fmt.Errorf("failed to load biases: %w", err)
	}

	a.biases = make(map[string]*Bias)
	for i := range biases {
		a.biases[biases[i].TokenText] = &biases[i]
	}

	log.Printf("[GRPO] Loaded %d token biases from database", len(a.biases))
	return nil
}

// SaveBiases saves all modified biases back to database
func (a *Agent) SaveBiases() error {
	count := 0
	for _, bias := range a.biases {
		if bias.ID > 0 {
			// Update existing
			if err := a.db.Save(bias).Error; err != nil {
				log.Printf("[GRPO][ERROR] Failed to save bias for '%s': %v", bias.TokenText, err)
				continue
			}
		} else {
			// Create new
			if err := a.db.Create(bias).Error; err != nil {
				log.Printf("[GRPO][ERROR] Failed to create bias for '%s': %v", bias.TokenText, err)
				continue
			}
		}
		count++
	}

	log.Printf("[GRPO] Saved %d token biases to database", count)
	return nil
}

// RecordReward records a reward from a trading decision
func (a *Agent) RecordReward(tradeID, traceID int, profitLoss, tradeSize, confidence float64, tokens []string) error {
	// Calculate reward value (-1 to +1)
	roi := profitLoss / tradeSize
	rewardValue := math.Copysign(math.Min(math.Abs(roi)*10, 1.0), roi) * (confidence / 100.0)

	// Determine outcome quality (0-100)
	quality := 50.0 + (roi * 500) // ROI of +0.1 = 100 quality
	if quality > 100 {
		quality = 100
	}
	if quality < 0 {
		quality = 0
	}

	// Convert tokens to PostgreSQL array format
	tokensJSON, _ := json.Marshal(tokens)

	reward := Reward{
		TradeID:           &tradeID,
		TraceID:           &traceID,
		RewardValue:       rewardValue,
		OutcomeQuality:    &quality,
		ProfitLoss:        &profitLoss,
		ConfidenceScore:   &confidence,
		DecisionTokens:    string(tokensJSON), // Store as JSON array
		RewardType:        "composite",
		RewardWeight:      1.0,
		LearningIteration: 0,
		AppliedToBiases:   false,
		CreatedAt:         time.Now(),
	}

	if err := a.db.Create(&reward).Error; err != nil {
		return fmt.Errorf("failed to record reward: %w", err)
	}

	log.Printf("[GRPO] Recorded reward %.4f for trade %d (P&L: $%.2f, ROI: %.2f%%)",
		rewardValue, tradeID, profitLoss, roi*100)

	return nil
}

// UpdateBiases applies pending rewards to token biases
func (a *Agent) UpdateBiases() (int, error) {
	// Get unapplied rewards
	var rewards []Reward
	if err := a.db.Where("applied_to_biases = ?", false).
		Order("created_at ASC").
		Limit(100). // Process in batches
		Find(&rewards).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch rewards: %w", err)
	}

	if len(rewards) == 0 {
		log.Println("[GRPO] No pending rewards to apply")
		return 0, nil
	}

	updatedCount := 0
	for _, reward := range rewards {
		// Parse tokens from JSON
		var tokens []string
		if err := json.Unmarshal([]byte(reward.DecisionTokens), &tokens); err != nil {
			log.Printf("[GRPO][WARN] Failed to parse tokens for reward %d: %v", reward.ID, err)
			continue
		}

		// Update bias for each token
		for _, token := range tokens {
			if token == "" {
				continue
			}

			bias, exists := a.biases[token]
			if !exists {
				// Create new bias
				bias = &Bias{
					TokenText:        token,
					BiasValue:        0.0,
					UpdateCount:      0,
					CumulativeReward: 0.0,
					CreatedAt:        time.Now(),
				}
				a.biases[token] = bias
			}

			// Apply gradient update: bias += learning_rate * reward
			delta := a.learningRate * reward.RewardValue
			bias.BiasValue += delta
			bias.UpdateCount++
			bias.LastReward = &reward.RewardValue
			bias.CumulativeReward += reward.RewardValue
			bias.LastUpdated = time.Now()

			// Clamp bias to [-1, 1]
			if bias.BiasValue > 1.0 {
				bias.BiasValue = 1.0
			}
			if bias.BiasValue < -1.0 {
				bias.BiasValue = -1.0
			}
		}

		// Mark reward as applied
		now := time.Now()
		reward.AppliedToBiases = true
		reward.AppliedAt = &now
		if err := a.db.Save(&reward).Error; err != nil {
			log.Printf("[GRPO][ERROR] Failed to mark reward %d as applied: %v", reward.ID, err)
		}

		updatedCount++
	}

	// Save updated biases to database
	if err := a.SaveBiases(); err != nil {
		return updatedCount, fmt.Errorf("failed to save biases: %w", err)
	}

	log.Printf("[GRPO] Applied %d rewards, updated %d unique tokens", updatedCount, len(a.biases))
	return updatedCount, nil
}

// GetBias returns the current bias for a token
func (a *Agent) GetBias(token string) float64 {
	if bias, exists := a.biases[token]; exists {
		return bias.BiasValue
	}
	return 0.0
}

// GetTopBiases returns tokens with highest absolute bias values
func (a *Agent) GetTopBiases(limit int) []Bias {
	type biasWithAbs struct {
		bias     *Bias
		absValue float64
	}

	biases := make([]biasWithAbs, 0, len(a.biases))
	for _, bias := range a.biases {
		biases = append(biases, biasWithAbs{
			bias:     bias,
			absValue: math.Abs(bias.BiasValue),
		})
	}

	// Sort by absolute value (descending)
	for i := 0; i < len(biases)-1; i++ {
		for j := i + 1; j < len(biases); j++ {
			if biases[j].absValue > biases[i].absValue {
				biases[i], biases[j] = biases[j], biases[i]
			}
		}
	}

	// Return top N
	if limit > len(biases) {
		limit = len(biases)
	}

	result := make([]Bias, limit)
	for i := 0; i < limit; i++ {
		result[i] = *biases[i].bias
	}

	return result
}

// GetStats returns current learning statistics
func (a *Agent) GetStats() map[string]interface{} {
	var totalRewards int64
	var appliedRewards int64
	var avgReward float64

	a.db.Model(&Reward{}).Count(&totalRewards)
	a.db.Model(&Reward{}).Where("applied_to_biases = ?", true).Count(&appliedRewards)
	a.db.Model(&Reward{}).Select("AVG(reward_value)").Row().Scan(&avgReward)

	return map[string]interface{}{
		"total_biases":     len(a.biases),
		"total_rewards":    totalRewards,
		"applied_rewards":  appliedRewards,
		"pending_rewards":  totalRewards - appliedRewards,
		"average_reward":   avgReward,
		"learning_rate":    a.learningRate,
		"biases_in_memory": len(a.biases),
	}
}
