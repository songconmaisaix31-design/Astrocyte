import { useId, useState } from 'react';
import type { components } from '../../../api/schema';
import styles from './ProposalWorkspace.module.css';

interface Props {
  status: components['schemas']['ProposalV1']['status'];
  revision: number;
  disabled: boolean;
  pending: boolean;
  onSubmit: () => void;
  onDecline: () => void;
  onApprove: (expiresAt: string) => void;
}

export function ProposalReviewActions({ status, revision, disabled, pending, onSubmit, onDecline, onApprove }: Props) {
  const [checked, setChecked] = useState(false);
  const [expiresAt, setExpiresAt] = useState('');
  const id = useId();
  if (status === 'draft') return <div className={styles.actions}><button className="ac-button" type="button" disabled={disabled || pending} onClick={onSubmit}>提交审阅</button></div>;
  if (status !== 'in_review') return null;
  return <section aria-label="批准当前提案" className={styles.review}>
    <h4>批准固定版本 r{revision}</h4>
    <p>批准记录保存当前版本与执行范围。后续修改产生新版本，需要重新审阅和批准。</p>
    <form className={styles.formFields} onSubmit={event => { event.preventDefault(); if (checked && !disabled && !pending) onApprove(new Date(expiresAt).toISOString()); }}>
      <label htmlFor={`${id}-expiry`}><span>批准有效期</span><input id={`${id}-expiry`} type="datetime-local" value={expiresAt} onChange={event => setExpiresAt(event.target.value)} required disabled={disabled || pending} /><small>有效期必须晚于当前时间并处于提案截止时间以内。</small></label>
      <label htmlFor={`${id}-review`} className={styles.confirmation}><input id={`${id}-review`} type="checkbox" checked={checked} onChange={event => setChecked(event.target.checked)} disabled={disabled || pending} /><span>我已核对当前固定版本、范围、资源上限和停止条件</span></label>
      <div className={styles.actions}><button className="ac-button" type="submit" disabled={!checked || !expiresAt || disabled || pending}>批准当前版本</button><button className="ac-button secondary" type="button" disabled={disabled || pending} onClick={onDecline}>拒绝当前版本</button></div>
    </form>
  </section>;
}
