/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Server
	Port      string
	GinMode   string
	JWTSecret string

	// LLM - Ollama
	OllamaBaseURL string
	OllamaModel   string

	// LLM - OpenAI
	OpenAIApiKey  string
	OpenAIBaseURL string
	OpenAIModel   string

	// SOLACE
	SolaceAPIKey string
	SolaceUserID string

	// Workspace
	AresWorkspaceRoot string
	AresRepoPath      string

	// Redis (optional)
	RedisAddr string

	// External APIs
	CoinGeckoAPIKey  string
	WhaleAlertAPIKey string
	JupiterAPIKey    string
}

func Load() (*Config, error) {
	godotenv.Load()

	return &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5433"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "ARESISWAKING"),
		DBName:     getEnv("DB_NAME", "ares_pgvector"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Server
		Port:      getEnv("PORT", "8080"),
		GinMode:   getEnv("GIN_MODE", "release"),
		JWTSecret: getEnv("JWT_SECRET", "ares-recognition-architecture-2025"),

		// LLM - Ollama
		OllamaBaseURL: getEnv("OLLAMA_BASE_URL", "http://127.0.0.1:11434/api"),
		OllamaModel:   getEnv("OLLAMA_MODEL", "deepseek-r1:14b"),

		// LLM - OpenAI
		OpenAIApiKey:  getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:   getEnv("OPENAI_MODEL", "gpt-4o-mini"),

		// SOLACE
		SolaceAPIKey: getEnv("SOLACE_API_KEY", "solace_autonomous_entity_2025_secure"),
		SolaceUserID: getEnv("SOLACE_USER_ID", "1"),

		// Workspace
		AresWorkspaceRoot: getEnv("ARES_WORKSPACE_ROOT", "C:/ARES_Workspace"),
		AresRepoPath:      getEnv("ARES_REPO_PATH", "C:/ARES_Workspace/ARES_API"),

		// Redis (optional)
		RedisAddr: getEnv("REDIS_ADDR", ""),

		// External APIs
		CoinGeckoAPIKey:  getEnv("COINGECKO_API_KEY", ""),
		WhaleAlertAPIKey: getEnv("WHALE_ALERT_API_KEY", ""),
		JupiterAPIKey:    getEnv("JUPITER_DEX_API", ""),
	}, nil
}

func (c *Config) DBDSN() string {
	return "host=" + c.DBHost + " port=" + c.DBPort + " user=" + c.DBUser + " dbname=" + c.DBName + " password=" + c.DBPassword + " sslmode=" + c.DBSSLMode
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
