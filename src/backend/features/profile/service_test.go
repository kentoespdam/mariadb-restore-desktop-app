package profile

import (
	"errors"
	"testing"

	crypto "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/catalog"
)

func mustSvc(t *testing.T) *Service {
	t.Helper()
	store, err := catalog.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i)
	}
	return New(store, key)
}

func TestCreateAndList(t *testing.T) {
	s := mustSvc(t)
	id, err := s.Create(Input{
		Name: "local", Host: "127.0.0.1", Port: 3306, User: "root", Password: "s3cret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("want ID")
	}
	views, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 {
		t.Fatalf("want 1 view, got %d", len(views))
	}
	v := views[0]
	if v.Name != "local" || v.Host != "127.0.0.1" || v.User != "root" || !v.HasPass {
		t.Fatalf("view = %+v", v)
	}
	if v.ID == "" {
		t.Fatal("view missing id")
	}
}

func TestUpdateKeepsPasswordWhenEmpty(t *testing.T) {
	s := mustSvc(t)
	id, _ := s.Create(Input{Name: "p", Host: "h", Port: 1, User: "u", Password: "old"})
	if err := s.Update(Input{ID: id, Name: "p", Host: "h2", Port: 2, User: "u", Password: ""}); err != nil {
		t.Fatal(err)
	}
	views, _ := s.List()
	if views[0].Host != "h2" || !views[0].HasPass {
		t.Fatalf("update lost password or host: %+v", views[0])
	}
}

func TestDelete(t *testing.T) {
	s := mustSvc(t)
	id, _ := s.Create(Input{Name: "p", Host: "h", Port: 1, User: "u", Password: "p"})
	if err := s.Delete(id); err != nil {
		t.Fatal(err)
	}
	views, _ := s.List()
	if len(views) != 0 {
		t.Fatalf("want 0, got %d", len(views))
	}
}

func TestDeleteUnknown(t *testing.T) {
	s := mustSvc(t)
	if err := s.Delete("nonexistent"); err == nil {
		t.Fatal("want error for unknown id")
	}
}

func TestCreateValidationNoName(t *testing.T) {
	s := mustSvc(t)
	_, err := s.Create(Input{Host: "h", Port: 1, User: "u", Password: "p"})
	if err == nil {
		t.Fatal("want error for empty name")
	}
}

// byID is a tiny helper, just smoke-test that it doesn't crash on
// an empty string (we use it as a no-op in this slice).
var _ = errors.New
