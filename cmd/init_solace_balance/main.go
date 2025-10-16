package main

import (
	"ares_api/internal/database"
	"ares_api/internal/models"
	"fmt"
	"log"
)

func main() {
	// Initialize database
	db := database.InitDB()

	// Create trading balance for solace_ai (user_id = 8)
	balance := models.Balance{
		UserID:           8,
		Asset:            "USDC",
		Amount:           10000.00,
		AutoTopup:        false,
		TopupThreshold:   1000.00,
		TopupAmount:      5000.00,
		TotalDeposits:    10000.00,
		TotalWithdrawals: 0.00,
		RealizedPnL:      0.00,
		UnrealizedPnL:    0.00,
	}

	// Upsert (insert or update)
	result := db.Where("user_id = ? AND asset = ?", 8, "USDC").Assign(balance).FirstOrCreate(&balance)
	if result.Error != nil {
		log.Fatalf("Failed to create/update balance: %v", result.Error)
	}

	fmt.Printf("✅ Trading balance initialized for solace_ai (user_id=8)\n")
	fmt.Printf("   Balance: $%.2f %s\n", balance.Amount, balance.Asset)
	fmt.Printf("   Auto top-up: %v\n", balance.AutoTopup)
}
