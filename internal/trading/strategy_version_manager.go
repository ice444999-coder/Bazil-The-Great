/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package trading

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// StrategyVersion represents a versioned configuration of a strategy
type StrategyVersion struct {
	ID               int       `json:"id"`
	StrategyName     string    `json:"strategy_name"`
	Version          int       `json:"version"`
	ConfigJSON       string    `json:"config_json"`
	CodeHash         string    `json:"code_hash,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        string    `json:"created_by"`
	Notes            string    `json:"notes,omitempty"`
	BacktestResultID *int      `json:"backtest_result_id,omitempty"`
	IsActive         bool      `json:"is_active"`
}

// BacktestResultDB represents a stored backtest result
type BacktestResultDB struct {
	ID                      int       `json:"id"`
	StrategyName            string    `json:"strategy_name"`
	Version                 *int      `json:"version,omitempty"`
	Symbol                  string    `json:"symbol"`
	Timeframe               string    `json:"timeframe"`
	StartDate               time.Time `json:"start_date"`
	EndDate                 time.Time `json:"end_date"`
	TotalCandles            int       `json:"total_candles"`
	TotalTrades             int       `json:"total_trades"`
	WinningTrades           int       `json:"winning_trades"`
	LosingTrades            int       `json:"losing_trades"`
	WinRate                 float64   `json:"win_rate"`
	TotalPnL                float64   `json:"total_pnl"`
	ReturnPct               float64   `json:"return_pct"`
	SharpeRatio             float64   `json:"sharpe_ratio"`
	MaxDrawdown             float64   `json:"max_drawdown"`
	ProfitFactor            float64   `json:"profit_factor"`
	AvgWin                  float64   `json:"avg_win"`
	AvgLoss                 float64   `json:"avg_loss"`
	LargestWin              float64   `json:"largest_win"`
	LargestLoss             float64   `json:"largest_loss"`
	AvgTradeDurationMinutes int       `json:"avg_trade_duration_minutes"`
	PassesCriteria          bool      `json:"passes_criteria"`
	ExecutionTimeMs         int       `json:"execution_time_ms"`
	CreatedAt               time.Time `json:"created_at"`
}

// RollbackHistoryEntry represents a strategy version rollback event
type RollbackHistoryEntry struct {
	ID                int       `json:"id"`
	StrategyName      string    `json:"strategy_name"`
	FromVersion       int       `json:"from_version"`
	ToVersion         int       `json:"to_version"`
	Reason            string    `json:"reason"`
	TriggeredBy       string    `json:"triggered_by"` // 'manual', 'auto', 'emergency'
	PerformanceBefore string    `json:"performance_before,omitempty"`
	PerformanceAfter  string    `json:"performance_after,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         string    `json:"created_by"`
}

// StrategyVersionManager handles versioning, rollback, and comparison
type StrategyVersionManager struct {
	db *sql.DB
}

// NewStrategyVersionManager creates a new version manager
func NewStrategyVersionManager(db *sql.DB) *StrategyVersionManager {
	return &StrategyVersionManager{db: db}
}

