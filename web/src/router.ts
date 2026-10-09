/**
 * Lightweight history-based router.
 * No external deps — uses the native History API.
 */
import { useState, useEffect, useCallback } from 'react';

export interface Route {
  path: string;
  params: Record<string, string>;
  query: URLSearchParams;
}

export function useRoute(): Route {
  const [route, setRoute] = useState<Route>(parseRoute());

  useEffect(() => {
    const onPop = () => setRoute(parseRoute());
    window.addEventListener('popstate', onPop);
    return () => window.removeEventListener('popstate', onPop);
  }, []);

  return route;
}

export function navigate(path: string, replace = false): void {
  if (replace) {
    window.history.replaceState(null, '', path);
  } else {
    window.history.pushState(null, '', path);
  }
  window.dispatchEvent(new PopStateEvent('popstate'));
}

function parseRoute(): Route {
  const url = new URL(window.location.href);
  const path = url.pathname || '/';
  const query = url.searchParams;
  return { path, params: {}, query };
}

export function useNavigate() {
  return useCallback((path: string, replace = false) => navigate(path, replace), []);
}
