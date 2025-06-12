package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration.
type Config struct {
	StorageBackend string
	PGHost         string
	PGPort         string
	PGUser         string
	PGPassword     string
	PGDbName       string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		StorageBackend: getEnv("STORAGE_BACKEND", "memory"),
		PGHost:         getEnv("PG_HOST", "localhost"),
		PGPort:         getEnv("PG_PORT", "5432"),
		PGUser:         getEnv("PG_USER", ""),
		PGPassword:     getEnv("PG_PASSWORD", ""),
		PGDbName:       getEnv("PG_DBNAME", ""),
	}

	if cfg.StorageBackend == "postgres" {
		if cfg.PGHost == "" || cfg.PGPort == "" || cfg.PGUser == "" || cfg.PGPassword == "" || cfg.PGDbName == "" {
			return nil, fmt.Errorf("missing required PostgreSQL configuration: PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DBNAME must be set when STORAGE_BACKEND is postgres")
		}
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
