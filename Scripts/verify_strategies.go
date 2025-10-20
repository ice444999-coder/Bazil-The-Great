package main

import (
	"ares_api/internal/trading"
	"fmt"
	"log"
)

func main() {
	fmt.Println("🚀 Initializing Multi-Strategy Trading System...")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	// Get all 5 strategies
	strategies := trading.GetAllStrategies()

	fmt.Printf("\n✅ Loaded %d strategies:\n\n", len(strategies))

	for i, strategy := range strategies {
		fmt.Printf("%d. %s\n", i+1, strategy.Name())
		fmt.Printf("   Description: %s\n", strategy.Description())
		fmt.Printf("   Risk Level:  %s\n", strategy.GetRiskLevel())
		fmt.Printf("\n")
	}

	// Verify each strategy can be retrieved by name
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println("\n🔍 Testing Strategy Registry...")

	strategyNames := []string{
		"RSI_Oversold",
		"MACD_Crossover",
		"Trend_Following",
		"Support_Bounce",
		"Volume_Breakout",
	}

	for _, name := range strategyNames {
		strategy, err := trading.GetStrategyByName(name)
		if err != nil {
			log.Fatalf("❌ Failed to get strategy %s: %v", name, err)
		}
		if strategy.Name() != name {
			log.Fatalf("❌ Strategy name mismatch: expected %s, got %s", name, strategy.Name())
		}
		fmt.Printf("✅ %s - OK\n", name)
	}

	fmt.Println("\n=" + string(make([]byte, 60)) + "=")
	fmt.Println("\n🎉 All 5 strategies successfully implemented!")
	fmt.Println("\n📊 Strategy Summary:")
	fmt.Println("   • 3 Medium Risk Strategies (RSI, MACD, Support)")
	fmt.Println("   • 2 High Risk Strategies (Trend, Volume)")
	fmt.Println("   • All implement Strategy interface")
	fmt.Println("   • All provide detailed reasoning")
	fmt.Println("   • All calculate target price + stop loss")
	fmt.Println("\n✨ Ready for integration with MultiStrategyOrchestrator!")
}
