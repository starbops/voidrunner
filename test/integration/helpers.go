package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/starbops/voidrunner/cmd/api"
	"github.com/starbops/voidrunner/internal/handlers"
	"github.com/starbops/voidrunner/internal/middleware"
	"github.com/starbops/voidrunner/internal/models"
	"github.com/starbops/voidrunner/internal/repositories"
	"github.com/starbops/voidrunner/internal/services"
	"github.com/starbops/voidrunner/pkg/auth"
	"github.com/starbops/voidrunner/pkg/config"
	_ "github.com/lib/pq"
)

type TestHelper struct {
	Server     *httptest.Server
	DB         *sql.DB
	UserRepo   repositories.UserRepository
	TaskRepo   repositories.TaskRepository
	Config     *config.Config
	AuthToken  string
	TestUserID int
}

func SetupTestHelper(t *testing.T) *TestHelper {
	helper := &TestHelper{}

	// Setup test configuration
	helper.Config = &config.Config{
		JWTSecret:      "test-secret-key-for-testing-only",
		StorageBackend: "postgres",
		PGHost:         getEnvOrDefault("TEST_PG_HOST", "localhost"),
		PGPort:         getEnvOrDefault("TEST_PG_PORT", "5432"),
		PGUser:         getEnvOrDefault("TEST_PG_USER", "voidrunner"),
		PGPassword:     getEnvOrDefault("TEST_PG_PASSWORD", "password"),
		PGDbName:       getEnvOrDefault("TEST_PG_DBNAME", "voidrunner_test"),
		JWTExpiration:  24 * time.Hour,
	}

	// Setup test database
	helper.setupTestDatabase(t)

	// Setup repositories
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		helper.Config.PGHost,
		helper.Config.PGPort,
		helper.Config.PGUser,
		helper.Config.PGPassword,
		helper.Config.PGDbName,
	)

	var err error
	helper.UserRepo, err = repositories.NewPostgresUserRepository(dataSourceName)
	if err != nil {
		t.Fatalf("Failed to create user repository: %v", err)
	}

	helper.TaskRepo, err = repositories.NewPostgresTaskRepository(dataSourceName)
	if err != nil {
		t.Fatalf("Failed to create task repository: %v", err)
	}

	// Setup HTTP server
	apiServer := api.NewAPIServer(":0", helper.Config, helper.TaskRepo, helper.UserRepo)
	mux := http.NewServeMux()
	
	// Setup routes manually for testing
	helper.setupTestRoutes(t, mux, apiServer)

	helper.Server = httptest.NewServer(mux)

	return helper
}

func (h *TestHelper) setupTestRoutes(t *testing.T, mux *http.ServeMux, apiServer *api.APIServer) {
	// Initialize services
	taskService := services.NewTaskService(h.TaskRepo)
	userService := services.NewUserService(h.UserRepo)
	
	tokenManager := auth.NewTokenManager(h.Config.JWTSecret, h.Config.JWTExpiration)
	authService := services.NewAuthService(h.UserRepo, tokenManager)

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(taskService)
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(tokenManager)

	// Welcome endpoint (no auth required)
	mux.HandleFunc("GET "+handlers.APIPrefix+"welcome", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Welcome to the VoidRunner API"}`))
	})

	// Register authentication routes (no auth required)
	mux.HandleFunc("POST "+handlers.APIPrefix+"register", authHandler.Register)
	mux.HandleFunc("POST "+handlers.APIPrefix+"login", authHandler.Login)
	mux.HandleFunc("POST "+handlers.APIPrefix+"logout", authHandler.Logout)

	// Register protected routes
	taskRouter := taskHandler.RegisterRoutes()
	userRouter := userHandler.RegisterRoutes()
	
	mux.Handle(handlers.APIPrefix+"tasks/", http.StripPrefix(handlers.APIPrefix+"tasks", authMiddleware(taskRouter)))
	mux.Handle(handlers.APIPrefix+"users/", http.StripPrefix(handlers.APIPrefix+"users", authMiddleware(userRouter)))
}

