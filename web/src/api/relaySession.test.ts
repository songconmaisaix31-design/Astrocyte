import { expect, it, vi } from 'vitest';
import { ApiError, createApiClient } from './client';
import { createRelaySessionApi } from './relaySession';

it('uses the existing session bootstrap and only GET with generated strict query fields', async () => {
  const body = { events: [], coverage: [], total: 0, page: 2, page_size: 50 };
  const fetchImpl = vi.fn<typeof fetch>().mockResolvedValueOnce(new Response('{}')).mockResolvedValueOnce(new Response(JSON.stringify(body)));
  const query = { task_id: 'cf1497b0-e092-43c0-80a9-33449bb57a40', start: '2026-10-14T00:00:00Z', end: '2026-10-15T00:00:00Z', page: 2, page_size: 50 };
  expect(await createRelaySessionApi(createApiClient('http://localhost/api/v1', fetchImpl)).events(query)).toEqual(body);
  const requests = fetchImpl.mock.calls.map(([request]) => request as Request);
  expect(requests.every(request => request.method === 'GET')).toBe(true);
  expect(new URL(requests[0].url).pathname).toBe('/api/v1/auth/session');
  expect(Object.fromEntries(new URL(requests[1].url).searchParams)).toEqual(Object.fromEntries(Object.entries(query).map(([key, value]) => [key, String(value)])));
});

it('preserves a 503 unknown result without retrying capture or issuing commands', async () => {
  const fetchImpl = vi.fn<typeof fetch>().mockResolvedValueOnce(new Response('{}')).mockResolvedValueOnce(new Response(JSON.stringify({ schema_version: 1, error: { code: 'provider_unavailable', message: 'Relay provider unavailable', required_action: 'inspect_local_relay_provider', retryable: false, request_id: 'public-request' } }), { status: 503 }));
  await expect(createRelaySessionApi(createApiClient('http://localhost/api/v1', fetchImpl)).events()).rejects.toBeInstanceOf(ApiError);
  expect(fetchImpl).toHaveBeenCalledTimes(2);
  expect(fetchImpl.mock.calls.every(([request]) => (request as Request).method === 'GET')).toBe(true);
});
