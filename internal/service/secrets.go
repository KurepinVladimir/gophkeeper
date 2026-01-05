package service

import (
	"context"
	"time"

	"gophkeeper/internal/model"
	"gophkeeper/internal/repository"
)

// SecretsService implements operations on encrypted secrets.
type SecretsService struct {
	repo repository.SecretRepository
}

// NewSecretsService creates SecretsService.
func NewSecretsService(repo repository.SecretRepository) *SecretsService {
	return &SecretsService{repo: repo}
}

// Upsert stores or updates secret.
func (s *SecretsService) Upsert(ctx context.Context, sec *model.Secret) (*model.Secret, error) {
	if sec.UpdatedAt.IsZero() {
		sec.UpdatedAt = time.Now().UTC()
	}
	if sec.Version == 0 {
		sec.Version = 1
	}
	return s.repo.Upsert(ctx, sec)
}

// List returns all secrets for user.
func (s *SecretsService) List(ctx context.Context, userID int64) ([]model.Secret, error) {
	return s.repo.List(ctx, userID)
}

// Get returns secret by id.
func (s *SecretsService) Get(ctx context.Context, id, userID int64) (*model.Secret, error) {
	return s.repo.Get(ctx, id, userID)
}

// Delete deletes secret.
func (s *SecretsService) Delete(ctx context.Context, id, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}
