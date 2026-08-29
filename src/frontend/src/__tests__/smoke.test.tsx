import { render, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const listServerProfiles = vi.fn();
vi.mock('@/api', () => ({
  ListServerProfiles: (...a: unknown[]) => listServerProfiles(...a),
}));

import { App } from '../app';

beforeEach(() => {
  listServerProfiles.mockReset();
  listServerProfiles.mockResolvedValue([]);
  window.location.hash = '';
});

describe('App', () => {
  it('renders the navigation and default route', async () => {
    const { getByRole, getByText } = render(<App />);
    expect(getByRole('heading', { level: 1, name: 'MariaDB Tools' })).toBeInTheDocument();
    expect(getByRole('heading', { level: 2, name: 'Dashboard' })).toBeInTheDocument();
    await waitFor(() => {
      expect(getByText(/No server profiles yet/i)).toBeInTheDocument();
    });
  });
});
