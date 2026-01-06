package service

import (
	"context"
	"errors"
	"time"

	"gophkeeper/internal/crypto/envelope"
	"gophkeeper/internal/model"
	"gophkeeper/internal/repository"
)

// SecretsService implements operations on encrypted secrets.
type SecretsService struct {
	repo      repository.SecretRepository
	encrypter *envelope.Encrypter
}

// NewSecretsService creates SecretsService.
func NewSecretsService(
	repo repository.SecretRepository,
	encrypter *envelope.Encrypter,
) *SecretsService {
	return &SecretsService{
		repo:      repo,
		encrypter: encrypter,
	}
}

// Upsert stores or updates secret.
func (s *SecretsService) Upsert(ctx context.Context, sec *model.Secret) (*model.Secret, error) {
	// 1. updated_at
	if sec.UpdatedAt.IsZero() {
		sec.UpdatedAt = time.Now().UTC()
	}

	// 2. versioning
	if sec.Version == 0 {
		sec.Version = 1
	}

	// 3. валидация входных данных
	if len(sec.EncryptedData) == 0 {
		return nil, errors.New("empty encrypted data")
	}

	const maxEncryptedSize = 1 << 20 // 1MB

	if len(sec.EncryptedData) > maxEncryptedSize {
		return nil, errors.New("encrypted data exceeds 1MB limit")
	}

	// 4. envelope encryption
	encData, encKey, err := s.encrypter.EncryptPayload(sec.EncryptedData)
	if err != nil {
		return nil, err
	}

	// 5. заменяем payload
	sec.EncryptedData = encData
	sec.EncryptedDataKey = encKey

	// 6. сохраняем
	return s.repo.Upsert(ctx, sec)
}

// List returns all secrets for user.
func (s *SecretsService) List(ctx context.Context, userID int64) ([]model.Secret, error) {
	return s.repo.List(ctx, userID)
}

// Get returns secret by id.
// Get returns secret by id.
func (s *SecretsService) Get(ctx context.Context, id, userID int64) (*model.Secret, error) {
	sec, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// legacy-запись (до envelope encryption)
	if len(sec.EncryptedDataKey) == 0 {
		return nil, errors.New("legacy secret without encrypted data key")
	}

	// снимаем envelope encryption
	plainEncryptedData, err := s.encrypter.DecryptPayload(
		sec.EncryptedData,
		sec.EncryptedDataKey,
	)
	if err != nil {
		return nil, err
	}

	// возвращаем клиенту данные БЕЗ envelope,
	// но всё ещё зашифрованные client-side
	sec.EncryptedData = plainEncryptedData
	sec.EncryptedDataKey = nil // никогда не отдаём

	return sec, nil
}

// Delete deletes secret.
func (s *SecretsService) Delete(ctx context.Context, id, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}
