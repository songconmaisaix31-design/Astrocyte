import { useState } from 'react';
import type { components } from '../../../api/client';
import { proposalsApi } from '../../../api/proposals';
import { useReadApi } from '../../../hooks/useReadApi';
import { useRankingProfile } from '../../../hooks/useAttention';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { QueryState } from '../../../components/QueryState';
import { CommandState } from '../../../components/CommandState';
import { Field } from '../../../components/DetailPanel';
import { SelectField, TextField } from '../FormControls';
import { JudgmentPolicyFields } from './JudgmentPolicyFields';
import { JudgmentEvidencePicker } from './JudgmentEvidencePicker';
import { judgmentApi, type EvidenceReference, type JudgmentPolicy, type JudgmentSnapshot, type QuestionType } from './judgmentApi';
import { policyDraftIssues } from './policyDraftIssues';
import { policyToDraft, requiredSafeInteger } from './policyDraft';
import { evidenceKey, useProjectEvidence } from './useProjectEvidence';
import styles from './DelegationWorkspace.module.css';

export function JudgmentPolicyEditor({ project, questionType, initial, disabled, onSaved }: {
  project: components['schemas']['LocalProjectV1'];
  questionType: QuestionType;
  initial?: JudgmentPolicy;
  disabled: boolean;
  onSaved: () => void;
}) {
  const goals = useReadApi(signal => proposalsApi.goals(signal));
  const taste = useReadApi(signal => judgmentApi.taste(signal));
  const ranking = useRankingProfile();
  const evidence = useProjectEvidence(project.id, project.space_id);
  const [draft, setDraft] = useState(() => policyToDraft(initial));
  const [expected, setExpected] = useState(initial?.revision ?? 0);
  const [goalReference, setGoalReference] = useState(initial?.sources ? `${initial.sources.goal.id}@${initial.sources.goal.revision}` : '');
  const [tasteReference, setTasteReference] = useState(initial?.sources ? `${initial.sources.taste.id}@${initial.sources.taste.version}` : '');
  const [rankingReference, setRankingReference] = useState(initial?.sources?.ranking ? `${initial.sources.ranking.id}@${initial.sources.ranking.version}` : '');
  const [goalVersion, setGoalVersion] = useState(initial?.versions.goal_version ?? '');
  const [tasteVersion, setTasteVersion] = useState(initial?.versions.taste_version ?? '');
  const [accessVersion, setAccessVersion] = useState(initial?.versions.access_epoch ?? '');
  const [files, setFiles] = useState(initial?.snapshot_files?.join('\n') ?? '');
  const [captured, setCaptured] = useState<{ snapshot: JudgmentSnapshot; files: string[]; settingsRevision: number } | null>(() => initial ? { snapshot: initial.snapshot, files: initial.snapshot_files ?? [], settingsRevision: initial.project_settings_revision } : null);
  const [externalEvidence, setExternalEvidence] = useState<EvidenceReference[]>(initial?.external_evidence ?? []);
  const [externalSourcesAllowed, setExternalSourcesAllowed] = useState(!!initial?.external_sources);
  const [validation, setValidation] = useState('');
  const [readLatest, setReadLatest] = useState(false);
  const latest = useReadApi(signal => judgmentApi.policy(project.id, questionType, signal), { enabled: readLatest });
  const capture = usePersistentCommand<Awaited<ReturnType<typeof judgmentApi.snapshot>>>(`judgment:snapshot:${project.id}:${questionType}`, () => ({ status: 'succeeded', message: '已读取真实项目快照' }));
  const save = usePersistentCommand<Awaited<ReturnType<typeof judgmentApi.savePolicy>>>(`judgment:policy:${project.id}:${questionType}`, () => ({ status: 'succeeded', message: '判断策略已保存' }));
  const blocked = disabled || capture.pending || capture.uncertain || save.pending || save.uncertain;
  const unavailable = blocked || goals.loading || goals.stale || !goals.data || taste.loading || taste.stale || !taste.data || ranking.loading || ranking.stale || !ranking.data || evidence.loading || evidence.stale || !evidence.data;
  const currentGoals = goals.data?.items.filter(goal => goal.project_id === project.id && goal.status === 'active') ?? [];
  const selectedGoal = currentGoals.find(goal => `${goal.id}@${goal.revision}` === goalReference);
  const currentTaste = taste.data?.configured && taste.data.profile?.enabled ? taste.data.profile : null;
  const currentTasteKey = currentTaste ? `${currentTaste.id}@${currentTaste.version}` : '';
  const currentRanking = ranking.data?.configured && ranking.data.profile?.enabled ? ranking.data.profile : null;
  const currentRankingKey = currentRanking ? `${currentRanking.id}@${currentRanking.version}` : '';
  const lines = files.split('\n').map(value => value.trim()).filter(Boolean);
  const matchingSnapshot = !!captured && captured.settingsRevision === project.settings.revision && lines.length === captured.files.length && lines.every((value, index) => value === captured.files[index]);
  return <section className={styles.result}>
    <Field label="项目权限版本">r{project.settings.revision}</Field><Field label="题型">{questionType}</Field><Field label="保存使用的策略版本">r{expected} → r{expected + 1}</Field>
    <form className={styles.result} onSubmit={event => {
      event.preventDefault();
      if (unavailable) return;
      try {
        const issues = policyDraftIssues(draft, questionType);
        if (issues.length) throw new Error(issues.join(' '));
        if (!draft.backend || !draft.modelRevision.trim() || !draft.policyVersion.trim() || !goalVersion.trim() || !tasteVersion.trim() || !accessVersion.trim()) throw new Error('请完整填写后端和全部版本标识');
        if (!selectedGoal || !currentTaste || !tasteReference || tasteReference !== currentTasteKey) throw new Error('请选择当前项目目标和当前启用的完整 Taste 版本');
        if (!captured || !matchingSnapshot || !lines.length) throw new Error('请明确文件范围并重新读取真实项目快照');
        if (!draft.languages.trim()) throw new Error('请明确填写允许的判断语言');
        if (externalEvidence.some(reference => !evidence.data?.some(item => evidenceKey(item.reference) === evidenceKey(reference)))) throw new Error('外发依据当前无法读取，请核查固定版本与权限');
        if (rankingReference && (!currentRanking || rankingReference !== currentRankingKey)) throw new Error('原排序来源已变化，请明确选择当前版本或移除此可选来源');
        const sources = { goal: { id: selectedGoal.id, revision: selectedGoal.revision }, taste: { id: currentTaste.id, version: currentTaste.version }, ...(rankingReference && currentRanking ? { ranking: { id: currentRanking.id, version: currentRanking.version } } : {}) };
        if (draft.backend === 'jev' && (!externalSourcesAllowed || !draft.externalScope.trim())) throw new Error('云端判断需要明确批准固定目标、完整 Taste 与外发范围');
        const policy: JudgmentPolicy = {
          revision: expected + 1, project_id: project.id, project_settings_revision: project.settings.revision,
          question_type: questionType, worktree_id: captured.snapshot.worktree_id, snapshot_id: captured.snapshot.generation, snapshot_files: captured.files, snapshot: captured.snapshot,
          backend: draft.backend, languages: draft.languages.split('\n').map(value => value.trim()).filter(Boolean),
          versions: { goal_version: goalVersion.trim(), taste_version: tasteVersion.trim(), policy_version: draft.policyVersion.trim(), model_revision: draft.modelRevision.trim(), access_epoch: accessVersion.trim() },
          budget: { timeout_ms: requiredSafeInteger(draft.timeoutMS, '判断超时上限', 1), max_input_bytes: requiredSafeInteger(draft.maxInputBytes, '判断输入字节上限', 1), max_input_tokens: requiredSafeInteger(draft.maxInputTokens, '判断输入 Token 上限', 1), max_output_tokens: requiredSafeInteger(draft.maxOutputTokens, '判断输出 Token 上限', 0) },
          min_confidence: draft.minConfidence.trim() ? Number(draft.minConfidence) : null, truth_accept_min: draft.truthAcceptMin.trim() ? Number(draft.truthAcceptMin) : null, truth_accept_max: draft.truthAcceptMax.trim() ? Number(draft.truthAcceptMax) : null,
          external_scope: draft.externalScope, external_evidence: externalEvidence, sources, ...(externalSourcesAllowed ? { external_sources: sources } : {}),
        };
        setValidation('');
        void save.run(() => judgmentApi.savePolicy(project.id, questionType, { expected_revision: expected, policy }), (_, outcome) => { if (outcome.status === 'succeeded') onSaved(); });
      } catch (error) { setValidation(error instanceof Error ? error.message : '请核查判断策略输入'); }
    }}>
      <QueryState state={goals}>{() => <SelectField label="判断固定目标" value={goalReference} disabled={unavailable} onChange={value => { setGoalReference(value); setExternalSourcesAllowed(false); }} options={[
        { value: '', label: '请选择当前项目的目标版本' }, ...(goalReference && !selectedGoal ? [{ value: goalReference, label: `原目标当前不可用：${goalReference}` }] : []), ...currentGoals.map(goal => ({ value: `${goal.id}@${goal.revision}`, label: `${goal.title} · r${goal.revision}` })),
      ]} />}</QueryState>
      <QueryState state={taste}>{() => <><SelectField label="判断完整 Taste 来源" value={tasteReference} disabled={unavailable} onChange={value => { setTasteReference(value); setExternalSourcesAllowed(false); }} options={[
        { value: '', label: '请选择当前完整 Taste 版本' }, ...(tasteReference && tasteReference !== currentTasteKey ? [{ value: tasteReference, label: `原 Taste 当前不可用：${tasteReference}` }] : []), ...(currentTaste ? [{ value: currentTasteKey, label: `${currentTaste.id} · v${currentTaste.version} · ${currentTaste.rubrics.length} 项准则` }] : []),
      ]} />{!currentTaste && <p>完整 Taste 尚未配置或尚未启用，请先在品味设置中明确保存。</p>}</>}</QueryState>
      <QueryState state={ranking}>{() => <SelectField label="可选的固定排序来源" value={rankingReference} disabled={unavailable} onChange={value => { setRankingReference(value); setExternalSourcesAllowed(false); }} options={[{ value: '', label: '不附加排序来源' }, ...(rankingReference && rankingReference !== currentRankingKey ? [{ value: rankingReference, label: `原排序来源当前不可用：${rankingReference}` }] : []), ...(currentRanking ? [{ value: currentRankingKey, label: `${currentRanking.id} · v${currentRanking.version}` }] : [])]} />}</QueryState>
      <fieldset className={styles.fields} disabled={blocked}><legend>真实项目快照</legend><TextField label="判断快照文件范围" value={files} onChange={setFiles} multiline required hint="每行填写一个项目允许目录内的具体文件；服务按当前权限实际读取。" />
        <button className="ac-button secondary compact" type="button" disabled={blocked || !lines.length} onClick={() => {
          if (new Set(lines).size !== lines.length) { setValidation('快照文件重复，请逐项核查'); return; }
          const capturedFiles = [...lines];
          void capture.run(() => judgmentApi.snapshot(project.id, capturedFiles), (response, outcome) => { if (outcome.status === 'succeeded') setCaptured({ snapshot: response.snapshot, files: capturedFiles, settingsRevision: project.settings.revision }); });
        }}>读取当前真实项目快照</button>
        {captured && <><Field label="快照版本"><code>{captured.snapshot.generation}</code></Field><Field label="实际 HEAD"><code>{captured.snapshot.head}</code></Field><Field label="实际工作树"><code>{captured.snapshot.worktree_id}</code></Field>{!matchingSnapshot && <p role="alert">文件范围或项目权限已变化，请明确重新读取快照。</p>}</>}
        <CommandState {...capture} />{capture.uncertain && <p role="alert">快照读取请求 {capture.operationID} 的结果未知，关闭或刷新后继续保持锁定，请核查原操作。</p>}
      </fieldset>
      <JudgmentPolicyFields draft={draft} questionType={questionType} disabled={unavailable} onChange={setDraft} />
      <fieldset className={styles.fields} disabled={unavailable}><legend>固定来源的版本标识</legend><TextField label="目标版本标识" value={goalVersion} onChange={setGoalVersion} required /><TextField label="Taste 版本标识" value={tasteVersion} onChange={setTasteVersion} required /><TextField label="授权版本标识" value={accessVersion} onChange={setAccessVersion} required /><p>版本标识与上方实际固定来源一并保存，变更来源后请重新核查。</p></fieldset>
      <QueryState state={evidence}>{items => <JudgmentEvidencePicker title="明确允许外发的固定资料版本" available={items} selected={externalEvidence} disabled={unavailable} onChange={setExternalEvidence} />}</QueryState>
      <label className={styles.confirmation}><input type="checkbox" checked={externalSourcesAllowed} disabled={unavailable} onChange={event => setExternalSourcesAllowed(event.target.checked)} /><span>明确允许所选固定目标、完整 Taste 和其排序来源用于云端判断</span></label>
      {validation && <p role="alert">{validation} 输入已保留。</p>}<CommandState {...save} />
      {save.uncertain && <p role="alert">保存请求 {save.operationID} 的结果未知，关闭和刷新后继续禁止重发。请只读取当前策略核查。</p>}
      <button className="ac-button" type="submit" disabled={unavailable || !matchingSnapshot}>保存当前题型的判断策略</button>
    </form>
    <button className="ac-button secondary compact" type="button" disabled={save.pending} onClick={() => { if (readLatest) latest.retry(); else setReadLatest(true); }}>读取最新策略以核查版本</button>
    {readLatest && <QueryState state={latest}>{({ policy }) => <div><Field label="当前已保存策略版本">r{policy.revision}</Field><Field label="当前固定快照"><code>{policy.snapshot_id}</code></Field><Field label="当前模型版本">{policy.versions.model_revision}</Field>{policy.revision !== expected && <button className="ac-button secondary compact" type="button" disabled={blocked || latest.loading || latest.stale} onClick={() => setExpected(policy.revision)}>使用已核对的最新版本保存原输入</button>}</div>}</QueryState>}
  </section>;
}
