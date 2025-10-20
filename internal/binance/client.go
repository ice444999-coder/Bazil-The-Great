package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// BinanceClient handles communication with Binance API
type BinanceClient struct {
	baseURL     string
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// KlineInterval represents Binance kline/candlestick intervals
type KlineInterval string

const (
	Interval1m  KlineInterval = "1m"
	Interval5m  KlineInterval = "5m"
	Interval15m KlineInterval = "15m"
	Interval1h  KlineInterval = "1h"
	Interval4h  KlineInterval = "4h"
	Interval1d  KlineInterval = "1d"
)

// Kline represents a Binance candlestick
type Kline struct {
	OpenTime                 int64
	Open                     string
	High                     string
	Low                      string
	Close                    string
	Volume                   string
	CloseTime                int64
	QuoteAssetVolume         string
	NumberOfTrades           int
	TakerBuyBaseAssetVolume  string
	TakerBuyQuoteAssetVolume string
}

// HistoricalCandle represents a parsed candlestick for our system
type HistoricalCandle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// RateLimiter implements token bucket algorithm for Binance rate limits
type RateLimiter struct {
	tokens         int
	maxTokens      int
	refillRate     int // tokens per second
	lastRefillTime time.Time
}

// NewRateLimiter creates a rate limiter
// Binance limits: 1200 requests/minute = 20 req/sec
func NewRateLimiter(maxTokens, refillRate int) *RateLimiter {
	return &RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	for {
		rl.refill()
		if rl.tokens > 0 {
			rl.tokens--
			return
		}
		// Sleep for 50ms before checking again
		time.Sleep(50 * time.Millisecond)
	}
}

// refill adds tokens based on elapsed time
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime).Seconds()
	tokensToAdd := int(elapsed * float64(rl.refillRate))

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefillTime = now
	}
}

// NewBinanceClient creates a new Binance API client
func NewBinanceClient() *BinanceClient {
	return &BinanceClient{
		baseURL: "https://api.binance.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		// Binance rate limit: 1200 req/min = 20 req/sec
		// We'll be conservative and use 15 req/sec
		rateLimiter: NewRateLimiter(15, 15),
	}
}

// GetHistoricalKlines fetches historical candlestick data from Binance
// symbol: trading pair (e.g., "BTCUSDT")
// interval: candlestick interval (e.g., "1m", "5m", "1h")
// startTime: start timestamp in milliseconds
// endTime: end timestamp in milliseconds
// limit: max number of candles (default 500, max 1000)
func (c *BinanceClient) GetHistoricalKlines(symbol string, interval KlineInterval, startTime, endTime int64, limit int) ([]HistoricalCandle, error) {
	// Apply rate limiting
	c.rateLimiter.Wait()

	// Build URL
	url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s", c.baseURL, symbol, interval)

	if startTime > 0 {
		url += fmt.Sprintf("&startTime=%d", startTime)
	}
	if endTime > 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}
	if limit > 0 {
		if limit > 1000 {
			limit = 1000 // Binance max
		}
		url += fmt.Sprintf("&limit=%d", limit)
	}

	log.Printf("[BINANCE] Fetching %s %s candles from %s to %s (limit: %d)",
		symbol, interval,
		time.UnixMilli(startTime).Format("2006-01-02 15:04"),
		time.UnixMilli(endTime).Format("2006-01-02 15:04"),
		limit)

	// Make request
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch klines: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var rawKlines [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawKlines); err != nil {
		return nil, fmt.Errorf("failed to decode klines: %w", err)
	}

	// Convert to HistoricalCandle
	candles := make([]HistoricalCandle, 0, len(rawKlines))
	for _, raw := range rawKlines {
		candle, err := c.parseKline(raw)
		if err != nil {
			log.Printf("[BINANCE][WARN] Failed to parse kline: %v", err)
			continue
		}
		candles = append(candles, candle)
	}

	log.Printf("[BINANCE] ✅ Fetched %d candles for %s %s", len(candles), symbol, interval)
	return candles, nil
}

