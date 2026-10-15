import { useEffect, useRef, useState } from 'react';
import { useReadApi } from '../../../hooks/useReadApi';
import { listProposals } from '../../../api/client';
import { SectionCard } from '../../../components/SectionCard';
import { QueryState } from '../../../components/QueryState';
import { EmptyState } from '../../../components/EmptyState';
import { DetailPanel } from '../../../components/DetailPanel';
import { StatusBadge } from '../../../components/StatusBadge';
import { proposalStatusLabel } from '../../../utils/format';
import { matchesQuery } from '../../../utils/search';
import { ProposalEditor } from './ProposalEditor';
import { GoalEditor } from './GoalEditor';
import { GoalDetailWorkspace } from './GoalDetailWorkspace';
import { ProposalDetailWorkspace } from './ProposalDetailWorkspace';
import { proposalsApi } from '../../../api/proposals';
import styles from './ProposalWorkspace.module.css';

interface Props {
  query: string;
  onClearQuery: () => void;
  refreshToken: number;
  onChanged: () => void;
}

export function ProposalWorkspace({ query, onClearQuery, refreshToken, onChanged }: Props) {
  const proposals = useReadApi(signal => listProposals({ signal }), { key: 'proposals' });
  const goals = useReadApi(signal => proposalsApi.goals(signal), { key: 'workspace-goals' });
  const { retry: refreshProposals } = proposals;
  const { retry: refreshGoals } = goals;
  const previousRefresh = useRef(refreshToken);
  useEffect(() => {
    if (previousRefresh.current !== refreshToken) {
      previousRefresh.current = refreshToken;
      refreshProposals();
      refreshGoals();
    }
  }, [refreshProposals, refreshGoals, refreshToken]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [creatingGoal, setCreatingGoal] = useState(false);
  const [selectedGoalID, setSelectedGoalID] = useState<string | null>(null);
  const changed = () => { proposals.retry(); goals.retry(); onChanged(); };
  return <SectionCard title="提案" tabs={['proposals']} count={proposals.data?.items.length} actions={<div className={styles.actions}><button className="ac-button secondary compact" type="button" disabled={goals.loading || goals.stale || !goals.data} onClick={() => setCreatingGoal(true)}>创建目标</button><button className="ac-button compact" type="button" disabled={proposals.loading || proposals.stale || !proposals.data} onClick={() => setCreating(true)}>创建提案</button></div>}>
    <details><summary>目标与成功条件</summary><QueryState state={goals} empty={data => !data.items.length} emptyTitle="暂无目标">{data => <ul className={styles.list}>{data.items.map(goal => <li key={goal.id}><button className={styles.proposal} type="button" aria-label={`查看目标 ${goal.title}`} onClick={() => setSelectedGoalID(goal.id)}><strong>{goal.title}</strong><span>r{goal.revision} · {goal.success_criteria.join('；')}</span><span>优先级 {goal.priority} · 项目 {goal.project_id}</span></button></li>)}</ul>}</QueryState></details>
    <QueryState state={proposals} empty={data => !data.items.length} emptyTitle="暂无提案">{data => {
      const items = data.items.filter(proposal => matchesQuery(query, [proposal.id, proposal.content.title, proposal.content.goal.id, proposal.content.scope.project_id]));
      return items.length ? <ul className={styles.list}>{items.map(proposal => <li key={proposal.id}><button type="button" className={styles.proposal} aria-label={`查看提案 ${proposal.content.title}`} onClick={() => setSelectedId(proposal.id)}><span className={styles.header}><strong>{proposal.content.title}</strong><StatusBadge value={proposal.status} label={proposalStatusLabel(proposal.status)} /></span><span>固定目标：{proposal.content.goal.id || '未分配'} · r{proposal.content.goal.revision}</span><span className={styles.meta}>r{proposal.revision} · {proposal.content.scope.project_id ? '已分配项目' : '未分配项目'} · {proposal.content.deliverables?.length ?? 0} 项产出</span><span>查看范围、预算和停止条件</span></button></li>)}</ul> : <><EmptyState title="没有匹配的提案" description="请调整搜索词" /><button type="button" className="ac-button secondary compact" onClick={onClearQuery}>清除搜索</button></>;
    }}</QueryState>
    {selectedId && <DetailPanel title="提案详情" onClose={() => setSelectedId(null)}><ProposalDetailWorkspace key={selectedId} id={selectedId} onChanged={changed} /></DetailPanel>}
    {creating && <DetailPanel title="创建提案" onClose={() => setCreating(false)}><ProposalEditor onSaved={() => { setCreating(false); changed(); }} /></DetailPanel>}
    {creatingGoal && <DetailPanel title="创建目标" onClose={() => setCreatingGoal(false)}><GoalEditor onSaved={() => { setCreatingGoal(false); changed(); }} /></DetailPanel>}
    {selectedGoalID && <DetailPanel title="目标详情" onClose={() => setSelectedGoalID(null)}><GoalDetailWorkspace key={selectedGoalID} id={selectedGoalID} onChanged={changed} /></DetailPanel>}
  </SectionCard>;
}
