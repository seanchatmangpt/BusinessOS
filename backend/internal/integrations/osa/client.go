// Package osa provides BusinessOS integration with the MIOSA sdk-go client.
// optional: OSA enabled when OSA_SHARED_SECRET (local mode) or OSA_API_KEY (cloud mode) is set.
package osa

import (
	"errors"
	"os"
	"time"

	osa "github.com/Miosa-osa/sdk-go"
)

// Mode represents which OSA deployment to connect to.
type Mode string

const (
	ModeLocal Mode = "local"
	ModeCloud Mode = "cloud"
)

// Config holds the resolved OSA connection parameters read from the environment.
type Config struct {
	// Mode is "local" (default) or "cloud".
	Mode Mode

	// BaseURL overrides the default endpoint.
	// Local default: http://localhost:8089
	// Cloud default: https://api.miosa.ai
	BaseURL string

	// SharedSecret is the JWT signing key used in local mode.
	// Maps to OSA_SHARED_SECRET environment variable.
	SharedSecret string

	// APIKey is the MIOSA Cloud API key used in cloud mode.
	// Maps to OSA_API_KEY environment variable.
	APIKey string

	// Timeout for HTTP requests. Default: 30s.
	Timeout time.Duration
}

// NewConfigFromEnv reads OSA configuration from environment variables:
//
//	OSA_MODE          "local" (default) or "cloud"
//	OSA_BASE_URL      override endpoint URL
//	OSA_SHARED_SECRET JWT signing key (local mode)
//	OSA_API_KEY       MIOSA Cloud API key (cloud mode)
//	OSA_TIMEOUT_SEC   request timeout in seconds (default 30)
func NewConfigFromEnv() Config {
	mode := Mode(os.Getenv("OSA_MODE"))
	if mode == "" {
		mode = ModeLocal
	}

	timeout := 30 * time.Second
	if v := os.Getenv("OSA_TIMEOUT_SEC"); v != "" {
		if d, err := time.ParseDuration(v + "s"); err == nil {
			timeout = d
		}
	}

	return Config{
		Mode:         mode,
		BaseURL:      os.Getenv("OSA_BASE_URL"),
		SharedSecret: os.Getenv("OSA_SHARED_SECRET"),
		APIKey:       os.Getenv("OSA_API_KEY"),
		Timeout:      timeout,
	}
}

// IsConfigured reports whether the config has the minimum credentials needed to
// initialise a client. Returns false when neither OSA_SHARED_SECRET nor
// OSA_API_KEY are set so callers can skip OSA init silently.
func (c Config) IsConfigured() bool {
	switch c.Mode {
	case ModeCloud:
		return c.APIKey != ""
	default: // local
		return c.SharedSecret != ""
	}
}

// NewOSAClient builds an osa.Client from the supplied Config.
// Returns an error if the config is missing required fields.
func NewOSAClient(cfg Config) (osa.Client, error) {
	switch cfg.Mode {
	case ModeCloud:
		if cfg.APIKey == "" {
			return nil, errors.New("osa: OSA_API_KEY is required for cloud mode")
		}
		cloudCfg := osa.CloudConfig{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL, // sdk applies default when empty
			Timeout: cfg.Timeout,
		}
		return osa.NewCloudClient(cloudCfg)

	default: // local
		localCfg := osa.LocalConfig{
			BaseURL:      cfg.BaseURL, // sdk applies default when empty
			SharedSecret: cfg.SharedSecret,
			Timeout:      cfg.Timeout,
		}
		return osa.NewLocalClient(localCfg)
	}
}
