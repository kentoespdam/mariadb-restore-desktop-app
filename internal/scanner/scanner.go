package scanner

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// ObjectType constants for catalog_objects.type column.
const (
	TypeTable        = "table"
	TypeRoutines     = "routines_block"
	TypeTriggers     = "triggers_block"
	TypeEvents       = "events_block"
)

// Progress is emitted during scanning.
type Progress struct {
	PercentBytes   float64 `json:"percentBytes"`
	DatabasesFound int     `json:"databasesFound"`
	TablesFound    int     `json:"tablesFound"`
	BlocksFound    int     `json:"blocksFound"`
}

// CatalogObject is a row to insert into catalog_objects.
type CatalogObject struct {
	DumpFile     string
	DatabaseName string
	ObjectType   string
	ObjectName   string
	StartByte    int64
	EndByte      int64
}

// Scan reads a dump file in a single pass, populating catalog_objects.
// It never loads the full file into RAM.
// onProgress is called throttled (every ~150ms). Pass nil to skip.
func Scan(ctx context.Context, db *sql.DB, dumpFile string, onProgress func(Progress)) error {
	f, err := os.Open(dumpFile)
	if err != nil {
		return fmt.Errorf("open dump: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat dump: %w", err)
	}
	totalSize := info.Size()

	// Single-slot: clear previous catalog objects for this dump file.
	if _, err := db.Exec("DELETE FROM catalog_objects WHERE dump_file=?", dumpFile); err != nil {
		return fmt.Errorf("clear catalog: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	stmt, err := tx.Prepare(
		"INSERT INTO catalog_objects (dump_file, database_name, object_type, object_name, start_byte, end_byte) VALUES (?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	reader := bufio.NewReader(f)
	var offset int64
	activeDB := ""
	var currentObj *CatalogObject
	dbs := map[string]bool{}
	tables := 0
	blocks := 0

	// Progress throttle
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	progressCh := make(chan Progress, 1)
	go func() {
		for range ticker.C {
			select {
			case p := <-progressCh:
				if onProgress != nil {
					onProgress(p)
				}
			default:
			}
		}
	}()

	for {
		if ctx.Err() != nil {
			if currentObj != nil {
				currentObj.EndByte = offset
				flushObj(stmt, currentObj)
			}
			tx.Commit()
			return ctx.Err()
		}

		line, err := reader.ReadBytes('\n')
		lineLen := int64(len(line))
		currentLine := string(line)

		// Emit progress
		select {
		case progressCh <- Progress{
			PercentBytes:   float64(offset) / float64(totalSize) * 100,
			DatabasesFound: len(dbs),
			TablesFound:    tables,
			BlocksFound:    blocks,
		}:
		default:
		}

		trimmed := strings.TrimSpace(currentLine)
		upper := strings.ToUpper(trimmed)

		// Detect USE dbname;
		if strings.HasPrefix(upper, "USE ") && strings.HasSuffix(upper, ";") {
			dbName := extractUseDB(trimmed)
			if dbName != "" {
				activeDB = dbName
				dbs[dbName] = true
			}
			offset += lineLen
			continue
		}

		// Detect comment block markers for routines/triggers/events
		if objType, dbName := parseBlockMarker(trimmed); objType != "" {
			if currentObj != nil {
				currentObj.EndByte = offset
				flushObj(stmt, currentObj)
			}
			currentObj = &CatalogObject{
				DumpFile:     dumpFile,
				DatabaseName: dbName,
				ObjectType:   objType,
				ObjectName:   objType,
				StartByte:    offset,
			}
			blocks++
			offset += lineLen
			continue
		}

		// Detect CREATE TABLE
		if objType, tableName := parseCreateTable(upper); objType != "" && activeDB != "" {
			if currentObj != nil {
				currentObj.EndByte = offset
				flushObj(stmt, currentObj)
			}
			currentObj = &CatalogObject{
				DumpFile:     dumpFile,
				DatabaseName: activeDB,
				ObjectType:   TypeTable,
				ObjectName:   tableName,
				StartByte:    offset,
			}
			tables++
			offset += lineLen
			continue
		}

		// Detect INSERT INTO (extends current table's range)
		if currentObj != nil && currentObj.ObjectType == TypeTable && strings.HasPrefix(upper, "INSERT INTO") {
			// INSERT INTO extends the current table's end_byte — no flush yet
			offset += lineLen
			continue
		}

		// Any other significant statement flushes the current block object
		if currentObj != nil && currentObj.ObjectType != TypeTable && len(trimmed) > 0 && !strings.HasPrefix(trimmed, "--") {
			currentObj.EndByte = offset
			flushObj(stmt, currentObj)
			currentObj = nil
		}

		offset += lineLen

		if err != nil {
			if err == io.EOF {
				break
			}
			tx.Rollback()
			return fmt.Errorf("read dump: %w", err)
		}
	}

	// Flush last object
	if currentObj != nil {
		currentObj.EndByte = offset
		flushObj(stmt, currentObj)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	// Final progress
	if onProgress != nil {
		onProgress(Progress{
			PercentBytes:   100,
			DatabasesFound: len(dbs),
			TablesFound:    tables,
			BlocksFound:    blocks,
		})
	}

	return nil
}

func flushObj(stmt *sql.Stmt, obj *CatalogObject) {
	stmt.Exec(obj.DumpFile, obj.DatabaseName, obj.ObjectType, obj.ObjectName, obj.StartByte, obj.EndByte)
}

// extractUseDB extracts the database name from "USE dbname;"
func extractUseDB(line string) string {
	s := strings.TrimSpace(line)
	upper := strings.ToUpper(s)
	if !strings.HasPrefix(upper, "USE ") {
		return ""
	}
	s = s[4:] // strip "USE " (4 chars)
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ";")
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "`\"'")
	return s
}

// parseCreateTable detects "CREATE TABLE `dbname`.`tablename"` or "CREATE TABLE tablename"
func parseCreateTable(upper string) (objType, name string) {
	if !strings.HasPrefix(upper, "CREATE TABLE") {
		return "", ""
	}
	// Skip CREATE TABLE IF NOT EXISTS
	s := upper
	s = strings.ReplaceAll(s, "CREATE TABLE IF NOT EXISTS", "CREATE TABLE")
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "CREATE TABLE")
	s = strings.TrimSpace(s)
	// Remove backticks
	s = strings.ReplaceAll(s, "`", "")
	s = strings.Trim(s, " \t")
	// Remove trailing (...
	if idx := strings.Index(s, "("); idx > 0 {
		s = s[:idx]
	}
	// Remove trailing semicolon
	s = strings.TrimSuffix(s, ";")
	// Remove schema prefix
	if idx := strings.Index(s, "."); idx > 0 {
		s = s[idx+1:]
	}
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"'")
	if s == "" {
		return "", ""
	}
	return TypeTable, strings.ToLower(s)
}

// parseBlockMarker detects "-- Dumped routines/triggers/events for database `dbname`"
// Returns the object type and the database name extracted from the comment.
func parseBlockMarker(line string) (objType, dbName string) {
	lower := strings.ToLower(line)
	var prefix string
	switch {
	case strings.Contains(lower, "-- dumped routines for database"):
		objType = TypeRoutines
		prefix = "-- dumped routines for database"
	case strings.Contains(lower, "-- dumped triggers for database"):
		objType = TypeTriggers
		prefix = "-- dumped triggers for database"
	case strings.Contains(lower, "-- dumped events for database"):
		objType = TypeEvents
		prefix = "-- dumped events for database"
	default:
		return "", ""
	}
	// Extract DB name after the prefix
	suffix := line[len(prefix):]
	suffix = strings.TrimSpace(suffix)
	suffix = strings.TrimSuffix(suffix, ";")
	suffix = strings.TrimSpace(suffix)
	suffix = strings.Trim(suffix, "`\"'")
	return objType, suffix
}
