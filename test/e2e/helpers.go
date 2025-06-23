package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/starbops/voidrunner/internal/models"
	_ "github.com/lib/pq"
)

type E2ETestHelper struct {
	ServerProcess *exec.Cmd
	ServerURL     string
	ServerPort    string
	TestDB        *sql.DB
	TestDBName    string
	AuthToken     string
	TestUserID    int
	BackendType   string // "memory" or "postgres"
}

// SetupE2ETestHelper sets up a complete E2E test environment with a real server
func SetupE2ETestHelper(t *testing.T, backendType string) *E2ETestHelper {
	helper := &E2ETestHelper{
		BackendType: backendType,
	}

	// Get a free port for the server
	helper.ServerPort = helper.getFreePort(t)
	helper.ServerURL = fmt.Sprintf("http://localhost:%s", helper.ServerPort)

	// Setup database if using postgres backend
	if backendType == "postgres" {
		helper.setupTestDatabase(t)
	}

	// Build the application
	helper.buildApplication(t)

	// Start the server
	helper.startServer(t)

	// Wait for server to be ready
	helper.waitForServer(t)

	return helper
}

func (h *E2ETestHelper) getFreePort(t *testing.T) string {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to find free port: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("%d", addr.Port)
}

func (h *E2ETestHelper) setupTestDatabase(t *testing.T) {
	// Generate unique test database name
	h.TestDBName = fmt.Sprintf("voidrunner_e2e_test_%d", time.Now().UnixNano())

	// Connect to postgres database to create test database
	pgHost := getEnvOrDefault("TEST_PG_HOST", "localhost")
	pgPort := getEnvOrDefault("TEST_PG_PORT", "5432")
	pgUser := getEnvOrDefault("TEST_PG_USER", "voidrunner")
	pgPassword := getEnvOrDefault("TEST_PG_PASSWORD", "password")

	rootDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		pgHost, pgPort, pgUser, pgPassword)

	rootDB, err := sql.Open("postgres", rootDSN)
	if err != nil {
		t.Fatalf("Failed to connect to root database: %v", err)
	}
	defer rootDB.Close()

	// Create test database
	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE %s", h.TestDBName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to test database
	testDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPassword, h.TestDBName)

	h.TestDB, err = sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	h.runMigrations(t)
}

