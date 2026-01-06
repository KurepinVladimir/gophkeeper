package envelope

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var ErrDecrypt = errors.New("decrypt failed")

type Encrypter struct {
	masterKey []byte
}

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

// // Шифрование перед сохранением
func (e *Encrypter) EncryptPayload(plain []byte) (encData, encDataKey []byte, err error) {
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

// Расшифровка при чтении
func (e *Encrypter) DecryptPayload(encData, encDataKey []byte) ([]byte, error) {
	// 1. расшифровываем data-key
	dataKey, err := decrypt(e.masterKey, encDataKey)
	if err != nil {
		return nil, err
	}

	// 2. расшифровываем данные
	return decrypt(dataKey, encData)
}
