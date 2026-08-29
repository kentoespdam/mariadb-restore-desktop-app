package crypto

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, KeySize)
	plaintext := []byte("the quick brown fox jumps over the lazy dog")

	blob, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Equal(blob[nonceSize:], plaintext) {
		t.Fatal("ciphertext should not equal plaintext")
	}
	got, err := Decrypt(key, blob)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("roundtrip mismatch: got %q want %q", got, plaintext)
	}
}

func TestEncryptWrongKeySize(t *testing.T) {
	if _, err := Encrypt(make([]byte, 16), []byte("x")); !errors.Is(err, ErrKeySize) {
		t.Fatalf("want ErrKeySize, got %v", err)
	}
}

func TestDecryptTampered(t *testing.T) {
	key := bytes.Repeat([]byte{0x11}, KeySize)
	blob, _ := Encrypt(key, []byte("hello"))
	blob[len(blob)-1] ^= 0xFF // flip a tag bit
	if _, err := Decrypt(key, blob); err == nil {
		t.Fatal("want error on tampered ciphertext")
	}
}

func TestGenerateAndLoadKey(t *testing.T) {
	dir := t.TempDir()
	path := KeyPath(dir)

	k1, err := GenerateKey(path)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(k1) != KeySize {
		t.Fatalf("key length = %d, want %d", len(k1), KeySize)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("permissions = %o, want 0600", perm)
	}

	k2, err := LoadKey(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !bytes.Equal(k1, k2) {
		t.Fatal("loaded key differs from generated key")
	}
}

func TestLoadKeyMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadKey(filepath.Join(dir, "app.key"))
	if !errors.Is(err, ErrMissingKey) {
		t.Fatalf("want ErrMissingKey, got %v", err)
	}
}

func TestLoadKeyWrongSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.key")
	if err := os.WriteFile(path, []byte("short"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadKey(path)
	if !errors.Is(err, ErrKeySize) {
		t.Fatalf("want ErrKeySize, got %v", err)
	}
}
