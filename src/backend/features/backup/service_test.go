package backup

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/catalog"
	crypto "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/profile"
)

// rec is a thread-safe emitter that captures every progress/done
// event the backup service emits. Used by the regression test to
// assert the UI receives at least one progress tick per job — the
// fix for the "backup looks stuck" bug where sub-ticker-interval
// dumps produced zero events.
type rec struct {
	mu   sync.Mutex
	evts map[string][]any
}

func newRec() *rec { return &rec{evts: map[string][]any{}} }

func (r *rec) Emit(_ context.Context, name string, payload any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.evts[name] = append(r.evts[name], payload)
	return nil
}

func (r *rec) snapshot(name string) []any {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]any(nil), r.evts[name]...)
}

// liveDBConfig skips the test when docker mariadb-test isn't
// reachable; the regression only makes sense against a real server.
func liveDBConfig(t *testing.T) (host string, port int, user, pass string) {
	t.Helper()
	if os.Getenv("SKIP_LIVE_DB") != "" {
		t.Skip("SKIP_LIVE_DB set")
	}
	host = os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port = 3307
	user = "root"
	pass = os.Getenv("TEST_DB_PASS")
	if pass == "" {
		pass = "testpass"
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 500*time.Millisecond)
	if err != nil {
		t.Skipf("live mariadb-test not reachable on %s:%d (%v); skipping live test", host, port, err)
	}
	_ = conn.Close()
	return host, port, user, pass
}

// TestStartEmitsAtLeastOneProgressEvent is the regression for
// "backup looks stuck at Scanning…". A successful backup must emit
// at least one backup:progress event before backup:done fires,
// regardless of how fast the subprocess finishes. Without this,
// jobs shorter than the ticker interval (100-150ms) silently emit
// nothing and the UI cannot distinguish a fast job from a stuck one.
func TestStartEmitsAtLeastOneProgressEvent(t *testing.T) {
	host, port, user, pass := liveDBConfig(t)

	dir := t.TempDir()
	catPath := filepath.Join(dir, "catalog.sqlite")
	cat, err := catalog.Open(catPath)
	if err != nil {
		t.Fatal(err)
	}
	defer cat.Close()

	keyPath := crypto.KeyPath(dir)
	if _, err := crypto.GenerateKey(keyPath); err != nil {
		t.Fatal(err)
	}
	key, err := crypto.LoadKey(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	profSvc := profile.New(cat, key)
	profID, err := profSvc.Create(profile.Input{
		Name: "docker-root", Host: host, Port: port, User: user, Password: pass, SSLMode: "preferred",
	})
	if err != nil {
		t.Skipf("live mariadb-test not reachable on %s:%d (%v); skipping live regression", host, port, err)
	}

	rec := newRec()
	svc := New(profSvc, rec, "mariadb-dump")
	dumpPath := filepath.Join(dir, "backup.sql")

	jobID, err := svc.Start(context.Background(), Request{
		ProfileID:  profID,
		Databases:  []string{"shop"},
		OutputPath: dumpPath,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if len(rec.snapshot("backup:done")) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	done := rec.snapshot("backup:done")
	if len(done) == 0 {
		t.Fatalf("backup:done never fired")
	}
	if d, ok := done[0].(Done); !ok || d.Status != "success" {
		t.Fatalf("backup:done payload = %+v; want status=success", done[0])
	}

	progress := rec.snapshot("backup:progress")
	if len(progress) == 0 {
		t.Fatalf("backup produced 0 progress events; UI will show 'Scanning…' forever. "+
			"This is the bug: sub-ticker-interval dumps emit no events.")
	}

	info, err := os.Stat(dumpPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() < 100 {
		t.Fatalf("dump suspiciously small: %d bytes", info.Size())
	}

	// Every emitted progress event must carry a jobId that matches
	// the job. A dropped or empty jobId would silently desync the FE.
	for i, p := range progress {
		pp, ok := p.(Progress)
		if !ok {
			t.Fatalf("progress[%d] wrong type %T", i, p)
		}
		if pp.JobID != jobID {
			t.Fatalf("progress[%d] jobId=%q want %q", i, pp.JobID, jobID)
		}
	}
}