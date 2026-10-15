import { expect, test, type Page } from '@playwright/test';
import { join } from 'node:path';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import { importTrackedDocument, registerCurrentProject } from './realService';
import type { components } from '../src/api/schema';

type S = components['schemas'];

async function contextService() {
  const server = await startS1Server({ browser: true, env: { ASTROCYTE_REPOSITORY_ENVIRONMENT_ID: 'product-ui-context-browser' } });
  return { server, api: await humanAPI(server.apiURL) };
}

async function approvedContext(api: Awaited<ReturnType<typeof humanAPI>>) {
  const project = await registerCurrentProject(api);
  const relativeRoot = 'web/src/pages/workspace/proposals';
  const actions = ['read_context'];
  const { project: configured } = await api.write(`/local-projects/${project.id}/settings`, { expected_revision: project.settings.revision, settings: { allow_directory: true, allowed_subdirs: [relativeRoot], allowed_actions: actions, allowed_tools: [], external_model_cli: '', allow_agent_control: false } }, { method: 'PUT' }) as S['LocalProjectResultV1'];
  const imported = await importTrackedDocument(api);
  const material = imported.detail.material;
  const { space } = await api.get(`/project-spaces/${project.space_id}`) as S['ProjectSpaceResultV1'];
  await api.write(`/project-spaces/${space.id}/references`, { expected_version: space.version, material_id: material.id, revision: material.current_revision });
  const { goal } = await api.write('/goals', { expected_version: 0, goal: { project_id: project.id, title: '核查固定上下文来源与实际快照', success_criteria: ['重启后固定来源和后继关系保持'], priority: 1, status: 'active' } }) as S['S2GoalResultV1'];
  const content: S['S2ProposalContentV1'] = {
    title: '真实上下文页面核查', goal: { id: goal.id, revision: goal.revision }, inputs: [{ material_id: material.id, revision: material.current_revision }],
    scope: { project_id: project.id, settings_revision: configured.settings.revision, roots: [join(project.root, relativeRoot)], write_paths: [], actions, tools: [], external_targets: [], max_concurrency: 1 },
    budget: { calls: null, tokens: null, money_minor: null, currency: '' }, deliverables: ['固定项目上下文记录'], success_criteria: ['真实来源、批准和快照可回查'],
    stops: { deadline: new Date(Date.now() + 86400_000).toISOString(), conditions: ['用户撤销后禁止建立新上下文'], max_work_items: 1, max_attempts: 1, no_progress_seconds: 3600, max_no_progress_rounds: 1 },
  };
  const { proposal } = await api.write('/proposals', { expected_version: 0, content }) as S['S2ProposalResultV1'];
  const { proposal: review } = await api.write(`/proposals/${proposal.id}/submit`, { expected_version: proposal.version }) as S['S2ProposalResultV1'];
  const { approval } = await api.write(`/proposals/${proposal.id}/approvals`, { expected_version: review.version, expected_revision: review.revision, expires_at: new Date(Date.now() + 3600_000).toISOString() }) as S['S2ApprovalResultV1'];
  return { project, material, goal, proposal, approval, relativeRoot };
}

async function openContext(page: Page, webURL: string, title: string) {
  await page.goto(`${webURL}/workspace`);
  await page.getByRole('tab', { name: '提案与批准', exact: true }).click();
  await page.getByRole('button', { name: `查看提案 ${title}`, exact: true }).click();
  return page.getByRole('region', { name: '批准范围内的固定上下文', exact: true });
}

