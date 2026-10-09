import { useState } from 'react';
import { attentionApi } from '../../api/client';
import { useOpportunityDetail, useDistillations } from '../../hooks/useAttention';
import { useCommand } from '../../hooks/useCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { Field, FieldRow, MutedValue } from '../../components/DetailPanel';
import { StatusBadge } from '../../components/StatusBadge';
import { formatDateTime, formatDimScore, opportunityStateLabel } from '../../utils/format';
import { SourceRefs } from './SourceFields';
import { TextField, SelectField } from './FormControls';
import { OpportunityForm, DimensionEditor } from './OpportunityForm';
import { dimensionLabels, feedbacks, type Dimensions, type Material, type Opportunity } from './model';
import type { components } from '../../api/schema';
import styles from './AttentionPage.module.css';

type Feedback = components['schemas']['ReviewOpportunityRequestV1']['feedback'];
type Detail = components['schemas']['OpportunityDetailV1'];
export function OpportunityWorkspace({ item, fixture, materials, onChanged }: { item: Opportunity; fixture: boolean; materials: Material[]; onChanged: () => void }) {
  const state = useOpportunityDetail(item.id, !fixture);
  const distillations = useDistillations(!fixture);
  if (fixture) return <><OpportunitySummary item={item} /><p role="note">示例数据 · 所有写操作尚未启用，不提供真实反馈或版本历史。</p></>;
  return <><button className="ac-button secondary compact" type="button" onClick={state.retry} disabled={state.loading}>刷新候选详情</button><QueryState state={state}>{data => <OpportunityBody detail={data} materials={materials} distillations={distillations} disabled={state.stale || state.loading} onChanged={() => { state.retry(); onChanged(); }} />}</QueryState></>;
}

function OpportunityBody({ detail, materials, distillations, disabled, onChanged }: { detail: Detail; materials: Material[]; distillations: ReturnType<typeof useDistillations>; disabled: boolean; onChanged: () => void }) {
  const [editing, setEditing] = useState(false);
  const [version, setVersion] = useState(detail.opportunity.revision);
  const item = detail.opportunity;
  const snapshot = detail.revisions.find(entry => entry.revision === version);
  return <>
    <OpportunitySummary item={item} />
    <p className={styles.note}>“以后再做”保留资料与候选，不当作拒绝。采用反馈只记录人工判断；准入、批准和任务执行属于后续切片。</p>
    {!item.version && <p role="alert">服务未提供候选元数据版本，写操作禁用。</p>}
    <button className="ac-button secondary compact" type="button" disabled={disabled || !item.version} onClick={() => setEditing(!editing)}>编辑候选</button>
    {editing && <QueryState state={distillations}>{data => <OpportunityForm initial={item} materials={materials} distillations={data.items} disabled={disabled || distillations.stale || distillations.loading} onSaved={onChanged} />}</QueryState>}
    <ReviewForm item={item} disabled={disabled || !item.version} onSaved={onChanged} />
    <section className={styles.record} aria-label="候选版本历史"><h4>候选版本历史</h4><div className={styles.form}><SelectField label="候选历史版本" value={String(version)} onChange={value => setVersion(Number(value))} options={detail.revisions.map(entry => ({ value: String(entry.revision), label: `r${entry.revision} · ${formatDateTime(entry.created_at)}` }))} /></div>{snapshot ? <OpportunitySummary item={snapshot} /> : <p>此历史版本暂未返回。</p>}</section>
    <Field label="人工反馈历史">{detail.reviews.length ? <ul className={styles.timeline}>{detail.reviews.map(review => <li key={review.id}>{feedbacks.find(feedback => feedback.value === review.feedback)?.label ?? review.feedback} · r{review.revision} · {formatDateTime(review.created_at)}<p>{review.reason || '理由未提供'}</p>{review.dimensions && <DimensionScores value={review.dimensions} />}</li>)}</ul> : <MutedValue>暂无人工反馈</MutedValue>}</Field>
  </>;
}

function ReviewForm({ item, disabled, onSaved }: { item: Opportunity; disabled: boolean; onSaved: () => void }) {
  const [feedback, setFeedback] = useState<Feedback>('later');
  const [reason, setReason] = useState('');
  const [dimensions, setDimensions] = useState<Dimensions>(item.dimensions);
  const command = useCommand(disabled);
  return <form className={styles.form} onSubmit={event => { event.preventDefault(); const request = command.prepare({ expected_version: item.version!, feedback, reason: reason.trim(), ...(feedback === 'reject' || feedback === 'revise' ? { dimensions } : {}) }); void command.run(() => attentionApi.reviewOpportunity(item.id, request.body, request.key), onSaved, '已保存人工反馈；未启动任务'); }}>
    <fieldset disabled={disabled || command.pending}><legend>人工反馈 · 当前候选 r{item.revision}</legend>
      <SelectField label="反馈类型" value={feedback} onChange={value => setFeedback(value as Feedback)} options={feedbacks} />
      <TextField label="反馈理由" value={reason} onChange={setReason} multiline required />
      {(feedback === 'reject' || feedback === 'revise') && <DimensionEditor value={dimensions} onChange={setDimensions} />}
    </fieldset>
    <CommandState {...command} />
    <button className="ac-button" type="submit" disabled={disabled || command.pending}>保存人工反馈</button>
  </form>;
}

function OpportunitySummary({ item }: { item: Opportunity }) {
  return <>
    <Field label="标题">{item.title || '未提供'}</Field>
    <Field label="状态"><StatusBadge value={item.state} label={opportunityStateLabel(item.state)} /></Field>
    <Field label="用途 / 为什么值得做">{item.purpose || '未提供'}</Field>
    <Field label="下一步">{item.next_step || '未提供'}</Field>
    <Field label="评估维度"><DimensionScores value={item.dimensions} /></Field>
    <Field label="证据引用"><SourceRefs refs={item.evidence_refs} /></Field>
    <Field label="关联目标">{item.goal_refs?.join('；') || '未提供'}</Field>
    <Field label="缺失证据">{item.missing_evidence?.length ? item.missing_evidence.join('；') : '缺失依据未记录'}</Field>
    <FieldRow><Field label="版本">r{item.revision}</Field><Field label="元数据版本">{item.version == null ? '未知' : `v${item.version}`}</Field></FieldRow>
  </>;
}
export function DimensionScores({ value }: { value: Dimensions }) {
  return <div className={styles.dimensions}>{(Object.keys(dimensionLabels) as (keyof Dimensions)[]).map(key => <div key={key}><span>{dimensionLabels[key]}</span><p>{value[key].value == null ? '未知' : formatDimScore(value[key].value)} · {value[key].reason || '依据未提供'}</p></div>)}</div>;
}
