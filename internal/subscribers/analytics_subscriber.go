package subscribers

import (
	"ares_api/internal/eventbus"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// TradeAnalytics stores real-time trading metrics
type TradeAnalytics struct {
	mu                 sync.RWMutex
	TotalTrades        int64
	TotalVolume        float64
	BuyCount           int64
	SellCount          int64
	AverageExecutionMS float64
	LastTradeTimestamp time.Time
	TradesPerMinute    float64
	PairVolumes        map[string]float64
	lastMinuteTrades   []time.Time
}

// AnalyticsSubscriber tracks trading analytics in real-time
type AnalyticsSubscriber struct {
	analytics *TradeAnalytics
}

// NewAnalyticsSubscriber creates a new analytics subscriber
func NewAnalyticsSubscriber() *AnalyticsSubscriber {
	return &AnalyticsSubscriber{
		analytics: &TradeAnalytics{
			PairVolumes:      make(map[string]float64),
			lastMinuteTrades: make([]time.Time, 0),
		},
	}
}

// HandleTradeExecuted processes trade_executed events for analytics
func (s *AnalyticsSubscriber) HandleTradeExecuted(data []byte) {
	var event eventbus.TradeExecutedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("[ANALYTICS][ERROR] Failed to unmarshal trade event: %v", err)
		return
	}

	s.analytics.mu.Lock()
	defer s.analytics.mu.Unlock()

	// Update counters
	s.analytics.TotalTrades++
	s.analytics.TotalVolume += event.Data.Amount
	s.analytics.LastTradeTimestamp = event.Timestamp

	// Track direction
	if event.Data.Side == "BUY" {
		s.analytics.BuyCount++
	} else {
		s.analytics.SellCount++
	}

	// Update average execution time
	if s.analytics.TotalTrades == 1 {
		s.analytics.AverageExecutionMS = float64(event.Data.ExecutionTime)
	} else {
		// Exponential moving average
		s.analytics.AverageExecutionMS = (s.analytics.AverageExecutionMS*0.9 + float64(event.Data.ExecutionTime)*0.1)
	}

	// Track volume by pair
	s.analytics.PairVolumes[event.Data.Symbol] += event.Data.Amount

	// Track trades per minute (keep last 60 seconds)
	now := time.Now()
	s.analytics.lastMinuteTrades = append(s.analytics.lastMinuteTrades, now)

	// Remove trades older than 1 minute
	cutoff := now.Add(-1 * time.Minute)
	validTrades := make([]time.Time, 0)
	for _, t := range s.analytics.lastMinuteTrades {
		if t.After(cutoff) {
			validTrades = append(validTrades, t)
		}
	}
	s.analytics.lastMinuteTrades = validTrades
	s.analytics.TradesPerMinute = float64(len(validTrades))

	log.Printf("[ANALYTICS][UPDATE] Total: %d trades | Volume: $%.2f | TPM: %.1f | Avg Exec: %.0fms",
		s.analytics.TotalTrades, s.analytics.TotalVolume, s.analytics.TradesPerMinute, s.analytics.AverageExecutionMS)
}

// GetStats returns current analytics (thread-safe)
func (s *AnalyticsSubscriber) GetStats() map[string]interface{} {
	s.analytics.mu.RLock()
	defer s.analytics.mu.RUnlock()

	pairVolumes := make(map[string]float64)
	for k, v := range s.analytics.PairVolumes {
		pairVolumes[k] = v
	}

	return map[string]interface{}{
		"total_trades":         s.analytics.TotalTrades,
		"total_volume":         s.analytics.TotalVolume,
		"buy_count":            s.analytics.BuyCount,
		"sell_count":           s.analytics.SellCount,
		"average_execution_ms": s.analytics.AverageExecutionMS,
		"last_trade":           s.analytics.LastTradeTimestamp.Format(time.RFC3339),
		"trades_per_minute":    s.analytics.TradesPerMinute,
		"pair_volumes":         pairVolumes,
	}
}

// Subscribe registers this subscriber with the EventBus
func (s *AnalyticsSubscriber) Subscribe(eb *eventbus.EventBus) {
	eb.Subscribe(eventbus.EventTypeTradeExecuted, s.HandleTradeExecuted)
	log.Println("[ANALYTICS][INFO] Subscribed to trade_executed events")
}
