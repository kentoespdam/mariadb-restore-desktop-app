package catalog

import (
	"fmt"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/scanner"
)

// Object is the catalog row for one offset recorded by the Byte-Offset
// Scanner. ObjectType uses scanner.ObjectType so the wire and the in-memory
// struct stay in sync; DatabaseName comes from the most-recent USE
// statement at scan time.
type Object struct {
	ID           int
	DumpPath     string
	DatabaseName string
	ObjectName   string
	ObjectType   string
	StartByte    int64
	EndByte      int64
}

// ReplaceObjectsForDump wipes every previously recorded object for the
// given dump path and inserts the supplied slice in a single transaction.
// The catalog uses (dump_path, database_name, object_name) as the natural
// key, so re-analyzing the same file is idempotent at the row level.
func (s *Store) ReplaceObjectsForDump(dumpPath string, objs []scanner.Object) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("catalog: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM objects WHERE dump_path = ?`, dumpPath); err != nil {
		return fmt.Errorf("catalog: clear objects: %w", err)
	}
	stmt, err := tx.Prepare(`
        INSERT INTO objects(dump_path, database_name, object_name, object_type, start_byte, end_byte)
        VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("catalog: prepare insert: %w", err)
	}
	defer stmt.Close()
	for _, o := range objs {
		if _, err := stmt.Exec(
			dumpPath,
			o.DatabaseName,
			o.ObjectName,
			string(o.ObjectType),
			o.StartByte,
			o.EndByte,
		); err != nil {
			return fmt.Errorf("catalog: insert object: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("catalog: commit: %w", err)
	}
	return nil
}

// ListObjectsForDump returns every catalog row for dumpPath ordered by
// (database_name, object_name) so the UI can render a stable list.
func (s *Store) ListObjectsForDump(dumpPath string) ([]Object, error) {
	rows, err := s.db.Query(`
        SELECT id, dump_path, database_name, object_name, object_type, start_byte, end_byte
        FROM objects
        WHERE dump_path = ?
        ORDER BY database_name, object_name, id`, dumpPath)
	if err != nil {
		return nil, fmt.Errorf("catalog: list objects: %w", err)
	}
	defer rows.Close()

	var out []Object
	for rows.Next() {
		var o Object
		if err := rows.Scan(
			&o.ID,
			&o.DumpPath,
			&o.DatabaseName,
			&o.ObjectName,
			&o.ObjectType,
			&o.StartByte,
			&o.EndByte,
		); err != nil {
			return nil, fmt.Errorf("catalog: scan object: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// DeleteObjectsForDump removes every object recorded for dumpPath.
// Used by reset-and-reinit so the catalog starts empty after a wipe.
func (s *Store) DeleteObjectsForDump(dumpPath string) error {
	_, err := s.db.Exec(`DELETE FROM objects WHERE dump_path = ?`, dumpPath)
	return err
}