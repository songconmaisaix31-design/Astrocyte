import { useState } from 'react';
import { listOpportunities } from '../../../api/client';
import { useReadApi } from '../../../hooks/useReadApi';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { QueryState } from '../../../components/QueryState';
import { CommandState } from '../../../components/CommandState';
import { Field } from '../../../components/DetailPanel';
import { SelectField } from '../FormControls';
import { judgmentApi, type DelegationRule, type JudgmentPolicy } from './judgmentApi';
import { delegationStates } from './judgmentLabels';
import styles from './DelegationWorkspace.module.css';

export function DelegationRuleWorkspace({ projectID, ruleID, disabled, onEdit, onResult, onChanged }: {
  projectID: string; ruleID: string; disabled: boolean;
  onEdit: (rule: DelegationRule) => void;
  onResult: (decisionID: string) => void;
  onChanged: () => void;
}) {
  const rule = useReadApi(signal => judgmentApi.rule(projectID, ruleID, signal));
  const questionType = rule.data?.delegation_rule.question.question_type;
  const policy = useReadApi(signal => { if (!questionType) throw new Error('请先读取实际规则题型'); return judgmentApi.policy(projectID, questionType, signal); }, { enabled: !!questionType, key: questionType ?? '' });
  const blocked = disabled || rule.loading || rule.stale || !rule.data || policy.loading || policy.stale || !policy.data;
  return <section className={styles.result}>
    <button className="ac-button secondary compact" type="button" onClick={() => { rule.retry(); if (questionType) policy.retry(); }}>只读刷新规则与策略</button>
    <QueryState state={rule}>{({ delegation_rule: current }) => <>
      <Field label="规则标识"><code>{current.id}</code> · r{current.revision}</Field><Field label="委托状态">{delegationStates[current.state as keyof typeof delegationStates] ?? current.state}</Field>
      <Field label="工作流"><code>{current.workflow_id}</code></Field><Field label="人类配置者"><code>{current.owner_id}</code></Field>
      <Field label="固定题目">{current.question.instructions}</Field><Field label="固定策略版本">r{current.policy_revision}</Field><Field label="判断语言">{current.language}</Field>
      <Field label="准入选项">{current.admission_choice ?? '该题型使用数值条件'}</Field><Field label="准入数值范围">{current.admission_value_min === null || current.admission_value_max === null ? '该题型使用选项条件' : `${current.admission_value_min} 至 ${current.admission_value_max}`}</Field>
      <Field label="候选元数据外发许可">{current.allow_opportunity_external === null ? '尚未决定' : current.allow_opportunity_external ? '明确允许' : '禁止外发'}</Field><Field label="人工复核条件">{current.review_conditions?.map(condition => condition === 'fixed_inputs' ? '核对固定资料、候选版本和权限' : condition === 'human_adopt' ? '本人采用当前固定判断' : condition).join('；') || '未提供'}</Field>
      <details><summary>固定资料与判断选项</summary><ul>{current.materials?.map(reference => <li key={`${reference.material_id}:${reference.revision}`}><code>{reference.material_id}</code> · r{reference.revision}<br /><code>{reference.digest}</code></li>)}</ul>{current.question.options?.map(option => <Field key={option.id} label={option.id}>{option.description}</Field>)}{current.question.levels && <ol start={0}>{current.question.levels.map((level, index) => <li key={index}>{level}</li>)}</ol>}</details>
      <button className="ac-button secondary compact" type="button" disabled={blocked || current.state === 'revoked'} onClick={() => onEdit(current)}>编辑并保存未开启的新版本</button>
      <QueryState state={policy}>{({ policy: actual }) => <><Field label="当前策略">r{actual.revision} · {actual.backend} · {actual.versions.model_revision}</Field>{actual.revision !== current.policy_revision && <p role="alert">当前策略已变化，请编辑规则并明确核查新策略版本。</p>}<RuleActions projectID={projectID} rule={current} policy={actual} disabled={blocked} onChanged={() => { rule.retry(); onChanged(); }} onResult={onResult} /></>}</QueryState>
    </>}</QueryState>
  </section>;
}

