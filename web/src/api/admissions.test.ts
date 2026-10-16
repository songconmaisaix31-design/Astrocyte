/// <reference types="node" />
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { startS1Server, type S1Server } from '../../../tests/s1/server.mjs';
import { createApiClient } from './client';
import { createAdmissionsApi } from './admissions';

describe('准入生成客户端的实际 Go HTTP 边界', () => {
  let server: S1Server;
  let cookie = '';
  const requests: Request[] = [];
  const transport: typeof fetch = async (input, init) => {
    const request = new Request(input, init);
    if (cookie) request.headers.set('Cookie', cookie);
    if (request.method !== 'GET') request.headers.set('Origin', server.apiURL);
    requests.push(request.clone());
    const response = await fetch(request);
    if (response.headers.getSetCookie().length) cookie = response.headers.getSetCookie().map(value => value.split(';')[0]).join('; ');
    return response;
  };
  const client = () => createAdmissionsApi(createApiClient(`${server.apiURL}/api/v1`, transport));
  beforeAll(async () => { server = await startS1Server(); }, 120_000);
  afterAll(async () => { if (server) await server.close({ preserveData: true }); });

  it('请求身份不一致时在发送前立即失败', async () => {
    const count = requests.length;
    await expect(client().admit('uncreated-opportunity', { schema_version: 1, request_id: crypto.randomUUID(), expected_version: 1, expected_revision: 1 }, crypto.randomUUID())).rejects.toThrow('原命令标识');
    expect(requests).toHaveLength(count);
  });

  it('实际输入校验拒绝 revision0，保留真实状态及相同正文和请求头身份', async () => {
    const count = requests.length;
    const key = crypto.randomUUID();
    await expect(client().admit('uncreated-opportunity', { schema_version: 1, request_id: key, expected_version: 1, expected_revision: 0 }, key)).rejects.toMatchObject({ status: 400, code: 'validation_failed' });
    expect(requests.slice(count).map(request => request.method)).toEqual(['GET', 'POST']);
    expect(requests.at(-1)!.headers.get('Idempotency-Key')).toBe(key);
    expect(requests.at(-1)!.headers.get('X-CSRF-Token')).toBeTruthy();
    expect((await requests.at(-1)!.json()).request_id).toBe(key);
    expect(server.query('SELECT COUNT(*) AS count FROM attention_admissions')).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_proposals')).toEqual([{ count: 0 }]);
  });

  it('实际不存在的准入读取保留404及请求身份', async () => {
    await expect(client().read('uncreated-admission')).rejects.toMatchObject({ status: 404, code: 'not_found', requestId: expect.any(String) });
    expect(server.query('SELECT COUNT(*) AS count FROM attention_admissions')).toEqual([{ count: 0 }]);
  });
});
