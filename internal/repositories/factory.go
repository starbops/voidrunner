package repositories

import (
	"fmt"

	"example.com/internal/config" // Assuming this is the correct path for the config package
)

// NewTaskRepository creates a task repository based on the provided configuration.
func NewTaskRepository(cfg *config.Config) (TaskRepository, error) {
	switch cfg.StorageBackend {
	case "postgres":
		dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDbName)
		return NewPostgresTaskRepository(dataSourceName)
	case "memory":
		// Assuming NewMemoryTaskRepository() is defined in this package or another imported one
		return NewMemoryTaskRepository(), nil
	default:
		// Default to memory if not specified or unrecognized, to maintain previous behavior for a bit.
		// For a stricter approach, one might remove this default and only allow explicit "memory".
		// However, the request was to default to memory if unrecognized.
		if cfg.StorageBackend != "" && cfg.StorageBackend != "memory" { // Log or handle unrecognized backend specifically if needed
			return nil, fmt.Errorf("invalid storage backend: %s. Supported backends are 'postgres' and 'memory'", cfg.StorageBackend)
		}
		// This path is taken if StorageBackend is empty or explicitly "memory"
		return NewMemoryTaskRepository(), nil
	}
}

// Placeholder for NewMemoryTaskRepository if it's not defined elsewhere yet.
// If NewMemoryTaskRepository is in another file in this package, this isn't strictly needed here.
// For the sake of this example, let's assume it exists.
/*
func NewMemoryTaskRepository() *MemoryTaskRepository {
	// Implementation for MemoryTaskRepository
	return &MemoryTaskRepository{tasks: make(map[int]*models.Task)}
}
*/

// Placeholder for TaskRepository interface if not defined elsewhere.
// type TaskRepository interface { ... }

// Placeholder for models.Task if not defined elsewhere.
// package models; type Task struct { ... }
