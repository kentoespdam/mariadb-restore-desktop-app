package app

import "github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/restore"

// RestoreRequest is the wire shape for Full Restore.
type RestoreRequest struct {
	ProfileID  string `json:"profileId"`
	FilePath   string `json:"filePath"`
	BinaryPath string `json:"binaryPath,omitempty"`
}

// StartFullRestore pipes the entire dump file to mariadb CLI on the
// target server. No scanning. (CONTEXT: "triggered when the user
// clicks Restore immediately after selecting a file, bypassing the
// Analyze step.")
func (a *App) StartFullRestore(req RestoreRequest) (string, error) {
	return a.Restore.StartFull(bgCtx(), restore.FullRequest{
		ProfileID:  req.ProfileID,
		FilePath:   req.FilePath,
		BinaryPath: req.BinaryPath,
	})
}

// PartialRestoreRequest is the wire shape for Partial Restore.
type PartialRestoreRequest struct {
	ProfileID       string `json:"profileId"`
	FilePath        string `json:"filePath"`
	SelectedIDs     []int  `json:"selectedIds"`
	IncludeRoutines bool   `json:"includeRoutines"`
	IncludeTriggers bool   `json:"includeTriggers"`
	IncludeEvents   bool   `json:"includeEvents"`
	BinaryPath      string `json:"binaryPath,omitempty"`
}

// StartPartialRestore pipes header + selected byte ranges + footer
// to mariadb CLI. CONTEXT: requires Analyze first; user picks
// tables and optionally routines/triggers/events.
func (a *App) StartPartialRestore(req PartialRestoreRequest) (string, error) {
	return a.Restore.StartPartial(bgCtx(), restore.PartialRequest{
		ProfileID:       req.ProfileID,
		FilePath:        req.FilePath,
		SelectedIDs:     req.SelectedIDs,
		IncludeRoutines: req.IncludeRoutines,
		IncludeTriggers: req.IncludeTriggers,
		IncludeEvents:   req.IncludeEvents,
		BinaryPath:      req.BinaryPath,
	})
}

// CancelRestore aborts the subprocess for jobID.
func (a *App) CancelRestore(jobID string) error {
	a.Restore.Cancel(jobID)
	return nil
}

// AnalyzeDump scans the dump file at path and writes the byte-range
// catalogue to the local SQLite. Returns the number of objects
// recorded.
func (a *App) AnalyzeDump(path string) (int, error) {
	return a.Restore.AnalyzeDump(path)
}

// CatalogObject is the wire shape the FE's ObjectGrid renders.
type CatalogObject struct {
	ID        int    `json:"id"`
	Database  string `json:"database"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	StartByte int64  `json:"startByte"`
	EndByte   int64  `json:"endByte"`
}

// ListCatalogObjects returns every catalog row for path so the FE
// can render the virtualized object grid.
func (a *App) ListCatalogObjects(path string) ([]CatalogObject, error) {
	rows, err := a.Restore.ListCatalogObjects(path)
	if err != nil {
		return nil, err
	}
	out := make([]CatalogObject, 0, len(rows))
	for _, r := range rows {
		out = append(out, CatalogObject{
			ID:        r.ID,
			Database:  r.Database,
			Name:      r.Name,
			Type:      r.Type,
			StartByte: r.StartByte,
			EndByte:   r.EndByte,
		})
	}
	return out, nil
}