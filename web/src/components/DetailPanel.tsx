import { useEffect, useRef, useCallback, type ReactNode } from 'react';
import styles from './DetailPanel.module.css';

const FOCUSABLE = 'a[href],button:not([disabled]),textarea,input,select,[tabindex]:not([tabindex="-1"])';

interface DetailPanelProps {
  title: string;
  onClose: () => void;
  children: ReactNode;
  /** Show disabled notice for unavailable features */
  showDisabledNotice?: boolean;
  disabledNoticeText?: string;
}

export function DetailPanel({
  title,
  onClose,
  children,
  showDisabledNotice,
  disabledNoticeText = '此功能的操作尚未启用',
}: DetailPanelProps) {
  const panelRef = useRef<HTMLDivElement>(null);
  const closeRef = useRef<HTMLButtonElement>(null);
  // Remember the element that opened the dialog for focus restoration
  const triggerRef = useRef<Element | null>(
    typeof document !== 'undefined' ? document.activeElement : null,
  );

  // Focus close button on mount.
  // Uses rAF so the focus happens after React's commit phase.
  const mountRafRef = useRef<number | null>(null);
  useEffect(() => {
    mountRafRef.current = requestAnimationFrame(() => {
      closeRef.current?.focus();
    });
    return () => {
      if (mountRafRef.current !== null) {
        cancelAnimationFrame(mountRafRef.current);
      }
    };
  }, []);

  // Restore focus to trigger element on final unmount.
  // Captures trigger in setup (lint rule). Each mount cancels any pending
  // restore rAF so StrictMode's mount-1 cleanup doesn't steal focus
  // after the remount effect refocuses the close button.
  const restoreRafRef = useRef<number | null>(null);
  useEffect(() => {
    // Cancel any pending restore from a previous (StrictMode-cleaned-up) mount
    if (restoreRafRef.current !== null) {
      cancelAnimationFrame(restoreRafRef.current);
      restoreRafRef.current = null;
    }
    const triggerEl = triggerRef.current;
    return () => {
      if (triggerEl && triggerEl instanceof HTMLElement) {
        restoreRafRef.current = requestAnimationFrame(() => {
          requestAnimationFrame(() => triggerEl.focus());
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
    <div
      ref={panelRef}
      className={styles.overlay}
      role="dialog"
      aria-label={title}
      aria-modal="true"
      onKeyDown={handleKeyDown}
    >
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
