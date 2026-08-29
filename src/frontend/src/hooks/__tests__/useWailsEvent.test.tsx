import { act, render } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const listeners = new Map<string, (...data: unknown[]) => void>();
let unsubscribeCalls = 0;

vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn((name: string, cb: (...data: unknown[]) => void) => {
    listeners.set(name, cb);
    return () => {
      unsubscribeCalls += 1;
      listeners.delete(name);
    };
  }),
}));

import { useWailsEvent } from '../useWailsEvent';

type Payload = { n: number };

function Probe({
  name,
  onPayload,
  throttleMs,
}: {
  name: string;
  onPayload: (p: Payload) => void;
  throttleMs?: number;
}) {
  useWailsEvent<Payload>(name, onPayload, { throttleMs });
  return null;
}

function emit(name: string, p: Payload) {
  const cb = listeners.get(name);
  if (cb) cb(p);
}

beforeEach(() => {
  listeners.clear();
  unsubscribeCalls = 0;
  vi.useRealTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('useWailsEvent', () => {
  it('delivers events to the handler', () => {
    const onPayload = vi.fn();
    render(<Probe name="evt" onPayload={onPayload} />);
    act(() => emit('evt', { n: 1 }));
    expect(onPayload).toHaveBeenCalledWith({ n: 1 });
  });

  it('uses the latest handler without re-subscribing', () => {
    const a = vi.fn();
    const b = vi.fn();
    const { rerender } = render(<Probe name="evt" onPayload={a} />);
    rerender(<Probe name="evt" onPayload={b} />);
    act(() => emit('evt', { n: 1 }));
    expect(a).not.toHaveBeenCalled();
    expect(b).toHaveBeenCalledWith({ n: 1 });
    // subscription is reused; the listener map still has one entry
    expect(listeners.size).toBe(1);
  });

  it('unsubscribes on unmount', () => {
    const onPayload = vi.fn();
    const { unmount } = render(<Probe name="evt" onPayload={onPayload} />);
    unmount();
    expect(unsubscribeCalls).toBe(1);
    expect(listeners.size).toBe(0);
  });

  it('throttle coalesces a burst into a single delivery', () => {
    vi.useFakeTimers();
    const onPayload = vi.fn();
    render(<Probe name="evt" onPayload={onPayload} throttleMs={100} />);
    act(() => emit('evt', { n: 1 }));
    act(() => emit('evt', { n: 2 }));
    act(() => emit('evt', { n: 3 }));
    expect(onPayload).not.toHaveBeenCalled();
    act(() => {
      vi.advanceTimersByTime(100);
    });
    expect(onPayload).toHaveBeenCalledTimes(1);
    expect(onPayload).toHaveBeenCalledWith({ n: 3 });
  });

  it('flushes a pending throttle on unmount without firing it', () => {
    vi.useFakeTimers();
    const onPayload = vi.fn();
    const { unmount } = render(<Probe name="evt" onPayload={onPayload} throttleMs={100} />);
    act(() => emit('evt', { n: 1 }));
    unmount();
    act(() => {
      vi.advanceTimersByTime(100);
    });
    expect(onPayload).not.toHaveBeenCalled();
  });
});
