package repositories

import (
	"testing"

	"github.com/starbops/voidrunner/internal/models"
)

func TestMemoryTaskRepository_CreateTask(t *testing.T) {
	repo := NewMemoryTaskRepository()

	task := &models.Task{
		Name:        "Test Task",
		Description: "Test Description",
		Status:      models.TaskStatusPending,
	}

	createdTask, err := repo.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if createdTask.ID == 0 {
		t.Error("CreateTask() should assign an ID")
	}
	if createdTask.Name != task.Name {
		t.Errorf("CreateTask() Name = %v, want %v", createdTask.Name, task.Name)
	}
	if createdTask.CreatedAt == "" {
		t.Error("CreateTask() should set CreatedAt")
	}
}

func TestMemoryTaskRepository_CreateTask_NilTask(t *testing.T) {
	repo := NewMemoryTaskRepository()

	createdTask, err := repo.CreateTask(nil)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if createdTask != nil {
		t.Error("CreateTask(nil) should return nil")
	}
}

func TestMemoryTaskRepository_GetTask(t *testing.T) {
	repo := NewMemoryTaskRepository()

	task := &models.Task{
		Name:        "Test Task",
		Description: "Test Description",
		Status:      models.TaskStatusPending,
	}

	createdTask, _ := repo.CreateTask(task)

	retrievedTask, err := repo.GetTask(createdTask.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if retrievedTask.ID != createdTask.ID {
		t.Errorf("GetTask() ID = %v, want %v", retrievedTask.ID, createdTask.ID)
	}
	if retrievedTask.Name != createdTask.Name {
		t.Errorf("GetTask() Name = %v, want %v", retrievedTask.Name, createdTask.Name)
	}
}

func TestMemoryTaskRepository_GetTask_NotFound(t *testing.T) {
	repo := NewMemoryTaskRepository()

	task, err := repo.GetTask(999)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task != nil {
		t.Error("GetTask() should return nil for non-existent task")
	}
}

func TestMemoryTaskRepository_GetTasks(t *testing.T) {
	repo := NewMemoryTaskRepository()

	task1 := &models.Task{Name: "Task 1", Status: models.TaskStatusPending}
	task2 := &models.Task{Name: "Task 2", Status: models.TaskStatusCompleted}

	_, err := repo.CreateTask(task1)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	_, err = repo.CreateTask(task2)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	tasks, err := repo.GetTasks()
	if err != nil {
		t.Fatalf("GetTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("GetTasks() returned %v tasks, want 2", len(tasks))
	}
}

func TestMemoryTaskRepository_UpdateTask(t *testing.T) {
	repo := NewMemoryTaskRepository()

	task := &models.Task{
		Name:        "Test Task",
		Description: "Test Description",
		Status:      models.TaskStatusPending,
	}

	createdTask, _ := repo.CreateTask(task)

	updatedTask := &models.Task{
		ID:          createdTask.ID,
		Name:        "Updated Task",
		Description: "Updated Description",
		Status:      models.TaskStatusCompleted,
	}

	result, err := repo.UpdateTask(createdTask.ID, updatedTask)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if result.Name != "Updated Task" {
		t.Errorf("UpdateTask() Name = %v, want %v", result.Name, "Updated Task")
	}
	if result.Status != models.TaskStatusCompleted {
		t.Errorf("UpdateTask() Status = %v, want %v", result.Status, models.TaskStatusCompleted)
	}
}

func TestMemoryTaskRepository_UpdateTask_NotFound(t *testing.T) {
	repo := NewMemoryTaskRepository()

	task := &models.Task{
		ID:   999,
		Name: "Non-existent Task",
	}

	result, err := repo.UpdateTask(999, task)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if result != nil {
		t.Error("UpdateTask() should return nil for non-existent task")
	}
}

func TestMemoryTaskRepository_DeleteTask(t *testing.T) {
	repo := NewMemoryTaskRepository()

	task := &models.Task{
		Name:   "Test Task",
		Status: models.TaskStatusPending,
	}

	createdTask, _ := repo.CreateTask(task)

	err := repo.DeleteTask(createdTask.ID)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	deletedTask, _ := repo.GetTask(createdTask.ID)
	if deletedTask != nil {
		t.Error("Task should be deleted")
	}
}

func TestMemoryTaskRepository_DeleteTask_NotFound(t *testing.T) {
	repo := NewMemoryTaskRepository()

	err := repo.DeleteTask(999)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}
}