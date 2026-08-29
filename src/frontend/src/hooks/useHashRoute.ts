import { useEffect, useState } from 'react';

export interface RouteMatch {
  // The path portion of the hash, e.g. "/backup".
  path: string;
  // Parsed query string. Always defined (empty if no query).
  query: URLSearchParams;
}

function readHash(): RouteMatch {
  const raw = window.location.hash.replace(/^#/, '') || '/';
  const [path, qs = ''] = raw.split('?');
  return { path: path || '/', query: new URLSearchParams(qs) };
}

// ponytail: tiny hash router. Five routes does not justify react-router.
// The hook just exposes the current path+query and re-renders on
// `hashchange`. Components navigate by writing to `window.location.hash`.
export function useHashRoute(): RouteMatch {
  const [route, setRoute] = useState<RouteMatch>(readHash);

  useEffect(() => {
    const onChange = () => setRoute(readHash());
    window.addEventListener('hashchange', onChange);
    return () => window.removeEventListener('hashchange', onChange);
  }, []);

  return route;
}

export function navigate(path: string, query?: Record<string, string>): void {
  const qs = query ? `?${new URLSearchParams(query).toString()}` : '';
  window.location.hash = `${path}${qs}`;
}
