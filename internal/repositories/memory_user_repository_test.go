package repositories

import (
	"testing"
	"time"

	"github.com/starbops/voidrunner/internal/models"
)

func TestMemoryUserRepository_CreateUser(t *testing.T) {
	repo := NewMemoryUserRepository()

	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	createdUser, err := repo.CreateUser(user)
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
	if createdUser.CreatedAt == "" {
		t.Error("CreateUser() should set CreatedAt")
	}
	if createdUser.UpdatedAt == "" {
		t.Error("CreateUser() should set UpdatedAt")
	}
}

func TestMemoryUserRepository_CreateUser_NilUser(t *testing.T) {
	repo := NewMemoryUserRepository()

	createdUser, err := repo.CreateUser(nil)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if createdUser != nil {
		t.Error("CreateUser(nil) should return nil")
	}
}

func TestMemoryUserRepository_GetUser(t *testing.T) {
	repo := NewMemoryUserRepository()

	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	createdUser, _ := repo.CreateUser(user)

	retrievedUser, err := repo.GetUser(createdUser.ID)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}

	if retrievedUser.ID != createdUser.ID {
		t.Errorf("GetUser() ID = %v, want %v", retrievedUser.ID, createdUser.ID)
	}
	if retrievedUser.Username != createdUser.Username {
		t.Errorf("GetUser() Username = %v, want %v", retrievedUser.Username, createdUser.Username)
	}
	if retrievedUser.Email != createdUser.Email {
		t.Errorf("GetUser() Email = %v, want %v", retrievedUser.Email, createdUser.Email)
	}
}

func TestMemoryUserRepository_GetUser_NotFound(t *testing.T) {
	repo := NewMemoryUserRepository()

	user, err := repo.GetUser(999)
	if err == nil {
		t.Fatal("GetUser() should return error for non-existent user")
	}
	if err != ErrUserNotFound {
		t.Errorf("GetUser() error = %v, want %v", err, ErrUserNotFound)
	}
	if user != nil {
		t.Error("GetUser() should return nil for non-existent user")
	}
}

func TestMemoryUserRepository_GetUsers(t *testing.T) {
	repo := NewMemoryUserRepository()

	user1 := &models.User{Username: "user1", Email: "user1@example.com"}
	user2 := &models.User{Username: "user2", Email: "user2@example.com"}

	_, err := repo.CreateUser(user1)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	_, err = repo.CreateUser(user2)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	users, err := repo.GetUsers()
	if err != nil {
		t.Fatalf("GetUsers() error = %v", err)
	}

	if len(users) != 2 {
		t.Errorf("GetUsers() returned %v users, want 2", len(users))
	}
}

func TestMemoryUserRepository_UpdateUser(t *testing.T) {
	repo := NewMemoryUserRepository()

	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	createdUser, _ := repo.CreateUser(user)

	// Small delay to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	updatedUser := &models.User{
		ID:        createdUser.ID,
		Username:  "updateduser",
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "User",
	}

	result, err := repo.UpdateUser(createdUser.ID, updatedUser)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	if result.Username != "updateduser" {
		t.Errorf("UpdateUser() Username = %v, want %v", result.Username, "updateduser")
	}
	if result.Email != "updated@example.com" {
		t.Errorf("UpdateUser() Email = %v, want %v", result.Email, "updated@example.com")
	}
	if result.FirstName != "Updated" {
		t.Errorf("UpdateUser() FirstName = %v, want %v", result.FirstName, "Updated")
	}
	if result.UpdatedAt == createdUser.UpdatedAt {
		t.Errorf("UpdateUser() should update UpdatedAt timestamp: result=%s, created=%s", result.UpdatedAt, createdUser.UpdatedAt)
	}
}

func TestMemoryUserRepository_UpdateUser_NotFound(t *testing.T) {
	repo := NewMemoryUserRepository()

	user := &models.User{
		ID:       999,
		Username: "nonexistent",
		Email:    "nonexistent@example.com",
	}

	result, err := repo.UpdateUser(999, user)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	if result != nil {
		t.Error("UpdateUser() should return nil for non-existent user")
	}
}

func TestMemoryUserRepository_UpdateUser_IDMismatch(t *testing.T) {
	repo := NewMemoryUserRepository()

	user := &models.User{
		ID:       2,
		Username: "testuser",
		Email:    "test@example.com",
	}

	result, err := repo.UpdateUser(1, user)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	if result != nil {
		t.Error("UpdateUser() should return nil for ID mismatch")
	}
}

func TestMemoryUserRepository_UpdateUser_NilUser(t *testing.T) {
	repo := NewMemoryUserRepository()

	result, err := repo.UpdateUser(1, nil)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	if result != nil {
		t.Error("UpdateUser(nil) should return nil")
	}
}

func TestMemoryUserRepository_DeleteUser(t *testing.T) {
	repo := NewMemoryUserRepository()

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	createdUser, _ := repo.CreateUser(user)

	err := repo.DeleteUser(createdUser.ID)
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}

	deletedUser, _ := repo.GetUser(createdUser.ID)
	if deletedUser != nil {
		t.Error("User should be deleted")
	}
}

func TestMemoryUserRepository_DeleteUser_NotFound(t *testing.T) {
	repo := NewMemoryUserRepository()

	err := repo.DeleteUser(999)
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
}