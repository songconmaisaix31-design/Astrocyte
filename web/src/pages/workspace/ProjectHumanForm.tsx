import { useState } from 'react';
import { SelectField, TextField } from '../attention/FormControls';
import { intentLabels } from './projectBoardView';
import formStyles from '../attention/AttentionPage.module.css';

export interface ProjectHumanFields { notes: string; review: string; group: string; intent: string; archived: boolean; revision: number }
export function ProjectHumanForm({ human, disabled, onSave }: { human: ProjectHumanFields; disabled: boolean; onSave: (fields: ProjectHumanFields) => void }) {
  const [draft, setDraft] = useState(human);
  const changed = draft.revision !== human.revision;
  return <form className={formStyles.form} onSubmit={event => { event.preventDefault(); onSave(draft); }}>
    <p className={formStyles.note}>记录你希望继续的方向；实际活动与完成度单独显示。归档可恢复，不删除目录、资料或会话。</p>
    {changed && <div role="status">记录已有更新，当前输入保留。<button className="ac-button secondary compact" type="button" onClick={() => setDraft(human)}>载入最新人工记录</button></div>}
    <fieldset disabled={disabled || changed}><legend>下一步与人工记录</legend>
      <TextField label="备注 / 下一步" value={draft.notes} onChange={notes => setDraft({ ...draft, notes })} multiline hint="写下下一次打开项目时要做的事。" />
      <TextField label="复盘" value={draft.review} onChange={review => setDraft({ ...draft, review })} multiline />
      <TextField label="项目分组名称" value={draft.group} onChange={group => setDraft({ ...draft, group })} />
      <SelectField label="我对项目的继续意愿" value={draft.intent} onChange={intent => setDraft({ ...draft, intent })} options={Object.entries(intentLabels).map(([value, label]) => ({ value, label }))} />
      <label><input type="checkbox" style={{ width: 'auto' }} checked={draft.archived} onChange={event => setDraft({ ...draft, archived: event.target.checked })} /> 归档此项目（可取消并保存恢复）</label>
    </fieldset>
    <button className="ac-button" type="submit" disabled={disabled || changed}>保存项目记录</button>
  </form>;
}
