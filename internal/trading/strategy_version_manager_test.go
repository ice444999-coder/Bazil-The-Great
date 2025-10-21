/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package trading

import (
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create temporary database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS strategy_versions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			strategy_name TEXT NOT NULL,
			version INTEGER NOT NULL,
			config_json TEXT NOT NULL,
			code_hash TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT 'system',
			notes TEXT,
			backtest_result_id INTEGER,
			is_active BOOLEAN DEFAULT 0,
			UNIQUE(strategy_name, version)
		)`,
		`CREATE INDEX idx_strategy_versions_name ON strategy_versions(strategy_name)`,
		`CREATE INDEX idx_strategy_versions_active ON strategy_versions(strategy_name, is_active)`,
		`CREATE TABLE IF NOT EXISTS backtest_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			strategy_name TEXT NOT NULL,
			version INTEGER,
			symbol TEXT NOT NULL,
			timeframe TEXT NOT NULL,
			start_date DATETIME NOT NULL,
			end_date DATETIME NOT NULL,
			total_candles INTEGER NOT NULL,
			total_trades INTEGER DEFAULT 0,
			winning_trades INTEGER DEFAULT 0,
			losing_trades INTEGER DEFAULT 0,
			win_rate REAL DEFAULT 0.0,
			total_pnl REAL DEFAULT 0.0,
			return_pct REAL DEFAULT 0.0,
			sharpe_ratio REAL DEFAULT 0.0,
			max_drawdown REAL DEFAULT 0.0,
			profit_factor REAL DEFAULT 0.0,
			avg_win REAL DEFAULT 0.0,
			avg_loss REAL DEFAULT 0.0,
			largest_win REAL DEFAULT 0.0,
			largest_loss REAL DEFAULT 0.0,
			avg_trade_duration_minutes INTEGER DEFAULT 0,
			passes_criteria BOOLEAN DEFAULT 0,
			execution_time_ms INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS strategy_rollback_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			strategy_name TEXT NOT NULL,
			from_version INTEGER NOT NULL,
			to_version INTEGER NOT NULL,
			reason TEXT,
			triggered_by TEXT DEFAULT 'manual',
			performance_before TEXT,
			performance_after TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT 'system'
		)`,
		`CREATE TABLE IF NOT EXISTS trades (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			strategy_name TEXT,
			pnl REAL,
			closed_at DATETIME
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			t.Fatalf("Failed to run migration: %v", err)
		}
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestStrategyVersionManager_CreateVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svm := NewStrategyVersionManager(db)

	config := map[string]interface{}{
		"rsi_period":         14,
		"oversold_threshold": 30,
		"position_size_pct":  2.0,
		"stop_loss_pct":      2.0,
		"target_profit_pct":  3.0,
	}
	configJSON, _ := json.Marshal(config)

	// Create first version
	v1, err := svm.CreateVersion("RSI_Oversold", string(configJSON), "Initial version", "test_user", nil)
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}

	if v1.Version != 1 {
		t.Errorf("Expected version 1, got %d", v1.Version)
	}
	if !v1.IsActive {
		t.Error("Expected version to be active")
	}

	// Create second version
	config["rsi_period"] = 21
	configJSON, _ = json.Marshal(config)
	v2, err := svm.CreateVersion("RSI_Oversold", string(configJSON), "Optimized RSI period", "test_user", nil)
	if err != nil {
		t.Fatalf("Failed to create second version: %v", err)
	}

	if v2.Version != 2 {
		t.Errorf("Expected version 2, got %d", v2.Version)
	}
	if !v2.IsActive {
		t.Error("Expected version 2 to be active")
	}

	// Check that v1 is now inactive
	v1Check, err := svm.GetVersion("RSI_Oversold", 1)
	if err != nil {
		t.Fatalf("Failed to get version 1: %v", err)
	}
	if v1Check.IsActive {
		t.Error("Expected version 1 to be inactive after creating version 2")
	}
}

