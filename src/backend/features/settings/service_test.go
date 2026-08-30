package settings

import (
	"path/filepath"
	"testing"
)

func TestDefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	s := New(dir, 256)
	got, err := s.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.MariadbPath != DefaultMariadbPath || got.MariadbDumpPath != DefaultMariadbDumpPath {
		t.Errorf("defaults wrong: %+v", got)
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	s := New(dir, 256)
	in := Input{MariadbPath: "/opt/mariadb/bin/mariadb", MariadbDumpPath: "/opt/mariadb/bin/mariadb-dump"}
	if err := s.Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := filepath.Glob(filepath.Join(dir, "settings.json")); err != nil {
		t.Fatalf("glob: %v", err)
	}
	// Re-open with a fresh service so we know it's the disk file.
	s2 := New(dir, 256)
	got, err := s2.Get()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.MariadbPath != in.MariadbPath || got.MariadbDumpPath != in.MariadbDumpPath {
		t.Errorf("roundtrip wrong: %+v", got)
	}
}

func TestBuildView(t *testing.T) {
	dir := t.TempDir()
	s := New(dir, 256)
	v, err := s.BuildView("/x/catalog.sqlite", "/x/app.key")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if v.ExeDir != dir || v.CatalogPath != "/x/catalog.sqlite" || v.AppKeyPath != "/x/app.key" || v.KeyBits != 256 {
		t.Errorf("view wrong: %+v", v)
	}
}
