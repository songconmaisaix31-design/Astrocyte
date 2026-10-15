import { useState } from 'react';
import { swarmApi } from '../../../api/swarm';
import type { components } from '../../../api/schema';
import { useReadApi } from '../../../hooks/useReadApi';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { QueryState } from '../../../components/QueryState';
import { CommandState } from '../../../components/CommandState';
import { Field, FieldRow, MutedValue } from '../../../components/DetailPanel';
import { StatusBadge } from '../../../components/StatusBadge';
import { workItemStatusLabel } from '../../../utils/format';
import { commandError } from '../../../utils/commandError';
import { ArtifactDiff, ArtifactReviewActions } from './ArtifactReviewActions';
import { ResourceRecord } from './MissionRecords';
import styles from './MissionWorkspace.module.css';

export function ArtifactWorkspace({ id, missionId, disabled, onChanged }: { id: string; missionId: string; disabled: boolean; onChanged: () => void }) {
  const [version, setVersion] = useState(0);
  const [showDiff, setShowDiff] = useState(false);
  const current = useReadApi(signal => swarmApi.artifact(id, { mission_id: missionId }, signal), { key: `${missionId}:${id}` });
  const historical = useReadApi(signal => swarmApi.artifact(id, { mission_id: missionId, version }, signal), { enabled: version > 0, key: `${missionId}:${id}@${version}` });
  const state = version ? historical : current;
  const shown = state.data?.artifact;
  const diff = useReadApi(signal => swarmApi.artifactDiff(id, { mission_id: missionId, version: shown?.version }, signal), { enabled: showDiff && !!shown, key: `${missionId}:${id}@${shown?.version}:diff` });
  const command = usePersistentCommand<components['schemas']['SwarmArtifactEnvelopeV1']>(`swarm:${missionId}:artifact:${id}:review`, result => ({ status: 'succeeded', message: result.artifact.state === 'accepted' ? '产物已由人工采用' : '产物已退回' }), { failureMessage: cause => commandError(cause).message });
  const changed = () => { current.retry(); onChanged(); };
  const blocked = disabled || state.loading || state.stale || !shown || !!version || command.pending || command.uncertain;
  return <section aria-label="产物详情" className={styles.record}>
    <h4>产物与人工验收</h4>
    <div className={styles.actions}><button type="button" className="ac-button secondary compact" onClick={() => { current.retry(); if (version) historical.retry(); if (showDiff) diff.retry(); }}>刷新产物记录</button>
      {current.data && <label>查看产物版本<select aria-label="查看产物版本" value={version} onChange={event => { setVersion(Number(event.target.value)); setShowDiff(false); }}><option value={0}>当前版本 v{current.data.artifact.version}</option>{Array.from({ length: current.data.artifact.version }, (_, index) => index + 1).map(value => <option key={value} value={value}>固定版本 v{value}</option>)}</select></label>}
    </div>
    <QueryState state={state}>{({ artifact }) => <>
      {!!version && <p>正在查看不可变历史产物版本，人工验收使用当前记录。</p>}
      <Field label="产物 ID"><code>{artifact.id}</code></Field><FieldRow><Field label="状态"><StatusBadge value={artifact.state} label={workItemStatusLabel(artifact.state)} /></Field><Field label="版本">v{artifact.version}</Field></FieldRow>
      <Field label="所属工作项"><code>{artifact.work_item_id}</code></Field><Field label="提交时间">{artifact.created_at}</Field>
      <Field label="产物 Commit"><code>{artifact.commit}</code></Field><Field label="基准 Commit"><code>{artifact.base_commit}</code></Field>
      <Field label="产物对象引用"><code>{artifact.object_ref}</code></Field><Field label="检查结果引用"><code>{artifact.checks_ref}</code></Field><Field label="项目检查引用"><code>{artifact.project_checks_ref}</code></Field>
      <Field label="实际修改路径">{artifact.changed_paths?.join('；') || <MutedValue>无</MutedValue>}</Field><Field label="固定输入版本">{artifact.inputs?.map(input => `${input.id} · ${input.revision}`).join('；') || <MutedValue>无</MutedValue>}</Field>
      <Field label="原生执行身份"><code>{artifact.operation_id}</code> · <code>{artifact.session_id}</code> · fencing token {artifact.fence}</Field><Field label="上下文包 ID">{artifact.context_packet_id || <MutedValue>未提供</MutedValue>}</Field>
      <Field label="执行结果">{artifact.execution_outcome || <MutedValue>未知</MutedValue>}</Field><Field label="执行失败代码">{artifact.execution_failure_code || <MutedValue>无</MutedValue>}</Field>
      {artifact.usage ? <ResourceRecord label="实际用量" resources={artifact.usage} /> : <Field label="实际用量"><MutedValue>未知</MutedValue></Field>}
      <Field label="人工采用记录">{artifact.acceptance_id ? <><code>{artifact.acceptance_id}</code> · {artifact.accepted_by} · {artifact.accepted_actor_kind}</> : <MutedValue>尚未采用</MutedValue>}</Field><Field label="人工复核">{artifact.reviewed_by ? `${artifact.reviewed_by} · ${artifact.reviewed_at} · ${artifact.review_reason}` : <MutedValue>暂无复核记录</MutedValue>}</Field><Field label="经验候选状态">{artifact.experience_status || <MutedValue>尚未形成</MutedValue>}</Field><Field label="经验引用">{artifact.experience_id || <MutedValue>无</MutedValue>}</Field>
      <button className="ac-button secondary" type="button" disabled={state.loading || state.stale} onClick={() => setShowDiff(true)}>查看 diff</button>
      {showDiff && <QueryState state={diff}>{data => <ArtifactDiff diff={data.diff} commit={data.commit} baseCommit={data.base_commit} />}</QueryState>}
      <ArtifactReviewActions key={`${artifact.id}@${artifact.version}`} state={artifact.state} disabled={blocked} pending={command.pending}
        onAccept={() => { void command.run(key => swarmApi.acceptArtifact(id, { schema_version: 1, request_id: key, expected_version: artifact.version, mission_id: missionId }, key), (_, outcome) => { if (outcome.status === 'succeeded') changed(); }); }}
        onReject={reason => { void command.run(key => swarmApi.rejectArtifact(id, { schema_version: 1, request_id: key, expected_version: artifact.version, mission_id: missionId, reason }, key), (_, outcome) => { if (outcome.status === 'succeeded') changed(); }); }} />
    </>}</QueryState>
    <CommandState {...command} />{command.uncertain && <p role="alert">人工验收请求 {command.operationID} 的结果未知；当前产物验收保持锁定，重新读取不会自动重发。</p>}
  </section>;
}
