/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package auth

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret     []byte
	refreshSecret []byte
	once          sync.Once
)

// initSecrets initializes JWT secrets from environment (lazy-loaded)
func initSecrets() {
	once.Do(func() {
		jwtSecret = []byte(os.Getenv("JWT_SECRET"))
		refreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))

		if len(jwtSecret) == 0 {
			log.Println("⚠️  WARNING: JWT_SECRET is empty! Using fallback (INSECURE)")
			jwtSecret = []byte("fallback-secret-change-me")
		}
		if len(refreshSecret) == 0 {
			log.Println("⚠️  WARNING: JWT_REFRESH_SECRET is empty! Using JWT_SECRET as fallback")
			refreshSecret = jwtSecret
		}

		log.Printf("✅ JWT secrets initialized (JWT_SECRET: %d bytes, REFRESH_SECRET: %d bytes)", len(jwtSecret), len(refreshSecret))
	})
}

// Claims defines JWT claims for access token
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// RefreshClaims defines JWT claims for refresh token
type RefreshClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT generates an access token (short-lived)
func GenerateJWT(userID uint) (string, error) {
	initSecrets() // Ensure secrets are loaded

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 min
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ares-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateRefreshToken generates a refresh token (long-lived)
func GenerateRefreshToken(userID uint) (string, error) {
	initSecrets() // Ensure secrets are loaded

	claims := &RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ares-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecret)
}

// ValidateJWT validates access token
func ValidateJWT(tokenStr string) (*Claims, error) {
	initSecrets() // Ensure secrets are loaded

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid access token")
}

// ValidateRefreshToken validates refresh token
func ValidateRefreshToken(tokenStr string) (*RefreshClaims, error) {
	initSecrets() // Ensure secrets are loaded

	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return refreshSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}
