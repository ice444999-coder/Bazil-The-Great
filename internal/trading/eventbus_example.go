package trading

import (
	"ares_api/internal/eventbus"
	"encoding/json"
	"fmt"
	"log"
)

// ========================================================================
// EVENTBUS SUBSCRIPTION EXAMPLE
// ========================================================================
// This demonstrates how the master dashboard subscribes to strategy events
// Real-time updates flow: Strategy → EventBus → Dashboard → WebSocket → UI
// ========================================================================

// DashboardSubscriber - Master dashboard subscribes to all strategy events
type DashboardSubscriber struct {
	eventBus *eventbus.EventBus
}

// NewDashboardSubscriber creates a new subscriber
func NewDashboardSubscriber(eb *eventbus.EventBus) *DashboardSubscriber {
	return &DashboardSubscriber{
		eventBus: eb,
	}
}

// SubscribeToAllStrategies - Subscribe to all 5 strategies
func (ds *DashboardSubscriber) SubscribeToAllStrategies() {
	strategies := []string{
		"RSI_Oversold",
		"MACD_Crossover",
		"Trend_Following",
		"Support_Bounce",
		"Volume_Breakout",
	}

	// Subscribe to signal events
	for _, strategyName := range strategies {
		topic := fmt.Sprintf("strategy.%s.signal", strategyName)
		ds.eventBus.Subscribe(topic, ds.handleSignalEvent)
		log.Printf("[DASHBOARD] Subscribed to %s", topic)
	}

	// Subscribe to metrics events
	for _, strategyName := range strategies {
		topic := fmt.Sprintf("strategy.%s.metrics", strategyName)
		ds.eventBus.Subscribe(topic, ds.handleMetricsEvent)
		log.Printf("[DASHBOARD] Subscribed to %s", topic)
	}

	// Subscribe to master metrics
	ds.eventBus.Subscribe("strategy.master.metrics", ds.handleMasterMetrics)
	log.Printf("[DASHBOARD] Subscribed to strategy.master.metrics")

	// Subscribe to strategy lifecycle events
	ds.eventBus.Subscribe("strategy.registered", ds.handleLifecycleEvent)
	ds.eventBus.Subscribe("strategy.unregistered", ds.handleLifecycleEvent)
	ds.eventBus.Subscribe("strategy.toggled", ds.handleLifecycleEvent)

	log.Println("[DASHBOARD] All event subscriptions initialized")
}

// handleSignalEvent - Process trade signals from strategies
func (ds *DashboardSubscriber) handleSignalEvent(data []byte) {
	var signal map[string]interface{}
	if err := json.Unmarshal(data, &signal); err != nil {
		log.Printf("[ERROR] Failed to unmarshal signal event: %v", err)
		return
	}

	jsonData, _ := json.MarshalIndent(signal, "", "  ")
	log.Printf("[SIGNAL EVENT] %s", string(jsonData))

	// In production, this would:
	// 1. Update dashboard metrics in real-time
	// 2. Push to WebSocket clients
	// 3. Trigger notifications if high-confidence signal
	// 4. Update UI charts and trade feed
	// 5. Log to database for historical analysis
}

// handleMetricsEvent - Process strategy metrics updates
func (ds *DashboardSubscriber) handleMetricsEvent(data []byte) {
	var metrics map[string]interface{}
	if err := json.Unmarshal(data, &metrics); err != nil {
		log.Printf("[ERROR] Failed to unmarshal metrics event: %v", err)
		return
	}

	jsonData, _ := json.MarshalIndent(metrics, "", "  ")
	log.Printf("[METRICS EVENT] %s", string(jsonData))

	// In production:
	// 1. Update performance cards in UI
	// 2. Recalculate leaderboard rankings
	// 3. Check auto-graduate criteria
	// 4. Update 3D P&L visualization
	// 5. Trigger alerts if metrics degrade
}

// handleMasterMetrics - Process aggregated master metrics
func (ds *DashboardSubscriber) handleMasterMetrics(data []byte) {
	var master map[string]interface{}
	if err := json.Unmarshal(data, &master); err != nil {
		log.Printf("[ERROR] Failed to unmarshal master metrics: %v", err)
		return
	}

	jsonData, _ := json.MarshalIndent(master, "", "  ")
	log.Printf("[MASTER METRICS] %s", string(jsonData))

	// In production:
	// 1. Update master dashboard header
	// 2. Show overall system health
	// 3. Highlight best/worst strategies
	// 4. Calculate portfolio allocation
	// 5. Update total P&L display
}

// handleLifecycleEvent - Process strategy lifecycle changes
func (ds *DashboardSubscriber) handleLifecycleEvent(data []byte) {
	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("[ERROR] Failed to unmarshal lifecycle event: %v", err)
		return
	}

	jsonData, _ := json.MarshalIndent(event, "", "  ")
	log.Printf("[LIFECYCLE EVENT] %s", string(jsonData))

	// In production:
	// 1. Update strategy list in UI
	// 2. Show enable/disable state changes
	// 3. Notify users of hot-swaps
	// 4. Refresh tab navigation
	// 5. Update system status indicators
}

// ========================================================================
// EXAMPLE USAGE
// ========================================================================
// func main() {
//     eb := eventbus.NewEventBus()
//     orchestrator := NewMultiStrategyOrchestrator(db, eb)
//     dashboard := NewDashboardSubscriber(eb)
//
//     // Subscribe to all events
//     dashboard.SubscribeToAllStrategies()
//
//     // Register strategies (they will auto-publish events)
//     orchestrator.RegisterStrategy(NewRSIOversoldStrategy(eb), &StrategyConfig{Enabled: true})
//     orchestrator.RegisterStrategy(NewMACDCrossoverStrategy(eb), &StrategyConfig{Enabled: true})
//
//     // Execute strategies (signals published automatically)
//     decisions := orchestrator.ExecuteAll(marketData, history)
//
//     // Publish metrics manually when calculated
//     orchestrator.PublishStrategyMetrics(metrics)
//     orchestrator.PublishMasterMetrics(masterMetrics)
// }
// ========================================================================
