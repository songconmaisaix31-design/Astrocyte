import { useEffect, useRef, useCallback, type ReactNode } from 'react';
import styles from './DetailPanel.module.css';
import { Icon } from './DesignIcons';

const FOCUSABLE = 'a[href],button:not([disabled]),textarea,input,select,[tabindex]:not([tabindex="-1"])';

interface DetailPanelProps {
  title: string;
  onClose: () => void;
  children: ReactNode;
  /** 显示当前不可用功能的说明。 */
  showDisabledNotice?: boolean;
  disabledNoticeText?: string;
  disabledActions?: string[];
}

export function DetailPanel({
  title,
  onClose,
  children,
  showDisabledNotice,
  disabledNoticeText = '此功能的操作尚未启用',
  disabledActions = [],
}: DetailPanelProps) {
  const panelRef = useRef<HTMLDivElement>(null);
  const closeRef = useRef<HTMLButtonElement>(null);
  const triggerRef = useRef<Element | null>(
    typeof document !== 'undefined' ? document.activeElement : null,
  );

  // 挂载时聚焦关闭按钮，最终卸载时恢复触发元素；取消前次 rAF，避免 StrictMode 重复回调。
  const rafRef = useRef<number | null>(null);
  useEffect(() => {
    if (rafRef.current !== null) {
      cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
    }
    const triggerEl = triggerRef.current;
    rafRef.current = requestAnimationFrame(() => {
      closeRef.current?.focus();
    });
    return () => {
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
      }
      if (triggerEl instanceof HTMLElement) {
        rafRef.current = requestAnimationFrame(() => {
          rafRef.current = null;
          if (triggerEl.isConnected) triggerEl.focus();
        });
      }
    };
  }, []);

  // 从整个文档接收 Escape，兼容详情外的焦点。
  useEffect(() => {
    const onDocKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', onDocKey);
    return () => document.removeEventListener('keydown', onDocKey);
  }, [onClose]);

  useEffect(() => {
    const previous = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = previous; };
  }, []);

  // Tab 和 Shift+Tab 在详情内部循环焦点。
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Tab') {
        const panel = panelRef.current;
        if (!panel) return;

        const focusables = panel.querySelectorAll<HTMLElement>(FOCUSABLE);
        if (focusables.length === 0) {
          e.preventDefault();
          return;
        }

        const first = focusables[0];
        const last = focusables[focusables.length - 1];

        if (e.shiftKey) {
          if (document.activeElement === first) {
            e.preventDefault();
            last.focus();
          }
        } else {
          if (document.activeElement === last) {
            e.preventDefault();
            first.focus();
          }
        }
      }
    },
    [],
  );

  return (
    <div className="ac-overlay" onClick={event => { if (event.target === event.currentTarget) onClose(); }}>
    <div
      ref={panelRef}
      className={styles.overlay}
      role="dialog"
      aria-label={title}
      aria-modal="true"
      onKeyDown={handleKeyDown}
    >
      <div className={styles.header}>
        <div><span className="ac-overline">CONTEXT & PROVENANCE</span><h3 style={{ fontSize: '20px', fontWeight: 'var(--weight-semibold)' }}>{title}</h3></div>
        <button
          ref={closeRef}
          className={styles.closeBtn}
          onClick={onClose}
          aria-label="关闭"
          type="button"
        >
          <Icon name="close" size={18} />
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
      {disabledActions.length > 0 && <footer className="ac-detail-actions">{disabledActions.map(label => <button key={label} className="ac-button disabled compact" type="button" disabled title={disabledNoticeText}>{label}</button>)}</footer>}
    </div>
    </div>
  );
}

/** 详情面板中复用的字段显示。 */
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
