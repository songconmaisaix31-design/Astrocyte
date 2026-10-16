import { useState } from 'react';
import type { components } from '../../../api/client';
import { localProjectsApi } from '../../../api/s1';
import { useReadApi } from '../../../hooks/useReadApi';
import { SectionCard } from '../../../components/SectionCard';
import { QueryState } from '../../../components/QueryState';
import { DetailPanel, Field } from '../../../components/DetailPanel';
import { SelectField, TextField } from '../FormControls';
import { JudgmentPolicyEditor } from './JudgmentPolicyEditor';
import { DelegationRuleEditor } from './DelegationRuleEditor';
import { DelegationRuleWorkspace } from './DelegationRuleWorkspace';
import { delegationStates } from './judgmentLabels';
import { JudgmentRequestEditor } from './JudgmentRequestEditor';
import { JudgmentResultWorkspace } from './JudgmentResultWorkspace';
import { TasteProfileEditor } from './TasteProfileEditor';
import { judgmentApi, type DelegationRule, type JudgmentPolicy, type QuestionType } from './judgmentApi';
import styles from './DelegationWorkspace.module.css';

export function JudgmentDelegationWorkspace() {
  const projects = useReadApi(signal => localProjectsApi.listProjects({ signal }));
  const [projectID, setProjectID] = useState('');
  const project = projects.data?.items.find(item => item.id === projectID);
  const disabled = projects.loading || projects.stale || !projects.data;
  return <SectionCard title="判断与委托" tabs={['overview', 'opportunities']}>
    <div className={styles.result}><p>按项目配置固定判断来源与策略，查看判断、记录人工反馈，分别确认资料准入和任务批准。</p>
      <QueryState state={projects} empty={data => !data.items.length} emptyTitle="请先登记实际项目">{data => <SelectField label="判断与委托的项目" value={projectID} disabled={disabled} onChange={setProjectID} options={[{ value: '', label: '请选择已登记的实际项目' }, ...(projectID && !project ? [{ value: projectID, label: `原项目当前不可用：${projectID}` }] : []), ...data.items.map(item => ({ value: item.id, label: item.name }))]} />}</QueryState>
      {project && <JudgmentProjectWorkspace key={project.id} project={project} disabled={disabled} />}
      <TasteProfileEditor />
    </div>
  </SectionCard>;
}

type Selected = { kind: 'policy'; questionType: QuestionType; initial?: JudgmentPolicy } | { kind: 'rule'; id: string } | { kind: 'rule-editor'; initial?: DelegationRule } | { kind: 'request' } | { kind: 'result'; id: string };

