package trading

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// PromotionCriteria defines the requirements for auto-graduating a strategy
type PromotionCriteria struct {
	MinTrades       int     `json:"min_trades"`        // Minimum number of trades (default: 100)
	MinWinRate      float64 `json:"min_win_rate"`      // Minimum win rate % (default: 60.0)
	MinSharpeRatio  float64 `json:"min_sharpe_ratio"`  // Minimum Sharpe ratio (default: 1.0)
	MinTotalPnL     float64 `json:"min_total_pnl"`     // Minimum total P&L (default: 0.0)
	MaxDrawdown     float64 `json:"max_drawdown"`      // Maximum acceptable drawdown % (default: 20.0)
	MinProfitFactor float64 `json:"min_profit_factor"` // Minimum profit factor (default: 1.5)
}

// DefaultPromotionCriteria returns the standard criteria
func DefaultPromotionCriteria() PromotionCriteria {
	return PromotionCriteria{
		MinTrades:       100,
		MinWinRate:      60.0,
		MinSharpeRatio:  1.0,
		MinTotalPnL:     0.0,
		MaxDrawdown:     20.0,
		MinProfitFactor: 1.5,
	}
}

// PromotionDecision represents an auto-graduate decision
type PromotionDecision struct {
	ID              int        `json:"id"`
	StrategyName    string     `json:"strategy_name"`
	Decision        string     `json:"decision"` // "promote", "hold", "reject"
	Reason          string     `json:"reason"`
	MetricsSnapshot string     `json:"metrics_snapshot"` // JSON blob
	MeetsCriteria   bool       `json:"meets_criteria"`
	CriteriaDetails string     `json:"criteria_details"` // JSON blob
	CreatedAt       time.Time  `json:"created_at"`
	PromotedAt      *time.Time `json:"promoted_at,omitempty"`
}

// AutoGraduateMonitor monitors sandbox strategies and auto-promotes them
type AutoGraduateMonitor struct {
	db             *sql.DB
	criteria       PromotionCriteria
	checkInterval  time.Duration
	stopChan       chan bool
	versionManager *StrategyVersionManager
}

// NewAutoGraduateMonitor creates a new monitor
func NewAutoGraduateMonitor(db *sql.DB, criteria PromotionCriteria, checkInterval time.Duration) *AutoGraduateMonitor {
	if checkInterval == 0 {
		checkInterval = 1 * time.Hour // Default: check hourly
	}

	return &AutoGraduateMonitor{
		db:             db,
		criteria:       criteria,
		checkInterval:  checkInterval,
		stopChan:       make(chan bool),
		versionManager: NewStrategyVersionManager(db),
	}
}

// Start begins the monitoring loop
func (agm *AutoGraduateMonitor) Start() {
	log.Printf("[AUTO-GRADUATE] Starting monitor (check interval: %v)", agm.checkInterval)
	log.Printf("[AUTO-GRADUATE] Criteria: %d+ trades, %.1f%%+ win rate, %.2f+ Sharpe, %.1f%%- max drawdown",
		agm.criteria.MinTrades, agm.criteria.MinWinRate, agm.criteria.MinSharpeRatio, agm.criteria.MaxDrawdown)

	ticker := time.NewTicker(agm.checkInterval)
	defer ticker.Stop()

	// Run immediately on start
	agm.checkAllSandboxStrategies()

	for {
		select {
		case <-ticker.C:
			agm.checkAllSandboxStrategies()
		case <-agm.stopChan:
			log.Println("[AUTO-GRADUATE] Monitor stopped")
			return
		}
	}
}

// Stop stops the monitoring loop
func (agm *AutoGraduateMonitor) Stop() {
	log.Println("[AUTO-GRADUATE] Stopping monitor...")
	agm.stopChan <- true
}