// CreateVersion creates a new version of a strategy
func (svm *StrategyVersionManager) CreateVersion(strategyName, configJSON, notes, createdBy string, backtestResultID *int) (*StrategyVersion, error) {
	// Get next version number
	var maxVersion int
	err := svm.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM strategy_versions WHERE strategy_name = ?`, strategyName).Scan(&maxVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get max version: %w", err)
	}
	nextVersion := maxVersion + 1

	// Calculate code hash (for future use when storing code)
	codeHash := fmt.Sprintf("%x", sha256.Sum256([]byte(configJSON)))

	// Deactivate all previous versions
	_, err = svm.db.Exec(`UPDATE strategy_versions SET is_active = 0 WHERE strategy_name = ?`, strategyName)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate previous versions: %w", err)
	}

	// Insert new version
	result, err := svm.db.Exec(`
		INSERT INTO strategy_versions (strategy_name, version, config_json, code_hash, notes, created_by, backtest_result_id, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1)
	`, strategyName, nextVersion, configJSON, codeHash, notes, createdBy, backtestResultID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert version: %w", err)
	}

	id, _ := result.LastInsertId()

	return &StrategyVersion{
		ID:               int(id),
		StrategyName:     strategyName,
		Version:          nextVersion,
		ConfigJSON:       configJSON,
		CodeHash:         codeHash,
		CreatedAt:        time.Now(),
		CreatedBy:        createdBy,
		Notes:            notes,
		BacktestResultID: backtestResultID,
		IsActive:         true,
	}, nil
}

// GetActiveVersion returns the currently active version of a strategy
func (svm *StrategyVersionManager) GetActiveVersion(strategyName string) (*StrategyVersion, error) {
	var version StrategyVersion
	var backtestResultID sql.NullInt64

	err := svm.db.QueryRow(`
		SELECT id, strategy_name, version, config_json, code_hash, created_at, created_by, notes, backtest_result_id, is_active
		FROM strategy_versions
		WHERE strategy_name = ? AND is_active = 1
		ORDER BY version DESC
		LIMIT 1
	`, strategyName).Scan(
		&version.ID, &version.StrategyName, &version.Version, &version.ConfigJSON,
		&version.CodeHash, &version.CreatedAt, &version.CreatedBy, &version.Notes,
		&backtestResultID, &version.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no active version found for strategy %s", strategyName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}

	if backtestResultID.Valid {
		id := int(backtestResultID.Int64)
		version.BacktestResultID = &id
	}

	return &version, nil
}

// GetVersion returns a specific version of a strategy
func (svm *StrategyVersionManager) GetVersion(strategyName string, versionNum int) (*StrategyVersion, error) {
	var version StrategyVersion
	var backtestResultID sql.NullInt64

	err := svm.db.QueryRow(`
		SELECT id, strategy_name, version, config_json, code_hash, created_at, created_by, notes, backtest_result_id, is_active
		FROM strategy_versions
		WHERE strategy_name = ? AND version = ?
	`, strategyName, versionNum).Scan(
		&version.ID, &version.StrategyName, &version.Version, &version.ConfigJSON,
		&version.CodeHash, &version.CreatedAt, &version.CreatedBy, &version.Notes,
		&backtestResultID, &version.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("version %d not found for strategy %s", versionNum, strategyName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	if backtestResultID.Valid {
		id := int(backtestResultID.Int64)
		version.BacktestResultID = &id
	}

	return &version, nil
}

// ListVersions returns all versions of a strategy
func (svm *StrategyVersionManager) ListVersions(strategyName string) ([]StrategyVersion, error) {
	rows, err := svm.db.Query(`
		SELECT id, strategy_name, version, config_json, code_hash, created_at, created_by, notes, backtest_result_id, is_active
		FROM strategy_versions
		WHERE strategy_name = ?
		ORDER BY version DESC
	`, strategyName)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}
	defer rows.Close()

	var versions []StrategyVersion
	for rows.Next() {
		var version StrategyVersion
		var backtestResultID sql.NullInt64

		err := rows.Scan(
			&version.ID, &version.StrategyName, &version.Version, &version.ConfigJSON,
			&version.CodeHash, &version.CreatedAt, &version.CreatedBy, &version.Notes,
			&backtestResultID, &version.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}

		if backtestResultID.Valid {
			id := int(backtestResultID.Int64)
			version.BacktestResultID = &id
		}

		versions = append(versions, version)
	}

	return versions, nil
}

// RollbackToVersion rolls back a strategy to a specific version
func (svm *StrategyVersionManager) RollbackToVersion(strategyName string, targetVersion int, reason, triggeredBy, createdBy string) error {
	// Get current active version
	currentVersion, err := svm.GetActiveVersion(strategyName)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion.Version == targetVersion {
		return fmt.Errorf("already on version %d", targetVersion)
	}

	// Verify target version exists
	targetVer, err := svm.GetVersion(strategyName, targetVersion)
	if err != nil {
		return fmt.Errorf("target version not found: %w", err)
	}

	// Get current performance metrics (optional - for audit)
	currentMetrics, _ := svm.getStrategyMetrics(strategyName)
	currentMetricsJSON, _ := json.Marshal(currentMetrics)

	// Deactivate all versions
	_, err = svm.db.Exec(`UPDATE strategy_versions SET is_active = 0 WHERE strategy_name = ?`, strategyName)
	if err != nil {
		return fmt.Errorf("failed to deactivate versions: %w", err)
	}

	// Activate target version
	_, err = svm.db.Exec(`UPDATE strategy_versions SET is_active = 1 WHERE id = ?`, targetVer.ID)
	if err != nil {
		return fmt.Errorf("failed to activate target version: %w", err)
	}

	// Record rollback in history
	_, err = svm.db.Exec(`
		INSERT INTO strategy_rollback_history (strategy_name, from_version, to_version, reason, triggered_by, performance_before, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, strategyName, currentVersion.Version, targetVersion, reason, triggeredBy, string(currentMetricsJSON), createdBy)
	if err != nil {
		return fmt.Errorf("failed to record rollback: %w", err)
	}

	return nil
}

