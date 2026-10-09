/**
 * useReadApi — lightweight data fetching hook.
 * Wraps createReadApi with loading/error/retry states.
 * Does NOT silently fall back to fixtures on error.
 */
import { useState, useEffect, useCallback, useRef } from 'react';
import { createReadApi, ApiError, api } from '../api/client';

const readApi = createReadApi(api);

export interface ReadApiState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  stale: boolean;
  retry: () => void;
}

export function useReadApi<T>(
  fetcher: () => Promise<T>,
  deps: unknown[] = [],
): ReadApiState<T> {
  const [state, setState] = useState<{
    data: T | null;
    loading: boolean;
    error: string | null;
    stale: boolean;
  }>({ data: null, loading: true, error: null, stale: false });

  const mountedRef = useRef(true);
  const fetchIdRef = useRef(0);

  const execute = useCallback(async () => {
    const fetchId = ++fetchIdRef.current;
    setState(prev => ({ ...prev, loading: true, error: null }));
    try {
      const result = await fetcher();
      if (mountedRef.current && fetchId === fetchIdRef.current) {
        setState({ data: result, loading: false, error: null, stale: false });
      }
    } catch (err: unknown) {
      if (mountedRef.current && fetchId === fetchIdRef.current) {
        const msg = err instanceof ApiError
          ? `API ${err.status}: ${err.message}`
          : err instanceof Error
            ? err.message
            : '未知错误';
        // Retain prior data; mark stale so UI can show stale banner
        setState(prev => ({ ...prev, loading: false, error: msg, stale: !!prev.data }));
      }
    }
  }, deps); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    mountedRef.current = true;
    execute();
    return () => { mountedRef.current = false; };
  }, [execute]);

  const retry = useCallback(() => { execute(); }, [execute]);

  return { ...state, retry };
}

// ── Pre-bound fetchers for each endpoint ──

export function useHealth() {
  return useReadApi(() => readApi.getHealth());
}

export function useFoundation() {
  return useReadApi(() => readApi.getFoundation());
}

export function useMaterials() {
  return useReadApi(() => readApi.listMaterials());
}

export function useOpportunities() {
  return useReadApi(() => readApi.listOpportunities());
}

export function useProjects() {
  return useReadApi(() => readApi.listProjects());
}

export function useProposals() {
  return useReadApi(() => readApi.listProposals());
}

export function useSessions() {
  return useReadApi(() => readApi.listSessions());
}

export function useMissions() {
  return useReadApi(() => readApi.listMissions());
}
