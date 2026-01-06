package envelope

import (
	"bytes"
	"encoding/base64"
	"testing"
)

// Тест 1: encrypt → decrypt возвращает исходные данные
func TestEncryptDecryptPayload(t *testing.T) {
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	mk := base64.StdEncoding.EncodeToString(masterKey)

	enc, err := New(mk)
	if err != nil {
		t.Fatalf("failed to create encrypter: %v", err)
	}

	plain := []byte("super secret data")

	encData, encKey, err := enc.EncryptPayload(plain)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	result, err := enc.DecryptPayload(encData, encKey)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if !bytes.Equal(result, plain) {
		t.Fatalf("expected %q, got %q", plain, result)
	}
}

// Тест 2: неверный master-key → ошибка
func TestDecryptWithWrongMasterKey(t *testing.T) {
	key1 := base64.StdEncoding.EncodeToString(make([]byte, 32))
	key2 := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))

	enc1, _ := New(key1)
	enc2, _ := New(key2)

	data := []byte("secret")

	encData, encKey, _ := enc1.EncryptPayload(data)

	_, err := enc2.DecryptPayload(encData, encKey)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
