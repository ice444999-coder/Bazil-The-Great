package dto

type MarketOrderRequest struct {
	CoinID   string  `json:"coin_id" binding:"required"`
	Currency string  `json:"currency" binding:"required"`
	Symbol   string  `json:"symbol" binding:"required"`
	Side     string  `json:"side" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required"`

}

type LimitOrderRequest struct {
    CoinID    string  `json:"coin_id"`
    Symbol    string  `json:"symbol"`
    Side      string  `json:"side"`
    Quantity  float64 `json:"quantity"`
    LimitPrice float64 `json:"limit_price"`
    Currency  string  `json:"currency"`
}


type TradeResponse struct {
	ID       uint    `json:"id"`
	UserID   uint    `json:"user_id"`
	CoinID   string  `json:"coin_id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Type     string  `json:"type"`
	Status   string  `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
