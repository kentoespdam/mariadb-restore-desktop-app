import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { App } from '../app';

describe('App', () => {
  it('renders the navigation and default route', () => {
    render(<App />);
    expect(screen.getByRole('heading', { level: 1, name: 'MariaDB Tools' })).toBeTruthy();
    expect(screen.getByRole('heading', { level: 2, name: 'Dashboard' })).toBeTruthy();
    expect(screen.getByText(/Layout bootstrap ready/i)).toBeTruthy();
  });
});
