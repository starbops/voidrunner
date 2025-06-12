package repositories

import (
	"fmt"
	"testing"

	"example.com/internal/config" // Adjust to your actual config path
	"example.com/internal/models" // Adjust to your actual models path
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This is a stand-in for the actual MemoryTaskRepository type.
// For these factory tests, we mainly care that the factory dispatches
// to the correct constructor, and we can check the type of the returned repo.
// Assume the real NewMemoryTaskRepository returns a pointer to a struct that
// implements TaskRepository.
type mockMemoryTaskRepository struct{}

func (m *mockMemoryTaskRepository) GetTasks() ([]*models.Task, error) { return nil, nil }
func (m *mockMemoryTaskRepository) GetTask(id int) (*models.Task, error) { return nil, nil }
func (m *mockMemoryTaskRepository) CreateTask(task *models.Task) (*models.Task, error) { return nil, nil }
func (m *mockMemoryTaskRepository) UpdateTask(id int, task *models.Task) (*models.Task, error) {
	return nil, nil
}
func (m *mockMemoryTaskRepository) DeleteTask(id int) error { return nil }

// This is a placeholder for the actual NewMemoryTaskRepository constructor.
// In a real scenario, the factory would call the actual repositories.NewMemoryTaskRepository().
// For this test, we need to control what happens when "memory" is selected.
// We'll assume that the real NewMemoryTaskRepository exists in the package
// and the factory calls it. The factory itself doesn't define it.

// To test the factory's dispatch, we need to know the type returned by the actual NewMemoryTaskRepository.
// Let's assume the actual NewMemoryTaskRepository is defined elsewhere in the package:
// func NewMemoryTaskRepository() (TaskRepository, error) { return &someConcreteMemoryType{}, nil }
// We will use our mockMemoryTaskRepository as that concrete type for the purpose of this test.

// To make this testable without complex mocking of the package-level NewMemoryTaskRepository itself,
// we rely on the fact that NewTaskRepository will return whatever NewMemoryTaskRepository returns.
// If NewMemoryTaskRepository is also in this 'repositories' package, the factory calls it directly.

func TestNewTaskRepository_Memory(t *testing.T) {
	cfg := &config.Config{StorageBackend: "memory"}

	// Temporarily replace the actual NewMemoryTaskRepository if complex setup is needed
	// or if it's not available in the test environment.
	// However, the factory directly calls `return NewMemoryTaskRepository(), nil`.
	// So, we are testing that this path is taken.
	// We need to ensure that the type check `assert.IsType` matches what
	// the actual NewMemoryTaskRepository() is expected to return.
	// For this test, we assume it returns something identifiable,
	// for instance, if NewMemoryTaskRepository was:
	// func NewMemoryTaskRepository() (TaskRepository, error) { return &MemoryTaskRepository{}, nil }
	// then we would assert for `*MemoryTaskRepository`.
	// Let's assume the real `NewMemoryTaskRepository` (not shown here) returns a type
	// that we can identify. For the sake of this example, we'll assume it returns a type
	// that, if it were our mock, would be `*mockMemoryTaskRepository`.
	// The key is that the factory itself doesn't instantiate it, it delegates.

	// To truly test this without modifying the original factory code or NewMemoryTaskRepository,
	// you'd have to know the concrete type returned by the *actual* NewMemoryTaskRepository.
	// Let's assume `NewMemoryTaskRepository` is defined in the same package and returns `*MemoryTaskRepository` (a real one).
	// The test below will pass if `NewMemoryTaskRepository` is callable and returns a non-nil repo and nil error.
	// Type assertion is the tricky part without a real `MemoryTaskRepository` type defined in this test scope.

	repo, err := NewTaskRepository(cfg)

	assert.NoError(t, err)
	require.NotNil(t, repo)
	// This assertion depends on the actual concrete type returned by NewMemoryTaskRepository.
	// If NewMemoryTaskRepository is defined in the same package and returns e.g. *ConcreteMemoryRepo,
	// you would assert.IsType(t, &ConcreteMemoryRepo{}, repo)
	// For now, just check it's not nil and no error, implying the path was taken.
	// A more robust test would involve ensuring it's not a Postgres repo, for example.
	// Or, if MemoryTaskRepository is an exported type: assert.IsType(t, &MemoryTaskRepository{}, repo)
}

func TestNewTaskRepository_Postgres_ConnectionError(t *testing.T) {
	cfg := &config.Config{
		StorageBackend: "postgres",
		PGHost:         "invalid-pg-host-for-test", // Invalid host
		PGPort:         "5432",
		PGUser:         "user",
		PGPassword:     "pw",
		PGDbName:       "db",
	}

	repo, err := NewTaskRepository(cfg)

	assert.Error(t, err) // Expect an error from NewPostgresTaskRepository due to bad DSN
	assert.Nil(t, repo)
	// We can check that the error message is consistent with a connection failure
	// and not an "invalid backend" error from the factory itself.
	assert.NotContains(t, err.Error(), "invalid storage backend")
}

func TestNewTaskRepository_InvalidBackend(t *testing.T) {
	cfg := &config.Config{StorageBackend: "non_existent_backend"}
	repo, err := NewTaskRepository(cfg)

	assert.Error(t, err)
	assert.Nil(t, repo)
	expectedErr := fmt.Sprintf("invalid storage backend: %s. Supported backends are 'postgres' and 'memory'", cfg.StorageBackend)
	assert.EqualError(t, err, expectedErr)
}

func TestNewTaskRepository_DefaultToMemory_EmptyBackend(t *testing.T) {
	cfg := &config.Config{StorageBackend: ""} // Empty string should default to memory
	repo, err := NewTaskRepository(cfg)

	assert.NoError(t, err)
	require.NotNil(t, repo)
	// Similar to TestNewTaskRepository_Memory, asserting the exact type depends on
	// the concrete type returned by the actual NewMemoryTaskRepository.
	// For now, assert it's not nil and no error.
}
