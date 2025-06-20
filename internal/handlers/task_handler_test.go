package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/starbops/voidrunner/internal/models"
)

type mockTaskService struct {
	tasks   map[int]*models.Task
	nextID  int
	failGet bool
}

func newMockTaskService() *mockTaskService {
	return &mockTaskService{
		tasks:  make(map[int]*models.Task),
		nextID: 1,
	}
}

func (m *mockTaskService) GetTasks() ([]*models.Task, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	tasks := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *mockTaskService) GetTask(id int) (*models.Task, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	task, exists := m.tasks[id]
	if !exists {
		return nil, nil
	}
	return task, nil
}

func (m *mockTaskService) CreateTask(task *models.Task) (*models.Task, error) {
	if task == nil {
		return nil, nil
	}
	task.ID = m.nextID
	m.nextID++
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskService) UpdateTask(id int, task *models.Task) (*models.Task, error) {
	if task == nil || task.ID != id {
		return nil, nil
	}
	if _, exists := m.tasks[id]; !exists {
		return nil, nil
	}
	m.tasks[id] = task
	return task, nil
}

func (m *mockTaskService) DeleteTask(id int) error {
	if _, exists := m.tasks[id]; !exists {
		return nil
	}
	delete(m.tasks, id)
	return nil
}

func TestTaskHandler_GetTasks(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	task1 := &models.Task{Name: "Task 1", Status: models.TaskStatusPending}
	task2 := &models.Task{Name: "Task 2", Status: models.TaskStatusCompleted}
	mockService.CreateTask(task1)
	mockService.CreateTask(task2)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.getTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("getTasks() status = %v, want %v", w.Code, http.StatusOK)
	}

	var tasks []*models.Task
	if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("getTasks() returned %v tasks, want 2", len(tasks))
	}
}

func TestTaskHandler_GetTasks_ServiceError(t *testing.T) {
	mockService := newMockTaskService()
	mockService.failGet = true
	handler := NewTaskHandler(mockService)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.getTasks(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("getTasks() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestTaskHandler_GetTask(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	task := &models.Task{Name: "Test Task", Status: models.TaskStatusPending}
	createdTask, _ := mockService.CreateTask(task)

	req := httptest.NewRequest("GET", "/1/", nil)
	req.SetPathValue("id", strconv.Itoa(createdTask.ID))
	w := httptest.NewRecorder()

	handler.getTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("getTask() status = %v, want %v", w.Code, http.StatusOK)
	}

	var retrievedTask models.Task
	if err := json.NewDecoder(w.Body).Decode(&retrievedTask); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if retrievedTask.ID != createdTask.ID {
		t.Errorf("getTask() ID = %v, want %v", retrievedTask.ID, createdTask.ID)
	}
}

func TestTaskHandler_GetTask_NotFound(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	req := httptest.NewRequest("GET", "/999/", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	handler.getTask(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("getTask() status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestTaskHandler_GetTask_InvalidID(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	req := httptest.NewRequest("GET", "/invalid/", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	handler.getTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("getTask() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestTaskHandler_CreateTask(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	task := models.Task{
		Name:        "New Task",
		Description: "New Description",
		Status:      models.TaskStatusPending,
	}

	body, _ := json.Marshal(task)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.createTask(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("createTask() status = %v, want %v", w.Code, http.StatusCreated)
	}

	var createdTask models.Task
	if err := json.NewDecoder(w.Body).Decode(&createdTask); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdTask.Name != task.Name {
		t.Errorf("createTask() Name = %v, want %v", createdTask.Name, task.Name)
	}
}

func TestTaskHandler_CreateTask_InvalidJSON(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.createTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("createTask() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	task := &models.Task{Name: "Test Task", Status: models.TaskStatusPending}
	createdTask, _ := mockService.CreateTask(task)

	updatedTask := models.Task{
		ID:     createdTask.ID,
		Name:   "Updated Task",
		Status: models.TaskStatusCompleted,
	}

	body, _ := json.Marshal(updatedTask)
	req := httptest.NewRequest("PUT", "/1/", bytes.NewReader(body))
	req.SetPathValue("id", strconv.Itoa(createdTask.ID))
	w := httptest.NewRecorder()

	handler.updateTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("updateTask() status = %v, want %v", w.Code, http.StatusOK)
	}

	var result models.Task
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Name != "Updated Task" {
		t.Errorf("updateTask() Name = %v, want %v", result.Name, "Updated Task")
	}
}

func TestTaskHandler_UpdateTask_NotFound(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	task := models.Task{
		ID:   999,
		Name: "Non-existent Task",
	}

	body, _ := json.Marshal(task)
	req := httptest.NewRequest("PUT", "/999/", bytes.NewReader(body))
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	handler.updateTask(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("updateTask() status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	task := &models.Task{Name: "Test Task", Status: models.TaskStatusPending}
	createdTask, _ := mockService.CreateTask(task)

	req := httptest.NewRequest("DELETE", "/1/", nil)
	req.SetPathValue("id", strconv.Itoa(createdTask.ID))
	w := httptest.NewRecorder()

	handler.deleteTask(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("deleteTask() status = %v, want %v", w.Code, http.StatusNoContent)
	}

	deletedTask, _ := mockService.GetTask(createdTask.ID)
	if deletedTask != nil {
		t.Error("Task should be deleted")
	}
}

func TestTaskHandler_DeleteTask_InvalidID(t *testing.T) {
	mockService := newMockTaskService()
	handler := NewTaskHandler(mockService)

	req := httptest.NewRequest("DELETE", "/invalid/", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	handler.deleteTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("deleteTask() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}