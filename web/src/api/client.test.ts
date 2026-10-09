import { describe, expect, it, vi } from 'vitest';
import { ApiError, createApiClient, createReadApi, createAttentionApi } from './client';

describe('generated HTTP client read boundary', () => {
  it('returns the backend empty envelope without substituting fixtures', async () => {
    const fetchImpl = vi.fn<typeof fetch>().mockImplementation(async () => new Response(JSON.stringify({ schema_version: 1, items: [], next_cursor: null }), {
      headers: { 'Content-Type': 'application/json' },
    }));
    const data = await createReadApi(createApiClient('http://127.0.0.1:8787/api/v1', fetchImpl)).listMaterials();
    expect(data).toEqual({ schema_version: 1, items: [], next_cursor: null });
    expect(fetchImpl).toHaveBeenCalledTimes(2);
    expect((fetchImpl.mock.calls[0][0] as Request).url).toBe('http://127.0.0.1:8787/api/v1/auth/session');
    expect((fetchImpl.mock.calls[1][0] as Request).url).toBe('http://127.0.0.1:8787/api/v1/materials');
  });

  it('preserves unsupported capability and recovery action without retrying', async () => {
    const detail = { schema_version: 1, error: {
      code: 'unsupported_capability', message: 'Not implemented in S0', retryable: false,
      required_action: 'Read S0 capabilities', request_id: 'request-1',
    } };
    const fetchImpl = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(detail), {
      status: 501, headers: { 'Content-Type': 'application/json' },
    }));
    try {
      await createReadApi(createApiClient('http://localhost/api/v1', fetchImpl)).getFoundation();
      throw new Error('Expected rejection');
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      expect((error as ApiError).status).toBe(501);
      expect((error as ApiError).detail).toEqual(detail);
      expect((error as ApiError).requestId).toBe('request-1');
    }
    expect(fetchImpl).toHaveBeenCalledTimes(1);
  });

  it('propagates transport failures and cancellation without returning synthetic data', async () => {
    const fetchImpl = vi.fn<typeof fetch>().mockRejectedValue(new DOMException('Aborted', 'AbortError'));
    const controller = new AbortController();
    controller.abort();
    await expect(createReadApi(createApiClient('http://localhost/api/v1', fetchImpl)).listMissions({ signal: controller.signal })).rejects.toMatchObject({ name: 'AbortError' });
    expect(fetchImpl).toHaveBeenCalledTimes(1);
    expect((fetchImpl.mock.calls[0][0] as Request).signal.aborted).toBe(true);
  });
});

const jsonResponse = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
describe('session and human command client', () => {
  it('refreshes expired human read credentials once without changing attention', async () => {
    const fetchImpl = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(jsonResponse({ csrf_token: 'first' }))
      .mockResolvedValueOnce(jsonResponse({ schema_version: 1, error: { code: 'scope_denied', message: 'local session expired' } }, 403))
      .mockResolvedValueOnce(jsonResponse({ csrf_token: 'second' }))
      .mockResolvedValueOnce(jsonResponse({ schema_version: 1, items: [], next_cursor: null }));
    await expect(createReadApi(createApiClient('http://localhost/api/v1', fetchImpl)).listMaterials()).resolves.toMatchObject({ items: [] });
    expect(fetchImpl.mock.calls.map(([request]) => (request as Request).url)).toEqual([
      'http://localhost/api/v1/auth/session', 'http://localhost/api/v1/materials',
      'http://localhost/api/v1/auth/session', 'http://localhost/api/v1/materials',
    ]);
  });
  it('does not bootstrap an Agent or convert its denied read into human access', async () => {
    const fetchImpl = vi.fn<typeof fetch>().mockResolvedValueOnce(jsonResponse({ schema_version: 1, error: { code: 'scope_denied', message: 'local session required' } }, 403));
    await createApiClient('http://localhost/api/v1', fetchImpl).GET('/jobs', { headers: { Authorization: 'Bearer invalid' } });
    expect(fetchImpl).toHaveBeenCalledTimes(1);
  });
  it('sends current CSRF and exact caller idempotency key, and never retries a failed write', async () => {
    const fetchImpl = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(jsonResponse({ schema_version: 1, actor_kind: 'human', actor_id: 'local-human', csrf_token: 'current-csrf' }))
      .mockRejectedValueOnce(new Error('lost response'));
    const service = createAttentionApi(createApiClient('http://localhost/api/v1', fetchImpl));
    await expect(service.importMaterial({ schema_version: 1, request_id: 'request', expected_version: 1, source_locator: 'arxiv:2401.00001v1', source_key: '', content_digest: '', kind: 'paper' }, 'stable-key')).rejects.toThrow('lost response');
    expect(fetchImpl).toHaveBeenCalledTimes(2);
    const request = fetchImpl.mock.calls[1][0] as Request;
    expect(request.headers.get('X-CSRF-Token')).toBe('current-csrf');
    expect(request.headers.get('Idempotency-Key')).toBe('stable-key');
    expect(request.method).toBe('POST');
  });
});
