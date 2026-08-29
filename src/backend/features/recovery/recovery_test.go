package recovery

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	crypto "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/catalog"
)

type staticModal struct{ d Decision }

func (s staticModal) Show(_ context.Context) (Decision, error) { return s.d, nil }

func setup(t *testing.T) (*catalog.Store, Paths) {
	t.Helper()
	dir := t.TempDir()
	paths := Paths{KeyPath: crypto.KeyPath(dir), CatalogDir: dir}
	// pre-create the catalog file (simulate an old install)
	if err := os.WriteFile(filepath.Join(dir, "catalog.sqlite"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	cat, err := catalog.Open(filepath.Join(dir, "catalog.sqlite"))
	if err != nil {
		t.Fatalf("catalog open: %v", err)
	}
	t.Cleanup(func() { cat.Close() })
	return cat, paths
}

func TestHandleMissingKeyCancel(t *testing.T) {
	cat, paths := setup(t)
	err := HandleMissingKey(context.Background(), staticModal{d: DecisionCancel}, cat, paths)
	if !errors.Is(err, ErrUserCancelled) {
		t.Fatalf("want ErrUserCancelled, got %v", err)
	}
	// catalog still present
	if _, err := os.Stat(filepath.Join(paths.CatalogDir, "catalog.sqlite")); err != nil {
		t.Fatalf("catalog should still exist: %v", err)
	}
}

func TestHandleMissingKeyReset(t *testing.T) {
	cat, paths := setup(t)
	// add a profile so we can verify wipe removed it
	if err := cat.SaveProfile(&catalog.Profile{Name: "p", Host: "h", Port: 1, User: "u", Password: "p"},
		make([]byte, crypto.KeySize)); err != nil {
		t.Fatal(err)
	}

	err := HandleMissingKey(context.Background(), staticModal{d: DecisionReset}, cat, paths)
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	// catalog file removed
	if _, err := os.Stat(filepath.Join(paths.CatalogDir, "catalog.sqlite")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("catalog should be gone, got %v", err)
	}
	// new key in place
	k, err := crypto.LoadKey(paths.KeyPath)
	if err != nil {
		t.Fatalf("load key: %v", err)
	}
	if len(k) != crypto.KeySize {
		t.Fatalf("key length = %d", len(k))
	}
}

func TestHandleMissingKeyUnknownDecision(t *testing.T) {
	cat, paths := setup(t)
	if err := HandleMissingKey(context.Background(), staticModal{d: Decision("maybe")}, cat, paths); err == nil {
		t.Fatal("want error on unknown decision")
	}
}
