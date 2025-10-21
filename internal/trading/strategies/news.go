/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package strategies

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// NewsStrategy implements news-based trading using alert patterns and sentiment analysis
// Target: 40-80% annual returns from news-driven volatility
// Focus: Alert-based dumps, earnings surprises, regulatory news
type NewsStrategy struct {
	config map[string]interface{}
}

// NewNewsStrategy creates a new news strategy
func NewNewsStrategy() *NewsStrategy {
	return &NewsStrategy{
		config: map[string]interface{}{
			"minGain":              0.02,  // 2% min gain
			"maxHold":              3600,  // 1 hour max hold
			"stopLoss":             -0.05, // 5% stop loss
			"targetGain":           0.08,  // 8% target gain
			"sentimentThreshold":   0.7,   // Sentiment score threshold
			"volumeSpikeThreshold": 3.0,   // Volume spike multiplier
			"priceGapThreshold":    0.03,  // Price gap threshold
			"newsCooldown":         300,   // 5 minutes between news trades
		},
	}
}

// Generate generates trading signals based on news analysis
func (s *NewsStrategy) Generate(data *MarketData) (*TradeSignal, error) {
	if len(data.PriceHistory) < 20 {
		return &TradeSignal{Action: "hold"}, nil
	}

	// Analyze recent price action for news patterns
	priceGap := s.detectPriceGap(data)
	volumeSpike := s.detectVolumeSpike(data)
	sentimentScore := s.analyzeNewsSentiment(data)

	// Check for news-driven patterns
	newsEvent := s.detectNewsEvent(data, priceGap, volumeSpike, sentimentScore)

	currentPrice := data.CurrentPrice
	targetGain := s.getConfigFloat("targetGain")
	stopLoss := s.getConfigFloat("stopLoss")

	action := "hold"
	confidence := 0.0
	reasoning := ""

	// BUY conditions: Positive news + price gap up + volume spike
	if newsEvent == "positive_news" && priceGap > s.getConfigFloat("priceGapThreshold") && volumeSpike > s.getConfigFloat("volumeSpikeThreshold") {
		action = "buy"
		confidence = 0.85
		reasoning = fmt.Sprintf("Positive news event detected with %.2f%% price gap and %.1fx volume spike. Sentiment: %.2f",
			priceGap*100, volumeSpike, sentimentScore)
	}

	// SELL conditions: Negative news + price gap down + volume spike
	if newsEvent == "negative_news" && priceGap < -s.getConfigFloat("priceGapThreshold") && volumeSpike > s.getConfigFloat("volumeSpikeThreshold") {
		action = "sell"
		confidence = 0.85
		reasoning = fmt.Sprintf("Negative news event detected with %.2f%% price gap and %.1fx volume spike. Sentiment: %.2f",
			priceGap*100, volumeSpike, sentimentScore)
	}

	// Additional exit conditions for open positions
	if action == "hold" {
		// Check for profit taking or stop loss based on price history
		if len(data.PriceHistory) >= 2 {
			prevPrice := data.PriceHistory[len(data.PriceHistory)-2]
			priceChange := (currentPrice - prevPrice) / prevPrice

			// Take profit if target reached
			if priceChange >= targetGain {
				action = "sell"
				confidence = 0.9
				reasoning = fmt.Sprintf("Profit target reached: +%.2f%%", priceChange*100)
			}

			// Stop loss if loss too large
			if priceChange <= stopLoss {
				action = "sell"
				confidence = 0.95
				reasoning = fmt.Sprintf("Stop loss triggered: %.2f%%", priceChange*100)
			}
		}
	}

	return &TradeSignal{
		Action:      action,
		Symbol:      data.Symbol,
		Confidence:  confidence,
		Reasoning:   reasoning,
		Strategy:    "news",
		Timestamp:   time.Now(),
		TargetGain:  targetGain,
		StopLoss:    stopLoss,
		MaxHoldTime: int(s.getConfigFloat("maxHold")),
		Priority:    9, // Very high priority for news events
	}, nil
}

// Analyze performs detailed analysis for the strategy
func (s *NewsStrategy) Analyze(data *MarketData) *StrategyAnalysis {
	signal, _ := s.Generate(data)

	score := 0.0
	recommendation := "NO_SIGNAL"

	// Calculate score based on signal confidence and market conditions
	if signal.Confidence > 0.8 {
		score = signal.Confidence * 100
		if signal.Action == "buy" {
			recommendation = "STRONG_BUY"
		} else if signal.Action == "sell" {
			recommendation = "STRONG_SELL"
		}
	} else if signal.Confidence > 0.6 {
		score = signal.Confidence * 80
		if signal.Action == "buy" {
			recommendation = "MODERATE_BUY"
		} else if signal.Action == "sell" {
			recommendation = "MODERATE_SELL"
		}
	}

	// Calculate current indicators for analysis
	indicators := map[string]float64{}
	if len(data.PriceHistory) >= 20 {
		priceGap := s.detectPriceGap(data)
		volumeSpike := s.detectVolumeSpike(data)
		sentimentScore := s.analyzeNewsSentiment(data)
		newsEvent := s.detectNewsEvent(data, priceGap, volumeSpike, sentimentScore)

		indicators["price_gap"] = priceGap
		indicators["volume_spike"] = volumeSpike
		indicators["sentiment_score"] = sentimentScore
		indicators["news_event"] = s.newsEventToFloat(newsEvent)
	}

	return &StrategyAnalysis{
		StrategyName:   "news",
		Score:          score,
		Indicators:     indicators,
		Recommendation: recommendation,
		Timestamp:      time.Now(),
	}
}

