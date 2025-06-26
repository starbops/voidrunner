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

// TestBackendCompatibilityE2E tests cross-backend compatibility
func TestBackendCompatibilityE2E(t *testing.T) {
	// Test memory backend
	t.Run("MemoryBackend", func(t *testing.T) {
		helper := SetupE2ETestHelper(t, "memory")
		defer helper.TearDown(t)

		testBackendFunctionality(t, helper, "memory")
	})

	// Test PostgreSQL backend
	t.Run("PostgreSQLBackend", func(t *testing.T) {
		helper := SetupE2ETestHelper(t, "postgres")
		defer helper.TearDown(t)

		testBackendFunctionality(t, helper, "postgres")
	})

	// Test backend consistency
	t.Run("BackendConsistency", func(t *testing.T) {
		testBackendConsistency(t)
	})
}

func testBackendFunctionality(t *testing.T, helper *E2ETestHelper, backendType string) {
	t.Logf("Testing %s backend functionality", backendType)

	// Register and login user
	token := helper.RegisterAndLoginUser(t)

	// Test basic CRUD operations
	t.Run("BasicCRUD", func(t *testing.T) {
		// Create task
		task := helper.CreateTask(t, fmt.Sprintf("Backend Test Task (%s)", backendType), "Testing backend functionality")
		
		if task.Name != fmt.Sprintf("Backend Test Task (%s)", backendType) {
			t.Errorf("Expected task name with backend type, got %s", task.Name)
		}

		// Read task
		retrievedTask := helper.GetTask(t, task.ID)
		if retrievedTask.ID != task.ID {
			t.Errorf("Expected task ID %d, got %d", task.ID, retrievedTask.ID)
		}

		// Update task
		updatedTask := helper.UpdateTask(t, task.ID, "Updated Task Name", "Updated description", models.TaskStatusCompleted)
		if updatedTask.Status != models.TaskStatusCompleted {
			t.Errorf("Expected status %s, got %s", models.TaskStatusCompleted, updatedTask.Status)
		}

		// Delete task
		helper.DeleteTask(t, task.ID)

		// Verify deletion
		resp := helper.MakeRequest(t, "GET", fmt.Sprintf("/api/v1/tasks/%d/", task.ID), nil, token)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 for deleted task, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	// Test data persistence (for memory backend, this tests in-memory consistency)
	t.Run("DataPersistence", func(t *testing.T) {
		tasks := make([]*models.Task, 3)
		
		// Create multiple tasks
		for i := 0; i < 3; i++ {
			task := helper.CreateTask(t, fmt.Sprintf("Persistence Task %d", i), fmt.Sprintf("Testing persistence %d", i))
			tasks[i] = task
		}

		// List all tasks
		resp := helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, token)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Failed to list tasks with status %d", resp.StatusCode)
		}

		var taskList []models.Task
		err := json.NewDecoder(resp.Body).Decode(&taskList)
		if err != nil {
			t.Fatalf("Failed to decode task list for %s backend: %v", backendType, err)
		}
		resp.Body.Close()

		if len(taskList) < 3 {
			t.Errorf("Expected at least 3 tasks, got %d", len(taskList))
		}

		// Verify all created tasks exist
		taskIDs := make(map[int]bool)
		for _, task := range taskList {
			taskIDs[task.ID] = true
		}

		for i, task := range tasks {
			if !taskIDs[task.ID] {
				t.Errorf("Task %d (ID: %d) not found in task list", i, task.ID)
			}
		}
	})

	// Test user management
	t.Run("UserManagement", func(t *testing.T) {
		// Register additional user
		registerData := map[string]string{
			"username":   fmt.Sprintf("backenduser_%s_%d", backendType, time.Now().UnixNano()),
			"email":      fmt.Sprintf("backend_%s_%d@example.com", backendType, time.Now().UnixNano()),
			"password":   "backendpassword123",
			"first_name": "Backend",
			"last_name":  "User",
		}

		registerJSON, _ := json.Marshal(registerData)
		resp := helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("User registration failed with status %d", resp.StatusCode)
		}

		var user models.User
		err := json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			t.Fatalf("Failed to decode user registration response for %s backend: %v", backendType, err)
		}
		resp.Body.Close()

		// Verify user data
		if user.Username != registerData["username"] {
			t.Errorf("Expected username %s, got %s", registerData["username"], user.Username)
		}
		if user.Email != registerData["email"] {
			t.Errorf("Expected email %s, got %s", registerData["email"], user.Email)
		}

		// Login with new user
		loginData := map[string]string{
			"identifier": registerData["username"],
			"password":   registerData["password"],
		}

		loginJSON, _ := json.Marshal(loginData)
		resp = helper.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(loginJSON), "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("User login failed with status %d", resp.StatusCode)
		}

		var loginResponse struct {
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&loginResponse)
		resp.Body.Close()

		if loginResponse.Token == "" {
			t.Error("Expected non-empty token")
		}

		// Test authenticated request with new user
		resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, loginResponse.Token)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 for authenticated request, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}