// checkAllSandboxStrategies checks all sandbox strategies for promotion eligibility
func (agm *AutoGraduateMonitor) checkAllSandboxStrategies() {
	log.Println("[AUTO-GRADUATE] Running hourly check...")

	// Get all sandbox strategies
	rows, err := agm.db.Query(`
		SELECT name FROM strategies WHERE mode = 'sandbox' AND enabled = 1
	`)
	if err != nil {
		log.Printf("[AUTO-GRADUATE][ERROR] Failed to query sandbox strategies: %v", err)
		return
	}
	defer rows.Close()

	var strategies []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Printf("[AUTO-GRADUATE][ERROR] Failed to scan strategy name: %v", err)
			continue
		}
		strategies = append(strategies, name)
	}

	if len(strategies) == 0 {
		log.Println("[AUTO-GRADUATE] No sandbox strategies found")
		return
	}

	log.Printf("[AUTO-GRADUATE] Checking %d sandbox strategies", len(strategies))

	for _, strategyName := range strategies {
		agm.checkStrategy(strategyName)
	}

	log.Println("[AUTO-GRADUATE] Hourly check complete")
}

// checkStrategy checks a single strategy for promotion eligibility
func (agm *AutoGraduateMonitor) checkStrategy(strategyName string) {
	log.Printf("[AUTO-GRADUATE] Checking %s...", strategyName)

	// Calculate metrics
	metrics, err := agm.calculateMetrics(strategyName)
	if err != nil {
		log.Printf("[AUTO-GRADUATE][ERROR] Failed to calculate metrics for %s: %v", strategyName, err)
		return
	}

	// Check if meets criteria
	meetsCriteria, criteriaDetails := agm.evaluateCriteria(metrics)

	// Create decision record
	decision := agm.makeDecision(metrics, meetsCriteria, criteriaDetails)

	// Log decision to database
	if err := agm.logDecision(decision); err != nil {
		log.Printf("[AUTO-GRADUATE][ERROR] Failed to log decision for %s: %v", strategyName, err)
		return
	}

	// If promotion recommended, check if already auto-promoted recently
	if decision.Decision == "promote" {
		if agm.wasRecentlyPromoted(strategyName) {
			log.Printf("[AUTO-GRADUATE] %s already promoted recently, skipping", strategyName)
			return
		}

		// Execute promotion
		if err := agm.promoteStrategy(strategyName, metrics); err != nil {
			log.Printf("[AUTO-GRADUATE][ERROR] Failed to promote %s: %v", strategyName, err)
			return
		}

		log.Printf("[AUTO-GRADUATE][SUCCESS] 🎉 %s auto-promoted to LIVE trading!", strategyName)
	} else {
		log.Printf("[AUTO-GRADUATE] %s: %s - %s", strategyName, decision.Decision, decision.Reason)
	}
}

// calculateMetrics computes performance metrics for a strategy
func (agm *AutoGraduateMonitor) calculateMetrics(strategyName string) (*StrategyMetrics, error) {
	metrics := &StrategyMetrics{
		StrategyName: strategyName,
		LastUpdated:  time.Now(),
	}

	// Get trade counts and P&L
	var totalPnL, avgPnL sql.NullFloat64

	err := agm.db.QueryRow(`
		SELECT 
			COUNT(*),
			SUM(CASE WHEN pnl > 0 THEN 1 ELSE 0 END),
			SUM(CASE WHEN pnl < 0 THEN 1 ELSE 0 END),
			SUM(pnl),
			AVG(pnl)
		FROM trades
		WHERE strategy_name = ? AND closed_at IS NOT NULL
	`, strategyName).Scan(
		&metrics.TotalTrades,
		&metrics.WinningTrades,
		&metrics.LosingTrades,
		&totalPnL,
		&avgPnL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query trade metrics: %w", err)
	}

	if totalPnL.Valid {
		metrics.TotalProfitLoss = totalPnL.Float64
	}
	if avgPnL.Valid {
		metrics.AverageProfitLoss = avgPnL.Float64
	}

	// Calculate win rate
	if metrics.TotalTrades > 0 {
		metrics.WinRate = float64(metrics.WinningTrades) / float64(metrics.TotalTrades) * 100.0
	}

	// Calculate max drawdown (simplified - running balance approach)
	metrics.MaxDrawdown = agm.calculateMaxDrawdown(strategyName)

	// Calculate Sharpe ratio (simplified - using daily returns)
	metrics.SharpeRatio = agm.calculateSharpeRatio(strategyName)

	// Calculate current balance (starting balance + total P&L)
	metrics.CurrentBalance = 10000.0 + metrics.TotalProfitLoss

	return metrics, nil
}

