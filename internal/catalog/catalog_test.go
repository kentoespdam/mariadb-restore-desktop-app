package catalog

import (
	"testing"
)

func TestOpenAndSchema(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	tables := []string{"server_profiles", "catalog_objects", "app_last_op"}
	for _, tbl := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl).Scan(&name)
		if err != nil {
			t.Fatalf("table %q not found: %v", tbl, err)
		}
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if Exists(dir) {
		t.Fatal("Exists returned true for empty dir")
	}
	db, _ := Open(dir)
	db.Close()
	if !Exists(dir) {
		t.Fatal("Exists returned false after Open")
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	db, _ := Open(dir)
	db.Close()
	if err := Remove(dir); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if Exists(dir) {
		t.Fatal("Exists returned true after Remove")
	}
}

func TestServerProfileCRUD(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Insert
	res, err := db.Exec(
		"INSERT INTO server_profiles (name, host, port, username) VALUES (?, ?, ?, ?)",
		"test-server", "127.0.0.1", 3306, "root",
	)
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	id, _ := res.LastInsertId()

	// Read
	var name, host, username string
	var port int
	err = db.QueryRow("SELECT name, host, port, username FROM server_profiles WHERE id=?", id).
		Scan(&name, &host, &port, &username)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if name != "test-server" || host != "127.0.0.1" || port != 3306 || username != "root" {
		t.Fatalf("unexpected values: %s %s %d %s", name, host, port, username)
	}

	// Update
	_, err = db.Exec("UPDATE server_profiles SET host=? WHERE id=?", "10.0.0.1", id)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	// Delete
	_, err = db.Exec("DELETE FROM server_profiles WHERE id=?", id)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestCatalogObjectInsert(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(
		`INSERT INTO catalog_objects (dump_file, database_name, object_type, object_name, start_byte, end_byte)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"/tmp/dump.sql", "mydb", "TABLE", "users", 0, 1024,
	)
	if err != nil {
		t.Fatalf("Insert catalog_object: %v", err)
	}

	var objName string
	var startByte, endByte int
	err = db.QueryRow(
		"SELECT object_name, start_byte, end_byte FROM catalog_objects WHERE database_name=?",
		"mydb",
	).Scan(&objName, &startByte, &endByte)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if objName != "users" || startByte != 0 || endByte != 1024 {
		t.Fatalf("unexpected: %s %d %d", objName, startByte, endByte)
	}
}

func TestAppLastOp(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT OR REPLACE INTO app_last_op (id, operation, dump_file) VALUES (1, ?, ?)",
		"restore", "/tmp/dump.sql",
	)
	if err != nil {
		t.Fatalf("Insert last_op: %v", err)
	}

	// Upsert again
	_, err = db.Exec(
		"INSERT OR REPLACE INTO app_last_op (id, operation, dump_file) VALUES (1, ?, ?)",
		"backup", "/tmp/backup.sql",
	)
	if err != nil {
		t.Fatalf("Upsert last_op: %v", err)
	}

	var op, file string
	err = db.QueryRow("SELECT operation, dump_file FROM app_last_op WHERE id=1").Scan(&op, &file)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if op != "backup" || file != "/tmp/backup.sql" {
		t.Fatalf("unexpected: %s %s", op, file)
	}
}
