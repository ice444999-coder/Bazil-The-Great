/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package Repositories
import "ares_api/internal/models"

type BalanceRepository interface {
	// Legacy methods
	GetUSDBalanceModel(userID uint) (*models.Balance, error)  // Returns full model
	UpdateUSDBalance(userID uint, delta float64) (*models.Balance, error)
	ResetUSDBalance(userID uint, defaultBalance float64) error
	CreateUSDBalance(userID uint, defaultBalance float64) (*models.Balance, error)

	// New trading methods
	GetUSDBalance(userID uint) (float64, error)               // Returns just amount
	GetBalanceRecord(userID uint) (*models.Balance, error)    // Get full record with auto-topup settings
	UpdateBalance(userID uint, newAmount float64) error       // Update balance amount
	UpdateBalanceRecord(balance *models.Balance) error        // Save entire record
}

