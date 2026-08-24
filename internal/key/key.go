package key

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const KeyFileName = "app.key"
const KeySize = 32 // 256-bit AES

var (
	ErrKeyNotFound = errors.New("app.key not found")
	ErrKeyCorrupt  = errors.New("app.key is corrupt or wrong size")
)

// Generate creates a new 256-bit AES-GCM key and writes it to app.key in dir.
func Generate(dir string) ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	path := filepath.Join(dir, KeyFileName)
	if err := os.WriteFile(path, key, 0600); err != nil {
		return nil, fmt.Errorf("write key: %w", err)
	}
	return key, nil
}

// Load reads app.key from dir. Returns ErrKeyNotFound if missing, ErrKeyCorrupt if wrong size.
func Load(dir string) ([]byte, error) {
	path := filepath.Join(dir, KeyFileName)
	key, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}
	if len(key) != KeySize {
		return nil, ErrKeyCorrupt
	}
	return key, nil
}

// Exists checks if app.key exists in dir.
func Exists(dir string) bool {
	path := filepath.Join(dir, KeyFileName)
	_, err := os.Stat(path)
	return err == nil
}

// Encrypt encrypts plaintext using AES-GCM with the given key.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext using AES-GCM with the given key.
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrKeyCorrupt
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ct, nil)
}
