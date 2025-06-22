package integration

import (
	"testing"
	"time"

	"github.com/starbops/voidrunner/internal/models"
)

func TestDatabaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration tests in short mode")
	}

	helper := SetupTestHelper(t)
	defer helper.TearDown()

	t.Run("User Repository Operations", func(t *testing.T) {
		helper.CleanupTestData(t)

		// Test Create User
		user := &models.User{
			Username:     "testuser_db",
			Email:        "testdb@example.com",
			FirstName:    "Test",
			LastName:     "User",
			PasswordHash: "$2a$10$samplehash",
		}

		createdUser, err := helper.UserRepo.Create(user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if createdUser.ID == 0 {
			t.Error("Expected user ID to be set")
		}

		if createdUser.CreatedAt == "" {
			t.Error("Expected CreatedAt to be set")
		}

		// Test Get User
		retrievedUser, err := helper.UserRepo.GetUser(createdUser.ID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if retrievedUser.Username != user.Username {
			t.Errorf("Expected username %s, got %s", user.Username, retrievedUser.Username)
		}

		// Test GetByUsernameOrEmail
		foundUser, err := helper.UserRepo.GetByUsernameOrEmail(user.Username, "")
		if err != nil {
			t.Fatalf("Failed to get user by username: %v", err)
		}

		if foundUser.ID != createdUser.ID {
			t.Errorf("Expected user ID %d, got %d", createdUser.ID, foundUser.ID)
		}

		foundUser, err = helper.UserRepo.GetByUsernameOrEmail("", user.Email)
		if err != nil {
			t.Fatalf("Failed to get user by email: %v", err)
		}

		if foundUser.ID != createdUser.ID {
			t.Errorf("Expected user ID %d, got %d", createdUser.ID, foundUser.ID)
		}

		// Test Update User
		updatedUser := &models.User{
			Username:  "updated_user",
			Email:     "updated@example.com",
			FirstName: "Updated",
			LastName:  "User",
		}

		result, err := helper.UserRepo.UpdateUser(createdUser.ID, updatedUser)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		if result.Username != updatedUser.Username {
			t.Errorf("Expected username %s, got %s", updatedUser.Username, result.Username)
		}

		// Test Get All Users
		users, err := helper.UserRepo.GetUsers()
		if err != nil {
			t.Fatalf("Failed to get users: %v", err)
		}

		if len(users) == 0 {
			t.Error("Expected at least one user")
		}

		// Test Delete User
		err = helper.UserRepo.DeleteUser(createdUser.ID)
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// Verify user is deleted
		deletedUser, err := helper.UserRepo.GetUser(createdUser.ID)
		if err != nil {
			t.Fatalf("Error checking deleted user: %v", err)
		}

		if deletedUser != nil {
			t.Error("Expected user to be deleted")
		}
	})

	t.Run("Task Repository Operations", func(t *testing.T) {
		helper.CleanupTestData(t)

		// Test Create Task
		task := &models.Task{
			Name:        "Test Task",
			Description: "Test task description",
			Status:      models.TaskStatusPending,
		}

		createdTask, err := helper.TaskRepo.CreateTask(task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		if createdTask.ID == 0 {
			t.Error("Expected task ID to be set")
		}

		// Test Get Task
		retrievedTask, err := helper.TaskRepo.GetTask(createdTask.ID)
		if err != nil {
			t.Fatalf("Failed to get task: %v", err)
		}

		if retrievedTask.Name != task.Name {
			t.Errorf("Expected task name %s, got %s", task.Name, retrievedTask.Name)
		}

		if retrievedTask.Status != task.Status {
			t.Errorf("Expected task status %s, got %s", task.Status, retrievedTask.Status)
		}

		// Test Update Task
		updatedTask := &models.Task{
			Name:        "Updated Task",
			Description: "Updated description",
			Status:      models.TaskStatusCompleted,
		}

		result, err := helper.TaskRepo.UpdateTask(createdTask.ID, updatedTask)
		if err != nil {
			t.Fatalf("Failed to update task: %v", err)
		}

		if result.Name != updatedTask.Name {
			t.Errorf("Expected task name %s, got %s", updatedTask.Name, result.Name)
		}

		if result.Status != updatedTask.Status {
			t.Errorf("Expected task status %s, got %s", updatedTask.Status, result.Status)
		}

		// Test Get All Tasks
		tasks, err := helper.TaskRepo.GetTasks()
		if err != nil {
			t.Fatalf("Failed to get tasks: %v", err)
		}

		if len(tasks) == 0 {
			t.Error("Expected at least one task")
		}

		// Test Delete Task
		err = helper.TaskRepo.DeleteTask(createdTask.ID)
		if err != nil {
			t.Fatalf("Failed to delete task: %v", err)
		}

		// Verify task is deleted
		deletedTask, err := helper.TaskRepo.GetTask(createdTask.ID)
		if err != nil {
			t.Fatalf("Error checking deleted task: %v", err)
		}

		if deletedTask != nil {
			t.Error("Expected task to be deleted")
		}
	})

	t.Run("Database Connection Handling", func(t *testing.T) {
		// Test connection resilience by creating multiple operations
		helper.CleanupTestData(t)

		for i := 0; i < 5; i++ {
			user := &models.User{
				Username:     "testuser_" + string(rune(i+97)), // a, b, c, etc.
				Email:        "test" + string(rune(i+97)) + "@example.com",
				FirstName:    "Test",
				LastName:     "User",
				PasswordHash: "$2a$10$samplehash",
			}

			_, err := helper.UserRepo.Create(user)
			if err != nil {
				t.Fatalf("Failed to create user %d: %v", i, err)
			}

			// Small delay to test concurrent access
			time.Sleep(10 * time.Millisecond)
		}

		users, err := helper.UserRepo.GetUsers()
		if err != nil {
			t.Fatalf("Failed to get users: %v", err)
		}

		if len(users) != 5 {
			t.Errorf("Expected 5 users, got %d", len(users))
		}
	})

	t.Run("Transaction Rollback Scenarios", func(t *testing.T) {
		helper.CleanupTestData(t)

		// Test creating user with duplicate username (should fail)
		user1 := &models.User{
			Username:     "duplicate_user",
			Email:        "user1@example.com",
			FirstName:    "Test",
			LastName:     "User",
			PasswordHash: "$2a$10$samplehash",
		}

		_, err := helper.UserRepo.Create(user1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Try to create user with same username
		user2 := &models.User{
			Username:     "duplicate_user",
			Email:        "user2@example.com",
			FirstName:    "Test",
			LastName:     "User",
			PasswordHash: "$2a$10$samplehash",
		}

		_, err = helper.UserRepo.Create(user2)
		if err == nil {
			t.Error("Expected error when creating user with duplicate username")
		}

		// Verify only one user exists
		users, err := helper.UserRepo.GetUsers()
		if err != nil {
			t.Fatalf("Failed to get users: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user after duplicate creation, got %d", len(users))
		}
	})
}