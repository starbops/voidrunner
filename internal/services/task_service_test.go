package services

import (
	"errors"
	"testing"

	"github.com/starbops/voidrunner/internal/models"
)

type mockTaskRepository struct {
	tasks   map[int]*models.Task
	nextID  int
	failGet bool
}

func newMockTaskRepository() *mockTaskRepository {
	return &mockTaskRepository{
		tasks:  make(map[int]*models.Task),
		nextID: 1,
	}
}

func (m *mockTaskRepository) GetTasks() ([]*models.Task, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	tasks := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *mockTaskRepository) GetTask(id int) (*models.Task, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	task, exists := m.tasks[id]
	if !exists {
		return nil, nil
	}
	return task, nil
}

func (m *mockTaskRepository) CreateTask(task *models.Task) (*models.Task, error) {
	if task == nil {
		return nil, nil
	}
	task.ID = m.nextID
	m.nextID++
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskRepository) UpdateTask(id int, task *models.Task) (*models.Task, error) {
	if task == nil || task.ID != id {
		return nil, nil
	}
	if _, exists := m.tasks[id]; !exists {
		return nil, nil
	}
	m.tasks[id] = task
	return task, nil
}

func (m *mockTaskRepository) DeleteTask(id int) error {
	if _, exists := m.tasks[id]; !exists {
		return nil
	}
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepository) GetTasksByUserID(userID int) ([]*models.Task, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	tasks := make([]*models.Task, 0)
	for _, task := range m.tasks {
		if task.UserID == userID {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (m *mockTaskRepository) GetTaskByUserID(id, userID int) (*models.Task, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	task, exists := m.tasks[id]
	if !exists || task.UserID != userID {
		return nil, nil
	}
	return task, nil
}

func (m *mockTaskRepository) UpdateTaskByUserID(id, userID int, task *models.Task) (*models.Task, error) {
	if task == nil || task.ID != id {
		return nil, nil
	}
	existingTask, exists := m.tasks[id]
	if !exists || existingTask.UserID != userID {
		return nil, nil
	}
	task.UserID = userID
	m.tasks[id] = task
	return task, nil
}

func (m *mockTaskRepository) DeleteTaskByUserID(id, userID int) error {
	task, exists := m.tasks[id]
	if !exists || task.UserID != userID {
		return nil
	}
	delete(m.tasks, id)
	return nil
}

func TestTaskService_CreateTask(t *testing.T) {
	mockRepo := newMockTaskRepository()
	service := NewTaskService(mockRepo)

	task := &models.Task{
		Name:        "Test Task",
		Description: "Test Description",
	}

	createdTask, err := service.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if createdTask.Status != models.TaskStatusPending {
		t.Errorf("CreateTask() Status = %v, want %v", createdTask.Status, models.TaskStatusPending)
	}
	if createdTask.ID == 0 {
		t.Error("CreateTask() should assign an ID")
	}
}

func TestTaskService_CreateTask_Nil(t *testing.T) {
	mockRepo := newMockTaskRepository()
	service := NewTaskService(mockRepo)

	createdTask, err := service.CreateTask(nil)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if createdTask != nil {
		t.Error("CreateTask(nil) should return nil")
	}
}

func TestTaskService_GetTask(t *testing.T) {
	mockRepo := newMockTaskRepository()
	service := NewTaskService(mockRepo)

	task := &models.Task{
		Name:   "Test Task",
		Status: models.TaskStatusPending,
	}
	createdTask, _ := service.CreateTask(task)

	retrievedTask, err := service.GetTask(createdTask.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if retrievedTask.ID != createdTask.ID {
		t.Errorf("GetTask() ID = %v, want %v", retrievedTask.ID, createdTask.ID)
	}
}

func TestTaskService_GetTasks(t *testing.T) {
	mockRepo := newMockTaskRepository()
	service := NewTaskService(mockRepo)

	task1 := &models.Task{Name: "Task 1", Status: models.TaskStatusPending}
	task2 := &models.Task{Name: "Task 2", Status: models.TaskStatusCompleted}

	service.CreateTask(task1)
	service.CreateTask(task2)

	tasks, err := service.GetTasks()
	if err != nil {
		t.Fatalf("GetTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("GetTasks() returned %v tasks, want 2", len(tasks))
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	mockRepo := newMockTaskRepository()
	service := NewTaskService(mockRepo)

	task := &models.Task{
		Name:   "Test Task",
		Status: models.TaskStatusPending,
	}
	createdTask, _ := service.CreateTask(task)

	updatedTask := &models.Task{
		ID:     createdTask.ID,
		Name:   "Updated Task",
		Status: models.TaskStatusCompleted,
	}

	result, err := service.UpdateTask(createdTask.ID, updatedTask)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if result.Name != "Updated Task" {
		t.Errorf("UpdateTask() Name = %v, want %v", result.Name, "Updated Task")
	}
}

func TestTaskService_UpdateTask_Nil(t *testing.T) {
	mockRepo := newMockTaskRepository()
	service := NewTaskService(mockRepo)

	result, err := service.UpdateTask(1, nil)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if result != nil {
		t.Error("UpdateTask(nil) should return nil")
	}
}

func TestTaskService_UpdateTask_IDMismatch(t *testing.T) {
	mockRepo := newMockTaskRepository()
	service := NewTaskService(mockRepo)

	task := &models.Task{
		ID:   2,
		Name: "Test Task",
	}

	result, err := service.UpdateTask(1, task)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if result != nil {
		t.Error("UpdateTask() with ID mismatch should return nil")
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	mockRepo := newMockTaskRepository()
	service := NewTaskService(mockRepo)

	task := &models.Task{
		Name:   "Test Task",
		Status: models.TaskStatusPending,
	}
	createdTask, _ := service.CreateTask(task)

	err := service.DeleteTask(createdTask.ID)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	deletedTask, _ := service.GetTask(createdTask.ID)
	if deletedTask != nil {
		t.Error("Task should be deleted")
	}
}

func TestTaskService_DeleteTask_InvalidID(t *testing.T) {
	mockRepo := newMockTaskRepository()
	service := NewTaskService(mockRepo)

	err := service.DeleteTask(0)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	err = service.DeleteTask(-1)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}
}