package services

import (
	"ares_api/internal/eventbus"
	"ares_api/internal/grpo"
	repository "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"ares_api/internal/repositories"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type TradingService struct {
	TradingRepo *repositories.TradingRepository
	BalanceRepo repository.BalanceRepository
	AssetRepo   repository.AssetRepository
	EventBus    *eventbus.EventBus // Phase 2: Event-driven architecture
	GRPOAgent   *grpo.Agent        // GRPO: Learning from outcomes
}

func NewTradingService(
	tradingRepo *repositories.TradingRepository,
	balanceRepo repository.BalanceRepository,
	assetRepo repository.AssetRepository,
	eb *eventbus.EventBus,
	grpoAgent *grpo.Agent,
) *TradingService {
	return &TradingService{
		TradingRepo: tradingRepo,
		BalanceRepo: balanceRepo,
		AssetRepo:   assetRepo,
		EventBus:    eb,
		GRPOAgent:   grpoAgent,
	}
}

// ExecuteTrade executes a sandbox trade for SOLACE with realistic slippage, fees, and leverage
func (s *TradingService) ExecuteTrade(
	userID uint,
	sessionID uuid.UUID,
	tradingPair string,
	direction string,
	sizeUSD float64,
	reasoning string,
) (*models.SandboxTrade, error) {
	// Validate direction
	if direction != "BUY" && direction != "SELL" {
		return nil, fmt.Errorf("invalid direction: must be BUY or SELL")
	}

	// Check user balance - create if doesn't exist
	balance, err := s.BalanceRepo.GetUSDBalance(userID)
	if err != nil {
		// If balance doesn't exist, create it with $10,000 starting balance
		fmt.Printf("Creating initial balance for user %d\n", userID)
		balanceRecord, createErr := s.BalanceRepo.CreateUSDBalance(userID, 10000.00)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create balance: %w", createErr)
		}
		balance = balanceRecord.Amount
	}

	if balance < sizeUSD {
		return nil, fmt.Errorf("insufficient balance: have $%.2f, need $%.2f", balance, sizeUSD)
	}

	// Get current market price from CoinGecko
	symbol := extractSymbol(tradingPair)
	marketData, err := s.AssetRepo.FetchCoinMarket(symbol, "usd")
	if err != nil {
		return nil, fmt.Errorf("failed to get market price: %w", err)
	}

	currentPrice := marketData.PriceUSD

	// Calculate realistic slippage based on trade size
	slippageBps := calculateSlippage(sizeUSD, marketData.MarketCap)
	slippageMultiplier := 1.0
	if direction == "BUY" {
		slippageMultiplier = 1.0 + (slippageBps / 10000.0) // Pay more when buying
	} else {
		slippageMultiplier = 1.0 - (slippageBps / 10000.0) // Receive less when selling
	}
	executionPrice := currentPrice * slippageMultiplier

	// Calculate Jupiter Exchange-equivalent fees
	// Jupiter: 0.25% platform fee + 0.01% referral = 0.26% total
	jupiterPlatformFee := sizeUSD * 0.0025 // 0.25%
	jupiterReferralFee := sizeUSD * 0.0001 // 0.01%
	totalFees := jupiterPlatformFee + jupiterReferralFee

	// Create market conditions snapshot with slippage data
	marketConditions := models.JSONB{
		"price":           currentPrice,
		"execution_price": executionPrice,
		"slippage_bps":    slippageBps,
		"slippage_cost":   (executionPrice - currentPrice) * (sizeUSD / currentPrice),
		"market_cap":      marketData.MarketCap,
		"change_24h":      marketData.Change24h,
		"timestamp":       time.Now().Unix(),
		"platform_fee":    jupiterPlatformFee,
		"referral_fee":    jupiterReferralFee,
		"total_fees":      totalFees,
	}

	// Generate trade hash
	tradeHash := generateTradeHashString(userID, tradingPair, direction, sizeUSD, executionPrice)

	// Create trade with execution price (includes slippage)
	trade := &models.SandboxTrade{
		UserID:           userID,
		SessionID:        sessionID,
		TradingPair:      tradingPair,
		Direction:        direction,
		Size:             sizeUSD,
		EntryPrice:       executionPrice, // Use execution price, not market price
		Fees:             totalFees,
		Status:           "OPEN",
		OpenedAt:         time.Now(),
		Reasoning:        reasoning,
		MarketConditions: marketConditions,
		TradeHash:        tradeHash,
		LineageTrail:     models.JSONB{},
		SolaceOverride:   false,
	}

	// Save trade
	if err := s.TradingRepo.SaveTrade(trade); err != nil {
		return nil, fmt.Errorf("failed to save trade: %w", err)
	}

	// Deduct balance (size + fees)
	newBalance := balance - sizeUSD - totalFees
	if err := s.BalanceRepo.UpdateBalance(userID, newBalance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// 🚀 Phase 2: Publish trade_executed event
	if s.EventBus != nil {
		event := eventbus.NewTradeExecutedEvent(
			int64(trade.ID),
			trade.TradingPair,
			trade.Direction,
			trade.Size,
			trade.EntryPrice,
			trade.OpenedAt.Format(time.RFC3339),
			"sandbox",
			trade.Status,
			int64(time.Since(trade.OpenedAt).Milliseconds()),
		)
		if err := s.EventBus.Publish(eventbus.EventTypeTradeExecuted, event); err != nil {
			log.Printf("[TRADING][WARN] Failed to publish trade_executed event: %v", err)
			// Don't fail the trade if event publishing fails
		}
	}

	// Check if auto top-up needed
	if err := s.checkAutoTopup(userID); err != nil {
		// Log but don't fail the trade
		fmt.Printf("Auto top-up check failed: %v\n", err)
	}

	return trade, nil
}

