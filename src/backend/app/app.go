// Package app is the assembly layer. It owns the long-lived
// references (catalog store, key, services) and exposes the Wails
// binding surface. main.go is the only place that calls wails.Run;
// everything else lives here.
package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/catalog"
	crypto "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/backup"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/profile"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/recovery"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/restore"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/settings"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/platform/events"
)

// App is the Wails binding target. Every public method on *App is
// callable from the frontend via the generated bindings.
type App struct {
	Catalog  *catalog.Store
	Key      []byte
	KeyPath  string
	CatPath  string
	Profile  *profile.Service
	Backup   *backup.Service
	Restore  *restore.Service
	Settings *settings.Service
	Recovery *RecoveryService
}

// RecoveryService is the Smart Recovery wiring. The Decision channel
// is exposed so the binding layer can receive the user's choice.
type RecoveryService struct {
	*recovery.Service
	Decision chan recovery.Decision
}

// New assembles the app: loads or generates the app.key, opens the
// catalog, and wires the profile + recovery services. The Wails event
// bus is registered via events.SetWailsContext(ctx) before this is
// called.
//
// exeDir is the directory returned by filepath.Dir(os.Executable());
// pass "" in tests to fall back to a temp dir.
func New(ctx context.Context, exeDir string) (*App, error) {
	if exeDir == "" {
		d, err := os.MkdirTemp("", "mariadb-restore-*")
		if err != nil {
			return nil, err
		}
		exeDir = d
	}
	events.SetWailsContext(ctx)

	keyPath := crypto.KeyPath(exeDir)
	catPath := filepath.Join(exeDir, "catalog.sqlite")

	key, err := crypto.LoadKey(keyPath)
	if err != nil {
		if err != crypto.ErrMissingKey {
			return nil, fmt.Errorf("app: load key: %w", err)
		}
		key, err = crypto.GenerateKey(keyPath)
		if err != nil {
			return nil, err
		}
	}

	cat, err := catalog.Open(catPath)
	if err != nil {
		return nil, fmt.Errorf("app: open catalog: %w", err)
	}

	decisions := make(chan recovery.Decision, 1)
	rs := &RecoveryService{
		Service: &recovery.Service{
			Cat:   cat,
			Paths: recovery.DefaultPaths(exeDir),
			Modal: &recovery.EventModal{
				Bus:      events.Default(ctx),
				Decision: decisions,
			},
		},
		Decision: decisions,
	}

	profileSvc := profile.New(cat, key)
	settingsSvc := settings.New(exeDir, len(key)*8)
	emitter := events.Default(ctx)

	return &App{
		Catalog:  cat,
		Key:      key,
		KeyPath:  keyPath,
		CatPath:  catPath,
		Profile:  profileSvc,
		Backup:   backup.New(profileSvc, emitter, settingsSvc.GetMariadbDumpPath()),
		Restore:  restore.New(profileSvc, cat, emitter, settingsSvc.GetMariadbPath()),
		Settings: settingsSvc,
		Recovery: rs,
	}, nil
}

// Close releases the catalog handle. Call from wails OnShutdown.
func (a *App) Close() error {
	if a.Catalog != nil {
		return a.Catalog.Close()
	}
	return nil
}

// RebindCtx re-wires the event bus to use the live Wails ctx. Call
// from OnStartup. Safe to call once.
func (a *App) RebindCtx(ctx context.Context) {
	events.SetWailsContext(ctx)
	if a.Recovery != nil {
		a.Recovery.Modal = &recovery.EventModal{
			Bus:      events.Default(ctx),
			Decision: a.Recovery.Decision,
		}
	}
	setBgCtx(ctx)
}

// bgCtx is the long-lived background context the binding layer
// passes to backup/restore services. Bound on RebindCtx (which is
// called from Wails OnStartup) so the context lives as long as
// the app.
var (
	bgCtxOnce sync.RWMutex
	bgCtxVal  context.Context
)

func setBgCtx(ctx context.Context) {
	bgCtxOnce.Lock()
	bgCtxVal = ctx
	bgCtxOnce.Unlock()
}

// bgCtx returns the Wails ctx for use as a parent context in
// subprocess commands. If RebindCtx has not been called yet (e.g.
// in a unit test), it falls back to context.Background() so the
// subprocess can still start.
func bgCtx() context.Context {
	bgCtxOnce.RLock()
	defer bgCtxOnce.RUnlock()
	if bgCtxVal != nil {
		return bgCtxVal
	}
	return context.Background()
}
