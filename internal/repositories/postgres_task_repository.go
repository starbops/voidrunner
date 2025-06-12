package repositories

import (
	"database/sql"
	"fmt" // Added for error formatting

	_ "github.com/lib/pq" // PostgreSQL driver
	"example.com/internal/models" // Assuming models.Task is defined here
)

// TaskRepository defines the interface for task operations.
// This is a placeholder, assuming it's defined elsewhere or will be.
type TaskRepository interface {
	GetTasks() ([]*models.Task, error)
	GetTask(id int) (*models.Task, error)
	CreateTask(task *models.Task) (*models.Task, error)
	UpdateTask(id int, task *models.Task) (*models.Task, error)
	DeleteTask(id int) error
}

// PostgresTaskRepository implements TaskRepository for PostgreSQL.
type PostgresTaskRepository struct {
	*sql.DB
}

// NewPostgresTaskRepository creates a new PostgresTaskRepository.
func NewPostgresTaskRepository(dataSourceName string) (*PostgresTaskRepository, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// SQL to create tasks table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		status VARCHAR(50) NOT NULL
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close() // Close the connection if table creation fails
		return nil, fmt.Errorf("failed to create tasks table: %w", err)
	}

	return &PostgresTaskRepository{DB: db}, nil
}

// GetTasks retrieves all tasks from the database.
func (r *PostgresTaskRepository) GetTasks() ([]*models.Task, error) {
	rows, err := r.DB.Query("SELECT id, title, description, status FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return tasks, nil
}

// GetTask retrieves a single task by its ID.
func (r *PostgresTaskRepository) GetTask(id int) (*models.Task, error) {
	task := &models.Task{}
	err := r.DB.QueryRow("SELECT id, title, description, status FROM tasks WHERE id = $1", id).Scan(&task.ID, &task.Title, &task.Description, &task.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to query task with id %d: %w", id, err)
	}
	return task, nil
}

// CreateTask creates a new task in the database.
func (r *PostgresTaskRepository) CreateTask(task *models.Task) (*models.Task, error) {
	err := r.DB.QueryRow(
		"INSERT INTO tasks (title, description, status) VALUES ($1, $2, $3) RETURNING id",
		task.Title, task.Description, task.Status,
	).Scan(&task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

// UpdateTask updates an existing task in the database.
func (r *PostgresTaskRepository) UpdateTask(id int, task *models.Task) (*models.Task, error) {
	result, err := r.DB.Exec(
		"UPDATE tasks SET title = $1, description = $2, status = $3 WHERE id = $4",
		task.Title, task.Description, task.Status, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task with id %d: %w", id, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected for update task with id %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("task with id %d not found for update", id)
	}
	task.ID = id // Ensure the ID is set on the returned task
	return task, nil
}

// DeleteTask deletes a task from the database by its ID.
func (r *PostgresTaskRepository) DeleteTask(id int) error {
	result, err := r.DB.Exec("DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete task with id %d: %w", id, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for delete task with id %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task with id %d not found for deletion", id)
	}
	return nil
}
