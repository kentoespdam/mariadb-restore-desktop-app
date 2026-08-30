package app

import "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/backup"

// BackupRequest is the wire shape the Wails binding accepts. The
// FE-facing TS type lives in src/frontend/src/api/backup.ts; the
// Wails codegen derives a matching TS interface from these tags.
type BackupRequest struct {
	ProfileID  string   `json:"profileId"`
	Databases  []string `json:"databases"`
	OutputPath string   `json:"outputPath"`
	BinaryPath string   `json:"binaryPath,omitempty"`
}

// StartBackup kicks off the dump and returns the JobID. The
// caller subscribes to "backup:progress" + "backup:done" events
// to follow the job. UI throttles progress at 150ms.
func (a *App) StartBackup(req BackupRequest) (string, error) {
	return a.Backup.Start(bgCtx(), backup.Request{
		ProfileID:  req.ProfileID,
		Databases:  req.Databases,
		OutputPath: req.OutputPath,
		BinaryPath: req.BinaryPath,
	})
}

// CancelBackup aborts the subprocess for jobID. No-op if the job
// already finished or never existed.
func (a *App) CancelBackup(jobID string) error {
	a.Backup.Cancel(jobID)
	return nil
}
