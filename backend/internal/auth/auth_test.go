package auth

import (
	"testing"
)

func TestArgon2PasswordHashing(t *testing.T) {
	password := "SecureStudentPass2026!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Fatalf("Password hash should not be empty")
	}

	valid, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("Failed to check password hash: %v", err)
	}
	if !valid {
		t.Errorf("Password validation should succeed for correct password")
	}

	invalidCheck, _ := CheckPasswordHash("WrongPassword!", hash)
	if invalidCheck {
		t.Errorf("Password validation should fail for incorrect password")
	}
}

func TestJWTTokens(t *testing.T) {
	secret := "test_jwt_secret_key_12345"
	userID := "user-uuid-12345"
	email := "student@example.com"
	role := "student"

	tokens, err := GenerateTokenPair(userID, email, role, secret, 15, 7)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("Tokens should not be empty")
	}

	claims, err := ValidateAccessToken(tokens.AccessToken, secret)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}

	if claims.UserID != userID || claims.Email != email || claims.Role != role {
		t.Errorf("Claims mismatch: got %+v", claims)
	}

	// Validate with wrong secret fails
	_, err = ValidateAccessToken(tokens.AccessToken, "wrong_secret")
	if err == nil {
		t.Errorf("ValidateAccessToken should fail with wrong secret")
	}
}
