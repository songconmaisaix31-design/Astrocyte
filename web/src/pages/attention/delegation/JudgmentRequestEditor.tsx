import { useState } from 'react';
import type { components } from '../../../api/client';
import { useReadApi } from '../../../hooks/useReadApi';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { QueryState } from '../../../components/QueryState';
import { CommandState } from '../../../components/CommandState';
import { Field } from '../../../components/DetailPanel';
import { SelectField } from '../FormControls';
import { JudgmentQuestionFields } from './JudgmentQuestionFields';
import { JudgmentEvidencePicker } from './JudgmentEvidencePicker';
import { judgmentApi, type EvidenceReference, type JudgmentCommand } from './judgmentApi';
import { draftToQuestion, questionToDraft } from './ruleDraft';
import { evidenceKey, useProjectEvidence } from './useProjectEvidence';
import styles from './DelegationWorkspace.module.css';

export function JudgmentRequestEditor({ project, disabled, onResult }: { project: components['schemas']['LocalProjectV1']; disabled: boolean; onResult: (decisionID: string) => void }) {
  const [question, setQuestion] = useState(() => questionToDraft());
  const [materials, setMaterials] = useState<EvidenceReference[]>([]);
  const [language, setLanguage] = useState('');
  const [acceptedPolicy, setAcceptedPolicy] = useState('');
  const [validation, setValidation] = useState('');
  const evidence = useProjectEvidence(project.id, project.space_id);
  const policy = useReadApi(signal => { if (!question.type) throw new Error('请先选择题型'); return judgmentApi.policy(project.id, question.type, signal); }, { key: question.type, enabled: !!question.type });
  const command = usePersistentCommand<Awaited<ReturnType<typeof judgmentApi.judge>>>(`judgment:manual:${project.id}`, () => ({ status: 'succeeded', message: '判断结果已返回' }));
  const blocked = disabled || command.pending || command.uncertain;
  const unavailable = blocked || evidence.loading || evidence.stale || !evidence.data || policy.loading || policy.stale || !policy.data;
  const current = policy.data?.policy;
  const policyKey = current ? `${current.question_type}:${current.revision}:${current.snapshot_id}` : '';
  return <form className={styles.result} onSubmit={event => {
    event.preventDefault();
    if (unavailable || !current) return;
    try {
      if (!acceptedPolicy || acceptedPolicy !== policyKey) throw new Error('请核查并明确选择当前固定策略');
      if (!language || !current.languages?.includes(language)) throw new Error('请明确选择当前策略允许的语言');
      if (!materials.length || materials.some(reference => !evidence.data?.some(item => evidenceKey(item.reference) === evidenceKey(reference)))) throw new Error('请选择当前项目可读取的固定资料');
      if (current.budget.max_output_tokens === undefined) throw new Error('当前策略的输出预算未知，请先明确保存策略');
      const request: JudgmentCommand = { schema_version: '1.0', project_id: project.id, worktree_id: current.worktree_id, snapshot_id: current.snapshot_id, question: draftToQuestion(question, materials), evidence_refs: materials, language, backend: current.backend, versions: current.versions, budget: current.budget, external_scope: current.external_scope, sources: current.sources };
      setValidation('');
      void command.run(() => judgmentApi.judge(project.id, request), (response, outcome) => { if (outcome.status === 'succeeded') onResult(response.judgment.decision_id); });
    } catch (error) { setValidation(error instanceof Error ? error.message : '请核查当前判断输入'); }
  }}>
    <p>本次判断使用明确选择的固定策略、资料、语言和调用预算。提交后可查看实际结果并单独记录人工反馈。</p>
    <JudgmentQuestionFields draft={question} evidence={materials.map(reference => ({ key: evidenceKey(reference), label: `${reference.material_id} · r${reference.revision}` }))} disabled={blocked} onChange={next => { if (next.type !== question.type) setAcceptedPolicy(''); setQuestion(next); }} />
    <QueryState state={policy}>{({ policy: actual }) => <fieldset className={styles.fields} disabled={unavailable}><legend>本次实际判断策略</legend><Field label="策略版本">r{actual.revision}</Field><Field label="实际后端与模型">{actual.backend} · {actual.versions.model_revision}</Field><Field label="实际快照"><code>{actual.snapshot_id}</code></Field><Field label="调用上限">{actual.budget.timeout_ms} 毫秒 · {actual.budget.max_input_bytes} 字节 · 输入 {actual.budget.max_input_tokens} Token · 输出 {actual.budget.max_output_tokens === undefined ? '未知' : actual.budget.max_output_tokens} Token</Field>
      <button className="ac-button secondary compact" type="button" onClick={() => setAcceptedPolicy(policyKey)}>确认本次使用该固定策略</button><p>{acceptedPolicy && acceptedPolicy === policyKey ? '已明确选择当前策略。' : '请核查策略来源和调用上限后明确选择。'}</p>
      <SelectField label="本次判断语言" value={language} disabled={unavailable} onChange={setLanguage} options={[{ value: '', label: '请选择当前策略允许的语言' }, ...(actual.languages ?? []).map(value => ({ value, label: value }))]} />
    </fieldset>}</QueryState>
    <QueryState state={evidence}>{items => <JudgmentEvidencePicker title="本次判断的固定资料" available={items} selected={materials} disabled={blocked} onChange={setMaterials} />}</QueryState>
    {validation && <p role="alert">{validation} 输入已保留。</p>}<CommandState {...command} />{command.uncertain && <p role="alert">判断请求 {command.operationID} 的结果未知，关闭或刷新后仍禁止重发，请通过原结果标识只读核查。</p>}
    <button className="ac-button" type="submit" disabled={unavailable || !acceptedPolicy || acceptedPolicy !== policyKey || !language || !materials.length}>执行这一次明确的判断</button>
    <button className="ac-button secondary compact" type="button" disabled={command.pending || !question.type} onClick={policy.retry}>只读刷新当前策略</button>
  </form>;
}