func testBackendConsistency(t *testing.T) {
	// This test ensures both backends behave identically
	testScenarios := []struct {
		name        string
		description string
		status      models.TaskStatus
	}{
		{"Consistency Task 1", "First task for consistency testing", models.TaskStatusPending},
		{"Consistency Task 2", "Second task for consistency testing", models.TaskStatusInProgress},
		{"Consistency Task 3", "Third task for consistency testing", models.TaskStatusCompleted},
	}

	// Test with memory backend
	memoryHelper := SetupE2ETestHelper(t, "memory")
	defer memoryHelper.TearDown(t)

	memoryToken := memoryHelper.RegisterAndLoginUser(t)
	memoryTasks := make([]*models.Task, len(testScenarios))

	for i, scenario := range testScenarios {
		task := memoryHelper.CreateTask(t, scenario.name, scenario.description)
		memoryHelper.UpdateTask(t, task.ID, scenario.name, scenario.description, scenario.status)
		memoryTasks[i] = task
	}

	// Test with PostgreSQL backend
	postgresHelper := SetupE2ETestHelper(t, "postgres")
	defer postgresHelper.TearDown(t)

	postgresToken := postgresHelper.RegisterAndLoginUser(t)
	postgresTasks := make([]*models.Task, len(testScenarios))

	for i, scenario := range testScenarios {
		task := postgresHelper.CreateTask(t, scenario.name, scenario.description)
		postgresHelper.UpdateTask(t, task.ID, scenario.name, scenario.description, scenario.status)
		postgresTasks[i] = task
	}

	// Compare task lists from both backends
	memoryResp := memoryHelper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, memoryToken)
	if memoryResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to get memory tasks with status %d", memoryResp.StatusCode)
	}

	var memoryTaskList []models.Task
	json.NewDecoder(memoryResp.Body).Decode(&memoryTaskList)
	memoryResp.Body.Close()

	postgresResp := postgresHelper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, postgresToken)
	if postgresResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to get postgres tasks with status %d", postgresResp.StatusCode)
	}

	var postgresTaskList []models.Task
	json.NewDecoder(postgresResp.Body).Decode(&postgresTaskList)
	postgresResp.Body.Close()

	// Both backends should have at least the test scenarios we created
	if len(memoryTaskList) < len(testScenarios) {
		t.Errorf("Memory backend has fewer tasks than expected: got=%d, expected at least=%d", len(memoryTaskList), len(testScenarios))
	}
	if len(postgresTaskList) < len(testScenarios) {
		t.Errorf("PostgreSQL backend has fewer tasks than expected: got=%d, expected at least=%d", len(postgresTaskList), len(testScenarios))
	}

	// Verify that both backends contain the expected tasks (order-independent)
	expectedTasks := make(map[string]struct {
		description string
		status      models.TaskStatus
	})
	
	for _, scenario := range testScenarios {
		expectedTasks[scenario.name] = struct {
			description string
			status      models.TaskStatus
		}{scenario.description, scenario.status}
	}

	// Check memory backend has all expected tasks
	memoryTaskMap := make(map[string]models.Task)
	for _, task := range memoryTaskList {
		if _, isExpected := expectedTasks[task.Name]; isExpected {
			memoryTaskMap[task.Name] = task
		}
	}

	// Check PostgreSQL backend has all expected tasks
	postgresTaskMap := make(map[string]models.Task)
	for _, task := range postgresTaskList {
		if _, isExpected := expectedTasks[task.Name]; isExpected {
			postgresTaskMap[task.Name] = task
		}
	}

	// Verify both backends have all expected tasks
	for taskName, expected := range expectedTasks {
		memoryTask, memoryHasTask := memoryTaskMap[taskName]
		postgresTask, postgresHasTask := postgresTaskMap[taskName]

		if !memoryHasTask {
			t.Errorf("Memory backend missing task: %s", taskName)
			continue
		}
		if !postgresHasTask {
			t.Errorf("PostgreSQL backend missing task: %s", taskName)
			continue
		}

		// Verify task properties match expected values
		if memoryTask.Description != expected.description {
			t.Errorf("Memory task %s description mismatch: expected=%s, got=%s", taskName, expected.description, memoryTask.Description)
		}
		if postgresTask.Description != expected.description {
			t.Errorf("PostgreSQL task %s description mismatch: expected=%s, got=%s", taskName, expected.description, postgresTask.Description)
		}
		if memoryTask.Status != expected.status {
			t.Errorf("Memory task %s status mismatch: expected=%s, got=%s", taskName, expected.status, memoryTask.Status)
		}
		if postgresTask.Status != expected.status {
			t.Errorf("PostgreSQL task %s status mismatch: expected=%s, got=%s", taskName, expected.status, postgresTask.Status)
		}
	}

	// Test update operations consistency
	if len(memoryTasks) > 0 && len(postgresTasks) > 0 {
		// Update first task in both backends
		updatedMemoryTask := memoryHelper.UpdateTask(t, memoryTasks[0].ID, "Updated Consistency Task", "Updated description", models.TaskStatusCompleted)
		updatedPostgresTask := postgresHelper.UpdateTask(t, postgresTasks[0].ID, "Updated Consistency Task", "Updated description", models.TaskStatusCompleted)

		// Verify updates are consistent
		if updatedMemoryTask.Name != updatedPostgresTask.Name {
			t.Errorf("Updated task name mismatch: memory=%s, postgres=%s", updatedMemoryTask.Name, updatedPostgresTask.Name)
		}

		if updatedMemoryTask.Status != updatedPostgresTask.Status {
			t.Errorf("Updated task status mismatch: memory=%s, postgres=%s", updatedMemoryTask.Status, updatedPostgresTask.Status)
		}
	}

	// Test deletion consistency
	if len(memoryTasks) > 1 && len(postgresTasks) > 1 {
		// Delete second task in both backends
		memoryHelper.DeleteTask(t, memoryTasks[1].ID)
		postgresHelper.DeleteTask(t, postgresTasks[1].ID)

		// Verify both return 404 for deleted tasks
		memoryResp := memoryHelper.MakeRequest(t, "GET", fmt.Sprintf("/api/v1/tasks/%d/", memoryTasks[1].ID), nil, memoryToken)
		postgresResp := postgresHelper.MakeRequest(t, "GET", fmt.Sprintf("/api/v1/tasks/%d/", postgresTasks[1].ID), nil, postgresToken)

		if memoryResp.StatusCode != http.StatusNotFound {
			t.Errorf("Memory backend should return 404 for deleted task, got %d", memoryResp.StatusCode)
		}

		if postgresResp.StatusCode != http.StatusNotFound {
			t.Errorf("Postgres backend should return 404 for deleted task, got %d", postgresResp.StatusCode)
		}

		memoryResp.Body.Close()
		postgresResp.Body.Close()
	}
}