import { expect, test, type Locator, type Page } from '@playwright/test';
import { join } from 'node:path';
import { humanAPI } from '../../tests/s1/api.mjs';
import { importTrackedDocument, realService, registerCurrentProject, writeWorks } from './realService';
import type { components } from '../src/api/schema';

type S = components['schemas'];
type API = Awaited<ReturnType<typeof realService>>['api'];

async function approvedScope(api: API) {
  const project = await registerCurrentProject(api);
  const actions = ['create_mission', 'manage_work_items', 'read_swarm', 'recover_mission', 'accept_artifact'];
  const settings = await api.write(`/local-projects/${project.id}/settings`, { expected_revision: project.settings.revision, settings: { allow_directory: true, allowed_subdirs: ['web'], allowed_actions: actions, allowed_tools: [], external_model_cli: '', allow_agent_control: false } }, { method: 'PUT' }) as S['LocalProjectResultV1'];
  const imported = await importTrackedDocument(api);
  const material = imported.detail.material;
  const { space } = await api.get(`/project-spaces/${project.space_id}`) as S['ProjectSpaceResultV1'];
  await api.write(`/project-spaces/${space.id}/references`, { expected_version: space.version, material_id: material.id, revision: material.current_revision });
  const { goal } = await api.write('/goals', { expected_version: 0, goal: { project_id: project.id, title: '验证真实 Swarm 人工任务账本', success_criteria: ['保存真实任务与事件并在进程重启后保持'], priority: 1, status: 'active' } }) as S['S2GoalResultV1'];
  const root = join(project.root, 'web');
  const content: S['S2ProposalContentV1'] = {
    title: '核查人工 Swarm 生命周期', goal: { id: goal.id, revision: goal.revision }, inputs: [{ material_id: material.id, revision: material.current_revision }],
    scope: { project_id: project.id, settings_revision: settings.project.settings.revision, roots: [root], write_paths: ['src/pages/swarm/missions'], actions, tools: [], external_targets: [], max_concurrency: 1 },
    budget: { calls: 3, tokens: null, money_minor: null, currency: '' }, deliverables: ['真实任务、工作项和事件记录'], success_criteria: ['任务停止、权限和重启行为可回查'],
    stops: { deadline: new Date(Date.now() + 86400_000).toISOString(), conditions: ['用户停止或撤销时禁止新受保护动作'], max_work_items: 4, max_attempts: 2, no_progress_seconds: 3600, max_no_progress_rounds: 3 },
  };
  const { proposal } = await api.write('/proposals', { expected_version: 0, content }) as S['S2ProposalResultV1'];
  const { proposal: review } = await api.write(`/proposals/${proposal.id}/submit`, { expected_version: proposal.version }) as S['S2ProposalResultV1'];
  const { approval } = await api.write(`/proposals/${proposal.id}/approvals`, { expected_version: review.version, expected_revision: review.revision, expires_at: new Date(Date.now() + 3600_000).toISOString() }) as S['S2ApprovalResultV1'];
  return { project, goal, proposal, approval, root };
}

async function createMission(api: API, scope: Awaited<ReturnType<typeof approvedScope>>) {
  const { mission } = await writeWorks(api, '/missions', { expected_version: 0, grant_id: scope.approval.id, epoch: scope.approval.epoch, proposal_revision: scope.approval.proposal_revision, project_id: scope.project.id, root: scope.root }) as S['SwarmMissionEnvelopeV1'];
  expect(mission.state).toBe('ready');
  return mission;
}

async function openMission(page: Page, webURL: string, mission: S['SwarmMissionV1']) {
  await page.goto(`${webURL}/swarm`);
  const button = page.getByRole('button', { name: `查看任务 ${mission.grant.goal}`, exact: true });
  await button.focus();
  await button.press('Enter');
  const dialog = page.getByRole('dialog', { name: '任务详情', exact: true });
  await expect(dialog.getByRole('button', { name: '暂停任务', exact: true })).toBeVisible();
  return dialog;
}

