import { act, render, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const listServerProfiles = vi.fn();
vi.mock('@/api', () => ({
  ListServerProfiles: (...a: unknown[]) => listServerProfiles(...a),
}));

import { Dashboard } from '../pages/Dashboard/Dashboard';

beforeEach(() => {
  listServerProfiles.mockReset();
  window.location.hash = '';
});

afterEach(() => {
  window.location.hash = '';
});

describe('Dashboard', () => {
  it('shows a loading state until the profile list resolves', () => {
    listServerProfiles.mockReturnValue(new Promise(() => {}));
    const { getByText } = render(<Dashboard />);
    expect(getByText(/Loading/i)).toBeInTheDocument();
  });

  it('renders the empty state when there are no profiles', async () => {
    listServerProfiles.mockResolvedValue([]);
    const { getByText } = render(<Dashboard />);
    await waitFor(() => {
      expect(getByText(/No server profiles yet/i)).toBeInTheDocument();
    });
  });

  it('renders a profile card per item with both action buttons', async () => {
    listServerProfiles.mockResolvedValue([
      {
        id: 'p1',
        name: 'prod',
        host: 'db.example.com',
        port: 3306,
        user: 'root',
        hasPassword: true,
        sslMode: 'preferred',
      },
    ]);
    const { getByText } = render(<Dashboard />);
    await waitFor(() => {
      expect(getByText('prod')).toBeInTheDocument();
    });
    expect(getByText('Use for Backup')).toBeInTheDocument();
    expect(getByText('Use for Restore')).toBeInTheDocument();
  });

  it('navigates to /backup with profile id when "Use for Backup" is clicked', async () => {
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
    const { getByText } = render(<Dashboard />);
    await waitFor(() => {
      expect(getByText('Use for Backup')).toBeInTheDocument();
    });

    await act(async () => {
      getByText('Use for Backup').click();
    });
    expect(window.location.hash).toBe('#/backup?profile=p1');
  });

  it('navigates to /restore with profile id when "Use for Restore" is clicked', async () => {
    listServerProfiles.mockResolvedValue([
      {
        id: 'p2',
        name: 'stage',
        host: 'db',
        port: 3306,
        user: 'u',
        hasPassword: false,
        sslMode: 'preferred',
      },
    ]);
    const { getByText } = render(<Dashboard />);
    await waitFor(() => {
      expect(getByText('Use for Restore')).toBeInTheDocument();
    });

    await act(async () => {
      getByText('Use for Restore').click();
    });
    expect(window.location.hash).toBe('#/restore?profile=p2');
  });

  it('renders an error message when listing fails', async () => {
    listServerProfiles.mockRejectedValue(new Error('boom'));
    const { getByText } = render(<Dashboard />);
    await waitFor(() => {
      expect(getByText('Error: boom')).toBeInTheDocument();
    });
  });
});
