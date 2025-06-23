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
	rows, err := ptr.Query("SELECT id, name, description, status, user_id, created_at, updated_at FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.Name, &task.Description, &task.Status, &task.UserID, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (ptr *PostgresTaskRepository) GetTask(id int) (*models.Task, error) {
	var task models.Task
	err := ptr.QueryRow("SELECT id, name, description, status, user_id, created_at, updated_at FROM tasks WHERE id = $1", id).Scan(
		&task.ID, &task.Name, &task.Description, &task.Status, &task.UserID, &task.CreatedAt, &task.UpdatedAt)
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
		"INSERT INTO tasks (name, description, status, user_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at",
		task.Name, task.Description, task.Status, task.UserID,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

func (ptr *PostgresTaskRepository) UpdateTask(id int, task *models.Task) (*models.Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	err := ptr.QueryRow("UPDATE tasks SET name = $1, description = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4 RETURNING updated_at",
		task.Name, task.Description, task.Status, id).Scan(&task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update task with id %d: %w", id, err)
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

func (ptr *PostgresTaskRepository) GetTasksByUserID(userID int) ([]*models.Task, error) {
	rows, err := ptr.Query("SELECT id, name, description, status, user_id, created_at, updated_at FROM tasks WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks for user %d: %w", userID, err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.Name, &task.Description, &task.Status, &task.UserID, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (ptr *PostgresTaskRepository) GetTaskByUserID(id, userID int) (*models.Task, error) {
	var task models.Task
	err := ptr.QueryRow("SELECT id, name, description, status, user_id, created_at, updated_at FROM tasks WHERE id = $1 AND user_id = $2", id, userID).Scan(
		&task.ID, &task.Name, &task.Description, &task.Status, &task.UserID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task with id %d for user %d: %w", id, userID, err)
	}
	return &task, nil
}

func (ptr *PostgresTaskRepository) UpdateTaskByUserID(id, userID int, task *models.Task) (*models.Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	err := ptr.QueryRow("UPDATE tasks SET name = $1, description = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4 AND user_id = $5 RETURNING updated_at",
		task.Name, task.Description, task.Status, id, userID).Scan(&task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update task with id %d for user %d: %w", id, userID, err)
	}

	task.ID = id
	task.UserID = userID

	return task, nil
}

func (ptr *PostgresTaskRepository) DeleteTaskByUserID(id, userID int) error {
	result, err := ptr.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete task with id %d for user %d: %w", id, userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for task with id %d for user %d: %w", id, userID, err)
	}
	if rowsAffected == 0 {
		return nil
	}

	return nil
}
