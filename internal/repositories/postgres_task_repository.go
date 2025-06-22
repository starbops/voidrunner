package repositories

import (
	"database/sql"
	"fmt"

	"github.com/starbops/voidrunner/internal/models"
	_ "github.com/lib/pq"
)

type PostgresTaskRepository struct {
	*sql.DB
}

func NewPostgresTaskRepository(dataSourceName string) (TaskRepository, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresTaskRepository{DB: db}, nil
}

func (ptr *PostgresTaskRepository) GetTasks() ([]*models.Task, error) {
	rows, err := ptr.Query("SELECT id, name, description, status, created_at FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.Name, &task.Description, &task.Status, &task.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (ptr *PostgresTaskRepository) GetTask(id int) (*models.Task, error) {
	var task models.Task
	err := ptr.QueryRow("SELECT id, name, description, status, created_at FROM tasks WHERE id = $1", id).Scan(
		&task.ID, &task.Name, &task.Description, &task.Status, &task.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task with id %d: %w", id, err)
	}
	return &task, nil
}

func (ptr *PostgresTaskRepository) CreateTask(task *models.Task) (*models.Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	err := ptr.QueryRow(
		"INSERT INTO tasks (name, description, status) VALUES ($1, $2, $3) RETURNING id",
		task.Name, task.Description, task.Status,
	).Scan(&task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

func (ptr *PostgresTaskRepository) UpdateTask(id int, task *models.Task) (*models.Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	result, err := ptr.Exec("UPDATE tasks SET name = $1, description = $2, status = $3 WHERE id = $4",
		task.Name, task.Description, task.Status, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update task with id %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected for task with id %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("no task found with id %d", id)
	}

	task.ID = id

	return task, nil
}

func (ptr *PostgresTaskRepository) DeleteTask(id int) error {
	result, err := ptr.Exec("DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete task with id %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for task with id %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no task found with id %d", id)
	}

	return nil
}
