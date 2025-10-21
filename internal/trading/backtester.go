/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package trading

import (
	"fmt"
	"log"
	"math"
	"time"
)

// ========================================================================
// BACKTESTING ENGINE
// ========================================================================
// Tests strategies against historical data to validate logic before live deployment.
// Calculates performance metrics: win rate, Sharpe ratio, max drawdown, total P&L.
// Prevents deploying broken strategies that would lose real money.
// ========================================================================

// HistoricalCandle - OHLCV data point for backtesting
type HistoricalCandle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Symbol    string
}

// BacktestResult - Performance metrics from backtesting
type BacktestResult struct {
	StrategyName    string          `json:"strategy_name"`
	Symbol          string          `json:"symbol"`
	StartDate       time.Time       `json:"start_date"`
	EndDate         time.Time       `json:"end_date"`
	TotalCandles    int             `json:"total_candles"`
	TotalTrades     int             `json:"total_trades"`
	WinningTrades   int             `json:"winning_trades"`
	LosingTrades    int             `json:"losing_trades"`
	WinRate         float64         `json:"win_rate"`          // Percentage
	TotalProfitLoss float64         `json:"total_profit_loss"` // USD
	AveragePnL      float64         `json:"average_pnl"`       // Per trade
	MaxDrawdown     float64         `json:"max_drawdown"`      // Percentage
	SharpeRatio     float64         `json:"sharpe_ratio"`      // Risk-adjusted return
	StartingBalance float64         `json:"starting_balance"`
	EndingBalance   float64         `json:"ending_balance"`
	ReturnPercent   float64         `json:"return_percent"`
	LargestWin      float64         `json:"largest_win"`
	LargestLoss     float64         `json:"largest_loss"`
	AverageWinSize  float64         `json:"average_win_size"`
	AverageLossSize float64         `json:"average_loss_size"`
	ProfitFactor    float64         `json:"profit_factor"`  // Gross profit / Gross loss
	ExpectedValue   float64         `json:"expected_value"` // Average per trade
	Trades          []BacktestTrade `json:"trades"`
	ExecutionTime   time.Duration   `json:"execution_time"`
	Pass            bool            `json:"pass"` // Meets promotion criteria
	FailureReasons  []string        `json:"failure_reasons,omitempty"`
}

// BacktestTrade - Individual trade from backtest
type BacktestTrade struct {
	EntryTime   time.Time `json:"entry_time"`
	ExitTime    time.Time `json:"exit_time"`
	Action      string    `json:"action"` // "buy" or "sell"
	EntryPrice  float64   `json:"entry_price"`
	ExitPrice   float64   `json:"exit_price"`
	Size        float64   `json:"size"`        // Position size
	ProfitLoss  float64   `json:"profit_loss"` // USD
	PnLPercent  float64   `json:"pnl_percent"` // Percentage
	Confidence  float64   `json:"confidence"`  // Strategy confidence (0-100)
	Reasoning   string    `json:"reasoning"`
	StopLoss    float64   `json:"stop_loss"`
	TargetPrice float64   `json:"target_price"`
	Duration    string    `json:"duration"` // Human readable
}

// BacktestConfig - Configuration for backtesting
type BacktestConfig struct {
	StartingBalance float64 // Initial capital (default: $10,000)
	PositionSize    float64 // % of balance per trade (default: 2%)
	MaxDailyTrades  int     // Limit trades per day (default: unlimited)
	StopLossEnabled bool    // Use strategy stop loss (default: true)
	Slippage        float64 // Simulated slippage % (default: 0.1%)
	Commission      float64 // Trading fee % (default: 0.075% per side)
}

// Backtester - Main backtesting engine
type Backtester struct {
	config BacktestConfig
}

// NewBacktester creates a new backtesting engine
func NewBacktester(config BacktestConfig) *Backtester {
	// Set defaults
	if config.StartingBalance == 0 {
		config.StartingBalance = 10000.0
	}
	if config.PositionSize == 0 {
		config.PositionSize = 2.0 // 2% per trade
	}
	if config.Slippage == 0 {
		config.Slippage = 0.1 // 0.1%
	}
	if config.Commission == 0 {
		config.Commission = 0.075 // 0.075% per side (Binance)
	}

	return &Backtester{
		config: config,
	}
}

