package dto

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}


type ChatResponse struct {
	Message  string `json:"message"`
	Response string `json:"response"`
}


type ChatHistoryResponse struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Message   string `json:"message"`
	Response  string `json:"response"`
	CreatedAt string `json:"created_at"`
}
