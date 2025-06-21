package services

import (
	"errors"
	"testing"
	"time"

	"github.com/starbops/voidrunner/internal/models"
	"github.com/starbops/voidrunner/internal/repositories"
	"github.com/starbops/voidrunner/pkg/auth"
)

type mockAuthUserRepository struct {
	users      map[int]*models.User
	usersByKey map[string]*models.User // username or email -> user
	nextID     int
	failGet    bool
	failCreate bool
}

func newMockAuthUserRepository() *mockAuthUserRepository {
	return &mockAuthUserRepository{
		users:      make(map[int]*models.User),
		usersByKey: make(map[string]*models.User),
		nextID:     1,
	}
}

func (m *mockAuthUserRepository) GetUsers() ([]*models.User, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	users := make([]*models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *mockAuthUserRepository) GetUser(id int) (*models.User, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	user, exists := m.users[id]
	if !exists {
		return nil, repositories.ErrUserNotFound
	}
	return user, nil
}

func (m *mockAuthUserRepository) GetByUsernameOrEmail(username, email string) (*models.User, error) {
	if m.failGet {
		return nil, errors.New("mock error")
	}
	if user, exists := m.usersByKey[username]; exists {
		return user, nil
	}
	if user, exists := m.usersByKey[email]; exists {
		return user, nil
	}
	return nil, repositories.ErrUserNotFound
}

func (m *mockAuthUserRepository) Create(user *models.User) (*models.User, error) {
	if m.failCreate {
		return nil, errors.New("mock create error")
	}
	if user == nil {
		return nil, errors.New("user cannot be nil")
	}
	user.ID = m.nextID
	m.nextID++
	user.CreatedAt = time.Now().Format(time.RFC3339Nano)
	user.UpdatedAt = user.CreatedAt
	
	m.users[user.ID] = user
	m.usersByKey[user.Username] = user
	m.usersByKey[user.Email] = user
	return user, nil
}

func (m *mockAuthUserRepository) CreateUser(user *models.User) (*models.User, error) {
	return m.Create(user)
}

func (m *mockAuthUserRepository) UpdateUser(id int, user *models.User) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAuthUserRepository) DeleteUser(id int) error {
	return errors.New("not implemented")
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := newMockAuthUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	service := NewAuthService(repo, tokenManager)
	
	req := models.RegisterRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}
	
	user, err := service.Register(req)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	
	if user == nil {
		t.Fatal("Register() should return user")
	}
	if user.Username != req.Username {
		t.Errorf("Register() Username = %v, want %v", user.Username, req.Username)
	}
	if user.Email != req.Email {
		t.Errorf("Register() Email = %v, want %v", user.Email, req.Email)
	}
	if user.PasswordHash == "" {
		t.Error("Register() should set password hash")
	}
	if user.PasswordHash == req.Password {
		t.Error("Register() should not store plain text password")
	}
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	repo := newMockAuthUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	service := NewAuthService(repo, tokenManager)
	
	req := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	
	// Create user first time
	_, err := service.Register(req)
	if err != nil {
		t.Fatalf("First Register() error = %v", err)
	}
	
	// Try to create same user again
	_, err = service.Register(req)
	if err == nil {
		t.Error("Register() should return error for duplicate user")
	}
	if err != ErrUserAlreadyExists {
		t.Errorf("Register() error = %v, want %v", err, ErrUserAlreadyExists)
	}
}

func TestAuthService_Register_MissingFields(t *testing.T) {
	repo := newMockAuthUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	service := NewAuthService(repo, tokenManager)
	
	tests := []models.RegisterRequest{
		{Email: "test@example.com", Password: "password123"},        // missing username
		{Username: "testuser", Password: "password123"},             // missing email
		{Username: "testuser", Email: "test@example.com"},           // missing password
		{},                                                          // missing all required fields
	}
	
	for i, req := range tests {
		_, err := service.Register(req)
		if err == nil {
			t.Errorf("Register() test %d should return error for missing fields", i)
		}
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := newMockAuthUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	service := NewAuthService(repo, tokenManager)
	
	// Register a user first
	regReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	user, err := service.Register(regReq)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	
	// Login with username
	loginReq := models.LoginRequest{
		Identifier: "testuser",
		Password:   "password123",
	}
	
	response, err := service.Login(loginReq)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	
	if response == nil {
		t.Fatal("Login() should return response")
	}
	if response.Token == "" {
		t.Error("Login() should return token")
	}
	if response.User.ID != user.ID {
		t.Errorf("Login() User.ID = %v, want %v", response.User.ID, user.ID)
	}
}

func TestAuthService_Login_WithEmail(t *testing.T) {
	repo := newMockAuthUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	service := NewAuthService(repo, tokenManager)
	
	// Register a user first
	regReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := service.Register(regReq)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	
	// Login with email
	loginReq := models.LoginRequest{
		Identifier: "test@example.com",
		Password:   "password123",
	}
	
	response, err := service.Login(loginReq)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	
	if response == nil {
		t.Fatal("Login() should return response")
	}
	if response.Token == "" {
		t.Error("Login() should return token")
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	repo := newMockAuthUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	service := NewAuthService(repo, tokenManager)
	
	// Register a user first
	regReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := service.Register(regReq)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	
	// Login with wrong password
	loginReq := models.LoginRequest{
		Identifier: "testuser",
		Password:   "wrongpassword",
	}
	
	_, err = service.Login(loginReq)
	if err == nil {
		t.Error("Login() should return error for wrong password")
	}
	if err != ErrInvalidCredentials {
		t.Errorf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := newMockAuthUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	service := NewAuthService(repo, tokenManager)
	
	loginReq := models.LoginRequest{
		Identifier: "nonexistent",
		Password:   "password123",
	}
	
	_, err := service.Login(loginReq)
	if err == nil {
		t.Error("Login() should return error for non-existent user")
	}
	if err != ErrInvalidCredentials {
		t.Errorf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	repo := newMockAuthUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	service := NewAuthService(repo, tokenManager)
	
	// Generate a token
	token, err := tokenManager.GenerateToken(1, "testuser", "test@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	
	// Logout should succeed
	err = service.Logout(token)
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	
	// Token should be invalid after logout
	_, err = tokenManager.ValidateToken(token)
	if err == nil {
		t.Error("Token should be invalid after logout")
	}
}

func TestAuthService_Logout_EmptyToken(t *testing.T) {
	repo := newMockAuthUserRepository()
	tokenManager := auth.NewTokenManager("test-secret", time.Hour)
	service := NewAuthService(repo, tokenManager)
	
	err := service.Logout("")
	if err == nil {
		t.Error("Logout() should return error for empty token")
	}
}