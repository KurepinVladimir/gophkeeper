package store

import (
	"os"
	"path/filepath"
)

// BaseDir returns "~/.gophkeeper".
func BaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gophkeeper"), nil
}
