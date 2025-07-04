package config

import (
	"log"
	"os"
	"time"
)

// Config holds all application-wide configuration settings.
type Config struct {
	WebPort         string
	OfflineTimeout  time.Duration
	MonitorInterval time.Duration
	// Add other configuration parameters here, e.g., database connection strings,
	// image registry URLs, etc.
}

// LoadConfig loads configuration from environment variables or uses default values.
func LoadConfig() *Config {
	cfg := &Config{
		WebPort:         getEnv("MACVMORX_WEB_PORT", "8080"),
		OfflineTimeout:  getEnvDuration("MACVMORX_OFFLINE_TIMEOUT", 30*time.Second),
		MonitorInterval: getEnvDuration("MACVMORX_MONITOR_INTERVAL", 5*time.Second),
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
