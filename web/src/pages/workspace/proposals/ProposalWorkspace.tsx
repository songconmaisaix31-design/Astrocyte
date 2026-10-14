import { useState } from 'react';
import { useReadApi } from '../../../hooks/useReadApi';
import { listProposals } from '../../../api/client';
import { SectionCard } from '../../../components/SectionCard';
import { QueryState } from '../../../components/QueryState';
import { EmptyState } from '../../../components/EmptyState';
import { DetailPanel } from '../../../components/DetailPanel';
import { StatusBadge } from '../../../components/StatusBadge';
import { proposalStatusLabel } from '../../../utils/format';
import { matchesQuery } from '../../../utils/search';
import { ProposalRevision } from './ProposalRevision';
import styles from './ProposalWorkspace.module.css';

interface Props {
  query: string;
  onClearQuery: () => void;
  refreshToken: number;
  onChanged: () => void;
}

export function ProposalWorkspace({ query, onClearQuery, refreshToken }: Props) {
  const proposals = useReadApi(signal => listProposals({ signal }), { key: `proposals:${refreshToken}` });
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const selected = proposals.data?.items.find(proposal => proposal.id === selectedId);
  return <SectionCard title="提案" tabs={['proposals']} count={proposals.data?.items.length}>
    <QueryState state={proposals} empty={data => !data.items.length} emptyTitle="暂无提案">{data => {
      const items = data.items.filter(proposal => matchesQuery(query, [proposal.id, proposal.title, proposal.goal]));
      return items.length ? <ul className={styles.list}>{items.map(proposal => <li key={proposal.id}><button type="button" className={styles.proposal} onClick={() => setSelectedId(proposal.id)}><span className={styles.header}><strong>{proposal.title}</strong><StatusBadge value={proposal.status} label={proposalStatusLabel(proposal.status)} /></span><span>{proposal.goal}</span><span className={styles.meta}>r{proposal.revision} · {proposal.scope.project_ids.length} 个项目 · {proposal.deliverables.length} 项产出</span><span>查看范围、预算和停止条件</span></button></li>)}</ul> : <><EmptyState title="没有匹配的提案" description="请调整搜索词" /><button type="button" className="ac-button secondary compact" onClick={onClearQuery}>清除搜索</button></>;
    }}</QueryState>
    {selected && <DetailPanel title="提案详情" onClose={() => setSelectedId(null)}><ProposalRevision proposal={selected} /></DetailPanel>}
  </SectionCard>;
}
