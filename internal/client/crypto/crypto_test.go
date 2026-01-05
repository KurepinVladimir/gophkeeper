package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	salt := []byte("1234567890abcdef")
	key, err := DeriveKey("master", salt)
	require.NoError(t, err)
	require.Len(t, key, 32)

	plain := []byte("hello")
	ct, err := Encrypt(key, plain)
	require.NoError(t, err)

	_, err = base64.StdEncoding.DecodeString(ct)
	require.NoError(t, err)

	got, err := Decrypt(key, ct)
	require.NoError(t, err)
	require.True(t, bytes.Equal(plain, got))
}

func TestDecryptBadCipher(t *testing.T) {
	key := make([]byte, 32)
	_, err := Decrypt(key, "not-base64!!!")
	require.ErrorIs(t, err, ErrBadCipher)
}
