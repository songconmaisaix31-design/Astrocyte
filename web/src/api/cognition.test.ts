import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import type { S1Server } from '../../../tests/s1/server.mjs';
import { ApiError } from './client';
import { createCognitionApi } from './cognition';
import { cognitionTestServer, recordedClient } from './realHttpTest';

describe('认知客户端使用生产服务、公开来源快照和既有会话', () => {
  let server: S1Server;
  beforeAll(async () => { server = await cognitionTestServer(); }, 120_000);
  afterAll(async () => { if (server) await server.close(); });

  it('并发受保护读取只建立一次实际会话', async () => {
    const recorded = recordedClient(server), api = createCognitionApi(recorded.client);
    const [health, focus] = await Promise.all([api.health(), api.focus()]);
    expect(health.timezone).toBe('Asia/Shanghai');
    expect(focus.topics.length).toBeGreaterThan(0);
    expect(recorded.requests[0].url).toBe(`${server.apiURL}/api/v1/auth/session`);
    expect(recorded.requests.filter(request => request.url.endsWith('/auth/session'))).toHaveLength(1);
    expect(recorded.requests.every(request => request.method === 'GET')).toBe(true);
    expect(health).toEqual(await recorded.responses.find(response => response.url.endsWith('/cognition/health'))!.json());
  });

  it('实际服务重启使旧会话失效，只重新建立一次会话', async () => {
    const recorded = recordedClient(server), api = createCognitionApi(recorded.client);
    const health = await api.health();
    await server.restart();
    const offset = recorded.requests.length;
    expect(await api.health()).toEqual(health);
    expect(recorded.requests.slice(offset).map(request => new URL(request.url).pathname)).toEqual(['/api/v1/cognition/health', '/api/v1/auth/session', '/api/v1/cognition/health']);
  }, 60_000);

  it('原健康接口直接返回实际健康状态', async () => {
    const recorded = recordedClient(server);
    const result = await recorded.client.GET('/health');
    expect(result.response.status).toBe(200);
    expect(recorded.requests).toHaveLength(1);
    expect(recorded.requests[0].url).toBe(`${server.apiURL}/api/v1/health`);
  });

  it('GET 完整保留调用者提供的过滤与分页参数', async () => {
    const recorded = recordedClient(server), api = createCognitionApi(recorded.client);
    const filters = { topic: '输入主题 / A', kind: 'code_created', source: 'manual', start: '2026-10-11', end: '2026-10-12', page: 3, page_size: 100 };
    const result = await api.events(filters);
    const request = recorded.requests.at(-1)!;
    expect(Object.fromEntries(new URL(request.url).searchParams)).toEqual({ ...filters, page: '3', page_size: '100' });
    expect(request.method).toBe('GET');
    expect(request.body).toBeNull();
    expect(result).toEqual(await recorded.responses.at(-1)!.json());
  });

  it('完整保留实际空日期、unknown 与连接失败', async () => {
    const recorded = recordedClient(server), api = createCognitionApi(recorded.client);
    const today = await api.today();
    expect(today).toEqual(await recorded.responses.at(-1)!.json());
    const empty = await api.events({ start: '1900-01-01', end: '1900-01-01' });
    expect(empty.events).toEqual([]);
    expect(empty.total).toBe(0);
    await server.stopAPI();
    try { await expect(api.focus()).rejects.toThrow(); }
    finally { await server.startAPI(); }
  }, 60_000);

  it('实际取消立即失败，无效 Bearer 保留拒绝且不自动读取认证', async () => {
    const recorded = recordedClient(server), api = createCognitionApi(recorded.client);
    const controller = new AbortController(); controller.abort();
    await expect(api.agents(controller.signal)).rejects.toMatchObject({ name: 'AbortError' });
    expect(recorded.requests).toHaveLength(1);
    const offset = recorded.requests.length;
    const result = await recorded.client.GET('/cognition/health', { headers: { Authorization: 'Bearer invalid' } });
    expect(result.response.status).toBe(403);
    expect(recorded.requests.length - offset).toBe(1);
    expect(new ApiError(result.response.status, result.error!).code).toBe('scope_denied');
  });
});
