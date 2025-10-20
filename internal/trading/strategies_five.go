package trading

import (
	"ares_api/internal/eventbus"
	"fmt"
	"math"
	"time"
)

// ========== 1. RSI OVERSOLD STRATEGY ==========
// Buys when RSI < 30 (oversold), sells when RSI > 70 (overbought)

type RSIOversoldStrategy struct {
	Period          int     // RSI calculation period (default: 14)
	OversoldLevel   float64 // Buy threshold (default: 30)
	OverboughtLevel float64 // Sell threshold (default: 70)
	eventBus        *eventbus.EventBus
}

func NewRSIOversoldStrategy(eb *eventbus.EventBus) *RSIOversoldStrategy {
	return &RSIOversoldStrategy{
		Period:          14,
		OversoldLevel:   30.0,
		OverboughtLevel: 70.0,
		eventBus:        eb,
	}
}

func (s *RSIOversoldStrategy) Name() string {
	return "RSI_Oversold"
}

func (s *RSIOversoldStrategy) Description() string {
	return "Buys when RSI indicates oversold conditions (<30), sells when overbought (>70). Mean reversion strategy."
}

func (s *RSIOversoldStrategy) GetRiskLevel() string {
	return "MEDIUM"
}

func (s *RSIOversoldStrategy) Analyze(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	if len(history) < s.Period {
		return &TradeSignal{
			Action:     "hold",
			Confidence: 0,
			Reasoning:  fmt.Sprintf("Insufficient data for RSI calculation (need %d periods)", s.Period),
			Strategy:   s.Name(),
			Symbol:     symbol,
		}, nil
	}

	// Calculate RSI
	rsi := s.calculateRSI(history, s.Period)
	currentPrice := marketData.GetPrice(symbol)

	// Generate signals based on RSI levels
	var signal *TradeSignal
	if rsi < s.OversoldLevel {
		// Oversold - BUY signal
		confidence := math.Min(100, (s.OversoldLevel-rsi)*2) // More oversold = higher confidence
		signal = &TradeSignal{
			Action:      "buy",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("RSI oversold at %.2f (threshold: %.2f). Strong mean reversion opportunity.", rsi, s.OversoldLevel),
			TargetPrice: currentPrice * 1.05, // 5% profit target
			StopLoss:    currentPrice * 0.97, // 3% stop loss
			Strategy:    s.Name(),
			Symbol:      symbol,
		}
	} else if rsi > s.OverboughtLevel {
		// Overbought - SELL signal
		confidence := math.Min(100, (rsi-s.OverboughtLevel)*2)
		signal = &TradeSignal{
			Action:      "sell",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("RSI overbought at %.2f (threshold: %.2f). Price likely to correct.", rsi, s.OverboughtLevel),
			TargetPrice: currentPrice * 0.95, // 5% down target
			StopLoss:    currentPrice * 1.03, // 3% stop loss
			Strategy:    s.Name(),
			Symbol:      symbol,
		}
	} else {
		// Neutral zone - HOLD
		signal = &TradeSignal{
			Action:     "hold",
			Confidence: 50,
			Reasoning:  fmt.Sprintf("RSI neutral at %.2f. Waiting for extreme levels.", rsi),
			Strategy:   s.Name(),
			Symbol:     symbol,
		}
	}

	// Publish signal event to EventBus
	s.publishSignal(signal)

	return signal, nil
}

// publishSignal publishes trade signal to EventBus
func (s *RSIOversoldStrategy) publishSignal(signal *TradeSignal) {
	if s.eventBus == nil {
		return
	}

	s.eventBus.Publish("strategy.RSI_Oversold.signal", map[string]interface{}{
		"strategy":     signal.Strategy,
		"action":       signal.Action,
		"symbol":       signal.Symbol,
		"confidence":   signal.Confidence,
		"reasoning":    signal.Reasoning,
		"target_price": signal.TargetPrice,
		"stop_loss":    signal.StopLoss,
		"timestamp":    time.Now(),
	})
}

