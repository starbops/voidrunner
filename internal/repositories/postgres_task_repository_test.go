package repositories

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/starbops/voidrunner/internal/models"
)

func TestPostgresTaskRepository_GetTasks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	rows := sqlmock.NewRows([]string{"id", "name", "description", "status", "user_id", "created_at", "updated_at"}).
		AddRow(1, "Task 1", "Description 1", "pending", 1, "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z").
		AddRow(2, "Task 2", "Description 2", "completed", 1, "2023-01-02T00:00:00Z", "2023-01-02T00:00:00Z")

	mock.ExpectQuery("SELECT id, name, description, status, user_id, created_at, updated_at FROM tasks").
		WillReturnRows(rows)

	tasks, err := repo.GetTasks()
	if err != nil {
		t.Fatalf("GetTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("GetTasks() returned %v tasks, want 2", len(tasks))
	}
	if tasks[0].Name != "Task 1" {
		t.Errorf("GetTasks() first task name = %v, want %v", tasks[0].Name, "Task 1")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_GetTasks_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	mock.ExpectQuery("SELECT id, name, description, status, user_id, created_at, updated_at FROM tasks").
		WillReturnError(sql.ErrConnDone)

	_, err = repo.GetTasks()
	if err == nil {
		t.Error("GetTasks() expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_GetTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	row := sqlmock.NewRows([]string{"id", "name", "description", "status", "user_id", "created_at", "updated_at"}).
		AddRow(1, "Test Task", "Test Description", "pending", 1, "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")

	mock.ExpectQuery("SELECT id, name, description, status, user_id, created_at, updated_at FROM tasks WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(row)

	task, err := repo.GetTask(1)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if task.ID != 1 {
		t.Errorf("GetTask() ID = %v, want %v", task.ID, 1)
	}
	if task.Name != "Test Task" {
		t.Errorf("GetTask() Name = %v, want %v", task.Name, "Test Task")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_GetTask_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	mock.ExpectQuery("SELECT id, name, description, status, user_id, created_at, updated_at FROM tasks WHERE id = \\$1").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	task, err := repo.GetTask(999)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task != nil {
		t.Error("GetTask() should return nil for non-existent task")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_CreateTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	task := &models.Task{
		Name:        "New Task",
		Description: "New Description",
		Status:      models.TaskStatusPending,
		UserID:      1,
	}

	mock.ExpectQuery("INSERT INTO tasks \\(name, description, status, user_id\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) RETURNING id, created_at, updated_at").
		WithArgs(task.Name, task.Description, task.Status, task.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(1, "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z"))

	createdTask, err := repo.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if createdTask.ID != 1 {
		t.Errorf("CreateTask() ID = %v, want %v", createdTask.ID, 1)
	}
	if createdTask.Name != task.Name {
		t.Errorf("CreateTask() Name = %v, want %v", createdTask.Name, task.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_CreateTask_Nil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	_, err = repo.CreateTask(nil)
	if err == nil {
		t.Error("CreateTask(nil) expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_UpdateTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	task := &models.Task{
		ID:          1,
		Name:        "Updated Task",
		Description: "Updated Description",
		Status:      models.TaskStatusCompleted,
	}

	mock.ExpectQuery("UPDATE tasks SET name = \\$1, description = \\$2, status = \\$3, updated_at = CURRENT_TIMESTAMP WHERE id = \\$4 RETURNING updated_at").
		WithArgs(task.Name, task.Description, task.Status, task.ID).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow("2023-01-01T00:00:00Z"))

	updatedTask, err := repo.UpdateTask(task.ID, task)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if updatedTask.Name != task.Name {
		t.Errorf("UpdateTask() Name = %v, want %v", updatedTask.Name, task.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_UpdateTask_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	task := &models.Task{
		ID:   999,
		Name: "Non-existent Task",
	}

	mock.ExpectQuery("UPDATE tasks SET name = \\$1, description = \\$2, status = \\$3, updated_at = CURRENT_TIMESTAMP WHERE id = \\$4 RETURNING updated_at").
		WithArgs(task.Name, task.Description, task.Status, task.ID).
		WillReturnError(sql.ErrNoRows)

	updatedTask, err := repo.UpdateTask(task.ID, task)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if updatedTask != nil {
		t.Error("UpdateTask() should return nil for non-existent task")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_UpdateTask_Nil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	_, err = repo.UpdateTask(1, nil)
	if err == nil {
		t.Error("UpdateTask(nil) expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_DeleteTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	mock.ExpectExec("DELETE FROM tasks WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteTask(1)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestPostgresTaskRepository_DeleteTask_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			_ = err // Log error if needed, but don't fail the test
		}
	}()

	repo := &PostgresTaskRepository{DB: db}

	mock.ExpectExec("DELETE FROM tasks WHERE id = \\$1").
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.DeleteTask(999)
	if err == nil {
		t.Error("DeleteTask() expected error for non-existent task, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}