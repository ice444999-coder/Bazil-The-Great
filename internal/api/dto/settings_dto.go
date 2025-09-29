package dto

// Save API Key
type APIKeyRequest struct {
	APIKey string `json:"api_key" binding:"required" example:"your-secret-api-key"`
}
type APIKeyResponse struct {
	Message string `json:"message" example:"API key saved successfully"`
}



