import { act, fireEvent, render, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const listServerProfiles = vi.fn();
const startBackup = vi.fn();
const cancelBackup = vi.fn();

vi.mock('@/api', () => ({
  ListServerProfiles: (...a: unknown[]) => listServerProfiles(...a),
  startBackup: (...a: unknown[]) => startBackup(...a),
  cancelBackup: (...a: unknown[]) => cancelBackup(...a),
}));

const listeners = new Map<string, (...data: unknown[]) => void>();
vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn((name: string, cb: (...data: unknown[]) => void) => {
    listeners.set(name, cb);
    return () => listeners.delete(name);
  }),
}));

import { Backup } from '../pages/Backup/Backup';

const emit = (name: string, payload: unknown) => {
  const cb = listeners.get(name);
  if (cb) cb(payload);
};

beforeEach(() => {
  listServerProfiles.mockReset();
  startBackup.mockReset();
  cancelBackup.mockReset();
  listeners.clear();
  window.location.hash = '';
  listServerProfiles.mockResolvedValue([
    {
      id: 'p1',
      name: 'prod',
      host: 'db',
      port: 3306,
      user: 'u',
      hasPassword: false,
      sslMode: 'preferred',
    },
  ]);
});

afterEach(() => {
  listeners.clear();
  window.location.hash = '';
});

describe('Backup', () => {
  it('renders the form', async () => {
    const { getByLabelText, getByText } = render(<Backup />);
    await waitFor(() => {
      expect(getByText('Start backup')).toBeInTheDocument();
    });
    expect(getByLabelText('Server profile')).toBeInTheDocument();
    expect(getByLabelText(/Databases/)).toBeInTheDocument();
    expect(getByLabelText(/Output file/)).toBeInTheDocument();
  });

  it('Start is disabled until all fields are filled', async () => {
    const { getByText } = render(<Backup />);
    await waitFor(() => {
      expect(getByText('Start backup')).toBeInTheDocument();
    });
    expect(getByText('Start backup').closest('button')).toBeDisabled();
  });

  it('Start calls api.startBackup with the parsed request', async () => {
    startBackup.mockResolvedValue({ jobId: 'job-1', cancel: vi.fn() });
    const { getByLabelText, getByText } = render(<Backup />);
    await waitFor(() => {
      expect(getByText('Start backup')).toBeInTheDocument();
    });

    fireEvent.change(getByLabelText('Server profile'), { target: { value: 'p1' } });
    fireEvent.change(getByLabelText(/Databases/), { target: { value: 'a, b ,c' } });
    fireEvent.change(getByLabelText(/Output file/), { target: { value: 'out.sql' } });

    await act(async () => {
      fireEvent.click(getByText('Start backup'));
    });

    expect(startBackup).toHaveBeenCalledWith({
      profileId: 'p1',
      databases: ['a', 'b', 'c'],
      outputPath: 'out.sql',
    });
  });

  it('progress events advance the ProgressBar', async () => {
    startBackup.mockResolvedValue({ jobId: 'job-1', cancel: vi.fn() });
    const { getByLabelText, getByText, getByRole } = render(<Backup />);
    await waitFor(() => {
      expect(getByText('Start backup')).toBeInTheDocument();
    });
    fireEvent.change(getByLabelText('Server profile'), { target: { value: 'p1' } });
    fireEvent.change(getByLabelText(/Databases/), { target: { value: 'a' } });
    fireEvent.change(getByLabelText(/Output file/), { target: { value: 'o.sql' } });
    await act(async () => {
      fireEvent.click(getByText('Start backup'));
    });

    await act(async () => {
      emit('backup:progress', { jobId: 'job-1', soFar: 50, total: 100 });
    });

    await waitFor(() => {
      const bar = getByRole('progressbar');
      expect(bar.getAttribute('aria-valuenow')).toBe('50');
    });
  });

  it('Cancel button calls api.cancelBackup', async () => {
    startBackup.mockResolvedValue({ jobId: 'job-1', cancel: vi.fn() });
    cancelBackup.mockResolvedValue(undefined);
    const { getByLabelText, getByText } = render(<Backup />);
    await waitFor(() => {
      expect(getByText('Start backup')).toBeInTheDocument();
    });
    fireEvent.change(getByLabelText('Server profile'), { target: { value: 'p1' } });
    fireEvent.change(getByLabelText(/Databases/), { target: { value: 'a' } });
    fireEvent.change(getByLabelText(/Output file/), { target: { value: 'o.sql' } });
    await act(async () => {
      fireEvent.click(getByText('Start backup'));
    });

    await waitFor(() => {
      expect(getByText('Cancel')).toBeInTheDocument();
    });

    await act(async () => {
      fireEvent.click(getByText('Cancel'));
    });

    expect(cancelBackup).toHaveBeenCalledWith('job-1');
  });

  it('error from startBackup renders inline', async () => {
    startBackup.mockRejectedValue(new Error('dump failed'));
    const { getByLabelText, getByText } = render(<Backup />);
    await waitFor(() => {
      expect(getByText('Start backup')).toBeInTheDocument();
    });
    fireEvent.change(getByLabelText('Server profile'), { target: { value: 'p1' } });
    fireEvent.change(getByLabelText(/Databases/), { target: { value: 'a' } });
    fireEvent.change(getByLabelText(/Output file/), { target: { value: 'o.sql' } });
    await act(async () => {
      fireEvent.click(getByText('Start backup'));
    });

    await waitFor(() => {
      expect(getByText(/Error: dump failed/)).toBeInTheDocument();
    });
  });

  it('done event arriving the same tick as startBackup resolves still updates UI', async () => {
    // Regression for "backup UI stuck at Backup Progress": a fast
    // dump fires backup:done from the BE in the same tick as the
    // startBackup promise resolves. If the handler reads jobId from
    // closed-over state (not a ref), the render that reflects the
    // new jobId hasn't committed yet, and the done event is dropped.
    startBackup.mockImplementation(async () => {
      // Schedule the emit to run AFTER the awaiting microtask that
      // sets jobId, but in the same tick (no setTimeout) — mirroring
      // the real Wails ordering where the Go goroutine returns the
      // jobId and emits events into the bus before the FE promise
      // resolves.
      const handle = { jobId: 'job-fast', cancel: vi.fn() };
      Promise.resolve().then(() => {
        Promise.resolve().then(() => {
          emit('backup:done', { jobId: 'job-fast', status: 'success' });
        });
      });
      return handle;
    });

    const { getByLabelText, getByText } = render(<Backup />);
    await waitFor(() => {
      expect(getByText('Start backup')).toBeInTheDocument();
    });
    fireEvent.change(getByLabelText('Server profile'), { target: { value: 'p1' } });
    fireEvent.change(getByLabelText(/Databases/), { target: { value: 'a' } });
    fireEvent.change(getByLabelText(/Output file/), { target: { value: 'o.sql' } });

    await act(async () => {
      fireEvent.click(getByText('Start backup'));
    });

    await waitFor(() => {
      expect(getByText('Backup completed')).toBeInTheDocument();
    });
  });

  it('pre-selects the profile from the ?profile= query param', async () => {
    window.location.hash = '#/backup?profile=p1';
    const { getByLabelText, getByText } = render(<Backup />);
    await waitFor(() => {
      expect(getByText('Start backup')).toBeInTheDocument();
    });
    const select = getByLabelText('Server profile') as HTMLSelectElement;
    expect(select.value).toBe('p1');
  });
});
