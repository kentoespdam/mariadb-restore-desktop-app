// Package crypto provides AES-256-GCM primitives and the app.key lifecycle.
//
// core has no Wails or SQLite dependency — pure stdlib only.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// KeySize is the required length of an AES-256 key in bytes.
const KeySize = 32

// nonceSize is the standard GCM nonce length.
const nonceSize = 12

// ErrKeySize is returned when a key has the wrong length.
var ErrKeySize = fmt.Errorf("crypto: key must be %d bytes", KeySize)

// Encrypt seals plaintext with key using AES-256-GCM. The returned blob is
// `nonce || ciphertext || tag` (nonce prepended, 12 bytes).
func Encrypt(key, plaintext []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrKeySize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: read nonce: %w", err)
	}
	sealed := gcm.Seal(nil, nonce, plaintext, nil)
	return append(nonce, sealed...), nil
}

// Decrypt opens a blob produced by Encrypt (nonce-prepended).
func Decrypt(key, blob []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrKeySize
	}
	if len(blob) < nonceSize {
		return nil, fmt.Errorf("crypto: blob too short")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}
	nonce, ct := blob[:nonceSize], blob[nonceSize:]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: open: %w", err)
	}
	return pt, nil
}
