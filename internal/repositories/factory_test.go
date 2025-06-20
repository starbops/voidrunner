package repositories

import (
	"testing"

	"github.com/starbops/voidrunner/pkg/config"
)

func TestNewTaskRepository_Memory(t *testing.T) {
	cfg := &config.Config{
		StorageBackend: config.StorageBackendMemory,
	}

	repo, err := NewTaskRepository(cfg)
	if err != nil {
		t.Fatalf("NewTaskRepository() error = %v", err)
	}

	if repo == nil {
		t.Error("NewTaskRepository() should return a repository")
	}

	if _, ok := repo.(*MemoryTaskRepository); !ok {
		t.Error("NewTaskRepository() should return MemoryTaskRepository for memory backend")
	}
}

func TestNewTaskRepository_InvalidBackend(t *testing.T) {
	cfg := &config.Config{
		StorageBackend: "invalid",
	}

	_, err := NewTaskRepository(cfg)
	if err == nil {
		t.Error("NewTaskRepository() should return error for invalid backend")
	}
}

func TestNewTaskRepository_EmptyBackend(t *testing.T) {
	cfg := &config.Config{
		StorageBackend: "",
	}

	repo, err := NewTaskRepository(cfg)
	if err != nil {
		t.Fatalf("NewTaskRepository() error = %v", err)
	}

	if _, ok := repo.(*MemoryTaskRepository); !ok {
		t.Error("NewTaskRepository() should default to MemoryTaskRepository for empty backend")
	}
}