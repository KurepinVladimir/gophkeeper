// Package middleware contains HTTP middleware used by the server.
// It provides request-level functionality such as authentication,
// authorization, and propagation of user identity through context.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"gophkeeper/internal/security"
)

type ctxKey string

const userIDKey ctxKey = "uid"

// UserIDFromContext extracts the authenticated user identifier
// from the provided context. It returns the user ID and a boolean
// indicating whether the value was present.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(userIDKey)
	id, ok := v.(int64)
	return id, ok
}

// Auth is an HTTP middleware that validates a JWT token from the
// Authorization header. On successful validation, it extracts
// the user identifier and injects it into the request context.
//
// Requests without a valid token are rejected with an HTTP 401 status.
func Auth(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				http.Error(w, "missing authorization", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid authorization", http.StatusUnauthorized)
				return
			}
			claims, err := security.ParseToken(jwtSecret, parts[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
