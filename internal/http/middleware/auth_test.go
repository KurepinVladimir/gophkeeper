package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gophkeeper/internal/security"
)

func TestAuthMiddlewareOK(t *testing.T) {
	secret := "s"
	tok, err := security.IssueToken(secret, 42, time.Minute)
	require.NoError(t, err)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := UserIDFromContext(r.Context())
		require.True(t, ok)
		require.Equal(t, int64(42), uid)
		w.WriteHeader(http.StatusOK)
	})

	h := Auth(secret)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddlewareMissing(t *testing.T) {
	h := Auth("s")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}
