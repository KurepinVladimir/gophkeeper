package service

import (
	"context"
	"time"

	"gophkeeper/internal/model"
	"gophkeeper/internal/repository"
	"gophkeeper/internal/security"
)

// AuthService implements user registration/login.
type AuthService struct {
	users     repository.UserRepository
	jwtSecret string
	jwtTTL    time.Duration
}

// NewAuthService creates AuthService.
func NewAuthService(users repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
		jwtTTL:    24 * time.Hour,
	}
}

// Register creates a new user.
func (a *AuthService) Register(ctx context.Context, login, password string) error {
	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	u := &model.User{Login: login, PasswordHash: hash}
	return a.users.Create(ctx, u)
}

// Login authenticates user and returns JWT.
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
