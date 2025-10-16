package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SOLACEAPIKeyMiddleware protects SOLACE endpoints with API key authentication
func SOLACEAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from environment
		validAPIKey := os.Getenv("SOLACE_API_KEY")

		// If no API key is set, allow all requests (development mode)
		if validAPIKey == "" {
			c.Next()
			return
		}

		// Check for API key in header
		providedKey := c.GetHeader("X-SOLACE-API-KEY")

		// Also check Authorization header (Bearer token format)
		if providedKey == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				providedKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Also check query parameter (for browser requests)
		if providedKey == "" {
			providedKey = c.Query("api_key")
		}

		// Validate API key
		if providedKey != validAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: Invalid or missing API key",
				"hint":  "Provide API key via X-SOLACE-API-KEY header, Authorization Bearer token, or ?api_key= query param",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
