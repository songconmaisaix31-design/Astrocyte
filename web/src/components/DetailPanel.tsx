import { useEffect, useRef, useCallback, type ReactNode } from 'react';
import styles from './DetailPanel.module.css';
import { Icon } from './DesignIcons';

const FOCUSABLE = 'a[href],button:not([disabled]),textarea,input,select,[tabindex]:not([tabindex="-1"])';

interface DetailPanelProps {
  title: string;
  onClose: () => void;
  children: ReactNode;
  /** Show disabled notice for unavailable features */
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

  // Single focus-lifecycle effect: close-button focus on mount, trigger
  // restoration on final unmount.  Uses one cancellable rAF slot so
  // StrictMode setup₁→cleanup₁→setup₂ never schedules competing callbacks:
  //   setup  – cancel any pending rAF from the previous cycle, capture the
  //            trigger element (lint: refs only read in effects), schedule
  //            close-button focus.
  //   cleanup – cancel close-focus rAF, schedule a single-rAF trigger
  //             restore (no nesting).  The next setup (StrictMode remount)
  //             cancels this restore before it fires, so only the final
  //             unmount actually restores focus.
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

  // Document-level Escape listener (works regardless of focus position)
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

  // Focus trap (Tab/Shift+Tab cycle within dialog)
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
