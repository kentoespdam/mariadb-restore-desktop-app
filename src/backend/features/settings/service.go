// Package settings persists user-level configuration (binary paths)
// beside the catalog. The schema is intentionally minimal: a single
// JSON file holding the keys the UI cares about. The catalog and the
// app.key are managed elsewhere; settings does not touch them.
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Default binary paths. These are what the BE falls back to when no
// override is stored. On Linux they match the distro package layout;
// on Windows they're the conventional PATH names; we let the Wails
// runtime's PATH lookup find them at exec time.
const (
	DefaultMariadbPath     = "mariadb"
	DefaultMariadbDumpPath = "mariadb-dump"
)

// Settings is the JSON shape persisted to disk. Field tags are stable
// for any future on-disk reader; the View/Input split keeps the wire
// shape explicit.
type Settings struct {
	MariadbPath     string `json:"mariadbPath"`
	MariadbDumpPath string `json:"mariadbDumpPath"`
}

// Input is the FE-facing payload for SaveSettings.
type Input struct {
	MariadbPath     string `json:"mariadbPath"`
	MariadbDumpPath string `json:"mariadbDumpPath"`
}

// Service is the settings store. exeDir is the directory returned by
// filepath.Dir(os.Executable()) and is the directory the JSON file
// lives in. keyBits is reported to the FE for the Settings screen
// fingerprint line (CONTEXT: never the key bytes).
type Service struct {
	exeDir  string
	path    string
	keyBits int

	mu sync.RWMutex
}

// New returns a service rooted at exeDir. The settings file is read
// lazily on first Get; Save persists atomically.
func New(exeDir string, keyBits int) *Service {
	return &Service{
		exeDir:  exeDir,
		path:    filepath.Join(exeDir, "settings.json"),
		keyBits: keyBits,
	}
}

// Get returns the current settings, loading from disk on first call.
// Missing file -> defaults. Disk read failure other than missing
// is propagated.
func (s *Service) Get() (Settings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	raw, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaults(), nil
		}
		return Settings{}, fmt.Errorf("settings: read: %w", err)
	}
	var out Settings
	if err := json.Unmarshal(raw, &out); err != nil {
		return Settings{}, fmt.Errorf("settings: parse: %w", err)
	}
	if out.MariadbPath == "" {
		out.MariadbPath = DefaultMariadbPath
	}
	if out.MariadbDumpPath == "" {
		out.MariadbDumpPath = DefaultMariadbDumpPath
	}
	return out, nil
}

// GetMariadbPath returns just the mariadb binary path. Used by the
// restore service to discover the binary at app start.
func (s *Service) GetMariadbPath() string {
	cur, err := s.Get()
	if err != nil {
		return DefaultMariadbPath
	}
	return cur.MariadbPath
}

// GetMariadbDumpPath returns just the mariadb-dump binary path. Used
// by the backup service to discover the binary at app start.
func (s *Service) GetMariadbDumpPath() string {
	cur, err := s.Get()
	if err != nil {
		return DefaultMariadbDumpPath
	}
	return cur.MariadbDumpPath
}

// Save atomically replaces the on-disk settings via a temp file +
// rename, so a crash mid-write cannot corrupt the JSON.
func (s *Service) Save(in Input) error {
	if in.MariadbPath == "" {
		in.MariadbPath = DefaultMariadbPath
	}
	if in.MariadbDumpPath == "" {
		in.MariadbDumpPath = DefaultMariadbDumpPath
	}
	blob, err := json.MarshalIndent(Settings{
		MariadbPath:     in.MariadbPath,
		MariadbDumpPath: in.MariadbDumpPath,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("settings: marshal: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, blob, 0o600); err != nil {
		return fmt.Errorf("settings: write tmp: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("settings: rename: %w", err)
	}
	return nil
}

// KeyBits is reported to the FE so the Settings screen can show the
// fingerprint without ever exposing the key bytes themselves.
func (s *Service) KeyBits() int { return s.keyBits }

func defaults() Settings {
	return Settings{
		MariadbPath:     DefaultMariadbPath,
		MariadbDumpPath: DefaultMariadbDumpPath,
	}
}