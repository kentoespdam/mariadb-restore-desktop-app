// Settings: real Wails bindings. The Go side persists binary-path
// overrides in settings.json beside the binary (CONTEXT:
// Executable Scope).
import { GetSettings, ResetAndReinit, SaveSettings } from '../../wailsjs/go/app/App';
import type { app } from '../../wailsjs/go/models';

export type Settings = app.SettingsView;
export type SaveSettingsInput = app.SaveSettingsInput;
export type ResetAndReinitResult = app.ResetAndReinitResult;

export async function getSettings(): Promise<Settings> {
  return GetSettings();
}

export async function saveSettings(input: SaveSettingsInput): Promise<void> {
  await SaveSettings(input);
}

export async function resetAndReinit(): Promise<ResetAndReinitResult> {
  return ResetAndReinit();
}
