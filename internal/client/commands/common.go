package commands

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"gophkeeper/internal/client/api"
	"gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/store"
)

func loadKey(master string) ([]byte, error) {
	ki, err := store.LoadOrCreateKeyInfo()
	if err != nil {
		return nil, err
	}
	salt, err := base64.StdEncoding.DecodeString(ki.SaltB64)
	if err != nil {
		return nil, err
	}
	return crypto.DeriveKey(master, salt)
}

func mustSession() (store.Session, error) {
	s, err := store.LoadSession()
	if err != nil {
		return store.Session{}, errors.New("not logged in: run login first")
	}
	if s.Token == "" || s.Server == "" {
		return store.Session{}, errors.New("not logged in: run login first")
	}
	return s, nil
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func decodeData[T any](b []byte, out *T) error {
	return json.Unmarshal(b, out)
}

func encodeData(v any) ([]byte, error) {
	return json.Marshal(v)
}

func apiClient(server string) *api.Client {
	return api.New(server)
}