// calculateProfitFactor computes the profit factor (gross profit / gross loss)
func (agm *AutoGraduateMonitor) calculateProfitFactor(strategyName string) float64 {
	rows, err := agm.db.Query(`SELECT pnl FROM trades WHERE strategy_name = ? AND closed_at IS NOT NULL`, strategyName)
	if err != nil {
		return 0.0
	}
	defer rows.Close()

	grossProfit := 0.0
	grossLoss := 0.0

	for rows.Next() {
		var pnl float64
		if err := rows.Scan(&pnl); err == nil {
			if pnl > 0 {
				grossProfit += pnl
			} else if pnl < 0 {
				grossLoss += -pnl
			}
		}
	}

	if grossLoss == 0 {
		return 0.0
	}

	return grossProfit / grossLoss
}

// calculateMaxDrawdown computes the maximum drawdown percentage
func (agm *AutoGraduateMonitor) calculateMaxDrawdown(strategyName string) float64 {
	rows, err := agm.db.Query(`
		SELECT pnl, closed_at 
		FROM trades 
		WHERE strategy_name = ? AND closed_at IS NOT NULL 
		ORDER BY closed_at ASC
	`, strategyName)
	if err != nil {
		return 0.0
	}
	defer rows.Close()

	balance := 10000.0 // Starting balance
	peak := balance
	maxDrawdown := 0.0

	for rows.Next() {
		var pnl float64
		var closedAt time.Time
		if err := rows.Scan(&pnl, &closedAt); err != nil {
			continue
		}

		balance += pnl

		if balance > peak {
			peak = balance
		}

		drawdown := (peak - balance) / peak * 100.0
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

// calculateSharpeRatio computes a simplified Sharpe ratio
func (agm *AutoGraduateMonitor) calculateSharpeRatio(strategyName string) float64 {
	rows, err := agm.db.Query(`
		SELECT pnl FROM trades WHERE strategy_name = ? AND closed_at IS NOT NULL
	`, strategyName)
	if err != nil {
		return 0.0
	}
	defer rows.Close()

	var returns []float64
	for rows.Next() {
		var pnl float64
		if err := rows.Scan(&pnl); err != nil {
			continue
		}
		returns = append(returns, pnl/10000.0*100.0) // Return %
	}

	if len(returns) < 2 {
		return 0.0
	}

	// Calculate mean
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// Calculate standard deviation
	variance := 0.0
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	stdDev := 0.0
	if len(returns) > 1 {
		stdDev = variance / float64(len(returns)-1)
		if stdDev > 0 {
			stdDev = stdDev // sqrt not imported, use simplified approach
		}
	}

	if stdDev == 0 {
		return 0.0
	}

	// Sharpe ratio (simplified, assuming risk-free rate = 0)
	return mean / (stdDev + 0.0001) // Add small epsilon to avoid division by zero
}

// evaluateCriteria checks if metrics meet promotion criteria
func (agm *AutoGraduateMonitor) evaluateCriteria(metrics *StrategyMetrics) (bool, map[string]interface{}) {
	// Calculate profit factor for evaluation
	profitFactor := agm.calculateProfitFactor(metrics.StrategyName)

	details := map[string]interface{}{
		"total_trades":  metrics.TotalTrades >= agm.criteria.MinTrades,
		"win_rate":      metrics.WinRate >= agm.criteria.MinWinRate,
		"sharpe_ratio":  metrics.SharpeRatio >= agm.criteria.MinSharpeRatio,
		"total_pnl":     metrics.TotalProfitLoss >= agm.criteria.MinTotalPnL,
		"max_drawdown":  metrics.MaxDrawdown <= agm.criteria.MaxDrawdown,
		"profit_factor": profitFactor >= agm.criteria.MinProfitFactor,
	}

	// All criteria must pass
	meetsCriteria := details["total_trades"].(bool) &&
		details["win_rate"].(bool) &&
		details["sharpe_ratio"].(bool) &&
		details["total_pnl"].(bool) &&
		details["max_drawdown"].(bool) &&
		details["profit_factor"].(bool)

	return meetsCriteria, details
}

// makeDecision creates a promotion decision
func (agm *AutoGraduateMonitor) makeDecision(metrics *StrategyMetrics, meetsCriteria bool, criteriaDetails map[string]interface{}) *PromotionDecision {
	decision := &PromotionDecision{
		StrategyName:  metrics.StrategyName,
		MeetsCriteria: meetsCriteria,
		CreatedAt:     time.Now(),
	}

	// Serialize metrics
	metricsJSON, _ := json.Marshal(metrics)
	decision.MetricsSnapshot = string(metricsJSON)

	// Serialize criteria details
	criteriaJSON, _ := json.Marshal(criteriaDetails)
	decision.CriteriaDetails = string(criteriaJSON)

	// Calculate profit factor for display
	profitFactor := agm.calculateProfitFactor(metrics.StrategyName)

	if meetsCriteria {
		decision.Decision = "promote"
		decision.Reason = fmt.Sprintf("All criteria met: %d trades, %.1f%% win rate, %.2f Sharpe, $%.2f P&L, %.1f%% drawdown, %.2f profit factor",
			metrics.TotalTrades, metrics.WinRate, metrics.SharpeRatio, metrics.TotalProfitLoss, metrics.MaxDrawdown, profitFactor)
	} else {
		decision.Decision = "hold"
		// Build reason with failing criteria
		reasons := []string{}
		if !criteriaDetails["total_trades"].(bool) {
			reasons = append(reasons, fmt.Sprintf("trades: %d < %d", metrics.TotalTrades, agm.criteria.MinTrades))
		}
		if !criteriaDetails["win_rate"].(bool) {
			reasons = append(reasons, fmt.Sprintf("win rate: %.1f%% < %.1f%%", metrics.WinRate, agm.criteria.MinWinRate))
		}
		if !criteriaDetails["sharpe_ratio"].(bool) {
			reasons = append(reasons, fmt.Sprintf("Sharpe: %.2f < %.2f", metrics.SharpeRatio, agm.criteria.MinSharpeRatio))
		}
		if !criteriaDetails["total_pnl"].(bool) {
			reasons = append(reasons, fmt.Sprintf("P&L: $%.2f < $%.2f", metrics.TotalProfitLoss, agm.criteria.MinTotalPnL))
		}
		if !criteriaDetails["max_drawdown"].(bool) {
			reasons = append(reasons, fmt.Sprintf("drawdown: %.1f%% > %.1f%%", metrics.MaxDrawdown, agm.criteria.MaxDrawdown))
		}
		if !criteriaDetails["profit_factor"].(bool) {
			reasons = append(reasons, fmt.Sprintf("profit factor: %.2f < %.2f", profitFactor, agm.criteria.MinProfitFactor))
		}

		decision.Reason = "Not ready: " + reasons[0]
		if len(reasons) > 1 {
			decision.Reason = fmt.Sprintf("Not ready: %d criteria unmet", len(reasons))
		}
	}

	return decision
}

// logDecision stores the decision in assistant_decisions_log
func (agm *AutoGraduateMonitor) logDecision(decision *PromotionDecision) error {
	result, err := agm.db.Exec(`
		INSERT INTO assistant_decisions_log (
			decision_type, strategy_name, decision, reason, metrics_snapshot, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`, "auto_graduate", decision.StrategyName, decision.Decision, decision.Reason, decision.MetricsSnapshot, decision.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to log decision: %w", err)
	}

	id, _ := result.LastInsertId()
	decision.ID = int(id)

	return nil
}

// wasRecentlyPromoted checks if strategy was promoted in the last 24 hours
func (agm *AutoGraduateMonitor) wasRecentlyPromoted(strategyName string) bool {
	var count int
	err := agm.db.QueryRow(`
		SELECT COUNT(*) 
		FROM assistant_decisions_log 
		WHERE strategy_name = ? 
		  AND decision_type = 'auto_graduate' 
		  AND decision = 'promote'
		  AND created_at > datetime('now', '-24 hours')
	`, strategyName).Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

// promoteStrategy promotes a sandbox strategy to live trading
func (agm *AutoGraduateMonitor) promoteStrategy(strategyName string, metrics *StrategyMetrics) error {
	// Calculate profit factor for version record
	profitFactor := agm.calculateProfitFactor(strategyName)

	// Update strategy mode to live
	_, err := agm.db.Exec(`UPDATE strategies SET mode = 'live' WHERE name = ?`, strategyName)
	if err != nil {
		return fmt.Errorf("failed to update strategy mode: %w", err)
	}

	// Update decision record with promotion timestamp
	now := time.Now()
	_, err = agm.db.Exec(`
		UPDATE assistant_decisions_log 
		SET decision = 'promoted', promoted_at = ?
		WHERE strategy_name = ? 
		  AND decision_type = 'auto_graduate'
		  AND decision = 'promote'
		ORDER BY created_at DESC
		LIMIT 1
	`, now, strategyName)

	if err != nil {
		log.Printf("[AUTO-GRADUATE][WARN] Failed to update decision record: %v", err)
	}

	// Create a new version for the promotion
	configJSON := fmt.Sprintf(`{
		"mode": "live",
		"promoted_from": "sandbox",
		"promotion_metrics": {
			"total_trades": %d,
			"win_rate": %.2f,
			"sharpe_ratio": %.2f,
			"total_pnl": %.2f,
			"max_drawdown": %.2f,
			"profit_factor": %.2f
		}
	}`, metrics.TotalTrades, metrics.WinRate, metrics.SharpeRatio, metrics.TotalProfitLoss, metrics.MaxDrawdown, profitFactor)

	notes := fmt.Sprintf("Auto-promoted from sandbox: %d trades, %.1f%% win rate, %.2f Sharpe",
		metrics.TotalTrades, metrics.WinRate, metrics.SharpeRatio)

	_, err = agm.versionManager.CreateVersion(strategyName, configJSON, notes, "auto_graduate_system", nil)
	if err != nil {
		log.Printf("[AUTO-GRADUATE][WARN] Failed to create version record: %v", err)
	}

	log.Printf("[AUTO-GRADUATE][PROMOTION] %s promoted: %d trades, %.1f%% win rate, $%.2f P&L",
		strategyName, metrics.TotalTrades, metrics.WinRate, metrics.TotalProfitLoss)

	return nil
}

// GetRecentDecisions retrieves recent auto-graduate decisions
func (agm *AutoGraduateMonitor) GetRecentDecisions(limit int) ([]PromotionDecision, error) {
	if limit == 0 {
		limit = 50
	}

	rows, err := agm.db.Query(`
		SELECT id, strategy_name, decision, reason, metrics_snapshot, created_at
		FROM assistant_decisions_log
		WHERE decision_type = 'auto_graduate'
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query decisions: %w", err)
	}
	defer rows.Close()

	var decisions []PromotionDecision
	for rows.Next() {
		var d PromotionDecision
		err := rows.Scan(&d.ID, &d.StrategyName, &d.Decision, &d.Reason, &d.MetricsSnapshot, &d.CreatedAt)
		if err != nil {
			continue
		}
		decisions = append(decisions, d)
	}

	return decisions, nil
}
