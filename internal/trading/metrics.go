package trading

import (
	"fmt"
	"math"
	"time"
)

// PerformanceMetrics tracks trading performance
type PerformanceMetrics struct {
	TotalTrades       int       `json:"total_trades"`
	WinningTrades     int       `json:"winning_trades"`
	LosingTrades      int       `json:"losing_trades"`
	WinRate           float64   `json:"win_rate"`
	TotalProfitLoss   float64   `json:"total_profit_loss"`
	AverageProfitLoss float64   `json:"average_profit_loss"`
	BestTrade         float64   `json:"best_trade"`
	WorstTrade        float64   `json:"worst_trade"`
	MaxDrawdown       float64   `json:"max_drawdown"`
	SharpeRatio       float64   `json:"sharpe_ratio"`
	StartBalance      float64   `json:"start_balance"`
	CurrentBalance    float64   `json:"current_balance"`
	ReturnOnInvestment float64  `json:"return_on_investment"`
	LastUpdated       time.Time `json:"last_updated"`
}

// MetricsCalculator computes trading performance metrics
type MetricsCalculator struct {
	Trades        []VirtualTrade
	InitialBalance float64
}

// NewMetricsCalculator creates a new metrics calculator
func NewMetricsCalculator(trades []VirtualTrade, initialBalance float64) *MetricsCalculator {
	return &MetricsCalculator{
		Trades:        trades,
		InitialBalance: initialBalance,
	}
}

// Calculate computes all performance metrics
func (mc *MetricsCalculator) Calculate() *PerformanceMetrics {
	if len(mc.Trades) == 0 {
		return &PerformanceMetrics{
			TotalTrades:    0,
			WinRate:        0,
			StartBalance:   mc.InitialBalance,
			CurrentBalance: mc.InitialBalance,
			LastUpdated:    time.Now(),
		}
	}

	metrics := &PerformanceMetrics{
		TotalTrades:    len(mc.Trades),
		StartBalance:   mc.InitialBalance,
		LastUpdated:    time.Now(),
	}

	// Count wins/losses and calculate P&L
	var totalPL float64
	var bestTrade float64 = math.Inf(-1)
	var worstTrade float64 = math.Inf(1)
	returns := make([]float64, 0)

	for _, trade := range mc.Trades {
		if trade.Status == "closed" {
			totalPL += trade.ProfitLoss

			// Count wins/losses
			if trade.ProfitLoss > 0 {
				metrics.WinningTrades++
			} else if trade.ProfitLoss < 0 {
				metrics.LosingTrades++
			}

			// Track best/worst
			if trade.ProfitLoss > bestTrade {
				bestTrade = trade.ProfitLoss
			}
			if trade.ProfitLoss < worstTrade {
				worstTrade = trade.ProfitLoss
			}

			// Store return for Sharpe ratio
			returns = append(returns, trade.ProfitLossPct)
		}
	}

	// Calculate win rate
	closedTrades := metrics.WinningTrades + metrics.LosingTrades
	if closedTrades > 0 {
		metrics.WinRate = (float64(metrics.WinningTrades) / float64(closedTrades)) * 100
	}

	// Calculate average P&L
	if closedTrades > 0 {
		metrics.AverageProfitLoss = totalPL / float64(closedTrades)
	}

	metrics.TotalProfitLoss = totalPL
	metrics.BestTrade = bestTrade
	metrics.WorstTrade = worstTrade
	metrics.CurrentBalance = mc.InitialBalance + totalPL

	// Calculate ROI
	if mc.InitialBalance > 0 {
		metrics.ReturnOnInvestment = (totalPL / mc.InitialBalance) * 100
	}

	// Calculate max drawdown
	metrics.MaxDrawdown = mc.calculateMaxDrawdown()

	// Calculate Sharpe ratio
	if len(returns) > 1 {
		metrics.SharpeRatio = mc.calculateSharpeRatio(returns)
	}

	return metrics
}

// calculateMaxDrawdown finds the maximum peak-to-trough decline
func (mc *MetricsCalculator) calculateMaxDrawdown() float64 {
	if len(mc.Trades) == 0 {
		return 0
	}

	balance := mc.InitialBalance
	peak := balance
	maxDrawdown := 0.0

	for _, trade := range mc.Trades {
		if trade.Status == "closed" {
			balance += trade.ProfitLoss

			// Update peak
			if balance > peak {
				peak = balance
			}

			// Calculate drawdown
			drawdown := ((peak - balance) / peak) * 100
			if drawdown > maxDrawdown {
				maxDrawdown = drawdown
			}
		}
	}

	return maxDrawdown
}

