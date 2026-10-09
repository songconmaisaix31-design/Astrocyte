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

export function formatError(err: unknown): string {
  if (err instanceof ApiError) return `API ${err.status} [${err.code}]: ${err.message} · ${err.detail.error.required_action || '请检查服务后重试'}（请求 ${err.requestId}）`;
  if (err instanceof Error) return err.message;
  return '未知错误';
}

export function useReadApi<T>(
  fetcher: (signal?: AbortSignal) => Promise<T>,
  options: { enabled?: boolean; key?: string; refreshMs?: number } = {},
): ReadApiState<T> {
  const { enabled = true, key = '', refreshMs } = options;
  const [state, setState] = useState<{
    data: T | null;
    loading: boolean;
    error: string | null;
    stale: boolean;
    key: string;
  }>({ data: null, loading: true, error: null, stale: false, key });

  // Keep fetcher in a ref so we don't re-run effects when it changes identity.
  const fetcherRef = useRef(fetcher);
  useEffect(() => { fetcherRef.current = fetcher; });

  const fetchIdRef = useRef(0);
  const mountedRef = useRef(false);
  const abortRef = useRef<AbortController | null>(null);

  // Stable fetch executor. All setState calls follow an `await` boundary,
  // so they execute as microtask continuations, not synchronously in the
  // effect body or useCallback body.
  const execute = useCallback(async () => {
    const id = ++fetchIdRef.current;
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    // True async boundary: ensures all subsequent code runs on the
    // microtask queue, not synchronously within the caller's frame.
    await Promise.resolve();
    if (!mountedRef.current || !enabled || id !== fetchIdRef.current) return;

    setState(prev => ({ data: prev.key === key ? prev.data : null, loading: true, error: null, stale: prev.key === key && prev.stale, key }));
    try {
      const result = await fetcherRef.current(controller.signal);
      if (mountedRef.current && id === fetchIdRef.current) {
        setState({ data: result, loading: false, error: null, stale: false, key });
      }
    } catch (err: unknown) {
      if (mountedRef.current && id === fetchIdRef.current) {
        setState(prev => ({ ...prev, loading: false, error: formatError(err), stale: !!prev.data }));
      }
    }
  }, [enabled, key]);

  // Initial fetch on mount.
  useEffect(() => {
    mountedRef.current = true;
    if (enabled) {
      // A different resource must never briefly display the old detail.
      void Promise.resolve().then(() => {
        if (!mountedRef.current) return;
        setState({ data: null, loading: true, error: null, stale: false, key });
        void execute();
      });
    }
    return () => {
      mountedRef.current = false;
      abortRef.current?.abort();
      // eslint-disable-next-line react-hooks/exhaustive-deps -- invalidate in-flight requests
      ++fetchIdRef.current;
    };
  }, [execute, enabled, key]);

  useEffect(() => {
    if (!enabled || !refreshMs) return;
    const timer = window.setInterval(() => { void execute(); }, refreshMs);
    return () => window.clearInterval(timer);
  }, [enabled, refreshMs, execute]);

  const retry = useCallback(() => {
    if (!mountedRef.current || !enabled) return;
    setState(prev => ({ ...prev, loading: true, error: null }));
    void execute();
  }, [execute, enabled]);

  return { data: enabled && state.key === key ? state.data : null, error: enabled && state.key === key ? state.error : null, stale: enabled && state.key === key && state.stale, loading: enabled && (state.key !== key || state.loading), retry };
}

// ── Pre-bound fetchers for each endpoint ──

export function useHealth() {
  return useReadApi(() => readApi.getHealth());
}

export function useFoundation() {
  return useReadApi(() => readApi.getFoundation());
}

export function useMaterials(enabled = true) {
  return useReadApi(signal => readApi.listMaterials({ signal }), { enabled });
}

export function useOpportunities(enabled = true) {
  return useReadApi(signal => readApi.listOpportunities({ signal }), { enabled });
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
