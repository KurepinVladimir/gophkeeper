package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Session struct {
	Server string `json:"server"`
	Token  string `json:"token"`
}

func sessionPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "session.json"), nil
}

func SaveSession(s Session) error {
	p, err := sessionPath()
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o700)
	b, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(p, b, 0o600)
}

func LoadSession() (Session, error) {
	p, err := sessionPath()
	if err != nil {
		return Session{}, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return Session{}, err
	}
	var s Session
	return s, json.Unmarshal(b, &s)
}
