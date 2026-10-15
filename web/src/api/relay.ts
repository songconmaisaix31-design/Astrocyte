import { api, ApiError, type ApiErrorBody } from './client';
import type { components } from './schema';

export type RelayAgent = components['schemas']['RelayAgentV1'];
export type RelayProject = components['schemas']['RelayProjectV1'];
export type RelayTaskView = components['schemas']['RelayTaskViewV1'];
export type RelayReceipt = components['schemas']['RelayReceiptV1'];
export type RelayActionResult = components['schemas']['RelayActionResultV1'];
export type RelayTaskResult = components['schemas']['RelayTaskResultV1'];
export type RelayEvent = components['schemas']['WorkbenchEventV1'];
export type RelayFacts = components['schemas']['RelayFactsV1'];

export function createRelayApi(client = api) {
  async function unwrap<T>(result: Promise<{ data?: T; error?: ApiErrorBody; response: Response }>): Promise<T> {
    const { data, error, response } = await result;
    if (error) throw new ApiError(response.status, error);
    if (!data) throw new Error('接力服务未返回数据');
    return data;
  }
  async function headers(key: string) {
    if (!key.trim()) throw new Error('接力操作需要唯一请求标识');
    const session = await unwrap(client.GET('/auth/session'));
    return { 'Idempotency-Key': key, 'X-CSRF-Token': session.csrf_token };
  }
  return {
    agents: (signal?: AbortSignal) => unwrap(client.GET('/relay/agents', { signal })),
    projects: (signal?: AbortSignal) => unwrap(client.GET('/relay/projects', { signal })),
    tasks: (signal?: AbortSignal) => unwrap(client.GET('/relay/tasks', { signal })),
    task: (id: string, signal?: AbortSignal) => unwrap(client.GET('/relay/tasks/{id}', { params: { path: { id } }, signal })),
    events: (id: string, signal?: AbortSignal) => unwrap(client.GET('/relay/tasks/{id}/events', { params: { path: { id } }, signal })),
    receipts: (id: string, signal?: AbortSignal) => unwrap(client.GET('/relay/tasks/{id}/receipts', { params: { path: { id } }, signal })),
    facts: (id: string, signal?: AbortSignal) => unwrap(client.GET('/relay/tasks/{id}/facts', { params: { path: { id } }, signal })),
    intent: async (body: components['schemas']['RelayIntentRequestV1'], key: string) => unwrap(client.POST('/relay/intents', { body, params: { header: await headers(key) } })),
    prepare: async (id: string, body: components['schemas']['RelayConfirmRequestV1'], key: string) => unwrap(client.POST('/relay/intents/{id}/prepare', { body, params: { path: { id }, header: await headers(key) } })),
    confirm: async (id: string, body: components['schemas']['RelayConfirmRequestV1'], key: string) => unwrap(client.POST('/relay/intents/{id}/confirm', { body, params: { path: { id }, header: await headers(key) } })),
    send: async (id: string, input: string, key: string) => unwrap(client.POST('/relay/tasks/{id}/send', { body: { input }, params: { path: { id }, header: await headers(key) } })),
    read: async (id: string, key: string) => unwrap(client.POST('/relay/tasks/{id}/read', { body: {}, params: { path: { id }, header: await headers(key) } })),
    stop: async (id: string, key: string) => unwrap(client.POST('/relay/tasks/{id}/stop', { body: {}, params: { path: { id }, header: await headers(key) } })),
    resume: async (id: string, key: string) => unwrap(client.POST('/relay/tasks/{id}/resume', { body: {}, params: { path: { id }, header: await headers(key) } })),
    transition: async (id: string, body: components['schemas']['RelayTransitionRequestV1'], key: string) => unwrap(client.POST('/relay/tasks/{id}/transition', { body, params: { path: { id }, header: await headers(key) } })),
    reconcileEvents: async (id: string, key: string) => unwrap(client.POST('/relay/tasks/{id}/events/reconcile', { body: {}, params: { path: { id }, header: await headers(key) } })),
  };
}
export const relayApi = createRelayApi();
