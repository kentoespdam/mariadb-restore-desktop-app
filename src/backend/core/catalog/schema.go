package catalog

import "fmt"

// schema is the canonical DDL. Run on every Open.
const schema = `
CREATE TABLE IF NOT EXISTS profiles (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    host          BLOB NOT NULL, -- encrypted
    port          INTEGER NOT NULL,
    user          BLOB NOT NULL, -- encrypted
    password      BLOB NOT NULL, -- encrypted
    ssl_mode      TEXT NOT NULL DEFAULT 'preferred',
    created_at    INTEGER NOT NULL,
    updated_at    INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS objects (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    dump_path     TEXT NOT NULL,
    database_name TEXT NOT NULL,
    object_name   TEXT NOT NULL,
    object_type   TEXT NOT NULL,
    start_byte    INTEGER NOT NULL,
    end_byte      INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_objects_dump
    ON objects(dump_path, database_name, object_name);
`

// migrate applies the schema. Idempotent (every statement uses IF NOT EXISTS).
func (s *Store) migrate() error {
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("catalog: migrate: %w", err)
	}
	return nil
}
