package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ErrMissingKey is returned by LoadKey when the key file is absent.
var ErrMissingKey = errors.New("crypto: app.key missing")

// KeyPath returns the on-disk path of app.key beside the given executable
// directory. Pass the directory returned by filepath.Dir(os.Executable()).
func KeyPath(exeDir string) string {
	return filepath.Join(exeDir, "app.key")
}

// GenerateKey creates a new 32-byte random key and writes it to path with
// 0600 permissions. Overwrites any existing file.
func GenerateKey(path string) ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("crypto: read random: %w", err)
	}
	if err := os.WriteFile(path, key, 0o600); err != nil {
		return nil, fmt.Errorf("crypto: write key: %w", err)
	}
	return key, nil
}

// LoadKey reads the key from path. Returns ErrMissingKey if the file does
// not exist.
func LoadKey(path string) ([]byte, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrMissingKey
		}
		return nil, fmt.Errorf("crypto: read key: %w", err)
	}
	if len(key) != KeySize {
		return nil, ErrKeySize
	}
	return key, nil
}