// RunBacktest executes strategy against historical data
func (bt *Backtester) RunBacktest(strategy Strategy, candles []HistoricalCandle) (*BacktestResult, error) {
	startTime := time.Now()

	if len(candles) < 50 {
		return nil, fmt.Errorf("insufficient historical data: need at least 50 candles, got %d", len(candles))
	}

	result := &BacktestResult{
		StrategyName:    strategy.Name(),
		Symbol:          candles[0].Symbol,
		StartDate:       candles[0].Timestamp,
		EndDate:         candles[len(candles)-1].Timestamp,
		TotalCandles:    len(candles),
		StartingBalance: bt.config.StartingBalance,
		Trades:          []BacktestTrade{},
	}

	balance := bt.config.StartingBalance
	var openPosition *BacktestTrade
	tradesToday := 0
	currentDay := candles[0].Timestamp.Day()
	peakBalance := balance
	maxDrawdown := 0.0

	// Build price history for strategy (as VirtualTrade slice)
	priceHistory := []VirtualTrade{}

	log.Printf("[BACKTEST] Starting backtest for %s on %s (%d candles)", strategy.Name(), result.Symbol, len(candles))

	// Iterate through each candle
	for i, candle := range candles {
		// Reset daily trade counter
		if candle.Timestamp.Day() != currentDay {
			tradesToday = 0
			currentDay = candle.Timestamp.Day()
		}

		// Add current candle to price history
		priceHistory = append(priceHistory, VirtualTrade{
			Price:  candle.Close,
			Amount: candle.Volume,
		})

		// Keep only recent history (last 200 candles for efficiency)
		if len(priceHistory) > 200 {
			priceHistory = priceHistory[len(priceHistory)-200:]
		}

		// Create mock market data for strategy
		mockData := &MockMarketData{
			Prices: map[string]float64{
				result.Symbol: candle.Close,
			},
		}

		// Execute strategy
		signal, err := strategy.Analyze(result.Symbol, mockData, priceHistory)
		if err != nil {
			log.Printf("[BACKTEST] Strategy error at candle %d: %v", i, err)
			continue
		}

		// Check for exit conditions on open position
		if openPosition != nil {
			exitTriggered := false
			exitPrice := candle.Close
			exitReason := ""

			// Stop loss check
			if bt.config.StopLossEnabled && openPosition.StopLoss > 0 {
				if openPosition.Action == "buy" && candle.Low <= openPosition.StopLoss {
					exitTriggered = true
					exitPrice = openPosition.StopLoss
					exitReason = "Stop loss hit"
				} else if openPosition.Action == "sell" && candle.High >= openPosition.StopLoss {
					exitTriggered = true
					exitPrice = openPosition.StopLoss
					exitReason = "Stop loss hit"
				}
			}

			// Target price check
			if openPosition.TargetPrice > 0 {
				if openPosition.Action == "buy" && candle.High >= openPosition.TargetPrice {
					exitTriggered = true
					exitPrice = openPosition.TargetPrice
					exitReason = "Target price reached"
				} else if openPosition.Action == "sell" && candle.Low <= openPosition.TargetPrice {
					exitTriggered = true
					exitPrice = openPosition.TargetPrice
					exitReason = "Target price reached"
				}
			}

			// Opposite signal (strategy reverses)
			if (openPosition.Action == "buy" && signal.Action == "sell") ||
				(openPosition.Action == "sell" && signal.Action == "buy") {
				exitTriggered = true
				exitPrice = candle.Close
				exitReason = "Strategy reversal signal"
			}

			// Close position if exit triggered
			if exitTriggered {
				openPosition.ExitTime = candle.Timestamp
				openPosition.ExitPrice = bt.applySlippage(exitPrice, openPosition.Action == "buy")
				openPosition.Duration = openPosition.ExitTime.Sub(openPosition.EntryTime).String()

				// Calculate P&L
				if openPosition.Action == "buy" {
					openPosition.ProfitLoss = (openPosition.ExitPrice - openPosition.EntryPrice) * openPosition.Size
				} else { // sell (short)
					openPosition.ProfitLoss = (openPosition.EntryPrice - openPosition.ExitPrice) * openPosition.Size
				}

				// Apply commission (both entry and exit)
				commission := (openPosition.EntryPrice * openPosition.Size * bt.config.Commission / 100) +
					(openPosition.ExitPrice * openPosition.Size * bt.config.Commission / 100)
				openPosition.ProfitLoss -= commission

				openPosition.PnLPercent = (openPosition.ProfitLoss / (openPosition.EntryPrice * openPosition.Size)) * 100
				openPosition.Reasoning += fmt.Sprintf(" | Exit: %s", exitReason)

				// Update balance
				balance += openPosition.ProfitLoss

				// Track stats
				result.Trades = append(result.Trades, *openPosition)
				result.TotalTrades++

				if openPosition.ProfitLoss > 0 {
					result.WinningTrades++
				} else {
					result.LosingTrades++
				}

				// Track drawdown
				if balance > peakBalance {
					peakBalance = balance
				}
				drawdown := ((peakBalance - balance) / peakBalance) * 100
				if drawdown > maxDrawdown {
					maxDrawdown = drawdown
				}

				log.Printf("[BACKTEST] Trade #%d: %s %s at %.2f, exit %.2f | P&L: $%.2f (%.2f%%) | Balance: $%.2f",
					result.TotalTrades, openPosition.Action, result.Symbol,
					openPosition.EntryPrice, openPosition.ExitPrice,
					openPosition.ProfitLoss, openPosition.PnLPercent, balance)

				openPosition = nil
			}
		}

		// Open new position if signal and no open position
		if openPosition == nil && (signal.Action == "buy" || signal.Action == "sell") {
			// Check daily trade limit
			if bt.config.MaxDailyTrades > 0 && tradesToday >= bt.config.MaxDailyTrades {
				continue
			}

			// Calculate position size
			positionValue := balance * (bt.config.PositionSize / 100)
			size := positionValue / candle.Close

			if size > 0 {
				entryPrice := bt.applySlippage(candle.Close, signal.Action == "sell")

				openPosition = &BacktestTrade{
					EntryTime:   candle.Timestamp,
					Action:      signal.Action,
					EntryPrice:  entryPrice,
					Size:        size,
					Confidence:  signal.Confidence,
					Reasoning:   signal.Reasoning,
					StopLoss:    signal.StopLoss,
					TargetPrice: signal.TargetPrice,
				}

				tradesToday++
			}
		}
	}

	// Close any remaining open position at last candle
	if openPosition != nil {
		lastCandle := candles[len(candles)-1]
		openPosition.ExitTime = lastCandle.Timestamp
		openPosition.ExitPrice = lastCandle.Close
		openPosition.Duration = openPosition.ExitTime.Sub(openPosition.EntryTime).String()

		if openPosition.Action == "buy" {
			openPosition.ProfitLoss = (openPosition.ExitPrice - openPosition.EntryPrice) * openPosition.Size
		} else {
			openPosition.ProfitLoss = (openPosition.EntryPrice - openPosition.ExitPrice) * openPosition.Size
		}

		balance += openPosition.ProfitLoss
		result.Trades = append(result.Trades, *openPosition)
		result.TotalTrades++

		if openPosition.ProfitLoss > 0 {
			result.WinningTrades++
		} else {
			result.LosingTrades++
		}
	}

	// Calculate final metrics
	result.EndingBalance = balance
	result.ReturnPercent = ((balance - bt.config.StartingBalance) / bt.config.StartingBalance) * 100
	result.MaxDrawdown = maxDrawdown

	if result.TotalTrades > 0 {
		result.WinRate = (float64(result.WinningTrades) / float64(result.TotalTrades)) * 100
		result.TotalProfitLoss = balance - bt.config.StartingBalance
		result.AveragePnL = result.TotalProfitLoss / float64(result.TotalTrades)

		// Calculate Sharpe ratio (simplified)
		returns := []float64{}
		for _, trade := range result.Trades {
			returns = append(returns, trade.PnLPercent)
		}
		result.SharpeRatio = bt.calculateSharpeRatio(returns)

		// Calculate largest win/loss
		for _, trade := range result.Trades {
			if trade.ProfitLoss > result.LargestWin {
				result.LargestWin = trade.ProfitLoss
			}
			if trade.ProfitLoss < result.LargestLoss {
				result.LargestLoss = trade.ProfitLoss
			}
		}

		// Calculate average win/loss size
		if result.WinningTrades > 0 {
			totalWins := 0.0
			for _, trade := range result.Trades {
				if trade.ProfitLoss > 0 {
					totalWins += trade.ProfitLoss
				}
			}
			result.AverageWinSize = totalWins / float64(result.WinningTrades)
		}

		if result.LosingTrades > 0 {
			totalLosses := 0.0
			for _, trade := range result.Trades {
				if trade.ProfitLoss < 0 {
					totalLosses += math.Abs(trade.ProfitLoss)
				}
			}
			result.AverageLossSize = totalLosses / float64(result.LosingTrades)
		}

		// Profit factor
		grossProfit := 0.0
		grossLoss := 0.0
		for _, trade := range result.Trades {
			if trade.ProfitLoss > 0 {
				grossProfit += trade.ProfitLoss
			} else {
				grossLoss += math.Abs(trade.ProfitLoss)
			}
		}
		if grossLoss > 0 {
			result.ProfitFactor = grossProfit / grossLoss
		}

		result.ExpectedValue = result.TotalProfitLoss / float64(result.TotalTrades)
	}

	result.ExecutionTime = time.Since(startTime)

	// Check if strategy passes criteria (same as promotion criteria)
	result.Pass = true
	result.FailureReasons = []string{}

	if result.TotalTrades < 100 {
		result.Pass = false
		result.FailureReasons = append(result.FailureReasons, fmt.Sprintf("Insufficient trades: %d < 100", result.TotalTrades))
	}
	if result.WinRate < 60.0 {
		result.Pass = false
		result.FailureReasons = append(result.FailureReasons, fmt.Sprintf("Win rate too low: %.2f%% < 60%%", result.WinRate))
	}
	if result.SharpeRatio < 1.0 {
		result.Pass = false
		result.FailureReasons = append(result.FailureReasons, fmt.Sprintf("Sharpe ratio too low: %.2f < 1.0", result.SharpeRatio))
	}
	if result.TotalProfitLoss <= 0 {
		result.Pass = false
		result.FailureReasons = append(result.FailureReasons, fmt.Sprintf("Unprofitable: $%.2f", result.TotalProfitLoss))
	}

	log.Printf("[BACKTEST] Complete: %d trades, %.2f%% win rate, Sharpe %.2f, $%.2f P&L (%.2f%% return) | Pass: %v",
		result.TotalTrades, result.WinRate, result.SharpeRatio,
		result.TotalProfitLoss, result.ReturnPercent, result.Pass)

	return result, nil
}

