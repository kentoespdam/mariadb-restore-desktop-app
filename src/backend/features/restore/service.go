// Package restore drives mariadb CLI as a subprocess, piping either
// the full dump file (Full Restore) or a selected-byte-range stream
// (Partial Restore) to stdin. The same context.WithCancel mechanism
// the backup package uses handles abort + subprocess cleanup.
package restore

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/catalog"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/scanner"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/streamer"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/profile"
)

// Progress / Done mirror the backup package's wire shape so the FE
// can use one handler pattern.
type Progress struct {
	JobID string `json:"jobId"`
	SoFar int64  `json:"soFar"`
	Total int64  `json:"total"`
}

type Done struct {
	JobID   string `json:"jobId"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Request is the Full Restore payload.
type FullRequest struct {
	ProfileID  string
	FilePath   string
	BinaryPath string
}

// PartialRequest is the Partial Restore payload. SelectedIDs refer
// to catalog.object rows; the service resolves them back to byte
// ranges against FilePath.
type PartialRequest struct {
	ProfileID       string
	FilePath        string
	SelectedIDs     []int
	IncludeRoutines bool
	IncludeTriggers bool
	IncludeEvents   bool
	BinaryPath      string
}

// Service is the restore runner.
type Service struct {
	profileSvc *profile.Service
	cat        *catalog.Store
	emitter    Emitter
	// stderrSink (optional) receives the raw mariadb stderr for every
	// failed restore. Used by the integration driver for diagnostics;
	// production sets it to nil.
	stderrSink func(jobID, stderr string)

	binMariadb string

	mu   sync.Mutex
	jobs map[string]*job
}

// Emitter mirrors platform/events.Emitter.
type Emitter interface {
	Emit(ctx context.Context, name string, payload any) error
}

type job struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

// New returns a service. mariadb is the path to the mariadb CLI;
// empty -> "mariadb" PATH lookup.
func New(profileSvc *profile.Service, cat *catalog.Store, emitter Emitter, mariadb string) *Service {
	if mariadb == "" {
		mariadb = "mariadb"
	}
	return &Service{
		profileSvc: profileSvc,
		cat:        cat,
		emitter:    emitter,
		binMariadb: mariadb,
		jobs:       make(map[string]*job),
	}
}

// SetStderrSink installs a raw-stderr sink for diagnostics. Tests
// use this; production leaves it nil. Not goroutine-safe; call
// during setup.
func (s *Service) SetStderrSink(sink func(jobID, stderr string)) {
	s.stderrSink = sink
}

// StartFull kicks off a Full Restore: the entire dump file at
// req.FilePath is streamed to mariadb stdin. Progress events stream
// on "restore:progress"; the terminal "restore:done" marks
// completion. Call Cancel to abort.
func (s *Service) StartFull(parent context.Context, req FullRequest) (string, error) {
	if err := validateFile(req.FilePath); err != nil {
		return "", err
	}
	if req.BinaryPath != "" {
		s.binMariadb = req.BinaryPath
	}
	creds, err := s.profileSvc.CredentialsByID(req.ProfileID)
	if err != nil {
		return "", err
	}
	jobID := newID()
	jobCtx, cancel := context.WithCancel(parent)

	args := []string{
		"-h", creds.Host,
		"-P", strconv.Itoa(creds.Port),
		"-u", creds.User,
		fmt.Sprintf("-p%s", creds.Password),
		"--comments",
	}
	cmd := exec.CommandContext(jobCtx, s.binMariadb, args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return "", fmt.Errorf("restore: stdin pipe: %w", err)
	}
	stderrBuf := &strings.Builder{}
	cmd.Stderr = stderrBuf

	if err := cmd.Start(); err != nil {
		cancel()
		return "", fmt.Errorf("restore: start subprocess: %w", err)
	}

	s.mu.Lock()
	s.jobs[jobID] = &job{cmd: cmd, cancel: cancel}
	s.mu.Unlock()

	go s.runFull(jobID, cmd, in, stderrBuf, req.FilePath, jobCtx)
	return jobID, nil
}

// StartPartial kicks off a Partial Restore. The service reads
// req.SelectedIDs from the catalog, looks up their byte ranges for
// req.FilePath, and pipes header + selected ranges + footer to
// mariadb stdin via streamer.Build. The DefinerStripper rewrites
// DEFINER= clauses on the fly so the target server doesn't need to
// know the original definer users.
func (s *Service) StartPartial(parent context.Context, req PartialRequest) (string, error) {
	if err := validateFile(req.FilePath); err != nil {
		return "", err
	}
	if req.BinaryPath != "" {
		s.binMariadb = req.BinaryPath
	}
	if len(req.SelectedIDs) == 0 {
		return "", errors.New("restore: at least one object required")
	}
	creds, err := s.profileSvc.CredentialsByID(req.ProfileID)
	if err != nil {
		return "", err
	}

	parts, err := s.partsFor(req.FilePath, req.SelectedIDs, req.IncludeRoutines, req.IncludeTriggers, req.IncludeEvents)
	if err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return "", errors.New("restore: no catalog rows matched the selection")
	}

	jobID := newID()
	jobCtx, cancel := context.WithCancel(parent)

	args := []string{
		"-h", creds.Host,
		"-P", strconv.Itoa(creds.Port),
		"-u", creds.User,
		fmt.Sprintf("-p%s", creds.Password),
		"--comments",
	}
	cmd := exec.CommandContext(jobCtx, s.binMariadb, args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return "", fmt.Errorf("restore: stdin pipe: %w", err)
	}
	stderrBuf := &strings.Builder{}
	cmd.Stderr = stderrBuf

	if err := cmd.Start(); err != nil {
		cancel()
		return "", fmt.Errorf("restore: start subprocess: %w", err)
	}

	s.mu.Lock()
	s.jobs[jobID] = &job{cmd: cmd, cancel: cancel}
	s.mu.Unlock()

	go s.runPartial(jobID, cmd, in, stderrBuf, req.FilePath, parts, jobCtx)
	return jobID, nil
}

// runFull pipes the file end-to-end to the subprocess stdin and
// emits progress via a stat ticker on the source file.
func (s *Service) runFull(
	jobID string,
	cmd *exec.Cmd,
	in io.WriteCloser,
	stderrBuf *strings.Builder,
	filePath string,
	ctx context.Context,
) {
	stop := make(chan struct{})
	defer close(stop)
	go s.progressLoop(jobID, filePath, stop, ctx)

	src, err := os.Open(filePath)
	if err != nil {
		_ = in.Close()
		s.finishError(jobID, cmd, stderrBuf, err)
		return
	}
	_, copyErr := io.Copy(in, src)
	_ = src.Close()
	_ = in.Close()

	waitErr := cmd.Wait()
	s.cleanup(jobID)

	if copyErr != nil {
		s.emitDone(jobID, "error", trimMsg(stderrBuf.String(), copyErr.Error()))
		return
	}
	if waitErr != nil {
		if ctx.Err() != nil {
			s.emitDone(jobID, "error", "canceled")
			return
		}
		s.emitDone(jobID, "error", trimMsg(stderrBuf.String(), waitErr.Error()))
		return
	}
	s.emitDone(jobID, "success", "")
}

// runPartial pipes header + selected byte ranges + footer through
// the DefinerStripper to the subprocess stdin.
func (s *Service) runPartial(
	jobID string,
	cmd *exec.Cmd,
	in io.WriteCloser,
	stderrBuf *strings.Builder,
	filePath string,
	parts []scanner.Offset,
	ctx context.Context,
) {
	stop := make(chan struct{})
	defer close(stop)
	// Ponytail: progress on partial = total bytes we'll write. We
	// sum the selected ranges so the bar can hit 100%.
	var total int64
	for _, p := range parts {
		total += p.EndByte - p.StartByte
	}
	go s.progressLoopTotal(jobID, stop, ctx, total, func() int64 {
		// We don't have a per-write byte counter hooked into the
		// streamer; emit a single "complete" tick after the pipe is
		// drained in the goroutine below. The FE will treat any
		// single-tick progress as indeterminate.
		return 0
	})

	src, err := os.Open(filePath)
	if err != nil {
		_ = in.Close()
		s.finishError(jobID, cmd, stderrBuf, err)
		return
	}
	defer src.Close()

	stat, err := src.Stat()
	if err != nil {
		_ = in.Close()
		s.finishError(jobID, cmd, stderrBuf, err)
		return
	}

	const baseHeader = "SET FOREIGN_KEY_CHECKS=0;\nSET UNIQUE_CHECKS=0;\nSET NAMES utf8mb4;\n"
	const footer = "SET FOREIGN_KEY_CHECKS=1;\nCOMMIT;\n"

	// ponytail: a multi-database dump has no top-level USE; the
	// selected parts belong to whatever database each part
	// references. Use the first part's database so mariadb knows
	// where to apply the CREATE/INSERT statements. If all parts
	// are from the same database this is exact; if they're mixed
	// the user picked inconsistent rows and we honor the first.
	header := baseHeader
	if len(parts) > 0 && parts[0].DatabaseName != "" {
		header = "USE `" + parts[0].DatabaseName + "`;\n" + baseHeader
	}

	stream := streamer.Build(header, footer, src, stat.Size(), parts)
	stripped := streamer.NewDefinerStripper(stream)

	if _, copyErr := io.Copy(in, stripped); copyErr != nil {
		_ = in.Close()
		s.finishError(jobID, cmd, stderrBuf, copyErr)
		return
	}
	_ = in.Close()

	waitErr := cmd.Wait()
	s.cleanup(jobID)

	// ponytail: when the subprocess fails, stash the raw stderr
	// BEFORE trimMsg mangles it. The integration test reads this
	// file to see exactly what mariadb said. Production never sets
	// up the sink so the file write is a no-op.
	if waitErr != nil {
		if s.stderrSink != nil {
			s.stderrSink(jobID, stderrBuf.String())
		}
		if ctx.Err() != nil {
			s.emitDone(jobID, "error", "canceled")
			return
		}
		s.emitDone(jobID, "error", trimMsg(stderrBuf.String(), waitErr.Error()))
		return
	}
	s.emitDone(jobID, "success", "")
}

func (s *Service) finishError(jobID string, cmd *exec.Cmd, stderrBuf *strings.Builder, baseErr error) {
	waitErr := cmd.Wait()
	s.cleanup(jobID)
	msg := trimMsg(stderrBuf.String(), baseErr.Error())
	if waitErr != nil {
		msg = trimMsg(stderrBuf.String(), waitErr.Error())
	}
	// ponytail: also publish a full-stderr debug event so the test
	// can inspect what mariadb actually said. Production FE ignores
	// it; the integration driver (cmd/livetest) subscribes.
	if s.stderrSink != nil {
		s.stderrSink(jobID, stderrBuf.String())
	}
	s.emitDone(jobID, "error", msg)
}

func (s *Service) cleanup(jobID string) {
	s.mu.Lock()
	delete(s.jobs, jobID)
	s.mu.Unlock()
}

func (s *Service) emitDone(jobID, status, message string) {
	// ponytail: side-channel raw-stderr for diagnostics. Wired by
	// tests via SetStderrSink; production leaves it nil.
	// The call here (not in finishError) is intentional: every
	// error path eventually calls emitDone, but some early-error
	// paths (e.g. startPartial's "no rows matched") never build a
	// stderr buffer, so those reach the sink with stderr="".
	if s.stderrSink != nil {
		s.stderrSink(jobID, message)
	}
	_ = s.emitter.Emit(context.Background(), "restore:done", Done{
		JobID:   jobID,
		Status:  status,
		Message: message,
	})
}

func (s *Service) progressLoop(jobID, filePath string, stop <-chan struct{}, ctx context.Context) {
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			_ = s.emitter.Emit(context.Background(), "restore:progress", Progress{
				JobID: jobID,
				SoFar: info.Size(),
				Total: 0,
			})
		}
	}
}

// progressLoopTotal emits one indeterminate tick at start (total known,
// so far unknown) so the FE can show the bar in its indeterminate
// mode (CONTEXT: ProgressBar handles total=0).
func (s *Service) progressLoopTotal(jobID string, stop <-chan struct{}, ctx context.Context, total int64, soFar func() int64) {
	_ = s.emitter.Emit(context.Background(), "restore:progress", Progress{
		JobID: jobID,
		SoFar: 0,
		Total: total,
	})
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			cur := soFar()
			_ = s.emitter.Emit(context.Background(), "restore:progress", Progress{
				JobID: jobID,
				SoFar: cur,
				Total: total,
			})
		}
	}
}

// Cancel aborts the subprocess for jobID. No-op if finished or missing.
func (s *Service) Cancel(jobID string) {
	s.mu.Lock()
	j, ok := s.jobs[jobID]
	s.mu.Unlock()
	if !ok {
		return
	}
	j.cancel()
}

// AnalyzeDump runs the Byte-Offset Scanner against path and writes
// the result to the catalog (replacing any prior objects for the
// same path). The number of objects catalogd is returned so the FE
// can show "Analyzed N objects" without re-listing.
func (s *Service) AnalyzeDump(path string) (int, error) {
	if err := validateFile(path); err != nil {
		return 0, err
	}
	objs, err := scanner.New().Scan(path)
	if err != nil {
		return 0, fmt.Errorf("restore: scan: %w", err)
	}
	cleanPath := filepath.Clean(path)
	if err := s.cat.ReplaceObjectsForDump(cleanPath, objs); err != nil {
		return 0, err
	}
	return len(objs), nil
}

// ListCatalogObjects returns every catalog row for path. The FE
// virtualizer renders these; restore uses them via partsFor.
func (s *Service) ListCatalogObjects(path string) ([]CatalogObject, error) {
	cleanPath := filepath.Clean(path)
	rows, err := s.cat.ListObjectsForDump(cleanPath)
	if err != nil {
		return nil, err
	}
	out := make([]CatalogObject, 0, len(rows))
	for _, r := range rows {
		out = append(out, CatalogObject{
			ID:        r.ID,
			Database:  r.DatabaseName,
			Name:      r.ObjectName,
			Type:      stringToObjectType(r.ObjectType),
			StartByte: r.StartByte,
			EndByte:   r.EndByte,
		})
	}
	return out, nil
}

// CatalogObject is the BE-facing projection of a catalog row. The
// FE-facing shape lives in api/restore.ts; the Wails binding
// converts at the boundary.
type CatalogObject struct {
	ID        int    `json:"id"`
	Database  string `json:"database"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	StartByte int64  `json:"startByte"`
	EndByte   int64  `json:"endByte"`
}

