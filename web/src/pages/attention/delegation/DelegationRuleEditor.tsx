import { useState } from 'react';
import type { components } from '../../../api/client';
import { useReadApi } from '../../../hooks/useReadApi';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { QueryState } from '../../../components/QueryState';
import { CommandState } from '../../../components/CommandState';
import { Field } from '../../../components/DetailPanel';
import { SelectField, TextField } from '../FormControls';
import { JudgmentQuestionFields } from './JudgmentQuestionFields';
import { DelegationReviewFields } from './DelegationReviewFields';
import { JudgmentEvidencePicker } from './JudgmentEvidencePicker';
import { judgmentApi, type DelegationRule, type EvidenceReference } from './judgmentApi';
import { draftToQuestion, draftToReview, questionToDraft, reviewToDraft } from './ruleDraft';
import { evidenceKey, useProjectEvidence } from './useProjectEvidence';
import styles from './DelegationWorkspace.module.css';

export function DelegationRuleEditor({ project, initial, disabled, onSaved }: {
  project: components['schemas']['LocalProjectV1'];
  initial?: DelegationRule;
  disabled: boolean;
  onSaved: (ruleID: string) => void;
}) {
  const draftStorage = `astrocyte:judgment:rule-draft:${project.id}`;
  const [ruleID] = useState(() => initial?.id ?? window.sessionStorage.getItem(draftStorage) ?? crypto.randomUUID());
  const [question, setQuestion] = useState(() => questionToDraft(initial?.question));
  const [review, setReview] = useState(() => reviewToDraft(initial));
  const [workflow, setWorkflow] = useState(initial?.workflow_id ?? '');
  const [materials, setMaterials] = useState<EvidenceReference[]>(initial?.materials ?? []);
  const [policyRevision, setPolicyRevision] = useState<number | null>(initial?.policy_revision ?? null);
  const [language, setLanguage] = useState(initial?.language ?? '');
  const [expected, setExpected] = useState(initial?.revision ?? 0);
  const [validation, setValidation] = useState('');
  const [readLatest, setReadLatest] = useState(false);
  const latest = useReadApi(signal => judgmentApi.rule(project.id, ruleID, signal), { enabled: readLatest });
  const session = useReadApi(() => judgmentApi.session());
  const evidence = useProjectEvidence(project.id, project.space_id);
  const policy = useReadApi(signal => { if (!question.type) throw new Error('请先选择题型'); return judgmentApi.policy(project.id, question.type, signal); }, { key: question.type, enabled: !!question.type });
  const command = usePersistentCommand<Awaited<ReturnType<typeof judgmentApi.saveRule>>>(initial ? `judgment:rule:${project.id}:${ruleID}` : `judgment:rule-create:${project.id}`, () => ({ status: 'succeeded', message: '委托规则已保存为未开启版本' }));
  const blocked = disabled || command.pending || command.uncertain || initial?.state === 'revoked' || latest.data?.delegation_rule.state === 'revoked';
  const unavailable = blocked || evidence.loading || evidence.stale || !evidence.data || session.loading || session.stale || !session.data || policy.loading || policy.stale || !policy.data;
  const currentPolicy = policy.data?.policy;
  const languages = currentPolicy?.languages ?? [];
  return <section className={styles.result}>
    <Field label="规则标识"><code>{ruleID}</code></Field><Field label="规则版本">r{expected} → r{expected + 1}</Field>
    <p>保存的新规则或新版本保持未开启。开启委托需要返回规则详情，核查当前策略与固定资料后单独确认。</p>
    <form className={styles.result} onSubmit={event => {
      event.preventDefault();
      if (unavailable || !session.data || !currentPolicy) return;
      try {
        if (!workflow.trim() || !materials.length) throw new Error('请明确填写工作流标识并选择固定资料');
        if (policyRevision !== currentPolicy.revision || !language || !languages.includes(language)) throw new Error('请核查并明确选择当前策略版本及其允许的语言');
        if (materials.some(reference => !evidence.data?.some(item => evidenceKey(item.reference) === evidenceKey(reference)))) throw new Error('固定资料当前不可读取，请核查权限与版本');
        const actualQuestion = draftToQuestion(question, materials);
        const rule: DelegationRule = { id: ruleID, project_id: project.id, workflow_id: workflow.trim(), owner_id: session.data.actor_id, revision: expected + 1, state: 'disabled', purpose: 'opportunity_admission', question: actualQuestion, policy_revision: policyRevision, language, materials, ...draftToReview(review, actualQuestion) };
        setValidation('');
        if (!initial) window.sessionStorage.setItem(draftStorage, ruleID);
        void command.run(async () => { const response = await judgmentApi.saveRule(project.id, ruleID, { expected_revision: expected, rule }); if (!initial) window.sessionStorage.removeItem(draftStorage); return response; }, (response, outcome) => { if (outcome.status === 'succeeded') onSaved(response.delegation_rule.id); });
      } catch (error) { setValidation(error instanceof Error ? error.message : '请核查委托规则输入'); }
    }}>
      <fieldset className={styles.fields} disabled={blocked}><legend>委托身份</legend><TextField label="委托工作流标识" value={workflow} onChange={setWorkflow} required disabled={!!initial} /><QueryState state={session}>{data => <Field label="当前人类控制面身份"><code>{data.actor_id}</code></Field>}</QueryState></fieldset>
      <JudgmentQuestionFields draft={question} evidence={materials.map(reference => ({ key: evidenceKey(reference), label: evidence.data?.find(item => evidenceKey(item.reference) === evidenceKey(reference))?.label ?? `${reference.material_id} · r${reference.revision}` }))} disabled={blocked} onChange={next => { if (next.type !== question.type) setPolicyRevision(null); setQuestion(next); }} />
      <QueryState state={policy}>{({ policy: current }) => <fieldset className={styles.fields} disabled={unavailable}><legend>当前实际判断策略</legend><Field label="策略版本">r{current.revision}</Field><Field label="判断后端">{current.backend}</Field><Field label="模型版本">{current.versions.model_revision}</Field><Field label="每次调用限制">{current.budget.timeout_ms} 毫秒 · {current.budget.max_input_bytes} 字节 · 输入 {current.budget.max_input_tokens} Token · 输出 {current.budget.max_output_tokens === undefined ? '未知' : current.budget.max_output_tokens} Token</Field>
        <button className="ac-button secondary compact" type="button" onClick={() => setPolicyRevision(current.revision)}>使用已核对的策略 r{current.revision}</button><Field label="本规则固定策略版本">{policyRevision === null ? '尚未选择' : `r${policyRevision}`}</Field>{policyRevision !== null && policyRevision !== current.revision && <p role="alert">原策略版本已变化，请核查新策略后再明确选择。</p>}
        <SelectField label="本规则允许的判断语言" value={language} disabled={unavailable} onChange={setLanguage} options={[{ value: '', label: '请选择当前策略允许的语言' }, ...(language && !languages.includes(language) ? [{ value: language, label: `原语言当前未获许可：${language}` }] : []), ...languages.map(value => ({ value, label: value }))]} />
      </fieldset>}</QueryState>
      <QueryState state={evidence}>{items => <JudgmentEvidencePicker title="本规则的固定资料范围" available={items} selected={materials} disabled={blocked} onChange={setMaterials} />}</QueryState>
      <DelegationReviewFields draft={review} questionType={question.type} options={question.options.filter(option => option.id.trim()).map(option => ({ id: option.id, label: `${option.id} · ${option.description}` }))} scoreLevels={question.levels} disabled={blocked} onChange={setReview} />
      {validation && <p role="alert">{validation} 输入已保留。</p>}<CommandState {...command} />
      {command.uncertain && <p role="alert">规则保存请求 {command.operationID} 的结果未知，原规则标识为 {ruleID}。关闭或刷新后仍禁止重发，请只读取原规则核查。</p>}
      <button className="ac-button" type="submit" disabled={unavailable || policyRevision !== currentPolicy?.revision}>保存为未开启的委托规则</button>
    </form>
    <div className={styles.actions}><button className="ac-button secondary compact" type="button" onClick={policy.retry} disabled={!question.type || command.pending}>刷新当前判断策略</button><button className="ac-button secondary compact" type="button" onClick={() => { if (readLatest) latest.retry(); else setReadLatest(true); }} disabled={command.pending}>只读核查原规则版本</button></div>
    {readLatest && <QueryState state={latest}>{({ delegation_rule: rule }) => <div><Field label="当前规则版本">r{rule.revision} · {rule.state}</Field>{rule.revision !== expected && <button className="ac-button secondary compact" type="button" disabled={blocked || latest.loading || latest.stale} onClick={() => setExpected(rule.revision)}>使用已核对的规则版本保存原输入</button>}</div>}</QueryState>}
  </section>;
}
