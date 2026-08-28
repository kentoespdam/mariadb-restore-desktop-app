package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"mariadb-restore-desktop-app/internal/catalog"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := catalog.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func writeDump(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write dump: %v", err)
	}
	return path
}

const testDumpMultiDB = `-- MariaDB dump
-- Server version	10.11.6-MariaDB

USE ` + "`testdb`" + `;

CREATE TABLE ` + "`users`" + ` (
  id INT PRIMARY KEY,
  name VARCHAR(100)
) ENGINE=InnoDB;

INSERT INTO ` + "`users`" + ` VALUES (1, 'Alice');
INSERT INTO ` + "`users`" + ` VALUES (2, 'Bob');

CREATE TABLE ` + "`orders`" + ` (
  id INT PRIMARY KEY,
  user_id INT
) ENGINE=InnoDB;

INSERT INTO ` + "`orders`" + ` VALUES (10, 1);

USE ` + "`analytics`" + `;

CREATE TABLE ` + "`events`" + ` (
  id INT PRIMARY KEY,
  event_name VARCHAR(50)
) ENGINE=InnoDB;

INSERT INTO ` + "`events`" + ` VALUES (100, 'click');

-- Dumped routines for database ` + "`testdb`" + `;
CREATE DEFINER=... PROCEDURE sp_test() BEGIN SELECT 1; END;

-- Dumped triggers for database ` + "`testdb`" + `;
CREATE TRIGGER trg_test BEFORE INSERT ON users FOR EACH ROW BEGIN SET NEW.name = NEW.name; END;

-- Dumped events for database ` + "`analytics`" + `;
CREATE EVENT evt_test ON SCHEDULE EVERY 1 HOUR DO BEGIN SELECT 1; END;
`

func TestScanMultiDB(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	dumpPath := writeDump(t, dir, "dump.sql", testDumpMultiDB)

	var lastProgress Progress
	err := Scan(context.Background(), db, dumpPath, func(p Progress) {
		lastProgress = p
	})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	// Verify progress reached 100%
	if lastProgress.PercentBytes != 100 {
		t.Fatalf("percent = %f, want 100", lastProgress.PercentBytes)
	}
	if lastProgress.DatabasesFound != 2 {
		t.Fatalf("databases = %d, want 2", lastProgress.DatabasesFound)
	}
	if lastProgress.TablesFound != 3 {
		t.Fatalf("tables = %d, want 3", lastProgress.TablesFound)
	}
	if lastProgress.BlocksFound != 3 {
		t.Fatalf("blocks = %d, want 3", lastProgress.BlocksFound)
	}

	// Verify catalog objects
	rows, err := db.Query("SELECT database_name, object_type, object_name, start_byte, end_byte FROM catalog_objects WHERE dump_file=? ORDER BY start_byte", dumpPath)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var objects []CatalogObject
	for rows.Next() {
		var o CatalogObject
		rows.Scan(&o.DatabaseName, &o.ObjectType, &o.ObjectName, &o.StartByte, &o.EndByte)
		objects = append(objects, o)
	}

	// Expect: users, orders (testdb), events (analytics), routines, triggers, events_block
	if len(objects) < 5 {
		t.Fatalf("expected at least 5 objects, got %d", len(objects))
	}

	// Check users table is in testdb
	found := false
	for _, o := range objects {
		if o.ObjectName == "users" && o.DatabaseName == "testdb" && o.ObjectType == TypeTable {
			found = true
			if o.EndByte <= o.StartByte {
				t.Fatalf("users: end_byte (%d) <= start_byte (%d)", o.EndByte, o.StartByte)
			}
		}
	}
	if !found {
		t.Fatal("users table not found in testdb")
	}

	// Check events table is in analytics
	found = false
	for _, o := range objects {
		if o.ObjectName == "events" && o.DatabaseName == "analytics" && o.ObjectType == TypeTable {
			found = true
		}
	}
	if !found {
		t.Fatal("events table not found in analytics")
	}

	// Check routines block
	found = false
	for _, o := range objects {
		if o.ObjectType == TypeRoutines && o.DatabaseName == "testdb" {
			found = true
		}
	}
	if !found {
		t.Fatal("routines block not found for testdb")
	}
}

