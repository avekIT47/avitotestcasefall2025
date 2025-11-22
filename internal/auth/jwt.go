package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/user/pr-reviewer/internal/logger"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
	ErrMissingToken = errors.New("missing authorization header")
)

// Claims кастомные JWT claims
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	TeamID int64  `json:"team_id,omitempty"`
	jwt.RegisteredClaims
}

// JWTAuth управляет JWT аутентификацией
type JWTAuth struct {
	secretKey       []byte
	tokenExpiration time.Duration
	logger          *logger.Logger
}

// NewJWTAuth создает новый JWT auth
func NewJWTAuth(secretKey string, tokenExpiration time.Duration, log *logger.Logger) *JWTAuth {
	return &JWTAuth{
		secretKey:       []byte(secretKey),
		tokenExpiration: tokenExpiration,
		logger:          log,
	}
}

// GenerateToken генерирует новый JWT token
func (a *JWTAuth) GenerateToken(userID int64, email, role string, teamID int64) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		TeamID: teamID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(a.tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "pr-reviewer",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken валидирует JWT token
func (a *JWTAuth) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// Middleware JWT authentication middleware
func (a *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.logger.Warnw("Missing authorization header", "path", r.URL.Path)
			a.sendError(w, ErrMissingToken, http.StatusUnauthorized)
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			a.logger.Warnw("Invalid authorization header format", "path", r.URL.Path)
			a.sendError(w, ErrInvalidToken, http.StatusUnauthorized)
			return
		}

		// Валидируем токен
		claims, err := a.ValidateToken(parts[1])
		if err != nil {
			a.logger.Warnw("Token validation failed", "error", err, "path", r.URL.Path)
			status := http.StatusUnauthorized
			if errors.Is(err, ErrExpiredToken) {
				status = http.StatusUnauthorized
			}
			a.sendError(w, err, status)
			return
		}

		// Добавляем claims в контекст
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "team_id", claims.TeamID)

		a.logger.Debugw("Request authenticated", "user_id", claims.UserID, "path", r.URL.Path)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalMiddleware опциональная аутентификация (не требует токен)
func (a *JWTAuth) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				if claims, err := a.ValidateToken(parts[1]); err == nil {
					ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
					ctx = context.WithValue(ctx, "email", claims.Email)
					ctx = context.WithValue(ctx, "role", claims.Role)
					ctx = context.WithValue(ctx, "team_id", claims.TeamID)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole middleware для проверки роли
func (a *JWTAuth) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value("role").(string)
			if !ok {
				a.sendError(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// Проверяем наличие роли
			hasRole := false
			for _, r := range roles {
				if role == r {
					hasRole = true
					break
				}
			}

			if !hasRole {
				a.logger.Warnw("Insufficient permissions",
					"user_role", role,
					"required_roles", roles,
					"path", r.URL.Path,
				)
				a.sendError(w, errors.New("forbidden"), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID извлекает user ID из контекста
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value("user_id").(int64)
	return userID, ok
}

// GetUserEmail извлекает email из контекста
func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value("email").(string)
	return email, ok
}

// GetUserRole извлекает роль из контекста
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value("role").(string)
	return role, ok
}

// GetTeamID извлекает team ID из контекста
func GetTeamID(ctx context.Context) (int64, bool) {
	teamID, ok := ctx.Value("team_id").(int64)
	return teamID, ok
}

// sendError отправляет ошибку в формате JSON
func (a *JWTAuth) sendError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
}
