/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package services

import (
	"ares_api/internal/api/dto"
	repository "ares_api/internal/interfaces/repository"
	service "ares_api/internal/interfaces/service"
	"ares_api/internal/models"
	"fmt"
	"time"
)

var _ service.TradeService = &TradeService{}

type TradeService struct {
	Repo        repository.TradeRepository
	BalanceRepo repository.BalanceRepository
	AssetRepo   repository.AssetRepository
}

func NewTradeService(r repository.TradeRepository, b repository.BalanceRepository, a repository.AssetRepository) *TradeService {
	return &TradeService{
		Repo:        r,
		BalanceRepo: b,
		AssetRepo:   a,
	}
}

// MarketOrder executes immediately and updates USD balance
func (s *TradeService) MarketOrder(userID uint, req dto.MarketOrderRequest) (*dto.TradeResponse, error) {
	// Always transact in USD
	const baseCurrency = "usd"

	// Fetch current price from CoinGecko
	coinMarket, err := s.AssetRepo.FetchCoinMarket(req.CoinID, baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market price: %w", err)
	}
	price := coinMarket.PriceUSD
	cost := req.Quantity * price

	// Get user USD balance
	balance, err := s.BalanceRepo.GetUSDBalanceModel(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get USD balance: %w", err)
	}

	// Check balance
	if req.Side == "buy" && balance.Amount < cost {
		return nil, fmt.Errorf("insufficient USD balance")
	}
	if req.Side == "sell" {
		// For demo: we only check if user has coins hypothetically
		// In real-world, you’d maintain user holdings per coin
		// Here, assume user can always sell (or extend with holdings later)
	}

	// Update USD balance
	switch req.Side {
	case "buy":
		// Subtract cost
		if _, err := s.BalanceRepo.UpdateUSDBalance(userID, -cost); err != nil {
			return nil, err
		}
	case "sell":
		// Add proceeds
		if _, err := s.BalanceRepo.UpdateUSDBalance(userID, cost); err != nil {
			return nil, err
		}
	}

	// Record the trade
	trade := &models.Trade{
		UserID:   userID,
		CoinID:  req.CoinID,
		Symbol:   req.Symbol,
		Side:     req.Side,
		Quantity: req.Quantity,
		Price:    price,
		Type:     "market",
		Status:   "filled",
	}
	if err := s.Repo.Create(trade); err != nil {
		return nil, err
	}

	return &dto.TradeResponse{
		ID:        trade.ID,
		UserID:    trade.UserID,
		CoinID:    trade.CoinID,
		Symbol:    trade.Symbol,
		Side:      trade.Side,
		Quantity:  trade.Quantity,
		Price:     trade.Price,
		Type:      trade.Type,
		Status:    trade.Status,
		CreatedAt: trade.CreatedAt.Format(time.RFC3339),
		UpdatedAt: trade.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// LimitOrder places a conditional order
func (s *TradeService) LimitOrder(userID uint, req dto.LimitOrderRequest) (*dto.TradeResponse, error) {
	const baseCurrency = "usd"

	// Default status
	status := "open"

	// Fetch current market price
	coinMarket, err := s.AssetRepo.FetchCoinMarket(req.CoinID, baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market price: %w", err)
	}
	currentPrice := coinMarket.PriceUSD

	// Immediate execution if limit condition met
	if (req.Side == "buy" && currentPrice <= req.LimitPrice) ||
		(req.Side == "sell" && currentPrice >= req.LimitPrice) {

		status = "filled"

		// Execute as market order
		_, err := s.MarketOrder(userID, dto.MarketOrderRequest{
			CoinID:   req.CoinID,
			Symbol:   req.Symbol,
			Side:     req.Side,
			Quantity: req.Quantity,
			Currency: baseCurrency,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to execute market order: %w", err)
		}
	}

	// Record limit order in DB
	trade := &models.Trade{
		UserID:   userID,
		CoinID:  req.CoinID,
		Symbol:   req.Symbol,
		Side:     req.Side,
		Quantity: req.Quantity,
		Price:    req.LimitPrice,
		Type:     "limit",
		Status:   status,
	}

	if err := s.Repo.Create(trade); err != nil {
		return nil, fmt.Errorf("failed to create limit order: %w", err)
	}

	return &dto.TradeResponse{
		ID:        trade.ID,
		UserID:    trade.UserID,
		CoinID:    trade.CoinID,
		Symbol:    trade.Symbol,
		Side:      trade.Side,
		Quantity:  trade.Quantity,
		Price:     trade.Price,
		Type:      trade.Type,
		Status:    trade.Status,
		CreatedAt: trade.CreatedAt.Format(time.RFC3339),
		UpdatedAt: trade.UpdatedAt.Format(time.RFC3339),
	}, nil
}


// GetHistory returns last N trades for a user
func (s *TradeService) GetHistory(userID uint, limit int) ([]dto.TradeResponse, error) {
	trades, err := s.Repo.GetByUserID(userID, limit)
	if err != nil {
		return nil, err
	}

	var responses []dto.TradeResponse
	for _, t := range trades {
		responses = append(responses, dto.TradeResponse{
			ID:        t.ID,
			UserID:    t.UserID,
			CoinID:    t.CoinID,
			Symbol:    t.Symbol,
			Side:      t.Side,
			Quantity:  t.Quantity,
			Price:     t.Price,
			Type:      t.Type,
			Status:    t.Status,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
			UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
		})
	}
	return responses, nil
}


func (s *TradeService) ProcessOpenLimitOrders() {
	const baseCurrency = "usd"

	// Fetch all open limit orders
	openOrders, err := s.Repo.GetOpenLimitOrders()
	if err != nil {
		fmt.Println("Error fetching open limit orders:", err)
		return
	}

	for _, order := range openOrders {
		coinMarket, err := s.AssetRepo.FetchCoinMarket(order.CoinID, baseCurrency)
		if err != nil {
			continue // skip if coin data not available
		}

		currentPrice := coinMarket.PriceUSD

		// Check if limit condition is met
		if (order.Side == "buy" && currentPrice <= order.Price) ||
			(order.Side == "sell" && currentPrice >= order.Price) {

			// Execute as market order
			_, err := s.MarketOrder(order.UserID, dto.MarketOrderRequest{
				CoinID:   order.CoinID,
				Symbol:   order.Symbol,
				Side:     order.Side,
				Quantity: order.Quantity,
				Currency: baseCurrency,
			})
			if err == nil {
				s.Repo.MarkOrderFilled(order.ID)
			}
		}
	}
}


func (s *TradeService) GetPendingLimitOrders(userID uint) ([]dto.TradeResponse, error) {
	trades, err := s.Repo.GetOpenLimitOrdersByUser(userID)
	if err != nil {
		return nil, err
	}

	var responses []dto.TradeResponse
	for _, t := range trades {
		responses = append(responses, dto.TradeResponse{
			ID:        t.ID,
			UserID:    t.UserID,
			CoinID:    t.CoinID,
			Symbol:    t.Symbol,
			Side:      t.Side,
			Quantity:  t.Quantity,
			Price:     t.Price,
			Type:      t.Type,
			Status:    t.Status,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
			UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
		})
	}
	return responses, nil
}