func TestStrategyVersionManager_GetActiveVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svm := NewStrategyVersionManager(db)

	// No active version yet
	_, err := svm.GetActiveVersion("RSI_Oversold")
	if err == nil {
		t.Error("Expected error when no active version exists")
	}

	// Create a version
	config := `{"rsi_period":14}`
	_, err = svm.CreateVersion("RSI_Oversold", config, "Test", "test_user", nil)
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}

	// Get active version
	active, err := svm.GetActiveVersion("RSI_Oversold")
	if err != nil {
		t.Fatalf("Failed to get active version: %v", err)
	}

	if active.Version != 1 {
		t.Errorf("Expected version 1, got %d", active.Version)
	}
	if !active.IsActive {
		t.Error("Expected version to be active")
	}
}

func TestStrategyVersionManager_ListVersions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svm := NewStrategyVersionManager(db)

	// Empty list initially
	versions, err := svm.ListVersions("RSI_Oversold")
	if err != nil {
		t.Fatalf("Failed to list versions: %v", err)
	}
	if len(versions) != 0 {
		t.Errorf("Expected 0 versions, got %d", len(versions))
	}

	// Create 3 versions
	for i := 1; i <= 3; i++ {
		config := `{"version":` + string(rune(i+'0')) + `}`
		_, err := svm.CreateVersion("RSI_Oversold", config, "Version "+string(rune(i+'0')), "test_user", nil)
		if err != nil {
			t.Fatalf("Failed to create version %d: %v", i, err)
		}
	}

	// List all versions
	versions, err = svm.ListVersions("RSI_Oversold")
	if err != nil {
		t.Fatalf("Failed to list versions: %v", err)
	}
	if len(versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(versions))
	}

	// Versions should be in descending order
	if versions[0].Version != 3 {
		t.Errorf("Expected first version to be 3, got %d", versions[0].Version)
	}
	if versions[2].Version != 1 {
		t.Errorf("Expected last version to be 1, got %d", versions[2].Version)
	}

	// Only latest should be active
	activeCount := 0
	for _, v := range versions {
		if v.IsActive {
			activeCount++
		}
	}
	if activeCount != 1 {
		t.Errorf("Expected 1 active version, got %d", activeCount)
	}
}

func TestStrategyVersionManager_RollbackToVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svm := NewStrategyVersionManager(db)

	// Create 3 versions
	for i := 1; i <= 3; i++ {
		config := `{"version":` + string(rune(i+'0')) + `}`
		_, err := svm.CreateVersion("RSI_Oversold", config, "Version "+string(rune(i+'0')), "test_user", nil)
		if err != nil {
			t.Fatalf("Failed to create version %d: %v", i, err)
		}
	}

	// Active should be version 3
	active, _ := svm.GetActiveVersion("RSI_Oversold")
	if active.Version != 3 {
		t.Errorf("Expected active version 3, got %d", active.Version)
	}

	// Rollback to version 1
	err := svm.RollbackToVersion("RSI_Oversold", 1, "Performance degradation", "manual", "test_user")
	if err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Active should now be version 1
	active, _ = svm.GetActiveVersion("RSI_Oversold")
	if active.Version != 1 {
		t.Errorf("Expected active version 1 after rollback, got %d", active.Version)
	}

	// Check rollback history
	history, err := svm.GetRollbackHistory("RSI_Oversold", 10)
	if err != nil {
		t.Fatalf("Failed to get rollback history: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("Expected 1 rollback entry, got %d", len(history))
	}
	if history[0].FromVersion != 3 || history[0].ToVersion != 1 {
		t.Errorf("Expected rollback from 3 to 1, got from %d to %d", history[0].FromVersion, history[0].ToVersion)
	}
	if history[0].Reason != "Performance degradation" {
		t.Errorf("Expected reason 'Performance degradation', got '%s'", history[0].Reason)
	}

	// Try to rollback to current version (should fail)
	err = svm.RollbackToVersion("RSI_Oversold", 1, "Test", "manual", "test_user")
	if err == nil {
		t.Error("Expected error when rolling back to current version")
	}

	// Try to rollback to non-existent version (should fail)
	err = svm.RollbackToVersion("RSI_Oversold", 99, "Test", "manual", "test_user")
	if err == nil {
		t.Error("Expected error when rolling back to non-existent version")
	}
}

