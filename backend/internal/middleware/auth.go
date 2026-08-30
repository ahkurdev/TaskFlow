package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"taskflow/internal/service"
)

// contextKey avoids collisions with other packages' context keys.
type contextKey string

const userIDContextKey contextKey = "userID"

// Auth returns middleware that requires a valid "Authorization: Bearer <token>"
// header, verifies it via the AuthService, and stores the user ID in the
// request context.
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				writeUnauthorized(w, "missing or malformed Authorization header")
				return
			}

			userID, err := authService.ParseToken(parts[1])
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user's ID set by Auth middleware.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
