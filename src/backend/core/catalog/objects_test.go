package catalog

import (
	"testing"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/scanner"
)

func TestReplaceAndListObjects(t *testing.T) {
	s := mustStore(t)

	dumpPath := "/tmp/dump-A.sql"
	objs := []scanner.Object{
		{StartByte: 0, EndByte: 100, ObjectType: scanner.TypeCreateTable, ObjectName: "t1", DatabaseName: "shop"},
		{StartByte: 100, EndByte: 200, ObjectType: scanner.TypeInsert, ObjectName: "t1", DatabaseName: "shop"},
		{StartByte: 200, EndByte: 300, ObjectType: scanner.TypeCreateTable, ObjectName: "t2", DatabaseName: "auth"},
	}
	if err := s.ReplaceObjectsForDump(dumpPath, objs); err != nil {
		t.Fatalf("replace: %v", err)
	}

	rows, err := s.ListObjectsForDump(dumpPath)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("want 3 rows, got %d", len(rows))
	}
	// Confirm ordering by (database_name, object_name, id): auth.t2
	// comes first.
	if rows[0].ObjectName != "t2" || rows[0].DatabaseName != "auth" {
		t.Errorf("ordering wrong: %+v", rows[0])
	}

	// Re-replace should wipe + insert, not duplicate.
	if err := s.ReplaceObjectsForDump(dumpPath, objs[:1]); err != nil {
		t.Fatalf("replace 2: %v", err)
	}
	rows, err = s.ListObjectsForDump(dumpPath)
	if err != nil {
		t.Fatalf("list 2: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("replace should have wiped: got %d", len(rows))
	}
}

func TestDeleteObjectsForDump(t *testing.T) {
	s := mustStore(t)
	dumpPath := "/tmp/dump-B.sql"
	if err := s.ReplaceObjectsForDump(dumpPath, []scanner.Object{
		{StartByte: 0, EndByte: 50, ObjectType: scanner.TypeCreateTable, ObjectName: "x", DatabaseName: "d"},
	}); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if err := s.DeleteObjectsForDump(dumpPath); err != nil {
		t.Fatalf("delete: %v", err)
	}
	rows, _ := s.ListObjectsForDump(dumpPath)
	if len(rows) != 0 {
		t.Errorf("delete failed: %d rows remain", len(rows))
	}
}
