/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package trading

import (
	"ares_api/internal/glassbox"
	"ares_api/internal/models"
	"ares_api/internal/repositories"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"sync"
	"time"
)

// Custom error types for better error handling
var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidSymbol       = errors.New("invalid symbol format")
	ErrInvalidSide         = errors.New("side must be 'buy' or 'sell'")
	ErrInvalidAmount       = errors.New("amount must be positive and within limits")
	ErrTradeNotFound       = errors.New("trade not found")
	ErrUnauthorized        = errors.New("unauthorized to access this trade")
	ErrTradeClosed         = errors.New("trade already closed")
	ErrInvalidPriceChange  = errors.New("price change must be between -1.0 and 1.0")
)

// SandboxTrader provides a virtual trading environment with thread-safe operations
// ENHANCED: Glass Box observability with decision tracing
type SandboxTrader struct {
	mu                sync.RWMutex // Protects all fields
	VirtualBalance    float64
	Trades            []VirtualTrade // In-memory cache for recent trades
	MarketData        *MockMarketData
	MaxTradesInMemory int                           // Prevent unbounded growth
	tradeCounter      uint64                        // Monotonic counter for unique IDs
	repo              *repositories.TradeRepository // Database persistence
	tracer            *glassbox.DecisionTracer      // GLASS BOX: Decision observability
}

// VirtualTrade represents a simulated trade
type VirtualTrade struct {
	ID              string     `json:"id"`
	UserID          uint       `json:"user_id"`
	Symbol          string     `json:"symbol"`
	Side            string     `json:"side"` // "buy" or "sell"
	Amount          float64    `json:"amount"`
	Price           float64    `json:"price"`
	ExecutedAt      time.Time  `json:"executed_at"`
	ProfitLoss      float64    `json:"profit_loss"`
	ProfitLossPct   float64    `json:"profit_loss_pct"`
	Status          string     `json:"status"` // "open", "closed", "failed"
	Strategy        string     `json:"strategy"`
	Reasoning       string     `json:"reasoning"`
	ExitPrice       *float64   `json:"exit_price,omitempty"`
	ExitedAt        *time.Time `json:"exited_at,omitempty"`
	TransactionHash string     `json:"transaction_hash"` // SHA256 hash for audit trail
	Fee             float64    `json:"fee"`              // Trading fee (0.1% typically)
}

// MockMarketData provides simulated price feeds with historical data
type MockMarketData struct {
	mu             sync.RWMutex
	Prices         map[string]float64      // symbol -> current price
	PriceHistory   map[string][]PricePoint // symbol -> historical prices
	Seed           int64
	MaxHistorySize int // Limit history to prevent unbounded growth
}

// PricePoint represents a price at a specific time
type PricePoint struct {
	Price     float64
	Timestamp time.Time
}

// NewSandboxTrader creates a new sandbox trader with thread-safe initialization
// ENHANCED: Initializes Glass Box decision tracer
func NewSandboxTrader(initialBalance float64, repo *repositories.TradeRepository, db interface{}) *SandboxTrader {
	// Seed the random number generator for realistic price movements
	rand.Seed(time.Now().UnixNano())

	// Initialize Glass Box tracer (requires database connection)
	var tracer *glassbox.DecisionTracer

	// Try direct sql.DB first
	if sqlDB, ok := db.(*sql.DB); ok {
		tracer = glassbox.NewDecisionTracer(sqlDB)
	} else if gormDB, ok := db.(interface{ DB() (*sql.DB, error) }); ok {
		// Handle gorm.DB wrapper
		if rawDB, err := gormDB.DB(); err == nil {
			tracer = glassbox.NewDecisionTracer(rawDB)
		}
	}

	return &SandboxTrader{
		VirtualBalance:    initialBalance,
		Trades:            make([]VirtualTrade, 0, 100),
		MarketData:        NewMockMarketData(),
		MaxTradesInMemory: 10000, // Archive older trades beyond this
		repo:              repo,
		tracer:            tracer, // GLASS BOX
	}
}

