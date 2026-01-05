package store

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
)

type KeyInfo struct {
	SaltB64 string `json:"salt_b64"`
}

func keyPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "key.json"), nil
}

// LoadOrCreateKeyInfo returns stored salt (base64) or creates a new one.
// Salt is used for deriving encryption key from master password (scrypt).
func LoadOrCreateKeyInfo() (KeyInfo, error) {
	p, err := keyPath()
	if err != nil {
		return KeyInfo{}, err
	}
	if b, err := os.ReadFile(p); err == nil {
		var ki KeyInfo
		return ki, json.Unmarshal(b, &ki)
	}

	salt := make([]byte, 16)
	_, _ = rand.Read(salt)

	ki := KeyInfo{SaltB64: base64.StdEncoding.EncodeToString(salt)}
	_ = os.MkdirAll(filepath.Dir(p), 0o700)

	out, _ := json.MarshalIndent(ki, "", "  ")
	if err := os.WriteFile(p, out, 0o600); err != nil {
		return KeyInfo{}, err
	}
	return ki, nil
}
