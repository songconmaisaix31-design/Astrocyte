import { SelectField, TextField } from '../FormControls';
import styles from './DelegationWorkspace.module.css';

export interface DelegationReviewDraft {
  admissionChoice: string;
  admissionValueMin: string;
  admissionValueMax: string;
  externalPermission: '' | 'allow' | 'deny';
  fixedInputs: boolean;
  humanAdopt: boolean;
}

export function DelegationReviewFields({ draft, questionType, options, scoreLevels, disabled, onChange }: {
  draft: DelegationReviewDraft;
  questionType: '' | 'choice' | 'score' | 'truth';
  options: readonly { id: string; label: string }[];
  scoreLevels: readonly string[];
  disabled: boolean;
  onChange: (draft: DelegationReviewDraft) => void;
}) {
  const missingChoice = !!draft.admissionChoice && !options.some(option => option.id === draft.admissionChoice);
  const incompleteScore = questionType === 'score' && (scoreLevels.length < 2 || scoreLevels.length > 10);
  return <fieldset className={styles.fields} disabled={disabled}>
    <legend>准入条件与人工复核</legend>
    {questionType === '' && <p>请先明确判断题型和题目。</p>}
    {questionType === 'choice' && <>
      <SelectField label="允许资料准入的选项" value={draft.admissionChoice} disabled={disabled} options={[
        { value: '', label: '请选择明确的准入选项' },
        ...(missingChoice ? [{ value: draft.admissionChoice, label: `原选择已不在当前候选中：${draft.admissionChoice}` }] : []),
        ...options.map(option => ({ value: option.id, label: option.label })),
      ]} onChange={admissionChoice => onChange({ ...draft, admissionChoice })} />
      {missingChoice && <p role="alert">题目候选已变化，请核对并重新选择准入选项。</p>}
    </>}
    {incompleteScore && <p role="alert">请先填写 2 至 10 个有序评分等级，再设置准入数值范围。</p>}
    {(questionType === 'truth' || questionType === 'score' && !incompleteScore) && <>
      {questionType === 'score' && <div><p>当前评分等级与数值：</p><ol start={0}>{scoreLevels.map((level, index) => <li key={index}>{level}</li>)}</ol></div>}
      <TextField label="资料准入数值下限" value={draft.admissionValueMin} onChange={admissionValueMin => onChange({ ...draft, admissionValueMin })} required disabled={disabled} type="number" numberRange={{ min: 0, max: questionType === 'truth' ? 1 : scoreLevels.length - 1, step: 'any' }} />
      <TextField label="资料准入数值上限" value={draft.admissionValueMax} onChange={admissionValueMax => onChange({ ...draft, admissionValueMax })} required disabled={disabled} type="number" numberRange={{ min: 0, max: questionType === 'truth' ? 1 : scoreLevels.length - 1, step: 'any' }} />
      <p>下限不得高于上限。该范围用于判断结果的数值，模型置信信息单独显示。</p>
    </>}
    <SelectField label="候选元数据外发许可" value={draft.externalPermission} disabled={disabled} options={[
      { value: '', label: '尚未决定' }, { value: 'deny', label: '禁止外发' }, { value: 'allow', label: '明确允许外发' },
    ]} onChange={externalPermission => {
      if (externalPermission !== '' && externalPermission !== 'allow' && externalPermission !== 'deny') throw new Error('外发许可选项无效');
      onChange({ ...draft, externalPermission });
    }} />
    <p>云端判断需要明确许可。允许候选元数据外发后，资料范围仍受当前策略和固定来源限制。</p>
    <label className={styles.confirmation}><input type="checkbox" checked={draft.fixedInputs} required onChange={event => onChange({ ...draft, fixedInputs: event.target.checked })} /><span>准入前重新核对固定资料、候选版本和当前权限（必选）</span></label>
    <label className={styles.confirmation}><input type="checkbox" checked={draft.humanAdopt} onChange={event => onChange({ ...draft, humanAdopt: event.target.checked })} /><span>准入前要求本人采用当前固定版本的判断建议</span></label>
    <p>复核条件由你明确选择。保存条件与开启委托分别确认，任务批准继续使用独立入口。</p>
  </fieldset>;
}
