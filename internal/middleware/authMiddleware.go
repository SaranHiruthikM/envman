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

// MaxRequestBody caps every JSON body read by the server. Realistic encrypted
// .env payloads are well under 32 KB; 256 KB is a generous ceiling that still
// kills slowloris / huge-body DoS attempts.
const MaxRequestBody = 256 * 1024

// LimitBody attaches an http.MaxBytesReader to every request.
func LimitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBody)
		next.ServeHTTP(w, r)
	})
}

func RequireAuth(manager *token.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				response.Error(w, http.StatusUnauthorized, "Authorization header missing")
				return
			}

			// Strict Bearer parse: exactly two whitespace-separated parts.
			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				response.Error(w, http.StatusUnauthorized, "Authorization must be 'Bearer <token>'")
				return
			}

			userID, err := manager.Verify(parts[1])
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(userIDKey).(string)
	return userID, ok
}
