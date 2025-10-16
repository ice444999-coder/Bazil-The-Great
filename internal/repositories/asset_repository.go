package repositories

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/cache"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	repository "ares_api/internal/interfaces/repository"
)

type AssetRepositoryImpl struct {
	BaseURL    string
	APIKey     string
	priceCache *cache.PriceCache // Phase 3: Graceful degradation
}

func NewAssetRepository() repository.AssetRepository {
	baseURL := os.Getenv("COINGECKO_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.coingecko.com/api/v3"
	}

	// Initialize price cache with 2-minute TTL
	priceCache := cache.NewPriceCache(2 * time.Minute)

	return &AssetRepositoryImpl{
		BaseURL:    baseURL,
		APIKey:     os.Getenv("COINGECKO_API_KEY"),
		priceCache: priceCache,
	}
}

func (r *AssetRepositoryImpl) FetchAllCoins() ([]dto.CoinDTO, error) {
	url := fmt.Sprintf("%s/coins/list", r.BaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	if r.APIKey != "" {
		req.Header.Add("X-CoinGecko-API-Key", r.APIKey)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var coins []dto.CoinDTO
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, err
	}
	return coins, nil
}

func (r *AssetRepositoryImpl) FetchCoinMarket(id, vsCurrency string) (*dto.CoinMarketDTO, error) {
	// Phase 3: Check cache first
	if cached, found := r.priceCache.Get(id); found {
		return cached, nil
	}

	url := fmt.Sprintf("%s/coins/markets?vs_currency=%s&ids=%s&order=market_cap_desc&sparkline=false",
		r.BaseURL, vsCurrency, id)

	req, _ := http.NewRequest("GET", url, nil)
	if r.APIKey != "" {
		req.Header.Add("X-CoinGecko-API-Key", r.APIKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Phase 3: API unavailable - try stale cache
		if stale, age, found := r.priceCache.GetStale(id); found {
			log.Printf("⚠️ CoinGecko API error, using stale cache (age: %v): %v", age.Round(time.Second), err)
			return stale, nil
		}

		log.Printf("⚠️ CoinGecko API error, no cache available: %v", err)
		return nil, fmt.Errorf("market data unavailable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[DEBUG] CoinGecko response for %s: %s", id, string(body))

	// Check for rate limit (429) - use sandbox fallback
	if resp.StatusCode == 429 {
		// Phase 3: Rate limited - try stale cache first
		if stale, age, found := r.priceCache.GetStale(id); found {
			log.Printf("⚠️ CoinGecko rate limit hit, using stale cache (age: %v)", age.Round(time.Second))
			return stale, nil
		}

		log.Println("⚠️ CoinGecko rate limit hit - using sandbox fallback prices")
		return r.getSandboxFallbackPrice(id)
	}

	var data []struct {
		ID          string  `json:"id"`
		Symbol      string  `json:"symbol"`
		Name        string  `json:"name"`
		PriceUSD    float64 `json:"current_price"`
		MarketCap   float64 `json:"market_cap"`
		Change24h   float64 `json:"price_change_percentage_24h"`
		LastUpdated string  `json:"last_updated"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("coin not found for id=%s", id)
	}

	t, _ := time.Parse(time.RFC3339, data[0].LastUpdated)
	result := &dto.CoinMarketDTO{
		ID:          data[0].ID,
		Symbol:      data[0].Symbol,
		Name:        data[0].Name,
		PriceUSD:    data[0].PriceUSD,
		MarketCap:   data[0].MarketCap,
		Change24h:   data[0].Change24h,
		LastUpdated: t,
	}

	// Phase 3: Cache the fresh data
	r.priceCache.Set(id, result)

	return result, nil
}

func (r *AssetRepositoryImpl) FetchTopMovers(limit int) ([]dto.TopMoverDTO, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1&sparkline=false", r.BaseURL, limit)

	req, _ := http.NewRequest("GET", url, nil)
	if r.APIKey != "" {
		req.Header.Add("X-CoinGecko-API-Key", r.APIKey)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []struct {
		ID        string  `json:"id"`
		Symbol    string  `json:"symbol"`
		Name      string  `json:"name"`
		Price     float64 `json:"current_price"`
		MarketCap float64 `json:"market_cap"`
		Change24h float64 `json:"price_change_percentage_24h"`
		LastUpd   string  `json:"last_updated"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var movers []dto.TopMoverDTO
	for _, d := range data {
		lastUpd, _ := parseTime(d.LastUpd)
		movers = append(movers, dto.TopMoverDTO{
			ID:          d.ID,
			Symbol:      d.Symbol,
			Name:        d.Name,
			PriceUSD:    d.Price,
			MarketCap:   d.MarketCap,
			Change24h:   d.Change24h,
			LastUpdated: lastUpd,
		})
	}
	return movers, nil
}

func (r *AssetRepositoryImpl) FetchSupportedVSCurrencies() ([]string, error) {
	url := fmt.Sprintf("%s/simple/supported_vs_currencies", r.BaseURL)

	req, _ := http.NewRequest("GET", url, nil)
	if r.APIKey != "" {
		req.Header.Add("X-CoinGecko-API-Key", r.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var currencies []string
	if err := json.NewDecoder(resp.Body).Decode(&currencies); err != nil {
		return nil, err
	}

	return currencies, nil
}

// getSandboxFallbackPrice returns simulated prices when CoinGecko rate limit is hit
// This allows sandbox trading to continue without external API dependency
func (r *AssetRepositoryImpl) getSandboxFallbackPrice(id string) (*dto.CoinMarketDTO, error) {
	// Sandbox prices (based on approximate current market values + small random variation)
	sandboxPrices := map[string]struct {
		price     float64
		marketCap float64
		change24h float64
	}{
		"bitcoin":     {115459.00, 2285000000000, 4.30},
		"ethereum":    {4164.84, 500700000000, 10.65},
		"solana":      {196.55, 93400000000, 11.18},
		"cardano":     {0.713, 25100000000, 13.40},
		"polkadot":    {3.25, 4800000000, 9.41},
		"avalanche-2": {19.87, 8200000000, 7.23},
		"polygon":     {0.352, 3400000000, 5.12},
		"chainlink":   {12.45, 7800000000, 6.89},
		"uniswap":     {6.89, 5200000000, 8.34},
		"cosmos":      {4.12, 2900000000, 4.56},
	}

	data, exists := sandboxPrices[id]
	if !exists {
		return nil, fmt.Errorf("sandbox fallback: coin not found for id=%s", id)
	}

	// Add small random variation (±0.5%) to simulate market movement
	variation := (float64(time.Now().UnixNano()%100) - 50) / 10000 // -0.5% to +0.5%
	priceWithVariation := data.price * (1 + variation)

	return &dto.CoinMarketDTO{
		ID:          id,
		Symbol:      id, // Simplified for sandbox
		Name:        id,
		PriceUSD:    priceWithVariation,
		MarketCap:   data.marketCap,
		Change24h:   data.change24h,
		LastUpdated: time.Now(),
	}, nil
}

func parseTime(t string) (time.Time, error) {
	return time.Parse(time.RFC3339, t)
}
