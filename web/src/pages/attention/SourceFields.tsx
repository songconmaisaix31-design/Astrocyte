import { Field, MutedValue } from '../../components/DetailPanel';
import { provenanceLabel, sourceIdentity, type Material, type Provenance, type SourceRef } from './model';
import styles from './AttentionPage.module.css';

export function SourceRefs({ refs }: { refs: SourceRef[] }) {
  return refs.length ? <ul>{refs.map(ref => <li key={sourceIdentity(ref)}>{ref.locator} · 资料 {ref.material_id} · v{ref.revision} · {ref.span || '真实位置未提供'}</li>)}</ul> : <MutedValue>依据未提供</MutedValue>;
}
export function ProvenanceFields({ value }: { value: Provenance }) {
  return <Field label="处理来源">{provenanceLabel(value)} · {value.processor} · {value.version || '工具版本未提供'}<br />{value.source || '来源未提供'}<br />模型 · {value.model || '未提供'}</Field>;
}
export function ReferencePicker({ label, materials, value, onChange, disabled = false }: { label: string; materials: Material[]; value: SourceRef[]; onChange: (refs: SourceRef[]) => void; disabled?: boolean }) {
  return <fieldset disabled={disabled}><legend>{label}</legend>
    <p className={styles.note}>固定所选内容版本；原文位置只填写实际页码、章节或时间区间。</p>
    {value.map((ref, index) => <div key={sourceIdentity({ ...ref, span: null })} className={styles.record}>
      <span>{materials.find(material => material.id === ref.material_id)?.title || ref.locator} · v{ref.revision}</span>
      <label>真实位置（可选）<input value={ref.span ?? ''} onChange={event => onChange(value.map((entry, i) => i === index ? { ...entry, span: event.target.value || null } : entry))} /></label>
      <button type="button" className="ac-button secondary compact" onClick={() => onChange(value.filter((_, i) => i !== index))}>移除该引用</button>
    </div>)}
    <label>添加关联资料<select value="" onChange={event => {
      const material = materials.find(item => item.id === event.target.value);
      if (material) onChange([...value, { material_id: material.id, revision: material.current_revision, locator: material.source_locator, span: null }]);
    }}><option value="">选择资料…</option>{materials.filter(item => item.lifecycle !== 'withdrawn' && !value.some(ref => ref.material_id === item.id && ref.revision === item.current_revision)).map(item => <option key={item.id} value={item.id}>{item.title || item.source_locator} · v{item.current_revision}</option>)}</select></label>
    {!materials.length && <p>暂无可关联资料；可先保存待查问题。</p>}
  </fieldset>;
}
