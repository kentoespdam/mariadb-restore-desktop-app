import { act, fireEvent, render } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const listeners = new Map<string, (...data: unknown[]) => void>();

vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn((name: string, cb: (...data: unknown[]) => void) => {
    listeners.set(name, cb);
    return () => listeners.delete(name);
  }),
}));

const decision = vi.fn();
vi.mock('@/api', () => ({ RecoveryDecision: (...a: unknown[]) => decision(...a) }));

import { RecoveryDialog } from '../components/RecoveryDialog';

const emit = (payload: unknown) => {
  const cb = listeners.get('recovery:show');
  if (cb) cb(payload);
};

beforeEach(() => {
  listeners.clear();
  decision.mockReset();
  decision.mockResolvedValue(undefined);
});

afterEach(() => {
  listeners.clear();
});

describe('RecoveryDialog', () => {
  it('renders nothing until the recovery:show event fires', () => {
    const { queryByRole } = render(<RecoveryDialog />);
    expect(queryByRole('dialog')).toBeNull();
  });

  it('opens on recovery:show with the missing_key reason', async () => {
    const { getByRole, getByText } = render(<RecoveryDialog />);
    await act(async () => {
      emit({ Reason: 'missing_key' });
    });
    expect(getByRole('dialog', { name: 'Smart Recovery' })).toBeInTheDocument();
    expect(getByText(/app\.key/)).toBeInTheDocument();
  });

  it('Reset button calls RecoveryDecision("reset") and closes', async () => {
    const { getByText, queryByRole } = render(<RecoveryDialog />);
    await act(async () => {
      emit({ Reason: 'missing_key' });
    });

    await act(async () => {
      fireEvent.click(getByText('Reset & Re-init'));
    });

    expect(decision).toHaveBeenCalledWith('reset');
    expect(queryByRole('dialog')).toBeNull();
  });

  it('Cancel button calls RecoveryDecision("cancel") and closes', async () => {
    const { getByText, queryByRole } = render(<RecoveryDialog />);
    await act(async () => {
      emit({ Reason: 'missing_key' });
    });

    await act(async () => {
      fireEvent.click(getByText('Cancel'));
    });

    expect(decision).toHaveBeenCalledWith('cancel');
    expect(queryByRole('dialog')).toBeNull();
  });

  it('unsubscribes on unmount', () => {
    const { unmount } = render(<RecoveryDialog />);
    expect(listeners.has('recovery:show')).toBe(true);
    unmount();
    expect(listeners.has('recovery:show')).toBe(false);
  });
});
