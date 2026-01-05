package model

import "time"

// User is an application user.
type User struct {
	ID           int64     `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// SecretType describes a kind of stored information.
// Supported: login, text, card, bin.
type SecretType string

const (
	SecretLogin SecretType = "login"
	SecretText  SecretType = "text"
	SecretCard  SecretType = "card"
	SecretBin   SecretType = "bin"
)

// Secret is a server-side record. EncryptedData contains ciphertext produced by the client.
type Secret struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"-"`
	Type          SecretType `json:"type"`
	Title         string     `json:"title"`
	Meta          string     `json:"meta"`
	EncryptedData []byte     `json:"encrypted_data"`
	Version       int64      `json:"version"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Deleted       bool       `json:"deleted"`
}
