import { describe, expect, it, vi } from 'vitest';
import { createApiClient, ApiError } from './client';
import { createCognitionApi } from './cognition';

const json = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
const denied = { schema_version: 1, error: { code: 'scope_denied', message: 'local session expired', retryable: true, required_action: '请重新建立本机会话', request_id: 'contract-request' } };
describe('cognition uses the existing authenticated read client', () => {
  it('bootstraps the protected cognition health and coalesces parallel reads', async () => {
    const fetchImpl = vi.fn<typeof fetch>().mockImplementation(async request => {
      const path = new URL((request as Request).url).pathname;
      return json(path.endsWith('/auth/session') ? { csrf_token: 'contract-session' } : path.endsWith('/health') ? { status: 'ok', snapshot_at: '2026-10-12T16:01:00Z', timezone: 'Asia/Shanghai' } : { topics: [], series: [] });
    });
    const api = createCognitionApi(createApiClient('http://localhost/api/v1', fetchImpl));
    const [health, focus] = await Promise.all([api.health(), api.focus()]);
    expect(health.timezone).toBe('Asia/Shanghai');
    expect(focus).toEqual({ topics: [], series: [] });
    const requests = fetchImpl.mock.calls.map(([request]) => request as Request);
    expect(requests[0].url).toBe('http://localhost/api/v1/auth/session');
    expect(requests.filter(request => request.url.endsWith('/auth/session'))).toHaveLength(1);
    expect(requests.every(request => request.method === 'GET')).toBe(true);
  });
  it('recovers one expired session on cognition health without adding another authentication flow', async () => {
    const health = { status: 'ok', snapshot_at: '2026-10-13T08:00:00Z', timezone: 'Asia/Shanghai' };
    const fetchImpl = vi.fn<typeof fetch>().mockResolvedValueOnce(json({ csrf_token: 'first' })).mockResolvedValueOnce(json(denied, 403)).mockResolvedValueOnce(json({ csrf_token: 'second' })).mockResolvedValueOnce(json(health));
    await expect(createCognitionApi(createApiClient('http://localhost/api/v1', fetchImpl)).health()).resolves.toEqual(health);
    expect(fetchImpl.mock.calls.map(([request]) => new URL((request as Request).url).pathname)).toEqual(['/api/v1/auth/session', '/api/v1/cognition/health', '/api/v1/auth/session', '/api/v1/cognition/health']);
  });
  it('keeps original root health public', async () => {
    const fetchImpl = vi.fn<typeof fetch>().mockResolvedValue(json({ status: 'ok' }));
    await createApiClient('http://localhost/api/v1', fetchImpl).GET('/health');
    expect(fetchImpl).toHaveBeenCalledTimes(1);
    expect((fetchImpl.mock.calls[0][0] as Request).url).toBe('http://localhost/api/v1/health');
  });
  it('preserves all filters and paging in the GET query, including source only as query', async () => {
    const fetchImpl = vi.fn<typeof fetch>().mockImplementation(async () => json({ events: [], total: 0, page: 3, page_size: 100 }));
    const api = createCognitionApi(createApiClient('http://localhost/api/v1', fetchImpl));
    await api.events({ topic: '真实主题 / A', kind: 'code_created', source: 'manual', start: '2026-10-11', end: '2026-10-12', page: 3, page_size: 100 });
    const request = fetchImpl.mock.calls[1][0] as Request;
    expect(Object.fromEntries(new URL(request.url).searchParams)).toEqual({ topic: '真实主题 / A', kind: 'code_created', source: 'manual', start: '2026-10-11', end: '2026-10-12', page: '3', page_size: '100' });
    expect(request.method).toBe('GET');
    expect(request.body).toBeNull();
  });
  it('preserves null values, empty days and fixed API failures without fixture fallback', async () => {
    const today = { day: '2026-10-13', focus: [{ topic: '主题', attention: 1, momentum: 0, confidence: 0.5, attention_delta: null }], activity: [], sleep: { day: '2026-10-12', status: 'not_produced', completed_at: null }, daily_counts: [], sources: [] };
    const fetchImpl = vi.fn<typeof fetch>().mockResolvedValueOnce(json({ csrf_token: 'human' })).mockResolvedValueOnce(json(today)).mockResolvedValueOnce(json({ ...denied, error: { ...denied.error, code: 'snapshot_unavailable', message: '工作台快照不可用' } }, 503));
    const api = createCognitionApi(createApiClient('http://localhost/api/v1', fetchImpl));
    await expect(api.today()).resolves.toEqual(today);
    await expect(api.focus()).rejects.toBeInstanceOf(ApiError);
    expect(fetchImpl).toHaveBeenCalledTimes(3);
  });
  it('forwards cancellation and does not bootstrap bearer requests', async () => {
    const controller = new AbortController(); controller.abort();
    const fetchImpl = vi.fn<typeof fetch>().mockRejectedValue(new DOMException('Aborted', 'AbortError'));
    await expect(createCognitionApi(createApiClient('http://localhost/api/v1', fetchImpl)).agents(controller.signal)).rejects.toMatchObject({ name: 'AbortError' });
    expect(fetchImpl).toHaveBeenCalledTimes(1);
    const bearerFetch = vi.fn<typeof fetch>().mockResolvedValue(json(denied, 403));
    await createApiClient('http://localhost/api/v1', bearerFetch).GET('/cognition/health', { headers: { Authorization: 'Bearer denied' } });
    expect(bearerFetch).toHaveBeenCalledTimes(1);
  });
});
