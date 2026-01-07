// Package envelope implements server-side envelope encryption.
// It uses a master key to encrypt per-record data encryption keys,
// which are then used to encrypt secret payloads before storage.
package envelope

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// ErrDecrypt is returned when encrypted data cannot be decrypted
// using the provided key or when the payload is malformed.
var ErrDecrypt = errors.New("decrypt failed")

// Encrypter performs envelope encryption using a master key.
// It encrypts secret payloads with a randomly generated data key
// and protects that data key by encrypting it with the master key.
type Encrypter struct {
	masterKey []byte
}

// New creates a new Encrypter instance using the provided master key.
// The master key must be base64-encoded and is decoded during initialization.
func New(masterKeyB64 string) (*Encrypter, error) {
	key, err := base64.StdEncoding.DecodeString(masterKeyB64)
	if err != nil {
		return nil, err
	}
	return &Encrypter{masterKey: key}, nil
}

//Вспомогательные функции AES-GCM

func encrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func decrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, ErrDecrypt
	}

	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Envelope-логика

// EncryptPayload encrypts the provided payload using envelope encryption.
// A random data key is generated for encrypting the payload, and this
// data key is then encrypted with the master key.
func (e *Encrypter) EncryptPayload(plain []byte) (encData, encDataKey []byte, err error) { // Шифрование перед сохранением
	// 1. генерируем data-key
	dataKey := make([]byte, 32)
	if _, err = rand.Read(dataKey); err != nil {
		return nil, nil, err
	}

	// 2. шифруем данные data-key
	encData, err = encrypt(dataKey, plain)
	if err != nil {
		return nil, nil, err
	}

	// 3. шифруем data-key мастер-ключом
	encDataKey, err = encrypt(e.masterKey, dataKey)
	if err != nil {
		return nil, nil, err
	}

	return encData, encDataKey, nil
}

// DecryptPayload decrypts a payload encrypted with envelope encryption.
// It first decrypts the data encryption key using the master key and
// then uses that key to decrypt the payload.
func (e *Encrypter) DecryptPayload(encData, encDataKey []byte) ([]byte, error) { // Расшифровка при чтении
	// 1. расшифровываем data-key
	dataKey, err := decrypt(e.masterKey, encDataKey)
	if err != nil {
		return nil, err
	}

	// 2. расшифровываем данные
	return decrypt(dataKey, encData)
}
