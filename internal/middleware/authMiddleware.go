package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/SaranHiruthikM/envman/internal/token"
)

type contextKey string

const userIDKey contextKey = "userID"

func RequireAuth(manager *token.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := r.Header.Get("Authorization")
			if header == "" {
				response.Error(w, http.StatusUnauthorized, "Authorization Header Missing")
				return
			}

			tokenString := strings.TrimPrefix(header, "Bearer ")

			userId, err := manager.Verify(tokenString)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Something went wrong")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userId)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(userIDKey).(string)
	return userID, ok
}
