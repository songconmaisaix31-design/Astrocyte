/// <reference types="node" />
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { startS1Server, type S1Server } from '../../../tests/s1/server.mjs';
import { humanAPI, importMaterial } from '../../../tests/s1/api.mjs';
import { createProposalsApi, type ProposalContent } from './proposals';
import { recordedClient } from './realHttpTest';
import { createSwarmApi, type SwarmCreateRequest } from './swarm';
import type { components } from './schema';

type S = components['schemas'];

describe('蜂群客户端使用生产 Go、实际 HTTP 与 SQLite', () => {
  let server: S1Server;
  const root = fileURLToPath(new URL('../../../', import.meta.url));
  beforeAll(async () => { server = await startS1Server(); }, 120_000);
  afterAll(async () => { if (server) await server.close(); });
  const createRequest = (key: string): SwarmCreateRequest => ({ schema_version: 1, request_id: key, expected_version: 0, grant_id: crypto.randomUUID(), epoch: 1, proposal_revision: 1, project_id: crypto.randomUUID(), root });

  it('实际空集合和不存在的任务不产生执行事实', async () => {
    const recorded = recordedClient(server), swarm = createSwarmApi(recorded.client);
    expect(await swarm.missions()).toEqual({ schema_version: 1, items: [], next_cursor: null });
    const id = crypto.randomUUID();
    for (const read of [() => swarm.mission(id), () => swarm.workItems(id), () => swarm.timeline(id)]) {
      await expect(read()).rejects.toMatchObject({ status: 404, code: 'not_found' });
    }
    expect(recorded.requests.every(request => request.method === 'GET')).toBe(true);
    expect(server.query('SELECT COUNT(*) AS count FROM swarm_missions')).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
  });

  it('请求身份不一致在建立会话前拒绝，缺少真实批准返回404', async () => {
    const recorded = recordedClient(server), swarm = createSwarmApi(recorded.client), key = crypto.randomUUID(), body = createRequest(key);
    await expect(swarm.createMission(body, crypto.randomUUID())).rejects.toThrow('请求身份必须与原命令标识相同');
    expect(recorded.requests).toHaveLength(0);
    await expect(swarm.createMission(body, key)).rejects.toMatchObject({ status: 404, code: 'not_found' });
    expect(recorded.requests.map(request => request.method)).toEqual(['GET', 'POST']);
    expect(recorded.requests[1].headers.get('Idempotency-Key')).toBe(key);
    expect(recorded.requests[1].headers.get('X-CSRF-Token')).toBe((await recorded.responses[0].json()).csrf_token);
    expect(await recorded.requests[1].json()).toEqual(body);
    expect(server.query('SELECT COUNT(*) AS count FROM swarm_missions')).toEqual([{ count: 0 }]);
  });

  it('实际缺少CSRF和无效Agent凭据返回403，产物版本输入实际校验失败', async () => {
    const recorded = recordedClient(server), swarm = createSwarmApi(recorded.client), key = crypto.randomUUID();
    await swarm.missions();
    const rejected = await recorded.client.POST('/missions', { body: createRequest(key), params: { header: { 'Idempotency-Key': key } } });
    expect(rejected.response.status).toBe(403);
    const denied = await recorded.client.GET('/missions', { headers: { Authorization: 'Bearer invalid' } });
    expect(denied.response.status).toBe(403);
    expect(denied.error?.error.code).toBe('scope_denied');
    await expect(swarm.artifact(crypto.randomUUID(), { mission_id: crypto.randomUUID(), version: 0 })).rejects.toMatchObject({ status: 400, code: 'validation_failed' });
    expect(server.query('SELECT COUNT(*) AS count FROM swarm_artifacts')).toEqual([{ count: 0 }]);
  });

  it('实际人工批准后保存任务、工作项和时间线，HTTP重启后保留固定版本', async () => {
    const human = await humanAPI(server.apiURL), recorded = recordedClient(server), proposals = createProposalsApi(recorded.client), swarm = createSwarmApi(recorded.client);
    const space = await human.write('/project-spaces', { title: '共享蜂群客户端实际核查' }, { status: 201 }) as S['ProjectSpaceResultV1'];
    const registered = await human.write('/local-projects', { root, name: 'standards-frontend', space_id: space.space.id }) as S['LocalProjectResultV1'];
    const { project } = await human.write(`/local-projects/${registered.project.id}/settings`, { expected_revision: registered.project.settings.revision, settings: { allow_directory: true, allowed_subdirs: ['web'], allowed_actions: ['create_mission', 'manage_work_items', 'read_swarm'], allowed_tools: [], external_model_cli: '', allow_agent_control: false } }, { method: 'PUT' }) as S['LocalProjectResultV1'];
    const head = execFileSync('git', ['rev-parse', 'HEAD'], { cwd: root, encoding: 'utf8' }).trim();
    const text = execFileSync('git', ['show', `${head}:AGENTS.md`], { cwd: root, encoding: 'utf8' });
    const imported = await importMaterial(human, { kind: 'text', adapter: 'manual', source_key: '', content_digest: '', source_locator: `https://github.com/songconmaisaix31-design/Astrocyte/blob/${head}/AGENTS.md`, export_text: text, title: 'AGENTS.md', collection_reason: null, source_spans: ['whole document'] });
    const material = imported.detail.material;
    await human.write(`/project-spaces/${space.space.id}/references`, { material_id: material.id, revision: material.current_revision });
    const goalKey = crypto.randomUUID();
    const { goal } = await proposals.createGoal({ schema_version: 1, request_id: goalKey, expected_version: 0, goal: { project_id: project.id, title: '核查共享客户端的真实任务账本', success_criteria: ['实际重启后保留任务和工作项'], priority: 1, status: 'active' } }, goalKey);
    const executionRoot = join(root, 'web');
    const content: ProposalContent = { title: '共享客户端固定协议核查', goal: { id: goal.id, revision: goal.revision }, inputs: [{ material_id: material.id, revision: material.current_revision }], scope: { project_id: project.id, settings_revision: project.settings.revision, roots: [executionRoot], write_paths: ['src/api'], actions: ['create_mission', 'manage_work_items', 'read_swarm'], tools: [], external_targets: [], max_concurrency: 1 }, budget: { calls: 3, tokens: null, money_minor: null, currency: '' }, stops: { deadline: new Date(Date.now() + 86400_000).toISOString(), conditions: ['完成本次账本核查后停止'], max_work_items: 2, max_attempts: 2, max_no_progress_rounds: 2, no_progress_seconds: 300 }, deliverables: ['真实HTTP和SQLite记录'], success_criteria: ['固定输入和持久记录保持一致'] };
    const proposalKey = crypto.randomUUID();
    const { proposal } = await proposals.createProposal({ schema_version: 1, request_id: proposalKey, expected_version: 0, content }, proposalKey);
    const submitKey = crypto.randomUUID();
    const submitted = await proposals.submitProposal(proposal.id, { schema_version: 1, request_id: submitKey, expected_version: proposal.version }, submitKey);
    const approveKey = crypto.randomUUID();
    const { approval } = await proposals.approveProposal(proposal.id, { schema_version: 1, request_id: approveKey, expected_version: submitted.proposal.version, expected_revision: submitted.proposal.revision, expires_at: new Date(Date.now() + 3600_000).toISOString() }, approveKey);
    const key = crypto.randomUUID();
    const { mission } = await swarm.createMission({ schema_version: 1, request_id: key, expected_version: 0, grant_id: approval.id, epoch: approval.epoch, proposal_revision: approval.proposal_revision, project_id: project.id, root: executionRoot }, key);
    expect(mission.grant.id).toBe(approval.id);
    expect(mission.state).toBe('ready');
    const itemKey = crypto.randomUUID();
    const { work_item: item } = await swarm.addWorkItem(mission.id, { schema_version: 1, request_id: itemKey, expected_version: mission.version, title: '核查实际固定协议引用', inputs: [{ id: material.id, revision: String(material.current_revision) }], write_paths: ['src/api'], dependencies: [], required_capabilities: [], artifact_inputs: [] }, itemKey);
    expect(item.state).toBe('open');
    expect(item.inputs).toEqual([{ id: material.id, revision: String(material.current_revision) }]);
    expect((await swarm.workItems(mission.id)).items).toEqual([item]);
    const timeline = await swarm.timeline(mission.id);
    expect(timeline.items.map(event => event.type)).toEqual(['swarm.mission_created', 'swarm.work_item_created']);
    for (const event of timeline.items) expect(server.query('SELECT event_id FROM foundation_outbox WHERE event_id=?', event.event_id)).toEqual([{ event_id: event.event_id }]);
    const detail = await swarm.mission(mission.id);
    expect(detail.work_items).toEqual([item]);
    expect(detail.leases).toEqual([]);
    expect(detail.executions).toEqual([]);
    await server.restart();
    expect(await swarm.mission(mission.id)).toEqual(detail);
    expect((await swarm.missions()).items).toEqual([detail.mission]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
  }, 60_000);

  it('实际取消在写入前失败，服务断开后传播连接错误', async () => {
    const recorded = recordedClient(server), swarm = createSwarmApi(recorded.client), key = crypto.randomUUID(), controller = new AbortController();
    controller.abort();
    await expect(swarm.createMission(createRequest(key), key, controller.signal)).rejects.toMatchObject({ name: 'AbortError' });
    expect(recorded.requests).toHaveLength(1);
    expect(recorded.requests[0].method).toBe('GET');
    expect(recorded.requests[0].signal.aborted).toBe(true);
    await server.close();
    await expect(swarm.missions()).rejects.toThrow();
    expect(recorded.requests.every(request => request.method === 'GET')).toBe(true);
  });
});