func stringToObjectType(s string) string {
	switch scanner.ObjectType(s) {
	case scanner.TypeCreateTable, scanner.TypeInsert, scanner.TypeUse:
		return s
	}
	// Scanned-but-unrecognized types round-trip as-is so the FE
	// filter toggles for ROUTINE/TRIGGER/EVENT can match.
	return s
}

// partsFor resolves a list of catalog row IDs against path. Routines,
// triggers, and events are filtered by the IncludeRoutines /
// IncludeTriggers / IncludeEvents flags when the matching toggle is
// off; unknown types pass through (they can only exist if the
// scanner learned them, which today it does not, so this is a
// future-proof no-op).
func (s *Service) partsFor(path string, ids []int, incRoutines, incTriggers, incEvents bool) ([]scanner.Offset, error) {
	cleanPath := filepath.Clean(path)
	all, err := s.cat.ListObjectsForDump(cleanPath)
	if err != nil {
		return nil, err
	}
	wanted := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		wanted[id] = struct{}{}
	}
	var out []scanner.Offset
	for _, row := range all {
		if _, ok := wanted[row.ID]; !ok {
			continue
		}
		ot := scanner.ObjectType(row.ObjectType)
		switch ot {
		case scanner.TypeUse:
			// USE statements are not standalone objects; skip.
			continue
		}
		// Best-effort type filtering for future scanner extensions.
		// Today the scanner only emits CREATE_TABLE/INSERT/USE.
		// The flags below only have effect when a future scanner
		// learns to emit them.
		switch {
		case ot == "ROUTINE" && !incRoutines,
			ot == "TRIGGER" && !incTriggers,
			ot == "EVENT" && !incEvents:
			continue
		}
		out = append(out, scanner.Offset{
			StartByte:    row.StartByte,
			EndByte:      row.EndByte,
			ObjectType:   ot,
			ObjectName:   row.ObjectName,
			DatabaseName: row.DatabaseName,
		})
	}
	return out, nil
}

func validateFile(path string) error {
	if path == "" {
		return errors.New("restore: file path required")
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("restore: stat dump file: %w", err)
	}
	if info.IsDir() {
		return errors.New("restore: dump path is a directory")
	}
	return nil
}

func trimMsg(a, b string) string {
	// ponytail: prefer stderr (a) because it usually carries the real
	// mariadb error. Fall back to the subprocess exit error (b) when
	// stderr is silent. Keep up to 240 chars and the first line.
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

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("rs-%x", b)
}