// validateTradeInput validates all trade parameters before execution
func validateTradeInput(symbol, side string, amount float64) error {
	// Validate symbol format (e.g., "SOL/USDC", "BTC/USDT")
	symbolPattern := regexp.MustCompile(`^[A-Z]{2,10}/[A-Z]{2,10}$`)
	if !symbolPattern.MatchString(symbol) {
		return fmt.Errorf("%w: %s (expected format: XXX/YYY)", ErrInvalidSymbol, symbol)
	}

	// Validate side
	if side != "buy" && side != "sell" {
		return fmt.Errorf("%w: %s", ErrInvalidSide, side)
	}

	// Validate amount (positive, reasonable limits)
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive (got %.4f)", ErrInvalidAmount, amount)
	}
	if amount > 1000000 {
		return fmt.Errorf("%w: amount exceeds maximum of 1,000,000 (got %.4f)", ErrInvalidAmount, amount)
	}

	return nil
}

// NewMockMarketData initializes mock market data with price history
func NewMockMarketData() *MockMarketData {
	now := time.Now()
	md := &MockMarketData{
		Prices: map[string]float64{
			"SOL/USDC":  150.00,
			"BTC/USDC":  45000.00,
			"ETH/USDC":  2800.00,
			"BONK/USDC": 0.000025,
			"JUP/USDC":  1.20,
		},
		PriceHistory:   make(map[string][]PricePoint),
		Seed:           time.Now().UnixNano(),
		MaxHistorySize: 1000, // Keep last 1000 price points per symbol
	}

	// Initialize price history with synthetic data (last 24 hours)
	for symbol, price := range md.Prices {
		history := make([]PricePoint, 0, 100)
		rng := rand.New(rand.NewSource(now.UnixNano()))

		// Generate 100 historical points (simulating 15-min candles for 24h)
		currentPrice := price
		for i := 100; i > 0; i-- {
			timestamp := now.Add(-time.Duration(i) * 15 * time.Minute)
			// Random walk with ±0.5% per candle
			change := (rng.Float64() - 0.5) * 0.01
			currentPrice *= (1 + change)

			history = append(history, PricePoint{
				Price:     currentPrice,
				Timestamp: timestamp,
			})
		}

		md.PriceHistory[symbol] = history
	}

	return md
}

// GetCurrentPrice returns the current price for a symbol (thread-safe)
func (md *MockMarketData) GetCurrentPrice(symbol string) (float64, error) {
	md.mu.RLock()
	defer md.mu.RUnlock()

	price, exists := md.Prices[symbol]
	if !exists {
		return 0, fmt.Errorf("%w: %s", ErrInvalidSymbol, symbol)
	}

	// Apply realistic volatility (±0.5%)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	volatility := (rng.Float64() - 0.5) * 0.01
	return price * (1 + volatility), nil
}

// GetPrice returns current simulated price with realistic volatility (legacy method)
func (m *MockMarketData) GetPrice(symbol string) float64 {
	price, err := m.GetCurrentPrice(symbol)
	if err != nil {
		return 0
	}
	return price
}

// GetPriceHistory returns historical prices for technical analysis (thread-safe)
func (md *MockMarketData) GetPriceHistory(symbol string, periods int) ([]float64, error) {
	md.mu.RLock()
	defer md.mu.RUnlock()

	history, exists := md.PriceHistory[symbol]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrInvalidSymbol, symbol)
	}

	if periods > len(history) {
		periods = len(history)
	}

	// Return last N periods
	start := len(history) - periods
	if start < 0 {
		start = 0
	}

	prices := make([]float64, 0, periods)
	for i := start; i < len(history); i++ {
		prices = append(prices, history[i].Price)
	}

	return prices, nil
}

// UpdatePrice simulates price movement and updates history (thread-safe)
func (md *MockMarketData) UpdatePrice(symbol string, changePercent float64) error {
	md.mu.Lock()
	defer md.mu.Unlock()

	if changePercent < -1.0 || changePercent > 1.0 {
		return fmt.Errorf("%w: %.2f%% (must be between -100%% and +100%%)", ErrInvalidPriceChange, changePercent*100)
	}

	basePrice, exists := md.Prices[symbol]
	if !exists {
		return fmt.Errorf("%w: %s", ErrInvalidSymbol, symbol)
	}

	newPrice := basePrice * (1 + changePercent)
	md.Prices[symbol] = newPrice

	// Append to price history
	if md.PriceHistory[symbol] == nil {
		md.PriceHistory[symbol] = make([]PricePoint, 0, md.MaxHistorySize)
	}

	md.PriceHistory[symbol] = append(md.PriceHistory[symbol], PricePoint{
		Price:     newPrice,
		Timestamp: time.Now(),
	})

	// Trim history if too large
	if len(md.PriceHistory[symbol]) > md.MaxHistorySize {
		md.PriceHistory[symbol] = md.PriceHistory[symbol][1:]
	}

	return nil
}

