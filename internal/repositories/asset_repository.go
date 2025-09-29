package repositories

import (
	"ares_api/internal/api/dto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	repository "ares_api/internal/interfaces/repository"
)

type AssetRepositoryImpl struct {
	BaseURL string
	APIKey  string
}

func NewAssetRepository() repository.AssetRepository {
	baseURL := os.Getenv("COINGECKO_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.coingecko.com/api/v3"
	}
	return &AssetRepositoryImpl{
		BaseURL: baseURL,
		APIKey:  os.Getenv("COINGECKO_API_KEY"),
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
	url := fmt.Sprintf("%s/coins/markets?vs_currency=%s&ids=%s&order=market_cap_desc&sparkline=false",
		r.BaseURL, vsCurrency, id)

	req, _ := http.NewRequest("GET", url, nil)
	if r.APIKey != "" {
		req.Header.Add("X-CoinGecko-API-Key", r.APIKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("DEBUG CoinGecko response:", string(body)) // 👈 log raw response

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
	return &dto.CoinMarketDTO{
		ID:          data[0].ID,
		Symbol:      data[0].Symbol,
		Name:        data[0].Name,
		PriceUSD:    data[0].PriceUSD,
		MarketCap:   data[0].MarketCap,
		Change24h:   data[0].Change24h,
		LastUpdated: t,
	}, nil
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




func parseTime(t string) (time.Time, error) {
	return time.Parse(time.RFC3339, t)
}
