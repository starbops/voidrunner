package auth

import (
	"testing"
	"time"
)

func TestTokenManager_GenerateToken(t *testing.T) {
	secret := "test-secret"
	expiry := time.Hour
	tm := NewTokenManager(secret, expiry)
	
	userID := 1
	username := "testuser"
	email := "test@example.com"
	
	token, err := tm.GenerateToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	
	if token == "" {
		t.Error("GenerateToken() should return non-empty token")
	}
}

func TestTokenManager_ValidateToken_Valid(t *testing.T) {
	secret := "test-secret"
	expiry := time.Hour
	tm := NewTokenManager(secret, expiry)
	
	userID := 1
	username := "testuser"
	email := "test@example.com"
	
	token, err := tm.GenerateToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	
	claims, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	
	if claims.UserID != userID {
		t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Username != username {
		t.Errorf("ValidateToken() Username = %v, want %v", claims.Username, username)
	}
	if claims.Email != email {
		t.Errorf("ValidateToken() Email = %v, want %v", claims.Email, email)
	}
}

func TestTokenManager_ValidateToken_Invalid(t *testing.T) {
	secret := "test-secret"
	expiry := time.Hour
	tm := NewTokenManager(secret, expiry)
	
	invalidToken := "invalid.token.here"
	
	_, err := tm.ValidateToken(invalidToken)
	if err == nil {
		t.Error("ValidateToken() should return error for invalid token")
	}
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestTokenManager_ValidateToken_Expired(t *testing.T) {
	secret := "test-secret"
	expiry := time.Millisecond
	tm := NewTokenManager(secret, expiry)
	
	userID := 1
	username := "testuser"
	email := "test@example.com"
	
	token, err := tm.GenerateToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	
	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)
	
	_, err = tm.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should return error for expired token")
	}
	if err != ErrExpiredToken {
		t.Errorf("ValidateToken() error = %v, want %v", err, ErrExpiredToken)
	}
}

func TestTokenManager_RevokeToken(t *testing.T) {
	secret := "test-secret"
	expiry := time.Hour
	tm := NewTokenManager(secret, expiry)
	
	userID := 1
	username := "testuser"
	email := "test@example.com"
	
	token, err := tm.GenerateToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	
	// Token should be valid before revocation
	_, err = tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	
	// Revoke the token
	err = tm.RevokeToken(token)
	if err != nil {
		t.Fatalf("RevokeToken() error = %v", err)
	}
	
	// Token should be invalid after revocation
	_, err = tm.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should return error for revoked token")
	}
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestTokenManager_DifferentSecrets(t *testing.T) {
	secret1 := "secret1"
	secret2 := "secret2"
	expiry := time.Hour
	
	tm1 := NewTokenManager(secret1, expiry)
	tm2 := NewTokenManager(secret2, expiry)
	
	userID := 1
	username := "testuser"
	email := "test@example.com"
	
	token, err := tm1.GenerateToken(userID, username, email)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	
	// Token from tm1 should not be valid for tm2
	_, err = tm2.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should return error for token with different secret")
	}
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken() error = %v, want %v", err, ErrInvalidToken)
	}
}