func TestStrategyVersionManager_SaveBacktestResult(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svm := NewStrategyVersionManager(db)

	// Create a backtest result
	result := &BacktestResult{
		StrategyName:    "RSI_Oversold",
		Symbol:          "BTC/USDT",
		StartDate:       time.Now().Add(-24 * time.Hour),
		EndDate:         time.Now(),
		TotalCandles:    500,
		TotalTrades:     25,
		WinningTrades:   18,
		LosingTrades:    7,
		WinRate:         72.0,
		TotalProfitLoss: 523.45,
		ReturnPercent:   5.23,
		SharpeRatio:     1.87,
		MaxDrawdown:     -8.3,
		ProfitFactor:    2.45,
		AverageWinSize:  42.15,
		AverageLossSize: -18.90,
		LargestWin:      125.50,
		LargestLoss:     -45.20,
		Pass:            true,
		ExecutionTime:   350 * time.Millisecond,
	}

	version := 1
	savedResult, err := svm.SaveBacktestResult("RSI_Oversold", &version, result, "BTC/USDT", "1h", result.StartDate, result.EndDate)
	if err != nil {
		t.Fatalf("Failed to save backtest result: %v", err)
	}

	if savedResult.ID == 0 {
		t.Error("Expected non-zero ID for saved result")
	}
	if savedResult.TotalTrades != 25 {
		t.Errorf("Expected 25 trades, got %d", savedResult.TotalTrades)
	}
	if savedResult.WinRate != 72.0 {
		t.Errorf("Expected 72%% win rate, got %.2f%%", savedResult.WinRate)
	}
	if !savedResult.PassesCriteria {
		t.Error("Expected passes_criteria to be true")
	}

	// Retrieve the result
	retrieved, err := svm.GetBacktestResult(savedResult.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve backtest result: %v", err)
	}

	if retrieved.TotalTrades != savedResult.TotalTrades {
		t.Errorf("Mismatch in total_trades: expected %d, got %d", savedResult.TotalTrades, retrieved.TotalTrades)
	}
	if retrieved.SharpeRatio != savedResult.SharpeRatio {
		t.Errorf("Mismatch in sharpe_ratio: expected %.2f, got %.2f", savedResult.SharpeRatio, retrieved.SharpeRatio)
	}
}

func TestStrategyVersionManager_ListBacktestResults(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svm := NewStrategyVersionManager(db)

	// Create 5 backtest results
	for i := 1; i <= 5; i++ {
		result := &BacktestResult{
			StrategyName:    "RSI_Oversold",
			Symbol:          "BTC/USDT",
			StartDate:       time.Now().Add(-24 * time.Hour),
			EndDate:         time.Now(),
			TotalCandles:    500,
			TotalTrades:     i * 10,
			WinningTrades:   i * 6,
			LosingTrades:    i * 4,
			WinRate:         60.0,
			TotalProfitLoss: float64(i * 100),
			ReturnPercent:   float64(i),
			SharpeRatio:     1.5,
			MaxDrawdown:     -5.0,
			ProfitFactor:    1.8,
			AverageWinSize:  50.0,
			AverageLossSize: -30.0,
			LargestWin:      100.0,
			LargestLoss:     -50.0,
			Pass:            i > 2, // Only last 3 pass
			ExecutionTime:   time.Duration(i*100) * time.Millisecond,
		}

		version := i
		_, err := svm.SaveBacktestResult("RSI_Oversold", &version, result, "BTC/USDT", "1h", result.StartDate, result.EndDate)
		if err != nil {
			t.Fatalf("Failed to save backtest result %d: %v", i, err)
		}
	}

	// List all results
	results, err := svm.ListBacktestResults("RSI_Oversold", 50)
	if err != nil {
		t.Fatalf("Failed to list backtest results: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	// Results should be in descending order by created_at (most recent first)
	// Check that passing criteria count is correct
	passingCount := 0
	for _, r := range results {
		if r.PassesCriteria {
			passingCount++
		}
	}
	if passingCount != 3 {
		t.Errorf("Expected 3 passing results, got %d", passingCount)
	}

	// List with limit
	results, err = svm.ListBacktestResults("RSI_Oversold", 2)
	if err != nil {
		t.Fatalf("Failed to list backtest results with limit: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results with limit=2, got %d", len(results))
	}
}

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()
	// Teardown
	os.Exit(code)
}
