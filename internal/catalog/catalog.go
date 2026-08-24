package catalog

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const CatalogFileName = "catalog.db"

const schema = `
CREATE TABLE IF NOT EXISTS server_profiles (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	host TEXT NOT NULL DEFAULT 'localhost',
	port INTEGER NOT NULL DEFAULT 3306,
	username TEXT NOT NULL DEFAULT 'root',
	encrypted_password BLOB,
	ssl_mode TEXT NOT NULL DEFAULT 'disabled',
	ssl_ca TEXT,
	ssl_cert TEXT,
	ssl_key TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS catalog_objects (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	dump_file TEXT NOT NULL,
	database_name TEXT NOT NULL,
	object_type TEXT NOT NULL,
	object_name TEXT NOT NULL,
	start_byte INTEGER NOT NULL,
	end_byte INTEGER NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS app_last_op (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	operation TEXT NOT NULL,
	dump_file TEXT,
	server_profile_id INTEGER,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS app_state (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
`

// Open opens (or creates) the SQLite catalog in dir.
func Open(dir string) (*sql.DB, error) {
	path := filepath.Join(dir, CatalogFileName)
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open catalog: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping catalog: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return db, nil
}

// Exists checks if catalog.db exists in dir.
func Exists(dir string) bool {
	path := filepath.Join(dir, CatalogFileName)
	_, err := os.Stat(path)
	return err == nil
}

// Remove deletes catalog.db from dir.
func Remove(dir string) error {
	path := filepath.Join(dir, CatalogFileName)
	os.Remove(path + "-wal")
	os.Remove(path + "-shm")
	return os.Remove(path)
}
