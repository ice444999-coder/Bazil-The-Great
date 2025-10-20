package trading

import (
	"ares_api/internal/eventbus"
	"testing"
	"time"
)

func TestAllStrategiesImplementInterface(t *testing.T) {
	eb := eventbus.NewEventBus()
	defer eb.Close()
	strategies := GetAllStrategies(eb)

	if len(strategies) != 5 {
		t.Errorf("Expected 5 strategies, got %d", len(strategies))
	}

	expectedNames := map[string]bool{
		"RSI_Oversold":    false,
		"MACD_Crossover":  false,
		"Trend_Following": false,
		"Support_Bounce":  false,
		"Volume_Breakout": false,
	}

	for _, s := range strategies {
		name := s.Name()
		if _, exists := expectedNames[name]; !exists {
			t.Errorf("Unexpected strategy name: %s", name)
		}
		expectedNames[name] = true

		// Verify interface methods
		if s.Description() == "" {
			t.Errorf("Strategy %s has empty description", name)
		}
		if s.GetRiskLevel() == "" {
			t.Errorf("Strategy %s has empty risk level", name)
		}
	}

	// Verify all expected strategies were found
	for name, found := range expectedNames {
		if !found {
			t.Errorf("Strategy %s not found in GetAllStrategies()", name)
		}
	}
}

func TestGetStrategyByName(t *testing.T) {
	eb := eventbus.NewEventBus()
	defer eb.Close()

	tests := []struct {
		name      string
		wantError bool
	}{
		{"RSI_Oversold", false},
		{"MACD_Crossover", false},
		{"Trend_Following", false},
		{"Support_Bounce", false},
		{"Volume_Breakout", false},
		{"NonExistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := GetStrategyByName(tt.name, eb)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for strategy %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for strategy %s: %v", tt.name, err)
				}
				if strategy == nil {
					t.Errorf("Expected strategy for %s, got nil", tt.name)
				}
				if strategy.Name() != tt.name {
					t.Errorf("Expected name %s, got %s", tt.name, strategy.Name())
				}
			}
		})
	}
}

func TestRSIOversoldStrategy(t *testing.T) {
	eb := eventbus.NewEventBus()
	defer eb.Close()
	strategy := NewRSIOversoldStrategy(eb)

	// Create mock market data (uses default prices including BTC/USDC)
	marketData := NewMockMarketData()

	// Create history with trending prices
	history := make([]VirtualTrade, 20)
	basePrice := 45000.0
	for i := 0; i < 20; i++ {
		history[i] = VirtualTrade{
			ID:         string(rune(i)),
			Symbol:     "BTC/USDC",
			Price:      basePrice + float64(i)*100, // Increasing prices
			Amount:     1.0,
			ExecutedAt: time.Now().Add(-time.Duration(20-i) * time.Hour),
			Status:     "closed",
		}
	}

	// Analyze
	signal, err := strategy.Analyze("BTC/USDC", marketData, history)
	if err != nil {
		t.Errorf("RSI strategy failed: %v", err)
	}

	if signal == nil {
		t.Fatal("Expected signal, got nil")
	}

	if signal.Strategy != "RSI_Oversold" {
		t.Errorf("Expected strategy name RSI_Oversold, got %s", signal.Strategy)
	}

	if signal.Action == "" {
		t.Error("Signal action is empty")
	}

	t.Logf("RSI Signal: %s (%.2f%% confidence) - %s", signal.Action, signal.Confidence, signal.Reasoning)
}

func TestMACDCrossoverStrategy(t *testing.T) {
	eb := eventbus.NewEventBus()
	defer eb.Close()
	strategy := NewMACDCrossoverStrategy(eb)

	marketData := NewMockMarketData()

	// Create sufficient history for MACD
	history := make([]VirtualTrade, 50)
	for i := 0; i < 50; i++ {
		history[i] = VirtualTrade{
			ID:         string(rune(i)),
			Symbol:     "BTC/USDC",
			Price:      44000.0 + float64(i)*50,
			Amount:     1.0,
			ExecutedAt: time.Now().Add(-time.Duration(50-i) * time.Hour),
			Status:     "closed",
		}
	}

	signal, err := strategy.Analyze("BTC/USDC", marketData, history)
	if err != nil {
		t.Errorf("MACD strategy failed: %v", err)
	}

	if signal == nil {
		t.Fatal("Expected signal, got nil")
	}

	t.Logf("MACD Signal: %s (%.2f%% confidence) - %s", signal.Action, signal.Confidence, signal.Reasoning)
}

