package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/pkg/utils"
)

type contextKey uint64

const userIDKey contextKey = iota

func AuthMiddleware(tokens *utils.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Неверный формат заголовка", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			userID, err := tokens.ParseToken(tokenString)
			if err != nil {
				http.Error(w, "Невалидный токен", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok && userID > 0
}
