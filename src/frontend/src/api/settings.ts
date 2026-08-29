// Settings: FE contract.
//
// Executable Scope (CONTEXT): all app files live beside the binary.
// The Settings screen surfaces this as read-only info and lets the
// user override the mariadb / mariadb-dump binary paths.

export interface Settings {
  exeDir: string;
  catalogPath: string;
  appKeyPath: string;
  mariadbPath: string;
  mariadbDumpPath: string;
  keyBits: 256;
}

export interface SaveSettingsInput {
  mariadbPath: string;
  mariadbDumpPath: string;
}

function notImplemented(name: string): never {
  throw new Error(`not implemented: ${name}`);
}

export async function getSettings(): Promise<Settings> {
  notImplemented('getSettings');
}

export async function saveSettings(_input: SaveSettingsInput): Promise<void> {
  notImplemented('saveSettings');
}

// ponytail: returns "unknown" sentinel so the Settings screen can
// render before any real binding lands; replace with the real
// recovery entry once the BE wires one.
export async function resetAndReinit(): Promise<{ triggered: 'unknown' }> {
  notImplemented('resetAndReinit');
}