test('真实固定上下文创建、后继重建、403恢复、撤销拒绝及重启持久化', async ({ page, context }) => {
  test.setTimeout(240_000);
  const { server, api } = await contextService();
  try {
    const scope = await approvedContext(api);
    let panel = await openContext(page, server.webURL, scope.proposal.content.title);
    await panel.getByLabel('上下文预算（字节）', { exact: true }).fill('131072');
    await panel.getByLabel('上下文未解决问题', { exact: true }).fill('本轨真实HTTP与页面持久化是否通过');
    const firstResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/context-packets') && response.request().method() === 'POST');
    await panel.getByRole('button', { name: '建立固定上下文包', exact: true }).focus();
    await page.keyboard.press('Enter');
    const first = await firstResponse;
    expect(first.status()).toBe(200);
    const { context_packet: packet } = await first.json() as S['S2ContextPacketResultV1'];
    expect(packet.goal).toEqual(scope.goal);
    expect(packet.sources).toEqual([expect.objectContaining({ reference: { material_id: scope.material.id, revision: scope.material.current_revision } })]);
    expect(packet.scope.roots).toEqual([scope.relativeRoot]);
    expect(packet.snapshot).not.toHaveProperty('root');
    expect(packet.snapshot.head).toMatch(/^[0-9a-f]{40}$/);
    expect(packet.snapshot.index_digest).not.toBe('');
    expect(packet.snapshot.worktree_digest).not.toBe('');
    let reference = panel.getByRole('region', { name: `固定上下文包 ${packet.id}`, exact: true });
    await reference.getByRole('button', { name: '展开固定上下文', exact: true }).click();
    await expect(reference).toContainText(packet.snapshot.worktree_id);
    await expect(reference).toContainText(packet.sources[0].object_ref);
    await expect(reference).toContainText('本轨真实HTTP与页面持久化是否通过');
    await context.setExtraHTTPHeaders({ Authorization: 'Bearer invalid' });
    const denied = page.waitForResponse(response => response.url().endsWith(`/context-packets/${packet.id}`));
    await reference.getByRole('button', { name: '重新读取固定上下文', exact: true }).click();
    expect((await denied).status()).toBe(403);
    await expect(reference.getByRole('alert')).toContainText('数据可能已过期');
    await expect(reference).toContainText(packet.snapshot.worktree_id);
    await context.setExtraHTTPHeaders({});
    await reference.getByRole('button', { name: '重新读取固定上下文', exact: true }).click();
    await expect(reference.getByRole('alert')).toHaveCount(0);
    await panel.getByLabel('前序上下文包 ID', { exact: true }).fill(packet.id);
    await panel.getByLabel('上下文关键决定', { exact: true }).fill('保留原固定包，建立后继记录');
    const nextResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/context-packets') && response.request().method() === 'POST');
    await panel.getByRole('button', { name: '建立后继上下文包', exact: true }).click();
    const next = await nextResponse;
    expect(next.status()).toBe(200);
    const { context_packet: successor } = await next.json() as S['S2ContextPacketResultV1'];
    expect(successor.id).not.toBe(packet.id);
    expect(successor.predecessor_id).toBe(packet.id);
    expect((await api.get(`/context-packets/${packet.id}`) as S['S2ContextPacketResultV1']).context_packet).toEqual(packet);
    reference = panel.getByRole('region', { name: `固定上下文包 ${successor.id}`, exact: true });
    await reference.getByRole('button', { name: '展开固定上下文', exact: true }).click();
    await reference.getByRole('region', { name: `固定上下文包 ${packet.id}`, exact: true }).getByRole('button', { name: '展开固定上下文', exact: true }).click();
    await expect(reference.getByRole('region', { name: `固定上下文包 ${packet.id}`, exact: true })).toContainText(packet.snapshot.head);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_context_packets')).toEqual([{ count: 2 }]);
    const events = server.query('SELECT event_id FROM foundation_outbox WHERE event_type=?', 'context_packet_built');
    expect(events).toHaveLength(2);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
    await server.restart();
    const restarted = await humanAPI(server.apiURL);
    expect((await restarted.get(`/context-packets/${successor.id}`) as S['S2ContextPacketResultV1']).context_packet).toEqual(successor);
    await page.reload();
    panel = await openContext(page, server.webURL, scope.proposal.content.title);
    await panel.getByLabel('读取固定上下文包 ID', { exact: true }).fill(successor.id);
    await panel.getByRole('button', { name: '读取指定上下文包', exact: true }).click();
    await panel.getByRole('region', { name: `固定上下文包 ${successor.id}`, exact: true }).getByRole('button', { name: '展开固定上下文', exact: true }).click();
    await expect(panel).toContainText(successor.snapshot.worktree_digest);
    await restarted.write(`/approvals/${scope.approval.id}/revoke`, { expected_version: scope.approval.epoch, reason: '核查新上下文建立拒绝' });
    await page.getByRole('button', { name: '重新读取批准记录', exact: true }).click();
    await expect(panel.getByRole('button', { name: '建立固定上下文包', exact: true })).toBeDisabled();
    const key = crypto.randomUUID();
    const refused = await restarted.request('/context-packets', { method: 'POST', key, body: { schema_version: 1, request_id: key, expected_version: 0, grant_id: scope.approval.id, grant_epoch: scope.approval.epoch + 1, max_bytes: 131072 } });
    expect(refused.status).toBe(403);
    expect((refused.data as S['ErrorV1']).error.code).toBe('approval_revoked');
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_context_packets')).toEqual([{ count: 2 }]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  } finally { await context.setExtraHTTPHeaders({}); await server.close({ preserveData: true }); }
});

test('真实离线建立上下文保留输入与unknown锁，关闭重开和刷新不重发', async ({ page, context }) => {
  test.setTimeout(180_000);
  const { server, api } = await contextService();
  try {
    const scope = await approvedContext(api);
    let panel = await openContext(page, server.webURL, scope.proposal.content.title);
    await panel.getByLabel('上下文预算（字节）', { exact: true }).fill('131072');
    await panel.getByLabel('上下文失败记录', { exact: true }).fill('离线命令结果待核查');
    const writes: string[] = [];
    page.on('request', request => { if (request.method() === 'POST' && request.url().endsWith('/api/v1/context-packets')) writes.push(request.url()); });
    await context.setOffline(true);
    await panel.getByRole('button', { name: '建立固定上下文包', exact: true }).click();
    await expect(panel.getByRole('alert').filter({ hasText: '上下文建立请求' })).toContainText('结果未知');
    await expect(panel.getByLabel('上下文失败记录', { exact: true })).toHaveValue('离线命令结果待核查');
    await context.setOffline(false);
    await page.keyboard.press('Escape');
    await page.getByRole('button', { name: `查看提案 ${scope.proposal.content.title}`, exact: true }).click();
    panel = page.getByRole('region', { name: '批准范围内的固定上下文', exact: true });
    await expect(panel.getByRole('button', { name: '建立固定上下文包', exact: true })).toBeDisabled();
    await page.reload();
    panel = await openContext(page, server.webURL, scope.proposal.content.title);
    await expect(panel.getByRole('alert').filter({ hasText: '上下文建立请求' })).toContainText('结果未知');
    expect(writes).toHaveLength(0);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_context_packets')).toEqual([{ count: 0 }]);
  } finally { await context.setOffline(false); await server.close({ preserveData: true }); }
});
