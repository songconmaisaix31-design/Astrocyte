import type { ReactNode } from 'react';
import type { ReadApiState } from '../hooks/useReadApi';
import { EmptyState } from './EmptyState';
import { ErrorState } from './ErrorState';

/** The previous successful snapshot stays visible when a refresh fails. */
export function QueryState<T>({ state, children, empty, emptyTitle = '暂无记录' }: { state: ReadApiState<T>; children: (data: T) => ReactNode; empty?: (data: T) => boolean; emptyTitle?: string }) {
  if (!state.data) {
    if (state.loading) return <p role="status">加载中…</p>;
    if (state.error) return <ErrorState message={state.error} onRetry={state.retry} />;
    return <EmptyState title={emptyTitle} />;
  }
  return <>
    {state.loading && <p role="status">正在刷新…</p>}
    {state.stale && <div role="alert">数据可能已过期：暂未取得新记录，保留上次已加载内容。<button className="ac-button secondary compact" type="button" onClick={state.retry}>重试加载</button><details><summary>加载详情</summary>{state.error}</details></div>}
    {empty?.(state.data) ? <EmptyState title={emptyTitle} /> : children(state.data)}
  </>;
}
