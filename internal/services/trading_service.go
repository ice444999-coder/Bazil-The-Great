package services

import (
	"ares_api/internal/models"
	"ares_api/internal/repositories"
	repository "ares_api/internal/interfaces/repository"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TradingService struct {
	TradingRepo  *repositories.TradingRepository
	BalanceRepo  repository.BalanceRepository
	AssetRepo    repository.AssetRepository
}

func NewTradingService(
	tradingRepo *repositories.TradingRepository,
	balanceRepo repository.BalanceRepository,
	assetRepo repository.AssetRepository,
) *TradingService {
	return &TradingService{
		TradingRepo: tradingRepo,
		BalanceRepo: balanceRepo,
		AssetRepo:   assetRepo,
	}
}

// ExecuteTrade executes a sandbox trade for SOLACE
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

	// Check user balance
	balance, err := s.BalanceRepo.GetUSDBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
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

	// Calculate fees (0.1% for sandbox)
	fees := sizeUSD * 0.001

	// Create market conditions snapshot
	marketConditions := models.JSONB{
		"price":      currentPrice,
		"market_cap": marketData.MarketCap,
		"change_24h": marketData.Change24h,
		"timestamp":  time.Now().Unix(),
	}

	// Generate trade hash
	tradeHash := generateTradeHashString(userID, tradingPair, direction, sizeUSD, currentPrice)

	// Create trade
	trade := &models.SandboxTrade{
		UserID:           userID,
		SessionID:        sessionID,
		TradingPair:      tradingPair,
		Direction:        direction,
		Size:             sizeUSD,
		EntryPrice:       currentPrice,
		Fees:             fees,
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

	// Deduct balance
	newBalance := balance - sizeUSD - fees
	if err := s.BalanceRepo.UpdateBalance(userID, newBalance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
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
