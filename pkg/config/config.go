package config

import (
	"fmt"
	"os"
	"time"
)

const (
	StorageBackendMemory   = "memory"
	StorageBackendPostgres = "postgres"

	DefaultPGHost = "localhost"
	DefaultPGPort = "5432"
	
	DefaultJWTSecret     = "voidrunner-secret-change-in-production"
	DefaultJWTExpiration = 24 * time.Hour
)

type Config struct {
	Port           string
	StorageBackend string
	PGHost         string
	PGPort         string
	PGUser         string
	PGPassword     string
	PGDbName       string
	JWTSecret      string
	JWTExpiration  time.Duration
	EnableDocs     bool
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		StorageBackend: getEnv("STORAGE_BACKEND", StorageBackendMemory),
		PGHost:         getEnv("PG_HOST", DefaultPGHost),
		PGPort:         getEnv("PG_PORT", DefaultPGPort),
		PGUser:         getEnv("PG_USER", ""),
		PGPassword:     getEnv("PG_PASSWORD", ""),
		PGDbName:       getEnv("PG_DBNAME", ""),
		JWTSecret:      getEnv("JWT_SECRET", DefaultJWTSecret),
		JWTExpiration:  DefaultJWTExpiration,
		EnableDocs:     getBoolEnv("ENABLE_DOCS", false),
	}

	if cfg.StorageBackend == "postgres" {
		if cfg.PGHost == "" || cfg.PGPort == "" || cfg.PGUser == "" || cfg.PGPassword == "" || cfg.PGDbName == "" {
			return nil, fmt.Errorf("missing required PostgreSQL configuration: PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DBNAME must be set when STORAGE_BACKEND is postgres")
		}
	}

	if cfg.JWTSecret == DefaultJWTSecret {
		fmt.Println("WARNING: Using default JWT secret. Please set JWT_SECRET environment variable in production.")
	}

	return cfg, nil
}
