interface ErrorStateProps {
  message: string;
  onRetry?: () => void;
}

/** Recoverable error with retry. No silent fallback. */
export function ErrorState({ message, onRetry }: ErrorStateProps) {
  return (
    <div
      role="alert"
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        padding: 'var(--space-8) var(--space-6)',
        textAlign: 'center',
      }}
    >
      <span style={{ fontSize: '1.5rem', marginBottom: 'var(--space-3)' }} aria-hidden="true">⚠</span>
      <div style={{ fontSize: 'var(--text-md)', fontWeight: 'var(--weight-medium)', color: 'var(--color-text-secondary)', marginBottom: 'var(--space-2)' }}>
        暂未完成，请检查本地服务或重新加载
      </div>
      <div style={{ fontSize: 'var(--text-sm)', color: 'var(--color-text-muted)', marginBottom: 'var(--space-4)', maxWidth: '400px', wordBreak: 'break-word' }}>
        <details><summary>加载详情</summary>{message}</details>
      </div>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          style={{
            padding: 'var(--space-2) var(--space-5)',
            border: '1px solid var(--color-border)',
            borderRadius: 'var(--radius-md)',
            background: 'var(--color-surface)',
            color: 'var(--color-text)',
            cursor: 'pointer',
            fontSize: 'var(--text-base)',
            fontWeight: 'var(--weight-medium)',
          }}
        >
          重试
        </button>
      )}
    </div>
  );
}
