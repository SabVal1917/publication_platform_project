package utils

import "testing"

func TestTokenCanBeVerifiedByAnotherInstanceWithSharedSecret(t *testing.T) {
	const sharedSecret = "shared-secret-at-least-thirty-two-bytes"
	issuerInstance := NewTokenManager(sharedSecret)
	verifierInstance := NewTokenManager(sharedSecret)

	token, err := issuerInstance.GenerateToken(42)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	userID, err := verifierInstance.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() on another instance error = %v", err)
	}
	if userID != 42 {
		t.Fatalf("ParseToken() user ID = %d, want 42", userID)
	}
}

func TestTokenIsRejectedByInstanceWithDifferentSecret(t *testing.T) {
	issuerInstance := NewTokenManager("first-secret-at-least-thirty-two-bytes")
	otherInstance := NewTokenManager("second-secret-at-least-thirty-two-bytes")

	token, err := issuerInstance.GenerateToken(42)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := otherInstance.ParseToken(token); err == nil {
		t.Fatal("ParseToken() error = nil, want signature validation error")
	}
}
