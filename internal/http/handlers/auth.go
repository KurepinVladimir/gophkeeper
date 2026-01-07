// Package handlers contains HTTP handlers for the GophKeeper server.
// Handlers are responsible for request parsing, response formatting,
// and delegating business logic to the service layer.
package handlers

import (
	"encoding/json"
	"net/http"

	"gophkeeper/internal/service"
)

// AuthHandler handles HTTP requests related to user authentication,
// including registration and login operations.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler creates a new AuthHandler using the provided AuthService.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register handles user registration requests.
// It parses credentials from the request body and delegates
// user creation to the authentication service.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "login/password required", http.StatusBadRequest)
		return
	}
	if err := h.auth.Register(r.Context(), req.Login, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type loginResp struct {
	Token string `json:"token"`
}

// Login handles user authentication requests.
// On successful authentication, it returns a JWT token
// that can be used for authorized API calls.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	tok, _, err := h.auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResp{Token: tok})
}
