package auth

import (
	"testing"
	"time"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("super-secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if !CheckPassword(hash, "super-secret") {
		t.Error("expected password to match hash")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("expected wrong password not to match hash")
	}
}

func TestTokenManager_GenerateAndParse(t *testing.T) {
	tm := NewTokenManager("test-secret", time.Minute)

	token, err := tm.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	userID, err := tm.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if userID != "user-123" {
		t.Errorf("expected user-123, got %s", userID)
	}
}

func TestTokenManager_ExpiredToken(t *testing.T) {
	tm := NewTokenManager("test-secret", -time.Minute)

	token, err := tm.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if _, err := tm.ParseToken(token); err == nil {
		t.Error("expected expired token to fail parsing")
	}
}

func TestTokenManager_InvalidSecret(t *testing.T) {
	tm := NewTokenManager("secret-one", time.Minute)
	other := NewTokenManager("secret-two", time.Minute)

	token, err := tm.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if _, err := other.ParseToken(token); err == nil {
		t.Error("expected token signed with different secret to fail parsing")
	}
}
