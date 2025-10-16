package config

import (
	"os"
	"strconv"
	"strings"
)

// FeatureFlags controls which features are enabled
type FeatureFlags struct {
	// Rate limiting
	RateLimitingEnabled bool
	RateLimitPerMinute  int

	// Monitoring
	MonitoringEnabled    bool
	MetricsPort          int
	HealthCheckEndpoint  string

	// Trading
	SandboxMode          bool
	MaxPositionSize      float64
	MaxOpenTrades        int

	// Database
	DatabasePersistence  bool
	AutoArchiveDays      int

	// Security
	JWTAuthRequired      bool
	FileAccessWhitelist  []string

	// LLM
	CircuitBreakerEnabled bool
	MaxRetries            int
}

// DefaultFeatureFlags returns default feature flags for single-user mode
func DefaultFeatureFlags() *FeatureFlags {
	return &FeatureFlags{
		// Rate limiting OFF for single user
		RateLimitingEnabled: getEnvBool("FEATURE_RATE_LIMITING", false),
		RateLimitPerMinute:  getEnvInt("RATE_LIMIT_PER_MINUTE", 1000),

		// Monitoring ON (always useful)
		MonitoringEnabled:    getEnvBool("FEATURE_MONITORING", true),
		MetricsPort:          getEnvInt("METRICS_PORT", 9090),
		HealthCheckEndpoint:  getEnv("HEALTH_CHECK_ENDPOINT", "/health"),

		// Trading defaults
		SandboxMode:          getEnvBool("SANDBOX_MODE", true),
		MaxPositionSize:      getEnvFloat("MAX_POSITION_SIZE", 10000.0),
		MaxOpenTrades:        getEnvInt("MAX_OPEN_TRADES", 10),

		// Database persistence ON (critical)
		DatabasePersistence:  getEnvBool("FEATURE_DATABASE_PERSISTENCE", true),
		AutoArchiveDays:      getEnvInt("AUTO_ARCHIVE_DAYS", 90),

		// Security
		JWTAuthRequired:      getEnvBool("JWT_AUTH_REQUIRED", false), // OFF for single user
		FileAccessWhitelist:  getEnvList("FILE_ACCESS_WHITELIST", []string{"C:/ARES_Workspace"}),

		// LLM
		CircuitBreakerEnabled: getEnvBool("FEATURE_CIRCUIT_BREAKER", true),
		MaxRetries:            getEnvInt("LLM_MAX_RETRIES", 3),
	}
}

// Helper functions to read from environment
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolVal, err := strconv.ParseBool(value)
		if err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intVal, err := strconv.Atoi(value)
		if err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		floatVal, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvList(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// IsFeatureEnabled checks if a specific feature is enabled
func (ff *FeatureFlags) IsFeatureEnabled(feature string) bool {
	switch feature {
	case "rate_limiting":
		return ff.RateLimitingEnabled
	case "monitoring":
		return ff.MonitoringEnabled
	case "database_persistence":
		return ff.DatabasePersistence
	case "circuit_breaker":
		return ff.CircuitBreakerEnabled
	case "jwt_auth":
		return ff.JWTAuthRequired
	default:
		return false
	}
}

// GetFeatureConfig returns configuration for a feature
func (ff *FeatureFlags) GetFeatureConfig(feature string) interface{} {
	switch feature {
	case "rate_limiting":
		return map[string]interface{}{
			"enabled":     ff.RateLimitingEnabled,
			"per_minute":  ff.RateLimitPerMinute,
		}
	case "monitoring":
		return map[string]interface{}{
			"enabled":     ff.MonitoringEnabled,
			"port":        ff.MetricsPort,
			"endpoint":    ff.HealthCheckEndpoint,
		}
	case "trading":
		return map[string]interface{}{
			"sandbox_mode":      ff.SandboxMode,
			"max_position_size": ff.MaxPositionSize,
			"max_open_trades":   ff.MaxOpenTrades,
		}
	default:
		return nil
	}
}
