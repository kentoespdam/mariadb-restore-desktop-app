package settings

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	KeyMariaDBPath = "mariadb_path"
	KeyDumpPath    = "mariadbdump_path"
)

// Settings holds the resolved binary paths and discovery status.
type Settings struct {
	MariaDBPath    string `json:"mariadbPath"`
	MariaDBDumpPath string `json:"mariadbdumpPath"`
	MariaDBFound   bool   `json:"mariadbFound"`
}

// Discover finds mariadb and mariadb-dump executables.
// Priority: bundled (beside binary) → PATH → platform-specific defaults.
func Discover(appDir string) Settings {
	s := Settings{}

	// Try bundled path first (for AppImage / portable)
	s.MariaDBPath = findBundled(appDir, "mariadb")
	s.MariaDBDumpPath = findBundled(appDir, "mariadb-dump")

	// Fall back to PATH
	if s.MariaDBPath == "" {
		s.MariaDBPath = findInPATH("mariadb")
	}
	if s.MariaDBDumpPath == "" {
		s.MariaDBDumpPath = findInPATH("mariadb-dump")
	}

	// Platform-specific fallbacks
	if s.MariaDBPath == "" {
		s.MariaDBPath = platformDefault("mariadb")
	}
	if s.MariaDBDumpPath == "" {
		s.MariaDBDumpPath = platformDefault("mariadb-dump")
	}

	s.MariaDBFound = s.MariaDBPath != ""
	return s
}

// GetPath reads a setting from app_state.
func GetPath(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM app_state WHERE key=?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting %s: %w", key, err)
	}
	return value, nil
}

// SetPath writes a setting to app_state (upsert).
func SetPath(db *sql.DB, key, value string) error {
	_, err := db.Exec(
		"INSERT INTO app_state (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value",
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set setting %s: %w", key, err)
	}
	return nil
}

// ResolveMariaDB returns the effective mariadb path: stored override > discovered.
func ResolveMariaDB(db *sql.DB, appDir string) string {
	if stored, _ := GetPath(db, KeyMariaDBPath); stored != "" {
		return stored
	}
	return Discover(appDir).MariaDBPath
}

// ResolveMariaDBDump returns the effective mariadb-dump path.
func ResolveMariaDBDump(db *sql.DB, appDir string) string {
	if stored, _ := GetPath(db, KeyDumpPath); stored != "" {
		return stored
	}
	return Discover(appDir).MariaDBDumpPath
}

func findBundled(appDir, name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(appDir, name)
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path
	}
	return ""
}

func findInPATH(name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func platformDefault(name string) string {
	switch runtime.GOOS {
	case "linux":
		path := filepath.Join("/usr/bin", name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	case "windows":
		// Common MariaDB install locations on Windows
		candidates := []string{
			`C:\Program Files\MariaDB 10.11\bin` + `\` + name + `.exe`,
			`C:\Program Files\MariaDB 11.4\bin` + `\` + name + `.exe`,
			`C:\Program Files\MariaDB\bin` + `\` + name + `.exe`,
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	}
	return ""
}

// IsMissing returns true if the mariadb binary path is empty (not found).
func IsMissing(s Settings) bool {
	return s.MariaDBPath == ""
}

// BannerText returns a user-facing message if mariadb is not found.
func BannerText(s Settings) string {
	if s.MariaDBPath == "" {
		return "mariadb CLI not found — configure path in Settings"
	}
	return ""
}

// ValidateBinary checks if a path points to an executable file.
func ValidateBinary(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not a binary", path)
	}
	// On Unix, also check executable permission
	if runtime.GOOS != "windows" {
		if info.Mode().Perm()&0111 == 0 {
			return fmt.Errorf("%s is not executable", path)
		}
	}
	return nil
}

// FormatPathForDisplay shortens a path for UI display.
func FormatPathForDisplay(path string) string {
	if path == "" {
		return "(not configured)"
	}
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
