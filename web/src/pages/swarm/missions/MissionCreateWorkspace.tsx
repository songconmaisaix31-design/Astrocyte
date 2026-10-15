import { useState } from 'react';
import { proposalsApi } from '../../../api/proposals';
import { swarmApi } from '../../../api/swarm';
import type { components } from '../../../api/schema';
import { useReadApi } from '../../../hooks/useReadApi';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { QueryState } from '../../../components/QueryState';
import { CommandState } from '../../../components/CommandState';
import { Field } from '../../../components/DetailPanel';
import { commandError } from '../../../utils/commandError';
import styles from './MissionWorkspace.module.css';

export function MissionCreateWorkspace({ onCreated }: { onCreated: (id: string) => void }) {
  const proposals = useReadApi(signal => proposalsApi.proposals(signal));
  const [grantID, setGrantID] = useState('');
  const grant = useReadApi(signal => proposalsApi.approval(grantID, signal), { enabled: !!grantID, key: grantID });
  return <>
    <p>选择已经人工批准的固定提案和根目录。建立 ready 任务后，Agent 认领仍需核查当前授权及实际执行能力。</p>
    <button className="ac-button secondary compact" type="button" onClick={() => { proposals.retry(); if (grantID) grant.retry(); }}>刷新可用批准</button>
    <QueryState state={proposals}>{data => {
      const approved = data.items.filter(item => item.status === 'approved' && item.grant_id);
      return approved.length ? <label className={styles.fields}>已批准提案<select aria-label="已批准提案" value={grantID} disabled={proposals.loading || proposals.stale} onChange={event => setGrantID(event.target.value)}><option value="">请选择已批准的固定提案</option>{approved.map(item => <option key={item.id} value={item.grant_id ?? ''}>{item.content.title} · r{item.revision}</option>)}</select></label> : <p>暂无已批准提案，请先在共同工作区完成提案审阅和人工批准。</p>;
    }}</QueryState>
    {!!grantID && <QueryState state={grant}>{({ approval }) => <MissionFromGrant key={approval.id} approval={approval} disabled={proposals.loading || proposals.stale || !proposals.data || grant.loading || grant.stale} onCreated={onCreated} />}</QueryState>}
  </>;
}

function MissionFromGrant({ approval, disabled, onCreated }: { approval: components['schemas']['S2ApprovalGrantV1']; disabled: boolean; onCreated: (id: string) => void }) {
  const [root, setRoot] = useState('');
  const command = usePersistentCommand<components['schemas']['SwarmMissionEnvelopeV1']>(`swarm:create:${approval.id}`, () => ({ status: 'succeeded', message: '任务账本已保存' }), { failureMessage: cause => commandError(cause).message });
  const roots = approval.scope.roots ?? [];
  const blocked = disabled || approval.status !== 'active' || command.pending || command.uncertain;
  return <section aria-label="从固定批准建立任务">
    <Field label="Grant ID"><code>{approval.id}</code> · Epoch {approval.epoch}</Field><Field label="授权状态">{approval.status}</Field><Field label="批准有效期">{approval.expires_at}</Field><Field label="项目"><code>{approval.scope.project_id}</code></Field><Field label="提案固定版本"><code>{approval.proposal_id}</code> · r{approval.proposal_revision}</Field>
    <form onSubmit={event => { event.preventDefault(); if (blocked || !roots.includes(root)) return; void command.run(key => swarmApi.createMission({ schema_version: 1, request_id: key, expected_version: 0, grant_id: approval.id, epoch: approval.epoch, proposal_revision: approval.proposal_revision, project_id: approval.scope.project_id, root }, key), (result, outcome) => { if (outcome.status === 'succeeded') onCreated(result.mission.id); }); }}>
      <fieldset className={styles.fields} disabled={blocked}><legend>选择批准范围</legend><label>任务根目录<select required aria-label="任务根目录" value={root} onChange={event => setRoot(event.target.value)}><option value="">请选择一个已批准的根目录</option>{roots.map(value => <option key={value} value={value}>{value}</option>)}</select></label><button className="ac-button" type="submit" disabled={!root || !roots.includes(root)}>建立任务账本</button></fieldset>
    </form>
    <CommandState {...command} />
    {command.uncertain && <p role="alert">建立任务请求 {command.operationID} 的结果未知，保留原请求身份；关闭或刷新后仍禁止重发。</p>}
  </section>;
}
