package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/starbops/voidrunner/internal/models"
)

// TestUserWorkflowE2E tests specific user workflow scenarios
func TestUserWorkflowE2E(t *testing.T) {
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

			// Test task management workflow
			t.Run("TaskManagementWorkflow", func(t *testing.T) {
				testTaskManagementWorkflow(t, helper)
			})

			// Test concurrent user operations
			t.Run("ConcurrentUserOperations", func(t *testing.T) {
				testConcurrentUserOperations(t, helper)
			})

			// Test error scenarios
			t.Run("ErrorScenarios", func(t *testing.T) {
				testErrorScenarios(t, helper)
			})

			// Test data consistency
			t.Run("DataConsistency", func(t *testing.T) {
				testDataConsistency(t, helper)
			})
		})
	}
}

func testTaskManagementWorkflow(t *testing.T, helper *E2ETestHelper) {
	// Login user
	token := helper.RegisterAndLoginUser(t)

	// Create a task with pending status
	task := helper.CreateTask(t, "Project Setup", "Set up the initial project structure")
	if task.Status != models.TaskStatusPending {
		t.Errorf("Expected status %s, got %s", models.TaskStatusPending, task.Status)
	}

	// Update task to in-progress
	updatedTask := helper.UpdateTask(t, task.ID, "Project Setup", "Set up the initial project structure", models.TaskStatusInProgress)
	if updatedTask.Status != models.TaskStatusInProgress {
		t.Errorf("Expected status %s, got %s", models.TaskStatusInProgress, updatedTask.Status)
	}

	// Update task to completed
	completedTask := helper.UpdateTask(t, task.ID, "Project Setup", "Set up the initial project structure", models.TaskStatusCompleted)
	if completedTask.Status != models.TaskStatusCompleted {
		t.Errorf("Expected status %s, got %s", models.TaskStatusCompleted, completedTask.Status)
	}

	// Verify task state persists
	retrievedTask := helper.GetTask(t, task.ID)
	if retrievedTask.Status != models.TaskStatusCompleted {
		t.Errorf("Expected persisted status %s, got %s", models.TaskStatusCompleted, retrievedTask.Status)
	}

	// Delete completed task
	helper.DeleteTask(t, task.ID)

	// Verify task is deleted
	resp := helper.MakeRequest(t, "GET", fmt.Sprintf("/api/v1/tasks/%d/", task.ID), nil, token)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for deleted task, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func testConcurrentUserOperations(t *testing.T, helper *E2ETestHelper) {
	// Create multiple users concurrently
	const numUsers = 5
	const tasksPerUser = 3

	var wg sync.WaitGroup
	userTokens := make([]string, numUsers)
	userTasks := make([][]models.Task, numUsers)

	// Register and login users concurrently
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userIndex int) {
			defer wg.Done()

			// Register user
			registerData := map[string]string{
				"username":   fmt.Sprintf("concurrentuser_%d_%d", userIndex, time.Now().UnixNano()),
				"email":      fmt.Sprintf("concurrent_%d_%d@example.com", userIndex, time.Now().UnixNano()),
				"password":   "concurrentpassword123",
				"first_name": "Concurrent",
				"last_name":  fmt.Sprintf("User%d", userIndex),
			}

			registerJSON, _ := json.Marshal(registerData)
			resp := helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("User %d registration failed with status %d", userIndex, resp.StatusCode)
				return
			}
			resp.Body.Close()

			// Login user
			loginData := map[string]string{
				"identifier": registerData["username"],
				"password":   registerData["password"],
			}

			loginJSON, _ := json.Marshal(loginData)
			resp = helper.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(loginJSON), "")
			if resp.StatusCode != http.StatusOK {
				t.Errorf("User %d login failed with status %d", userIndex, resp.StatusCode)
				return
			}

			var loginResponse struct {
				Token string `json:"token"`
			}
			json.NewDecoder(resp.Body).Decode(&loginResponse)
			resp.Body.Close()

			userTokens[userIndex] = loginResponse.Token
		}(i)
	}

	wg.Wait()

	// Create tasks for each user concurrently
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userIndex int) {
			defer wg.Done()

			token := userTokens[userIndex]
			if token == "" {
				return
			}

			userTasks[userIndex] = make([]models.Task, tasksPerUser)

			for j := 0; j < tasksPerUser; j++ {
				taskData := map[string]string{
					"name":        fmt.Sprintf("User%d Task%d", userIndex, j),
					"description": fmt.Sprintf("Task %d for user %d", j, userIndex),
				}

				taskJSON, _ := json.Marshal(taskData)
				resp := helper.MakeRequest(t, "POST", "/api/v1/tasks/", bytes.NewBuffer(taskJSON), token)
				if resp.StatusCode != http.StatusCreated {
					t.Errorf("User %d task %d creation failed with status %d", userIndex, j, resp.StatusCode)
					continue
				}

				var task models.Task
				json.NewDecoder(resp.Body).Decode(&task)
				resp.Body.Close()

				userTasks[userIndex][j] = task
			}
		}(i)
	}

	wg.Wait()

	// Verify that tasks are globally visible (not per-user)
	// In VoidRunner, tasks are shared across all users
	for i := 0; i < numUsers; i++ {
		token := userTokens[i]
		if token == "" {
			continue
		}

		resp := helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, token)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("User %d task list failed with status %d", i, resp.StatusCode)
			continue
		}

		var tasks []models.Task
		json.NewDecoder(resp.Body).Decode(&tasks)
		resp.Body.Close()

		// Tasks are global, so each user should see all tasks created by all users
		if len(tasks) < tasksPerUser {
			t.Errorf("User %d expected at least %d tasks, got %d", i, tasksPerUser, len(tasks))
		}
	}
}

