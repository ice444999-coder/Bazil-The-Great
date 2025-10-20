package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"ares_api/config"
	"ares_api/internal/eventbus"
	"ares_api/internal/trading"
)

// Test EventBus integration with all 5 trading strategies
func main() {
	fmt.Println("🧪 ARES EventBus Strategy Validation Test")
	fmt.Println("==========================================")
	fmt.Println()

	// Initialize config and database
	cfg := config.Load()
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize EventBus
	eb := eventbus.NewEventBus()
	fmt.Println("✅ EventBus initialized")

	// Test Topic: Subscribe to all strategy topics
	topics := []string{
		"strategy.RSIOversold.signal",
		"strategy.RSIOversold.metrics",
		"strategy.MACDCrossover.signal",
		"strategy.MACDCrossover.metrics",
		"strategy.TrendFollowing.signal",
		"strategy.TrendFollowing.metrics",
		"strategy.SupportBounce.signal",
		"strategy.SupportBounce.metrics",
		"strategy.VolumeBreakout.signal",
		"strategy.VolumeBreakout.metrics",
		"strategy.master.metrics",
	}

	// Create subscribers for each topic
	receivedEvents := make(map[string]int)
	for _, topic := range topics {
		t := topic // Capture loop variable
		eb.Subscribe(t, func(data []byte) {
			receivedEvents[t]++
			fmt.Printf("📨 Received event on '%s': %s\n", t, string(data))
		})
		fmt.Printf("👂 Subscribed to: %s\n", t)
	}

	fmt.Println()
	fmt.Println("🚀 Initializing trading strategies...")

	// Initialize all 5 strategies
	strategies := []trading.Strategy{
		trading.NewRSIOversoldStrategy(eb),
		trading.NewMACDCrossoverStrategy(eb),
		trading.NewTrendFollowingStrategy(eb),
		trading.NewSupportBounceStrategy(eb),
		trading.NewVolumeBreakoutStrategy(eb),
	}

	// Create test market data
	testData := []trading.MarketData{
		{
			Symbol:    "BTCUSDT",
			Price:     65000.0,
			High24h:   66000.0,
			Low24h:    64000.0,
			Volume24h: 15000.0,
			Timestamp: time.Now(),
			RSI:       25.0, // Oversold
			MACD:      -50.0,
			Signal:    -45.0,
			Histogram: -5.0,
		},
		{
			Symbol:    "ETHUSDT",
			Price:     3200.0,
			High24h:   3250.0,
			Low24h:    3150.0,
			Volume24h: 8000.0,
			Timestamp: time.Now(),
			RSI:       45.0,
			MACD:      10.0, // Bullish crossover
			Signal:    5.0,
			Histogram: 5.0,
		},
		{
			Symbol:    "SOLUSDT",
			Price:     150.0,
			High24h:   155.0,
			Low24h:    145.0,
			Volume24h: 5000.0,
			Timestamp: time.Now(),
			RSI:       65.0,
			MACD:      20.0,
			Signal:    18.0,
			Histogram: 2.0,
		},
	}

	fmt.Println("📊 Testing strategies with market data...")
	fmt.Println()

	// Run each strategy with test data
	ctx := context.Background()
	for i, strategy := range strategies {
		fmt.Printf("\n🔄 Testing Strategy %d: %s\n", i+1, getStrategyName(strategy))
		fmt.Println(strings.Repeat("-", 50))

		for _, data := range testData {
			signal, err := strategy.Analyze(ctx, data)
			if err != nil {
				fmt.Printf("❌ Error analyzing %s: %v\n", data.Symbol, err)
				continue
			}

			if signal != nil && signal.Action != "HOLD" {
				fmt.Printf("✅ Signal generated for %s: %s (Confidence: %.2f)\n",
					data.Symbol, signal.Action, signal.Confidence)
			} else {
				fmt.Printf("⏸️  No signal for %s (HOLD)\n", data.Symbol)
			}
		}

		// Get and publish metrics
		metrics := strategy.GetMetrics()
		fmt.Printf("📊 Metrics: Analyzed=%d, Signals=%d, Win Rate=%.2f%%\n",
			metrics.TotalAnalyzed, metrics.SignalsGenerated, metrics.WinRate*100)
	}

	// Wait for events to be processed
	fmt.Println()
	fmt.Println("⏳ Waiting for EventBus to process events...")
	time.Sleep(2 * time.Second)

	// Print results
	fmt.Println()
	fmt.Println("📊 EventBus Validation Results")
	fmt.Println("================================")

	allTopicsReceived := true
	totalEvents := 0

	for _, topic := range topics {
		count := receivedEvents[topic]
		totalEvents += count
		status := "❌"
		if count > 0 {
			status = "✅"
		} else {
			allTopicsReceived = false
		}
		fmt.Printf("%s %s: %d events\n", status, topic, count)
	}

	fmt.Println()
	fmt.Printf("📈 Total events received: %d\n", totalEvents)

	if allTopicsReceived && totalEvents > 0 {
		fmt.Println()
		fmt.Println("✅ ✅ ✅ EVENTBUS VALIDATION PASSED! ✅ ✅ ✅")
		fmt.Println("All strategies are properly wired to EventBus")
	} else {
		fmt.Println()
		fmt.Println("❌ ❌ ❌ EVENTBUS VALIDATION FAILED! ❌ ❌ ❌")
		fmt.Println("Some topics did not receive events")
	}

	// Cleanup
	eb.Close()
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

func getStrategyName(s trading.Strategy) string {
	switch s.(type) {
	case *trading.RSIOversoldStrategy:
		return "RSIOversoldStrategy"
	case *trading.MACDCrossoverStrategy:
		return "MACDCrossoverStrategy"
	case *trading.TrendFollowingStrategy:
		return "TrendFollowingStrategy"
	case *trading.SupportBounceStrategy:
		return "SupportBounceStrategy"
	case *trading.VolumeBreakoutStrategy:
		return "VolumeBreakoutStrategy"
	default:
		return "Unknown Strategy"
	}
}