// SaveBacktestResult saves a backtest result to the database
func (svm *StrategyVersionManager) SaveBacktestResult(strategyName string, version *int, result *BacktestResult, symbol, timeframe string, startDate, endDate time.Time) (*BacktestResultDB, error) {
	execResult, err := svm.db.Exec(`
		INSERT INTO backtest_results (
			strategy_name, version, symbol, timeframe, start_date, end_date, total_candles,
			total_trades, winning_trades, losing_trades, win_rate, total_pnl, return_pct,
			sharpe_ratio, max_drawdown, profit_factor, avg_win, avg_loss, largest_win, largest_loss,
			avg_trade_duration_minutes, passes_criteria, execution_time_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, strategyName, version, symbol, timeframe, startDate, endDate, result.TotalCandles,
		result.TotalTrades, result.WinningTrades, result.LosingTrades, result.WinRate,
		result.TotalProfitLoss, result.ReturnPercent, result.SharpeRatio, result.MaxDrawdown,
		result.ProfitFactor, result.AverageWinSize, result.AverageLossSize, result.LargestWin, result.LargestLoss,
		0, result.Pass, result.ExecutionTime.Milliseconds(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save backtest result: %w", err)
	}

	id, _ := execResult.LastInsertId()

	return &BacktestResultDB{
		ID:                      int(id),
		StrategyName:            strategyName,
		Version:                 version,
		Symbol:                  symbol,
		Timeframe:               timeframe,
		StartDate:               startDate,
		EndDate:                 endDate,
		TotalCandles:            result.TotalCandles,
		TotalTrades:             result.TotalTrades,
		WinningTrades:           result.WinningTrades,
		LosingTrades:            result.LosingTrades,
		WinRate:                 result.WinRate,
		TotalPnL:                result.TotalProfitLoss,
		ReturnPct:               result.ReturnPercent,
		SharpeRatio:             result.SharpeRatio,
		MaxDrawdown:             result.MaxDrawdown,
		ProfitFactor:            result.ProfitFactor,
		AvgWin:                  result.AverageWinSize,
		AvgLoss:                 result.AverageLossSize,
		LargestWin:              result.LargestWin,
		LargestLoss:             result.LargestLoss,
		AvgTradeDurationMinutes: 0,
		PassesCriteria:          result.Pass,
		ExecutionTimeMs:         int(result.ExecutionTime.Milliseconds()),
		CreatedAt:               time.Now(),
	}, nil
}

// GetBacktestResult retrieves a backtest result by ID
func (svm *StrategyVersionManager) GetBacktestResult(id int) (*BacktestResultDB, error) {
	var result BacktestResultDB
	var version sql.NullInt64

	err := svm.db.QueryRow(`
		SELECT id, strategy_name, version, symbol, timeframe, start_date, end_date, total_candles,
			total_trades, winning_trades, losing_trades, win_rate, total_pnl, return_pct,
			sharpe_ratio, max_drawdown, profit_factor, avg_win, avg_loss, largest_win, largest_loss,
			avg_trade_duration_minutes, passes_criteria, execution_time_ms, created_at
		FROM backtest_results
		WHERE id = ?
	`, id).Scan(
		&result.ID, &result.StrategyName, &version, &result.Symbol, &result.Timeframe,
		&result.StartDate, &result.EndDate, &result.TotalCandles, &result.TotalTrades,
		&result.WinningTrades, &result.LosingTrades, &result.WinRate, &result.TotalPnL,
		&result.ReturnPct, &result.SharpeRatio, &result.MaxDrawdown, &result.ProfitFactor,
		&result.AvgWin, &result.AvgLoss, &result.LargestWin, &result.LargestLoss,
		&result.AvgTradeDurationMinutes, &result.PassesCriteria, &result.ExecutionTimeMs,
		&result.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("backtest result %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get backtest result: %w", err)
	}

	if version.Valid {
		v := int(version.Int64)
		result.Version = &v
	}

	return &result, nil
}

// ListBacktestResults lists backtest results for a strategy
func (svm *StrategyVersionManager) ListBacktestResults(strategyName string, limit int) ([]BacktestResultDB, error) {
	if limit == 0 {
		limit = 50
	}

	rows, err := svm.db.Query(`
		SELECT id, strategy_name, version, symbol, timeframe, start_date, end_date, total_candles,
			total_trades, winning_trades, losing_trades, win_rate, total_pnl, return_pct,
			sharpe_ratio, max_drawdown, profit_factor, avg_win, avg_loss, largest_win, largest_loss,
			avg_trade_duration_minutes, passes_criteria, execution_time_ms, created_at
		FROM backtest_results
		WHERE strategy_name = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, strategyName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list backtest results: %w", err)
	}
	defer rows.Close()

	var results []BacktestResultDB
	for rows.Next() {
		var result BacktestResultDB
		var version sql.NullInt64

		err := rows.Scan(
			&result.ID, &result.StrategyName, &version, &result.Symbol, &result.Timeframe,
			&result.StartDate, &result.EndDate, &result.TotalCandles, &result.TotalTrades,
			&result.WinningTrades, &result.LosingTrades, &result.WinRate, &result.TotalPnL,
			&result.ReturnPct, &result.SharpeRatio, &result.MaxDrawdown, &result.ProfitFactor,
			&result.AvgWin, &result.AvgLoss, &result.LargestWin, &result.LargestLoss,
			&result.AvgTradeDurationMinutes, &result.PassesCriteria, &result.ExecutionTimeMs,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backtest result: %w", err)
		}

		if version.Valid {
			v := int(version.Int64)
			result.Version = &v
		}

		results = append(results, result)
	}

	return results, nil
}

