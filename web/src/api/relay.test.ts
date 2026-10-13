import { describe, expect, it, vi } from 'vitest';
import { ApiError, createApiClient } from './client';
import { createRelayApi } from './relay';

describe('relay generated client uses existing human session boundaries', () => {
  it('bootstraps safe reads and preserves the exact registry response', async () => {
    const body = { schema_version: 1, agents: [], inventory: [] };
    const fetchImpl = vi.fn<typeof fetch>().mockResolvedValueOnce(new Response('{}')).mockResolvedValueOnce(new Response(JSON.stringify(body)));
    expect(await createRelayApi(createApiClient('http://localhost/api/v1', fetchImpl)).agents()).toEqual(body);
    expect(fetchImpl.mock.calls.map(([request]) => (request as Request).url)).toEqual(['http://localhost/api/v1/auth/session', 'http://localhost/api/v1/relay/agents']);
  });
  it('sends one CSRF-protected confirm and never replays delivery_unknown', async () => {
    const fetchImpl = vi.fn<typeof fetch>().mockResolvedValueOnce(new Response(JSON.stringify({ csrf_token: 'local-test-csrf' }))).mockResolvedValueOnce(new Response(JSON.stringify({ schema_version: 1, error: { code: 'delivery_unknown', message: 'unknown first attempt', request_id: 'first', retryable: false, required_action: 'read_receipts' } }), { status: 503 }));
    const key = '63d38b74-1655-43eb-98a7-5f1e2d9a4edb';
    await expect(createRelayApi(createApiClient('http://localhost/api/v1', fetchImpl)).confirm('task-test', { permission_confirmed: true, permission_level: 'A', project_settings_revision: 4 }, key)).rejects.toBeInstanceOf(ApiError);
    expect(fetchImpl).toHaveBeenCalledTimes(2);
    const request = fetchImpl.mock.calls[1][0] as Request;
    expect(request.headers.get('X-CSRF-Token')).toBe('local-test-csrf');
    expect(request.headers.get('Idempotency-Key')).toBe(key);
    expect(await request.json()).toEqual({ permission_confirmed: true, permission_level: 'A', project_settings_revision: 4 });
  });
});
