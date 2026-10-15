import { useState } from 'react';
import { proposalsApi, type Proposal } from '../../../api/proposals';
import type { components } from '../../../api/schema';
import { useReadApi } from '../../../hooks/useReadApi';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { QueryState } from '../../../components/QueryState';
import { CommandState } from '../../../components/CommandState';
import { commandError } from '../../../utils/commandError';
import { ProposalRevision, ProposalRevisionDiff } from './ProposalRevision';
import { ProposalReviewActions } from './ProposalReviewActions';
import { ProposalEditor } from './ProposalEditor';
import { ApprovalPanel } from './ApprovalPanel';
import { FixedGoalReference } from './FixedGoalReference';
import styles from './ProposalWorkspace.module.css';

type CommandResult = components['schemas']['S2ProposalResultV1'] | components['schemas']['S2ApprovalResultV1'];

export function ProposalDetailWorkspace({ id, onChanged }: { id: string; onChanged: () => void }) {
  const current = useReadApi(signal => proposalsApi.proposal(id, undefined, signal), { key: id });
  const [revision, setRevision] = useState(0);
  const [editing, setEditing] = useState<Proposal | null>(null);
  const historical = useReadApi(signal => proposalsApi.proposal(id, revision, signal), { enabled: revision > 0, key: `${id}@${revision}` });
  const state = revision ? historical : current;
  const shown = state.data?.proposal;
  const previousRevision = shown ? shown.revision - 1 : 0;
  const previous = useReadApi(signal => proposalsApi.proposal(id, previousRevision, signal), { enabled: previousRevision > 0, key: `${id}@previous:${previousRevision}` });
  const command = usePersistentCommand<CommandResult>(`proposal:${id}:review`, result => ({ status: 'succeeded', message: 'proposal' in result ? '提案状态已保存' : '当前固定版本已批准' }), { failureMessage: cause => commandError(cause).message });
  const changed = () => { current.retry(); onChanged(); };
  const blocked = !!revision || state.loading || state.stale || !state.data || command.pending || command.uncertain;
  return <>
    <div className={styles.actions}><button className="ac-button secondary compact" type="button" onClick={() => { current.retry(); if (revision) historical.retry(); }}>刷新当前提案</button>
      {current.data && <label>查看提案版本<select aria-label="查看提案版本" value={revision} onChange={event => setRevision(Number(event.target.value))}><option value={0}>当前版本 r{current.data.proposal.revision}</option>{Array.from({ length: current.data.proposal.revision }, (_, index) => index + 1).map(value => <option key={value} value={value}>固定版本 r{value}</option>)}</select></label>}
    </div>
    <QueryState state={state}>{({ proposal }) => <>
      {!!revision && <p>正在查看固定历史版本；当前执行内容通过新版本修改。</p>}
      <ProposalRevision proposal={proposal} />
      <FixedGoalReference key={`${proposal.content.goal.id}@${proposal.content.goal.revision}`} goal={proposal.content.goal} />
      {previousRevision > 0 && <QueryState state={previous}>{({ proposal: before }) => <ProposalRevisionDiff before={before} after={proposal} />}</QueryState>}
      <div className={styles.actions}><button className="ac-button secondary" type="button" disabled={blocked || !!editing} onClick={() => setEditing(proposal)}>编辑新版本</button></div>
      <ProposalReviewActions key={`${proposal.id}:${proposal.revision}`} status={proposal.status} revision={proposal.revision} disabled={blocked || !!editing || !proposal.content.goal.id || !proposal.content.scope.project_id} pending={command.pending}
        onSubmit={() => { void command.run(key => proposalsApi.submitProposal(id, { schema_version: 1, request_id: key, expected_version: proposal.version }, key), (_, outcome) => { if (outcome.status === 'succeeded') changed(); }); }}
        onDecline={() => { void command.run(key => proposalsApi.declineProposal(id, { schema_version: 1, request_id: key, expected_version: proposal.version }, key), (_, outcome) => { if (outcome.status === 'succeeded') changed(); }); }}
        onApprove={expires_at => { void command.run(key => proposalsApi.approveProposal(id, { schema_version: 1, request_id: key, expected_version: proposal.version, expected_revision: proposal.revision, expires_at }, key), (_, outcome) => { if (outcome.status === 'succeeded') changed(); }); }} />
      <ApprovalPanel key={proposal.grant_id || 'none'} id={proposal.grant_id ?? ''} onChanged={changed} />
    </>}</QueryState>
    <CommandState {...command} />{command.uncertain && <p role="alert">提案请求 {command.operationID} 的结果未知；只读取记录核对，重新打开或刷新后仍禁止重发。</p>}
    {editing && <section aria-label="编辑提案新版本"><h4>编辑新版本</h4><ProposalEditor key={editing.id} initial={editing} latest={current.data?.proposal} disabled={current.loading || current.stale || !current.data} onSaved={() => { setEditing(null); setRevision(0); changed(); }} /><button className="ac-button secondary compact" type="button" onClick={() => setEditing(null)}>关闭编辑</button></section>}
  </>;
}
