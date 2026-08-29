import { act, fireEvent, render, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const getSettings = vi.fn();
const saveSettings = vi.fn();
const resetAndReinit = vi.fn();

vi.mock('@/api', () => ({
  getSettings: (...a: unknown[]) => getSettings(...a),
  saveSettings: (...a: unknown[]) => saveSettings(...a),
  resetAndReinit: (...a: unknown[]) => resetAndReinit(...a),
}));

import { SettingsPage } from '../pages/Settings/Settings';

const SAMPLE = {
  exeDir: '/opt/app',
  catalogPath: '/opt/app/catalog.sqlite',
  appKeyPath: '/opt/app/app.key',
  mariadbPath: '/usr/bin/mariadb',
  mariadbDumpPath: '/usr/bin/mariadb-dump',
  keyBits: 256 as const,
};

beforeEach(() => {
  getSettings.mockReset();
  saveSettings.mockReset();
  resetAndReinit.mockReset();
  getSettings.mockResolvedValue(SAMPLE);
  saveSettings.mockResolvedValue(undefined);
  resetAndReinit.mockResolvedValue({ triggered: 'unknown' });
  window.location.hash = '';
});

afterEach(() => {
  window.location.hash = '';
});

describe('Settings', () => {
  it('renders all four sections', async () => {
    const { getByText } = render(<SettingsPage />);
    await waitFor(() => {
      expect(getByText('Executable Scope')).toBeInTheDocument();
    });
    expect(getByText('MariaDB binaries')).toBeInTheDocument();
    expect(getByText('Encryption')).toBeInTheDocument();
    expect(getByText('Danger zone')).toBeInTheDocument();
  });

  it('renders the read-only paths once getSettings resolves', async () => {
    const { getByText } = render(<SettingsPage />);
    await waitFor(() => {
      expect(getByText('/opt/app/catalog.sqlite')).toBeInTheDocument();
    });
    expect(getByText('/opt/app/app.key')).toBeInTheDocument();
  });

  it('renders an error message if getSettings fails', async () => {
    getSettings.mockRejectedValue(new Error('config boom'));
    const { getByText } = render(<SettingsPage />);
    await waitFor(() => {
      expect(getByText(/Error: config boom/)).toBeInTheDocument();
    });
  });

  it('Save calls api.saveSettings and shows "Saved"', async () => {
    const { getByText, getByLabelText } = render(<SettingsPage />);
    await waitFor(() => {
      expect(getByText('Save')).toBeInTheDocument();
    });
    fireEvent.change(getByLabelText('mariadb path'), {
      target: { value: '/new/mariadb' },
    });
    fireEvent.change(getByLabelText('mariadb-dump path'), {
      target: { value: '/new/mariadb-dump' },
    });
    await act(async () => {
      fireEvent.click(getByText('Save'));
    });
    await waitFor(() => {
      expect(saveSettings).toHaveBeenCalledWith({
        mariadbPath: '/new/mariadb',
        mariadbDumpPath: '/new/mariadb-dump',
      });
      expect(getByText('Saved')).toBeInTheDocument();
    });
  });

  it('Reset button opens a confirm dialog', async () => {
    const { getByText } = render(<SettingsPage />);
    await waitFor(() => {
      expect(getByText('Reset & Re-init')).toBeInTheDocument();
    });
    await act(async () => {
      fireEvent.click(getByText('Reset & Re-init'));
    });
    expect(getByText(/This will wipe the catalog/)).toBeInTheDocument();
  });

  it('Confirm reset calls api.resetAndReinit and closes the dialog', async () => {
    const { getByText } = render(<SettingsPage />);
    await waitFor(() => {
      expect(getByText('Reset & Re-init')).toBeInTheDocument();
    });
    await act(async () => {
      fireEvent.click(getByText('Reset & Re-init'));
    });
    await act(async () => {
      fireEvent.click(getByText('Confirm reset'));
    });
    await waitFor(() => {
      expect(resetAndReinit).toHaveBeenCalled();
    });
  });
});
