package recovery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	crypto "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/catalog"
)

// KeyDir is the directory app.key lives in. ExeDir is the directory
// returned by filepath.Dir(os.Executable()).
func KeyDir(exeDir string) string { return exeDir }

// Paths groups the on-disk locations recovery needs to wipe.
type Paths struct {
	KeyPath    string
	CatalogDir string
}

// DefaultPaths returns the on-disk paths derived from the executable
// directory.
func DefaultPaths(exeDir string) Paths {
	return Paths{
		KeyPath:    crypto.KeyPath(exeDir),
		CatalogDir: exeDir,
	}
}

// Wipe removes the catalog file and any side-files (WAL, journal).
// Returns the absolute path that was removed.
func Wipe(p Paths) (string, error) {
	pattern := filepath.Join(p.CatalogDir, "catalog.sqlite*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("recovery: glob: %w", err)
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("recovery: remove %s: %w", m, err)
		}
	}
	if len(matches) == 0 {
		return "", nil
	}
	return matches[0], nil
}

// HandleMissingKey is the entry point. It asks the modal what to do,
// then either returns ErrUserCancelled (Cancel) or wipes the catalog
// and regenerates a fresh key (Reset).
func HandleMissingKey(ctx context.Context, modal Modal, cat *catalog.Store, paths Paths) error {
	if modal == nil {
		return fmt.Errorf("recovery: nil modal")
	}
	decision, err := modal.Show(ctx)
	if err != nil {
		return fmt.Errorf("recovery: show: %w", err)
	}
	switch decision {
	case DecisionCancel:
		return ErrUserCancelled
	case DecisionReset:
		// Regenerate a key first so the user can keep going after
		// wipe. Wipe before the new key is written.
		removed, err := Wipe(paths)
		if err != nil {
			return err
		}
		if cat != nil {
			_ = cat.Close()
		}
		if _, err := crypto.GenerateKey(paths.KeyPath); err != nil {
			return err
		}
		_ = removed // caller can log; not fatal
		return nil
	default:
		return fmt.Errorf("recovery: unknown decision %q", decision)
	}
}
