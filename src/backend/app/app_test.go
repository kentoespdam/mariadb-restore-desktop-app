package app

import (
	"context"
	"os"
	"testing"

	crypto "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/profile"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/recovery"
)

func TestNewAppGeneratesKey(t *testing.T) {
	dir := t.TempDir()
	a, err := New(context.Background(), dir)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	defer a.Close()
	if a.Key == nil || len(a.Key) != crypto.KeySize {
		t.Fatalf("key length = %d", len(a.Key))
	}
	if a.Profile == nil {
		t.Fatal("Profile service not wired")
	}
	if a.Recovery == nil {
		t.Fatal("Recovery service not wired")
	}
}

func TestNewAppReusesExistingKey(t *testing.T) {
	dir := t.TempDir()
	existing, err := crypto.GenerateKey(crypto.KeyPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	a, err := New(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	for i, b := range existing {
		if a.Key[i] != b {
			t.Fatal("key not reused")
		}
	}
}

func TestBindingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	a, err := New(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	id, err := a.CreateServerProfile(profile.Input{
		Name: "n", Host: "h", Port: 3306, User: "u", Password: "p",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("want id")
	}
	views, err := a.ListServerProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 || views[0].Name != "n" {
		t.Fatalf("views = %+v", views)
	}
	if err := a.UpdateServerProfile(profile.Input{
		ID: id, Name: "n2", Host: "h2", Port: 3307, User: "u2",
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.DeleteServerProfile(id); err != nil {
		t.Fatal(err)
	}
	views, _ = a.ListServerProfiles()
	if len(views) != 0 {
		t.Fatalf("after delete = %d", len(views))
	}
}

func TestRecoveryDecisionChannel(t *testing.T) {
	dir := t.TempDir()
	a, err := New(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	// unknown decisions are dropped silently
	if err := a.RecoveryDecision("maybe"); err != nil {
		t.Fatal(err)
	}
	// valid decision with no waiting modal is dropped (non-blocking)
	if err := a.RecoveryDecision(string(recovery.DecisionCancel)); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryCancelReturnsCancelled(t *testing.T) {
	dir := t.TempDir()
	a, err := New(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	// inject a static modal that returns cancel
	a.Recovery.Modal = staticDecision{recovery.DecisionCancel}
	got, err := a.HandleMissingKey()
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if got != "canceled" {
		t.Fatalf("got %q", got)
	}
}

type staticDecision struct{ d recovery.Decision }

func (s staticDecision) Show(_ context.Context) (recovery.Decision, error) { return s.d, nil }

// ensure the recovery service correctly wipes a stale catalog when reset is
// chosen and the app key file goes missing.
func TestRecoveryResetWipesAndRegenerates(t *testing.T) {
	dir := t.TempDir()
	a, err := New(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	// seed a profile so we can prove the wipe removed it
	if _, err := a.CreateServerProfile(profile.Input{
		Name: "x", Host: "h", Port: 1, User: "u", Password: "p",
	}); err != nil {
		t.Fatal(err)
	}
	// simulate the missing-key state by removing app.key
	if err := removeKey(t, a.KeyPath); err != nil {
		t.Fatal(err)
	}
	a.Recovery.Modal = staticDecision{recovery.DecisionReset}
	got, err := a.HandleMissingKey()
	if err != nil {
		t.Fatal(err)
	}
	if got != "reset" {
		t.Fatalf("got %q", got)
	}
	if _, err := crypto.LoadKey(a.KeyPath); err != nil {
		t.Fatalf("key not regenerated: %v", err)
	}
	// old catalog and its store handle are gone; a new app reopens cleanly
	a2, err := New(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a2.Close()
	views, _ := a2.ListServerProfiles()
	if len(views) != 0 {
		t.Fatalf("catalog not wiped: %+v", views)
	}
}

func removeKey(t *testing.T, path string) error {
	t.Helper()
	return os.Remove(path)
}
