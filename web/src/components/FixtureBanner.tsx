import { navigate } from '../router';

/**
 * Clearly indicates fixture/sample data mode across all pages.
 * Shows a dismissible banner with instructions to exit.
 */
export function FixtureBanner() {
  const exitFixture = () => {
    const url = new URL(window.location.href);
    url.searchParams.delete('fixture');
    navigate(url.pathname + url.search, true);
  };

  return (
    <div
      role="alert"
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '8px 14px',
        marginBottom: 'var(--space-4)',
        background: 'var(--color-accent-subtle)',
        border: '1px solid #dceee1',
        borderRadius: 'var(--radius-md)',
        fontSize: '12px',
        fontWeight: 'var(--weight-medium)',
        color: 'var(--color-accent-hover)',
      }}
    >
      <span>
        <strong>⚑ 示例数据模式</strong> — 当前显示固定样本数据，非真实 API 结果
      </span>
      <button
        type="button"
        onClick={exitFixture}
        style={{
          padding: 'var(--space-1) var(--space-3)',
          border: '1px solid #b2cfbc',
          borderRadius: 'var(--radius-sm)',
          background: 'transparent',
          color: 'var(--color-accent-hover)',
          cursor: 'pointer',
          fontSize: 'var(--text-sm)',
          fontWeight: 'var(--weight-medium)',
        }}
      >
        退出示例模式
      </button>
    </div>
  );
}
