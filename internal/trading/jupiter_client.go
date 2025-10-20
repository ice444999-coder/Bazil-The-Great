package trading

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// JupiterClient handles Solana DEX operations via Jupiter API
type JupiterClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string // Optional, for enhanced rate limits
}

// JupiterQuote represents a swap quote from Jupiter
type JupiterQuote struct {
	InputMint            string          `json:"inputMint"`
	OutputMint           string          `json:"outputMint"`
	Amount               string          `json:"amount"`
	OtherAmountThreshold string          `json:"otherAmountThreshold"`
	SwapMode             string          `json:"swapMode"`
	SlippageBps          int             `json:"slippageBps"`
	PlatformFee          *PlatformFee    `json:"platformFee,omitempty"`
	PriceImpactPct       string          `json:"priceImpactPct"`
	RoutePlan            []RoutePlanStep `json:"routePlan"`
	ContextSlot          uint64          `json:"contextSlot"`
	TimeTaken            float64         `json:"timeTaken"`
}

// PlatformFee represents platform fee configuration
type PlatformFee struct {
	Amount string `json:"amount"`
	FeeBps int    `json:"feeBps"`
}

// RoutePlanStep represents a step in the swap route
type RoutePlanStep struct {
	SwapInfo SwapInfo `json:"swapInfo"`
	Percent  int      `json:"percent"`
}

// SwapInfo contains swap execution details
type SwapInfo struct {
	AmmKey     string `json:"ammKey"`
	Label      string `json:"label"`
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	InAmount   string `json:"inAmount"`
	OutAmount  string `json:"outAmount"`
	FeeAmount  string `json:"feeAmount"`
	FeeMint    string `json:"feeMint"`
}

// JupiterSwapRequest represents a swap transaction request
type JupiterSwapRequest struct {
	QuoteResponse         JupiterQuoteResponse `json:"quoteResponse"`
	UserPublicKey         string               `json:"userPublicKey"`
	WrapAndUnwrapSol      bool                 `json:"wrapAndUnwrapSol"`
	UseSharedAccounts     bool                 `json:"useSharedAccounts"`
	FeeAccount            string               `json:"feeAccount,omitempty"`
	ComputeUnitLimit      int                  `json:"computeUnitLimit,omitempty"`
	ComputeUnitPriceMicro int                  `json:"computeUnitPriceMicro,omitempty"`
	Blockhash             string               `json:"blockhash,omitempty"`
	ReturnAllInstructions bool                 `json:"returnAllInstructions"`
}

// JupiterQuoteResponse is the response from quote endpoint
type JupiterQuoteResponse struct {
	InputMint            string          `json:"inputMint"`
	OutputMint           string          `json:"outputMint"`
	InAmount             string          `json:"inAmount"`
	OutAmount            string          `json:"outAmount"`
	OtherAmountThreshold string          `json:"otherAmountThreshold"`
	SwapMode             string          `json:"swapMode"`
	SlippageBps          int             `json:"slippageBps"`
	PlatformFee          *PlatformFee    `json:"platformFee,omitempty"`
	PriceImpactPct       string          `json:"priceImpactPct"`
	RoutePlan            []RoutePlanStep `json:"routePlan"`
	ContextSlot          uint64          `json:"contextSlot"`
	TimeTaken            float64         `json:"timeTaken"`
}

// JupiterSwapResponse represents the swap transaction response
type JupiterSwapResponse struct {
	SwapTransaction      string `json:"swapTransaction"`
	LastValidBlockHeight uint64 `json:"lastValidBlockHeight"`
}