async function fillWorkItem(dialog: Locator, mission: S['SwarmMissionV1'], title: string) {
  await dialog.locator('summary').filter({ hasText: /^添加工作项$/ }).click();
  const editor = dialog.getByRole('region', { name: '添加任务工作项', exact: true });
  await editor.getByLabel('工作项标题', { exact: true }).fill(title);
  await editor.getByRole('button', { name: '添加固定输入版本', exact: true }).click();
  const input = mission.grant.inputs?.[0];
  expect(input).toBeDefined();
  if (!input) throw new Error('实际批准没有固定输入');
  await editor.getByLabel('输入 1 ID', { exact: true }).fill(input.id);
  await editor.getByLabel('输入 1 版本', { exact: true }).fill(input.revision);
  await editor.getByLabel('工作项允许修改的路径', { exact: true }).fill('src/pages/swarm/missions');
  return editor;
}

test('真实人工建立任务、添加工作项、时间线、暂停恢复取消与SQLite重启保持', async ({ page }) => {
  test.setTimeout(240_000);
  const { server, api } = await realService();
  try {
    const scope = await approvedScope(api);
    await page.goto(`${server.webURL}/swarm`);
    await page.getByRole('button', { name: '从批准建立任务', exact: true }).click();
    let dialog = page.getByRole('dialog', { name: '从批准建立任务', exact: true });
    await dialog.getByLabel('已批准提案', { exact: true }).selectOption(scope.approval.id);
    await dialog.getByLabel('任务根目录', { exact: true }).selectOption(scope.root);
    const createdResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/missions') && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '建立任务账本', exact: true }).click();
    const created = await createdResponse;
    expect(created.status()).toBe(200);
    const { mission } = await created.json() as S['SwarmMissionEnvelopeV1'];
    expect(mission.state).toBe('ready');
    expect(mission.grant.id).toBe(scope.approval.id);
    dialog = page.getByRole('dialog', { name: '任务详情', exact: true });
    await expect(dialog.getByText('未知', { exact: true }).first()).toBeVisible();
    const editor = await fillWorkItem(dialog, mission, '核查本轨任务记录');
    const addedResponse = page.waitForResponse(response => response.url().endsWith(`/missions/${mission.id}/work-items`) && response.request().method() === 'POST');
    await editor.getByRole('button', { name: '保存工作项', exact: true }).click();
    const added = await addedResponse;
    expect(added.status()).toBe(200);
    const { work_item: work } = await added.json() as S['SwarmWorkItemEnvelopeV1'];
    expect(work.state).toBe('open');
    await dialog.getByRole('button', { name: `查看工作项 ${work.title}`, exact: true }).click();
    const workDetail = dialog.getByRole('region', { name: '工作项详情', exact: true });
    await expect(workDetail).toContainText(work.id);
    await expect(workDetail).toContainText('未认领');
    await expect(dialog.getByRole('button', { name: '认领工作项', exact: true })).toBeDisabled();
    await expect(dialog.getByRole('button', { name: '提交产物', exact: true })).toBeDisabled();
    await dialog.getByLabel('暂停或取消原因', { exact: true }).fill('核查人工暂停');
    await dialog.getByRole('button', { name: '暂停任务', exact: true }).click();
    await expect(dialog.getByRole('status', { name: '状态: 已暂停', exact: true })).toBeVisible();
    await dialog.getByLabel('任务恢复或核查原因', { exact: true }).fill('原任务没有原生执行，核查后恢复');
    await dialog.getByRole('button', { name: '核查原执行并恢复任务', exact: true }).click();
    await expect(dialog.getByRole('status', { name: '状态: 待执行', exact: true })).toBeVisible();
    await dialog.getByLabel('暂停或取消原因', { exact: true }).fill('本轨验收完成，取消后续任务');
    await dialog.getByRole('button', { name: '取消任务', exact: true }).click();
    await expect(dialog.getByRole('region', { name: '任务账本与批准快照', exact: true }).getByRole('status', { name: '状态: 已取消', exact: true })).toBeVisible();
    const detail = await api.get(`/missions/${mission.id}`) as S['SwarmMissionDetailV1'];
    expect(detail.mission.state).toBe('cancelled');
    expect(detail.work_items).toEqual([expect.objectContaining({ id: work.id, state: 'cancelled' })]);
    expect(detail.leases).toEqual([]);
    expect(detail.artifacts).toEqual([]);
    expect(detail.executions).toEqual([]);
    expect(detail.timeline?.map(event => event.type)).toEqual(expect.arrayContaining(['swarm.mission_created', 'swarm.work_item_created', 'swarm.mission_paused', 'swarm.mission_recovered', 'swarm.mission_cancelled']));
    for (const event of detail.timeline ?? []) {
      expect(server.query('SELECT event_type,aggregate_id,aggregate_version FROM foundation_outbox WHERE event_id=?', event.event_id)).toEqual([{ event_type: event.type, aggregate_id: event.aggregate_id, aggregate_version: event.aggregate_version }]);
      await expect(dialog.getByRole('region', { name: '真实事件时间线', exact: true }).getByText(event.event_id, { exact: true })).toBeVisible();
    }
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions WHERE project_id=?', scope.project.id)).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM swarm_leases')).toEqual([{ count: 0 }]);
    await server.restart();
    const restarted = await humanAPI(server.apiURL);
    expect(await restarted.get(`/missions/${mission.id}`)).toEqual(detail);
    await page.reload();
    await page.getByRole('button', { name: `查看任务 ${mission.grant.goal}`, exact: true }).click();
    dialog = page.getByRole('dialog', { name: '任务详情', exact: true });
    await expect(dialog.getByRole('button', { name: '取消任务', exact: true })).toBeDisabled();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
    await page.keyboard.press('Escape');
    await expect(page.getByRole('button', { name: `查看任务 ${mission.grant.goal}`, exact: true })).toBeFocused();
  } finally { await server.close({ preserveData: true }); }
});

