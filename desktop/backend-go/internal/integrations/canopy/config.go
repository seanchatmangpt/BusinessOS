package canopy

import (
	"os"
	"time"
)

// Config holds the Canopy integration configuration.
type Config struct {
	// BaseURL is the Canopy backend base URL (env CANOPY_BASE_URL).
	BaseURL string
	// SharedSecret is the HMAC secret for X-BOS-Secret header (env CANOPY_BOS_SECRET).
	SharedSecret string
	// Timeout is the HTTP client timeout per attempt.
	Timeout time.Duration
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int
}

// LoadConfig reads config from environment variables.
// Returns nil if CANOPY_BASE_URL or CANOPY_BOS_SECRET are not set.
func LoadConfig() *Config {
	baseURL := os.Getenv("CANOPY_BASE_URL")
	secret := os.Getenv("CANOPY_BOS_SECRET")
	if baseURL == "" || secret == "" {
		return nil
	}
	return &Config{
		BaseURL:      baseURL,
		SharedSecret: secret,
		Timeout:      10 * time.Second,
		MaxRetries:   2,
	}
}
