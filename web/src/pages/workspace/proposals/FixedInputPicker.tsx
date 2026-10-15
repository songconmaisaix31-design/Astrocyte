import type { components } from '../../../api/schema';
import styles from './ProposalWorkspace.module.css';

type FixedReference = components['schemas']['FixedReferenceV1'];
type SourceReference = components['schemas']['SourceRefV1'];

export function FixedInputPicker({ available, selected, disabled, onChange }: { available: SourceReference[]; selected: FixedReference[]; disabled: boolean; onChange: (selected: FixedReference[]) => void }) {
  const same = (a: FixedReference, b: FixedReference) => a.material_id === b.material_id && a.revision === b.revision;
  const toggle = (reference: FixedReference) => onChange(selected.some(value => same(value, reference)) ? selected.filter(value => !same(value, reference)) : [...selected, { material_id: reference.material_id, revision: reference.revision }]);
  const unavailable = selected.filter(reference => !available.some(value => same(value, reference)));
  return <fieldset disabled={disabled} className={styles.inputs}>
    <legend>项目已纳入的固定资料版本</legend>
    {available.length ? <ul className={styles.list}>{available.map(reference => <li key={`${reference.material_id}:${reference.revision}`}><label className={styles.confirmation}><input type="checkbox" checked={selected.some(value => same(value, reference))} onChange={() => toggle(reference)} /><span><code>{reference.material_id}</code> · r{reference.revision}<span className={styles.sourceLocator}>{reference.locator}{reference.span && ` · ${reference.span}`}</span></span></label></li>)}</ul> : <p>所选项目空间尚未主动纳入资料。</p>}
    {unavailable.length > 0 && <div role="alert"><p>以下已保存引用当前未出现在所选项目空间中，请核对权限和版本后再提交。</p><ul className={styles.list}>{unavailable.map(reference => <li key={`${reference.material_id}:${reference.revision}`}><span><code>{reference.material_id}</code> · r{reference.revision}</span> <button className="ac-button secondary compact" type="button" onClick={() => toggle(reference)}>移除此固定引用</button></li>)}</ul></div>}
  </fieldset>;
}
