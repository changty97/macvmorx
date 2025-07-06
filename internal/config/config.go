package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all application-wide configuration settings.
type Config struct {
	WebPort                       string
	OfflineTimeout                time.Duration
	MonitorInterval               time.Duration
	GitHubWebhookSecret           string // GitHub Webhook secret for validation
	GitHubRunnerRegistrationToken string // Static token for runner registration

	// Removed mTLS Configuration
}

// LoadConfig loads configuration from environment variables or uses default values.
func LoadConfig() *Config {
	cfg := &Config{
		WebPort:                       getEnv("MACVMORX_WEB_PORT", "8080"),
		OfflineTimeout:                getEnvDuration("MACVMORX_OFFLINE_TIMEOUT", 45*time.Second),
		MonitorInterval:               getEnvDuration("MACVMORX_MONITOR_INTERVAL", 5*time.Second),
		GitHubWebhookSecret:           getEnv("GITHUB_WEBHOOK_SECRET", ""),
		GitHubRunnerRegistrationToken: getEnv("GITHUB_RUNNER_REGISTRATION_TOKEN", ""),

		// Removed mTLS Configuration Defaults
	}
	log.Printf("Loaded configuration: %+v", cfg)
	return cfg
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvDuration retrieves a duration environment variable or returns a default value.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			log.Printf("Warning: Could not parse duration for %s='%s', using default %v. Error: %v", key, value, defaultValue, err)
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}

// getEnvInt64 retrieves an int64 environment variable or returns a default value.
func getEnvInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Printf("Warning: Could not parse int64 for %s='%s', using default %d. Error: %v", key, value, defaultValue, err)
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}
