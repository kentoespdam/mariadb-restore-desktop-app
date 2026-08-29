import { render, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const listServerProfiles = vi.fn();
const getSettings = vi.fn();
const listCatalogObjects = vi.fn();
const analyzeDump = vi.fn();
const startFullRestore = vi.fn();
const startPartialRestore = vi.fn();
const startBackup = vi.fn();
const cancelBackup = vi.fn();
const cancelRestore = vi.fn();
const saveSettings = vi.fn();
const resetAndReinit = vi.fn();

vi.mock('@/api', () => ({
  ListServerProfiles: (...a: unknown[]) => listServerProfiles(...a),
  getSettings: (...a: unknown[]) => getSettings(...a),
  listCatalogObjects: (...a: unknown[]) => listCatalogObjects(...a),
  analyzeDump: (...a: unknown[]) => analyzeDump(...a),
  startFullRestore: (...a: unknown[]) => startFullRestore(...a),
  startPartialRestore: (...a: unknown[]) => startPartialRestore(...a),
  startBackup: (...a: unknown[]) => startBackup(...a),
  cancelBackup: (...a: unknown[]) => cancelBackup(...a),
  cancelRestore: (...a: unknown[]) => cancelRestore(...a),
  saveSettings: (...a: unknown[]) => saveSettings(...a),
  resetAndReinit: (...a: unknown[]) => resetAndReinit(...a),
}));

vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(() => () => {}),
  EventsOnMultiple: vi.fn(() => () => {}),
  EventsOff: vi.fn(),
  EventsOffAll: vi.fn(),
  EventsEmit: vi.fn(),
  EventsOnce: vi.fn(() => () => {}),
}));

import { App } from '../app';

beforeEach(() => {
  listServerProfiles.mockReset();
  listServerProfiles.mockResolvedValue([]);
  getSettings.mockReset();
  getSettings.mockResolvedValue({
    exeDir: '/x',
    catalogPath: '/x/catalog.sqlite',
    appKeyPath: '/x/app.key',
    mariadbPath: '/usr/bin/mariadb',
    mariadbDumpPath: '/usr/bin/mariadb-dump',
    keyBits: 256,
  });
  listCatalogObjects.mockReset();
  window.location.hash = '';
});

afterEach(() => {
  window.location.hash = '';
});

describe('App', () => {
  it('renders the navigation and default route', async () => {
    const { getByRole, getByText } = render(<App />);
    expect(getByRole('heading', { level: 1, name: 'MariaDB Tools' })).toBeInTheDocument();
    await waitFor(() => {
      expect(getByText(/No server profiles yet/i)).toBeInTheDocument();
    });
  });

  it('renders each route when navigated', async () => {
    const routes: ReadonlyArray<readonly [string, RegExp]> = [
      ['#/dashboard', /No server profiles yet/i],
      ['#/profiles', /Credentials are encrypted/i],
      ['#/backup', /Output file/],
      ['#/restore', /Full Restore pipes/i],
      ['#/restore/select', /No dump file selected/i],
      ['#/settings', /MariaDB binaries/i],
    ];

    for (const [hash, text] of routes) {
      window.location.hash = hash;
      const { getByText, unmount } = render(<App />);
      await waitFor(() => {
        expect(getByText(text)).toBeInTheDocument();
      });
      unmount();
    }
  });
});
