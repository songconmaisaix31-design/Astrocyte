import type { EvidenceReference } from './judgmentApi';
import { evidenceKey, type ProjectEvidence } from './useProjectEvidence';
import styles from './DelegationWorkspace.module.css';

export function JudgmentEvidencePicker({ title, available, selected, disabled, onChange }: {
  title: string;
  available: ProjectEvidence[];
  selected: EvidenceReference[];
  disabled: boolean;
  onChange: (references: EvidenceReference[]) => void;
}) {
  const selectedKeys = new Set(selected.map(evidenceKey));
  const unavailable = selected.filter(reference => !available.some(item => evidenceKey(item.reference) === evidenceKey(reference)));
  const toggle = (reference: EvidenceReference) => onChange(selectedKeys.has(evidenceKey(reference)) ? selected.filter(item => evidenceKey(item) !== evidenceKey(reference)) : [...selected, reference]);
  return <fieldset className={styles.fields} disabled={disabled}><legend>{title}</legend>
    {available.length ? available.map(item => <label className={styles.confirmation} key={evidenceKey(item.reference)}><input type="checkbox" checked={selectedKeys.has(evidenceKey(item.reference))} onChange={() => toggle(item.reference)} /><span>{item.label}<br /><code>{item.reference.digest}</code></span></label>) : <p>该项目空间没有可读取的固定资料版本。</p>}
    {unavailable.length > 0 && <div role="alert"><p>原选择的固定资料当前不在可读取范围内，请先核查权限与版本。</p>{unavailable.map(reference => <p key={evidenceKey(reference)}><code>{reference.material_id}</code> · r{reference.revision} <button type="button" className="ac-button secondary compact" onClick={() => toggle(reference)}>移除此固定依据</button></p>)}</div>}
  </fieldset>;
}
