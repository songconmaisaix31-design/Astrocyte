import { api, ApiError, type ApiErrorBody } from './client';
import type { operations } from './schema';

export type EventFilters = NonNullable<operations['getWorkbenchEvents']['parameters']['query']>;
export function createCognitionApi(client = api) {
  async function unwrap<T>(result: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
    const { data, error, response } = await result;
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error('认知服务未返回数据');
    return data;
  }
  return {
    health: (signal?: AbortSignal) => unwrap(client.GET('/cognition/health', { signal })),
    today: (source = '', signal?: AbortSignal) => unwrap(client.GET('/cognition/today', { params: { query: source ? { source } : {} }, signal })),
    focus: (signal?: AbortSignal) => unwrap(client.GET('/cognition/focus', { signal })),
    agents: (signal?: AbortSignal) => unwrap(client.GET('/cognition/agents', { signal })),
    filters: (signal?: AbortSignal) => unwrap(client.GET('/cognition/filters', { signal })),
    events: (filters: EventFilters, signal?: AbortSignal) => unwrap(client.GET('/cognition/events', { params: { query: filters }, signal })),
    evidence: (topic: string, page = 1, page_size = 20, signal?: AbortSignal) => unwrap(client.GET('/cognition/topics/evidence', { params: { query: { topic, page, page_size } }, signal })),
  };
}
export const cognitionApi = createCognitionApi();
