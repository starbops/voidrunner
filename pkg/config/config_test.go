package config

import (
	"os"
	"testing"
)

func TestLoadConfig_MemoryBackend(t *testing.T) {
	os.Clearenv()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.StorageBackend != StorageBackendMemory {
		t.Errorf("StorageBackend = %v, want %v", cfg.StorageBackend, StorageBackendMemory)
	}
}

func TestLoadConfig_PostgresBackend_Valid(t *testing.T) {
	env := map[string]string{
		"STORAGE_BACKEND": "postgres",
		"PG_HOST":         "testhost",
		"PG_PORT":         "5432",
		"PG_USER":         "testuser",
		"PG_PASSWORD":     "testpass",
		"PG_DBNAME":       "testdb",
	}

	for key, value := range env {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.StorageBackend != StorageBackendPostgres {
		t.Errorf("StorageBackend = %v, want %v", cfg.StorageBackend, StorageBackendPostgres)
	}
}

func TestLoadConfig_PostgresBackend_MissingConfig(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "missing PG_HOST",
			env: map[string]string{
				"STORAGE_BACKEND": "postgres",
				"PG_HOST":         "",
				"PG_PORT":         "5432",
				"PG_USER":         "testuser",
				"PG_PASSWORD":     "testpass",
				"PG_DBNAME":       "testdb",
			},
		},
		{
			name: "missing PG_USER",
			env: map[string]string{
				"STORAGE_BACKEND": "postgres",
				"PG_HOST":         "localhost",
				"PG_PORT":         "5432",
				"PG_PASSWORD":     "testpass",
				"PG_DBNAME":       "testdb",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for key, value := range tt.env {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			_, err := LoadConfig()
			if err == nil {
				t.Error("LoadConfig() expected error, got nil")
			}
		})
	}
}