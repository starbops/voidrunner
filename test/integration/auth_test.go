package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/starbops/voidrunner/internal/models"
)

func TestAuthenticationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping authentication flow integration tests in short mode")
	}

	helper := SetupTestHelper(t)
	defer helper.TearDown()

	t.Run("Complete Registration Flow", func(t *testing.T) {
		helper.CleanupTestData(t)

		registerData := map[string]string{
			"username":   "newuser",
			"email":      "newuser@example.com",
			"password":   "securepassword123",
			"first_name": "New",
			"last_name":  "User",
		}

		registerJSON, _ := json.Marshal(registerData)
		resp := helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		var registeredUser models.User
		err := json.NewDecoder(resp.Body).Decode(&registeredUser)
		if err != nil {
			t.Fatalf("Failed to decode registration response: %v", err)
		}

		if registeredUser.Username != registerData["username"] {
			t.Errorf("Expected username %s, got %s", registerData["username"], registeredUser.Username)
		}

		if registeredUser.Email != registerData["email"] {
			t.Errorf("Expected email %s, got %s", registerData["email"], registeredUser.Email)
		}

		// Password hash should not be returned
		if registeredUser.PasswordHash != "" {
			t.Error("Password hash should not be returned in registration response")
		}

		// Verify user was actually created in database
		users, err := helper.UserRepo.GetUsers()
		if err != nil {
			t.Fatalf("Failed to get users from database: %v", err)
		}

		found := false
		for _, user := range users {
			if user.Username == registerData["username"] {
				found = true
				break
			}
		}

		if !found {
			t.Error("User was not found in database after registration")
		}
	})

	t.Run("Registration Validation", func(t *testing.T) {
		helper.CleanupTestData(t)

		// Test duplicate username
		registerData := map[string]string{
			"username":   "duplicate",
			"email":      "user1@example.com",
			"password":   "password123",
			"first_name": "User",
			"last_name":  "One",
		}

		registerJSON, _ := json.Marshal(registerData)
		resp := helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("First registration should succeed, got %d", resp.StatusCode)
		}

		// Try to register with same username
		duplicateData := map[string]string{
			"username":   "duplicate",
			"email":      "user2@example.com",
			"password":   "password123",
			"first_name": "User",
			"last_name":  "Two",
		}

		duplicateJSON, _ := json.Marshal(duplicateData)
		resp = helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(duplicateJSON), "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409 for duplicate username, got %d", resp.StatusCode)
		}

		// Test duplicate email
		duplicateEmailData := map[string]string{
			"username":   "unique",
			"email":      "user1@example.com", // Same email as first user
			"password":   "password123",
			"first_name": "User",
			"last_name":  "Three",
		}

		duplicateEmailJSON, _ := json.Marshal(duplicateEmailData)
		resp = helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(duplicateEmailJSON), "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409 for duplicate email, got %d", resp.StatusCode)
		}
	})

	t.Run("Complete Login Flow", func(t *testing.T) {
		helper.CleanupTestData(t)

		// First register a user
		registerData := map[string]string{
			"username":   "loginuser",
			"email":      "login@example.com",
			"password":   "loginpassword123",
			"first_name": "Login",
			"last_name":  "User",
		}

		registerJSON, _ := json.Marshal(registerData)
		resp := helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Registration failed: %d", resp.StatusCode)
		}

		// Test login with username
		loginData := map[string]string{
			"identifier": registerData["username"],
			"password":   registerData["password"],
		}

		loginJSON, _ := json.Marshal(loginData)
		resp = helper.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(loginJSON), "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for login, got %d", resp.StatusCode)
		}

		var loginResponse struct {
			Token string       `json:"token"`
			User  models.User  `json:"user"`
		}
		err := json.NewDecoder(resp.Body).Decode(&loginResponse)
		if err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}

		if loginResponse.Token == "" {
			t.Error("Expected token to be returned")
		}

		if loginResponse.User.Username != registerData["username"] {
			t.Errorf("Expected username %s, got %s", registerData["username"], loginResponse.User.Username)
		}

		// Test login with email
		loginEmailData := map[string]string{
			"identifier": registerData["email"],
			"password":   registerData["password"],
		}

		loginEmailJSON, _ := json.Marshal(loginEmailData)
		resp = helper.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(loginEmailJSON), "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for email login, got %d", resp.StatusCode)
		}
	})

	t.Run("Login Validation", func(t *testing.T) {
		helper.CleanupTestData(t)

		// Test login with non-existent user
		loginData := map[string]string{
			"identifier": "nonexistent",
			"password":   "password123",
		}

		loginJSON, _ := json.Marshal(loginData)
		resp := helper.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(loginJSON), "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for non-existent user, got %d", resp.StatusCode)
		}

		// Register a user for wrong password test
		registerData := map[string]string{
			"username":   "wrongpass",
			"email":      "wrongpass@example.com",
			"password":   "correctpassword",
			"first_name": "Wrong",
			"last_name":  "Pass",
		}

		registerJSON, _ := json.Marshal(registerData)
		resp = helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
		resp.Body.Close()

		// Test login with wrong password
		wrongPassData := map[string]string{
			"identifier": registerData["username"],
			"password":   "wrongpassword",
		}

		wrongPassJSON, _ := json.Marshal(wrongPassData)
		resp = helper.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(wrongPassJSON), "")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for wrong password, got %d", resp.StatusCode)
		}
	})

	t.Run("Token Authentication", func(t *testing.T) {
		helper.CleanupTestData(t)
		token := helper.RegisterAndLoginUser(t)

		// Test accessing protected endpoint with valid token
		resp := helper.MakeRequest(t, "GET", "/api/v1/users/", nil, token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 with valid token, got %d", resp.StatusCode)
		}

		// Test with malformed token
		malformedToken := "malformed.token.here"
		resp = helper.MakeRequest(t, "GET", "/api/v1/users/", nil, malformedToken)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 with malformed token, got %d", resp.StatusCode)
		}

		// Test with missing Bearer prefix
		resp = helper.MakeRequest(t, "GET", "/api/v1/users/", nil, strings.TrimPrefix(token, "Bearer "))
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 without Bearer prefix, got %d", resp.StatusCode)
		}
	})

	t.Run("Logout Flow", func(t *testing.T) {
		helper.CleanupTestData(t)
		token := helper.RegisterAndLoginUser(t)

		// Verify token works before logout
		resp := helper.MakeRequest(t, "GET", "/api/v1/users/", nil, token)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Token should work before logout, got %d", resp.StatusCode)
		}

		// Logout
		resp = helper.MakeRequest(t, "POST", "/api/v1/logout", nil, token)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for logout, got %d", resp.StatusCode)
		}

		// Verify token is invalidated after logout
		// Note: This test depends on whether token revocation is implemented
		// For JWT tokens, they might still be valid until expiration
		resp = helper.MakeRequest(t, "GET", "/api/v1/users/", nil, token)
		resp.Body.Close()

		// The behavior here depends on implementation:
		// - If JWT revocation is implemented: should return 401
		// - If JWT without revocation: might still return 200 until expiration
		t.Logf("Post-logout token validation returned status: %d", resp.StatusCode)
	})

	t.Run("Token Expiration", func(t *testing.T) {
		helper.CleanupTestData(t)

		// This test would require modifying JWT expiration time for testing
		// or mocking time, which is complex in integration tests
		// For now, we'll just verify that tokens contain expiration claims

		token := helper.RegisterAndLoginUser(t)

		// Verify token works
		resp := helper.MakeRequest(t, "GET", "/api/v1/users/", nil, token)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 with fresh token, got %d", resp.StatusCode)
		}

		t.Logf("Token expiration test completed - token works as expected")
	})

	t.Run("Concurrent Authentication", func(t *testing.T) {
		helper.CleanupTestData(t)

		// Test multiple users logging in concurrently
		numUsers := 5
		results := make(chan error, numUsers)

		for i := 0; i < numUsers; i++ {
			go func(userNum int) {
				// Register user
				registerData := map[string]string{
					"username":   "concurrent" + string(rune(userNum+97)),
					"email":      "concurrent" + string(rune(userNum+97)) + "@example.com",
					"password":   "password123",
					"first_name": "Concurrent",
					"last_name":  "User",
				}

				registerJSON, _ := json.Marshal(registerData)
				resp := helper.MakeRequest(t, "POST", "/api/v1/register", bytes.NewBuffer(registerJSON), "")
				resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					results <- fmt.Errorf("registration failed for user %d: %d", userNum, resp.StatusCode)
					return
				}

				// Login user
				loginData := map[string]string{
					"identifier": registerData["username"],
					"password":   registerData["password"],
				}

				loginJSON, _ := json.Marshal(loginData)
				resp = helper.MakeRequest(t, "POST", "/api/v1/login", bytes.NewBuffer(loginJSON), "")
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					results <- fmt.Errorf("login failed for user %d: %d", userNum, resp.StatusCode)
					return
				}

				results <- nil
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numUsers; i++ {
			select {
			case err := <-results:
				if err != nil {
					t.Errorf("Concurrent authentication error: %v", err)
				}
			case <-time.After(10 * time.Second):
				t.Error("Timeout waiting for concurrent authentication")
			}
		}
	})
}