package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CryptoPriceController handles live cryptocurrency price data from CoinGecko
type CryptoPriceController struct {
	// Cache prices to avoid hitting CoinGecko too frequently
	cachedPrices map[string]*CryptoPrice
	lastFetch    time.Time
	cacheDuration time.Duration
}

// CryptoPrice represents cryptocurrency price data
type CryptoPrice struct {
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	CurrentPrice float64 `json:"current_price"`
	Change24h   float64 `json:"price_change_24h"`
	ChangePercent float64 `json:"price_change_percentage_24h"`
	MarketCap   float64 `json:"market_cap"`
	Volume24h   float64 `json:"total_volume"`
	High24h     float64 `json:"high_24h"`
	Low24h      float64 `json:"low_24h"`
	LastUpdated time.Time `json:"last_updated"`
}

// CoinGeckoResponse represents the CoinGecko API response
type CoinGeckoResponse []struct {
	ID                string  `json:"id"`
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	CurrentPrice      float64 `json:"current_price"`
	MarketCap         float64 `json:"market_cap"`
	TotalVolume       float64 `json:"total_volume"`
	High24h           float64 `json:"high_24h"`
	Low24h            float64 `json:"low_24h"`
	PriceChange24h    float64 `json:"price_change_24h"`
	PriceChangePercent float64 `json:"price_change_percentage_24h"`
}

// NewCryptoPriceController creates a new crypto price controller
func NewCryptoPriceController() *CryptoPriceController {
	return &CryptoPriceController{
		cachedPrices: make(map[string]*CryptoPrice),
		cacheDuration: 60 * time.Second, // Cache for 60 seconds to avoid API rate limits
	}
}

// GetPrices fetches live cryptocurrency prices
// @Summary Get live crypto prices
// @Description Get real-time cryptocurrency prices from CoinGecko API
// @Tags Trading
// @Produce json
// @Success 200 {array} CryptoPrice
// @Router /trading/prices [get]
func (cpc *CryptoPriceController) GetPrices(c *gin.Context) {
	// Check cache first
	if time.Since(cpc.lastFetch) < cpc.cacheDuration && len(cpc.cachedPrices) > 0 {
		prices := make([]CryptoPrice, 0, len(cpc.cachedPrices))
		for _, price := range cpc.cachedPrices {
			prices = append(prices, *price)
		}
		c.JSON(http.StatusOK, gin.H{
			"prices": prices,
			"cached": true,
			"last_update": cpc.lastFetch,
		})
		return
	}

	// Fetch from CoinGecko - BTC, ETH, SOL, ADA, DOT
	url := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=bitcoin,ethereum,solana,cardano,polkadot&order=market_cap_desc&sparkline=false"

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch prices from CoinGecko",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": fmt.Sprintf("CoinGecko API returned status %d", resp.StatusCode),
		})
		return
	}

	var coinGeckoData CoinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&coinGeckoData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse CoinGecko response",
			"details": err.Error(),
		})
		return
	}

	// Transform and cache
	prices := make([]CryptoPrice, 0, len(coinGeckoData))
	cpc.cachedPrices = make(map[string]*CryptoPrice)
	now := time.Now()

	for _, coin := range coinGeckoData {
		price := &CryptoPrice{
			Symbol:        coin.Symbol,
			Name:          coin.Name,
			CurrentPrice:  coin.CurrentPrice,
			Change24h:     coin.PriceChange24h,
			ChangePercent: coin.PriceChangePercent,
			MarketCap:     coin.MarketCap,
			Volume24h:     coin.TotalVolume,
			High24h:       coin.High24h,
			Low24h:        coin.Low24h,
			LastUpdated:   now,
		}
		prices = append(prices, *price)
		cpc.cachedPrices[coin.Symbol] = price
	}

	cpc.lastFetch = now

	c.JSON(http.StatusOK, gin.H{
		"prices": prices,
		"cached": false,
		"last_update": now,
	})
}
