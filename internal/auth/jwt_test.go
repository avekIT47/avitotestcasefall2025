package auth

import (
	"testing"
	"time"
)

func TestContextKeys(t *testing.T) {
	tests := []struct {
		name string
		key  contextKey
		want string
	}{
		{"user_id key", contextKeyUserID, "user_id"},
		{"email key", contextKeyEmail, "email"},
		{"role key", contextKeyRole, "role"},
		{"team_id key", contextKeyTeamID, "team_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.key) != tt.want {
				t.Errorf("expected %s, got %s", tt.want, string(tt.key))
			}
		})
	}
}

func TestErrors(t *testing.T) {
	if ErrInvalidToken == nil {
		t.Error("ErrInvalidToken should not be nil")
	}

	if ErrExpiredToken == nil {
		t.Error("ErrExpiredToken should not be nil")
	}

	if ErrMissingToken == nil {
		t.Error("ErrMissingToken should not be nil")
	}

	if ErrInvalidToken.Error() != "invalid token" {
		t.Errorf("unexpected error message: %s", ErrInvalidToken.Error())
	}

	if ErrExpiredToken.Error() != "token expired" {
		t.Errorf("unexpected error message: %s", ErrExpiredToken.Error())
	}

	if ErrMissingToken.Error() != "missing authorization header" {
		t.Errorf("unexpected error message: %s", ErrMissingToken.Error())
	}
}

func TestClaims_Structure(t *testing.T) {
	claims := &Claims{
		UserID: 123,
		Email:  "test@example.com",
		Role:   "admin",
		TeamID: 1,
	}

	if claims.UserID != 123 {
		t.Errorf("expected UserID 123, got %d", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", claims.Email)
	}

	if claims.Role != "admin" {
		t.Errorf("expected role 'admin', got '%s'", claims.Role)
	}

	if claims.TeamID != 1 {
		t.Errorf("expected TeamID 1, got %d", claims.TeamID)
	}
}

func TestJWTAuth_Structure(t *testing.T) {
	// Test that JWTAuth can be created
	auth := &JWTAuth{
		secretKey:       []byte("test-secret"),
		tokenExpiration: 1 * time.Hour,
		logger:          nil, // Would be a real logger in production
	}

	if auth == nil {
		t.Error("expected non-nil JWTAuth")
	}

	if string(auth.secretKey) != "test-secret" {
		t.Error("expected secret key to match")
	}

	if auth.tokenExpiration != 1*time.Hour {
		t.Error("expected 1 hour expiration")
	}
}

