import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { startS1Server, type S1Server } from '../../../tests/s1/server.mjs';
import { createApiClient } from './client';
import { createRelayApi } from './relay';
import { recordedClient } from './realHttpTest';

describe('接力客户端使用实际获准服务与人工会话', () => {
  let origin: string;
  let server: S1Server;
  beforeAll(async () => { server = await startS1Server({ relay: true }); origin = server.apiURL; }, 120_000);
  afterAll(async () => { if (server) await server.close(); });

  it('实际会话建立后原样保留注册表响应', async () => {
    const recorded = recordedClient({ apiURL: origin });
    const result = await createRelayApi(recorded.client).agents();
    expect(result).toEqual(await recorded.responses.at(-1)!.json());
    expect(recorded.requests.map(request => request.url)).toEqual([`${origin}/api/v1/auth/session`, `${origin}/api/v1/relay/agents`]);
    expect(recorded.requests.every(request => request.method === 'GET')).toBe(true);
  });

  it('真实请求取消后不重发，完整保留CSRF、命令身份与正文', async () => {
    const requests: Request[] = [], responses: Response[] = [];
    let cookie = '';
    const controller = new AbortController();
    const transport: typeof fetch = async (input, init) => {
      const original = new Request(input, init);
      if (cookie) original.headers.set('Cookie', cookie);
      if (original.method !== 'GET') {
        original.headers.set('Origin', origin);
        controller.abort();
      }
      const request = original.method === 'GET' ? original : new Request(original, { signal: controller.signal });
      requests.push(request.clone());
      const response = await fetch(request);
      responses.push(response.clone());
      const setCookie = response.headers.getSetCookie();
      if (setCookie.length) cookie = setCookie.map(value => value.split(';')[0]).join('; ');
      return response;
    };
    const key = crypto.randomUUID(), body = { permission_confirmed: true as const, permission_level: 'A' as const, project_settings_revision: 1 };
    await expect(createRelayApi(createApiClient(`${origin}/api/v1`, transport)).confirm(crypto.randomUUID(), body, key)).rejects.toMatchObject({ name: 'AbortError' });
    expect(requests).toHaveLength(2);
    expect(requests[0].method).toBe('GET');
    expect(requests[1].method).toBe('POST');
    expect(requests[1].signal.aborted).toBe(true);
    expect(requests[1].headers.get('X-CSRF-Token')).toBe((await responses[0].json()).csrf_token);
    expect(requests[1].headers.get('Idempotency-Key')).toBe(key);
    expect(await requests[1].json()).toEqual(body);
  });
});
