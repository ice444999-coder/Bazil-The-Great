package middleware

import (
	"ares_api/internal/auth"
	"ares_api/internal/common"
	"net/http"
	"strings"
	"sync"
	"time"

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

// RateLimiter implements basic rate limiting
func RateLimiter(requests int, window time.Duration) gin.HandlerFunc {
	// Simple in-memory rate limiter (for production, use Redis)
	type client struct {
		count   int
		resetAt time.Time
	}

	clients := make(map[string]*client)
	var mu sync.Mutex

	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		ip := c.ClientIP()
		now := time.Now()

		if cl, exists := clients[ip]; exists {
			if now.After(cl.resetAt) {
				cl.count = 1
				cl.resetAt = now.Add(window)
			} else if cl.count >= requests {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
				c.Abort()
				return
			} else {
				cl.count++
			}
		} else {
			clients[ip] = &client{
				count:   1,
				resetAt: now.Add(window),
			}
		}

		c.Next()
	}
}