// parseKline converts raw Binance kline data to HistoricalCandle
// Raw format: [openTime, open, high, low, close, volume, closeTime, quoteVolume, trades, ...]
func (c *BinanceClient) parseKline(raw []interface{}) (HistoricalCandle, error) {
	if len(raw) < 6 {
		return HistoricalCandle{}, fmt.Errorf("invalid kline data: expected at least 6 fields, got %d", len(raw))
	}

	// Parse timestamp (openTime is in milliseconds)
	openTimeFloat, ok := raw[0].(float64)
	if !ok {
		return HistoricalCandle{}, fmt.Errorf("invalid openTime type")
	}
	timestamp := time.UnixMilli(int64(openTimeFloat))

	// Parse OHLCV (all come as strings from Binance)
	open, err := strconv.ParseFloat(raw[1].(string), 64)
	if err != nil {
		return HistoricalCandle{}, fmt.Errorf("invalid open price: %w", err)
	}

	high, err := strconv.ParseFloat(raw[2].(string), 64)
	if err != nil {
		return HistoricalCandle{}, fmt.Errorf("invalid high price: %w", err)
	}

	low, err := strconv.ParseFloat(raw[3].(string), 64)
	if err != nil {
		return HistoricalCandle{}, fmt.Errorf("invalid low price: %w", err)
	}

	close, err := strconv.ParseFloat(raw[4].(string), 64)
	if err != nil {
		return HistoricalCandle{}, fmt.Errorf("invalid close price: %w", err)
	}

	volume, err := strconv.ParseFloat(raw[5].(string), 64)
	if err != nil {
		return HistoricalCandle{}, fmt.Errorf("invalid volume: %w", err)
	}

	return HistoricalCandle{
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}, nil
}

// GetKlinesBatch fetches large datasets by making multiple requests
// Handles pagination automatically (Binance max 1000 candles per request)
func (c *BinanceClient) GetKlinesBatch(symbol string, interval KlineInterval, startTime, endTime int64) ([]HistoricalCandle, error) {
	var allCandles []HistoricalCandle
	currentStart := startTime

	// Calculate interval duration in milliseconds
	intervalMs := c.getIntervalMilliseconds(interval)

	for currentStart < endTime {
		// Binance max 1000 candles per request
		limit := 1000

		// Fetch batch
		candles, err := c.GetHistoricalKlines(symbol, interval, currentStart, endTime, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch batch starting at %s: %w",
				time.UnixMilli(currentStart).Format("2006-01-02 15:04"), err)
		}

		if len(candles) == 0 {
			break // No more data
		}

		allCandles = append(allCandles, candles...)

		// Move to next batch (use last candle's timestamp + interval)
		lastCandle := candles[len(candles)-1]
		currentStart = lastCandle.Timestamp.UnixMilli() + intervalMs

		// If we got less than limit, we've reached the end
		if len(candles) < limit {
			break
		}

		log.Printf("[BINANCE] Batch complete, total candles: %d, next start: %s",
			len(allCandles), time.UnixMilli(currentStart).Format("2006-01-02 15:04"))
	}

	log.Printf("[BINANCE] ✅ Batch complete: %d total candles for %s %s",
		len(allCandles), symbol, interval)
	return allCandles, nil
}

// getIntervalMilliseconds converts interval to milliseconds
func (c *BinanceClient) getIntervalMilliseconds(interval KlineInterval) int64 {
	switch interval {
	case Interval1m:
		return 60 * 1000
	case Interval5m:
		return 5 * 60 * 1000
	case Interval15m:
		return 15 * 60 * 1000
	case Interval1h:
		return 60 * 60 * 1000
	case Interval4h:
		return 4 * 60 * 60 * 1000
	case Interval1d:
		return 24 * 60 * 60 * 1000
	default:
		return 60 * 1000 // Default to 1m
	}
}

// GetLatestPrice fetches the current price for a symbol
func (c *BinanceClient) GetLatestPrice(symbol string) (float64, error) {
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", c.baseURL, symbol)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("binance API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode price: %w", err)
	}

	price, err := strconv.ParseFloat(result.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid price format: %w", err)
	}

	return price, nil
}

// TestConnection tests connectivity to Binance API
func (c *BinanceClient) TestConnection() error {
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/api/v3/ping", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to ping Binance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("binance ping failed (status %d)", resp.StatusCode)
	}

	log.Println("[BINANCE] ✅ Connection test successful")
	return nil
}
