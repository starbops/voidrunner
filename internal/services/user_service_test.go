package services

import (
	"errors"
	"testing"

	"github.com/starbops/voidrunner/internal/models"
	"github.com/starbops/voidrunner/internal/repositories"
)

type mockUserRepository struct {
	users   map[int]*models.User
	nextID  int
	failGet bool
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:  make(map[int]*models.User),
		nextID: 1,
	}
}

func (m *mockUserRepository) GetUsers() ([]*models.User, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	users := make([]*models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *mockUserRepository) GetUser(id int) (*models.User, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	user, exists := m.users[id]
	if !exists {
		return nil, repositories.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepository) GetByUsernameOrEmail(username, email string) (*models.User, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	for _, user := range m.users {
		if user.Username == username || user.Email == email {
			return user, nil
		}
	}
	return nil, repositories.ErrUserNotFound
}

func (m *mockUserRepository) Create(user *models.User) (*models.User, error) {
	if user == nil {
		return nil, errors.New("user cannot be nil")
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserRepository) CreateUser(user *models.User) (*models.User, error) {
	if user == nil {
		return nil, nil
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserRepository) UpdateUser(id int, user *models.User) (*models.User, error) {
	if user == nil || user.ID != id {
		return nil, nil
	}
	if _, exists := m.users[id]; !exists {
		return nil, nil
	}
	m.users[id] = user
	return user, nil
}

func (m *mockUserRepository) DeleteUser(id int) error {
	if _, exists := m.users[id]; !exists {
		return nil
	}
	delete(m.users, id)
	return nil
}

func TestUserService_CreateUser(t *testing.T) {
	mockRepo := newMockUserRepository()
	service := NewUserService(mockRepo)

	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	createdUser, err := service.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if createdUser.ID == 0 {
		t.Error("CreateUser() should assign an ID")
	}
	if createdUser.Username != user.Username {
		t.Errorf("CreateUser() Username = %v, want %v", createdUser.Username, user.Username)
	}
	if createdUser.Email != user.Email {
		t.Errorf("CreateUser() Email = %v, want %v", createdUser.Email, user.Email)
	}
}

func TestUserService_CreateUser_Nil(t *testing.T) {
	mockRepo := newMockUserRepository()
	service := NewUserService(mockRepo)

	createdUser, err := service.CreateUser(nil)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if createdUser != nil {
		t.Error("CreateUser(nil) should return nil")
	}
}

func TestUserService_GetUser(t *testing.T) {
	mockRepo := newMockUserRepository()
	service := NewUserService(mockRepo)

	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	createdUser, _ := service.CreateUser(user)

	retrievedUser, err := service.GetUser(createdUser.ID)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}

	if retrievedUser.ID != createdUser.ID {
		t.Errorf("GetUser() ID = %v, want %v", retrievedUser.ID, createdUser.ID)
	}
	if retrievedUser.Username != createdUser.Username {
		t.Errorf("GetUser() Username = %v, want %v", retrievedUser.Username, createdUser.Username)
	}
}

func TestUserService_GetUsers(t *testing.T) {
	mockRepo := newMockUserRepository()
	service := NewUserService(mockRepo)

	user1 := &models.User{Username: "user1", Email: "user1@example.com"}
	user2 := &models.User{Username: "user2", Email: "user2@example.com"}

	service.CreateUser(user1)
	service.CreateUser(user2)

	users, err := service.GetUsers()
	if err != nil {
		t.Fatalf("GetUsers() error = %v", err)
	}

	if len(users) != 2 {
		t.Errorf("GetUsers() returned %v users, want 2", len(users))
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	mockRepo := newMockUserRepository()
	service := NewUserService(mockRepo)

	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	createdUser, _ := service.CreateUser(user)

	updatedUser := &models.User{
		ID:        createdUser.ID,
		Username:  "updateduser",
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "User",
	}

	result, err := service.UpdateUser(createdUser.ID, updatedUser)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	if result.Username != "updateduser" {
		t.Errorf("UpdateUser() Username = %v, want %v", result.Username, "updateduser")
	}
	if result.Email != "updated@example.com" {
		t.Errorf("UpdateUser() Email = %v, want %v", result.Email, "updated@example.com")
	}
}

func TestUserService_UpdateUser_Nil(t *testing.T) {
	mockRepo := newMockUserRepository()
	service := NewUserService(mockRepo)

	result, err := service.UpdateUser(1, nil)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	if result != nil {
		t.Error("UpdateUser(nil) should return nil")
	}
}

func TestUserService_UpdateUser_IDMismatch(t *testing.T) {
	mockRepo := newMockUserRepository()
	service := NewUserService(mockRepo)

	user := &models.User{
		ID:       2,
		Username: "testuser",
		Email:    "test@example.com",
	}

	result, err := service.UpdateUser(1, user)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	if result != nil {
		t.Error("UpdateUser() with ID mismatch should return nil")
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	mockRepo := newMockUserRepository()
	service := NewUserService(mockRepo)

	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	createdUser, _ := service.CreateUser(user)

	err := service.DeleteUser(createdUser.ID)
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}

	deletedUser, _ := service.GetUser(createdUser.ID)
	if deletedUser != nil {
		t.Error("User should be deleted")
	}
}

func TestUserService_DeleteUser_InvalidID(t *testing.T) {
	mockRepo := newMockUserRepository()
	service := NewUserService(mockRepo)

	err := service.DeleteUser(0)
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}

	err = service.DeleteUser(-1)
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
}