function RuleActions({ projectID, rule, policy, disabled, onChanged, onResult }: { projectID: string; rule: DelegationRule; policy: JudgmentPolicy; disabled: boolean; onChanged: () => void; onResult: (decisionID: string) => void }) {
  const [nextState, setNextState] = useState('');
  const [candidate, setCandidate] = useState('');
  const opportunities = useReadApi(signal => listOpportunities({ signal }));
  const state = usePersistentCommand<Awaited<ReturnType<typeof judgmentApi.setState>>>(`judgment:rule-state:${projectID}:${rule.id}`, () => ({ status: 'succeeded', message: '委托状态已保存' }));
  const evaluation = usePersistentCommand<Awaited<ReturnType<typeof judgmentApi.evaluate>>>(`judgment:rule-evaluation:${projectID}:${rule.id}`, () => ({ status: 'succeeded', message: '固定候选判断已返回' }));
  const busy = disabled || state.pending || state.uncertain || evaluation.pending || evaluation.uncertain;
  const ready = opportunities.data?.items.filter(item => item.state === 'ready_for_review') ?? [];
  const selected = ready.find(item => `${item.id}@${item.revision}` === candidate);
  const matchingPolicy = policy.revision === rule.policy_revision;
  return <>
    <fieldset className={styles.fields} disabled={busy || rule.state === 'revoked'}><legend>人工变更委托状态</legend>
      <SelectField label="新的委托状态" value={nextState} disabled={busy || rule.state === 'revoked'} onChange={setNextState} options={[{ value: '', label: '请选择本次确认的状态' }, ...Object.entries(delegationStates).map(([value, label]) => ({ value, label }))]} />
      <p>开启委托会重新核查固定资料、当前策略和实际权限。执行一次判断使用下方独立入口。</p>
      <button className="ac-button" type="button" disabled={busy || !nextState || nextState === rule.state || nextState === 'active' && !matchingPolicy || rule.state === 'revoked'} onClick={() => { void state.run(() => judgmentApi.setState(projectID, rule.id, { expected_revision: rule.revision, state: nextState }), (_, outcome) => { if (outcome.status === 'succeeded') onChanged(); }); }}>确认变更当前委托状态</button>
      <CommandState {...state} />{state.uncertain && <p role="alert">状态变更请求 {state.operationID} 的结果未知，关闭或刷新后禁止重发，请只读核查原规则。</p>}
    </fieldset>
    <fieldset className={styles.fields} disabled={busy || rule.state !== 'active' || !matchingPolicy}><legend>对固定候选执行一次判断</legend>
      <Field label="本次后端与模型">{policy.backend} · {policy.versions.model_revision}</Field><Field label="本次调用上限">{policy.budget.timeout_ms} 毫秒 · {policy.budget.max_input_bytes} 字节 · 输入 {policy.budget.max_input_tokens} Token · 输出 {policy.budget.max_output_tokens === undefined ? '未知' : policy.budget.max_output_tokens} Token</Field>
      <QueryState state={opportunities} empty={data => !data.items.length} emptyTitle="暂无实际候选">{() => <SelectField label="本次判断的固定候选" value={candidate} disabled={busy || opportunities.loading || opportunities.stale} onChange={setCandidate} options={[{ value: '', label: '请选择已可审阅的候选版本' }, ...(candidate && !selected ? [{ value: candidate, label: `原候选当前不可用：${candidate}` }] : []), ...ready.map(item => ({ value: `${item.id}@${item.revision}`, label: `${item.title || item.id} · r${item.revision}` }))]} />}</QueryState>
      <p>本次调用使用规则已固定的语言、资料、模型和预算。判断完成后可查看结果并单独记录反馈与准入。</p>
      <button className="ac-button" type="button" disabled={busy || rule.state !== 'active' || !matchingPolicy || !selected || opportunities.loading || opportunities.stale || !opportunities.data || policy.budget.max_output_tokens === undefined} onClick={() => {
        if (!selected || busy) return;
        void evaluation.run(() => judgmentApi.evaluate(projectID, rule.id, { rule_revision: rule.revision, opportunity_id: selected.id, opportunity_revision: selected.revision }), (response, outcome) => { if (outcome.status === 'succeeded') onResult(response.judgment.decision_id); });
      }}>执行该固定候选的一次判断</button>
      <CommandState {...evaluation} />{evaluation.uncertain && <p role="alert">判断请求 {evaluation.operationID} 的结果未知，原请求身份继续保留，关闭或刷新后禁止再次提交。</p>}
    </fieldset>
  </>;
}