// applySlippage simulates realistic order execution with slippage
func (bt *Backtester) applySlippage(price float64, isBuy bool) float64 {
	slippagePercent := bt.config.Slippage / 100
	if isBuy {
		return price * (1 + slippagePercent) // Buy higher
	}
	return price * (1 - slippagePercent) // Sell lower
}

// calculateSharpeRatio calculates risk-adjusted returns
func (bt *Backtester) calculateSharpeRatio(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Calculate mean return
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// Calculate standard deviation
	variance := 0.0
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}
	variance /= float64(len(returns))
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return 0
	}

	// Sharpe = (mean return - risk-free rate) / std dev
	// Assuming risk-free rate = 0 for simplicity
	return mean / stdDev
}

// GenerateSyntheticData creates realistic historical data for testing
// DEPRECATED: Use GetRealHistoricalData() for production backtesting
// This function remains for quick testing without API dependencies
func GenerateSyntheticData(symbol string, numCandles int, startPrice float64) []HistoricalCandle {
	candles := make([]HistoricalCandle, numCandles)
	currentPrice := startPrice
	startTime := time.Now().Add(-time.Duration(numCandles) * time.Minute)

	for i := 0; i < numCandles; i++ {
		// Random walk with slight upward bias
		change := (math.Sin(float64(i)/10)*0.02 + (math.Cos(float64(i)/20) * 0.015)) * currentPrice
		currentPrice += change

		// OHLCV generation
		open := currentPrice
		high := open * (1 + math.Abs(math.Sin(float64(i)))*0.01)
		low := open * (1 - math.Abs(math.Cos(float64(i)))*0.01)
		close := low + (high-low)*0.5
		volume := 100000 + math.Abs(math.Sin(float64(i)))*50000

		candles[i] = HistoricalCandle{
			Timestamp: startTime.Add(time.Duration(i) * time.Minute),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			Symbol:    symbol,
		}

		currentPrice = close
	}

	return candles
}

