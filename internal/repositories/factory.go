package repositories

import (
	"fmt"

	"github.com/starbops/voidrunner/pkg/config"
)

func NewTaskRepository(cfg *config.Config) (TaskRepository, error) {
	switch cfg.StorageBackend {
	case config.StorageBackendMemory:
		return NewMemoryTaskRepository(), nil
	case config.StorageBackendPostgres:
		dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDbName)
		return NewPostgresTaskRepository(dataSourceName)
	default:
		if cfg.StorageBackend != "" && cfg.StorageBackend != config.StorageBackendMemory {
			return nil, fmt.Errorf("invalid storage backend: %s. Supported backends are: %s, %s", cfg.StorageBackend, config.StorageBackendMemory, config.StorageBackendPostgres)
		}

		return NewMemoryTaskRepository(), nil
	}
}
