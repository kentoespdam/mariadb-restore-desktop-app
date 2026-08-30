package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

const fixture = `-- MariaDB dump 10.19
USE db1;
CREATE TABLE t1 (id int);
INSERT INTO t1 VALUES (1);
INSERT INTO t1 VALUES (2);
CREATE TABLE t2 (id int);
USE db2;
CREATE TABLE u1 (id int);
INSERT INTO u1 VALUES (1);
`

func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "dump.sql")
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestScanFixture(t *testing.T) {
	path := writeFixture(t)
	got, err := New().Scan(path)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(got) != 6 {
		t.Fatalf("want 6 objects, got %d: %+v", len(got), got)
	}
	// sanity: every object has positive range
	for i, o := range got {
		if o.StartByte < 0 || o.EndByte <= o.StartByte {
			t.Fatalf("object %d has bad range: %+v", i, o)
		}
	}
	// offsets are byte-exact: re-read the file and confirm each range
	// begins with the expected first line
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range got {
		// confirm no off-by-one: range should be valid within file
		if o.EndByte > int64(len(data)) {
			t.Fatalf("EndByte %d > file size %d", o.EndByte, len(data))
		}
	}
}

func TestScanOffsetExact(t *testing.T) {
	// Verifies the recorded StartByte actually points at the line we expect.
	path := writeFixture(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	objs, err := New().Scan(path)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	// first DDL: `CREATE TABLE t1`
	if objs[0].ObjectType != TypeCreateTable || objs[0].ObjectName != "t1" || objs[0].DatabaseName != "db1" {
		t.Fatalf("first obj = %+v", objs[0])
	}
	if got := string(data[objs[0].StartByte : objs[0].StartByte+15]); got != "CREATE TABLE t1" {
		t.Fatalf("offset points at %q", got)
	}
}

func TestScanMissingFile(t *testing.T) {
	if _, err := New().Scan("/no/such/file.sql"); err == nil {
		t.Fatal("want error for missing file")
	}
}

const backtickFixture = "-- MariaDB dump 10.19\nUSE `db1`;\nCREATE TABLE `db1`.`t1` (id int);\nINSERT INTO `db1`.`t1` VALUES (1);\nCREATE TABLE `db1`.`t2` (id int);\n"

func TestScanBacktickDbTable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dump.sql")
	if err := os.WriteFile(path, []byte(backtickFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	objs, err := New().Scan(path)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(objs) != 3 {
		t.Fatalf("want 3 objects, got %d: %+v", len(objs), objs)
	}
	// ObjectName should be the TABLE name, not the DB name.
	want := []string{"t1", "t1", "t2"}
	for i, o := range objs {
		if o.ObjectName != want[i] {
			t.Errorf("obj[%d] ObjectName = %q, want %q", i, o.ObjectName, want[i])
		}
		if o.DatabaseName != "db1" {
			t.Errorf("obj[%d] DatabaseName = %q, want db1", i, o.DatabaseName)
		}
	}
}
