package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ares_api/internal/eventbus"
)

// TestEventBus demonstrates EventBus functionality
func main() {
	// Create event bus
	eb := eventbus.NewEventBus()
	defer eb.Close()

	eventReceived := make(chan bool, 1)

	// Subscribe to trade_executed events
	eb.Subscribe(eventbus.EventTypeTradeExecuted, func(data []byte) {
		var event eventbus.TradeExecutedEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			return
		}

		fmt.Printf("\n🎯 TRADE EXECUTED EVENT RECEIVED:\n")
		fmt.Printf("   Trade ID: %d\n", event.Data.TradeID)
		fmt.Printf("   Symbol: %s\n", event.Data.Symbol)
		fmt.Printf("   Side: %s\n", event.Data.Side)
		fmt.Printf("   Amount: $%.2f\n", event.Data.Amount)
		fmt.Printf("   Price: $%.2f\n", event.Data.Price)
		fmt.Printf("   Status: %s\n", event.Data.Status)
		fmt.Printf("   Timestamp: %s\n\n", event.Timestamp.Format("2006-01-02 15:04:05"))

		eventReceived <- true
	})

	// Publish test event
	testEvent := eventbus.NewTradeExecutedEvent(
		12345,
		"BTC/USD",
		"BUY",
		1000.00,
		50000.00,
		"2025-01-16T09:48:00Z",
		"sandbox",
		"OPEN",
		125,
	)

	fmt.Println("📤 Publishing test event...")
	if err := eb.Publish(eventbus.EventTypeTradeExecuted, testEvent); err != nil {
		log.Fatalf("Failed to publish event: %v", err)
	}

	fmt.Println("✅ Event published successfully!")
	fmt.Println("\n📊 EventBus Health:")
	health := eb.Health()
	healthJSON, _ := json.MarshalIndent(health, "", "  ")
	fmt.Println(string(healthJSON))

	// Wait for event to be received (with timeout)
	select {
	case <-eventReceived:
		fmt.Println("\n✅ Test completed successfully! EventBus is working.")
	case <-time.After(2 * time.Second):
		fmt.Println("\n⚠️ Timeout waiting for event")
	}
}
