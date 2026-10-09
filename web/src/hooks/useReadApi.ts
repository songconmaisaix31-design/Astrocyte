/**
 * useReadApi — lightweight data fetching hook.
 * Wraps createReadApi with loading/error/retry/stale states.
 * Does NOT silently fall back to fixtures on error.
 *
 * Lifecycle:
 *   - mountedRef + fetchIdRef guard every setState so that unmounted or
 *     superseded request completions never write state.
 *   - execute() begins with `await Promise.resolve()` so all subsequent
 *     code (including the first setState) runs on the microtask queue,
 *     not synchronously inside the effect body.
 *   - The effect's cleanup sets mounted=false and bumps fetchId to
 *     invalidate any in-flight request from the previous mount cycle
 *     (important under StrictMode double-mount).
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

function formatError(err: unknown): string {
  if (err instanceof ApiError) return `API ${err.status}: ${err.message}`;
  if (err instanceof Error) return err.message;
  return '未知错误';
}

export function useReadApi<T>(
  fetcher: () => Promise<T>,
): ReadApiState<T> {
  const [state, setState] = useState<{
    data: T | null;
    loading: boolean;
    error: string | null;
    stale: boolean;
  }>({ data: null, loading: true, error: null, stale: false });

  // Keep fetcher in a ref so we don't re-run effects when it changes identity.
  const fetcherRef = useRef(fetcher);
  useEffect(() => { fetcherRef.current = fetcher; });

  const fetchIdRef = useRef(0);
  const mountedRef = useRef(false);

  // Stable fetch executor. All setState calls follow an `await` boundary,
  // so they execute as microtask continuations, not synchronously in the
  // effect body or useCallback body.
  const execute = useCallback(async () => {
    const id = fetchIdRef.current;
    // True async boundary: ensures all subsequent code runs on the
    // microtask queue, not synchronously within the caller's frame.
    await Promise.resolve();
    if (!mountedRef.current || id !== fetchIdRef.current) return;

    setState(prev => ({ data: prev.data, loading: true, error: null, stale: prev.stale }));
    try {
      const result = await fetcherRef.current();
      if (mountedRef.current && id === fetchIdRef.current) {
        setState({ data: result, loading: false, error: null, stale: false });
      }
    } catch (err: unknown) {
      if (mountedRef.current && id === fetchIdRef.current) {
        setState(prev => ({ ...prev, loading: false, error: formatError(err), stale: !!prev.data }));
      }
    }
  }, []);

  // Initial fetch on mount.
  useEffect(() => {
    mountedRef.current = true;
    void execute();
    return () => {
      mountedRef.current = false;
      // eslint-disable-next-line react-hooks/exhaustive-deps -- invalidate in-flight requests
      ++fetchIdRef.current;
    };
  }, [execute]);

  const retry = useCallback(() => {
    if (!mountedRef.current) return;
    ++fetchIdRef.current;
    setState(prev => ({ ...prev, loading: true, error: null }));
    void execute();
  }, [execute]);

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
