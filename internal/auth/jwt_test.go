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
	// Проверяем сообщения ошибок (проверки на nil не нужны для var с errors.New)
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

	// Проверка на nil не нужна для литерала структуры
	expectedKey := []byte("test-secret")
	if len(auth.secretKey) != len(expectedKey) {
		t.Error("expected secret key to match")
	}
	for i, b := range expectedKey {
		if auth.secretKey[i] != b {
			t.Error("expected secret key to match")
			break
		}
	}

	if auth.tokenExpiration != 1*time.Hour {
		t.Error("expected 1 hour expiration")
	}
}
