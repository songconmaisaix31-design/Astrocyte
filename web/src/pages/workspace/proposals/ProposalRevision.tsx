import type { components } from '../../../api/schema';
import { Field, FieldRow, MutedValue } from '../../../components/DetailPanel';
import { StatusBadge } from '../../../components/StatusBadge';
import { proposalStatusLabel } from '../../../utils/format';
import styles from './ProposalWorkspace.module.css';

type Proposal = components['schemas']['ProposalV1'];

function Values({ items }: { items: string[] }) {
  return items.length ? <ul className={styles.values}>{items.map((item, index) => <li key={`${index}:${item}`}>{item}</li>)}</ul> : <MutedValue>无</MutedValue>;
}

export function ProposalRevision({ proposal }: { proposal: Proposal }) {
  return <div className={styles.revision}>
    <Field label="标题">{proposal.title}</Field>
    <FieldRow><Field label="提案版本">r{proposal.revision}</Field><Field label="状态"><StatusBadge value={proposal.status} label={proposalStatusLabel(proposal.status)} /></Field></FieldRow>
    <Field label="目标">{proposal.goal}</Field>
    <Field label="允许项目"><Values items={proposal.scope.project_ids} /></Field>
    <Field label="允许根目录"><Values items={proposal.scope.allowed_roots} /></Field>
    <Field label="允许动作"><Values items={proposal.scope.allowed_actions} /></Field>
    <Field label="数据外发规则"><Values items={proposal.scope.data_egress_rules} /></Field>
    <FieldRow><Field label="调用上限">{proposal.budget.calls ?? <MutedValue>未知</MutedValue>}</Field><Field label="Token 上限">{proposal.budget.tokens ?? <MutedValue>未知</MutedValue>}</Field></FieldRow>
    <Field label="费用上限">{proposal.budget.money === null ? <MutedValue>未知</MutedValue> : `${proposal.budget.money} ${proposal.budget.currency ?? '币种未知'}`}</Field>
    <Field label="预期产出"><Values items={proposal.deliverables} /></Field>
    <Field label="停止条件"><Values items={proposal.stop_conditions} /></Field>
    <Field label="候选来源与固定版本">{proposal.candidate_refs?.length ? <ul className={styles.values}>{proposal.candidate_refs.map((reference, index) => <li key={`${index}:${reference.material_id}:${reference.revision}`}><code>{reference.material_id}</code> · r{reference.revision}<div>{reference.locator}</div>{reference.span && <div>{reference.span}</div>}</li>)}</ul> : <MutedValue>无候选来源引用</MutedValue>}</Field>
    <Field label="提案 ID"><code>{proposal.id}</code></Field>
  </div>;
}

export function ProposalRevisionDiff({ before, after }: { before: Proposal; after: Proposal }) {
  if (before.id !== after.id) throw new Error('提案版本比较必须属于同一提案');
  const budget = (proposal: Proposal) => `调用 ${proposal.budget.calls ?? '未知'}；Token ${proposal.budget.tokens ?? '未知'}；费用 ${proposal.budget.money ?? '未知'}；币种 ${proposal.budget.currency ?? '未知'}`;
  const sources = (proposal: Proposal) => (proposal.candidate_refs ?? []).map(reference => `${reference.material_id} · r${reference.revision} · ${reference.locator}${reference.span ? ` · ${reference.span}` : ''}`);
  const fields: { label: string; before: string | string[]; after: string | string[] }[] = [
    { label: '标题', before: before.title, after: after.title },
    { label: '目标', before: before.goal, after: after.goal },
    { label: '允许项目', before: before.scope.project_ids, after: after.scope.project_ids },
    { label: '允许根目录', before: before.scope.allowed_roots, after: after.scope.allowed_roots },
    { label: '允许动作', before: before.scope.allowed_actions, after: after.scope.allowed_actions },
    { label: '数据外发规则', before: before.scope.data_egress_rules, after: after.scope.data_egress_rules },
    { label: '预算', before: budget(before), after: budget(after) },
    { label: '预期产出', before: before.deliverables, after: after.deliverables },
    { label: '停止条件', before: before.stop_conditions, after: after.stop_conditions },
    { label: '候选来源与固定版本', before: sources(before), after: sources(after) },
  ];
  const changed = fields.filter(field => JSON.stringify(field.before) !== JSON.stringify(field.after));
  const value = (input: string | string[]) => Array.isArray(input) ? <Values items={input} /> : input;
  return <section aria-label="提案版本差分" className={styles.revision}>
    <h4>r{before.revision} → r{after.revision}</h4>
    {changed.length ? changed.map(field => <div key={field.label} className={styles.change}><h5>{field.label}</h5><div className={styles.comparison}><div><strong>原版本</strong>{value(field.before)}</div><div><strong>当前版本</strong>{value(field.after)}</div></div></div>) : <p>执行内容没有变化。</p>}
  </section>;
}