func TestTrendFollowingStrategy(t *testing.T) {
	eb := eventbus.NewEventBus()
	defer eb.Close()
	strategy := NewTrendFollowingStrategy(eb)

	marketData := NewMockMarketData()

	// Create strong uptrend
	history := make([]VirtualTrade, 60)
	for i := 0; i < 60; i++ {
		history[i] = VirtualTrade{
			ID:         string(rune(i)),
			Symbol:     "BTC/USDC",
			Price:      42000.0 + float64(i)*80, // Strong uptrend
			Amount:     1.0,
			ExecutedAt: time.Now().Add(-time.Duration(60-i) * time.Hour),
			Status:     "closed",
		}
	}

	signal, err := strategy.Analyze("BTC/USDC", marketData, history)
	if err != nil {
		t.Errorf("Trend strategy failed: %v", err)
	}

	if signal == nil {
		t.Fatal("Expected signal, got nil")
	}

	t.Logf("Trend Signal: %s (%.2f%% confidence) - %s", signal.Action, signal.Confidence, signal.Reasoning)
}

func TestSupportBounceStrategy(t *testing.T) {
	eb := eventbus.NewEventBus()
	defer eb.Close()
	strategy := NewSupportBounceStrategy(eb)

	marketData := NewMockMarketData()

	// Create history with support at 44000
	history := make([]VirtualTrade, 40)
	for i := 0; i < 40; i++ {
		price := 44000.0
		if i%5 == 0 {
			price = 44000.0 // Touch support
		} else {
			price = 44000.0 + float64((i%5))*200
		}
		history[i] = VirtualTrade{
			ID:         string(rune(i)),
			Symbol:     "BTC/USDC",
			Price:      price,
			Amount:     1.0,
			ExecutedAt: time.Now().Add(-time.Duration(40-i) * time.Hour),
			Status:     "closed",
		}
	}

	signal, err := strategy.Analyze("BTC/USDC", marketData, history)
	if err != nil {
		t.Errorf("Support strategy failed: %v", err)
	}

	if signal == nil {
		t.Fatal("Expected signal, got nil")
	}

	t.Logf("Support Signal: %s (%.2f%% confidence) - %s", signal.Action, signal.Confidence, signal.Reasoning)
}

func TestVolumeBreakoutStrategy(t *testing.T) {
	eb := eventbus.NewEventBus()
	defer eb.Close()
	strategy := NewVolumeBreakoutStrategy(eb)

	marketData := NewMockMarketData()

	// Create history with normal volume
	history := make([]VirtualTrade, 30)
	for i := 0; i < 30; i++ {
		amount := 1.0
		if i == 29 {
			amount = 5.0 // Volume spike on last trade
		}
		history[i] = VirtualTrade{
			ID:         string(rune(i)),
			Symbol:     "BTC/USDC",
			Price:      43000.0 + float64(i)*50,
			Amount:     amount,
			ExecutedAt: time.Now().Add(-time.Duration(30-i) * time.Hour),
			Status:     "closed",
		}
	}

	signal, err := strategy.Analyze("BTC/USDC", marketData, history)
	if err != nil {
		t.Errorf("Volume strategy failed: %v", err)
	}

	if signal == nil {
		t.Fatal("Expected signal, got nil")
	}

	t.Logf("Volume Signal: %s (%.2f%% confidence) - %s", signal.Action, signal.Confidence, signal.Reasoning)
}

func TestStrategyRiskLevels(t *testing.T) {
	eb := eventbus.NewEventBus()
	defer eb.Close()

	strategies := map[string]string{
		"RSI_Oversold":    "MEDIUM",
		"MACD_Crossover":  "MEDIUM",
		"Trend_Following": "HIGH",
		"Support_Bounce":  "MEDIUM",
		"Volume_Breakout": "HIGH",
	}

	for name, expectedRisk := range strategies {
		strategy, err := GetStrategyByName(name, eb)
		if err != nil {
			t.Errorf("Failed to get strategy %s: %v", name, err)
			continue
		}

		actualRisk := strategy.GetRiskLevel()
		if actualRisk != expectedRisk {
			t.Errorf("Strategy %s: expected risk level %s, got %s", name, expectedRisk, actualRisk)
		}
	}
}