// GetRollbackHistory returns the rollback history for a strategy
func (svm *StrategyVersionManager) GetRollbackHistory(strategyName string, limit int) ([]RollbackHistoryEntry, error) {
	if limit == 0 {
		limit = 20
	}

	rows, err := svm.db.Query(`
		SELECT id, strategy_name, from_version, to_version, reason, triggered_by, performance_before, performance_after, created_at, created_by
		FROM strategy_rollback_history
		WHERE strategy_name = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, strategyName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollback history: %w", err)
	}
	defer rows.Close()

	var history []RollbackHistoryEntry
	for rows.Next() {
		var entry RollbackHistoryEntry
		err := rows.Scan(
			&entry.ID, &entry.StrategyName, &entry.FromVersion, &entry.ToVersion,
			&entry.Reason, &entry.TriggeredBy, &entry.PerformanceBefore,
			&entry.PerformanceAfter, &entry.CreatedAt, &entry.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollback entry: %w", err)
		}
		history = append(history, entry)
	}

	return history, nil
}

// getStrategyMetrics is a helper to get current strategy performance
func (svm *StrategyVersionManager) getStrategyMetrics(strategyName string) (map[string]interface{}, error) {
	var totalTrades, winningTrades int
	var totalPnL float64

	err := svm.db.QueryRow(`
		SELECT COUNT(*), 
			   SUM(CASE WHEN pnl > 0 THEN 1 ELSE 0 END),
			   SUM(pnl)
		FROM trades
		WHERE strategy_name = ? AND closed_at IS NOT NULL
	`, strategyName).Scan(&totalTrades, &winningTrades, &totalPnL)

	if err != nil {
		return nil, err
	}

	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winningTrades) / float64(totalTrades) * 100
	}

	return map[string]interface{}{
		"total_trades":   totalTrades,
		"winning_trades": winningTrades,
		"win_rate":       winRate,
		"total_pnl":      totalPnL,
	}, nil
}
