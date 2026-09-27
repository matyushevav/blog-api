package middleware

import (
	"advanced-blog-management-system/pkg/auth"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// contextKey - кастомный тип, чтобы избежать коллизии
type contextKey string

const (
	// Ключи, под которыми AuthMiddleware кладет данные пользователя в контекст запроса
	UserIDKey    contextKey = "userID"
	UserEmailKey contextKey = "userEmail"
	UserNameKey  contextKey = "username"
)

// AuthMiddleware проверяет JWT и кладет данные пользователя в контекст.
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware создает middleware авторизации.
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth пропускает запрос дальше только с валидным JWT.
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			respondWithError(w, "missing or malformed authorization header", http.StatusUnauthorized)

			return
		}

		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			message := "invalid token"
			if errors.Is(err, auth.ErrExpiredToken) {
				message = "token expired"
			}

			respondWithError(w, message, http.StatusUnauthorized)

			return
		}

		// Контекст неизменяемый: WithValue возвращает новый, дополненный
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserNameKey, claims.Username)

		next(w, r.WithContext(ctx))
	}
}

// OptionalAuth проверяет токен. Добавляет данные пользователя в контекст в случае валидного токена.
func (m *AuthMiddleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			next(w, r)

			return
		}

		claims, err := m.jwtManager.ValidateToken(token)
		if err == nil {
			// Контекст неизменяемый: WithValue возвращает новый, дополненный
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserNameKey, claims.Username)

			next(w, r.WithContext(ctx))

			return
		}

		next(w, r)
	}
}

// GetUserIDFromContext возвращает ID пользователя, положенный в контекст
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)

	return userID, ok
}

// GetUserEmailFromContext возвращает email пользователя, положенный в контекст
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	userEmail, ok := ctx.Value(UserEmailKey).(string)

	return userEmail, ok
}

// GetUsernameFromContext возвращает username пользователя, положенный в контекст
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	userName, ok := ctx.Value(UserNameKey).(string)

	return userName, ok
}

// extractToken достает JWT из заголовка "Authorization: Bearer <token>".
func extractToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, message string, code int) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, ErrorResponse{Error: message})
}
