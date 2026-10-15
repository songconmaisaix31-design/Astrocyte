/** 读取请求保留加载、错误与过期状态；已卸载或被取代的请求不再更新页面。 */
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

  // 保存最新 fetcher，避免函数引用变化触发重复请求。
  const fetcherRef = useRef(fetcher);
  useEffect(() => { fetcherRef.current = fetcher; });

  const fetchIdRef = useRef(0);
  const mountedRef = useRef(false);
  const abortRef = useRef<AbortController | null>(null);

  // 状态更新在 await 后执行，避免在 effect 内同步更新。
  const execute = useCallback(async () => {
    const id = ++fetchIdRef.current;
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    // 后续代码进入微任务队列。
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

  // 挂载时读取当前资源。
  useEffect(() => {
    mountedRef.current = true;
    if (enabled) {
      // 切换资源时先清除先前详情。
      void Promise.resolve().then(() => {
        if (!mountedRef.current) return;
        setState({ data: null, loading: true, error: null, stale: false, key });
        void execute();
      });
    }
    return () => {
      mountedRef.current = false;
      abortRef.current?.abort();
      // eslint-disable-next-line react-hooks/exhaustive-deps -- 使未完成请求失效
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

// 各接口复用同一个生成客户端。

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

export function useLocalAgents(enabled = true) {
  return useReadApi(signal => readApi.listLocalAgents({ signal }), { enabled });
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