// calculateSharpeRatio computes risk-adjusted return (annualized)
// Formula: (Mean Return - Risk-Free Rate) / Std Dev of Returns
// Annualized using sqrt(252) for daily returns (trading days per year)
func (mc *MetricsCalculator) calculateSharpeRatio(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Calculate mean return
	var sum float64
	for _, r := range returns {
		sum += r
	}
	meanReturn := sum / float64(len(returns))

	// Calculate standard deviation (sample std dev)
	var variance float64
	for _, r := range returns {
		variance += math.Pow(r-meanReturn, 2)
	}
	variance /= float64(len(returns) - 1) // Sample variance (n-1)
	stdDev := math.Sqrt(variance)

	// Prevent division by zero
	if stdDev == 0 {
		// If no volatility, check if returns are positive
		if meanReturn > 0 {
			return 10.0 // Cap at high value for perfect consistency
		}
		return 0
	}

	// Risk-free rate (assume 4% annual = 0.04/252 per trading day ≈ 0.000159)
	const riskFreeRateDaily = 0.04 / 252.0
	
	// Sharpe ratio: (Return - Risk-Free Rate) / Volatility
	sharpeDaily := (meanReturn - riskFreeRateDaily) / stdDev
	
	// Annualize by multiplying by sqrt(252) trading days
	sharpeAnnualized := sharpeDaily * math.Sqrt(252)
	
	return sharpeAnnualized
}

// CalculateWinRate computes win rate percentage
func CalculateWinRate(trades []VirtualTrade) float64 {
	if len(trades) == 0 {
		return 0
	}

	wins := 0
	total := 0

	for _, trade := range trades {
		if trade.Status == "closed" {
			total++
			if trade.ProfitLoss > 0 {
				wins++
			}
		}
	}

	if total == 0 {
		return 0
	}

	return (float64(wins) / float64(total)) * 100
}

// GetRecentPerformance returns metrics for last N trades
func (mc *MetricsCalculator) GetRecentPerformance(lastN int) *PerformanceMetrics {
	if len(mc.Trades) == 0 {
		return mc.Calculate()
	}

	// Get last N trades
	start := len(mc.Trades) - lastN
	if start < 0 {
		start = 0
	}
	recentTrades := mc.Trades[start:]

	// Create temporary calculator for recent trades
	tempCalc := NewMetricsCalculator(recentTrades, mc.InitialBalance)
	return tempCalc.Calculate()
}

// TradeBreakdown provides detailed trade analysis
type TradeBreakdown struct {
	ByStrategy map[string]*PerformanceMetrics `json:"by_strategy"`
	BySymbol   map[string]*PerformanceMetrics `json:"by_symbol"`
	BySide     map[string]*PerformanceMetrics `json:"by_side"`
}

// GetTradeBreakdown analyzes trades by different dimensions
func (mc *MetricsCalculator) GetTradeBreakdown() *TradeBreakdown {
	breakdown := &TradeBreakdown{
		ByStrategy: make(map[string]*PerformanceMetrics),
		BySymbol:   make(map[string]*PerformanceMetrics),
		BySide:     make(map[string]*PerformanceMetrics),
	}

	// Group trades by strategy
	strategyTrades := make(map[string][]VirtualTrade)
	symbolTrades := make(map[string][]VirtualTrade)
	sideTrades := make(map[string][]VirtualTrade)

	for _, trade := range mc.Trades {
		// By strategy
		if _, exists := strategyTrades[trade.Strategy]; !exists {
			strategyTrades[trade.Strategy] = make([]VirtualTrade, 0)
		}
		strategyTrades[trade.Strategy] = append(strategyTrades[trade.Strategy], trade)

		// By symbol
		if _, exists := symbolTrades[trade.Symbol]; !exists {
			symbolTrades[trade.Symbol] = make([]VirtualTrade, 0)
		}
		symbolTrades[trade.Symbol] = append(symbolTrades[trade.Symbol], trade)

		// By side
		if _, exists := sideTrades[trade.Side]; !exists {
			sideTrades[trade.Side] = make([]VirtualTrade, 0)
		}
		sideTrades[trade.Side] = append(sideTrades[trade.Side], trade)
	}

	// Calculate metrics for each group
	for strategy, trades := range strategyTrades {
		calc := NewMetricsCalculator(trades, mc.InitialBalance)
		breakdown.ByStrategy[strategy] = calc.Calculate()
	}

	for symbol, trades := range symbolTrades {
		calc := NewMetricsCalculator(trades, mc.InitialBalance)
		breakdown.BySymbol[symbol] = calc.Calculate()
	}

	for side, trades := range sideTrades {
		calc := NewMetricsCalculator(trades, mc.InitialBalance)
		breakdown.BySide[side] = calc.Calculate()
	}

	return breakdown
}

// FormatMetrics returns human-readable metrics string
func FormatMetrics(metrics *PerformanceMetrics) string {
	return fmt.Sprintf(`
Trading Performance Metrics:
============================
Total Trades:    %d
Win Rate:        %.2f%% (%d wins / %d losses)
Total P&L:       $%.2f
Average P&L:     $%.2f
Best Trade:      $%.2f
Worst Trade:     $%.2f
Max Drawdown:    %.2f%%
Sharpe Ratio:    %.2f
ROI:             %.2f%%
Start Balance:   $%.2f
Current Balance: $%.2f
`,
		metrics.TotalTrades,
		metrics.WinRate, metrics.WinningTrades, metrics.LosingTrades,
		metrics.TotalProfitLoss,
		metrics.AverageProfitLoss,
		metrics.BestTrade,
		metrics.WorstTrade,
		metrics.MaxDrawdown,
		metrics.SharpeRatio,
		metrics.ReturnOnInvestment,
		metrics.StartBalance,
		metrics.CurrentBalance,
	)
}
