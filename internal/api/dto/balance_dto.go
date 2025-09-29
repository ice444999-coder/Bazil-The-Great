package dto

type BalanceDTO struct {
	UserID uint    `json:"user_id"`
	Asset  string  `json:"asset"`  // Always USD
	Amount float64 `json:"amount"`
}
