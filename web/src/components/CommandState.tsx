export function CommandState({ pending, error, errorDetail, notice }: { pending: boolean; error: string | null; errorDetail?: string | null; notice: string | null }) {
  return <>
    {pending && <p role="status">正在提交…</p>}
    {error && <p role="alert">{error} 输入已保留。</p>}
    {error && errorDetail && <details><summary>处理详情</summary><p>{errorDetail}</p></details>}
    {notice && <p role="status">{notice}</p>}
  </>;
}
