import type { components } from '../../../api/schema';
import { Field, FieldRow, MutedValue } from '../../../components/DetailPanel';
import { StatusBadge } from '../../../components/StatusBadge';
import { proposalStatusLabel } from '../../../utils/format';
import styles from './ProposalWorkspace.module.css';

type Proposal = components['schemas']['ProposalV1'];

function Values({ items }: { items: string[] | null }) {
  if (items === null) return <MutedValue>未填写</MutedValue>;
  return items.length ? <ul className={styles.values}>{items.map((item, index) => <li key={`${index}:${item}`}>{item}</li>)}</ul> : <MutedValue>无</MutedValue>;
}

export function ProposalRevision({ proposal }: { proposal: Proposal }) {
  return <div className={styles.revision}>
    <Field label="标题">{proposal.content.title}</Field>
    <FieldRow><Field label="提案版本">r{proposal.revision}</Field><Field label="状态"><StatusBadge value={proposal.status} label={proposalStatusLabel(proposal.status)} /></Field></FieldRow>
    <Field label="固定目标">{proposal.content.goal.id ? <><code>{proposal.content.goal.id}</code> · r{proposal.content.goal.revision}</> : <MutedValue>未分配目标</MutedValue>}</Field>
    <Field label="允许项目">{proposal.content.scope.project_id || <MutedValue>未分配项目</MutedValue>}</Field>
    <Field label="项目权限版本">r{proposal.content.scope.settings_revision}</Field>
    <Field label="允许根目录"><Values items={proposal.content.scope.roots} /></Field>
    <Field label="允许修改的路径"><Values items={proposal.content.scope.write_paths} /></Field>
    <Field label="允许动作"><Values items={proposal.content.scope.actions} /></Field>
    <Field label="允许工具"><Values items={proposal.content.scope.tools} /></Field>
    <Field label="允许外发目标"><Values items={proposal.content.scope.external_targets} /></Field>
    <Field label="最大并发数">{proposal.content.scope.max_concurrency}</Field>
    <FieldRow><Field label="调用上限">{proposal.content.budget.calls ?? <MutedValue>未知</MutedValue>}</Field><Field label="Token 上限">{proposal.content.budget.tokens ?? <MutedValue>未知</MutedValue>}</Field></FieldRow>
    <Field label="费用上限">{proposal.content.budget.money_minor === null ? <MutedValue>未知</MutedValue> : `${proposal.content.budget.money_minor} ${proposal.content.budget.currency}（最小货币单位）`}</Field>
    <Field label="预期产出"><Values items={proposal.content.deliverables} /></Field>
    <Field label="成功条件"><Values items={proposal.content.success_criteria} /></Field>
    <Field label="停止条件"><Values items={proposal.content.stops.conditions} /></Field>
    <Field label="截止时间">{proposal.content.stops.deadline}</Field>
    <FieldRow><Field label="最大工作项数">{proposal.content.stops.max_work_items}</Field><Field label="最大尝试次数">{proposal.content.stops.max_attempts}</Field></FieldRow>
    <FieldRow><Field label="连续无进展轮数上限">{proposal.content.stops.max_no_progress_rounds}</Field><Field label="无进展时限（秒）">{proposal.content.stops.no_progress_seconds}</Field></FieldRow>
    <Field label="固定输入版本">{proposal.content.inputs?.length ? <ul className={styles.values}>{proposal.content.inputs.map((reference, index) => <li key={`${index}:${reference.material_id}:${reference.revision}`}><code>{reference.material_id}</code> · r{reference.revision}</li>)}</ul> : <MutedValue>无固定输入引用</MutedValue>}</Field>
    <Field label="提案 ID"><code>{proposal.id}</code></Field>
  </div>;
}

export function ProposalRevisionDiff({ before, after }: { before: Proposal; after: Proposal }) {
  if (before.id !== after.id) throw new Error('提案版本比较必须属于同一提案');
  const budget = ({ content }: Proposal) => `调用 ${content.budget.calls ?? '未知'}；Token ${content.budget.tokens ?? '未知'}；费用 ${content.budget.money_minor ?? '未知'}（最小货币单位）；币种 ${content.budget.currency || '未知'}`;
  const sources = ({ content }: Proposal) => (content.inputs ?? []).map(reference => `${reference.material_id} · r${reference.revision}`);
  const fields: { label: string; before: string | string[] | null; after: string | string[] | null }[] = [
    { label: '标题', before: before.content.title, after: after.content.title },
    { label: '固定目标', before: `${before.content.goal.id} · r${before.content.goal.revision}`, after: `${after.content.goal.id} · r${after.content.goal.revision}` },
    { label: '允许项目', before: before.content.scope.project_id, after: after.content.scope.project_id },
    { label: '项目权限版本', before: String(before.content.scope.settings_revision), after: String(after.content.scope.settings_revision) },
    { label: '允许根目录', before: before.content.scope.roots, after: after.content.scope.roots },
    { label: '允许修改的路径', before: before.content.scope.write_paths, after: after.content.scope.write_paths },
    { label: '允许动作', before: before.content.scope.actions, after: after.content.scope.actions },
    { label: '允许工具', before: before.content.scope.tools, after: after.content.scope.tools },
    { label: '允许外发目标', before: before.content.scope.external_targets, after: after.content.scope.external_targets },
    { label: '最大并发数', before: String(before.content.scope.max_concurrency), after: String(after.content.scope.max_concurrency) },
    { label: '预算', before: budget(before), after: budget(after) },
    { label: '预期产出', before: before.content.deliverables, after: after.content.deliverables },
    { label: '成功条件', before: before.content.success_criteria, after: after.content.success_criteria },
    { label: '停止条件', before: before.content.stops.conditions, after: after.content.stops.conditions },
    { label: '截止时间', before: before.content.stops.deadline, after: after.content.stops.deadline },
    { label: '最大工作项数', before: String(before.content.stops.max_work_items), after: String(after.content.stops.max_work_items) },
    { label: '最大尝试次数', before: String(before.content.stops.max_attempts), after: String(after.content.stops.max_attempts) },
    { label: '连续无进展轮数上限', before: String(before.content.stops.max_no_progress_rounds), after: String(after.content.stops.max_no_progress_rounds) },
    { label: '无进展时限（秒）', before: String(before.content.stops.no_progress_seconds), after: String(after.content.stops.no_progress_seconds) },
    { label: '固定输入版本', before: sources(before), after: sources(after) },
  ];
  const changed = fields.filter(field => JSON.stringify(field.before) !== JSON.stringify(field.after));
  const value = (input: string | string[] | null) => input === null || Array.isArray(input) ? <Values items={input} /> : input;
  return <section aria-label="提案版本差分" className={styles.revision}>
    <h4>r{before.revision} → r{after.revision}</h4>
    {changed.length ? changed.map(field => <div key={field.label} className={styles.change}><h5>{field.label}</h5><div className={styles.comparison}><div><strong>原版本</strong>{value(field.before)}</div><div><strong>当前版本</strong>{value(field.after)}</div></div></div>) : <p>执行内容没有变化。</p>}
  </section>;
}
