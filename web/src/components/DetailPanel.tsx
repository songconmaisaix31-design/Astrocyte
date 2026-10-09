import { useEffect, useRef, type ReactNode } from 'react';
import styles from './DetailPanel.module.css';

interface DetailPanelProps {
  title: string;
  onClose: () => void;
  children: ReactNode;
  /** Show disabled notice for future slice features */
  showDisabledNotice?: boolean;
  disabledNoticeText?: string;
}

export function DetailPanel({
  title,
  onClose,
  children,
  showDisabledNotice,
  disabledNoticeText = '此功能的操作按钮将在后续版本中启用',
}: DetailPanelProps) {
  const closeRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    closeRef.current?.focus();
  }, []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [onClose]);

  return (
    <div className={styles.overlay} role="dialog" aria-label={title} aria-modal="true">
      <div className={styles.header}>
        <h3 style={{ fontSize: 'var(--text-md)', fontWeight: 'var(--weight-semibold)' }}>{title}</h3>
        <button
          ref={closeRef}
          className={styles.closeBtn}
          onClick={onClose}
          aria-label="关闭"
          type="button"
        >
          ✕
        </button>
      </div>
      <div className={styles.body}>
        {children}
        {showDisabledNotice && (
          <div className={styles.disabledNotice}>
            {disabledNoticeText}
          </div>
        )}
      </div>
    </div>
  );
}

/** Reusable field display within DetailPanel */
export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className={styles.field}>
      <span className={styles.fieldLabel}>{label}</span>
      <div className={styles.fieldValue}>{children}</div>
    </div>
  );
}

export function FieldRow({ children }: { children: ReactNode }) {
  return <div className={styles.fieldRow}>{children}</div>;
}

export function MutedValue({ children }: { children: ReactNode }) {
  return <span className={styles.fieldValueMuted}>{children}</span>;
}
