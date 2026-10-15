import { SelectField, TextField } from '../FormControls';
import { questionDraftIssues } from './questionDraftIssues';
import styles from './DelegationWorkspace.module.css';

export interface JudgmentQuestionDraft {
  id: string;
  type: '' | 'choice' | 'score' | 'truth';
  instructions: string;
  rule: string;
  options: { id: string; description: string; evidenceKeys: string[] }[];
  levels: string[];
}

export function JudgmentQuestionFields({ draft, evidence, disabled, onChange }: {
  draft: JudgmentQuestionDraft;
  evidence: readonly { key: string; label: string }[];
  disabled: boolean;
  onChange: (draft: JudgmentQuestionDraft) => void;
}) {
  const rules = draft.type === 'choice' ? [{ value: 'single_supported_option', label: '依据齐全的唯一选项' }]
    : draft.type === 'score' ? [{ value: 'evidence_coverage', label: '依据覆盖程度' }]
      : draft.type === 'truth' ? [{ value: 'all_evidence_present', label: '全部必要依据齐全' }] : [];
  const missingRule = !!draft.rule && draft.rule !== 'none' && !rules.some(item => item.value === draft.rule);
  const issues = questionDraftIssues(draft);
  return <fieldset className={styles.fields} disabled={disabled}>
    <legend>委托判断的题目</legend>
    <TextField label="判断题目标识" value={draft.id} onChange={id => onChange({ ...draft, id })} required />
    <SelectField label="判断题型" value={draft.type} disabled={disabled} options={[
      { value: '', label: '请选择明确的题型' }, { value: 'choice', label: '选项判断' }, { value: 'score', label: '分级评分' }, { value: 'truth', label: '真实性判断' },
    ]} onChange={type => {
      if (type !== '' && type !== 'choice' && type !== 'score' && type !== 'truth') throw new Error('判断题型不在可选范围内');
      onChange({ ...draft, type });
    }} />
    <TextField label="完整判断题目与要求" value={draft.instructions} onChange={instructions => onChange({ ...draft, instructions })} multiline required />
    {draft.type === 'choice' && <fieldset className={styles.fields}><legend>判断候选选项</legend>
      <p>填写 2 至 255 个选项，并为规则判断选择当前资料中明确的固定依据。</p>
      {draft.options.map((option, index) => {
        const update = (next: typeof option) => onChange({ ...draft, options: draft.options.map((item, position) => position === index ? next : item) });
        const missingEvidence = option.evidenceKeys.filter(key => !evidence.some(item => item.key === key));
        return <fieldset key={index} className={styles.fields}><legend>选项 {index + 1}</legend>
          <TextField label={`选项 ${index + 1} 标识`} value={option.id} onChange={id => update({ ...option, id })} required />
          <TextField label={`选项 ${index + 1} 描述`} value={option.description} onChange={description => update({ ...option, description })} multiline />
          <fieldset className={styles.fields}><legend>选项 {index + 1} 的必要依据</legend>
            {evidence.length ? evidence.map(item => <label key={item.key} className={styles.confirmation}><input type="checkbox" checked={option.evidenceKeys.includes(item.key)} onChange={event => update({ ...option, evidenceKeys: event.target.checked ? [...option.evidenceKeys, item.key] : option.evidenceKeys.filter(key => key !== item.key) })} /><span>{item.label}</span></label>) : <p>请先选择可用于判断的实际固定资料版本。</p>}
            {missingEvidence.length > 0 && <p role="alert">原选择的固定依据当前不可用：{missingEvidence.join('；')}。请重新核查资料和版本。</p>}
          </fieldset>
          <button type="button" className="ac-button secondary compact" onClick={() => onChange({ ...draft, options: draft.options.filter((_, position) => position !== index) })}>移除判断选项 {index + 1}</button>
        </fieldset>;
      })}
      <button type="button" className="ac-button secondary compact" disabled={draft.options.length >= 255} onClick={() => onChange({ ...draft, options: [...draft.options, { id: '', description: '', evidenceKeys: [] }] })}>添加判断选项</button>
    </fieldset>}
    {draft.type === 'score' && <fieldset className={styles.fields}><legend>有序评分等级</legend><p>填写 2 至 10 个完整等级，数值从 0 开始，顺序与准入数值条件对应。</p>
      {draft.levels.map((level, index) => <div key={index}><TextField label={`评分等级 ${index}`} value={level} onChange={value => onChange({ ...draft, levels: draft.levels.map((item, position) => position === index ? value : item) })} multiline required /><button type="button" className="ac-button secondary compact" onClick={() => onChange({ ...draft, levels: draft.levels.filter((_, position) => position !== index) })}>移除评分等级 {index}</button></div>)}
      <button type="button" className="ac-button secondary compact" disabled={draft.levels.length >= 10} onClick={() => onChange({ ...draft, levels: [...draft.levels, ''] })}>添加评分等级</button>
    </fieldset>}
    <SelectField label="确定性判断规则" value={draft.rule} disabled={disabled || !draft.type} options={[
      { value: '', label: '请选择规则使用方式' }, { value: 'none', label: '不使用确定性规则' }, ...(missingRule ? [{ value: draft.rule, label: `原规则与当前题型不匹配：${draft.rule}` }] : []), ...rules,
    ]} onChange={rule => onChange({ ...draft, rule })} />
    {missingRule && <p role="alert">题型已变化，请重新核对确定性规则。原填写内容继续保留。</p>}
    {issues.length > 0 && <div role="alert"><p>请核查当前判断题目：</p><ul>{issues.map(issue => <li key={issue}>{issue}</li>)}</ul></div>}
  </fieldset>;
}
