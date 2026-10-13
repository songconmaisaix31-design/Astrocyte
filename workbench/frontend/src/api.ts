import { useEffect, useState } from 'react';

export type Event = { event_id: string; kind: string; topic: string; tokens: number | null; scope: string; ts: string };
export type Topic = { topic: string; attention: number; momentum: number; confidence: number };
export type Health = { status: string; snapshot_at: string | null; timezone: 'Asia/Shanghai' };
export type Today = {
  day: string; focus: (Topic & { attention_delta: number | null })[];
  activity: Event[]; sleep: { day: string; status: string; completed_at: string | null };
  daily_counts: { day: string; count: number }[]; sources: { source: string; count: number }[];
};
export type Focus = { topics: Topic[]; series: { topic: string; points: { ts: string; attention: number }[] }[] };
export type Agents = { groups: { source: string; today: number; days7: number; days30: number; recent: Event[] }[] };
export type Events = { events: Event[]; total: number; page: number; page_size: number };
export type TopicEvidence = Events & { topic: string; aggregates: { day: string; kind: string; count: number }[] };
export type Filters = { topics: string[]; kinds: string[]; sources: string[] };

export function query(path: string, values: Record<string, string | number>) {
  const params = new URLSearchParams();
  Object.entries(values).forEach(([key, value]) => { if (value !== '') params.set(key, String(value)); });
  return `${path}?${params}`;
}

export function useApi<T>(url: string, revision = 0) {
  const [state, setState] = useState<{ data?: T; error?: string; loading: boolean }>({ loading: true });
  useEffect(() => {
    const controller = new AbortController();
    setState({ loading: true });
    fetch(url, { signal: controller.signal, cache: 'no-store' }).then(async response => {
      if (!response.ok) throw new Error(`请求未完成（HTTP ${response.status}）`);
      return response.json() as Promise<T>;
    }).then(data => setState({ data, loading: false })).catch(error => {
      if (!controller.signal.aborted) setState({ loading: false, error: error instanceof SyntaxError ? '服务响应格式异常' : error instanceof TypeError ? '无法连接本机服务' : error.message });
    });
    return () => controller.abort();
  }, [url, revision]);
  return state;
}
