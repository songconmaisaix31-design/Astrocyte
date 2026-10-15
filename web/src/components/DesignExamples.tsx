import { Icon } from './DesignIcons';
import { EmptyState } from './EmptyState';

export function ResearchRoutes() {
  return <div className="ac-panel"><p className="ac-rail-note">研究路线尚未接入。当前接口没有分支与依赖关系，项目详情保留真实路径和版本。</p></div>;
}

export function DesignTimeline() {
  return <EmptyState title="事件时间线尚未接入" description="当前 API 只返回任务快照；没有事件记录时不生成时间或协作动态。" />;
}

export function DesignGraph() {
  return <div className="ac-panel"><div className="ac-panel-heading"><div><h3>协作拓扑</h3><p>拓扑需有认领、交接或复核事件支撑</p></div><button type="button" className="ac-button compact secondary" disabled title="当前 API 未提供协作事件，拓扑尚未接入"><Icon name="swarm" size={16} />拓扑视图</button></div><p className="ac-panel-note">拓扑尚未接入；下面显示真实任务与工作项快照。</p></div>;
}
