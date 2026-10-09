export function CommandState({ pending, error, notice }: { pending: boolean; error: string | null; notice: string | null }) {
  return <>
    {pending && <p role="status">正在提交…</p>}
    {error && <p role="alert">{error}。输入已保留；请核对最新状态后重试。</p>}
    {notice && <p role="status">{notice}</p>}
  </>;
}