test('真实工作项版本冲突保留输入，人工核对后保存，Grant撤销后新动作拒绝', async ({ page }) => {
  test.setTimeout(240_000);
  const { server, api } = await realService();
  try {
    const scope = await approvedScope(api);
    const mission = await createMission(api, scope);
    const dialog = await openMission(page, server.webURL, mission);
    const editor = await fillWorkItem(dialog, mission, '版本冲突时保留的工作项');
    await writeWorks(api, `/missions/${mission.id}/work-items`, { expected_version: mission.version, title: '同一任务实际并发新增项', inputs: mission.grant.inputs, dependencies: [], write_paths: ['src/pages/swarm/missions'] });
    const failedResponse = page.waitForResponse(response => response.url().endsWith(`/missions/${mission.id}/work-items`) && response.request().method() === 'POST');
    await editor.getByRole('button', { name: '保存工作项', exact: true }).click();
    const conflict = await failedResponse;
    expect(conflict.status()).toBe(409);
    expect((await conflict.json() as S['ErrorV1']).error.code).toBe('version_conflict');
    await expect(editor.getByLabel('工作项标题', { exact: true })).toHaveValue('版本冲突时保留的工作项');
    await dialog.getByRole('button', { name: '刷新当前任务', exact: true }).click();
    await editor.getByRole('button', { name: '已核对当前任务 v2', exact: true }).click();
    const savedResponse = page.waitForResponse(response => response.url().endsWith(`/missions/${mission.id}/work-items`) && response.request().method() === 'POST');
    await editor.getByRole('button', { name: '保存工作项', exact: true }).click();
    expect((await savedResponse).status()).toBe(200);
    await api.write(`/approvals/${scope.approval.id}/revoke`, { expected_version: scope.approval.epoch, reason: '核查撤销后的新动作拒绝' });
    await dialog.getByRole('button', { name: '刷新当前任务', exact: true }).click();
    await expect(dialog.getByText('当前 Grant 有效状态', { exact: true }).locator('..')).toContainText('revoked');
    await expect(editor.getByRole('button', { name: '保存工作项', exact: true })).toBeDisabled();
    const detail = await api.get(`/missions/${mission.id}`) as S['SwarmMissionDetailV1'];
    const key = crypto.randomUUID();
    const refused = await api.request(`/missions/${mission.id}/work-items`, { method: 'POST', key, body: { schema_version: 1, request_id: key, expected_version: detail.mission.version, title: '实际撤销拒绝检查', inputs: mission.grant.inputs, write_paths: ['src/pages/swarm/missions'], dependencies: [] } });
    expect(refused.status).toBe(403);
    expect((refused.data as S['ErrorV1']).error.code).toBe('approval_revoked');
    expect(server.query('SELECT COUNT(*) AS count FROM swarm_work_items WHERE mission_id=?', mission.id)).toEqual([{ count: 2 }]);
    await server.restart();
    const restarted = await humanAPI(server.apiURL);
    expect((await restarted.get(`/approvals/${scope.approval.id}`) as S['S2ApprovalResultV1']).approval.status).toBe('revoked');
    expect((await restarted.get(`/missions/${mission.id}`) as S['SwarmMissionDetailV1']).work_items).toHaveLength(2);
  } finally { await server.close({ preserveData: true }); }
});

