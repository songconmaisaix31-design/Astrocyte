import type { components } from '../../../api/client';
import { useReadApi } from '../../../hooks/useReadApi';
import { QueryState } from '../../../components/QueryState';
import { SelectField, TextField } from '../FormControls';
import { judgmentApi, type TasteExample } from './judgmentApi';
import type { RubricDraft } from './tasteDraft';

type Material = components['schemas']['MaterialV1'];

export function TasteRubricFields({ rubric, index, disabled, onChange, onRemove }: {
  rubric: RubricDraft; index: number; disabled: boolean; onChange: (rubric: RubricDraft) => void; onRemove: () => void;
}) {
  return <fieldset className="ac-taste-rubric" disabled={disabled}>
    <legend>评判准则 {index + 1}</legend>
    <SelectField label="准则题型" value={rubric.question_type} onChange={value => {
      if (value !== '' && value !== 'choice' && value !== 'score' && value !== 'truth') throw new Error('题型无效');
      onChange({ ...rubric, question_type: value, options: [], levels: [] });
    }} options={[{ value: '', label: '请选择题型' }, { value: 'choice', label: '选择题' }, { value: 'score', label: '评分题' }, { value: 'truth', label: '确定性题' }]} />
    <TextField label="评判标准" value={rubric.criteria} onChange={criteria => onChange({ ...rubric, criteria })} multiline required />
    {rubric.question_type === 'choice' && <div role="group" aria-label="准则选项">
      {rubric.options.map((option, optionIndex) => <fieldset key={optionIndex}><legend>选项 {optionIndex + 1}</legend>
        <TextField label="选项标识" value={option.id} onChange={id => onChange({ ...rubric, options: rubric.options.map((item, i) => i === optionIndex ? { ...item, id } : item) })} required />
        <TextField label="选项说明" value={option.description} onChange={description => onChange({ ...rubric, options: rubric.options.map((item, i) => i === optionIndex ? { ...item, description } : item) })} required />
        <button className="ac-button secondary compact" type="button" onClick={() => onChange({ ...rubric, options: rubric.options.filter((_, i) => i !== optionIndex) })}>移除选项</button>
      </fieldset>)}
      <button className="ac-button secondary compact" type="button" disabled={rubric.options.length >= 255} onClick={() => onChange({ ...rubric, options: [...rubric.options, { id: '', description: '' }] })}>添加选项</button>
      <p>填写 2 至 255 个唯一选项标识及说明。</p>
    </div>}
    {rubric.question_type === 'score' && <div role="group" aria-label="准则等级">
      {rubric.levels.map((level, levelIndex) => <div key={levelIndex}>
        <TextField label={`等级 ${levelIndex}`} value={level} onChange={value => onChange({ ...rubric, levels: rubric.levels.map((item, i) => i === levelIndex ? value : item) })} required />
        <button className="ac-button secondary compact" type="button" onClick={() => onChange({ ...rubric, levels: rubric.levels.filter((_, i) => i !== levelIndex) })}>移除等级 {levelIndex}</button>
      </div>)}
      <button className="ac-button secondary compact" type="button" disabled={rubric.levels.length >= 10} onClick={() => onChange({ ...rubric, levels: [...rubric.levels, ''] })}>添加等级</button>
      <p>填写 2 至 10 个等级，预期数值从 0 开始。</p>
    </div>}
    <button className="ac-button secondary compact" type="button" onClick={onRemove}>移除准则</button>
  </fieldset>;
}

export function TasteExampleFields({ example, index, rubrics, materials, disabled, onChange, onRemove }: {
  example: TasteExample; index: number; rubrics: RubricDraft[]; materials: Material[]; disabled: boolean;
  onChange: (example: TasteExample) => void; onRemove: () => void;
}) {
  // 每个示例有独立且可取消的读取；读取响应只供展示，不能修改父级草稿。
  const detail = useReadApi(signal => judgmentApi.materialDetail(example.material_id, signal), { enabled: !!example.material_id, key: example.material_id });
  const rubric = rubrics.find(item => item.id === example.rubric_id);
  const active = detail.data?.material.lifecycle === 'active';
  const fixed = detail.data?.revisions.find(item => item.revision === example.revision);
  const validFixed = active && !!fixed && fixed.content_digest === example.content_digest;
  const available = materials.filter(material => material.lifecycle === 'active');
  return <fieldset className="ac-taste-example" disabled={disabled}>
    <legend>固定资料示例 {index + 1}</legend>
    <SelectField label="示例关联准则" value={example.rubric_id} onChange={rubric_id => onChange({ ...example, rubric_id, expected: '' })} options={[{ value: '', label: '请选择准则' }, ...rubrics.map((item, i) => ({ value: item.id, label: `准则 ${i + 1} · ${item.criteria}` }))]} />
    <SelectField label="示例资料" value={example.material_id} onChange={material_id => onChange({ ...example, material_id, revision: 0, content_digest: '' })} options={[
      { value: '', label: '请选择实际资料' }, ...(example.material_id && !available.some(item => item.id === example.material_id) ? [{ value: example.material_id, label: `原资料当前不可用：${example.material_id}` }] : []), ...available.map(item => ({ value: item.id, label: `${item.title ?? '未提供标题'} · ${item.id}` })),
    ]} />
    {example.material_id && <QueryState state={detail}>{data => <>
      {!active && <p role="alert">资料当前未处于 active 状态，请选择可用资料。</p>}
      <SelectField label="示例固定版本" value={example.revision ? String(example.revision) : ''} disabled={disabled || detail.loading || detail.stale || !active} onChange={value => {
        if (!value) { onChange({ ...example, revision: 0, content_digest: '' }); return; }
        const revision = data.revisions.find(item => String(item.revision) === value);
        if (!revision || !revision.content_digest) throw new Error('所选固定版本缺少真实摘要');
        onChange({ ...example, revision: revision.revision, content_digest: revision.content_digest });
      }} options={[{ value: '', label: '请明确选择固定版本' }, ...(example.revision && !fixed ? [{ value: String(example.revision), label: `原版本 r${example.revision} 当前不可用` }] : []), ...data.revisions.map(item => ({ value: String(item.revision), label: `r${item.revision} · ${item.created_at}` }))]} />
      {!!example.revision && !validFixed && <p role="alert">原固定版本缺失或内容摘要已变化，请核查后重新选择。</p>}
    </>}</QueryState>}
    {example.content_digest && <p>固定内容摘要：<code>{example.content_digest}</code></p>}
    {rubric?.question_type === 'choice' ? <SelectField label="示例预期选项" value={example.expected} onChange={expected => onChange({ ...example, expected })} options={[{ value: '', label: '请选择预期选项' }, ...(example.expected && !rubric.options.some(option => option.id === example.expected) ? [{ value: example.expected, label: `原选项当前不可用：${example.expected}` }] : []), ...rubric.options.filter(option => option.id.trim()).map(option => ({ value: option.id, label: `${option.id} · ${option.description}` }))]} /> : rubric?.question_type && <TextField label="示例预期数值" value={example.expected} onChange={expected => onChange({ ...example, expected })} type="number" numberRange={{ min: 0, max: rubric.question_type === 'score' ? rubric.levels.length - 1 : 1, step: 'any' }} required hint={rubric.question_type === 'score' ? `数值范围 0 至 ${rubric.levels.length - 1}` : '数值范围 0 至 1'} />}
    <button className="ac-button secondary compact" type="button" onClick={onRemove}>移除示例</button>
  </fieldset>;
}
