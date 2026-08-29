// Package catalog is the local SQLite store for Server Profiles and
// scanned object offsets. Sensitive profile fields are encrypted at rest
// using the core/crypto package.
package catalog

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store wraps the catalog SQLite file.
type Store struct {
	db *sql.DB
}

// Open opens the catalog at path. The file is created if missing.
// Ponytail: passing ":memory:" gives an in-memory store for tests.
func Open(path string) (*Store, error) {
	if path != ":memory:" {
		path = filepath.Clean(path)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("catalog: open: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// DB exposes the underlying *sql.DB for callers that need a transaction.
// Avoid in feature code — prefer typed methods on Store.
func (s *Store) DB() *sql.DB { return s.db }
