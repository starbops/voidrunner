package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Unset known environment variables to ensure defaults are tested
	os.Unsetenv("STORAGE_BACKEND")
	os.Unsetenv("PG_HOST")
	os.Unsetenv("PG_PORT")
	os.Unsetenv("PG_USER")
	os.Unsetenv("PG_PASSWORD")
	os.Unsetenv("PG_DBNAME")

	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "memory", cfg.StorageBackend, "Default StorageBackend should be memory")
	assert.Equal(t, "localhost", cfg.PGHost, "Default PGHost should be localhost")
	assert.Equal(t, "5432", cfg.PGPort, "Default PGPort should be 5432")
	assert.Equal(t, "", cfg.PGUser, "Default PGUser should be empty")       // Default for PGUser is empty
	assert.Equal(t, "", cfg.PGPassword, "Default PGPassword should be empty") // Default for PGPassword is empty
	assert.Equal(t, "", cfg.PGDbName, "Default PGDbName should be empty")     // Default for PGDbName is empty
}

func TestLoadConfig_Postgres_Success(t *testing.T) {
	// Set environment variables for PostgreSQL
	os.Setenv("STORAGE_BACKEND", "postgres")
	os.Setenv("PG_HOST", "testhost")
	os.Setenv("PG_PORT", "1234")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpassword")
	os.Setenv("PG_DBNAME", "testdb")

	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "postgres", cfg.StorageBackend)
	assert.Equal(t, "testhost", cfg.PGHost)
	assert.Equal(t, "1234", cfg.PGPort)
	assert.Equal(t, "testuser", cfg.PGUser)
	assert.Equal(t, "testpassword", cfg.PGPassword)
	assert.Equal(t, "testdb", cfg.PGDbName)

	// Clean up environment variables
	os.Unsetenv("STORAGE_BACKEND")
	os.Unsetenv("PG_HOST")
	os.Unsetenv("PG_PORT")
	os.Unsetenv("PG_USER")
	os.Unsetenv("PG_PASSWORD")
	os.Unsetenv("PG_DBNAME")
}

func TestLoadConfig_Postgres_MissingVars(t *testing.T) {
	tests := []struct {
		name          string
		unsetVar      string // The variable to unset for this test case
		expectedError string
	}{
		{
			name:          "Missing PG_HOST",
			unsetVar:      "PG_HOST",
			expectedError: "missing required PostgreSQL configuration: PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DBNAME must be set when STORAGE_BACKEND is postgres",
		},
		{
			name:          "Missing PG_PORT",
			unsetVar:      "PG_PORT",
			expectedError: "missing required PostgreSQL configuration: PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DBNAME must be set when STORAGE_BACKEND is postgres",
		},
		{
			name:          "Missing PG_USER",
			unsetVar:      "PG_USER",
			expectedError: "missing required PostgreSQL configuration: PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DBNAME must be set when STORAGE_BACKEND is postgres",
		},
		{
			name:          "Missing PG_PASSWORD",
			unsetVar:      "PG_PASSWORD",
			expectedError: "missing required PostgreSQL configuration: PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DBNAME must be set when STORAGE_BACKEND is postgres",
		},
		{
			name:          "Missing PG_DBNAME",
			unsetVar:      "PG_DBNAME",
			expectedError: "missing required PostgreSQL configuration: PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DBNAME must be set when STORAGE_BACKEND is postgres",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set necessary PG vars, then unset the one for the test case
			os.Setenv("STORAGE_BACKEND", "postgres")
			os.Setenv("PG_HOST", "testhost")
			os.Setenv("PG_PORT", "1234")
			os.Setenv("PG_USER", "testuser")
			os.Setenv("PG_PASSWORD", "testpassword")
			os.Setenv("PG_DBNAME", "testdb")

			// Unset the specific variable for this sub-test
			if tt.unsetVar != "" {
				os.Unsetenv(tt.unsetVar)
			}

			cfg, err := LoadConfig()

			assert.Error(t, err)
			assert.Nil(t, cfg)
			assert.EqualError(t, err, tt.expectedError)

			// Clean up all vars
			os.Unsetenv("STORAGE_BACKEND")
			os.Unsetenv("PG_HOST")
			os.Unsetenv("PG_PORT")
			os.Unsetenv("PG_USER")
			os.Unsetenv("PG_PASSWORD")
			os.Unsetenv("PG_DBNAME")
		})
	}
}

func TestLoadConfig_Postgres_AllPGVarsSetButBackendNotPostgres(t *testing.T) {
	// Set all PG vars
	os.Setenv("PG_HOST", "testhost")
	os.Setenv("PG_PORT", "1234")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpassword")
	os.Setenv("PG_DBNAME", "testdb")

	// But set STORAGE_BACKEND to memory (or leave it default)
	os.Setenv("STORAGE_BACKEND", "memory")

	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "memory", cfg.StorageBackend)
	// PG vars should still be loaded, even if not used by memory backend
	assert.Equal(t, "testhost", cfg.PGHost)
	assert.Equal(t, "1234", cfg.PGPort)
	assert.Equal(t, "testuser", cfg.PGUser)
	assert.Equal(t, "testpassword", cfg.PGPassword)
	assert.Equal(t, "testdb", cfg.PGDbName)


	// Clean up environment variables
	os.Unsetenv("STORAGE_BACKEND")
	os.Unsetenv("PG_HOST")
	os.Unsetenv("PG_PORT")
	os.Unsetenv("PG_USER")
	os.Unsetenv("PG_PASSWORD")
	os.Unsetenv("PG_DBNAME")
}
