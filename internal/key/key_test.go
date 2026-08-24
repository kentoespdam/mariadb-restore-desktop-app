package key

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateAndLoad(t *testing.T) {
	dir := t.TempDir()

	k, err := Generate(dir)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(k) != KeySize {
		t.Fatalf("key size = %d, want %d", len(k), KeySize)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if string(k) != string(loaded) {
		t.Fatal("loaded key differs from generated key")
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load(t.TempDir())
	if err != ErrKeyNotFound {
		t.Fatalf("Load missing: got %v, want ErrKeyNotFound", err)
	}
}

func TestLoadCorrupt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, KeyFileName), []byte("short"), 0600)

	_, err := Load(dir)
	if err != ErrKeyCorrupt {
		t.Fatalf("Load corrupt: got %v, want ErrKeyCorrupt", err)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if Exists(dir) {
		t.Fatal("Exists returned true for empty dir")
	}
	Generate(dir)
	if !Exists(dir) {
		t.Fatal("Exists returned false after Generate")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	dir := t.TempDir()
	k, _ := Generate(dir)

	plaintext := []byte("secret credential data")
	enc, err := Encrypt(k, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	dec, err := Decrypt(k, enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(dec) != string(plaintext) {
		t.Fatalf("decrypted = %q, want %q", dec, plaintext)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	dir := t.TempDir()
	k1, _ := Generate(dir)
	k2 := make([]byte, KeySize)
	k2[0] = 1 // different key

	plaintext := []byte("data")
	enc, _ := Encrypt(k1, plaintext)

	_, err := Decrypt(k2, enc)
	if err == nil {
		t.Fatal("Decrypt with wrong key should fail")
	}
}