func (h *TestHelper) setupTestDatabase(t *testing.T) {
	// Connect to postgres database to create test database
	rootDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		h.Config.PGHost,
		h.Config.PGPort,
		h.Config.PGUser,
		h.Config.PGPassword,
	)

	rootDB, err := sql.Open("postgres", rootDSN)
	if err != nil {
		t.Fatalf("Failed to connect to root database: %v", err)
	}
	defer rootDB.Close()

	// Drop and recreate test database
	testDBName := h.Config.PGDbName
	
	// First terminate all connections to the test database
	_, _ = rootDB.Exec(fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()", testDBName))
	
	_, err = rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		t.Logf("Warning: Failed to drop test database: %v", err)
		// Continue anyway, maybe database doesn't exist
	}

	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to test database and run migrations
	testDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		h.Config.PGHost,
		h.Config.PGPort,
		h.Config.PGUser,
		h.Config.PGPassword,
		h.Config.PGDbName,
	)

	h.DB, err = sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations manually for testing
	h.runTestMigrations(t)
}

func (h *TestHelper) runTestMigrations(t *testing.T) {
	migrations := []string{
		// Migration 001: Create users table
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// Migration 002: Create tasks table
		`CREATE TABLE tasks (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// Migration 003: Add password hash to users
		`ALTER TABLE users ADD COLUMN password_hash VARCHAR(255) NOT NULL DEFAULT ''`,
	}

	for i, migration := range migrations {
		_, err := h.DB.Exec(migration)
		if err != nil {
			t.Fatalf("Failed to run migration %d: %v", i+1, err)
		}
	}
}

func (h *TestHelper) TearDown() {
	if h.Server != nil {
		h.Server.Close()
	}
	if h.DB != nil {
		h.DB.Close()
	}
}

func (h *TestHelper) CreateTestUser(t *testing.T) *models.User {
	user := &models.User{
		Username:     fmt.Sprintf("testuser_%d", time.Now().UnixNano()),
		Email:        fmt.Sprintf("test_%d@example.com", time.Now().UnixNano()),
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: "$2a$10$hash", // Mock hash
	}

	createdUser, err := h.UserRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	h.TestUserID = createdUser.ID
	return createdUser
}

func (h *TestHelper) RegisterAndLoginUser(t *testing.T) string {
	// Register user
	registerData := map[string]string{
		"username":   fmt.Sprintf("testuser_%d", time.Now().UnixNano()),
		"email":      fmt.Sprintf("test_%d@example.com", time.Now().UnixNano()),
		"password":   "testpassword123",
		"first_name": "Test",
		"last_name":  "User",
	}

	registerJSON, _ := json.Marshal(registerData)
	resp := h.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Registration failed with status %d", resp.StatusCode)
	}

	// Login user
	loginData := map[string]string{
		"identifier": registerData["username"],
		"password":   registerData["password"],
	}

	loginJSON, _ := json.Marshal(loginData)
	resp = h.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(loginJSON), "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed with status %d", resp.StatusCode)
	}

	var loginResponse struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResponse)
	
	h.AuthToken = loginResponse.Token
	return loginResponse.Token
}

func (h *TestHelper) MakeRequest(t *testing.T, method, path string, body *bytes.Buffer, authToken string) *http.Response {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, h.Server.URL+path, body)
	} else {
		req, err = http.NewRequest(method, h.Server.URL+path, nil)
	}

	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

func (h *TestHelper) CleanupTestData(t *testing.T) {
	// Clean up test data
	_, err := h.DB.Exec("DELETE FROM tasks")
	if err != nil {
		t.Logf("Failed to clean up tasks: %v", err)
	}

	_, err = h.DB.Exec("DELETE FROM users")
	if err != nil {
		t.Logf("Failed to clean up users: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}