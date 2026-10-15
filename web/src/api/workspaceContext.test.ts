/// <reference types="node" />
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { startS1Server, type S1Server } from '../../../tests/s1/server.mjs';
import { recordedClient } from './realHttpTest';
import { createWorkspaceContextApi, type WorkspaceContextPacketRequest } from './workspaceContext';

describe('安全工作区上下文客户端使用生产 Go、HTTP 与 SQLite', () => {
  let server: S1Server;
  beforeAll(async () => { server = await startS1Server(); }, 120_000);
  afterAll(async () => { if (server) await server.close(); });
  const packetRequest = (key: string): WorkspaceContextPacketRequest => ({ schema_version: 1, request_id: key, expected_version: 0, grant_id: crypto.randomUUID(), grant_epoch: 1, max_bytes: 131072, checks: ['核查当前服务实际配置和原请求身份'] });

  it('请求身份不同立即拒绝，不建立会话或写入', async () => {
    const recorded = recordedClient(server), context = createWorkspaceContextApi(recorded.client);
    await expect(context.buildPacket(packetRequest(crypto.randomUUID()), crypto.randomUUID())).rejects.toThrow('请求身份必须与原命令标识相同');
    expect(recorded.requests).toHaveLength(0);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_context_packets')).toEqual([{ count: 0 }]);
  });

  it('实际未配置快照来源保留501，完整传递字节预算、正文和CSRF且不重发', async () => {
    const recorded = recordedClient(server), key = crypto.randomUUID(), body = packetRequest(key);
    await expect(createWorkspaceContextApi(recorded.client).buildPacket(body, key)).rejects.toMatchObject({ status: 501, code: 'unsupported_capability' });
    expect(recorded.requests.map(request => new URL(request.url).pathname)).toEqual(['/api/v1/auth/session', '/api/v1/context-packets']);
    const request = recorded.requests.at(-1)!;
    expect(request.method).toBe('POST');
    expect(request.headers.get('Idempotency-Key')).toBe(key);
    expect(request.headers.get('X-CSRF-Token')).toBe((await recorded.responses[0].json()).csrf_token);
    expect(await request.json()).toEqual(body);
    expect(recorded.responses.at(-1)!.status).toBe(501);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_context_packets')).toEqual([{ count: 0 }]);
  });

  it('实际缺少原控制器绑定明确拒绝，SQLite不生成会话身份', async () => {
    const recorded = recordedClient(server), key = crypto.randomUUID();
    const body = { schema_version: 1 as const, request_id: key, expected_version: 0 as const, session_id: crypto.randomUUID(), packet_id: crypto.randomUUID(), member_id: crypto.randomUUID() };
    await expect(createWorkspaceContextApi(recorded.client).bindSession(body, key)).rejects.toMatchObject({ status: 501, code: 'unsupported_capability' });
    expect(await recorded.requests.at(-1)!.json()).toEqual(body);
    expect(recorded.requests.at(-1)!.headers.get('Idempotency-Key')).toBe(key);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_session_bindings')).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
  });

  it('实际不存在的固定上下文和无效Agent凭据分别返回404及403', async () => {
    const recorded = recordedClient(server), id = crypto.randomUUID();
    await expect(createWorkspaceContextApi(recorded.client).packet(id)).rejects.toMatchObject({ status: 404, code: 'not_found' });
    expect(new URL(recorded.requests.at(-1)!.url).pathname).toBe(`/api/v1/context-packets/${id}`);
    const denied = await recorded.client.GET('/context-packets/{id}', { params: { path: { id } }, headers: { Authorization: 'Bearer invalid' } });
    expect(denied.response.status).toBe(403);
    expect(denied.error?.error.code).toBe('scope_denied');
  });

  it('实际取消在写入前失败，服务断开后传播原连接失败', async () => {
    const recorded = recordedClient(server), context = createWorkspaceContextApi(recorded.client), key = crypto.randomUUID();
    const controller = new AbortController(); controller.abort();
    await expect(context.buildPacket(packetRequest(key), key, controller.signal)).rejects.toMatchObject({ name: 'AbortError' });
    expect(recorded.requests).toHaveLength(1);
    expect(recorded.requests[0].signal.aborted).toBe(true);
    expect(recorded.requests.every(request => request.method === 'GET')).toBe(true);
    await server.close();
    await expect(context.packet(crypto.randomUUID())).rejects.toThrow();
  });
});
