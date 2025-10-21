/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package cache

import (
	"ares_api/internal/api/dto"
	"log"
	"sync"
	"time"
)

// PriceCache provides in-memory caching for market data with TTL
type PriceCache struct {
	prices map[string]*CachedPrice
	mu     sync.RWMutex
	ttl    time.Duration
}

// CachedPrice stores a price with metadata
type CachedPrice struct {
	Data      *dto.CoinMarketDTO
	Timestamp time.Time
}

// NewPriceCache creates a new price cache with specified TTL
func NewPriceCache(ttl time.Duration) *PriceCache {
	cache := &PriceCache{
		prices: make(map[string]*CachedPrice),
		ttl:    ttl,
	}

	// Start cleanup goroutine (every 5 minutes)
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a cached price if available and not expired
func (pc *PriceCache) Get(symbol string) (*dto.CoinMarketDTO, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	cached, exists := pc.prices[symbol]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(cached.Timestamp) > pc.ttl {
		log.Printf("[CACHE][DEBUG] Price for %s expired (age: %v)", symbol, time.Since(cached.Timestamp))
		return nil, false
	}

	age := time.Since(cached.Timestamp)
	log.Printf("[CACHE][HIT] Using cached price for %s (age: %v, price: $%.2f)",
		symbol, age.Round(time.Second), cached.Data.PriceUSD)

	return cached.Data, true
}

// Set stores a price in the cache
func (pc *PriceCache) Set(symbol string, data *dto.CoinMarketDTO) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.prices[symbol] = &CachedPrice{
		Data:      data,
		Timestamp: time.Now(),
	}

	log.Printf("[CACHE][SET] Cached price for %s: $%.2f", symbol, data.PriceUSD)
}

// GetStale retrieves any cached price, even if expired (for emergency fallback)
func (pc *PriceCache) GetStale(symbol string) (*dto.CoinMarketDTO, time.Duration, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	cached, exists := pc.prices[symbol]
	if !exists {
		return nil, 0, false
	}

	age := time.Since(cached.Timestamp)
	log.Printf("[CACHE][STALE] Using stale price for %s (age: %v, price: $%.2f)",
		symbol, age.Round(time.Second), cached.Data.PriceUSD)

	return cached.Data, age, true
}

// cleanupExpired removes expired entries periodically
func (pc *PriceCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		pc.mu.Lock()

		removed := 0
		for symbol, cached := range pc.prices {
			// Keep entries for 24 hours max (even if TTL is shorter)
			if time.Since(cached.Timestamp) > 24*time.Hour {
				delete(pc.prices, symbol)
				removed++
			}
		}

		if removed > 0 {
			log.Printf("[CACHE][CLEANUP] Removed %d expired entries (total remaining: %d)",
				removed, len(pc.prices))
		}

		pc.mu.Unlock()
	}
}

// Stats returns cache statistics
func (pc *PriceCache) Stats() map[string]interface{} {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	fresh := 0
	stale := 0

	for _, cached := range pc.prices {
		if time.Since(cached.Timestamp) <= pc.ttl {
			fresh++
		} else {
			stale++
		}
	}

	return map[string]interface{}{
		"total_entries": len(pc.prices),
		"fresh_entries": fresh,
		"stale_entries": stale,
		"ttl_seconds":   int(pc.ttl.Seconds()),
	}
}
