/// <reference types="node" />
import { fileURLToPath } from 'node:url';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { startS1Server, type S1Server } from '../../../tests/s1/server.mjs';
import { humanAPI } from '../../../tests/s1/api.mjs';
import { recordedClient } from './realHttpTest';
import { createRepositoryContextApi, type RepositoryMapQuery, type RepositoryQueryRequest } from './repositoryContext';

describe('仓库上下文客户端使用生产 Go、HTTP 与 SQLite', () => {
  let server: S1Server;
  let projectID: string;
  const query: RepositoryMapQuery = { purpose: '核查本轨公开仓库的实际接口', root: ['web/src/api', 'web/src/components'], max_files: 32, max_bytes: 65536, max_results: 10, max_depth: 2 };
  const body: RepositoryQueryRequest = { schema_version: 1, purpose: query.purpose, roots: query.root, budget: { max_files: query.max_files, max_bytes: query.max_bytes, max_results: query.max_results, max_depth: query.max_depth } };

  beforeAll(async () => {
    server = await startS1Server();
    const human = await humanAPI(server.apiURL);
    const space = await human.write('/project-spaces', { title: '仓库上下文客户端实际核查' }, { status: 201 });
    const registered = await human.write('/local-projects', { root: fileURLToPath(new URL('../../../', import.meta.url)), name: 'standards-frontend', space_id: space.space.id });
    projectID = registered.project.id;
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_projects WHERE id = ?', projectID)[0].count).toBe(1);
  }, 120_000);
  afterAll(async () => { if (server) await server.close(); });

  it('实际会话与重复 root 查询完整传递，未配置能力保留 503', async () => {
    const recorded = recordedClient(server), repository = createRepositoryContextApi(recorded.client);
    await expect(repository.map(projectID, query)).rejects.toMatchObject({ status: 503, code: 'provider_unavailable' });
    expect(recorded.requests.map(request => request.method)).toEqual(['GET', 'GET']);
    expect(new URL(recorded.requests[0].url).pathname).toBe('/api/v1/auth/session');
    const outgoing = new URL(recorded.requests[1].url);
    expect(outgoing.pathname).toBe(`/api/v1/local-projects/${projectID}/repository-context/map`);
    expect(outgoing.searchParams.getAll('root')).toEqual(query.root);
    outgoing.searchParams.delete('root');
    expect(Object.fromEntries(outgoing.searchParams)).toEqual({ purpose: query.purpose, max_files: '32', max_bytes: '65536', max_results: '10', max_depth: '2' });
    const response = await recorded.responses.at(-1)!.json();
    expect(response.error.required_action).toBe('configure_the_verified_local_repository_adapter');
    expect(response).not.toHaveProperty('repository_context');
  });

  it('四种实际 POST 保留正文与有效 CSRF，各命令仅发送一次', async () => {
    const recorded = recordedClient(server), repository = createRepositoryContextApi(recorded.client);
    const search = { ...body, query: 'QueryState' };
    const impact = { ...body, symbol_id: 'QueryState' };
    const index = { ...body, action: 'index' as const };
    for (const [path, input, invoke] of [
      ['search', search, () => repository.search(projectID, search)],
      ['impact', impact, () => repository.impact(projectID, impact)],
      ['index', index, () => repository.index(projectID, index)],
      ['map/refresh', body, () => repository.refreshMap(projectID, body)],
    ] as const) {
      const offset = recorded.requests.length;
      await expect(invoke()).rejects.toMatchObject({ status: 503, code: 'provider_unavailable' });
      const requests = recorded.requests.slice(offset);
      expect(requests.map(request => request.method)).toEqual(['GET', 'POST']);
      expect(new URL(requests[1].url).pathname).toBe(`/api/v1/local-projects/${projectID}/repository-context/${path}`);
      expect(requests[1].headers.has('X-CSRF-Token')).toBe(true);
      expect(requests[1].headers.has('Idempotency-Key')).toBe(false);
      expect(await requests[1].clone().json()).toEqual(input);
    }
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')[0].count).toBe(0);
  });

  it('实际输入校验与缺失 CSRF 分别返回 400 和 403', async () => {
    const recorded = recordedClient(server), repository = createRepositoryContextApi(recorded.client);
    await expect(repository.map(projectID, { ...query, max_files: 0 })).rejects.toMatchObject({ status: 400, code: 'validation_failed' });
    const offset = recorded.requests.length;
    const denied = await recorded.client.POST('/local-projects/{id}/repository-context/index', { params: { path: { id: projectID } }, body: { ...body, action: 'index' } });
    expect(denied.response.status).toBe(403);
    expect(recorded.requests.length - offset).toBe(1);
    expect(recorded.requests.at(-1)!.headers.has('X-CSRF-Token')).toBe(false);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')[0].count).toBe(0);
  });

  it('实际取消立即失败，预先取消的 POST 不提交命令', async () => {
    const recorded = recordedClient(server), repository = createRepositoryContextApi(recorded.client);
    const controller = new AbortController();
    controller.abort();
    await expect(repository.map(projectID, query, controller.signal)).rejects.toMatchObject({ name: 'AbortError' });
    await expect(repository.search(projectID, { ...body, query: 'QueryState' }, controller.signal)).rejects.toMatchObject({ name: 'AbortError' });
    expect(recorded.requests.every(request => request.method === 'GET' && request.signal.aborted)).toBe(true);
    expect(recorded.responses).toHaveLength(0);
  });

  it('实际服务关闭后传播连接失败，不返回地图或重新提交 POST', async () => {
    const recorded = recordedClient(server), repository = createRepositoryContextApi(recorded.client);
    await expect(repository.map(projectID, query)).rejects.toMatchObject({ status: 503 });
    await server.close();
    const offset = recorded.requests.length;
    await expect(repository.map(projectID, query)).rejects.toThrow();
    expect(recorded.requests.length - offset).toBe(1);
    expect(recorded.requests.at(-1)!.method).toBe('GET');
  });
});
