// Package service contains application business logic.
// It implements core use cases such as user authentication
// and secure management of encrypted secrets, independent
// of transport and storage implementations.
package service

import (
	"context"
	"time"

	"gophkeeper/internal/model"
	"gophkeeper/internal/repository"
	"gophkeeper/internal/security"
)

// AuthService provides user authentication and registration logic.
// It is responsible for creating users, verifying credentials,
// and issuing JWT tokens for authenticated sessions.
type AuthService struct {
	users     repository.UserRepository
	jwtSecret string
	jwtTTL    time.Duration
}

// NewAuthService creates a new AuthService.
// It accepts a user repository for persistence and a JWT secret
// used to sign authentication tokens.
func NewAuthService(users repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
		jwtTTL:    24 * time.Hour,
	}
}

// Register creates a new user with the provided login and password.
// The password is securely hashed before being stored.
func (a *AuthService) Register(ctx context.Context, login, password string) error {
	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	u := &model.User{Login: login, PasswordHash: hash}
	return a.users.Create(ctx, u)
}

// Login authenticates a user by login and password.
// On success, it returns a signed JWT token that can be used
// for authorized requests.
func (a *AuthService) Login(ctx context.Context, login, password string) (string, int64, error) {
	u, err := a.users.GetByLogin(ctx, login)
	if err != nil {
		return "", 0, err
	}
	if err := security.ComparePassword(u.PasswordHash, password); err != nil {
		return "", 0, err
	}
	tok, err := security.IssueToken(a.jwtSecret, u.ID, a.jwtTTL)
	if err != nil {
		return "", 0, err
	}
	return tok, u.ID, nil
}
