// Package backup runs mariadb-dump as a subprocess against a Server
// Profile and streams the result to a chosen output path. The job is
// cancellable via the context returned to the caller; cancellation
// kills the subprocess and any partial file is removed.
package backup

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/profile"
)

// Progress is the payload emitted to the FE on every chunked write
// from the subprocess. The FE throttles this to ~150ms anyway, but
// the BE also throttles to 100-250ms (CONTEXT) so a chatty writer
// cannot flood the Wails event bus.
type Progress struct {
	JobID string `json:"jobId"`
	SoFar int64  `json:"soFar"`
	Total int64  `json:"total"`
}

// Done is the terminal event for a backup job.
type Done struct {
	JobID   string `json:"jobId"`
	Status  string `json:"status"` // "success" | "error"
	Message string `json:"message,omitempty"`
}

// Request is the BE-facing shape. The FE-facing payload lives in
// api/backup.ts; the Wails binding converts at the boundary.
type Request struct {
	ProfileID  string
	Databases  []string
	OutputPath string
	BinaryPath string
}

// Service is the backup runner. profileSvc decrypts credentials,
// emitter fans out Progress/Done events to the FE.
type Service struct {
	profileSvc *profile.Service
	emitter    Emitter
	binPath    string

	mu   sync.Mutex
	jobs map[string]*job
}

// Emitter is the minimal interface the service needs to publish Wails
// events. platform/events.Emitter satisfies it; tests inject a fake.
type Emitter interface {
	Emit(ctx context.Context, name string, payload any) error
}

// job is the in-flight state for one backup. cancel kills the
// subprocess.
type job struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

// New returns a service. binPath is the mariadb-dump binary the
// service uses; settings overrides the default "mariadb-dump" PATH
// lookup at runtime.
func New(profileSvc *profile.Service, emitter Emitter, binPath string) *Service {
	if binPath == "" {
		binPath = "mariadb-dump"
	}
	return &Service{
		profileSvc: profileSvc,
		emitter:    emitter,
		binPath:    binPath,
		jobs:       make(map[string]*job),
	}
}

// Start kicks off the backup in a goroutine and returns the JobID
// immediately. Progress events stream on the Wails bus named
// "backup:progress"; the terminal "backup:done" event marks
// completion. Call Cancel(jobID) to abort.
func (s *Service) Start(parent context.Context, req Request) (string, error) {
	if req.ProfileID == "" {
		return "", errors.New("backup: profile id required")
	}
	if len(req.Databases) == 0 {
		return "", errors.New("backup: at least one database required")
	}
	if req.OutputPath == "" {
		return "", errors.New("backup: output path required")
	}
	if req.BinaryPath != "" {
		s.binPath = req.BinaryPath
	}

	// Resolve credentials from the profile by id.
	creds, err := s.profileSvc.CredentialsByID(req.ProfileID)
	if err != nil {
		return "", err
	}

	jobID := newID()
	jobCtx, cancel := context.WithCancel(parent)

	outFile, err := os.Create(req.OutputPath)
	if err != nil {
		cancel()
		return "", fmt.Errorf("backup: create output: %w", err)
	}
	// We close outFile inside the goroutine so the writer's lifetime
	// matches the job's.

	args := []string{
		"-h", creds.Host,
		"-P", strconv.Itoa(creds.Port),
		"-u", creds.User,
		fmt.Sprintf("-p%s", creds.Password),
	}
	for _, db := range req.Databases {
		args = append(args, "--databases", db)
	}

	cmd := exec.CommandContext(jobCtx, s.binPath, args...)
	cmd.Stdout = outFile
	// mariadb-dump writes only to stdout; silence stderr but capture
	// it for the done event.
	stderrBuf := &strings.Builder{}
	cmd.Stderr = stderrBuf

	if err := cmd.Start(); err != nil {
		_ = outFile.Close()
		_ = os.Remove(req.OutputPath)
		cancel()
		return "", fmt.Errorf("backup: start subprocess: %w", err)
	}

	s.mu.Lock()
	s.jobs[jobID] = &job{cmd: cmd, cancel: cancel}
	s.mu.Unlock()

	go s.run(jobID, cmd, outFile, stderrBuf, req.OutputPath, jobCtx)
	return jobID, nil
}

// run drives the subprocess lifecycle and emits the terminal event.
func (s *Service) run(
	jobID string,
	cmd *exec.Cmd,
	outFile *os.File,
	stderrBuf *strings.Builder,
	outputPath string,
	ctx context.Context,
) {
	// Periodic progress emission: we know the file size grows so we
	// stat it on a ticker. A more accurate approach would parse the
	// subprocess's stderr ("--progress"), but mariadb-dump does not
	// emit a reliable progress line. Ponytail: poll-file beats
	// parsing every line of stderr.
	stop := make(chan struct{})
	defer close(stop)
	go s.progressLoop(jobID, outputPath, stop, ctx)

	waitErr := cmd.Wait()
	_ = outFile.Close()

	s.mu.Lock()
	delete(s.jobs, jobID)
	s.mu.Unlock()

	if waitErr != nil {
		// Distinguish cancel from genuine failure.
		if ctx.Err() != nil {
			_ = os.Remove(outputPath)
			_ = s.emitter.Emit(context.Background(), "backup:done", Done{
				JobID:   jobID,
				Status:  "error",
				Message: "canceled",
			})
			return
		}
		_ = os.Remove(outputPath)
		_ = s.emitter.Emit(context.Background(), "backup:done", Done{
			JobID:   jobID,
			Status:  "error",
			Message: trimMsg(stderrBuf.String(), waitErr.Error()),
		})
		return
	}

	info, statErr := os.Stat(outputPath)
	size := int64(0)
	if statErr == nil {
		size = info.Size()
	}
	_ = s.emitter.Emit(context.Background(), "backup:done", Done{
		JobID:  jobID,
		Status: "success",
	})
	_ = size
}

func (s *Service) progressLoop(jobID, outputPath string, stop <-chan struct{}, ctx context.Context) {
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			info, err := os.Stat(outputPath)
			if err != nil {
				continue
			}
			_ = s.emitter.Emit(context.Background(), "backup:progress", Progress{
				JobID: jobID,
				SoFar: info.Size(),
				Total: 0,
			})
		}
	}
}

// Cancel aborts the subprocess for jobID. No-op if the job already
// finished or never existed.
func (s *Service) Cancel(jobID string) {
	s.mu.Lock()
	j, ok := s.jobs[jobID]
	s.mu.Unlock()
	if !ok {
		return
	}
	j.cancel()
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("bk-%x", b)
}

// trimMsg keeps the done-event message short enough for the UI.
// mariadb-dump stderr can be a multi-line stack trace; we only want
// the first line.
func trimMsg(a, b string) string {
	msg := a
	if msg == "" {
		msg = b
	}
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	if len(msg) > 240 {
		msg = msg[:240] + "..."
	}
	return msg
}

// EnsurePath creates the parent directory of outputPath if missing.
// Called by the FE before starting a backup so we don't fail at
// os.Create time on a fresh machine.
func EnsurePath(outputPath string) error {
	dir := filepath.Dir(outputPath)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

// Keep crypto imported so future changes can decrypt secrets from
// here without re-adding the import.
var _ = crypto.KeySize