// GetConfig returns the strategy configuration
func (s *NewsStrategy) GetConfig() map[string]interface{} {
	return s.config
}

// UpdateConfig updates the strategy configuration
func (s *NewsStrategy) UpdateConfig(params map[string]interface{}) error {
	for k, v := range params {
		s.config[k] = v
	}
	return nil
}

// Helper functions

func (s *NewsStrategy) getConfigFloat(key string) float64 {
	if val, ok := s.config[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

func (s *NewsStrategy) newsEventToFloat(event string) float64 {
	switch event {
	case "positive_news":
		return 1.0
	case "negative_news":
		return -1.0
	default:
		return 0.0
	}
}

// detectPriceGap detects significant price gaps that indicate news events
func (s *NewsStrategy) detectPriceGap(data *MarketData) float64 {
	if len(data.PriceHistory) < 2 {
		return 0.0
	}

	// Calculate gap between last two prices
	prevPrice := data.PriceHistory[len(data.PriceHistory)-2]
	currentPrice := data.CurrentPrice

	return (currentPrice - prevPrice) / prevPrice
}

// detectVolumeSpike detects unusual volume spikes
func (s *NewsStrategy) detectVolumeSpike(data *MarketData) float64 {
	if len(data.VolumeHistory) < 20 {
		return 1.0
	}

	// Calculate average volume over last 20 periods
	avgVolume := 0.0
	for _, vol := range data.VolumeHistory[len(data.VolumeHistory)-20:] {
		avgVolume += vol
	}
	avgVolume /= 20

	// Current volume
	currentVolume := data.VolumeHistory[len(data.VolumeHistory)-1]

	if avgVolume == 0 {
		return 1.0
	}

	return currentVolume / avgVolume
}

// analyzeNewsSentiment performs basic sentiment analysis on news data
// In a real implementation, this would integrate with news APIs and NLP
func (s *NewsStrategy) analyzeNewsSentiment(data *MarketData) float64 {
	// Mock sentiment analysis - in reality this would analyze news headlines
	// For now, we'll use price and volume patterns as sentiment proxies

	priceGap := s.detectPriceGap(data)
	volumeSpike := s.detectVolumeSpike(data)

	// Positive sentiment: upward price gap + volume spike
	if priceGap > 0.02 && volumeSpike > 2.0 {
		return 0.8
	}

	// Negative sentiment: downward price gap + volume spike
	if priceGap < -0.02 && volumeSpike > 2.0 {
		return -0.8
	}

	// Neutral sentiment
	return 0.0
}

// detectNewsEvent determines if a news event has occurred
func (s *NewsStrategy) detectNewsEvent(data *MarketData, priceGap, volumeSpike, sentiment float64) string {
	threshold := s.getConfigFloat("sentimentThreshold")
	volumeThreshold := s.getConfigFloat("volumeSpikeThreshold")
	gapThreshold := s.getConfigFloat("priceGapThreshold")

	// Positive news event
	if sentiment > threshold && volumeSpike > volumeThreshold && priceGap > gapThreshold {
		return "positive_news"
	}

	// Negative news event
	if sentiment < -threshold && volumeSpike > volumeThreshold && priceGap < -gapThreshold {
		return "negative_news"
	}

	return "no_event"
}

// Additional helper: analyzeNewsHeadlines (for future NLP integration)
func (s *NewsStrategy) analyzeNewsHeadlines(headlines []string) float64 {
	if len(headlines) == 0 {
		return 0.0
	}

	positiveWords := []string{"surge", "rally", "gains", "bullish", "upgrade", "beats", "earnings beat", "positive"}
	negativeWords := []string{"crash", "plunge", "losses", "bearish", "downgrade", "misses", "earnings miss", "negative"}

	score := 0.0
	totalWords := 0

	for _, headline := range headlines {
		lower := strings.ToLower(headline)
		totalWords += len(strings.Fields(headline))

		for _, word := range positiveWords {
			if matched, _ := regexp.MatchString("\\b"+word+"\\b", lower); matched {
				score += 1.0
			}
		}

		for _, word := range negativeWords {
			if matched, _ := regexp.MatchString("\\b"+word+"\\b", lower); matched {
				score -= 1.0
			}
		}
	}

	if totalWords == 0 {
		return 0.0
	}

	return score / float64(totalWords)
}
