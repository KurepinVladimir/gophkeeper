package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type LocalRecord struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Meta      string    `json:"meta"`
	CipherB64 string    `json:"cipher_b64"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
}

func recordsPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "records.json"), nil
}

func LoadRecords() ([]LocalRecord, error) {
	p, err := recordsPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []LocalRecord{}, nil
		}
		return nil, err
	}
	var rr []LocalRecord
	return rr, json.Unmarshal(b, &rr)
}

func SaveRecords(rr []LocalRecord) error {
	p, err := recordsPath()
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o700)
	b, _ := json.MarshalIndent(rr, "", "  ")
	return os.WriteFile(p, b, 0o600)
}
