import { Component, lazy, Suspense, type ComponentProps, type ReactNode } from 'react';
import type { MissionDAGView } from './MissionDAGView';

const LazyMissionDAGView = lazy(() => import('./MissionDAGView').then(module => ({ default: module.MissionDAGView })));

class DAGErrorBoundary extends Component<{ children: ReactNode }, { failed: boolean; error: unknown }> {
  state = { failed: false, error: null as unknown };

  static getDerivedStateFromError(error: unknown) {
    return { failed: true, error };
  }

  render() {
    if (this.state.failed) return <section aria-label="任务依赖加载错误">
      <p role="alert">任务依赖视图无法显示：{this.state.error instanceof Error ? this.state.error.message : String(this.state.error)}</p>
      <p>请确认服务可用，然后重新加载页面。</p>
      <button className="ac-button secondary compact" type="button" onClick={() => window.location.reload()}>重新加载页面</button>
    </section>;
    return this.props.children;
  }
}

export function MissionDAGRegion(props: ComponentProps<typeof MissionDAGView>) {
  return <DAGErrorBoundary><Suspense fallback={<p role="status">正在加载任务依赖视图…</p>}><LazyMissionDAGView {...props} /></Suspense></DAGErrorBoundary>;
}