func TestScanCancel(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()

	// Create a large-ish dump to give time to cancel
	content := "USE `db1`;\n"
	for i := 0; i < 1000; i++ {
		content += fmt.Sprintf("CREATE TABLE t%d (id INT);\n", i)
	}
	dumpPath := writeDump(t, dir, "large.sql", content)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := Scan(ctx, db, dumpPath, nil)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestScanSingleSlot(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	dumpPath := writeDump(t, dir, "dump.sql", testDumpMultiDB)

	// Scan twice
	Scan(context.Background(), db, dumpPath, nil)
	Scan(context.Background(), db, dumpPath, nil)

	// Should only have objects from the last scan (single-slot)
	var count int
	db.QueryRow("SELECT COUNT(*) FROM catalog_objects WHERE dump_file=?", dumpPath).Scan(&count)

	// Should be 6 (3 tables + 3 blocks), not 12
	if count != 6 {
		t.Fatalf("expected 6 objects after rescan, got %d", count)
	}
}

func TestScanEmptyFile(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	dumpPath := writeDump(t, dir, "empty.sql", "")

	err := Scan(context.Background(), db, dumpPath, nil)
	if err != nil {
		t.Fatalf("Scan empty: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM catalog_objects WHERE dump_file=?", dumpPath).Scan(&count)
	if count != 0 {
		t.Fatalf("expected 0 objects for empty file, got %d", count)
	}
}

func TestScanMissingFile(t *testing.T) {
	db := setupTestDB(t)
	err := Scan(context.Background(), db, "/nonexistent/dump.sql", nil)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestScanByteRangesValid(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	dumpPath := writeDump(t, dir, "dump.sql", testDumpMultiDB)

	Scan(context.Background(), db, dumpPath, nil)

	// Verify all byte ranges are valid (start < end, end <= file size)
	info, _ := os.Stat(dumpPath)
	fileSize := info.Size()

	rows, _ := db.Query("SELECT object_name, start_byte, end_byte FROM catalog_objects WHERE dump_file=?", dumpPath)
	defer rows.Close()

	for rows.Next() {
		var name string
		var start, end int64
		rows.Scan(&name, &start, &end)
		if start >= end {
			t.Fatalf("%s: start (%d) >= end (%d)", name, start, end)
		}
		if end > fileSize {
			t.Fatalf("%s: end (%d) > file size (%d)", name, end, fileSize)
		}
	}
}

func TestExtractUseDB(t *testing.T) {
	tests := []struct {
		line, want string
	}{
		{"USE `testdb`;", "testdb"},
		{"use mydb;", "mydb"},
		{"USE \"somedb\";", "somedb"},
		{"  USE db1;  ", "db1"},
	}
	for _, tt := range tests {
		got := extractUseDB(tt.line)
		if got != tt.want {
			t.Errorf("extractUseDB(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestParseCreateTable(t *testing.T) {
	tests := []struct {
		line, wantType, wantName string
	}{
		{"CREATE TABLE `users` (", TypeTable, "users"},
		{"CREATE TABLE IF NOT EXISTS `orders` (", TypeTable, "orders"},
		{"CREATE TABLE myschema.users (", TypeTable, "users"},
		{"CREATE TABLE foo;", TypeTable, "foo"},
		{"SELECT 1;", "", ""},
	}
	for _, tt := range tests {
		gotType, gotName := parseCreateTable(tt.line)
		if gotType != tt.wantType || gotName != tt.wantName {
			t.Errorf("parseCreateTable(%q) = (%q, %q), want (%q, %q)", tt.line, gotType, gotName, tt.wantType, tt.wantName)
		}
	}
}

const testDumpSingleDB = `-- MariaDB dump 10.19  Distrib 10.11.6-MariaDB, for debian-linux-gnu (x86_64)
--
-- Host: localhost    Database: my_single_db
-- ------------------------------------------------------
-- Server version	10.11.6-MariaDB

DROP TABLE IF EXISTS ` + "`customers`" + `;
CREATE TABLE ` + "`customers`" + ` (
  ` + "`id`" + ` int(11) NOT NULL,
  ` + "`name`" + ` varchar(50) DEFAULT NULL
) ENGINE=InnoDB;

INSERT INTO ` + "`customers`" + ` VALUES (1,'John Doe');
`

func TestScanSingleDBWithoutUseStatement(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	dumpPath := writeDump(t, dir, "singledb.sql", testDumpSingleDB)

	var lastProgress Progress
	err := Scan(context.Background(), db, dumpPath, func(p Progress) {
		lastProgress = p
	})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if lastProgress.TablesFound != 1 {
		t.Fatalf("tables found = %d, want 1", lastProgress.TablesFound)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM catalog_objects WHERE dump_file=? AND object_type='table'", dumpPath).Scan(&count)
	if count != 1 {
		t.Fatalf("catalog table count = %d, want 1", count)
	}
}

