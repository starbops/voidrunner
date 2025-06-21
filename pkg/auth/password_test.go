package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	
	if hash == "" {
		t.Error("HashPassword() should return non-empty hash")
	}
	
	if hash == password {
		t.Error("HashPassword() should not return the original password")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "testpassword123"
	
	hash1, err1 := HashPassword(password)
	if err1 != nil {
		t.Fatalf("HashPassword() error = %v", err1)
	}
	
	hash2, err2 := HashPassword(password)
	if err2 != nil {
		t.Fatalf("HashPassword() error = %v", err2)
	}
	
	if hash1 == hash2 {
		t.Error("HashPassword() should generate different hashes for same password due to salt")
	}
}

func TestCheckPasswordHash_Valid(t *testing.T) {
	password := "testpassword123"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	
	if !CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash() should return true for correct password")
	}
}

func TestCheckPasswordHash_Invalid(t *testing.T) {
	password := "testpassword123"
	wrongPassword := "wrongpassword"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	
	if CheckPasswordHash(wrongPassword, hash) {
		t.Error("CheckPasswordHash() should return false for incorrect password")
	}
}

func TestCheckPasswordHash_EmptyPassword(t *testing.T) {
	password := "testpassword123"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	
	if CheckPasswordHash("", hash) {
		t.Error("CheckPasswordHash() should return false for empty password")
	}
}

func TestCheckPasswordHash_EmptyHash(t *testing.T) {
	password := "testpassword123"
	
	if CheckPasswordHash(password, "") {
		t.Error("CheckPasswordHash() should return false for empty hash")
	}
}