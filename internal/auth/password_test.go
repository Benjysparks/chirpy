package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTCreationAndValidation(t *testing.T) {
	// Create a test user ID
	userID := uuid.New()
	
	// Define a secret for testing
	tokenSecret := "your-test-secret"
	
	// Test token creation
	token, err := MakeJWT(userID, tokenSecret, time.Hour*24)
	if err != nil {
		t.Fatalf("Error creating JWT: %v", err)
	}
	
	// Test token validation
	extractedID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Error validating JWT: %v", err)
	}
	
	// Check if the extracted ID matches the original
	if extractedID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, extractedID)
	}
}

func TestExpiredToken(t *testing.T) {
	// Create a test user ID
	userID := uuid.New()
	
	// Define a secret for testing
	tokenSecret := "your-test-secret"
	
	// Create a token that expires immediately
	token, err := MakeJWT(userID, tokenSecret, -time.Hour)
	if err != nil {
		t.Fatalf("Error creating JWT: %v", err)
	}
	
	// Try to validate the expired token
	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Error("Expected error for expired token, but got nil")
	}
}

