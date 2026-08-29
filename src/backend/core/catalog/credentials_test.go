package catalog

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	crypto "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
)

func mustStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func mustKey(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, crypto.KeySize)
	for i := range k {
		k[i] = byte(i)
	}
	return k
}

func TestSaveAndLoadProfile(t *testing.T) {
	s := mustStore(t)
	key := mustKey(t)
	p := &Profile{
		Name:     "local",
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "root",
		Password: "s3cret",
		SSLMode:  "preferred",
	}
	if err := s.SaveProfile(p, key); err != nil {
		t.Fatalf("save: %v", err)
	}
	if p.ID == "" {
		t.Fatal("ID should be assigned")
	}

	got, err := s.LoadProfile("local", key)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Host != "127.0.0.1" || got.User != "root" || got.Password != "s3cret" {
		t.Fatalf("decrypt mismatch: %+v", got)
	}
	if got.Port != 3306 {
		t.Fatalf("port = %d", got.Port)
	}
}

func TestPasswordEncryptedAtRest(t *testing.T) {
	s := mustStore(t)
	key := mustKey(t)
	plaintext := "supersecret"
	if err := s.SaveProfile(&Profile{
		Name: "p1", Host: "h", Port: 1, User: "u", Password: plaintext,
	}, key); err != nil {
		t.Fatal(err)
	}
	// raw row read — password must be ciphertext, not plaintext
	var raw []byte
	if err := s.DB().QueryRow(`SELECT password FROM profiles WHERE name='p1'`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(plaintext)) {
		t.Fatalf("password leaked into raw row: %q", raw)
	}
}

func TestLoadProfileNotFound(t *testing.T) {
	s := mustStore(t)
	if _, err := s.LoadProfile("nope", mustKey(t)); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("want ErrProfileNotFound, got %v", err)
	}
}

func TestListAndDelete(t *testing.T) {
	s := mustStore(t)
	key := mustKey(t)
	for _, n := range []string{"a", "b", "c"} {
		if err := s.SaveProfile(&Profile{Name: n, Host: "h", Port: 1, User: "u", Password: "p"}, key); err != nil {
			t.Fatal(err)
		}
	}
	names, err := s.ListProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(names, ",") != "a,b,c" {
		t.Fatalf("list = %v", names)
	}
	if err := s.DeleteProfile("b"); err != nil {
		t.Fatal(err)
	}
	names, _ = s.ListProfiles()
	if strings.Join(names, ",") != "a,c" {
		t.Fatalf("after delete = %v", names)
	}
}

func TestProfileTimestamps(t *testing.T) {
	s := mustStore(t)
	key := mustKey(t)
	p := &Profile{Name: "t", Host: "h", Port: 1, User: "u", Password: "p"}
	before := time.Now().Add(-time.Second)
	if err := s.SaveProfile(p, key); err != nil {
		t.Fatal(err)
	}
	if p.CreatedAt.Before(before) {
		t.Fatalf("created_at too old: %v", p.CreatedAt)
	}
}
