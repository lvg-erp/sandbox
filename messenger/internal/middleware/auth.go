package middleware

import (
	"context"
	"net/http"
	"strings"

	"messanger/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

func JWTAuth(jwtConfig *auth.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем токен из заголовка Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Проверяем формат "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Валидируем токен
			claims, err := jwtConfig.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Сохраняем данные пользователя в контексте
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext извлекает данные пользователя из контекста
func GetUserFromContext(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(UserContextKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
