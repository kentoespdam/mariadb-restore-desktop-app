package app

import (
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/crypto"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/recovery"
	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/features/settings"
)

// SettingsView is the wire shape the FE Settings screen renders.
type SettingsView struct {
	ExeDir      string `json:"exeDir"`
	CatalogPath string `json:"catalogPath"`
	AppKeyPath  string `json:"appKeyPath"`
	MariadbPath string `json:"mariadbPath"`
	MariadbDump string `json:"mariadbDumpPath"`
	KeyBits     int    `json:"keyBits"`
}

// GetSettings returns the read-only scope info plus the editable
// binary path overrides. Bundled so the FE only has to ask once.
func (a *App) GetSettings() (SettingsView, error) {
	v, err := a.Settings.BuildView(a.CatPath, a.KeyPath)
	if err != nil {
		return SettingsView{}, err
	}
	return SettingsView{
		ExeDir:      v.ExeDir,
		CatalogPath: v.CatalogPath,
		AppKeyPath:  v.AppKeyPath,
		MariadbPath: v.MariadbPath,
		MariadbDump: v.MariadbDump,
		KeyBits:     v.KeyBits,
	}, nil
}

// SaveSettingsInput is the wire shape for the MariaDB binaries
// section. Saves atomically (temp file + rename).
type SaveSettingsInput struct {
	MariadbPath string `json:"mariadbPath"`
	MariadbDump string `json:"mariadbDumpPath"`
}

// SaveSettings persists the binary-path overrides.
func (a *App) SaveSettings(in SaveSettingsInput) error {
	return a.Settings.Save(settings.Input{
		MariadbPath:     in.MariadbPath,
		MariadbDumpPath: in.MariadbDump,
	})
}

// ResetAndReinitResult is what the FE Settings screen reads after
// the user confirms the Danger-zone reset.
type ResetAndReinitResult struct {
	Triggered string `json:"triggered"`
}

// ResetAndReinit wipes the catalog + regenerates app.key. The
// Smart Recovery modal will subsequently fire via the boot probe
// when the user starts the next operation; we don't synthesize
// the event here because the user already confirmed the reset by
// clicking the Settings Reset button.
func (a *App) ResetAndReinit() (ResetAndReinitResult, error) {
	if err := a.Recovery.Cat.Close(); err != nil {
		return ResetAndReinitResult{}, err
	}
	if _, err := recovery.Wipe(a.Recovery.Paths); err != nil {
		return ResetAndReinitResult{}, err
	}
	if _, err := crypto.GenerateKey(a.Recovery.Paths.KeyPath); err != nil {
		return ResetAndReinitResult{}, err
	}
	return ResetAndReinitResult{Triggered: "reset"}, nil
}
