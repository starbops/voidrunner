package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/starbops/voidrunner/internal/middleware"
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
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserService) UpdateUser(id int, user *models.User) (*models.User, error) {
	if m.failUpdate {
		return nil, errors.New("mock error")
	}
	existingUser, exists := m.users[id]
	if !exists {
		return nil, nil
	}
	user.ID = id
	m.users[id] = user
	// Preserve the original created_at if it exists
	user.CreatedAt = existingUser.CreatedAt
	return user, nil
}

func (m *mockUserService) DeleteUser(id int) error {
	if m.failDelete {
		return errors.New("mock error")
	}
	delete(m.users, id)
	return nil
}

func setupUserHandler() (*UserHandler, *mockUserService) {
	mockService := newMockUserService()
	handler := NewUserHandler(mockService)
	return handler, mockService
}

func addUserContextForUser(req *http.Request, userID int) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.UsernameKey, "testuser")
	ctx = context.WithValue(ctx, middleware.UserEmailKey, "test@example.com")
	return req.WithContext(ctx)
}


func TestUserHandler_GetCurrentUser_Success(t *testing.T) {
	handler, mockService := setupUserHandler()

	// Add a test user
	testUser := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	createdUser, _ := mockService.CreateUser(testUser)

	req := httptest.NewRequest("GET", "/me", nil)
	req = addUserContextForUser(req, createdUser.ID)
	w := httptest.NewRecorder()

	handler.getCurrentUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var user models.User
	err := json.NewDecoder(w.Body).Decode(&user)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.Username != testUser.Username {
		t.Errorf("Expected username %s, got %s", testUser.Username, user.Username)
	}
}

func TestUserHandler_GetCurrentUser_NoContext(t *testing.T) {
	handler, _ := setupUserHandler()

	req := httptest.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()

	handler.getCurrentUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUserHandler_GetCurrentUser_ServiceError(t *testing.T) {
	handler, mockService := setupUserHandler()
	mockService.failGet = true

	req := httptest.NewRequest("GET", "/me", nil)
	req = addUserContextForUser(req, 1)
	w := httptest.NewRecorder()

	handler.getCurrentUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestUserHandler_UpdateCurrentUser_Success(t *testing.T) {
	handler, mockService := setupUserHandler()

	// Add a test user
	testUser := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	createdUser, _ := mockService.CreateUser(testUser)

	// Create update request
	updateReq := models.UpdateUserRequest{
		FirstName: stringPtr("Updated"),
		LastName:  stringPtr("Name"),
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest("PUT", "/me", bytes.NewBuffer(reqBody))
	req = addUserContextForUser(req, createdUser.ID)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.updateCurrentUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var user models.User
	err := json.NewDecoder(w.Body).Decode(&user)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.FirstName != "Updated" {
		t.Errorf("Expected first name %s, got %s", "Updated", user.FirstName)
	}
	if user.LastName != "Name" {
		t.Errorf("Expected last name %s, got %s", "Name", user.LastName)
	}
}

func TestUserHandler_UpdateCurrentUser_InvalidJSON(t *testing.T) {
	handler, mockService := setupUserHandler()

	// Add a test user
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	createdUser, _ := mockService.CreateUser(testUser)

	req := httptest.NewRequest("PUT", "/me", bytes.NewBuffer([]byte("invalid json")))
	req = addUserContextForUser(req, createdUser.ID)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.updateCurrentUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUserHandler_DeleteCurrentUser_Success(t *testing.T) {
	handler, mockService := setupUserHandler()

	// Add a test user
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	createdUser, _ := mockService.CreateUser(testUser)

	req := httptest.NewRequest("DELETE", "/me", nil)
	req = addUserContextForUser(req, createdUser.ID)
	w := httptest.NewRecorder()

	handler.deleteCurrentUser(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify user was deleted
	user, _ := mockService.GetUser(createdUser.ID)
	if user != nil {
		t.Error("Expected user to be deleted")
	}
}

func TestUserHandler_DeleteCurrentUser_NoContext(t *testing.T) {
	handler, _ := setupUserHandler()

	req := httptest.NewRequest("DELETE", "/me", nil)
	w := httptest.NewRecorder()

	handler.deleteCurrentUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUserHandler_DeleteCurrentUser_ServiceError(t *testing.T) {
	handler, mockService := setupUserHandler()
	mockService.failDelete = true

	// Add a test user
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	createdUser, _ := mockService.CreateUser(testUser)

	req := httptest.NewRequest("DELETE", "/me", nil)
	req = addUserContextForUser(req, createdUser.ID)
	w := httptest.NewRecorder()

	handler.deleteCurrentUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

