package middleware

import (
	"ares_api/internal/auth"
	"ares_api/internal/common"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware protects routes and extracts user ID from JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.JSON(c, http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Expect header format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			common.JSON(c, http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := auth.ValidateJWT(token)
		if err != nil {
			common.JSON(c, http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set userID in context for downstream handlers
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