func (h *E2ETestHelper) runMigrations(t *testing.T) {
	migrations := []string{
		// Create users table
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// Create tasks table
		`CREATE TABLE tasks (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for i, migration := range migrations {
		_, err := h.TestDB.Exec(migration)
		if err != nil {
			t.Fatalf("Failed to run migration %d: %v", i+1, err)
		}
	}
}

func (h *E2ETestHelper) buildApplication(t *testing.T) {
	// Get the project root directory
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to project root (assuming we're in test/e2e)
	for filepath.Base(projectRoot) != "voidrunner" {
		projectRoot = filepath.Dir(projectRoot)
		if projectRoot == "/" {
			t.Fatalf("Could not find project root")
		}
	}

	// Build the application
	cmd := exec.Command("go", "build", "-o", "./bin/voidrunner", "./cmd/main.go")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build application: %v\nOutput: %s", err, output)
	}
}

func (h *E2ETestHelper) startServer(t *testing.T) {
	// Get the project root directory
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to project root
	for filepath.Base(projectRoot) != "voidrunner" {
		projectRoot = filepath.Dir(projectRoot)
		if projectRoot == "/" {
			t.Fatalf("Could not find project root")
		}
	}

	// Prepare environment variables
	env := os.Environ()
	env = append(env, fmt.Sprintf("PORT=%s", h.ServerPort))
	env = append(env, fmt.Sprintf("STORAGE_BACKEND=%s", h.BackendType))

	if h.BackendType == "postgres" {
		env = append(env, fmt.Sprintf("PG_HOST=%s", getEnvOrDefault("TEST_PG_HOST", "localhost")))
		env = append(env, fmt.Sprintf("PG_PORT=%s", getEnvOrDefault("TEST_PG_PORT", "5432")))
		env = append(env, fmt.Sprintf("PG_USER=%s", getEnvOrDefault("TEST_PG_USER", "voidrunner")))
		env = append(env, fmt.Sprintf("PG_PASSWORD=%s", getEnvOrDefault("TEST_PG_PASSWORD", "password")))
		env = append(env, fmt.Sprintf("PG_DBNAME=%s", h.TestDBName))
	}

	// Start the server
	h.ServerProcess = exec.Command("./bin/voidrunner")
	h.ServerProcess.Dir = projectRoot
	h.ServerProcess.Env = env

	err = h.ServerProcess.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give the server a moment to start
	time.Sleep(2 * time.Second)
}

func (h *E2ETestHelper) waitForServer(t *testing.T) {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(h.ServerURL + "/api/v1/welcome")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Server did not become ready within %d seconds", maxRetries)
}

func (h *E2ETestHelper) TearDown(t *testing.T) {
	// Stop the server
	if h.ServerProcess != nil {
		err := h.ServerProcess.Process.Kill()
		if err != nil {
			t.Logf("Failed to kill server process: %v", err)
		}
		h.ServerProcess.Wait()
	}

	// Clean up test database
	if h.BackendType == "postgres" && h.TestDBName != "" {
		h.cleanupTestDatabase(t)
	}
}

func (h *E2ETestHelper) cleanupTestDatabase(t *testing.T) {
	if h.TestDB != nil {
		h.TestDB.Close()
	}

	// Connect to postgres database to drop test database
	pgHost := getEnvOrDefault("TEST_PG_HOST", "localhost")
	pgPort := getEnvOrDefault("TEST_PG_PORT", "5432")
	pgUser := getEnvOrDefault("TEST_PG_USER", "voidrunner")
	pgPassword := getEnvOrDefault("TEST_PG_PASSWORD", "password")

	rootDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		pgHost, pgPort, pgUser, pgPassword)

	rootDB, err := sql.Open("postgres", rootDSN)
	if err != nil {
		t.Logf("Failed to connect to root database for cleanup: %v", err)
		return
	}
	defer rootDB.Close()

	// Terminate connections to test database
	_, _ = rootDB.Exec(fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()", h.TestDBName))

	// Drop test database
	_, err = rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", h.TestDBName))
	if err != nil {
		t.Logf("Failed to drop test database: %v", err)
	}
}

func (h *E2ETestHelper) RegisterAndLoginUser(t *testing.T) string {
	// Register user
	registerData := map[string]string{
		"username":   fmt.Sprintf("e2euser_%d", time.Now().UnixNano()),
		"email":      fmt.Sprintf("e2e_%d@example.com", time.Now().UnixNano()),
		"password":   "e2epassword123",
		"first_name": "E2E",
		"last_name":  "User",
	}

	registerJSON, _ := json.Marshal(registerData)
	resp := h.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Registration failed with status %d", resp.StatusCode)
	}
	resp.Body.Close()

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
	resp.Body.Close()

	h.AuthToken = loginResponse.Token
	return loginResponse.Token
}

func (h *E2ETestHelper) MakeRequest(t *testing.T, method, path string, body *bytes.Buffer, authToken string) *http.Response {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, h.ServerURL+path, body)
	} else {
		req, err = http.NewRequest(method, h.ServerURL+path, nil)
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

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

func (h *E2ETestHelper) CreateTask(t *testing.T, name, description string) *models.Task {
	taskData := map[string]string{
		"name":        name,
		"description": description,
	}

	taskJSON, _ := json.Marshal(taskData)
	resp := h.MakeRequest(t, "POST", "/api/v1/tasks/", bytes.NewBuffer(taskJSON), h.AuthToken)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Task creation failed with status %d", resp.StatusCode)
	}

	var task models.Task
	json.NewDecoder(resp.Body).Decode(&task)
	resp.Body.Close()

	return &task
}

func (h *E2ETestHelper) GetTask(t *testing.T, taskID int) *models.Task {
	resp := h.MakeRequest(t, "GET", fmt.Sprintf("/api/v1/tasks/%d/", taskID), nil, h.AuthToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get task failed with status %d", resp.StatusCode)
	}

	var task models.Task
	json.NewDecoder(resp.Body).Decode(&task)
	resp.Body.Close()

	return &task
}

func (h *E2ETestHelper) UpdateTask(t *testing.T, taskID int, name, description string, status models.TaskStatus) *models.Task {
	taskData := map[string]string{
		"name":        name,
		"description": description,
		"status":      string(status),
	}

	taskJSON, _ := json.Marshal(taskData)
	resp := h.MakeRequest(t, "PUT", fmt.Sprintf("/api/v1/tasks/%d/", taskID), bytes.NewBuffer(taskJSON), h.AuthToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Task update failed with status %d", resp.StatusCode)
	}

	var task models.Task
	json.NewDecoder(resp.Body).Decode(&task)
	resp.Body.Close()

	return &task
}

func (h *E2ETestHelper) DeleteTask(t *testing.T, taskID int) {
	resp := h.MakeRequest(t, "DELETE", fmt.Sprintf("/api/v1/tasks/%d/", taskID), nil, h.AuthToken)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Task deletion failed with status %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func (h *E2ETestHelper) Logout(t *testing.T) {
	resp := h.MakeRequest(t, "POST", "/api/v1/logout", nil, h.AuthToken)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Logout failed with status %d", resp.StatusCode)
	}
	resp.Body.Close()
	h.AuthToken = ""
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}