import { useState } from 'react';
import { workspaceContextApi, type WorkspaceContextPacket } from '../../../api/workspaceContext';
import { useReadApi } from '../../../hooks/useReadApi';
import { QueryState } from '../../../components/QueryState';
import { Field, FieldRow, MutedValue } from '../../../components/DetailPanel';
import styles from './ProposalWorkspace.module.css';

export function ContextPacketReference({ id, projectId }: { id: string; projectId?: string }) {
  const [expanded, setExpanded] = useState(false);
  const packet = useReadApi(signal => workspaceContextApi.packet(id, signal), { key: id, enabled: expanded });
  return <section className={styles.inputs} aria-label={`固定上下文包 ${id}`}>
    <p>上下文包 <code>{id}</code></p>
    <button className="ac-button secondary compact" type="button" aria-expanded={expanded} onClick={() => setExpanded(value => !value)}>{expanded ? '收起固定上下文' : '展开固定上下文'}</button>
    {expanded && <>
      <button className="ac-button secondary compact" type="button" onClick={packet.retry}>重新读取固定上下文</button>
      <QueryState state={packet}>{({ context_packet }) => {
        if (context_packet.id !== id || projectId && context_packet.project_id !== projectId) throw new Error('上下文包与当前引用或项目身份不一致');
        return <PacketRecord packet={context_packet} />;
      }}</QueryState>
    </>}
  </section>;
}

function PacketRecord({ packet }: { packet: WorkspaceContextPacket }) {
  const snapshot = packet.snapshot;
  const notes: [string, string[] | null | undefined][] = [
    ['关键决定', packet.decisions], ['已有结果', packet.results], ['检查记录', packet.checks],
    ['失败记录', packet.failures], ['未解决问题', packet.open_questions], ['下一步', packet.next_steps],
  ];
  return <div className={styles.revision}>
    <p>固定记录保存创建时的来源、权限和项目快照；新的上下文通过后继包记录。</p>
    <Field label="上下文包对象引用"><code>{packet.object_ref}</code></Field>
    <Field label="创建时间">{packet.created_at}</Field>
    <Field label="前序上下文包">{packet.predecessor_id ? <ContextPacketReference id={packet.predecessor_id} projectId={packet.project_id} /> : <MutedValue>无</MutedValue>}</Field>
    <Field label="固定批准"><code>{packet.grant_id}</code> · Epoch {packet.grant_epoch}</Field>
    <Field label="固定提案"><code>{packet.proposal_id}</code> · r{packet.proposal_revision}</Field>
    <Field label="固定目标"><code>{packet.goal.id}</code> · r{packet.goal.revision} · {packet.goal.title}</Field>
    <Field label="目标成功条件"><Values items={packet.goal.success_criteria} /></Field>
    <FieldRow><Field label="项目"><code>{packet.project_id}</code></Field><Field label="权限版本">r{packet.project_settings_revision}</Field></FieldRow>
    <Field label="环境"><code>{snapshot.environment_id}</code></Field>
    <Field label="仓库及工作树"><code>{snapshot.repo_id}</code> · <code>{snapshot.worktree_id}</code></Field>
    <Field label="固定 HEAD"><code>{snapshot.head}</code></Field>
    <Field label="Index digest"><code>{snapshot.index_digest}</code></Field>
    <Field label="工作目录 digest"><code>{snapshot.worktree_digest}</code></Field>
    <Field label="快照 generation"><code>{snapshot.generation}</code> · {snapshot.observed_at}</Field>
    <Field label="固定资料来源">{packet.sources.length ? <ul className={styles.values}>{packet.sources.map(source => <li key={`${source.reference.material_id}@${source.reference.revision}`}>
      {source.title} · <code>{source.reference.material_id}</code> · r{source.reference.revision}
      <span className={styles.sourceLocator}>对象引用：{source.object_ref}</span>
      <Values items={source.locations} />
    </li>)}</ul> : <MutedValue>无</MutedValue>}</Field>
    <Field label="允许根目录"><Values items={packet.scope.roots} /></Field>
    <Field label="允许修改路径"><Values items={packet.scope.write_paths} /></Field>
    <Field label="允许动作"><Values items={packet.scope.actions} /></Field>
    <Field label="允许工具"><Values items={packet.scope.tools} /></Field>
    <Field label="允许外发目标"><Values items={packet.scope.external_targets} /></Field>
    <Field label="最大并发数">{packet.scope.max_concurrency}</Field>
    <Field label="停止条件"><Values items={packet.stops.conditions} /></Field>
    <Field label="执行截止时间">{packet.stops.deadline}</Field>
    <Field label="工作项与尝试上限">{packet.stops.max_work_items} 工作项 · 每项 {packet.stops.max_attempts} 次尝试</Field>
    <Field label="无进展限制">{packet.stops.no_progress_seconds} 秒 · {packet.stops.max_no_progress_rounds} 轮</Field>
    {notes.map(([label, values]) => <Field key={label} label={label}><Values items={values} /></Field>)}
  </div>;
}

function Values({ items }: { items: string[] | null | undefined }) {
  return items?.length ? <ul className={styles.values}>{items.map((item, index) => <li key={index}>{item}</li>)}</ul> : <MutedValue>无</MutedValue>;
}
