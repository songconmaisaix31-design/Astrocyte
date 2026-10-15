import type { Goal } from '../../../api/proposals';
import { Field, FieldRow } from '../../../components/DetailPanel';
import { goalStatusLabel } from './goalDraft';
import styles from './ProposalWorkspace.module.css';

export function GoalRevision({ goal }: { goal: Goal }) {
  return <div className={styles.revision}>
    <Field label="目标名称">{goal.title}</Field>
    <FieldRow><Field label="固定目标版本">r{goal.revision}</Field><Field label="目标状态">{goalStatusLabel(goal.status)}</Field></FieldRow>
    <Field label="目标成功条件"><ul className={styles.values}>{goal.success_criteria.map((value, index) => <li key={`${index}:${value}`}>{value}</li>)}</ul></Field>
    <Field label="目标优先级">{goal.priority}</Field><Field label="目标所属项目"><code>{goal.project_id}</code></Field>
    <Field label="目标 ID"><code>{goal.id}</code></Field><Field label="版本保存时间">{goal.created_at}</Field>
  </div>;
}

export function GoalRevisionDiff({ before, after }: { before: Goal; after: Goal }) {
  if (before.id !== after.id) throw new Error('目标差分需要同一目标的两个固定版本');
  const fields: { label: string; before: string; after: string }[] = [
    { label: '目标名称', before: before.title, after: after.title },
    { label: '目标成功条件', before: before.success_criteria.join('\n'), after: after.success_criteria.join('\n') },
    { label: '目标优先级', before: String(before.priority), after: String(after.priority) },
    { label: '目标状态', before: goalStatusLabel(before.status), after: goalStatusLabel(after.status) },
    { label: '目标所属项目', before: before.project_id, after: after.project_id },
  ];
  const changes = fields.filter(field => field.before !== field.after);
  return <section aria-label="目标版本差分"><h4>目标差分 r{before.revision} → r{after.revision}</h4>
    {changes.length ? changes.map(field => <div className={styles.change} key={field.label}><h5>{field.label}</h5><div className={styles.comparison}><Field label="原版本">{field.before}</Field><Field label="新版本">{field.after}</Field></div></div>) : <p>这两个版本的目标内容相同。</p>}
  </section>;
}