test('真实离线任务命令保留unknown锁，关闭重开和刷新不自动重发', async ({ page, context }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService();
  try {
    const scope = await approvedScope(api);
    const mission = await createMission(api, scope);
    let dialog = await openMission(page, server.webURL, mission);
    await dialog.getByLabel('暂停或取消原因', { exact: true }).fill('核查离线未知结果');
    const writes: string[] = [];
    page.on('request', request => { if (request.method() === 'POST' && request.url().endsWith(`/missions/${mission.id}/pause`)) writes.push(request.url()); });
    await context.setOffline(true);
    await dialog.getByRole('button', { name: '暂停任务', exact: true }).click();
    await expect(dialog.getByRole('alert').filter({ hasText: '任务命令' })).toContainText('结果未知');
    await expect(dialog.getByRole('button', { name: '暂停任务', exact: true })).toBeDisabled();
    await context.setOffline(false);
    await page.keyboard.press('Escape');
    await page.getByRole('button', { name: `查看任务 ${mission.grant.goal}`, exact: true }).click();
    dialog = page.getByRole('dialog', { name: '任务详情', exact: true });
    await expect(dialog.getByRole('alert').filter({ hasText: '任务命令' })).toContainText('结果未知');
    await page.reload();
    await page.getByRole('button', { name: `查看任务 ${mission.grant.goal}`, exact: true }).click();
    dialog = page.getByRole('dialog', { name: '任务详情', exact: true });
    await expect(dialog.getByRole('button', { name: '取消任务', exact: true })).toBeDisabled();
    expect(writes).toHaveLength(0);
    expect((await api.get(`/missions/${mission.id}`) as S['SwarmMissionDetailV1']).mission).toEqual(mission);
    expect(server.query('SELECT COUNT(*) AS count FROM foundation_outbox WHERE event_type=?', 'swarm.mission_paused')).toEqual([{ count: 0 }]);
  } finally { await context.setOffline(false); await server.close({ preserveData: true }); }
});

test('真实任务读取403保留快照并禁用命令，恢复人工身份后重新读取', async ({ page, context }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService();
  try {
    const scope = await approvedScope(api);
    const mission = await createMission(api, scope);
    const dialog = await openMission(page, server.webURL, mission);
    await dialog.getByLabel('暂停或取消原因', { exact: true }).fill('读取权限恢复检查');
    await expect(dialog.getByRole('button', { name: '暂停任务', exact: true })).toBeEnabled();
    await context.setExtraHTTPHeaders({ Authorization: 'Bearer invalid' });
    const refused = page.waitForResponse(response => response.url().endsWith(`/missions/${mission.id}`) && response.request().method() === 'GET');
    await dialog.getByRole('button', { name: '刷新当前任务', exact: true }).click();
    expect((await refused).status()).toBe(403);
    await expect(dialog.getByRole('alert').filter({ hasText: '数据可能已过期' }).first()).toBeVisible();
    await expect(dialog.getByRole('region', { name: '任务账本与批准快照', exact: true })).toContainText(mission.id);
    await expect(dialog.getByRole('button', { name: '暂停任务', exact: true })).toBeDisabled();
    await context.setExtraHTTPHeaders({});
    await dialog.getByRole('button', { name: '刷新当前任务', exact: true }).click();
    await expect(dialog.getByRole('alert').filter({ hasText: '数据可能已过期' })).toHaveCount(0);
    await expect(dialog.getByRole('button', { name: '暂停任务', exact: true })).toBeEnabled();
    await expect(dialog.getByLabel('暂停或取消原因', { exact: true })).toHaveValue('读取权限恢复检查');
    expect((await api.get(`/missions/${mission.id}`) as S['SwarmMissionDetailV1']).mission).toEqual(mission);
  } finally { await context.setExtraHTTPHeaders({}); await server.close({ preserveData: true }); }
});
