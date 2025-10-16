package trading

import (
	"fmt"
	"time"
)

// AuthorizationGate enforces trading requirements before allowing live trading
type AuthorizationGate struct {
	MinimumTrades       int     `json:"minimum_trades"`
	MinimumWinRate      float64 `json:"minimum_win_rate"`
	MinimumROI          float64 `json:"minimum_roi,omitempty"`
	MaxDrawdownLimit    float64 `json:"max_drawdown_limit,omitempty"`
	MinimumSharpeRatio  float64 `json:"minimum_sharpe_ratio,omitempty"`
}

// AuthorizationResult represents the result of an authorization check
type AuthorizationResult struct {
	Authorized      bool                   `json:"authorized"`
	Reason          string                 `json:"reason"`
	MissingCriteria []string               `json:"missing_criteria"`
	CurrentMetrics  *PerformanceMetrics    `json:"current_metrics"`
	Requirements    *AuthorizationGate     `json:"requirements"`
	CheckedAt       time.Time              `json:"checked_at"`
	Progress        *AuthorizationProgress `json:"progress"`
}

// AuthorizationProgress tracks progress toward requirements
type AuthorizationProgress struct {
	TradesProgress   float64 `json:"trades_progress"`    // percentage
	WinRateProgress  float64 `json:"win_rate_progress"`  // percentage
	ROIProgress      float64 `json:"roi_progress,omitempty"`
	DrawdownProgress float64 `json:"drawdown_progress,omitempty"`
}

// DefaultAuthorizationGate returns the standard authorization requirements
func DefaultAuthorizationGate() *AuthorizationGate {
	return &AuthorizationGate{
		MinimumTrades:      100,
		MinimumWinRate:     60.0,  // 60% win rate required
		MinimumROI:         10.0,  // Optional: 10% ROI
		MaxDrawdownLimit:   25.0,  // Optional: Max 25% drawdown
		MinimumSharpeRatio: 0.5,   // Optional: Positive risk-adjusted returns
	}
}

// StrictAuthorizationGate returns stricter requirements
func StrictAuthorizationGate() *AuthorizationGate {
	return &AuthorizationGate{
		MinimumTrades:      200,
		MinimumWinRate:     65.0,
		MinimumROI:         20.0,
		MaxDrawdownLimit:   20.0,
		MinimumSharpeRatio: 1.0,
	}
}

// RelaxedAuthorizationGate returns relaxed requirements for testing
func RelaxedAuthorizationGate() *AuthorizationGate {
	return &AuthorizationGate{
		MinimumTrades:  50,
		MinimumWinRate: 55.0,
	}
}

// CheckAuthorization verifies if trader meets live trading requirements
func (ag *AuthorizationGate) CheckAuthorization(metrics *PerformanceMetrics) *AuthorizationResult {
	result := &AuthorizationResult{
		Authorized:      true,
		MissingCriteria: make([]string, 0),
		CurrentMetrics:  metrics,
		Requirements:    ag,
		CheckedAt:       time.Now(),
		Progress:        &AuthorizationProgress{},
	}

	closedTrades := metrics.WinningTrades + metrics.LosingTrades

	// Check 1: Minimum trades requirement
	if closedTrades < ag.MinimumTrades {
		result.Authorized = false
		result.MissingCriteria = append(result.MissingCriteria,
			fmt.Sprintf("Insufficient trades: %d / %d completed", closedTrades, ag.MinimumTrades))
	}
	result.Progress.TradesProgress = (float64(closedTrades) / float64(ag.MinimumTrades)) * 100

	// Check 2: Win rate requirement
	if metrics.WinRate < ag.MinimumWinRate {
		result.Authorized = false
		result.MissingCriteria = append(result.MissingCriteria,
			fmt.Sprintf("Win rate below threshold: %.2f%% / %.2f%%", metrics.WinRate, ag.MinimumWinRate))
	}
	result.Progress.WinRateProgress = (metrics.WinRate / ag.MinimumWinRate) * 100

	// Check 3: ROI requirement (if set)
	if ag.MinimumROI > 0 {
		if metrics.ReturnOnInvestment < ag.MinimumROI {
			result.Authorized = false
			result.MissingCriteria = append(result.MissingCriteria,
				fmt.Sprintf("ROI below threshold: %.2f%% / %.2f%%", metrics.ReturnOnInvestment, ag.MinimumROI))
		}
		result.Progress.ROIProgress = (metrics.ReturnOnInvestment / ag.MinimumROI) * 100
	}

	// Check 4: Max drawdown requirement (if set)
	if ag.MaxDrawdownLimit > 0 {
		if metrics.MaxDrawdown > ag.MaxDrawdownLimit {
			result.Authorized = false
			result.MissingCriteria = append(result.MissingCriteria,
				fmt.Sprintf("Drawdown exceeds limit: %.2f%% / %.2f%%", metrics.MaxDrawdown, ag.MaxDrawdownLimit))
		}
		result.Progress.DrawdownProgress = ((ag.MaxDrawdownLimit - metrics.MaxDrawdown) / ag.MaxDrawdownLimit) * 100
	}

	// Check 5: Sharpe ratio requirement (if set)
	if ag.MinimumSharpeRatio > 0 {
		if metrics.SharpeRatio < ag.MinimumSharpeRatio {
			result.Authorized = false
			result.MissingCriteria = append(result.MissingCriteria,
				fmt.Sprintf("Sharpe ratio below threshold: %.2f / %.2f", metrics.SharpeRatio, ag.MinimumSharpeRatio))
		}
	}

	// Set reason
	if result.Authorized {
		result.Reason = "All requirements met - authorized for live trading"
	} else {
		result.Reason = fmt.Sprintf("Authorization denied: %d criteria not met", len(result.MissingCriteria))
	}

	return result
}

