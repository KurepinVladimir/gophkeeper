package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"gophkeeper/internal/crypto/envelope"
	"gophkeeper/internal/model"
	"testing"
)

// Тест 1: Upsert с пустыми данными → ошибка
type mockSecretRepo struct {
	upsertFn func(ctx context.Context, sec *model.Secret) (*model.Secret, error)
	getFn    func(ctx context.Context, id, userID int64) (*model.Secret, error)
}

func (m *mockSecretRepo) Upsert(ctx context.Context, sec *model.Secret) (*model.Secret, error) {
	return m.upsertFn(ctx, sec)
}

func (m *mockSecretRepo) Get(ctx context.Context, id, userID int64) (*model.Secret, error) {
	return m.getFn(ctx, id, userID)
}

// заглушки
func (m *mockSecretRepo) List(ctx context.Context, userID int64) ([]model.Secret, error) {
	return nil, nil
}

func (m *mockSecretRepo) Delete(ctx context.Context, id, userID int64) error {
	return nil
}

// Тест 2: Upsert шифрует данные
func TestSecretsService_Upsert_EncryptsData(t *testing.T) {
	masterKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	enc, _ := envelope.New(masterKey)

	repo := &mockSecretRepo{
		upsertFn: func(ctx context.Context, sec *model.Secret) (*model.Secret, error) {
			// проверяем, что данные УЖЕ зашифрованы envelope
			if len(sec.EncryptedDataKey) == 0 {
				t.Fatal("EncryptedDataKey is empty")
			}
			if bytes.Equal(sec.EncryptedData, []byte("data")) {
				t.Fatal("data was not encrypted")
			}
			return sec, nil
		},
	}

	svc := NewSecretsService(repo, enc)

	sec := &model.Secret{
		EncryptedData: []byte("data"),
	}

	_, err := svc.Upsert(context.Background(), sec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Тест 3: превышение лимита → ошибка
func TestSecretsService_Upsert_TooLarge(t *testing.T) {
	masterKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	enc, _ := envelope.New(masterKey)

	repo := &mockSecretRepo{}
	svc := NewSecretsService(repo, enc)

	sec := &model.Secret{
		EncryptedData: make([]byte, (1<<20)+1),
	}

	_, err := svc.Upsert(context.Background(), sec)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Тест 4: Get снимает envelope encryption
func TestSecretsService_Get_DecryptsEnvelope(t *testing.T) {
	masterKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	enc, _ := envelope.New(masterKey)

	inner := []byte("client-encrypted")

	encData, encKey, _ := enc.EncryptPayload(inner)

	repo := &mockSecretRepo{
		getFn: func(ctx context.Context, id, userID int64) (*model.Secret, error) {
			return &model.Secret{
				EncryptedData:    encData,
				EncryptedDataKey: encKey,
			}, nil
		},
	}

	svc := NewSecretsService(repo, enc)

	sec, err := svc.Get(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(sec.EncryptedData, inner) {
		t.Fatal("envelope encryption was not removed")
	}
}
