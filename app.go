package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"mariadb-restore-desktop-app/internal/catalog"
	"mariadb-restore-desktop-app/internal/key"
	"mariadb-restore-desktop-app/internal/profile"
	"mariadb-restore-desktop-app/internal/scanner"
	"mariadb-restore-desktop-app/internal/settings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App holds all application state. Methods exposed to the frontend via Wails.
type App struct {
	ctx     context.Context
	dir     string // executable scope: dir next to the binary
	keyData []byte
	db      *sql.DB
}

// NewApp creates an App using Executable Scope (dir next to the binary).
func NewApp() *App {
	dir := resolveAppDir()
	return &App{dir: dir}
}

// Dir returns the application directory (executable scope).
func (a *App) Dir() string { return a.dir }

// startup is called by Wails on application start.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// InitResult tells the frontend what state the app is in after launch.
type InitResult struct {
	Status          string `json:"status"` // "ready", "needs_recovery", "fresh"
	KeyExists       bool   `json:"keyExists"`
	CatalogExists   bool   `json:"catalogExists"`
}

// Init checks key + catalog state and returns the result to the frontend.
// The frontend uses this to decide whether to show Smart Recovery or proceed.
func (a *App) Init() InitResult {
	keyExists := key.Exists(a.dir)
	catalogExists := catalog.Exists(a.dir)

	switch {
	case keyExists && catalogExists:
		return InitResult{Status: "ready", KeyExists: true, CatalogExists: true}
	case !keyExists && catalogExists:
		return InitResult{Status: "needs_recovery", KeyExists: false, CatalogExists: true}
	default:
		return InitResult{Status: "fresh", KeyExists: keyExists, CatalogExists: catalogExists}
	}
}

// InitFresh generates a new key and opens the catalog. Call after "fresh" or "reset".
func (a *App) InitFresh() error {
	k, err := key.Generate(a.dir)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	a.keyData = k

	db, err := catalog.Open(a.dir)
	if err != nil {
		return fmt.Errorf("open catalog: %w", err)
	}
	a.db = db
	return nil
}

// ResetAndInit wipes catalog + old key, generates fresh key, opens catalog.
func (a *App) ResetAndInit() error {
	catalog.Remove(a.dir)
	os.Remove(filepath.Join(a.dir, key.KeyFileName))
	return a.InitFresh()
}

// Recover loads existing key and opens catalog (for Smart Recovery cancel path).
func (a *App) Recover() error {
	k, err := key.Load(a.dir)
	if err != nil {
		return fmt.Errorf("load key: %w", err)
	}
	a.keyData = k

	db, err := catalog.Open(a.dir)
	if err != nil {
		return fmt.Errorf("open catalog: %w", err)
	}
	a.db = db
	return nil
}

// Close closes the database connection.
func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// --- Server Profile methods ---

// ListProfiles returns all server profiles.
func (a *App) ListProfiles() ([]profile.Profile, error) {
	if a.db == nil {
		return nil, fmt.Errorf("catalog not initialized")
	}
	return profile.ListProfiles(a.db)
}

// SaveProfile creates or updates a profile.
func (a *App) SaveProfile(p *profile.Profile) (int64, error) {
	if a.db == nil {
		return 0, fmt.Errorf("catalog not initialized")
	}
	if p.ID == 0 {
		return profile.CreateProfile(a.db, a.keyData, p)
	}
	return p.ID, profile.UpdateProfile(a.db, a.keyData, p)
}

// DeleteProfile removes a profile by ID.
func (a *App) DeleteProfile(id int64) error {
	if a.db == nil {
		return fmt.Errorf("catalog not initialized")
	}
	return profile.DeleteProfile(a.db, id)
}

// TestConnection tests a MariaDB connection.
func (a *App) TestConnection(p *profile.Profile) profile.TestResult {
	pw, _ := profile.DecryptPassword(a.keyData, p.EncryptedPassword)
	if pw == "" {
		pw = p.Password
	}
	return profile.TestConnection(a.ctx, p, pw)
}

// GetProfile returns a single profile by ID.
func (a *App) GetProfile(id int64) (*profile.Profile, error) {
	if a.db == nil {
		return nil, fmt.Errorf("catalog not initialized")
	}
	return profile.GetProfile(a.db, id)
}

// --- Scanner / Analyze methods ---

// CatalogObject is exposed to the frontend for the table checklist.
type CatalogObject struct {
	ID           int64  `json:"id"`
	DatabaseName string `json:"databaseName"`
	ObjectType   string `json:"objectType"`
	ObjectName   string `json:"objectName"`
	StartByte    int64  `json:"startByte"`
	EndByte      int64  `json:"endByte"`
}

// Analyze scans a dump file and populates the catalog. Progress is emitted via Wails events.
func (a *App) Analyze(dumpFile string) error {
	if a.db == nil {
		return fmt.Errorf("catalog not initialized")
	}
	return scanner.Scan(a.ctx, a.db, dumpFile, func(p scanner.Progress) {
		runtime.EventsEmit(a.ctx, "scan:progress", p)
	})
}

// GetCatalogObjects returns all catalog objects for a dump file.
func (a *App) GetCatalogObjects(dumpFile string) ([]CatalogObject, error) {
	if a.db == nil {
		return nil, fmt.Errorf("catalog not initialized")
	}
	rows, err := a.db.Query(
		"SELECT id, database_name, object_type, object_name, start_byte, end_byte FROM catalog_objects WHERE dump_file=? ORDER BY start_byte",
		dumpFile,
	)
	if err != nil {
		return nil, fmt.Errorf("query catalog: %w", err)
	}
	defer rows.Close()

	var objs []CatalogObject
	for rows.Next() {
		var o CatalogObject
		if err := rows.Scan(&o.ID, &o.DatabaseName, &o.ObjectType, &o.ObjectName, &o.StartByte, &o.EndByte); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		objs = append(objs, o)
	}
	return objs, rows.Err()
}

// --- Settings methods ---

// GetSettings returns current binary paths and discovery status.
func (a *App) GetSettings() settings.Settings {
	s := settings.Discover(a.dir)
	if a.db != nil {
		if p, _ := settings.GetPath(a.db, settings.KeyMariaDBPath); p != "" {
			s.MariaDBPath = p
		}
		if p, _ := settings.GetPath(a.db, settings.KeyDumpPath); p != "" {
			s.MariaDBDumpPath = p
		}
	}
	s.MariaDBFound = s.MariaDBPath != ""
	return s
}

// SetMariaDBPath stores the mariadb binary path override.
func (a *App) SetMariaDBPath(path string) error {
	if err := settings.ValidateBinary(path); err != nil {
		return err
	}
	return settings.SetPath(a.db, settings.KeyMariaDBPath, path)
}

// SetMariaDBDumpPath stores the mariadb-dump binary path override.
func (a *App) SetMariaDBDumpPath(path string) error {
	if err := settings.ValidateBinary(path); err != nil {
		return err
	}
	return settings.SetPath(a.db, settings.KeyDumpPath, path)
}

// DiscoverBinaries re-runs auto-discovery and returns fresh results.
func (a *App) DiscoverBinaries() settings.Settings {
	return settings.Discover(a.dir)
}

// Greet is the Wails template method — kept for build compatibility.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// resolveAppDir returns the directory next to the executable.
func resolveAppDir() string {
	exe, err := os.Executable()
	if err != nil {
		// fallback to working directory
		wd, _ := os.Getwd()
		return wd
	}
	return filepath.Dir(exe)
}
