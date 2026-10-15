import { SelectField, TextField } from '../FormControls';
import { policyDraftIssues } from './policyDraftIssues';
import styles from './DelegationWorkspace.module.css';

export interface JudgmentPolicyDraft {
  backend: '' | 'rule' | 'jev' | 'laya';
  languages: string;
  modelRevision: string;
  policyVersion: string;
  externalScope: string;
  timeoutMS: string;
  maxInputBytes: string;
  maxInputTokens: string;
  maxOutputTokens: string;
  minConfidence: string;
  truthAcceptMin: string;
  truthAcceptMax: string;
}

export function JudgmentPolicyFields({ draft, questionType, disabled, onChange }: {
  draft: JudgmentPolicyDraft;
  questionType: '' | 'choice' | 'score' | 'truth';
  disabled: boolean;
  onChange: (draft: JudgmentPolicyDraft) => void;
}) {
  const update = (key: Exclude<keyof JudgmentPolicyDraft, 'backend'>, value: string) => onChange({ ...draft, [key]: value });
  const issues = policyDraftIssues(draft, questionType);
  return <fieldset className={styles.fields} disabled={disabled}>
    <legend>判断策略与调用范围</legend>
    <SelectField label="判断后端" value={draft.backend} disabled={disabled} options={[
      { value: '', label: '请选择明确的判断后端' },
      { value: 'rule', label: '确定性规则' },
      { value: 'laya', label: '本地 Laya' },
      { value: 'jev', label: '云端 Jev' },
    ]} onChange={backend => {
      if (backend !== '' && backend !== 'rule' && backend !== 'laya' && backend !== 'jev') throw new Error('判断后端不在当前选项中');
      onChange({ ...draft, backend });
    }} />
    <TextField label="允许的判断语言" value={draft.languages} onChange={value => update('languages', value)} multiline required hint="每行填写一种明确允许的语言。" />
    <TextField label="判断策略版本标识" value={draft.policyVersion} onChange={value => update('policyVersion', value)} required />
    <TextField label="判断模型或规则版本" value={draft.modelRevision} onChange={value => update('modelRevision', value)} required hint="使用已配置后端的实际版本；改变版本后重新核查策略和固定来源。" />
    <TextField label="允许外发的资料范围说明" value={draft.externalScope} onChange={value => update('externalScope', value)} multiline required={draft.backend === 'jev'} />
    {draft.backend === 'jev' && <p>请逐项确认允许外发的固定资料和来源版本。填写范围说明后，服务仍核查每项资料的实际外发许可。</p>}
    <fieldset className={styles.fields}><legend>每次判断的调用预算</legend>
      <TextField label="判断超时上限（毫秒）" value={draft.timeoutMS} onChange={value => update('timeoutMS', value)} required type="number" numberRange={{ min: 1, max: Number.MAX_SAFE_INTEGER, step: '1' }} />
      <TextField label="判断输入字节上限" value={draft.maxInputBytes} onChange={value => update('maxInputBytes', value)} required type="number" numberRange={{ min: 1, max: Number.MAX_SAFE_INTEGER, step: '1' }} />
      <TextField label="判断输入 Token 上限" value={draft.maxInputTokens} onChange={value => update('maxInputTokens', value)} required type="number" numberRange={{ min: 1, max: Number.MAX_SAFE_INTEGER, step: '1' }} />
      <TextField label="判断输出 Token 上限" value={draft.maxOutputTokens} onChange={value => update('maxOutputTokens', value)} required type="number" numberRange={{ min: 0, max: Number.MAX_SAFE_INTEGER, step: '1' }} />
      <p>输入预算包含完整题目、判断条件和必要资料。服务报告截断后，先重新准备资料并核查预算。</p>
    </fieldset>
    <TextField label="模型置信下限（可选）" value={draft.minConfidence} onChange={value => update('minConfidence', value)} type="number" numberRange={{ min: 0, max: 1, step: 'any' }} />
    <p>未提供模型置信信息时保留未知。模型建议及其置信信息分别显示。</p>
    {questionType === 'truth' && <fieldset className={styles.fields}><legend>真实性判断的可接受范围（可选）</legend>
      <TextField label="真实性判断数值下限" value={draft.truthAcceptMin} onChange={value => update('truthAcceptMin', value)} type="number" numberRange={{ min: 0, max: 1, step: 'any' }} required={draft.truthAcceptMax !== ''} />
      <TextField label="真实性判断数值上限" value={draft.truthAcceptMax} onChange={value => update('truthAcceptMax', value)} type="number" numberRange={{ min: 0, max: 1, step: 'any' }} required={draft.truthAcceptMin !== ''} />
      <p>同时填写两个边界，下限不得高于上限。题目或题型变更后请重新核查范围。</p>
    </fieldset>}
    {issues.length > 0 && <div role="alert"><p>请核查当前判断策略：</p><ul>{issues.map(issue => <li key={issue}>{issue}</li>)}</ul></div>}
  </fieldset>;
}
