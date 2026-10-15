import { SelectField, TextField } from '../FormControls';
import styles from './DelegationWorkspace.module.css';

const choices = [
  { value: '', label: '请选择反馈类型' },
  { value: 'adopt', label: '采用建议' },
  { value: 'later', label: '以后再看' },
  { value: 'reject', label: '拒绝建议' },
  { value: 'revise', label: '需要修改' },
  { value: 'already_solved', label: '已经解决' },
] as const;

export interface JudgmentFeedbackDraft {
  kind: (typeof choices)[number]['value'];
  reason: string;
}

export function JudgmentFeedbackFields({ draft, disabled, onChange }: {
  draft: JudgmentFeedbackDraft;
  disabled: boolean;
  onChange: (draft: JudgmentFeedbackDraft) => void;
}) {
  return <fieldset className={styles.fields} disabled={disabled}>
    <legend>人工判断反馈</legend>
    <SelectField label="判断反馈类型" value={draft.kind} options={[...choices]} disabled={disabled} onChange={value => {
      const choice = choices.find(item => item.value === value);
      if (!choice) throw new Error('反馈类型不在当前选项中');
      onChange({ ...draft, kind: choice.value });
    }} />
    <TextField label="判断反馈理由" value={draft.reason} onChange={reason => onChange({ ...draft, reason })} multiline required disabled={disabled} />
    <p>反馈记录关联所选判断结果的固定版本。采用建议、资料准入和任务批准分别确认。</p>
    <p>“以后再看”记录单独的反馈类型，保留原判断结果。</p>
  </fieldset>;
}
