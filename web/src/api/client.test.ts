import { describe, expect, it, vi } from 'vitest';
import { ApiError, createApiClient, createReadApi } from './client';

describe('generated HTTP client read boundary', () => {
  it('returns the backend empty envelope without substituting fixtures', async () => {
    const fetchImpl = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ schema_version: 1, items: [], next_cursor: null }), {
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
