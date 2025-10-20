package repositories

import (
	"ares_api/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TradingRepository struct {
	DB *gorm.DB
}

func NewTradingRepository(db *gorm.DB) *TradingRepository {
	return &TradingRepository{DB: db}
}

// SaveTrade creates a new sandbox trade
func (r *TradingRepository) SaveTrade(trade *models.SandboxTrade) error {
	return r.DB.Create(trade).Error
}

// GetTradeByID retrieves a trade by ID
func (r *TradingRepository) GetTradeByID(id uint) (*models.SandboxTrade, error) {
	var trade models.SandboxTrade
	err := r.DB.First(&trade, id).Error
	return &trade, err
}

// GetOpenTrades gets all open trades for a user
func (r *TradingRepository) GetOpenTrades(userID uint) ([]models.SandboxTrade, error) {
	var trades []models.SandboxTrade
	err := r.DB.Where("user_id = ? AND status = ?", userID, "OPEN").
		Order("opened_at DESC").
		Find(&trades).Error
	return trades, err
}

// GetTradeHistory gets all trades for a user with pagination
func (r *TradingRepository) GetTradeHistory(userID uint, limit int, offset int) ([]models.SandboxTrade, error) {
	var trades []models.SandboxTrade
	err := r.DB.Where("user_id = ?", userID).
		Order("opened_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&trades).Error
	return trades, err
}

// GetTradesBySession gets all trades for a session
func (r *TradingRepository) GetTradesBySession(sessionID uuid.UUID) ([]models.SandboxTrade, error) {
	var trades []models.SandboxTrade
	err := r.DB.Where("session_id = ?", sessionID).
		Order("opened_at DESC").
		Find(&trades).Error
	return trades, err
}

// GetTradesByStrategy gets all trades for a user filtered by strategy name
func (r *TradingRepository) GetTradesByStrategy(userID uint, strategyName string) ([]models.SandboxTrade, error) {
	var trades []models.SandboxTrade
	err := r.DB.Where("user_id = ? AND strategy_name = ?", userID, strategyName).
		Order("opened_at DESC").
		Find(&trades).Error
	return trades, err
}

// GetAllTrades gets all trades for a user (used for aggregated metrics)
func (r *TradingRepository) GetAllTrades(userID uint) ([]models.SandboxTrade, error) {
	var trades []models.SandboxTrade
	err := r.DB.Where("user_id = ?", userID).
		Order("opened_at DESC").
		Find(&trades).Error
	return trades, err
}

// UpdateTrade updates an existing trade
func (r *TradingRepository) UpdateTrade(trade *models.SandboxTrade) error {
	return r.DB.Save(trade).Error
}

// CloseTrade closes an open trade with exit price and P&L
func (r *TradingRepository) CloseTrade(tradeID uint, exitPrice float64) error {
	var trade models.SandboxTrade
	if err := r.DB.First(&trade, tradeID).Error; err != nil {
		return err
	}

	now := time.Now()
	trade.ExitPrice = &exitPrice
	trade.ClosedAt = &now
	trade.Status = "CLOSED"

	// Calculate P&L
	var profitLoss float64
	if trade.Direction == "BUY" {
		profitLoss = (exitPrice - trade.EntryPrice) * trade.Size / trade.EntryPrice
	} else { // SELL
		profitLoss = (trade.EntryPrice - exitPrice) * trade.Size / trade.EntryPrice
	}

	profitLoss -= trade.Fees
	trade.ProfitLoss = &profitLoss

	// Calculate P&L percentage
	plPercent := (profitLoss / trade.Size) * 100
	trade.ProfitLossPercent = &plPercent

	return r.DB.Save(&trade).Error
}

// GetPerformanceMetrics calculates trading performance for a user
func (r *TradingRepository) GetPerformanceMetrics(userID uint) (*models.TradingPerformance, error) {
	var closedTrades []models.SandboxTrade
	err := r.DB.Where("user_id = ? AND status = ?", userID, "CLOSED").
		Find(&closedTrades).Error
	if err != nil {
		return nil, err
	}

	if len(closedTrades) == 0 {
		return &models.TradingPerformance{
			UserID:       userID,
			CalculatedAt: time.Now(),
		}, nil
	}

	perf := &models.TradingPerformance{
		UserID:          userID,
		CalculatedAt:    time.Now(),
		TotalTrades:     len(closedTrades),
		StrategyVersion: 1,
	}

	var totalPnL float64
	var totalProfit float64
	var totalLoss float64
	var largestWin float64
	var largestLoss float64
	winningTrades := 0
	losingTrades := 0

	for _, trade := range closedTrades {
		if trade.ProfitLoss != nil {
			pnl := *trade.ProfitLoss
			totalPnL += pnl

			if pnl > 0 {
				winningTrades++
				totalProfit += pnl
				if pnl > largestWin {
					largestWin = pnl
				}
			} else {
				losingTrades++
				totalLoss += pnl
				if pnl < largestLoss {
					largestLoss = pnl
				}
			}
		}
	}

	perf.WinningTrades = winningTrades
	perf.LosingTrades = losingTrades
	perf.TotalProfitLoss = &totalPnL

	if winningTrades > 0 {
		avgProfit := totalProfit / float64(winningTrades)
		perf.AvgProfit = &avgProfit
		perf.LargestWin = &largestWin
	}

	if losingTrades > 0 {
		avgLoss := totalLoss / float64(losingTrades)
		perf.AvgLoss = &avgLoss
		perf.LargestLoss = &largestLoss
	}

	if perf.TotalTrades > 0 {
		winRate := (float64(winningTrades) / float64(perf.TotalTrades)) * 100
		perf.WinRate = &winRate
	}

	return perf, nil
}

// SavePerformance saves performance metrics
func (r *TradingRepository) SavePerformance(perf *models.TradingPerformance) error {
	return r.DB.Create(perf).Error
}

// SaveMarketData saves market data to cache
func (r *TradingRepository) SaveMarketData(data *models.MarketDataCache) error {
	return r.DB.Create(data).Error
}

// GetLatestMarketData gets the most recent market data for a symbol
func (r *TradingRepository) GetLatestMarketData(symbol string) (*models.MarketDataCache, error) {
	var data models.MarketDataCache
	err := r.DB.Where("symbol = ?", symbol).
		Order("timestamp DESC").
		First(&data).Error
	return &data, err
}

// SaveStrategyMutation saves a strategy mutation
func (r *TradingRepository) SaveStrategyMutation(mutation *models.StrategyMutation) error {
	return r.DB.Create(mutation).Error
}

// GetLatestStrategy gets the latest deployed strategy
func (r *TradingRepository) GetLatestStrategy(userID uint) (*models.StrategyMutation, error) {
	var strategy models.StrategyMutation
	err := r.DB.Where("user_id = ? AND status = ?", userID, "DEPLOYED").
		Order("deployed_at DESC").
		First(&strategy).Error
	return &strategy, err
}

// SaveRiskEvent logs a risk event
func (r *TradingRepository) SaveRiskEvent(event *models.RiskEvent) error {
	return r.DB.Create(event).Error
}

// GetRecentRiskEvents gets recent risk events
func (r *TradingRepository) GetRecentRiskEvents(userID uint, limit int) ([]models.RiskEvent, error) {
	var events []models.RiskEvent
	err := r.DB.Where("user_id = ?", userID).
		Order("detected_at DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}