// IsLiveTradingAuthorized checks if user is authorized for live trading
func IsLiveTradingAuthorized(trades []VirtualTrade, initialBalance float64) bool {
	gate := DefaultAuthorizationGate()
	calculator := NewMetricsCalculator(trades, initialBalance)
	metrics := calculator.Calculate()
	result := gate.CheckAuthorization(metrics)
	return result.Authorized
}

// GetAuthorizationStatus returns detailed authorization status
func GetAuthorizationStatus(trades []VirtualTrade, initialBalance float64) *AuthorizationResult {
	gate := DefaultAuthorizationGate()
	calculator := NewMetricsCalculator(trades, initialBalance)
	metrics := calculator.Calculate()
	return gate.CheckAuthorization(metrics)
}

// GetAuthorizationProgress returns progress toward live trading
func GetAuthorizationProgress(trades []VirtualTrade, initialBalance float64) string {
	result := GetAuthorizationStatus(trades, initialBalance)
	
	closedTrades := result.CurrentMetrics.WinningTrades + result.CurrentMetrics.LosingTrades
	
	progress := fmt.Sprintf(`
Authorization Progress:
======================
Status: %s

Requirements:
  Trades:    %d / %d (%.1f%%)
  Win Rate:  %.2f%% / %.2f%% (%.1f%%)
`,
		result.Reason,
		closedTrades, result.Requirements.MinimumTrades, result.Progress.TradesProgress,
		result.CurrentMetrics.WinRate, result.Requirements.MinimumWinRate, result.Progress.WinRateProgress,
	)

	if result.Requirements.MinimumROI > 0 {
		progress += fmt.Sprintf("  ROI:       %.2f%% / %.2f%% (%.1f%%)\n",
			result.CurrentMetrics.ReturnOnInvestment, result.Requirements.MinimumROI, result.Progress.ROIProgress)
	}

	if len(result.MissingCriteria) > 0 {
		progress += "\nMissing Criteria:\n"
		for i, criteria := range result.MissingCriteria {
			progress += fmt.Sprintf("  %d. %s\n", i+1, criteria)
		}
	}

	return progress
}

// BenchmarkGate represents specific benchmark requirements
type BenchmarkGate struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	MinTrades         int     `json:"min_trades"`
	MinWinRate        float64 `json:"min_win_rate"`
	MinROI            float64 `json:"min_roi"`
	RewardMultiplier  float64 `json:"reward_multiplier"`  // Bonus for exceeding
}

// PredefinedBenchmarks returns standard trading benchmarks
func PredefinedBenchmarks() []BenchmarkGate {
	return []BenchmarkGate{
		{
			Name:             "Bronze Trader",
			Description:      "Entry-level sandbox trader",
			MinTrades:        50,
			MinWinRate:       55.0,
			MinROI:           5.0,
			RewardMultiplier: 1.0,
		},
		{
			Name:             "Silver Trader",
			Description:      "Competent sandbox trader",
			MinTrades:        100,
			MinWinRate:       60.0,
			MinROI:           10.0,
			RewardMultiplier: 1.5,
		},
		{
			Name:             "Gold Trader",
			Description:      "Advanced sandbox trader - Live trading eligible",
			MinTrades:        200,
			MinWinRate:       65.0,
			MinROI:           20.0,
			RewardMultiplier: 2.0,
		},
		{
			Name:             "Platinum Trader",
			Description:      "Expert trader with exceptional performance",
			MinTrades:        500,
			MinWinRate:       70.0,
			MinROI:           50.0,
			RewardMultiplier: 3.0,
		},
	}
}

// GetHighestBenchmark returns the highest benchmark achieved
func GetHighestBenchmark(metrics *PerformanceMetrics) *BenchmarkGate {
	benchmarks := PredefinedBenchmarks()
	closedTrades := metrics.WinningTrades + metrics.LosingTrades
	
	var highest *BenchmarkGate
	
	for i := range benchmarks {
		bench := &benchmarks[i]
		if closedTrades >= bench.MinTrades &&
			metrics.WinRate >= bench.MinWinRate &&
			metrics.ReturnOnInvestment >= bench.MinROI {
			highest = bench
		}
	}
	
	return highest
}