func (s *RSIOversoldStrategy) calculateRSI(history []VirtualTrade, period int) float64 {
	if len(history) < period+1 {
		return 50.0 // Neutral RSI if insufficient data
	}

	gains := 0.0
	losses := 0.0

	// Calculate average gains and losses
	for i := len(history) - period; i < len(history); i++ {
		if i == 0 {
			continue
		}
		change := history[i].Price - history[i-1].Price
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100.0 // All gains = max RSI
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	return rsi
}

// ========== 2. MACD CROSSOVER STRATEGY ==========
// Buys on MACD bullish crossover, sells on bearish crossover

type MACDCrossoverStrategy struct {
	FastPeriod   int // Default: 12
	SlowPeriod   int // Default: 26
	SignalPeriod int // Default: 9
	eventBus     *eventbus.EventBus
}

func NewMACDCrossoverStrategy(eb *eventbus.EventBus) *MACDCrossoverStrategy {
	return &MACDCrossoverStrategy{
		FastPeriod:   12,
		SlowPeriod:   26,
		SignalPeriod: 9,
		eventBus:     eb,
	}
}

func (s *MACDCrossoverStrategy) Name() string {
	return "MACD_Crossover"
}

func (s *MACDCrossoverStrategy) Description() string {
	return "Trades based on MACD line crossing signal line. Bullish crossover = buy, bearish = sell."
}

func (s *MACDCrossoverStrategy) GetRiskLevel() string {
	return "MEDIUM"
}

func (s *MACDCrossoverStrategy) Analyze(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	requiredPeriods := s.SlowPeriod + s.SignalPeriod
	if len(history) < requiredPeriods {
		return &TradeSignal{
			Action:     "hold",
			Confidence: 0,
			Reasoning:  fmt.Sprintf("Insufficient data for MACD calculation (need %d periods)", requiredPeriods),
			Strategy:   s.Name(),
			Symbol:     symbol,
		}, nil
	}

	// Calculate MACD and signal line
	macdLine, signalLine := s.calculateMACD(history)
	macdHist := macdLine - signalLine
	currentPrice := marketData.GetPrice(symbol)

	// Detect crossover
	prevMACDLine, prevSignalLine := s.calculateMACDPrevious(history)
	prevHist := prevMACDLine - prevSignalLine

	// Bullish crossover (MACD crosses above signal)
	if prevHist <= 0 && macdHist > 0 {
		confidence := math.Min(100, math.Abs(macdHist)*10)
		return &TradeSignal{
			Action:      "buy",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("MACD bullish crossover detected. MACD: %.4f, Signal: %.4f, Histogram: %.4f", macdLine, signalLine, macdHist),
			TargetPrice: currentPrice * 1.06,
			StopLoss:    currentPrice * 0.96,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Bearish crossover (MACD crosses below signal)
	if prevHist >= 0 && macdHist < 0 {
		confidence := math.Min(100, math.Abs(macdHist)*10)
		return &TradeSignal{
			Action:      "sell",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("MACD bearish crossover detected. MACD: %.4f, Signal: %.4f, Histogram: %.4f", macdLine, signalLine, macdHist),
			TargetPrice: currentPrice * 0.94,
			StopLoss:    currentPrice * 1.04,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// No crossover - HOLD
	return &TradeSignal{
		Action:     "hold",
		Confidence: 50,
		Reasoning:  fmt.Sprintf("No MACD crossover. MACD: %.4f, Signal: %.4f, Histogram: %.4f", macdLine, signalLine, macdHist),
		Strategy:   s.Name(),
		Symbol:     symbol,
	}, nil
}

func (s *MACDCrossoverStrategy) calculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return prices[len(prices)-1]
	}

	multiplier := 2.0 / float64(period+1)
	ema := prices[0]

	for i := 1; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

func (s *MACDCrossoverStrategy) calculateMACD(history []VirtualTrade) (float64, float64) {
	prices := make([]float64, len(history))
	for i, trade := range history {
		prices[i] = trade.Price
	}

	fastEMA := s.calculateEMA(prices, s.FastPeriod)
	slowEMA := s.calculateEMA(prices, s.SlowPeriod)
	macdLine := fastEMA - slowEMA

	// Calculate signal line (EMA of MACD)
	macdHistory := []float64{macdLine} // Simplified - in production, track full MACD history
	signalLine := s.calculateEMA(macdHistory, s.SignalPeriod)

	return macdLine, signalLine
}

func (s *MACDCrossoverStrategy) calculateMACDPrevious(history []VirtualTrade) (float64, float64) {
	if len(history) < 2 {
		return 0, 0
	}
	prevHistory := history[:len(history)-1]
	return s.calculateMACD(prevHistory)
}

// ========== 3. TREND FOLLOWING STRATEGY ==========
// Follows strong directional trends using moving averages

type TrendFollowingStrategy struct {
	ShortMA  int     // Short-term MA period (default: 20)
	LongMA   int     // Long-term MA period (default: 50)
	MinTrend float64 // Minimum % difference for trend confirmation (default: 1.5%)
	eventBus *eventbus.EventBus
}

func NewTrendFollowingStrategy(eb *eventbus.EventBus) *TrendFollowingStrategy {
	return &TrendFollowingStrategy{
		ShortMA:  20,
		LongMA:   50,
		MinTrend: 1.5,
		eventBus: eb,
	}
}

func (s *TrendFollowingStrategy) Name() string {
	return "Trend_Following"
}

func (s *TrendFollowingStrategy) Description() string {
	return "Rides strong trends using MA crossovers. Short MA > Long MA = bullish, vice versa = bearish."
}

func (s *TrendFollowingStrategy) GetRiskLevel() string {
	return "HIGH"
}

func (s *TrendFollowingStrategy) Analyze(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	if len(history) < s.LongMA {
		return &TradeSignal{
			Action:     "hold",
			Confidence: 0,
			Reasoning:  fmt.Sprintf("Insufficient data for trend analysis (need %d periods)", s.LongMA),
			Strategy:   s.Name(),
			Symbol:     symbol,
		}, nil
	}

	// Calculate moving averages
	shortMA := s.calculateSMA(history, s.ShortMA)
	longMA := s.calculateSMA(history, s.LongMA)
	trendStrength := ((shortMA - longMA) / longMA) * 100
	currentPrice := marketData.GetPrice(symbol)

	// Strong uptrend
	if trendStrength > s.MinTrend {
		confidence := math.Min(100, math.Abs(trendStrength)*10)
		return &TradeSignal{
			Action:      "buy",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Strong uptrend detected. Short MA (%.2f) > Long MA (%.2f) by %.2f%%", shortMA, longMA, trendStrength),
			TargetPrice: currentPrice * 1.08, // Ride the trend
			StopLoss:    currentPrice * 0.95,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Strong downtrend
	if trendStrength < -s.MinTrend {
		confidence := math.Min(100, math.Abs(trendStrength)*10)
		return &TradeSignal{
			Action:      "sell",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Strong downtrend detected. Short MA (%.2f) < Long MA (%.2f) by %.2f%%", shortMA, longMA, trendStrength),
			TargetPrice: currentPrice * 0.92,
			StopLoss:    currentPrice * 1.05,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Weak trend or consolidation
	return &TradeSignal{
		Action:     "hold",
		Confidence: 30,
		Reasoning:  fmt.Sprintf("Weak trend or consolidation. Trend strength: %.2f%% (need >%.2f%%)", trendStrength, s.MinTrend),
		Strategy:   s.Name(),
		Symbol:     symbol,
	}, nil
}

func (s *TrendFollowingStrategy) calculateSMA(history []VirtualTrade, period int) float64 {
	if len(history) < period {
		period = len(history)
	}

	sum := 0.0
	start := len(history) - period
	for i := start; i < len(history); i++ {
		sum += history[i].Price
	}

	return sum / float64(period)
}

// ========== 4. SUPPORT BOUNCE STRATEGY ==========
// Buys when price bounces off support levels

type SupportBounceStrategy struct {
	LookbackPeriod  int     // Periods to identify support/resistance (default: 30)
	BounceThreshold float64 // % proximity to support to trigger (default: 0.5%)
	MinTouches      int     // Minimum touches to confirm support (default: 2)
	eventBus        *eventbus.EventBus
}

func NewSupportBounceStrategy(eb *eventbus.EventBus) *SupportBounceStrategy {
	return &SupportBounceStrategy{
		LookbackPeriod:  30,
		BounceThreshold: 0.5,
		MinTouches:      2,
		eventBus:        eb,
	}
}

func (s *SupportBounceStrategy) Name() string {
	return "Support_Bounce"
}

func (s *SupportBounceStrategy) Description() string {
	return "Buys when price bounces off identified support levels. Sells at resistance."
}

func (s *SupportBounceStrategy) GetRiskLevel() string {
	return "MEDIUM"
}

func (s *SupportBounceStrategy) Analyze(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	if len(history) < s.LookbackPeriod {
		return &TradeSignal{
			Action:     "hold",
			Confidence: 0,
			Reasoning:  fmt.Sprintf("Insufficient data for support/resistance analysis (need %d periods)", s.LookbackPeriod),
			Strategy:   s.Name(),
			Symbol:     symbol,
		}, nil
	}

	// Identify support and resistance levels
	support, supportTouches := s.findSupportLevel(history)
	resistance, resistanceTouches := s.findResistanceLevel(history)
	currentPrice := marketData.GetPrice(symbol)

	// Check if near support (bounce opportunity)
	distanceToSupport := ((currentPrice - support) / support) * 100
	if math.Abs(distanceToSupport) <= s.BounceThreshold && supportTouches >= s.MinTouches {
		confidence := math.Min(100, float64(supportTouches)*20)
		return &TradeSignal{
			Action:      "buy",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Price near strong support at %.2f (%.2f%% away, %d touches). Bounce expected.", support, distanceToSupport, supportTouches),
			TargetPrice: resistance,
			StopLoss:    support * 0.98, // Tight stop below support
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Check if near resistance (rejection likely)
	distanceToResistance := ((currentPrice - resistance) / resistance) * 100
	if math.Abs(distanceToResistance) <= s.BounceThreshold && resistanceTouches >= s.MinTouches {
		confidence := math.Min(100, float64(resistanceTouches)*20)
		return &TradeSignal{
			Action:      "sell",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Price near strong resistance at %.2f (%.2f%% away, %d touches). Rejection likely.", resistance, distanceToResistance, resistanceTouches),
			TargetPrice: support,
			StopLoss:    resistance * 1.02,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Price in mid-range
	return &TradeSignal{
		Action:     "hold",
		Confidence: 40,
		Reasoning:  fmt.Sprintf("Price at %.2f, mid-range between support (%.2f) and resistance (%.2f)", currentPrice, support, resistance),
		Strategy:   s.Name(),
		Symbol:     symbol,
	}, nil
}

func (s *SupportBounceStrategy) findSupportLevel(history []VirtualTrade) (float64, int) {
	recentHistory := history[len(history)-s.LookbackPeriod:]

	// Find local minima
	minPrice := math.MaxFloat64
	touches := 0
	tolerance := 0.01 // 1% tolerance for support level

	for _, trade := range recentHistory {
		if trade.Price < minPrice {
			minPrice = trade.Price
		}
	}

	// Count how many times price touched this level
	for _, trade := range recentHistory {
		if math.Abs((trade.Price-minPrice)/minPrice) <= tolerance {
			touches++
		}
	}

	return minPrice, touches
}

func (s *SupportBounceStrategy) findResistanceLevel(history []VirtualTrade) (float64, int) {
	recentHistory := history[len(history)-s.LookbackPeriod:]

	// Find local maxima
	maxPrice := 0.0
	touches := 0
	tolerance := 0.01 // 1% tolerance

	for _, trade := range recentHistory {
		if trade.Price > maxPrice {
			maxPrice = trade.Price
		}
	}

	// Count touches
	for _, trade := range recentHistory {
		if math.Abs((trade.Price-maxPrice)/maxPrice) <= tolerance {
			touches++
		}
	}

	return maxPrice, touches
}

// ========== 5. VOLUME BREAKOUT STRATEGY ==========
// Trades breakouts confirmed by volume spikes

type VolumeBreakoutStrategy struct {
	VolumePeriod     int     // Period for average volume (default: 20)
	VolumeMultiplier float64 // Volume spike threshold (default: 2.0x)
	PriceThreshold   float64 // Minimum % price move for breakout (default: 2%)
	eventBus         *eventbus.EventBus
}

func NewVolumeBreakoutStrategy(eb *eventbus.EventBus) *VolumeBreakoutStrategy {
	return &VolumeBreakoutStrategy{
		VolumePeriod:     20,
		VolumeMultiplier: 2.0,
		PriceThreshold:   2.0,
		eventBus:         eb,
	}
}

func (s *VolumeBreakoutStrategy) Name() string {
	return "Volume_Breakout"
}

func (s *VolumeBreakoutStrategy) Description() string {
	return "Trades price breakouts confirmed by high volume. Volume spike + price move = strong signal."
}

func (s *VolumeBreakoutStrategy) GetRiskLevel() string {
	return "HIGH"
}

func (s *VolumeBreakoutStrategy) Analyze(symbol string, marketData *MockMarketData, history []VirtualTrade) (*TradeSignal, error) {
	if len(history) < s.VolumePeriod+1 {
		return &TradeSignal{
			Action:     "hold",
			Confidence: 0,
			Reasoning:  fmt.Sprintf("Insufficient data for volume analysis (need %d periods)", s.VolumePeriod),
			Strategy:   s.Name(),
			Symbol:     symbol,
		}, nil
	}

	// Calculate average volume
	avgVolume := s.calculateAverageVolume(history)
	currentVolume := history[len(history)-1].Amount // Use latest trade amount as proxy for volume
	volumeRatio := currentVolume / avgVolume

	// Calculate price change
	previousPrice := history[len(history)-1].Price
	currentPrice := marketData.GetPrice(symbol)
	priceChange := ((currentPrice - previousPrice) / previousPrice) * 100

	// Bullish breakout (volume spike + strong upward move)
	if volumeRatio >= s.VolumeMultiplier && priceChange >= s.PriceThreshold {
		confidence := math.Min(100, volumeRatio*20+math.Abs(priceChange)*5)
		return &TradeSignal{
			Action:      "buy",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Bullish breakout! Volume spike %.2fx average with +%.2f%% price move. Strong momentum.", volumeRatio, priceChange),
			TargetPrice: currentPrice * 1.10,  // Aggressive target on breakout
			StopLoss:    previousPrice * 0.98, // Stop at breakout level
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// Bearish breakdown (volume spike + strong downward move)
	if volumeRatio >= s.VolumeMultiplier && priceChange <= -s.PriceThreshold {
		confidence := math.Min(100, volumeRatio*20+math.Abs(priceChange)*5)
		return &TradeSignal{
			Action:      "sell",
			Confidence:  confidence,
			Reasoning:   fmt.Sprintf("Bearish breakdown! Volume spike %.2fx average with %.2f%% price drop. Strong selling pressure.", volumeRatio, priceChange),
			TargetPrice: currentPrice * 0.90,
			StopLoss:    previousPrice * 1.02,
			Strategy:    s.Name(),
			Symbol:      symbol,
		}, nil
	}

	// No breakout signal
	return &TradeSignal{
		Action:     "hold",
		Confidence: 25,
		Reasoning:  fmt.Sprintf("No breakout. Volume ratio: %.2fx (need %.2fx), Price change: %.2f%% (need %.2f%%)", volumeRatio, s.VolumeMultiplier, priceChange, s.PriceThreshold),
		Strategy:   s.Name(),
		Symbol:     symbol,
	}, nil
}

func (s *VolumeBreakoutStrategy) calculateAverageVolume(history []VirtualTrade) float64 {
	period := s.VolumePeriod
	if len(history) < period {
		period = len(history)
	}

	sum := 0.0
	start := len(history) - period
	for i := start; i < len(history); i++ {
		// Use Amount as proxy for volume
		sum += history[i].Amount
	}

	return sum / float64(period)
}

// ========== STRATEGY REGISTRY ==========
// Factory function to create all strategies

func GetAllStrategies(eb *eventbus.EventBus) []Strategy {
	return []Strategy{
		NewRSIOversoldStrategy(eb),
		NewMACDCrossoverStrategy(eb),
		NewTrendFollowingStrategy(eb),
		NewSupportBounceStrategy(eb),
		NewVolumeBreakoutStrategy(eb),
	}
}

func GetStrategyByName(name string, eb *eventbus.EventBus) (Strategy, error) {
	strategies := map[string]Strategy{
		"RSI_Oversold":    NewRSIOversoldStrategy(eb),
		"MACD_Crossover":  NewMACDCrossoverStrategy(eb),
		"Trend_Following": NewTrendFollowingStrategy(eb),
		"Support_Bounce":  NewSupportBounceStrategy(eb),
		"Volume_Breakout": NewVolumeBreakoutStrategy(eb),
	}

	strategy, exists := strategies[name]
	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", name)
	}

	return strategy, nil
}
