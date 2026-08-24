package profile

import (
	"database/sql"
	"testing"

	"mariadb-restore-desktop-app/internal/catalog"
	"mariadb-restore-desktop-app/internal/key"
)

func setupTestDB(t *testing.T) (*sql.DB, []byte) {
	t.Helper()
	dir := t.TempDir()
	k, _ := key.Generate(dir)
	db, err := catalog.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, k
}

func TestCreateAndList(t *testing.T) {
	db, k := setupTestDB(t)

	id, err := CreateProfile(db, k, &Profile{
		Name:     "dev-server",
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}

	profiles, err := ListProfiles(db)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "dev-server" {
		t.Fatalf("name = %q, want dev-server", profiles[0].Name)
	}
	if len(profiles[0].EncryptedPassword) == 0 {
		t.Fatal("password should be encrypted, not empty")
	}
}

func TestGetProfile(t *testing.T) {
	db, k := setupTestDB(t)

	id, _ := CreateProfile(db, k, &Profile{
		Name: "prod", Host: "10.0.0.1", Port: 3307, Username: "admin", Password: "pw",
	})

	p, err := GetProfile(db, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if p.Host != "10.0.0.1" || p.Port != 3307 {
		t.Fatalf("unexpected: %s %d", p.Host, p.Port)
	}
}

func TestUpdateProfile(t *testing.T) {
	db, k := setupTestDB(t)

	id, _ := CreateProfile(db, k, &Profile{
		Name: "old-name", Host: "1.1.1.1", Port: 3306, Username: "u", Password: "p",
	})

	err := UpdateProfile(db, k, &Profile{
		ID: id, Name: "new-name", Host: "2.2.2.2", Port: 3307, Username: "u2", Password: "newpw",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	p, _ := GetProfile(db, id)
	if p.Name != "new-name" || p.Host != "2.2.2.2" || p.Port != 3307 {
		t.Fatalf("update didn't apply: %s %s %d", p.Name, p.Host, p.Port)
	}
}

func TestUpdateWithoutPassword(t *testing.T) {
	db, k := setupTestDB(t)

	id, _ := CreateProfile(db, k, &Profile{
		Name: "keep-pw", Host: "1.1.1.1", Port: 3306, Username: "u", Password: "secret",
	})

	// Update without password — should keep encrypted password intact
	err := UpdateProfile(db, k, &Profile{
		ID: id, Name: "keep-pw", Host: "2.2.2.2", Port: 3306, Username: "u",
	})
	if err != nil {
		t.Fatalf("Update no pw: %v", err)
	}

	p, _ := GetProfile(db, id)
	if len(p.EncryptedPassword) == 0 {
		t.Fatal("password should still be present after update without password")
	}
}

func TestDeleteProfile(t *testing.T) {
	db, k := setupTestDB(t)

	id, _ := CreateProfile(db, k, &Profile{
		Name: "to-delete", Host: "1.1.1.1", Port: 3306, Username: "u", Password: "p",
	})

	if err := DeleteProfile(db, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	profiles, _ := ListProfiles(db)
	if len(profiles) != 0 {
		t.Fatalf("expected 0 profiles after delete, got %d", len(profiles))
	}
}

func TestDecryptPassword(t *testing.T) {
	dir := t.TempDir()
	k, _ := key.Generate(dir)

	encrypted, err := key.Encrypt(k, []byte("my-secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	decrypted, err := DecryptPassword(k, encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if decrypted != "my-secret" {
		t.Fatalf("decrypted = %q, want my-secret", decrypted)
	}
}

func TestDecryptPasswordEmpty(t *testing.T) {
	decrypted, err := DecryptPassword(nil, nil)
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}
	if decrypted != "" {
		t.Fatalf("expected empty, got %q", decrypted)
	}
}

func TestDefaultsApplied(t *testing.T) {
	db, k := setupTestDB(t)

	id, _ := CreateProfile(db, k, &Profile{
		Name: "defaults", Host: "h", Username: "u",
	})

	p, _ := GetProfile(db, id)
	if p.Port != 3306 {
		t.Fatalf("port = %d, want 3306", p.Port)
	}
	if p.SSLMode != "disabled" {
		t.Fatalf("ssl_mode = %q, want disabled", p.SSLMode)
	}
}
