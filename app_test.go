package main

import (
	"os"
	"testing"

	"mariadb-restore-desktop-app/internal/catalog"
	"mariadb-restore-desktop-app/internal/key"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	dir := t.TempDir()
	return &App{dir: dir}
}

func TestInitFresh(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()
	result := app.Init()
	if result.Status != "fresh" {
		t.Fatalf("status = %q, want fresh", result.Status)
	}
	if result.KeyExists || result.CatalogExists {
		t.Fatal("expected both false for fresh state")
	}

	if err := app.InitFresh(); err != nil {
		t.Fatalf("InitFresh: %v", err)
	}
	if !key.Exists(app.dir) {
		t.Fatal("key not created")
	}
	if !catalog.Exists(app.dir) {
		t.Fatal("catalog not created")
	}
}

func TestInitReady(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()
	key.Generate(app.dir)
	db, _ := catalog.Open(app.dir)
	db.Close() // close so test cleanup can delete

	result := app.Init()
	if result.Status != "ready" {
		t.Fatalf("status = %q, want ready", result.Status)
	}

	// The app must have its DB opened and keyData loaded so that subsequent operations work!
	if app.db == nil {
		t.Fatal("expected app.db to be initialized on ready status")
	}
	if app.keyData == nil {
		t.Fatal("expected app.keyData to be loaded on ready status")
	}
	if _, err := app.ListProfiles(); err != nil {
		t.Fatalf("ListProfiles failed on ready app: %v", err)
	}
}

func TestInitNeedsRecovery(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()
	db, _ := catalog.Open(app.dir)
	db.Close() // close so test cleanup can delete

	result := app.Init()
	if result.Status != "needs_recovery" {
		t.Fatalf("status = %q, want needs_recovery", result.Status)
	}
	if result.KeyExists {
		t.Fatal("key should not exist")
	}
	if !result.CatalogExists {
		t.Fatal("catalog should exist")
	}
}

func TestRecover(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()
	key.Generate(app.dir)
	db, _ := catalog.Open(app.dir)
	db.Close() // close pre-existing so Recover opens fresh

	if err := app.Recover(); err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if app.keyData == nil {
		t.Fatal("keyData not loaded")
	}
	if app.db == nil {
		t.Fatal("db not opened")
	}
}

func TestResetAndInit(t *testing.T) {
	app := newTestApp(t)
	defer app.Close()
	key.Generate(app.dir)
	db, _ := catalog.Open(app.dir)
	db.Close() // close pre-existing

	if err := app.ResetAndInit(); err != nil {
		t.Fatalf("ResetAndInit: %v", err)
	}
	if !key.Exists(app.dir) {
		t.Fatal("key should exist after reset")
	}
	if !catalog.Exists(app.dir) {
		t.Fatal("catalog should exist after reset")
	}

	// Old catalog data should be gone
	var count int
	app.db.QueryRow("SELECT COUNT(*) FROM server_profiles").Scan(&count)
	if count != 0 {
		t.Fatalf("server_profiles count = %d, want 0", count)
	}
}

func TestResolveAppDir(t *testing.T) {
	dir := resolveAppDir()
	if dir == "" {
		t.Fatal("resolveAppDir returned empty")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("resolveAppDir dir does not exist: %v", err)
	}
}

func TestGreet(t *testing.T) {
	app := newTestApp(t)
	got := app.Greet("World")
	want := "Hello World, It's show time!"
	if got != want {
		t.Fatalf("Greet = %q, want %q", got, want)
	}
}