function JudgmentProjectWorkspace({ project, disabled }: { project: components['schemas']['LocalProjectV1']; disabled: boolean }) {
  const [questionType, setQuestionType] = useState<QuestionType | ''>('');
  const [selected, setSelected] = useState<Selected | null>(null);
  const [decisionID, setDecisionID] = useState('');
  const rules = useReadApi(signal => judgmentApi.rules(project.id, signal));
  const policy = useReadApi(signal => { if (!questionType) throw new Error('请先选择明确题型'); return judgmentApi.policy(project.id, questionType, signal); }, { enabled: !!questionType, key: questionType });
  const blocked = disabled || rules.loading || rules.stale || !rules.data;
  const result = (id: string) => setSelected({ kind: 'result', id });
  return <div className={styles.result}>
    <Field label="当前项目权限版本">r{project.settings.revision}</Field>
    <fieldset className={styles.fields} disabled={disabled}><legend>当前题型的判断策略</legend>
      <SelectField label="读取或配置策略的题型" value={questionType} disabled={disabled} options={[{ value: '', label: '请选择明确题型' }, { value: 'choice', label: '选项判断' }, { value: 'score', label: '分级评分' }, { value: 'truth', label: '真实性判断' }]} onChange={value => { if (value !== '' && value !== 'choice' && value !== 'score' && value !== 'truth') throw new Error('策略题型无效'); setQuestionType(value); }} />
      {questionType && <><QueryState state={policy}>{({ policy: current }) => <><Field label="当前策略版本">r{current.revision}</Field><Field label="当前后端与模型">{current.backend} · {current.versions.model_revision}</Field><Field label="固定快照"><code>{current.snapshot_id}</code></Field><button className="ac-button secondary compact" type="button" disabled={disabled || policy.loading || policy.stale} onClick={() => { if (!questionType) return; setSelected({ kind: 'policy', questionType, initial: current }); }}>查看并编辑当前判断策略</button></>}</QueryState><div className={styles.actions}><button className="ac-button secondary compact" type="button" disabled={disabled || policy.loading} onClick={() => { if (!questionType) return; setSelected({ kind: 'policy', questionType }); }}>创建此题型的新策略</button><button className="ac-button secondary compact" type="button" disabled={policy.loading} onClick={policy.retry}>只读刷新策略</button></div></>}
    </fieldset>
    <div className={styles.actions}><button className="ac-button" type="button" disabled={blocked} onClick={() => setSelected({ kind: 'rule-editor' })}>创建未开启的委托规则</button><button className="ac-button secondary compact" type="button" disabled={disabled} onClick={() => setSelected({ kind: 'request' })}>人工发起一次判断</button><button className="ac-button secondary compact" type="button" onClick={rules.retry}>只读刷新规则列表</button></div>
    <QueryState state={rules} empty={data => !data.items.length} emptyTitle="该项目暂无委托规则">{data => <ul className={styles.list}>{data.items.map(rule => <li key={rule.id}><button className={styles.record} type="button" onClick={() => setSelected({ kind: 'rule', id: rule.id })}><strong>{rule.question.instructions}</strong><span>{delegationStates[rule.state as keyof typeof delegationStates] ?? rule.state} · r{rule.revision} · 策略 r{rule.policy_revision}</span><span>{rule.workflow_id} · {rule.language}</span><code>{rule.id}</code></button></li>)}</ul>}</QueryState>
    <form className={styles.fields} onSubmit={event => { event.preventDefault(); if (!disabled && decisionID.trim()) result(decisionID.trim()); }}><TextField label="只读核查的判断记录标识" value={decisionID} onChange={setDecisionID} required /><button className="ac-button secondary compact" type="submit" disabled={disabled || !decisionID.trim()}>按标识读取原判断</button></form>
    {selected && <DetailPanel title={selected.kind === 'policy' ? '判断策略' : selected.kind === 'rule-editor' ? '委托规则编辑' : selected.kind === 'rule' ? '委托规则详情' : selected.kind === 'request' ? '人工判断' : '判断结果与反馈'} onClose={() => setSelected(null)}>
      {selected.kind === 'policy' && <JudgmentPolicyEditor key={`policy:${project.id}:${selected.questionType}:${selected.initial?.revision ?? 'new'}`} project={project} questionType={selected.questionType} initial={selected.initial} disabled={disabled} onSaved={() => { policy.retry(); setSelected(null); }} />}
      {selected.kind === 'rule-editor' && <DelegationRuleEditor key={`rule-edit:${project.id}:${selected.initial?.id ?? 'new'}`} project={project} initial={selected.initial} disabled={disabled} onSaved={id => { rules.retry(); setSelected({ kind: 'rule', id }); }} />}
      {selected.kind === 'rule' && <DelegationRuleWorkspace key={`rule:${project.id}:${selected.id}`} projectID={project.id} ruleID={selected.id} disabled={disabled} onChanged={rules.retry} onEdit={initial => setSelected({ kind: 'rule-editor', initial })} onResult={result} />}
      {selected.kind === 'request' && <JudgmentRequestEditor project={project} disabled={disabled} onResult={result} />}
      {selected.kind === 'result' && <JudgmentResultWorkspace key={`result:${project.id}:${selected.id}`} projectID={project.id} decisionID={selected.id} disabled={disabled} />}
    </DetailPanel>}
  </div>;
}
