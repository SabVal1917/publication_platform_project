package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/pkg/utils"
)

func AuthMiddleware(next http.Handler) http.Handler {
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

		userID, err := utils.ParseToken(tokenString)
		if err != nil {
			http.Error(w, "Невалидный токен", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}