/// <reference types="node" />
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { startS1Server, type S1Server } from '../../../tests/s1/server.mjs';
import { ApiError, createApiClient, createReadApi, createAttentionApi } from './client';

describe('生产 Go HTTP 客户端', () => {
  let server: S1Server;
  let cookie = '';
  const requests: Request[] = [];
  const responses: Response[] = [];
  // Node 没有浏览器 cookie jar；保留真实服务的 Set-Cookie，并记录实际请求和响应。
  const transport: typeof fetch = async (input, init) => {
    const request = new Request(input, init);
    if (cookie) request.headers.set('Cookie', cookie);
    if (request.method !== 'GET') request.headers.set('Origin', server.apiURL);
    requests.push(request.clone());
    const response = await fetch(request);
    responses.push(response.clone());
    const setCookie = response.headers.getSetCookie();
    if (setCookie.length) cookie = setCookie.map(value => value.split(';')[0]).join('; ');
    return response;
  };
  const client = () => createApiClient(`${server.apiURL}/api/v1`, transport);
  beforeAll(async () => { server = await startS1Server(); }, 120_000);
  afterAll(async () => { if (server) await server.close({ preserveData: true }); });

  it('读取本机实际独立观察值并完整保留 unknown', async () => {
    const offset = requests.length;
    const result = await createReadApi(client()).listLocalAgents();
    expect(requests.slice(offset).map(request => new URL(request.url).pathname)).toEqual(['/api/v1/auth/session', '/api/v1/local-agents']);
    expect(result).toEqual(await responses.at(-1)!.json());
    for (const item of result.items) {
      expect(['available', 'unavailable', 'unknown']).toContain(item.installed.status);
      for (const observation of Object.values(item.capabilities)) {
        expect(['supported', 'unsupported', 'unknown']).toContain(observation.status);
        expect(observation.reason).toBeTruthy();
      }
    }
  });

  it('返回 SQLite 的实际空集合', async () => {
    const offset = requests.length;
    const result = await createReadApi(client()).listMaterials();
    expect(result).toEqual({ schema_version: 1, items: [], next_cursor: null });
    expect(requests.slice(offset).map(request => new URL(request.url).pathname)).toEqual(['/api/v1/auth/session', '/api/v1/materials']);
    expect(server.query('SELECT COUNT(*) AS count FROM attention_materials')[0].count).toBe(0);
  });

  it('实际未实现接口保留 unsupported_capability 和恢复说明且不重发', async () => {
    const http = client();
    const session = await http.GET('/auth/session');
    expect(session.data?.csrf_token).toBeTruthy();
    const offset = requests.length;
    const key = crypto.randomUUID();
    const result = await http.POST('/sessions/{id}/resume', {
      params: { path: { id: 'unbound-session' }, header: { 'Idempotency-Key': key } },
      body: { schema_version: 1, request_id: key, expected_version: 1, context_packet_id: null },
      headers: { 'X-CSRF-Token': session.data!.csrf_token },
    });
    expect(result.response.status).toBe(501);
    expect(result.error?.error.code).toBe('unsupported_capability');
    expect(result.error?.error.required_action).toBeTruthy();
    expect(result.error?.error.request_id).toBeTruthy();
    expect(requests.length - offset).toBe(1);
  });

  it('实际取消请求立即失败且不返回集合', async () => {
    const controller = new AbortController();
    controller.abort();
    const offset = requests.length;
    await expect(createReadApi(client()).listMissions({ signal: controller.signal })).rejects.toMatchObject({ name: 'AbortError' });
    expect(requests.length - offset).toBe(1);
    expect(requests.at(-1)!.signal.aborted).toBe(true);
  });

  it('服务重启使旧会话失效，受保护读取只重新建立一次会话', async () => {
    const read = createReadApi(client());
    await read.listMaterials();
    await server.restart();
    const offset = requests.length;
    await expect(read.listMaterials()).resolves.toMatchObject({ items: [] });
    expect(requests.slice(offset).map(request => new URL(request.url).pathname)).toEqual(['/api/v1/materials', '/api/v1/auth/session', '/api/v1/materials']);
  }, 60_000);

  it('实际拒绝无效 Agent 凭据并保留 Agent 身份', async () => {
    const offset = requests.length;
    const result = await client().GET('/jobs', { headers: { Authorization: 'Bearer invalid' } });
    expect(result.response.status).toBe(403);
    expect(result.error?.error.code).toBe('scope_denied');
    expect(requests.length - offset).toBe(1);
    expect(requests.at(-1)!.headers.get('Authorization')).toBe('Bearer invalid');
  });

  it('实际校验拒绝保留 CSRF 和调用方 Idempotency-Key，写入只发送一次', async () => {
    const offset = requests.length;
    const responseOffset = responses.length;
    const key = crypto.randomUUID();
    await expect(createAttentionApi(client()).importMaterial({
      schema_version: 1, request_id: key, expected_version: 1,
      source_locator: '', source_key: '', content_digest: '', kind: 'paper',
    }, key)).rejects.toBeInstanceOf(ApiError);
    expect(requests.length - offset).toBe(2);
    const session = await responses[responseOffset].json() as { csrf_token: string };
    expect(requests.at(-1)!.headers.get('X-CSRF-Token')).toBe(session.csrf_token);
    expect(requests.at(-1)!.headers.get('Idempotency-Key')).toBe(key);
    expect(requests.at(-1)!.method).toBe('POST');
    expect(responses.at(-1)!.status).toBe(400);
    expect(server.query('SELECT COUNT(*) AS count FROM attention_materials')[0].count).toBe(0);
  });

  it('实际服务断开后传播连接失败', async () => {
    const read = createReadApi(client());
    await read.listMaterials();
    await server.close({ preserveData: true });
    await expect(read.listMaterials()).rejects.toThrow();
  });
});
