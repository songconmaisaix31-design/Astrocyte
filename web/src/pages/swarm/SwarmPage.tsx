import { useMemo, useState } from 'react';
import { DevelopmentSpacePanel } from './DevelopmentSpacePanel';
import { useMissions } from '../../hooks/useReadApi';
import { SectionCard } from '../../components/SectionCard';
import { QueryState } from '../../components/QueryState';
import { EmptyState } from '../../components/EmptyState';
import { DetailPanel } from '../../components/DetailPanel';
import { PageFrame, RailSummary } from '../../components/PageFrame';
import { DesignGraph } from '../../components/DesignExamples';
import type { SwarmMission } from '../../api/swarm';
import { matchesQuery } from '../../utils/search';
import { MissionList } from './missions/MissionRecords';
import { MissionDetailWorkspace } from './missions/MissionDetailWorkspace';
import { MissionCreateWorkspace } from './missions/MissionCreateWorkspace';
import styles from './SwarmPage.module.css';

export { MissionList, MissionDetail, WorkItemList, WorkItemDetail } from './missions/MissionRecords';

const EMPTY_MISSIONS: SwarmMission[] = [];

export function SwarmPage({ query }: { query: string }) {
  const missions = useMissions();
  const [selected, setSelected] = useState<{ kind: 'mission'; id: string } | { kind: 'create' } | null>(null);
  const items = useMemo(() => (missions.data?.items ?? EMPTY_MISSIONS).filter(item => matchesQuery(query, [item.id, item.grant.goal, item.grant.id, item.grant.project_id, item.reason])), [missions.data, query]);
  const workItemCount = items.reduce((count, item) => count + item.work_item_count, 0);
  const count = (value: number) => missions.loading && !missions.data ? '加载中…' : missions.error && !missions.data ? '无法获取' : value;
  return <PageFrame section="swarm" title="蜂群空间" subtitle="查看批准范围内的任务、执行记录和待验收产物。" onRefresh={missions.retry} rail={<RailSummary title="执行概览" rows={[{ label: '任务', value: count(items.length) }, { label: '工作项', value: count(workItemCount) }, { label: '产物引用', value: '按任务查看' }]} note="提交产物后等待人工验收；当前授权由服务逐次核查。" />}>
    <div className={styles.grid}>
      <DevelopmentSpacePanel query={query} />
      <SectionCard title="任务时间线" tabs={['timeline']} count={missions.data ? items.length : undefined} actions={<button className="ac-button compact" type="button" disabled={missions.loading || missions.stale || !missions.data} onClick={() => setSelected({ kind: 'create' })}>从批准建立任务</button>}>
        <p className={styles.itemDesc}>打开任务查看真实事件时间线、当前工作项、租约与产物。建立任务保存批准账本，执行由已绑定的 Agent 另行认领。</p>
        <QueryState state={missions} empty={data => !data.items.length} emptyTitle="暂无任务">{() => items.length ? <MissionList items={items} onSelect={item => setSelected({ kind: 'mission', id: item.id })} /> : <EmptyState title="没有匹配的任务" description="请调整搜索词" />}</QueryState>
      </SectionCard>
      <SectionCard title="协作视图" tabs={['timeline']}><DesignGraph /></SectionCard>
    </div>
    {selected?.kind === 'mission' && <DetailPanel title="任务详情" onClose={() => setSelected(null)}><MissionDetailWorkspace key={selected.id} id={selected.id} onChanged={missions.retry} /></DetailPanel>}
    {selected?.kind === 'create' && <DetailPanel title="从批准建立任务" onClose={() => setSelected(null)}><MissionCreateWorkspace onCreated={id => { missions.retry(); setSelected({ kind: 'mission', id }); }} /></DetailPanel>}
  </PageFrame>;
}
