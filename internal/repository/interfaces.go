package repository

import (
	"context"

	"gophkeeper/internal/model"
)

// UserRepository provides user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

// SecretRepository provides secret persistence.
type SecretRepository interface {
	Upsert(ctx context.Context, s *model.Secret) (*model.Secret, error)
	List(ctx context.Context, userID int64) ([]model.Secret, error)
	Get(ctx context.Context, id int64, userID int64) (*model.Secret, error)
	Delete(ctx context.Context, id int64, userID int64) error
}
