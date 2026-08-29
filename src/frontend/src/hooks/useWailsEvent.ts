import { useEffect, useRef } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';

type Options = {
  // ponytail: trailing-edge throttle; if the event fires faster than
  // throttleMs, only the latest payload reaches React state. 0 = no throttle.
  throttleMs?: number;
};

export function useWailsEvent<T = unknown>(
  name: string,
  handler: (payload: T) => void,
  { throttleMs = 0 }: Options = {},
): void {
  const handlerRef = useRef(handler);
  handlerRef.current = handler;

  const latestRef = useRef<T | undefined>(undefined);
  const timerRef = useRef<number | null>(null);

  useEffect(() => {
    const flush = () => {
      timerRef.current = null;
      if (latestRef.current !== undefined) {
        handlerRef.current(latestRef.current);
        latestRef.current = undefined;
      }
    };

    const unsubscribe = EventsOn(name, (...data: unknown[]) => {
      const payload = data[0] as T;
      if (throttleMs <= 0) {
        handlerRef.current(payload);
        return;
      }
      latestRef.current = payload;
      if (timerRef.current === null) {
        timerRef.current = window.setTimeout(flush, throttleMs);
      }
    });

    return () => {
      if (timerRef.current !== null) {
        window.clearTimeout(timerRef.current);
        timerRef.current = null;
      }
      latestRef.current = undefined;
      unsubscribe();
    };
  }, [name, throttleMs]);
}