// ConvertBinanceCandles converts Binance historical candles to backtester format
// This is used when fetching real historical data from Binance API
func ConvertBinanceCandles(binanceCandles interface{}, symbol string) []HistoricalCandle {
	// Type assertion based on the actual Binance candle type
	// For now, we'll handle a slice of maps (common JSON format)
	candleSlice, ok := binanceCandles.([]interface{})
	if !ok {
		log.Printf("[BACKTEST][ERROR] Invalid binance candles format")
		return []HistoricalCandle{}
	}

	result := make([]HistoricalCandle, 0, len(candleSlice))
	for _, item := range candleSlice {
		candleMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		candle := HistoricalCandle{
			Symbol: symbol,
		}

		// Parse timestamp
		if ts, ok := candleMap["Timestamp"].(time.Time); ok {
			candle.Timestamp = ts
		}

		// Parse OHLCV
		if val, ok := candleMap["Open"].(float64); ok {
			candle.Open = val
		}
		if val, ok := candleMap["High"].(float64); ok {
			candle.High = val
		}
		if val, ok := candleMap["Low"].(float64); ok {
			candle.Low = val
		}
		if val, ok := candleMap["Close"].(float64); ok {
			candle.Close = val
		}
		if val, ok := candleMap["Volume"].(float64); ok {
			candle.Volume = val
		}

		result = append(result, candle)
	}

	return result
}