// CloseTrade closes an open trade
func (s *TradingService) CloseTrade(userID uint, tradeID uint) (*models.SandboxTrade, error) {
	// Get trade
	trade, err := s.TradingRepo.GetTradeByID(tradeID)
	if err != nil {
		return nil, fmt.Errorf("trade not found: %w", err)
	}

	// Verify ownership
	if trade.UserID != userID {
		return nil, fmt.Errorf("unauthorized: trade belongs to different user")
	}

	// Verify trade is open
	if trade.Status != "OPEN" {
		return nil, fmt.Errorf("trade is already closed")
	}

	// Get current price
	symbol := extractSymbol(trade.TradingPair)
	marketData, err := s.AssetRepo.FetchCoinMarket(symbol, "usd")
	if err != nil {
		return nil, fmt.Errorf("failed to get market price: %w", err)
	}

	currentPrice := marketData.PriceUSD

	// Close trade with current price
	if err := s.TradingRepo.CloseTrade(tradeID, currentPrice); err != nil {
		return nil, fmt.Errorf("failed to close trade: %w", err)
	}

	// Reload trade to get calculated P&L
	trade, err = s.TradingRepo.GetTradeByID(tradeID)
	if err != nil {
		return nil, err
	}

	// Return capital + profit/loss to balance
	returnAmount := trade.Size
	if trade.ProfitLoss != nil {
		returnAmount += *trade.ProfitLoss
	}

	balance, _ := s.BalanceRepo.GetUSDBalance(userID)
	newBalance := balance + returnAmount
	if err := s.BalanceRepo.UpdateBalance(userID, newBalance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// Update realized P&L
	s.updateRealizedPnL(userID, trade.ProfitLoss)

	// 🧠 GRPO: Record reward for learning
	if s.GRPOAgent != nil && trade.ProfitLoss != nil {
		// Extract simple tokens from reasoning (split by spaces, filter empty)
		tokens := []string{trade.TradingPair, trade.Direction} // Use trading pair and direction as tokens
		if trade.Reasoning != "" {
			// Could add more sophisticated tokenization here
		}

		if err := s.GRPOAgent.RecordReward(
			int(trade.ID),
			0, // trace_id (would link to decision_traces if integrated)
			*trade.ProfitLoss,
			trade.Size,
			70.0, // Default confidence (could calculate from trade metrics)
			tokens,
		); err != nil {
			log.Printf("[GRPO][WARN] Failed to record reward for trade %d: %v", trade.ID, err)
			// Don't fail the trade if GRPO recording fails
		}
	}

	// 🚀 Phase 2: Publish trade_closed event
	if s.EventBus != nil {
		event := eventbus.NewTradeExecutedEvent(
			int64(trade.ID),
			trade.TradingPair,
			trade.Direction,
			trade.Size,
			trade.EntryPrice,
			trade.OpenedAt.Format(time.RFC3339),
			"sandbox",
			"CLOSED",
			int64(time.Since(trade.OpenedAt).Milliseconds()),
		)
		if err := s.EventBus.Publish(eventbus.EventTypeTradeExecuted, event); err != nil {
			log.Printf("[TRADING][WARN] Failed to publish trade_closed event: %v", err)
			// Don't fail the trade if event publishing fails
		}
	}

	return trade, nil
}

// CloseAllTrades closes all open trades (for kill-switch)
func (s *TradingService) CloseAllTrades(userID uint) (int, error) {
	openTrades, err := s.TradingRepo.GetOpenTrades(userID)
	if err != nil {
		return 0, err
	}

	closed := 0
	for _, trade := range openTrades {
		if _, err := s.CloseTrade(userID, trade.ID); err != nil {
			fmt.Printf("Failed to close trade %d: %v\n", trade.ID, err)
			continue
		}
		closed++
	}

	return closed, nil
}

// GetTradeHistory gets trade history with pagination
func (s *TradingService) GetTradeHistory(userID uint, limit int, offset int) ([]models.SandboxTrade, error) {
	if limit == 0 {
		limit = 50
	}
	return s.TradingRepo.GetTradeHistory(userID, limit, offset)
}

// GetOpenTrades gets all open trades
func (s *TradingService) GetOpenTrades(userID uint) ([]models.SandboxTrade, error) {
	return s.TradingRepo.GetOpenTrades(userID)
}

// GetPerformance calculates trading performance metrics
func (s *TradingService) GetPerformance(userID uint) (*models.TradingPerformance, error) {
	return s.TradingRepo.GetPerformanceMetrics(userID)
}

// checkAutoTopup checks if balance needs top-up and executes if enabled
func (s *TradingService) checkAutoTopup(userID uint) error {
	// Get balance record with top-up settings
	balanceRecord, err := s.BalanceRepo.GetBalanceRecord(userID)
	if err != nil {
		return err
	}

	// Check if auto top-up is enabled
	if !balanceRecord.AutoTopup {
		return nil
	}

	// Check if balance is below threshold
	if balanceRecord.Amount < balanceRecord.TopupThreshold {
		// Execute top-up
		newBalance := balanceRecord.Amount + balanceRecord.TopupAmount
		if err := s.BalanceRepo.UpdateBalance(userID, newBalance); err != nil {
			return err
		}

		// Update total deposits
		balanceRecord.TotalDeposits += balanceRecord.TopupAmount
		if err := s.BalanceRepo.UpdateBalanceRecord(balanceRecord); err != nil {
			return err
		}

		fmt.Printf("Auto top-up executed: Added $%.2f to user %d balance\n", balanceRecord.TopupAmount, userID)
	}

	return nil
}

// updateRealizedPnL updates the realized P&L in balance record
func (s *TradingService) updateRealizedPnL(userID uint, pnl *float64) error {
	if pnl == nil {
		return nil
	}

	balanceRecord, err := s.BalanceRepo.GetBalanceRecord(userID)
	if err != nil {
		return err
	}

	balanceRecord.RealizedPnL += *pnl
	return s.BalanceRepo.UpdateBalanceRecord(balanceRecord)
}

// Helper functions

func extractSymbol(tradingPair string) string {
	// Extract symbol from pair like "BTC/USDC" -> "bitcoin"
	// This is a simplified version - should map to CoinGecko IDs
	switch tradingPair {
	case "BTC/USDC", "BTC/USD":
		return "bitcoin"
	case "ETH/USDC", "ETH/USD":
		return "ethereum"
	case "SOL/USDC", "SOL/USD":
		return "solana"
	default:
		return "bitcoin" // fallback
	}
}

func generateTradeHashString(userID uint, pair string, direction string, size float64, price float64) string {
	data := fmt.Sprintf("%d-%s-%s-%.8f-%.8f-%d", userID, pair, direction, size, price, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// calculateSlippage calculates realistic slippage in basis points based on trade size and market cap
// Simulates real market impact similar to Jupiter Exchange
func calculateSlippage(tradeSizeUSD float64, marketCapUSD float64) float64 {
	if marketCapUSD == 0 {
		return 5.0 // Default 5 bps if market cap unknown
	}

	// Calculate trade size as percentage of market cap
	tradeImpact := (tradeSizeUSD / marketCapUSD) * 10000.0 // bps

	// Base slippage: 1-5 bps depending on trade size
	var baseSlippage float64
	if tradeSizeUSD < 100 {
		baseSlippage = 1.0 // Tiny trades: 1 bps
	} else if tradeSizeUSD < 1000 {
		baseSlippage = 2.0 // Small trades: 2 bps
	} else if tradeSizeUSD < 5000 {
		baseSlippage = 3.5 // Medium trades: 3.5 bps
	} else {
		baseSlippage = 5.0 // Large trades: 5 bps
	}

	// Add market impact (scales with trade size / market cap)
	impactSlippage := tradeImpact * 0.5 // 50% of impact translates to slippage

	// Total slippage (capped at 50 bps = 0.5% for realism)
	totalSlippage := baseSlippage + impactSlippage
	if totalSlippage > 50.0 {
		totalSlippage = 50.0
	}

	return totalSlippage
}

// ExecuteLeveragedTrade executes a trade with leverage (1x - 20x)
func (s *TradingService) ExecuteLeveragedTrade(
	userID uint,
	sessionID uuid.UUID,
	tradingPair string,
	direction string,
	sizeUSD float64,
	leverage float64,
	reasoning string,
) (*models.SandboxTrade, error) {
	// Validate leverage (1x - 20x)
	if leverage < 1.0 || leverage > 20.0 {
		return nil, fmt.Errorf("invalid leverage: must be between 1x and 20x (got %.2fx)", leverage)
	}

	// Calculate collateral required (size / leverage)
	collateralRequired := sizeUSD / leverage

	// Check user balance
	balance, err := s.BalanceRepo.GetUSDBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if balance < collateralRequired {
		return nil, fmt.Errorf("insufficient collateral: need $%.2f, have $%.2f (%.2fx leverage on $%.2f position)",
			collateralRequired, balance, leverage, sizeUSD)
	}

	// Get market price
	symbol := extractSymbol(tradingPair)
	marketData, err := s.AssetRepo.FetchCoinMarket(symbol, "usd")
	if err != nil {
		return nil, fmt.Errorf("failed to get market price: %w", err)
	}

	currentPrice := marketData.PriceUSD

	// Calculate slippage (higher for leveraged trades due to liquidation risk)
	baseSlippage := calculateSlippage(sizeUSD, marketData.MarketCap)
	leverageSlippage := baseSlippage * (1.0 + (leverage-1.0)*0.1) // +10% slippage per leverage unit
	slippageMultiplier := 1.0
	if direction == "BUY" {
		slippageMultiplier = 1.0 + (leverageSlippage / 10000.0)
	} else {
		slippageMultiplier = 1.0 - (leverageSlippage / 10000.0)
	}
	executionPrice := currentPrice * slippageMultiplier

	// Calculate fees (Jupiter base + leverage borrowing fee)
	jupiterPlatformFee := sizeUSD * 0.0025 // 0.25%
	jupiterReferralFee := sizeUSD * 0.0001 // 0.01%

	// Leverage borrowing fee: 0.01% per hour * leverage factor
	// Simulated as opening fee (in reality, this would accrue hourly)
	leverageFee := sizeUSD * 0.0001 * leverage

	totalFees := jupiterPlatformFee + jupiterReferralFee + leverageFee

	// Calculate liquidation price
	// For BUY: liquidation when price drops enough to wipe out collateral
	// For SELL: liquidation when price rises enough to wipe out collateral
	liquidationBuffer := 0.05 // 5% buffer before liquidation
	var liquidationPrice float64
	if direction == "BUY" {
		// Price can drop by (collateral - fees) / (size / execution_price) before liquidation
		maxLoss := collateralRequired - totalFees
		liquidationPrice = executionPrice * (1.0 - (maxLoss/sizeUSD)*(1.0-liquidationBuffer))
	} else {
		// Price can rise by (collateral - fees) / (size / execution_price) before liquidation
		maxLoss := collateralRequired - totalFees
		liquidationPrice = executionPrice * (1.0 + (maxLoss/sizeUSD)*(1.0-liquidationBuffer))
	}

	// Market conditions with leverage data
	marketConditions := models.JSONB{
		"price":             currentPrice,
		"execution_price":   executionPrice,
		"slippage_bps":      leverageSlippage,
		"market_cap":        marketData.MarketCap,
		"change_24h":        marketData.Change24h,
		"leverage":          leverage,
		"collateral":        collateralRequired,
		"position_size":     sizeUSD,
		"liquidation_price": liquidationPrice,
		"platform_fee":      jupiterPlatformFee,
		"referral_fee":      jupiterReferralFee,
		"leverage_fee":      leverageFee,
		"total_fees":        totalFees,
		"timestamp":         time.Now().Unix(),
	}

	// Generate trade hash
	tradeHash := generateTradeHashString(userID, tradingPair, direction, sizeUSD, executionPrice)

	// Create leveraged trade
	trade := &models.SandboxTrade{
		UserID:           userID,
		SessionID:        sessionID,
		TradingPair:      tradingPair,
		Direction:        direction,
		Size:             sizeUSD, // Full position size
		EntryPrice:       executionPrice,
		Fees:             totalFees,
		Status:           "OPEN",
		OpenedAt:         time.Now(),
		Reasoning:        fmt.Sprintf("[%.2fx LEVERAGE] %s", leverage, reasoning),
		MarketConditions: marketConditions,
		TradeHash:        tradeHash,
		LineageTrail: models.JSONB{
			"leverage":          leverage,
			"collateral":        collateralRequired,
			"liquidation_price": liquidationPrice,
		},
		SolaceOverride: false,
	}

	// Save trade
	if err := s.TradingRepo.SaveTrade(trade); err != nil {
		return nil, fmt.Errorf("failed to save trade: %w", err)
	}

	// Deduct ONLY collateral from balance (not full position size)
	newBalance := balance - collateralRequired - totalFees
	if err := s.BalanceRepo.UpdateBalance(userID, newBalance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	return trade, nil
}
