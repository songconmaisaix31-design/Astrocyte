import { cognitionApi } from '../../api/cognition';
import { useReadApi } from '../../hooks/useReadApi';
import { QueryState } from '../../components/QueryState';
import { EmptyState } from '../../components/EmptyState';
import { SectionCard } from '../../components/SectionCard';
import { Icon } from '../../components/DesignIcons';
import { EventTable } from './shared';
import { formatNumber } from './format';

export function AgentsView({ revision, query }: { revision: number; query: string }) {
  const state = useReadApi(signal => cognitionApi.agents(signal), { key: String(revision) });
  return <><div className="cg-section-heading"><div><h2>来源概览</h2><p>按存储来源分组，观察各来源的活动频率与最近事件。</p></div></div><QueryState state={state}>{data => {
    const groups = data.groups.filter(group => group.source.toLowerCase().includes(query.toLowerCase()));
    return groups.length ? <div className="cg-agent-list">{groups.map(group => <SectionCard key={group.source} title={group.source || '未标注来源'} padded><div className="cg-source-metrics"><Icon name="layers" size={30} />{(['today', 'days7', 'days30'] as const).map((key, index) => <div key={key}><span>{['今日', '近 7 日', '近 30 日'][index]}</span><strong data-testid={`agent-${group.source}-${key}`} data-value={group[key]}>{formatNumber(group[key])}</strong></div>)}</div><h3 className="cg-subheading">最近事件</h3><EventTable events={group.recent} /></SectionCard>)}</div> : <EmptyState title="没有来源活动记录" />;
  }}</QueryState></>;
}