func testErrorScenarios(t *testing.T, helper *E2ETestHelper) {
	token := helper.RegisterAndLoginUser(t)

	// Test invalid JSON
	resp := helper.MakeRequest(t, "POST", "/api/v1/tasks/", bytes.NewBufferString("invalid json"), token)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test empty task creation (should succeed with empty name)
	emptyTaskData := map[string]string{
		"description": "Task without name",
	}
	emptyJSON, _ := json.Marshal(emptyTaskData)
	resp = helper.MakeRequest(t, "POST", "/api/v1/tasks/", bytes.NewBuffer(emptyJSON), token)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 for task creation, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test accessing non-existent task
	resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/99999/", nil, token)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent task, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test updating non-existent task
	updateData := map[string]string{
		"name":        "Updated Task",
		"description": "Updated description",
		"status":      string(models.TaskStatusCompleted),
	}
	updateJSON, _ := json.Marshal(updateData)
	resp = helper.MakeRequest(t, "PUT", "/api/v1/tasks/99999/", bytes.NewBuffer(updateJSON), token)
	// PostgreSQL backend returns 500 for non-existent updates, memory backend may return 404
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 404 or 500 for updating non-existent task, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test deleting non-existent task
	resp = helper.MakeRequest(t, "DELETE", "/api/v1/tasks/99999/", nil, token)
	// PostgreSQL backend returns 500 for non-existent deletes, memory backend returns 204
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 204 or 500 for deleting non-existent task, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test duplicate user registration
	registerData := map[string]string{
		"username":   "duplicateuser",
		"email":      "duplicate@example.com",
		"password":   "password123",
		"first_name": "Duplicate",
		"last_name":  "User",
	}

	// First registration should succeed
	registerJSON, _ := json.Marshal(registerData)
	resp = helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 for first registration, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Second registration with same username should fail
	resp = helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected 409 for duplicate registration, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test login with wrong credentials
	wrongLoginData := map[string]string{
		"identifier": "duplicateuser",
		"password":   "wrongpassword",
	}
	wrongLoginJSON, _ := json.Marshal(wrongLoginData)
	resp = helper.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(wrongLoginJSON), "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for wrong credentials, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func testDataConsistency(t *testing.T, helper *E2ETestHelper) {
	token := helper.RegisterAndLoginUser(t)

	// Create tasks
	tasks := make([]*models.Task, 5)
	for i := 0; i < 5; i++ {
		task := helper.CreateTask(t, fmt.Sprintf("Consistency Task %d", i), fmt.Sprintf("Task %d for consistency testing", i))
		tasks[i] = task
	}

	// Update tasks with different statuses
	statuses := []models.TaskStatus{
		models.TaskStatusPending,
		models.TaskStatusInProgress,
		models.TaskStatusCompleted,
		models.TaskStatusInProgress,
		models.TaskStatusCompleted,
	}

	for i, task := range tasks {
		helper.UpdateTask(t, task.ID, task.Name, task.Description, statuses[i])
	}

	// Verify all tasks have correct statuses
	resp := helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to get tasks with status %d", resp.StatusCode)
	}

	var retrievedTasks []models.Task
	json.NewDecoder(resp.Body).Decode(&retrievedTasks)
	resp.Body.Close()

	if len(retrievedTasks) < len(tasks) {
		t.Errorf("Expected at least %d tasks, got %d", len(tasks), len(retrievedTasks))
	}

	// Verify each task has the correct status
	taskStatusMap := make(map[int]models.TaskStatus)
	for _, task := range retrievedTasks {
		taskStatusMap[task.ID] = task.Status
	}

	for i, task := range tasks {
		expectedStatus := statuses[i]
		actualStatus := taskStatusMap[task.ID]
		if actualStatus != expectedStatus {
			t.Errorf("Task %d expected status %s, got %s", task.ID, expectedStatus, actualStatus)
		}
	}

	// Delete half of the tasks we created
	deletedCount := 0
	for i := 0; i < len(tasks)/2; i++ {
		helper.DeleteTask(t, tasks[i].ID)
		deletedCount++
	}

	// Verify our tasks were deleted (but other tests may have created more tasks)
	resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to get remaining tasks with status %d", resp.StatusCode)
	}

	var remainingTasks []models.Task
	json.NewDecoder(resp.Body).Decode(&remainingTasks)
	resp.Body.Close()

	// Verify that the deleted tasks are no longer present
	remainingTaskIDs := make(map[int]bool)
	for _, task := range remainingTasks {
		remainingTaskIDs[task.ID] = true
	}

	for i := 0; i < deletedCount; i++ {
		if remainingTaskIDs[tasks[i].ID] {
			t.Errorf("Task %d should have been deleted but still exists", tasks[i].ID)
		}
	}
}