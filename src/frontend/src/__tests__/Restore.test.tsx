import { act, fireEvent, getByText, render, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const listServerProfiles = vi.fn();
const startFullRestore = vi.fn();
const cancelRestore = vi.fn();
const analyzeDump = vi.fn();
const openDumpFileDialog = vi.fn();

vi.mock('@/api', () => ({
  ListServerProfiles: (...a: unknown[]) => listServerProfiles(...a),
  startFullRestore: (...a: unknown[]) => startFullRestore(...a),
  cancelRestore: (...a: unknown[]) => cancelRestore(...a),
  analyzeDump: (...a: unknown[]) => analyzeDump(...a),
}));

vi.mock('../../wailsjs/go/app/App', () => ({
  OpenDumpFileDialog: (...a: unknown[]) => openDumpFileDialog(...a),
}));

const listeners = new Map<string, (...data: unknown[]) => void>();
vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn((name: string, cb: (...data: unknown[]) => void) => {
    listeners.set(name, cb);
    return () => listeners.delete(name);
  }),
}));

import { Restore } from '../pages/Restore/Restore';

const emit = (name: string, payload: unknown) => {
  const cb = listeners.get(name);
  if (cb) cb(payload);
};

beforeEach(() => {
  listServerProfiles.mockReset();
  startFullRestore.mockReset();
  cancelRestore.mockReset();
  analyzeDump.mockReset();
  openDumpFileDialog.mockReset();
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
  openDumpFileDialog.mockResolvedValue('/tmp/dump.sql');
});

afterEach(() => {
  listeners.clear();
  window.location.hash = '';
});

const restoreButton = (root: HTMLElement) => {
  const el = root.querySelector<HTMLButtonElement>('button[data-action="restore"]');
  if (!el) throw new Error('restore button not found');
  return el;
};
const analyzeButton = (root: HTMLElement) => {
  const el = root.querySelector<HTMLButtonElement>('button[data-action="analyze"]');
  if (!el) throw new Error('analyze button not found');
  return el;
};

const selectFile = async (container: HTMLElement) => {
  const fileButton = container.querySelector<HTMLButtonElement>('#file');
  if (!fileButton) throw new Error('file picker button not found');
  await act(async () => {
    fireEvent.click(fileButton);
  });
};

describe('Restore', () => {
  it('renders the form with both action buttons', async () => {
    const { container } = render(<Restore />);
    await waitFor(() => {
      expect(restoreButton(container)).toBeInTheDocument();
    });
    expect(analyzeButton(container)).toBeInTheDocument();
  });

  it('both action buttons are disabled until file + profile are set', async () => {
    const { container } = render(<Restore />);
    await waitFor(() => {
      expect(restoreButton(container)).toBeInTheDocument();
    });
    expect(restoreButton(container)).toBeDisabled();
    expect(analyzeButton(container)).toBeDisabled();
  });

  it('Restore calls api.startFullRestore with the picked inputs', async () => {
    startFullRestore.mockResolvedValue({ jobId: 'job-1', cancel: vi.fn() });
    const { container, getByLabelText } = render(<Restore />);
    await waitFor(() => {
      expect(restoreButton(container)).toBeInTheDocument();
    });
    await selectFile(container);
    fireEvent.change(getByLabelText(/Target server profile/), { target: { value: 'p1' } });

    await act(async () => {
      fireEvent.click(restoreButton(container));
    });

    expect(startFullRestore).toHaveBeenCalledWith({
      filePath: '/tmp/dump.sql',
      profileId: 'p1',
    });
  });

  it('progress events advance the ProgressBar', async () => {
    startFullRestore.mockResolvedValue({ jobId: 'job-1', cancel: vi.fn() });
    const { container, getByLabelText, getByRole } = render(<Restore />);
    await waitFor(() => {
      expect(restoreButton(container)).toBeInTheDocument();
    });
    await selectFile(container);
    fireEvent.change(getByLabelText(/Target server profile/), { target: { value: 'p1' } });
    await act(async () => {
      fireEvent.click(restoreButton(container));
    });

    await act(async () => {
      emit('restore:progress', { jobId: 'job-1', soFar: 25, total: 100 });
    });

    await waitFor(() => {
      expect(getByRole('progressbar').getAttribute('aria-valuenow')).toBe('25');
    });
  });

  it('Cancel calls api.cancelRestore', async () => {
    startFullRestore.mockResolvedValue({ jobId: 'job-1', cancel: vi.fn() });
    cancelRestore.mockResolvedValue(undefined);
    const { container, getByLabelText } = render(<Restore />);
    await waitFor(() => {
      expect(restoreButton(container)).toBeInTheDocument();
    });
    await selectFile(container);
    fireEvent.change(getByLabelText(/Target server profile/), { target: { value: 'p1' } });
    await act(async () => {
      fireEvent.click(restoreButton(container));
    });
    await waitFor(() => {
      expect(getByText(container, 'Cancel')).toBeInTheDocument();
    });
    await act(async () => {
      fireEvent.click(getByText(container, 'Cancel'));
    });
    expect(cancelRestore).toHaveBeenCalledWith('job-1');
  });

  it('Analyze navigates to /restore/select with file + profile', async () => {
    analyzeDump.mockResolvedValue({ objectCount: 0 });
    const { container, getByLabelText } = render(<Restore />);
    await waitFor(() => {
      expect(analyzeButton(container)).toBeInTheDocument();
    });
    await selectFile(container);
    fireEvent.change(getByLabelText(/Target server profile/), { target: { value: 'p1' } });

    await act(async () => {
      fireEvent.click(analyzeButton(container));
    });

    expect(analyzeDump).toHaveBeenCalledWith('/tmp/dump.sql');
    expect(window.location.hash).toBe('#/restore/select?file=%2Ftmp%2Fdump.sql&profile=p1');
  });

  it('error from startFullRestore renders inline', async () => {
    startFullRestore.mockRejectedValue(new Error('restore boom'));
    const { container, getByLabelText } = render(<Restore />);
    await waitFor(() => {
      expect(restoreButton(container)).toBeInTheDocument();
    });
    await selectFile(container);
    fireEvent.change(getByLabelText(/Target server profile/), { target: { value: 'p1' } });
    await act(async () => {
      fireEvent.click(restoreButton(container));
    });
    await waitFor(() => {
      expect(getByText(container, /Error: restore boom/)).toBeInTheDocument();
    });
  });
});
