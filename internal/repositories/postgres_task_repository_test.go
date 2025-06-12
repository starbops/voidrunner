package repositories

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp" // Required for sqlmock query matching
	"testing"

	"example.com/internal/models" // Adjust if your models path is different
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a new mock DB and repository
func newMockDBAndRepo(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *PostgresTaskRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	// In NewPostgresTaskRepository, we added a CREATE TABLE IF NOT EXISTS query.
	// We need to expect this query during the setup for tests that call NewPostgresTaskRepository.
	// However, for unit testing the repository methods, we are directly instantiating PostgresTaskRepository
	// with the mocked *sql.DB, so the NewPostgresTaskRepository function (and thus the CREATE TABLE query)
	// is not actually called in these unit tests.
	// If tests were to call NewPostgresTaskRepository, we would add:
	// mock.ExpectExec("CREATE TABLE IF NOT EXISTS tasks").WillReturnResult(sqlmock.NewResult(0, 0))

	repo := &PostgresTaskRepository{DB: db}
	return db, mock, repo
}

func TestCreateTask(t *testing.T) {
	db, mock, repo := newMockDBAndRepo(t)
	defer db.Close()

	taskToCreate := &models.Task{Title: "Test Task", Description: "Test Description", Status: "Pending"}
	expectedID := 1

	// Mock QueryRow for insert
	query := "INSERT INTO tasks (title, description, status) VALUES ($1, $2, $3) RETURNING id"
	mock.ExpectQuery(regexp.QuoteMeta(query)). // Use regexp.QuoteMeta for literal matching
		WithArgs(taskToCreate.Title, taskToCreate.Description, taskToCreate.Status).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	createdTask, err := repo.CreateTask(taskToCreate)

	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, expectedID, createdTask.ID)
	assert.Equal(t, taskToCreate.Title, createdTask.Title)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test DB error
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(taskToCreate.Title, taskToCreate.Description, taskToCreate.Status).
		WillReturnError(errors.New("db error"))

	_, err = repo.CreateTask(taskToCreate)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to create task: db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTask(t *testing.T) {
	db, mock, repo := newMockDBAndRepo(t)
	defer db.Close()

	taskID := 1
	expectedTask := &models.Task{ID: taskID, Title: "Found Task", Description: "Desc", Status: "Done"}

	query := "SELECT id, title, description, status FROM tasks WHERE id = $1"
	rows := sqlmock.NewRows([]string{"id", "title", "description", "status"}).
		AddRow(expectedTask.ID, expectedTask.Title, expectedTask.Description, expectedTask.Status)

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(taskID).WillReturnRows(rows)

	task, err := repo.GetTask(taskID)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask, task)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test sql.ErrNoRows
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(taskID).WillReturnError(sql.ErrNoRows)
	_, err = repo.GetTask(taskID)
	assert.Error(t, err)
	assert.EqualError(t, err, "task with id 1 not found")
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test other DB error
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(taskID).WillReturnError(errors.New("db error"))
	_, err = repo.GetTask(taskID)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to query task with id 1: db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTasks(t *testing.T) {
	db, mock, repo := newMockDBAndRepo(t)
	defer db.Close()

	query := "SELECT id, title, description, status FROM tasks"

	// Case 1: Multiple tasks found
	expectedTasks := []*models.Task{
		{ID: 1, Title: "Task 1", Description: "Desc 1", Status: "Pending"},
		{ID: 2, Title: "Task 2", Description: "Desc 2", Status: "Completed"},
	}
	rows := sqlmock.NewRows([]string{"id", "title", "description", "status"}).
		AddRow(expectedTasks[0].ID, expectedTasks[0].Title, expectedTasks[0].Description, expectedTasks[0].Status).
		AddRow(expectedTasks[1].ID, expectedTasks[1].Title, expectedTasks[1].Description, expectedTasks[1].Status)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows)

	tasks, err := repo.GetTasks()
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, expectedTasks, tasks)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case 2: No tasks found
	emptyRows := sqlmock.NewRows([]string{"id", "title", "description", "status"})
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(emptyRows)
	tasks, err = repo.GetTasks()
	assert.NoError(t, err)
	assert.Len(t, tasks, 0)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case 3: Database error on Query
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errors.New("db query error"))
	_, err = repo.GetTasks()
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to query tasks: db query error")
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case 4: Error during row scanning
	scanErrorRows := sqlmock.NewRows([]string{"id", "title"}).AddRow(1, "only two cols") // Mismatch columns
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(scanErrorRows)
	_, err = repo.GetTasks()
	assert.Error(t, err) // Error message will be specific to the scan error
	assert.Contains(t, err.Error(), "failed to scan task row:")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTask(t *testing.T) {
	db, mock, repo := newMockDBAndRepo(t)
	defer db.Close()

	taskToUpdate := &models.Task{Title: "Updated Task", Description: "Updated Desc", Status: "Completed"}
	taskID := 1

	query := "UPDATE tasks SET title = $1, description = $2, status = $3 WHERE id = $4"

	// Case 1: Successful update
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(taskToUpdate.Title, taskToUpdate.Description, taskToUpdate.Status, taskID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	updatedTask, err := repo.UpdateTask(taskID, taskToUpdate)
	assert.NoError(t, err)
	assert.NotNil(t, updatedTask)
	assert.Equal(t, taskID, updatedTask.ID) // ID should be set correctly
	assert.Equal(t, taskToUpdate.Title, updatedTask.Title)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case 2: Task not found (0 rows affected)
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(taskToUpdate.Title, taskToUpdate.Description, taskToUpdate.Status, taskID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	_, err = repo.UpdateTask(taskID, taskToUpdate)
	assert.Error(t, err)
	assert.EqualError(t, err, "task with id 1 not found for update")
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case 3: Database error on Exec
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(taskToUpdate.Title, taskToUpdate.Description, taskToUpdate.Status, taskID).
		WillReturnError(errors.New("db exec error"))

	_, err = repo.UpdateTask(taskID, taskToUpdate)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to update task with id 1: db exec error")
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case 4: Error on RowsAffected() (less common, but for completeness)
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(taskToUpdate.Title, taskToUpdate.Description, taskToUpdate.Status, taskID).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))
	_, err = repo.UpdateTask(taskID, taskToUpdate)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to get rows affected for update task with id 1: rows affected error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteTask(t *testing.T) {
	db, mock, repo := newMockDBAndRepo(t)
	defer db.Close()

	taskID := 1
	query := "DELETE FROM tasks WHERE id = $1"

	// Case 1: Successful delete
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(taskID).WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.DeleteTask(taskID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case 2: Task not found (0 rows affected)
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(taskID).WillReturnResult(sqlmock.NewResult(0, 0))
	err = repo.DeleteTask(taskID)
	assert.Error(t, err)
	assert.EqualError(t, err, "task with id 1 not found for deletion")
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case 3: Database error on Exec
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(taskID).WillReturnError(errors.New("db exec error"))
	err = repo.DeleteTask(taskID)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to delete task with id 1: db exec error")
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case 4: Error on RowsAffected()
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(taskID).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))
	err = repo.DeleteTask(taskID)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to get rows affected for delete task with id 1: rows affected error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper to satisfy driver.Valuer interface for sqlmock with WithArgs
// This might be needed if you pass custom types as arguments, not strictly for basic types.
// For basic types like string and int, sqlmock handles it.
// type AnyArgument struct{}

// func (a AnyArgument) Match(v driver.Value) bool {
// 	return true
// }
type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	// For time.Time, you might check the type or a range
	// For this example, we'll just accept any value if it's used for a timestamp.
	_, ok := v.(string) // Assuming timestamps are passed as strings or time.Time
	if !ok {
		_, ok = v.(int64) // Or int64 for Unix timestamps
	}
	return ok
	// return true // Simpler: always match, relying on query structure for correctness
}
