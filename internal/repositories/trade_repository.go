/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package repositories

import (
	"fmt"
	repo "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"time"

	"gorm.io/gorm"
)

type TradeRepository struct {
	db *gorm.DB
}

func NewTradeRepository(db *gorm.DB) repo.TradeRepository {
	return &TradeRepository{db: db}
}

func (r *TradeRepository) Create(trade *models.Trade) error {
	// Use atomic transaction to ensure balance + trade creation succeed together
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the trade record
		if err := tx.Create(trade).Error; err != nil {
			return err
		}

		// Update user's virtual balance atomically
		// Deduct cost for buy orders
		if trade.Side == "buy" {
			totalCost := (trade.Amount * trade.Price) * 1.001 // Include 0.1% fee
			result := tx.Exec(
				"UPDATE users SET virtual_balance = virtual_balance - ? WHERE id = ? AND virtual_balance >= ?",
				totalCost, trade.UserID, totalCost,
			)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("insufficient balance or user not found")
			}
		}

		return nil
	})
}

func (r *TradeRepository) GetByUserID(userID uint, limit int) ([]models.Trade, error) {
	var trades []models.Trade
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&trades).Error
	return trades, err
}

func (r *TradeRepository) GetOpenLimitOrders() ([]models.Trade, error) {
	var trades []models.Trade
	err := r.db.Where("type = ? AND status = ?", "limit", "open").Find(&trades).Error
	return trades, err
}

func (r *TradeRepository) MarkOrderFilled(tradeID uint) error {
	return r.db.Model(&models.Trade{}).Where("id = ?", tradeID).Updates(map[string]interface{}{
		"status":     "filled",
		"updated_at": time.Now(),
	}).Error
}

func (r *TradeRepository) GetOpenLimitOrdersByUser(userID uint) ([]models.Trade, error) {
	var trades []models.Trade
	err := r.db.Where("user_id = ? AND type = ? AND status = ?", userID, "limit", "open").Find(&trades).Error
	return trades, err
}

// Update updates a trade (for closing positions) with atomic balance update
func (r *TradeRepository) Update(trade *models.Trade) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update the trade
		if err := tx.Save(trade).Error; err != nil {
			return err
		}

		// If closing a position, update balance
		if trade.Status == "closed" && trade.ExitPrice != nil {
			// For buy orders: return principal + profit/loss
			if trade.Side == "buy" {
				proceeds := *trade.ExitPrice * trade.Amount
				closingFee := proceeds * 0.001 // 0.1% closing fee
				netProceeds := proceeds - closingFee
				
				result := tx.Exec(
					"UPDATE users SET virtual_balance = virtual_balance + ? WHERE id = ?",
					netProceeds, trade.UserID,
				)
				if result.Error != nil {
					return result.Error
				}
			}
		}

		return nil
	})
}

// FindByID finds a trade by its string ID (for sandbox trades)
func (r *TradeRepository) FindByID(id string) (*models.Trade, error) {
	var trade models.Trade
	if err := r.db.Where("trade_id = ?", id).First(&trade).Error; err != nil {
		return nil, err
	}
	return &trade, nil
}

// FindOpenByUserID finds all open trades for a user
func (r *TradeRepository) FindOpenByUserID(userID uint) ([]models.Trade, error) {
	var trades []models.Trade
	if err := r.db.Where("user_id = ? AND status = ?", userID, "open").
		Order("created_at DESC").
		Find(&trades).Error; err != nil {
		return nil, err
	}
	return trades, nil
}

// GetUserBalance gets user's current virtual balance
func (r *TradeRepository) GetUserBalance(userID uint) (float64, error) {
	var user models.User
	if err := r.db.Select("virtual_balance").Where("id = ?", userID).First(&user).Error; err != nil {
		return 0, err
	}
	return user.VirtualBalance, nil
}

// CountClosedTrades counts closed trades for metrics
func (r *TradeRepository) CountClosedTrades(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Trade{}).
		Where("user_id = ? AND status = ?", userID, "closed").
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