// TokenInfo represents token information from Jupiter
type TokenInfo struct {
	Address  string   `json:"address"`
	Symbol   string   `json:"symbol"`
	Name     string   `json:"name"`
	Decimals int      `json:"decimals"`
	LogoURI  string   `json:"logoURI,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// NewJupiterClient creates a new Jupiter DEX client
func NewJupiterClient(apiKey string) *JupiterClient {
	return &JupiterClient{
		baseURL: "https://quote-api.jup.ag/v6",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// GetQuote gets a swap quote from Jupiter
func (jc *JupiterClient) GetQuote(inputMint, outputMint string, amount uint64, slippageBps int) (*JupiterQuote, error) {
	url := fmt.Sprintf("%s/quote?inputMint=%s&outputMint=%s&amount=%d&slippageBps=%d",
		jc.baseURL, inputMint, outputMint, amount, slippageBps)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if jc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+jc.apiKey)
	}

	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("quote API error %d: %s", resp.StatusCode, string(body))
	}

	var quote JupiterQuote
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}

	return &quote, nil
}

// GetSwapTransaction gets the swap transaction for execution
func (jc *JupiterClient) GetSwapTransaction(quoteResponse *JupiterQuoteResponse, userPublicKey string) (*JupiterSwapResponse, error) {
	swapReq := JupiterSwapRequest{
		QuoteResponse:         *quoteResponse,
		UserPublicKey:         userPublicKey,
		WrapAndUnwrapSol:      true,
		UseSharedAccounts:     true,
		ReturnAllInstructions: false,
	}

	jsonData, err := json.Marshal(swapReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal swap request: %w", err)
	}

	url := jc.baseURL + "/swap"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create swap request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if jc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+jc.apiKey)
	}

	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap transaction: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("swap API error %d: %s", resp.StatusCode, string(body))
	}

	var swapResp JupiterSwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&swapResp); err != nil {
		return nil, fmt.Errorf("failed to decode swap response: %w", err)
	}

	return &swapResp, nil
}

// GetTokens gets all supported tokens
func (jc *JupiterClient) GetTokens() ([]TokenInfo, error) {
	url := "https://token.jup.ag/all"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokens request: %w", err)
	}

	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tokens API error %d: %s", resp.StatusCode, string(body))
	}

	var tokens []TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode tokens response: %w", err)
	}

	return tokens, nil
}

// GetTokenPrice gets the price for a specific token
func (jc *JupiterClient) GetTokenPrice(tokenAddress string) (decimal.Decimal, error) {
	url := fmt.Sprintf("https://price.jup.ag/v4/price?ids=%s", tokenAddress)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to create price request: %w", err)
	}

	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return decimal.Zero, fmt.Errorf("price API error %d: %s", resp.StatusCode, string(body))
	}

	var priceResp struct {
		Data map[string]struct {
			Price string `json:"price"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return decimal.Zero, fmt.Errorf("failed to decode price response: %w", err)
	}

	if tokenData, exists := priceResp.Data[tokenAddress]; exists {
		price, err := decimal.NewFromString(tokenData.Price)
		if err != nil {
			return decimal.Zero, fmt.Errorf("failed to parse price: %w", err)
		}
		return price, nil
	}

	return decimal.Zero, fmt.Errorf("price not found for token %s", tokenAddress)
}

// ConvertToLamports converts SOL amount to lamports (1 SOL = 1e9 lamports)
func ConvertToLamports(solAmount decimal.Decimal) uint64 {
	lamports := solAmount.Mul(decimal.NewFromInt(1000000000))
	return uint64(lamports.IntPart())
}

// ConvertFromLamports converts lamports to SOL amount
func ConvertFromLamports(lamports uint64) decimal.Decimal {
	return decimal.NewFromInt(int64(lamports)).Div(decimal.NewFromInt(1000000000))
}

// Common token addresses on Solana
const (
	SOLAddress  = "So11111111111111111111111111111111111111112"
	USDCAddress = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	USDTAddress = "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"
	BTCAddress  = "9n4nbM75f5Ui33ZbPYXn59EwSgE8CGsHtAeTH5YFeJ9E"
	ETHAddress  = "7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs"
	JUPAddress  = "JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN"
)
