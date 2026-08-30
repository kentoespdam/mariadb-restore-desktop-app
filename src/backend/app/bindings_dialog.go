package app

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// OpenDumpFileDialog opens a native file dialog filtered to .sql files
// and returns the selected absolute path. Returns empty string if the
// user cancels.
func (a *App) OpenDumpFileDialog() (string, error) {
	return runtime.OpenFileDialog(bgCtx(), runtime.OpenDialogOptions{
		Title: "Select Dump File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "SQL Files (*.sql)",
				Pattern:     "*.sql",
			},
		},
	})
}
