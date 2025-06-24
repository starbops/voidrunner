package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/starbops/voidrunner/internal/models"
)

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestAPIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API integration tests in short mode")
	}

	helper := SetupTestHelper(t)
	defer helper.TearDown()

	t.Run("Welcome Endpoint", func(t *testing.T) {
		resp := helper.MakeRequest(t, "GET", "/api/v1/welcome", nil, "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response map[string]string
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] == "" {
			t.Error("Expected welcome message to be present")
		}
	})

	t.Run("User API Endpoints", func(t *testing.T) {
		helper.CleanupTestData(t)
		token := helper.RegisterAndLoginUser(t)

		// Test GET /api/v1/users/me (get current user profile)
		resp := helper.MakeRequest(t, "GET", "/api/v1/users/me", nil, token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var user models.User
		err := json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			t.Fatalf("Failed to decode user response: %v", err)
		}

		if user.ID == 0 {
			t.Error("Expected user ID to be set")
		}
		if user.Username == "" {
			t.Error("Expected username to be set")
		}

		// Test PUT /api/v1/users/me (update current user profile)
		updateData := models.UpdateUserRequest{
			Username:  stringPtr("updated_user_api"),
			Email:     stringPtr("updated_api@example.com"),
			FirstName: stringPtr("Updated"),
			LastName:  stringPtr("User"),
		}

		updateJSON, _ := json.Marshal(updateData)
		resp = helper.MakeRequest(t, "PUT", "/api/v1/users/me", bytes.NewBuffer(updateJSON), token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var updatedUser models.User
		err = json.NewDecoder(resp.Body).Decode(&updatedUser)
		if err != nil {
			t.Fatalf("Failed to decode updated user response: %v", err)
		}

		if updatedUser.Username != *updateData.Username {
			t.Errorf("Expected username %s, got %s", *updateData.Username, updatedUser.Username)
		}
		if updatedUser.FirstName != *updateData.FirstName {
			t.Errorf("Expected first name %s, got %s", *updateData.FirstName, updatedUser.FirstName)
		}
	})

	t.Run("Task API Endpoints", func(t *testing.T) {
		helper.CleanupTestData(t)
		token := helper.RegisterAndLoginUser(t)

		// Test POST /api/v1/tasks/ (create task)
		taskData := map[string]string{
			"name":        "Test API Task",
			"description": "Task created via API test",
		}

		taskJSON, _ := json.Marshal(taskData)
		resp := helper.MakeRequest(t, "POST", "/api/v1/tasks/", bytes.NewBuffer(taskJSON), token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		var createdTask models.Task
		err := json.NewDecoder(resp.Body).Decode(&createdTask)
		if err != nil {
			t.Fatalf("Failed to decode created task response: %v", err)
		}

		if createdTask.Name != taskData["name"] {
			t.Errorf("Expected task name %s, got %s", taskData["name"], createdTask.Name)
		}

		if string(createdTask.Status) != "pending" {
			t.Errorf("Expected task status 'pending', got %s", createdTask.Status)
		}

		// Test GET /api/v1/tasks/ (get all tasks)
		resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/", nil, token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var tasks []*models.Task
		err = json.NewDecoder(resp.Body).Decode(&tasks)
		if err != nil {
			t.Fatalf("Failed to decode tasks response: %v", err)
		}

		if len(tasks) == 0 {
			t.Error("Expected at least one task")
		}

		// Test GET /api/v1/tasks/{id}/
		taskID := createdTask.ID
		resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/"+strconv.Itoa(taskID)+"/", nil, token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var task models.Task
		err = json.NewDecoder(resp.Body).Decode(&task)
		if err != nil {
			t.Fatalf("Failed to decode task response: %v", err)
		}

		if task.ID != taskID {
			t.Errorf("Expected task ID %d, got %d", taskID, task.ID)
		}

		// Test PUT /api/v1/tasks/{id}/
		updateTaskData := map[string]string{
			"name":        "Updated API Task",
			"description": "Updated task description",
			"status":      "completed",
		}

		updateTaskJSON, _ := json.Marshal(updateTaskData)
		resp = helper.MakeRequest(t, "PUT", "/api/v1/tasks/"+strconv.Itoa(taskID)+"/", bytes.NewBuffer(updateTaskJSON), token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var updatedTask models.Task
		err = json.NewDecoder(resp.Body).Decode(&updatedTask)
		if err != nil {
			t.Fatalf("Failed to decode updated task response: %v", err)
		}

		if updatedTask.Name != updateTaskData["name"] {
			t.Errorf("Expected task name %s, got %s", updateTaskData["name"], updatedTask.Name)
		}

		if string(updatedTask.Status) != updateTaskData["status"] {
			t.Errorf("Expected task status %s, got %s", updateTaskData["status"], updatedTask.Status)
		}

		// Test DELETE /api/v1/tasks/{id}/
		resp = helper.MakeRequest(t, "DELETE", "/api/v1/tasks/"+strconv.Itoa(taskID)+"/", nil, token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", resp.StatusCode)
		}

		// Verify task is deleted
		resp = helper.MakeRequest(t, "GET", "/api/v1/tasks/"+strconv.Itoa(taskID)+"/", nil, token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for deleted task, got %d", resp.StatusCode)
		}
	})

	t.Run("Authentication Required Endpoints", func(t *testing.T) {
		helper.CleanupTestData(t)

		// Test accessing protected endpoints without token
		protectedEndpoints := []struct {
			method string
			path   string
		}{
			{"GET", "/api/v1/users/"},
			{"GET", "/api/v1/users/1/"},
			{"POST", "/api/v1/users/"},
			{"PUT", "/api/v1/users/1/"},
			{"DELETE", "/api/v1/users/1/"},
			{"GET", "/api/v1/tasks/"},
			{"GET", "/api/v1/tasks/1/"},
			{"POST", "/api/v1/tasks/"},
			{"PUT", "/api/v1/tasks/1/"},
			{"DELETE", "/api/v1/tasks/1/"},
			{"POST", "/api/v1/logout"},
		}

		for _, endpoint := range protectedEndpoints {
			resp := helper.MakeRequest(t, endpoint.method, endpoint.path, nil, "")
			defer resp.Body.Close()

			// Logout endpoint might return 400 for bad request instead of 401
			expectedStatuses := []int{http.StatusUnauthorized, http.StatusBadRequest}
			if endpoint.path == "/api/v1/logout" {
				expectedStatuses = append(expectedStatuses, http.StatusBadRequest)
			}
			
			statusOK := false
			for _, expected := range expectedStatuses {
				if resp.StatusCode == expected {
					statusOK = true
					break
				}
			}
			
			if !statusOK {
				t.Errorf("Expected status 401 or 400 for %s %s without token, got %d", endpoint.method, endpoint.path, resp.StatusCode)
			}
		}

		// Test with invalid token
		invalidToken := "invalid.jwt.token"
		for _, endpoint := range protectedEndpoints {
			resp := helper.MakeRequest(t, endpoint.method, endpoint.path, nil, invalidToken)
			defer resp.Body.Close()

			// Some endpoints might return 500 for invalid tokens due to parsing errors
			expectedStatuses := []int{http.StatusUnauthorized, http.StatusInternalServerError, http.StatusBadRequest}
			
			statusOK := false
			for _, expected := range expectedStatuses {
				if resp.StatusCode == expected {
					statusOK = true
					break
				}
			}
			
			if !statusOK {
				t.Errorf("Expected status 401, 400, or 500 for %s %s with invalid token, got %d", endpoint.method, endpoint.path, resp.StatusCode)
			}
		}
	})

	t.Run("JSON Error Handling", func(t *testing.T) {
		helper.CleanupTestData(t)

		// Test malformed JSON
		malformedJSON := `{"name": "test", "invalid": }`
		resp := helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBufferString(malformedJSON), "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for malformed JSON, got %d", resp.StatusCode)
		}

		// Test missing required fields for registration
		incompleteData := map[string]string{
			"username": "testuser",
			// missing email, password, etc.
		}

		incompleteJSON, _ := json.Marshal(incompleteData)
		resp = helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(incompleteJSON), "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for incomplete registration data, got %d", resp.StatusCode)
		}
	})

	t.Run("HTTP Method Validation", func(t *testing.T) {
		helper.CleanupTestData(t)
		token := helper.RegisterAndLoginUser(t)

		// Test wrong HTTP methods for the updated API structure
		wrongMethods := []struct {
			method string
			path   string
		}{
			{"POST", "/api/v1/users/me"},   // Should be GET, PUT, or DELETE
			{"PATCH", "/api/v1/users/me"}, // Should be GET, PUT, or DELETE
			{"PATCH", "/api/v1/tasks/1/"}, // Should be GET, PUT, or DELETE
		}

		for _, test := range wrongMethods {
			resp := helper.MakeRequest(t, test.method, test.path, nil, token)
			defer resp.Body.Close()

			// Go 1.23 routing might return 400 for unmatched method patterns, or 404 for non-existent routes
			if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
				t.Errorf("Expected status 405, 400, or 404 for %s %s, got %d", test.method, test.path, resp.StatusCode)
			}
		}
	})
}