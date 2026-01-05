package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/scrypt"
)

var (
	// ErrBadCipher indicates malformed ciphertext.
	ErrBadCipher = errors.New("bad cipher")
)

// DeriveKey derives a 32-byte key from master password + salt using scrypt.
func DeriveKey(master string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(master), salt, 1<<15, 8, 1, 32)
}

// Encrypt encrypts plain bytes using AES-GCM and returns base64 ciphertext (nonce|cipher).
func Encrypt(key, plain []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, plain, nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt decrypts base64 ciphertext (nonce|cipher) using AES-GCM.
func Decrypt(key []byte, cipherB64 string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return nil, ErrBadCipher
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return nil, ErrBadCipher
	}
	nonce, ct := raw[:ns], raw[ns:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrBadCipher
	}
	return plain, nil
}