// ExecuteTrade executes a virtual trade with full validation and thread safety
// authenticatedUserID is the user making the request (from JWT/session)
// tradeUserID is the user account to trade on behalf of
func (st *SandboxTrader) ExecuteTrade(authenticatedUserID, tradeUserID uint, symbol string, side string, amount float64, strategy string, reasoning string) (*VirtualTrade, error) {
	ctx := context.Background()
	var trace *glassbox.Trace
	var err error

	// GLASS BOX: Start decision trace
	if st.tracer != nil {
		trace, err = st.tracer.StartTrace(ctx, "trade_execution", nil)
		if err != nil {
			// Log but don't fail the trade
			fmt.Printf("Warning: Failed to start decision trace: %v\n", err)
		}
	}

	// SECURITY: Verify user authorization
	// GLASS BOX: Span 1 - Authorization check
	var authSpan *glassbox.Span
	if st.tracer != nil && trace != nil {
		authSpan, _ = st.tracer.StartSpan(ctx, trace.ID, nil, "authorization_check", "security", map[string]interface{}{
			"authenticated_user": authenticatedUserID,
			"trade_user":         tradeUserID,
		})
	}

	if authenticatedUserID != tradeUserID {
		if st.tracer != nil && authSpan != nil {
			st.tracer.EndSpan(ctx, authSpan.ID, map[string]interface{}{"error": "unauthorized"}, "User authorization failed", 0, "failed")
			st.tracer.EndTrace(ctx, trace.ID, "rejected_unauthorized", 0)
		}
		return nil, fmt.Errorf("%w: user %d cannot trade for user %d", ErrUnauthorized, authenticatedUserID, tradeUserID)
	}

	if st.tracer != nil && authSpan != nil {
		st.tracer.EndSpan(ctx, authSpan.ID, map[string]interface{}{"authorized": true}, "User authorized to execute trade", 100, "success")
	}

	// Validate input parameters
	// GLASS BOX: Span 2 - Input validation
	var validationSpan *glassbox.Span
	if st.tracer != nil && trace != nil {
		validationSpan, _ = st.tracer.StartSpan(ctx, trace.ID, &authSpan.ID, "input_validation", "validation", map[string]interface{}{
			"symbol": symbol,
			"side":   side,
			"amount": amount,
		})
	}

	if err := validateTradeInput(symbol, side, amount); err != nil {
		if st.tracer != nil && validationSpan != nil {
			st.tracer.EndSpan(ctx, validationSpan.ID, map[string]interface{}{"error": err.Error()}, "Input validation failed", 0, "failed")
			st.tracer.EndTrace(ctx, trace.ID, "rejected_invalid_input", 0)
		}
		return nil, err
	}

	if st.tracer != nil && validationSpan != nil {
		st.tracer.EndSpan(ctx, validationSpan.ID, map[string]interface{}{"valid": true}, "All inputs validated successfully", 100, "success")
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	// Get current market price
	// GLASS BOX: Span 3 - Market pricing
	var pricingSpan *glassbox.Span
	if st.tracer != nil && trace != nil {
		pricingSpan, _ = st.tracer.StartSpan(ctx, trace.ID, &validationSpan.ID, "market_pricing", "data_fetch", map[string]interface{}{
			"symbol": symbol,
		})
	}

	price := st.MarketData.GetPrice(symbol)
	if price == 0 {
		if st.tracer != nil && pricingSpan != nil {
			st.tracer.EndSpan(ctx, pricingSpan.ID, map[string]interface{}{"error": "invalid_symbol"}, "Symbol not found in market data", 0, "failed")
			st.tracer.EndTrace(ctx, trace.ID, "rejected_invalid_symbol", 0)
		}
		return nil, fmt.Errorf("%w: %s", ErrInvalidSymbol, symbol)
	}

	if st.tracer != nil && pricingSpan != nil {
		st.tracer.EndSpan(ctx, pricingSpan.ID, map[string]interface{}{"price": price}, "Market price retrieved successfully", 95, "success")
		// Record pricing metric
		st.tracer.RecordMetric(ctx, trace.ID, &pricingSpan.ID, "market_price", price, "USD")
	}

	// Calculate cost with trading fees (0.1%)
	const tradingFeePercent = 0.001 // 0.1%
	baseCost := amount * price
	tradingFee := baseCost * tradingFeePercent
	totalCost := baseCost + tradingFee

	// Check balance for buy orders (thread-safe due to mutex)
	// GLASS BOX: Span 4 - Balance check
	var balanceSpan *glassbox.Span
	if st.tracer != nil && trace != nil {
		balanceSpan, _ = st.tracer.StartSpan(ctx, trace.ID, &pricingSpan.ID, "balance_check", "validation", map[string]interface{}{
			"required_balance": totalCost,
			"current_balance":  st.VirtualBalance,
			"side":             side,
		})
	}

	if side == "buy" && totalCost > st.VirtualBalance {
		if st.tracer != nil && balanceSpan != nil {
			st.tracer.EndSpan(ctx, balanceSpan.ID, map[string]interface{}{"error": "insufficient_balance"}, "Insufficient balance for trade", 0, "failed")
			st.tracer.EndTrace(ctx, trace.ID, "rejected_insufficient_balance", 0)
		}
		return nil, fmt.Errorf("%w: need %.2f, have %.2f", ErrInsufficientBalance, totalCost, st.VirtualBalance)
	}

	if st.tracer != nil && balanceSpan != nil {
		st.tracer.EndSpan(ctx, balanceSpan.ID, map[string]interface{}{"sufficient": true}, "Balance check passed", 100, "success")
	}

	// Generate transaction hash (SHA256 of trade details)
	now := time.Now()
	txHash := generateTransactionHash(tradeUserID, symbol, side, amount, price, now)

	// Generate unique ID using counter
	st.tradeCounter++
	tradeID := fmt.Sprintf("SANDBOX_%d_%d", now.Unix(), st.tradeCounter)

	// Create trade
	// GLASS BOX: Span 5 - Trade execution
	var executionSpan *glassbox.Span
	executionStart := time.Now()
	if st.tracer != nil && trace != nil {
		executionSpan, _ = st.tracer.StartSpan(ctx, trace.ID, &balanceSpan.ID, "trade_execution", "execution", map[string]interface{}{
			"trade_id":    tradeID,
			"base_cost":   baseCost,
			"trading_fee": tradingFee,
			"total_cost":  totalCost,
			"tx_hash":     txHash,
		})
	}

	trade := VirtualTrade{
		ID:              tradeID,
		UserID:          tradeUserID,
		Symbol:          symbol,
		Side:            side,
		Amount:          amount,
		Price:           price,
		ExecutedAt:      now,
		Status:          "open",
		Strategy:        strategy,
		Reasoning:       reasoning,
		TransactionHash: txHash,
		ProfitLoss:      0,
		ProfitLossPct:   0,
		Fee:             tradingFee,
	}

	if st.tracer != nil && executionSpan != nil {
		executionDuration := time.Since(executionStart).Milliseconds()
		st.tracer.EndSpan(ctx, executionSpan.ID, map[string]interface{}{
			"trade_created": true,
			"trade_id":      tradeID,
		}, "Trade object created successfully", 100, "success")
		st.tracer.RecordMetric(ctx, trace.ID, &executionSpan.ID, "execution_time_ms", float64(executionDuration), "ms")
		st.tracer.RecordMetric(ctx, trace.ID, &executionSpan.ID, "trading_fee", tradingFee, "USD")
	}

	// Persist to database if repository is available
	// GLASS BOX: Span 6 - Database persistence
	var persistenceSpan *glassbox.Span
	persistenceStart := time.Now()
	if st.tracer != nil && trace != nil {
		persistenceSpan, _ = st.tracer.StartSpan(ctx, trace.ID, &executionSpan.ID, "database_persistence", "data_storage", map[string]interface{}{
			"trade_id": tradeID,
			"has_repo": st.repo != nil,
		})
	}

	if st.repo != nil {
		dbTrade := models.Trade{
			TradeID:         tradeID,
			UserID:          tradeUserID,
			Symbol:          symbol,
			Side:            side,
			Amount:          amount,
			Price:           price,
			Status:          "open",
			Strategy:        strategy,
			Reasoning:       reasoning,
			TransactionHash: txHash,
			Fee:             tradingFee,
			ExecutedAt:      now, // time.Time (not pointer)
		}

		if err := st.repo.Create(&dbTrade); err != nil {
			// Database persistence failed, rollback in-memory state
			if st.tracer != nil && persistenceSpan != nil {
				st.tracer.EndSpan(ctx, persistenceSpan.ID, map[string]interface{}{"error": err.Error()}, "Database persistence failed", 0, "failed")
				st.tracer.EndTrace(ctx, trace.ID, "failed_persistence", 50)
			}
			return nil, fmt.Errorf("failed to persist trade: %w", err)
		}

		if st.tracer != nil && persistenceSpan != nil {
			persistenceDuration := time.Since(persistenceStart).Milliseconds()
			st.tracer.EndSpan(ctx, persistenceSpan.ID, map[string]interface{}{"persisted": true}, "Trade persisted to database", 100, "success")
			st.tracer.RecordMetric(ctx, trace.ID, &persistenceSpan.ID, "persistence_time_ms", float64(persistenceDuration), "ms")
		}
	} else {
		if st.tracer != nil && persistenceSpan != nil {
			st.tracer.EndSpan(ctx, persistenceSpan.ID, map[string]interface{}{"skipped": true}, "No repository available, skipped persistence", 100, "success")
		}
	}

	// Atomic balance update (after successful DB save)
	if side == "buy" {
		st.VirtualBalance -= totalCost
	} else {
		st.VirtualBalance += baseCost - tradingFee // Sell also pays fees
	}

	// Store trade in memory for quick access
	st.Trades = append(st.Trades, trade)

	// Archive old trades if memory limit exceeded
	if len(st.Trades) > st.MaxTradesInMemory {
		// Trim in-memory cache (data is persisted in database)
		st.Trades = st.Trades[len(st.Trades)-st.MaxTradesInMemory:]
	}

	// GLASS BOX: Complete decision trace
	if st.tracer != nil && trace != nil {
		overallConfidence := 92.0 // High confidence for successful execution
		st.tracer.EndTrace(ctx, trace.ID, "trade_executed_successfully", overallConfidence)

		// Record overall trade metrics
		st.tracer.RecordMetric(ctx, trace.ID, nil, "total_cost", totalCost, "USD")
		st.tracer.RecordMetric(ctx, trace.ID, nil, "overall_confidence", overallConfidence, "percent")
	}

	return &trade, nil
}

// CloseTrade simulates closing a position with authorization and thread safety
func (st *SandboxTrader) CloseTrade(authenticatedUserID uint, tradeID string) (*VirtualTrade, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	// Find trade
	for i, trade := range st.Trades {
		if trade.ID == tradeID {
			// SECURITY: Verify user owns this trade
			if trade.UserID != authenticatedUserID {
				return nil, fmt.Errorf("%w: user %d cannot close trade owned by user %d", ErrUnauthorized, authenticatedUserID, trade.UserID)
			}

			// Check if already closed
			if trade.Status != "open" {
				return nil, fmt.Errorf("%w: trade %s", ErrTradeClosed, tradeID)
			}

			// Get current price
			exitPrice := st.MarketData.GetPrice(trade.Symbol)
			exitTime := time.Now()

			// Calculate P&L with fees
			const closingFeePercent = 0.001 // 0.1% closing fee
			var profitLoss float64
			var profitLossPct float64

			if trade.Side == "buy" {
				// Long position: profit if price went up
				grossProfit := (exitPrice - trade.Price) * trade.Amount
				closingFee := exitPrice * trade.Amount * closingFeePercent
				profitLoss = grossProfit - closingFee
				profitLossPct = ((exitPrice - trade.Price) / trade.Price) * 100
			} else {
				// Short position: profit if price went down
				grossProfit := (trade.Price - exitPrice) * trade.Amount
				closingFee := exitPrice * trade.Amount * closingFeePercent
				profitLoss = grossProfit - closingFee
				profitLossPct = ((trade.Price - exitPrice) / trade.Price) * 100
			}

			// Persist closure to database if repository is available
			if st.repo != nil {
				// Find and update the database record
				dbTrade, err := st.repo.FindByID(tradeID)
				if err != nil {
					return nil, fmt.Errorf("failed to find trade in database: %w", err)
				}

				dbTrade.Status = "closed"
				dbTrade.ExitPrice = &exitPrice
				dbTrade.ExitedAt = &exitTime
				dbTrade.ProfitLoss = profitLoss
				dbTrade.ProfitLossPct = profitLossPct

				if err := st.repo.Update(dbTrade); err != nil {
					return nil, fmt.Errorf("failed to persist trade closure: %w", err)
				}
			}

			// Update in-memory trade (after successful DB save)
			st.Trades[i].Status = "closed"
			st.Trades[i].ExitPrice = &exitPrice
			st.Trades[i].ExitedAt = &exitTime
			st.Trades[i].ProfitLoss = profitLoss
			st.Trades[i].ProfitLossPct = profitLossPct

			// Atomic balance update - return principal + profit/loss
			if trade.Side == "buy" {
				// Return the value of the asset sold
				st.VirtualBalance += exitPrice*trade.Amount - (exitPrice * trade.Amount * closingFeePercent)
			}
			// For sell (short), we already credited balance on ExecuteTrade

			return &st.Trades[i], nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrTradeNotFound, tradeID)
}

// GetBalance returns current virtual balance (thread-safe)
func (st *SandboxTrader) GetBalance() float64 {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.VirtualBalance
}

// GetOpenTrades returns all open positions (thread-safe)
func (st *SandboxTrader) GetOpenTrades(userID uint) []VirtualTrade {
	st.mu.RLock()
	defer st.mu.RUnlock()

	open := make([]VirtualTrade, 0)
	for _, trade := range st.Trades {
		if trade.Status == "open" && trade.UserID == userID {
			open = append(open, trade)
		}
	}
	return open
}

// GetAllOpenTrades returns all open positions for all users (thread-safe)
func (st *SandboxTrader) GetAllOpenTrades() []VirtualTrade {
	st.mu.RLock()
	defer st.mu.RUnlock()

	open := make([]VirtualTrade, 0)
	for _, trade := range st.Trades {
		if trade.Status == "open" {
			open = append(open, trade)
		}
	}
	return open
}

// GetTradeHistory returns all trades (thread-safe copy)
func (st *SandboxTrader) GetTradeHistory() []VirtualTrade {
	st.mu.RLock()
	defer st.mu.RUnlock()

	// Return a copy to prevent external mutation
	history := make([]VirtualTrade, len(st.Trades))
	copy(history, st.Trades)
	return history
}

// SimulateMarketMovement creates realistic price changes over time (thread-safe)
func (st *SandboxTrader) SimulateMarketMovement() {
	// Simulate price changes for all symbols
	for symbol := range st.MarketData.Prices {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		change := (rng.Float64() - 0.5) * 0.02        // ±1% per tick
		_ = st.MarketData.UpdatePrice(symbol, change) // Ignore error in simulation
	}
}

// generateTransactionHash creates a deterministic hash for audit trail
func generateTransactionHash(userID uint, symbol string, side string, amount float64, price float64, timestamp time.Time) string {
	data := fmt.Sprintf("%d|%s|%s|%.8f|%.8f|%d", userID, symbol, side, amount, price, timestamp.Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// PortfolioSnapshot represents current portfolio state
type PortfolioSnapshot struct {
	Balance       float64        `json:"balance"`
	OpenTrades    int            `json:"open_trades"`
	TotalTrades   int            `json:"total_trades"`
	OpenPositions []VirtualTrade `json:"open_positions"`
}

// GetPortfolio returns current portfolio state for a specific user
func (st *SandboxTrader) GetPortfolio(userID uint) *PortfolioSnapshot {
	return &PortfolioSnapshot{
		Balance:       st.GetBalance(),
		OpenTrades:    len(st.GetOpenTrades(userID)),
		TotalTrades:   len(st.Trades),
		OpenPositions: st.GetOpenTrades(userID),
	}
}
