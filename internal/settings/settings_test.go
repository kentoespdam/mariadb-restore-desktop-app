package settings

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"mariadb-restore-desktop-app/internal/catalog"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := catalog.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestGetSetPath(t *testing.T) {
	db := setupTestDB(t)

	// Get non-existent
	val, err := GetPath(db, KeyMariaDBPath)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty, got %q", val)
	}

	// Set
	if err := SetPath(db, KeyMariaDBPath, "/usr/bin/mariadb"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Get
	val, err = GetPath(db, KeyMariaDBPath)
	if err != nil {
		t.Fatalf("Get after Set: %v", err)
	}
	if val != "/usr/bin/mariadb" {
		t.Fatalf("got %q, want /usr/bin/mariadb", val)
	}

	// Upsert
	SetPath(db, KeyMariaDBPath, "/opt/mariadb/bin/mariadb")
	val, _ = GetPath(db, KeyMariaDBPath)
	if val != "/opt/mariadb/bin/mariadb" {
		t.Fatalf("upsert failed: got %q", val)
	}
}

func TestResolveUsesStoredOverDiscovered(t *testing.T) {
	db := setupTestDB(t)
	SetPath(db, KeyMariaDBPath, "/custom/mariadb")

	resolved := ResolveMariaDB(db, t.TempDir())
	if resolved != "/custom/mariadb" {
		t.Fatalf("expected stored path, got %q", resolved)
	}
}

func TestResolveFallsBackToDiscovery(t *testing.T) {
	db := setupTestDB(t)

	resolved := ResolveMariaDB(db, t.TempDir())
	// Should be whatever discovery finds (may be empty in test env)
	_ = resolved // no panic is the assertion
}

func TestDiscoverBundled(t *testing.T) {
	dir := t.TempDir()
	// Create a fake mariadb binary
	bin := filepath.Join(dir, "mariadb")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	os.WriteFile(bin, []byte("#!/bin/sh\necho test"), 0755)

	s := Discover(dir)
	if s.MariaDBPath != bin {
		t.Fatalf("expected bundled path %q, got %q", bin, s.MariaDBPath)
	}
}

func TestDiscoverEmptyDir(t *testing.T) {
	s := Discover(t.TempDir())
	// Should not panic, just return empty paths
	_ = s
}

func TestValidateBinary(t *testing.T) {
	dir := t.TempDir()

	// Empty
	if err := ValidateBinary(""); err == nil {
		t.Fatal("expected error for empty path")
	}

	// Non-existent
	if err := ValidateBinary(filepath.Join(dir, "nope")); err == nil {
		t.Fatal("expected error for non-existent file")
	}

	// Directory
	if err := ValidateBinary(dir); err == nil {
		t.Fatal("expected error for directory")
	}

	// Valid file
	path := filepath.Join(dir, "mariadb")
	os.WriteFile(path, []byte("binary"), 0755)
	if err := ValidateBinary(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsMissing(t *testing.T) {
	s := Settings{MariaDBPath: ""}
	if !IsMissing(s) {
		t.Fatal("expected missing when path is empty")
	}

	s.MariaDBPath = "/usr/bin/mariadb"
	if IsMissing(s) {
		t.Fatal("expected not missing when path is set")
	}
}

func TestBannerText(t *testing.T) {
	s := Settings{MariaDBPath: ""}
	msg := BannerText(s)
	if msg == "" {
		t.Fatal("expected banner text when mariadb not found")
	}

	s.MariaDBPath = "/usr/bin/mariadb"
	msg = BannerText(s)
	if msg != "" {
		t.Fatalf("expected empty banner, got %q", msg)
	}
}

func TestFormatPathForDisplay(t *testing.T) {
	got := FormatPathForDisplay("")
	if got != "(not configured)" {
		t.Fatalf("empty: got %q", got)
	}

	got = FormatPathForDisplay("/usr/bin/mariadb")
	if got != "/usr/bin/mariadb" {
		t.Fatalf("got %q", got)
	}
}
