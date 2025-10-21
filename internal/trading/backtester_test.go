/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package trading

import (
	"ares_api/internal/eventbus"
	"fmt"
	"testing"
)

// TestBacktester_RSIStrategy tests backtesting with RSI strategy
func TestBacktester_RSIStrategy(t *testing.T) {
	// Generate synthetic historical data (500 candles)
	candles := GenerateSyntheticData("BTC/USDT", 500, 50000.0)

	// Create backtester with config
	config := BacktestConfig{
		StartingBalance: 10000.0,
		PositionSize:    2.0, // 2% per trade
		MaxDailyTrades:  10,
		StopLossEnabled: true,
		Slippage:        0.1,   // 0.1%
		Commission:      0.075, // 0.075% per side
	}

	backtester := NewBacktester(config)

	// Create RSI strategy
	eb := eventbus.NewEventBus()
	defer eb.Close()
	strategy := NewRSIOversoldStrategy(eb)

	// Run backtest
	result, err := backtester.RunBacktest(strategy, candles)
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	// Print results
	fmt.Printf("\n========== BACKTEST RESULTS ==========\n")
	fmt.Printf("Strategy: %s\n", result.StrategyName)
	fmt.Printf("Symbol: %s\n", result.Symbol)
	fmt.Printf("Period: %s to %s\n", result.StartDate.Format("2006-01-02"), result.EndDate.Format("2006-01-02"))
	fmt.Printf("Total Candles: %d\n", result.TotalCandles)
	fmt.Printf("\n--- TRADE STATISTICS ---\n")
	fmt.Printf("Total Trades: %d\n", result.TotalTrades)
	fmt.Printf("Winning Trades: %d\n", result.WinningTrades)
	fmt.Printf("Losing Trades: %d\n", result.LosingTrades)
	fmt.Printf("Win Rate: %.2f%%\n", result.WinRate)
	fmt.Printf("\n--- PERFORMANCE METRICS ---\n")
	fmt.Printf("Starting Balance: $%.2f\n", result.StartingBalance)
	fmt.Printf("Ending Balance: $%.2f\n", result.EndingBalance)
	fmt.Printf("Total P&L: $%.2f\n", result.TotalProfitLoss)
	fmt.Printf("Return: %.2f%%\n", result.ReturnPercent)
	fmt.Printf("Average P&L per Trade: $%.2f\n", result.AveragePnL)
	fmt.Printf("Expected Value: $%.2f\n", result.ExpectedValue)
	fmt.Printf("\n--- RISK METRICS ---\n")
	fmt.Printf("Max Drawdown: %.2f%%\n", result.MaxDrawdown)
	fmt.Printf("Sharpe Ratio: %.2f\n", result.SharpeRatio)
	fmt.Printf("Profit Factor: %.2f\n", result.ProfitFactor)
	fmt.Printf("Largest Win: $%.2f\n", result.LargestWin)
	fmt.Printf("Largest Loss: $%.2f\n", result.LargestLoss)
	fmt.Printf("Average Win: $%.2f\n", result.AverageWinSize)
	fmt.Printf("Average Loss: $%.2f\n", result.AverageLossSize)
	fmt.Printf("\n--- PROMOTION CRITERIA ---\n")
	fmt.Printf("Pass: %v\n", result.Pass)
	if !result.Pass {
		fmt.Printf("Failure Reasons:\n")
		for _, reason := range result.FailureReasons {
			fmt.Printf("  - %s\n", reason)
		}
	}
	fmt.Printf("\nExecution Time: %s\n", result.ExecutionTime)
	fmt.Printf("======================================\n\n")

	// Validate basic expectations
	if result.TotalTrades == 0 {
		t.Error("Expected some trades to be executed")
	}

	if result.EndingBalance == result.StartingBalance {
		t.Error("Expected balance to change")
	}

	fmt.Printf("✅ Backtest completed successfully\n")
}

// TestBacktester_AllStrategies tests all 5 strategies
func TestBacktester_AllStrategies(t *testing.T) {
	// Generate historical data
	candles := GenerateSyntheticData("BTC/USDT", 500, 50000.0)

	config := BacktestConfig{
		StartingBalance: 10000.0,
		PositionSize:    2.0,
		MaxDailyTrades:  10,
		StopLossEnabled: true,
	}

	backtester := NewBacktester(config)
	eb := eventbus.NewEventBus()
	defer eb.Close()

	// Test all strategies
	strategies := GetAllStrategies(eb)

	fmt.Printf("\n========== COMPARING ALL STRATEGIES ==========\n")

	bestStrategy := ""
	bestReturn := -999999.0

	for _, strategy := range strategies {
		result, err := backtester.RunBacktest(strategy, candles)
		if err != nil {
			t.Errorf("Backtest failed for %s: %v", strategy.Name(), err)
			continue
		}

		fmt.Printf("\n%s:\n", result.StrategyName)
		fmt.Printf("  Trades: %d | Win Rate: %.2f%% | Return: %.2f%% | Sharpe: %.2f | Pass: %v\n",
			result.TotalTrades, result.WinRate, result.ReturnPercent, result.SharpeRatio, result.Pass)

		if result.ReturnPercent > bestReturn {
			bestReturn = result.ReturnPercent
			bestStrategy = result.StrategyName
		}
	}

	fmt.Printf("\n🏆 Best Performer: %s (%.2f%% return)\n", bestStrategy, bestReturn)
	fmt.Printf("==============================================\n\n")
}

// TestBacktester_InsufficientData tests error handling
func TestBacktester_InsufficientData(t *testing.T) {
	candles := GenerateSyntheticData("BTC/USDT", 10, 50000.0) // Only 10 candles

	config := BacktestConfig{
		StartingBalance: 10000.0,
	}

	backtester := NewBacktester(config)
	eb := eventbus.NewEventBus()
	defer eb.Close()
	strategy := NewRSIOversoldStrategy(eb)

	_, err := backtester.RunBacktest(strategy, candles)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}

	fmt.Printf("✅ Correctly rejected insufficient data\n")
}

// TestBacktester_ConfigDefaults tests default config values
func TestBacktester_ConfigDefaults(t *testing.T) {
	config := BacktestConfig{} // Empty config
	backtester := NewBacktester(config)

	if backtester.config.StartingBalance != 10000.0 {
		t.Errorf("Expected default starting balance 10000, got %.2f", backtester.config.StartingBalance)
	}

	if backtester.config.PositionSize != 2.0 {
		t.Errorf("Expected default position size 2%%, got %.2f", backtester.config.PositionSize)
	}

	if backtester.config.Slippage != 0.1 {
		t.Errorf("Expected default slippage 0.1%%, got %.2f", backtester.config.Slippage)
	}

	if backtester.config.Commission != 0.075 {
		t.Errorf("Expected default commission 0.075%%, got %.2f", backtester.config.Commission)
	}

	fmt.Printf("✅ Default config values correct\n")
}

// BenchmarkBacktester benchmarks backtest performance
func BenchmarkBacktester(b *testing.B) {
	candles := GenerateSyntheticData("BTC/USDT", 500, 50000.0)
	config := BacktestConfig{StartingBalance: 10000.0}
	backtester := NewBacktester(config)
	eb := eventbus.NewEventBus()
	defer eb.Close()
	strategy := NewRSIOversoldStrategy(eb)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = backtester.RunBacktest(strategy, candles)
	}
}
