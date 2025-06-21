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

type mockUserService struct {
	users      map[int]*models.User
	nextID     int
	failGet    bool
	failCreate bool
	failUpdate bool
	failDelete bool
}

func newMockUserService() *mockUserService {
	return &mockUserService{
		users:  make(map[int]*models.User),
		nextID: 1,
	}
}

func (m *mockUserService) GetUsers() ([]*models.User, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	users := make([]*models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *mockUserService) GetUser(id int) (*models.User, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	user, exists := m.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *mockUserService) CreateUser(user *models.User) (*models.User, error) {
	if m.failCreate {
		return nil, errors.New("mock error")
	}
	if user == nil {
		return nil, nil
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserService) UpdateUser(id int, user *models.User) (*models.User, error) {
	if m.failUpdate {
		return nil, errors.New("mock error")
	}
	if user == nil || user.ID != id {
		return nil, nil
	}
	if _, exists := m.users[id]; !exists {
		return nil, nil
	}
	m.users[id] = user
	return user, nil
}

func (m *mockUserService) DeleteUser(id int) error {
	if m.failDelete {
		return errors.New("mock error")
	}
	if _, exists := m.users[id]; !exists {
		return nil
	}
	delete(m.users, id)
	return nil
}

func TestUserHandler_GetUsers(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	user1 := &models.User{Username: "user1", Email: "user1@example.com"}
	user2 := &models.User{Username: "user2", Email: "user2@example.com"}
	mockService.CreateUser(user1)
	mockService.CreateUser(user2)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.getUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("getUsers() status = %v, want %v", w.Code, http.StatusOK)
	}

	var users []*models.User
	if err := json.NewDecoder(w.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("getUsers() returned %v users, want 2", len(users))
	}
}

func TestUserHandler_GetUsers_ServiceError(t *testing.T) {
	mockService := newMockUserService()
	mockService.failGet = true
	handler := NewUserHandler(mockService)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.getUsers(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("getUsers() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	user := &models.User{Username: "testuser", Email: "test@example.com"}
	createdUser, _ := mockService.CreateUser(user)

	req := httptest.NewRequest("GET", "/1/", nil)
	req.SetPathValue("id", strconv.Itoa(createdUser.ID))
	w := httptest.NewRecorder()

	handler.getUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("getUser() status = %v, want %v", w.Code, http.StatusOK)
	}

	var retrievedUser models.User
	if err := json.NewDecoder(w.Body).Decode(&retrievedUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if retrievedUser.ID != createdUser.ID {
		t.Errorf("getUser() ID = %v, want %v", retrievedUser.ID, createdUser.ID)
	}
	if retrievedUser.Username != createdUser.Username {
		t.Errorf("getUser() Username = %v, want %v", retrievedUser.Username, createdUser.Username)
	}
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	req := httptest.NewRequest("GET", "/999/", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	handler.getUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("getUser() status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestUserHandler_GetUser_InvalidID(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	req := httptest.NewRequest("GET", "/invalid/", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	handler.getUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("getUser() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_GetUser_ServiceError(t *testing.T) {
	mockService := newMockUserService()
	mockService.failGet = true
	handler := NewUserHandler(mockService)

	req := httptest.NewRequest("GET", "/1/", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	handler.getUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("getUser() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.createUser(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("createUser() status = %v, want %v", w.Code, http.StatusCreated)
	}

	var createdUser models.User
	if err := json.NewDecoder(w.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdUser.Username != user.Username {
		t.Errorf("createUser() Username = %v, want %v", createdUser.Username, user.Username)
	}
	if createdUser.Email != user.Email {
		t.Errorf("createUser() Email = %v, want %v", createdUser.Email, user.Email)
	}
}

func TestUserHandler_CreateUser_InvalidJSON(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.createUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("createUser() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_CreateUser_ServiceError(t *testing.T) {
	mockService := newMockUserService()
	mockService.failCreate = true
	handler := NewUserHandler(mockService)

	user := &models.User{Username: "testuser", Email: "test@example.com"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.createUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("createUser() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	user := &models.User{Username: "testuser", Email: "test@example.com"}
	createdUser, _ := mockService.CreateUser(user)

	updatedUser := &models.User{
		ID:        createdUser.ID,
		Username:  "updateduser",
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "User",
	}

	body, _ := json.Marshal(updatedUser)
	req := httptest.NewRequest("PUT", "/1/", bytes.NewReader(body))
	req.SetPathValue("id", strconv.Itoa(createdUser.ID))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.updateUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("updateUser() status = %v, want %v", w.Code, http.StatusOK)
	}

	var result models.User
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Username != "updateduser" {
		t.Errorf("updateUser() Username = %v, want %v", result.Username, "updateduser")
	}
}

func TestUserHandler_UpdateUser_NotFound(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	user := &models.User{ID: 999, Username: "nonexistent", Email: "test@example.com"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("PUT", "/999/", bytes.NewReader(body))
	req.SetPathValue("id", "999")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.updateUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("updateUser() status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestUserHandler_UpdateUser_InvalidID(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	user := &models.User{Username: "testuser", Email: "test@example.com"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("PUT", "/invalid/", bytes.NewReader(body))
	req.SetPathValue("id", "invalid")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.updateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("updateUser() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_UpdateUser_InvalidJSON(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	req := httptest.NewRequest("PUT", "/1/", bytes.NewReader([]byte("invalid json")))
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.updateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("updateUser() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_UpdateUser_ServiceError(t *testing.T) {
	mockService := newMockUserService()
	mockService.failUpdate = true
	handler := NewUserHandler(mockService)

	user := &models.User{ID: 1, Username: "testuser", Email: "test@example.com"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("PUT", "/1/", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.updateUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("updateUser() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	user := &models.User{Username: "testuser", Email: "test@example.com"}
	createdUser, _ := mockService.CreateUser(user)

	req := httptest.NewRequest("DELETE", "/1/", nil)
	req.SetPathValue("id", strconv.Itoa(createdUser.ID))
	w := httptest.NewRecorder()

	handler.deleteUser(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("deleteUser() status = %v, want %v", w.Code, http.StatusNoContent)
	}

	deletedUser, _ := mockService.GetUser(createdUser.ID)
	if deletedUser != nil {
		t.Error("User should be deleted")
	}
}

func TestUserHandler_DeleteUser_InvalidID(t *testing.T) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)

	req := httptest.NewRequest("DELETE", "/invalid/", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	handler.deleteUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("deleteUser() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_DeleteUser_ServiceError(t *testing.T) {
	mockService := newMockUserService()
	mockService.failDelete = true
	handler := NewUserHandler(mockService)

	req := httptest.NewRequest("DELETE", "/1/", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	handler.deleteUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("deleteUser() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}