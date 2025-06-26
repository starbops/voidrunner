package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/starbops/voidrunner/internal/models"
)

// TestFullApplicationE2E tests complete user workflows with real server
func TestFullApplicationE2E(t *testing.T) {
	testCases := []struct {
		name        string
		backendType string
	}{
		{"Memory Backend", "memory"},
		{"PostgreSQL Backend", "postgres"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helper := SetupE2ETestHelper(t, tc.backendType)
			defer helper.TearDown(t)

			// Test complete user workflow
			t.Run("CompleteUserWorkflow", func(t *testing.T) {
				testCompleteUserWorkflow(t, helper)
			})

			// Test server health
			t.Run("ServerHealth", func(t *testing.T) {
				testServerHealth(t, helper)
			})

			// Test authentication lifecycle
			t.Run("AuthenticationLifecycle", func(t *testing.T) {
				testAuthenticationLifecycle(t, helper)
			})
		})
	}
}

func testCompleteUserWorkflow(t *testing.T, helper *E2ETestHelper) {
	// Step 1: Register a new user
	registerData := map[string]string{
		"username":   fmt.Sprintf("workflowuser_%d", time.Now().UnixNano()),
		"email":      fmt.Sprintf("workflow_%d@example.com", time.Now().UnixNano()),
		"password":   "workflowpassword123",
		"first_name": "Workflow",
		"last_name":  "User",
	}

	registerJSON, _ := json.Marshal(registerData)
	resp := helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Registration failed with status %d", resp.StatusCode)
	}

	var user models.User
	err := json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		t.Fatalf("Failed to decode user response: %v", err)
	}
	resp.Body.Close()

	if user.Username != registerData["username"] {
		t.Errorf("Expected username %s, got %s", registerData["username"], user.Username)
	}

	// Step 2: Login with registered user
	loginData := map[string]string{
		"identifier": registerData["username"],
		"password":   registerData["password"],
	}

	loginJSON, _ := json.Marshal(loginData)
	resp = helper.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(loginJSON), "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed with status %d", resp.StatusCode)
	}

	var loginResponse struct {
		Message string `json:"message"`
		Token   string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}
	resp.Body.Close()

	if loginResponse.Token == "" {
		t.Error("Expected non-empty token")
	}

	authToken := loginResponse.Token

	// Step 3: Create multiple tasks
	tasks := []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{
		{"Complete project", "Finish the VoidRunner project"},
		{"Write documentation", "Create comprehensive API documentation"},
		{"Deploy to production", "Set up production environment"},
	}

	var createdTasks []models.Task

	for _, taskData := range tasks {
		taskJSON, _ := json.Marshal(taskData)
		resp = helper.MakeRequest(t, "POST", "/api/v1/tasks/", bytes.NewBuffer(taskJSON), authToken)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Task creation failed with status %d", resp.StatusCode)
		}

		var task models.Task
		err = json.NewDecoder(resp.Body).Decode(&task)
		if err != nil {
			t.Fatalf("Failed to decode task response: %v", err)
		}
		resp.Body.Close()

		if task.Name != taskData.Name {
			t.Errorf("Expected task name %s, got %s", taskData.Name, task.Name)
		}
		if task.Status != models.TaskStatusPending {
			t.Errorf("Expected task status %s, got %s", models.TaskStatusPending, task.Status)
		}

		createdTasks = append(createdTasks, task)
	}

	// Step 4: List all tasks
	resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, authToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List tasks failed with status %d", resp.StatusCode)
	}

	var taskList []models.Task
	err = json.NewDecoder(resp.Body).Decode(&taskList)
	if err != nil {
		t.Fatalf("Failed to decode task list response: %v", err)
	}
	resp.Body.Close()

	if len(taskList) != len(createdTasks) {
		t.Errorf("Expected %d tasks, got %d", len(createdTasks), len(taskList))
	}

	// Step 5: Update task status
	updatedTask := createdTasks[0]
	updateData := map[string]string{
		"name":        updatedTask.Name,
		"description": updatedTask.Description,
		"status":      string(models.TaskStatusCompleted),
	}

	updateJSON, _ := json.Marshal(updateData)
	resp = helper.MakeRequest(t, "PUT", fmt.Sprintf("/api/v1/tasks/%d/", updatedTask.ID), bytes.NewBuffer(updateJSON), authToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Task update failed with status %d", resp.StatusCode)
	}

	var updatedTaskResponse models.Task
	err = json.NewDecoder(resp.Body).Decode(&updatedTaskResponse)
	if err != nil {
		t.Fatalf("Failed to decode updated task response: %v", err)
	}
	resp.Body.Close()

	if updatedTaskResponse.Status != models.TaskStatusCompleted {
		t.Errorf("Expected task status %s, got %s", models.TaskStatusCompleted, updatedTaskResponse.Status)
	}

	// Step 6: Get specific task
	resp = helper.MakeRequest(t, "GET", fmt.Sprintf("/api/v1/tasks/%d/", updatedTask.ID), nil, authToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get task failed with status %d", resp.StatusCode)
	}

	var retrievedTask models.Task
	err = json.NewDecoder(resp.Body).Decode(&retrievedTask)
	if err != nil {
		t.Fatalf("Failed to decode retrieved task response: %v", err)
	}
	resp.Body.Close()

	if retrievedTask.Status != models.TaskStatusCompleted {
		t.Errorf("Expected retrieved task status %s, got %s", models.TaskStatusCompleted, retrievedTask.Status)
	}

	// Step 7: Delete a task
	taskToDelete := createdTasks[1]
	resp = helper.MakeRequest(t, "DELETE", fmt.Sprintf("/api/v1/tasks/%d/", taskToDelete.ID), nil, authToken)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Task deletion failed with status %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify task was deleted
	resp = helper.MakeRequest(t, "GET", fmt.Sprintf("/api/v1/tasks/%d/", taskToDelete.ID), nil, authToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for deleted task, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Step 8: Logout user
	resp = helper.MakeRequest(t, "POST", "/api/v1/logout", nil, authToken)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Logout failed with status %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Step 9: Verify token is invalid after logout
	resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, authToken)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 after logout, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func testServerHealth(t *testing.T, helper *E2ETestHelper) {
	// Test welcome endpoint
	resp := helper.MakeRequest(t, "GET", "/api/v1/welcome", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Welcome endpoint failed with status %d", resp.StatusCode)
	}

	var welcomeResponse struct {
		Message string `json:"message"`
	}
	err := json.NewDecoder(resp.Body).Decode(&welcomeResponse)
	if err != nil {
		t.Fatalf("Failed to decode welcome response: %v", err)
	}
	resp.Body.Close()

	if welcomeResponse.Message != "Welcome to the VoidRunner API" {
		t.Errorf("Expected welcome message, got %s", welcomeResponse.Message)
	}

	// Test invalid endpoint
	resp = helper.MakeRequest(t, "GET", "/api/v1/nonexistent", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for invalid endpoint, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func testAuthenticationLifecycle(t *testing.T, helper *E2ETestHelper) {
	// Test authentication without token
	resp := helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth token, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test with invalid token
	resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, "Bearer invalid.token.here")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 with invalid token, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Register and login user
	token := helper.RegisterAndLoginUser(t)

	// Test with valid token
	resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 with valid token, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Logout
	helper.Logout(t)

	// Test with token after logout
	resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, token)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 after logout, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}