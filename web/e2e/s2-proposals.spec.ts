import { test, expect, type Locator, type Page } from '@playwright/test';
import { importTrackedDocument, realService, registerCurrentProject } from './realService';
import { humanAPI } from '../../tests/s1/api.mjs';
import { join } from 'node:path';
import { readdir } from 'node:fs/promises';
import type { components } from '../src/api/schema';

function localMinute(date: Date) {
  return new Date(date.getTime() - date.getTimezoneOffset() * 60_000).toISOString().slice(0, 16);
}

async function fillProposal(form: Locator, goalId: string, projectId: string, root: string, title: string) {
  await form.getByLabel('提案标题', { exact: true }).fill(title);
  await form.getByLabel('目标 ID', { exact: true }).fill(goalId);
  await form.getByLabel('目标版本', { exact: true }).fill('1');
  await form.getByLabel('项目 ID', { exact: true }).fill(projectId);
  await form.getByLabel('允许根目录', { exact: true }).fill(root);
  await form.getByLabel('允许修改的路径', { exact: true }).fill('src/pages/workspace/proposals');
  await form.getByLabel('允许动作', { exact: true }).fill('create_mission\nmanage_work_items\nread_swarm');
  await form.getByLabel('最大并发数', { exact: true }).fill('1');
  await form.getByLabel('调用上限', { exact: true }).fill('3');
  await form.getByLabel('预期产出', { exact: true }).fill('本轨实际 HTTP 与 SQLite 核查记录');
  await form.getByLabel('成功条件', { exact: true }).fill('固定提案版本和授权记录能够在重启后查询');
  await form.getByLabel('截止时间', { exact: true }).fill(localMinute(new Date(Date.now() + 86400_000)));
  await form.getByLabel('停止条件', { exact: true }).fill('达到调用上限或用户撤销批准时停止');
  await form.getByLabel('最大工作项数', { exact: true }).fill('2');
  await form.getByLabel('最大尝试次数', { exact: true }).fill('2');
  await form.getByLabel('连续无进展轮数上限', { exact: true }).fill('2');
  await form.getByLabel('无进展时限（秒）', { exact: true }).fill('300');
}

async function expectGrantSnapshot(panel: Locator, grant: components['schemas']['S2ApprovalGrantV1']) {
  const fields: [string, string][] = [
    ['批准提案版本', `r${grant.proposal_revision}`], ['授权 Epoch', String(grant.epoch)],
    ['批准者', grant.approver_id], ['批准时间', grant.approved_at], ['有效期', grant.expires_at], ['状态变更时间', grant.changed_at],
    ['批准的固定目标', `${grant.goal.id} · r${grant.goal.revision}`], ['授权项目', grant.scope.project_id],
    ['批准时的项目权限版本', `r${grant.scope.settings_revision}`],
    ['允许根目录', grant.scope.roots?.join('；') || '无'], ['允许修改的路径', grant.scope.write_paths?.join('；') || '无'],
    ['允许动作', grant.scope.actions?.join('；') || '无'], ['允许工具', grant.scope.tools?.join('；') || '无'],
    ['允许外发目标', grant.scope.external_targets?.join('；') || '无'], ['最大并发数', String(grant.scope.max_concurrency)],
    ['调用上限', String(grant.budget.calls ?? '未知')], ['Token 上限', String(grant.budget.tokens ?? '未知')],
    ['费用上限', grant.budget.money_minor === null ? '未知' : `${grant.budget.money_minor} ${grant.budget.currency}（最小货币单位）`],
    ['执行截止时间', grant.stops.deadline], ['最大工作项数', String(grant.stops.max_work_items)],
    ['最大尝试次数', String(grant.stops.max_attempts)], ['连续无进展轮数上限', String(grant.stops.max_no_progress_rounds)],
    ['无进展时限（秒）', String(grant.stops.no_progress_seconds)], ['Grant ID', grant.id],
  ];
  for (const [label, value] of fields) await expect(panel.getByText(label, { exact: true }).locator('..')).toContainText(value);
  for (const condition of grant.stops.conditions ?? []) await expect(panel.getByRole('listitem').filter({ hasText: condition })).toHaveText(condition);
}

test('真实提案创建、不可变版本、人工批准与撤销通过HTTP及SQLite重启核查', async ({ page }) => {
  test.setTimeout(240_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    await api.write(`/local-projects/${project.id}/settings`, {
      expected_revision: project.settings.revision,
      settings: { allow_directory: true, allowed_subdirs: ['web'], allowed_actions: ['create_mission', 'manage_work_items', 'read_swarm'], allowed_tools: [], external_model_cli: '', allow_agent_control: false },
    }, { method: 'PUT' });
    await page.goto(`${server.webURL}/workspace`);
    await page.getByRole('tab', { name: '提案与批准', exact: true }).click();
    await page.getByRole('button', { name: '创建目标', exact: true }).click();
    const goalDialog = page.getByRole('dialog', { name: '创建目标', exact: true });
    await goalDialog.getByLabel('目标所属项目', { exact: true }).selectOption(project.id);
    await goalDialog.getByLabel('目标名称', { exact: true }).fill('核查本轨真实提案与批准记录');
    await goalDialog.getByLabel('目标成功条件', { exact: true }).fill('重启后保留提案历史和撤销状态');
    await goalDialog.getByLabel('目标优先级', { exact: true }).fill('1');
    const goalResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/goals') && response.request().method() === 'POST');
    await goalDialog.getByRole('button', { name: '保存目标', exact: true }).click();
    const goalSaved = await goalResponse;
    expect(goalSaved.status()).toBe(200);
    const { goal } = await goalSaved.json() as components['schemas']['S2GoalResultV1'];
    await page.getByRole('button', { name: '创建提案', exact: true }).click();
    let dialog = page.getByRole('dialog', { name: '创建提案', exact: true });
    await fillProposal(dialog, goal.id, project.id, join(project.root, 'web'), '本轨真实记录核查 r1');
    const createdResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/proposals') && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '保存提案草稿', exact: true }).click();
    const created = await createdResponse;
    expect(created.status()).toBe(200);
    const { proposal: original } = await created.json() as components['schemas']['S2ProposalResultV1'];
    expect(original.revision).toBe(1);
    expect(original.status).toBe('draft');
    expect(original.content.budget.money_minor).toBeNull();
    expect((await api.get('/missions') as components['schemas']['MissionListV1']).items).toHaveLength(0);
    const originalButton = page.getByRole('button', { name: '查看提案 本轨真实记录核查 r1', exact: true });
    await originalButton.focus();
    await originalButton.press('Enter');
    dialog = page.getByRole('dialog', { name: '提案详情', exact: true });
    await expect(dialog.getByText('未知', { exact: true }).first()).toBeVisible();
    await dialog.getByRole('button', { name: '编辑新版本', exact: true }).click();
    await dialog.getByLabel('提案标题', { exact: true }).fill('本轨真实记录核查 r2');
    await dialog.getByRole('button', { name: '保存新版本', exact: true }).click();
    await expect(dialog.getByText('r2', { exact: true }).first()).toBeVisible();
    const fixedFirst = (await api.get(`/proposals/${original.id}?revision=1`) as components['schemas']['S2ProposalResultV1']).proposal;
    expect(fixedFirst.content).toEqual(original.content);
    await expect(dialog.getByRole('region', { name: '提案版本差分' })).toContainText('本轨真实记录核查 r1');
    await dialog.getByRole('button', { name: '提交审阅', exact: true }).click();
    await expect(dialog.getByRole('status', { name: '状态: 审核中', exact: true })).toBeVisible();
    await dialog.getByLabel('批准有效期', { exact: true }).fill(localMinute(new Date(Date.now() + 3600_000)));
    await dialog.getByRole('checkbox', { name: '我已核对当前固定版本、范围、资源上限和停止条件' }).check();
    const approvedResponse = page.waitForResponse(response => response.url().endsWith(`/proposals/${original.id}/approvals`));
    await dialog.getByRole('button', { name: '批准当前版本', exact: true }).click();
    const approved = await approvedResponse;
    expect(approved.status()).toBe(200);
    const { approval } = await approved.json() as components['schemas']['S2ApprovalResultV1'];
    expect(approval.proposal_revision).toBe(2);
    expect(approval.status).toBe('active');
    expect(approval.scope).toEqual(original.content.scope);
    expect(approval.budget).toEqual(original.content.budget);
    expect(approval.stops).toEqual(original.content.stops);
    const approvalPanel = dialog.getByRole('region', { name: '审批授权', exact: true });
    await expect(approvalPanel.getByRole('status', { name: '状态: 有效', exact: true })).toBeVisible();
    await expectGrantSnapshot(approvalPanel, approval);
    const missions = (await api.get('/missions') as components['schemas']['MissionListV1']).items;
    expect(missions.length).toBeLessThanOrEqual(1);
    for (const mission of missions) expect(mission).toMatchObject({ state: 'ready', grant: { id: approval.id } });
    expect(server.query('SELECT COUNT(*) AS count FROM foundation_outbox WHERE event_type=? AND aggregate_id=?', 'workspace.proposal_approved', original.id)).toEqual([{ count: 1 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions WHERE project_id=?', project.id)).toEqual([{ count: 0 }]);
    expect((await readdir(server.dataDir)).includes('native-sessions')).toBe(false);
    await dialog.getByLabel('撤销原因', { exact: true }).fill('本轨核查完成，停止后续新动作');
    await dialog.getByRole('button', { name: '撤销批准', exact: true }).click();
    await expect(dialog.getByRole('status', { name: '状态: 已撤销', exact: true })).toBeVisible();
    expect(server.query('SELECT epoch, json_extract(data, \'$.status\') AS status FROM workspace_approvals WHERE id=?', approval.id)).toEqual([{ epoch: 2, status: 'revoked' }]);
    expect(server.query('SELECT revision, json_extract(data, \'$.content.title\') AS title FROM workspace_proposal_revisions WHERE id=? ORDER BY revision', original.id)).toEqual([{ revision: 1, title: original.content.title }, { revision: 2, title: '本轨真实记录核查 r2' }]);
    await server.restart();
    const restarted = await humanAPI(server.apiURL);
    const restoredGrant = (await restarted.get(`/approvals/${approval.id}`) as components['schemas']['S2ApprovalResultV1']).approval;
    expect(restoredGrant).toMatchObject({ status: 'revoked', epoch: 2, scope: approval.scope, budget: approval.budget, stops: approval.stops });
    expect((await restarted.get(`/proposals/${original.id}?revision=1`) as components['schemas']['S2ProposalResultV1']).proposal.content).toEqual(original.content);
    await page.reload();
    await page.getByRole('button', { name: '查看提案 本轨真实记录核查 r2', exact: true }).click();
    await expect(page.getByRole('dialog').getByRole('status', { name: '状态: 已撤销', exact: true })).toBeVisible();
    await expectGrantSnapshot(page.getByRole('region', { name: '审批授权', exact: true }), restoredGrant);
    await expect(page.getByRole('button', { name: '撤销批准', exact: true })).toBeDisabled();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
    await page.getByRole('dialog').press('Escape');
    await expect(page.getByRole('dialog')).not.toBeVisible();
    await expect(page.getByRole('button', { name: '查看提案 本轨真实记录核查 r2', exact: true })).toBeFocused();
  } finally { await server.close({ preserveData: true }); }
});

async function prepareDraft(api: Awaited<ReturnType<typeof realService>>['api'], permitted = true) {
  const project = await registerCurrentProject(api);
  const settings = await api.write(`/local-projects/${project.id}/settings`, {
    expected_revision: project.settings.revision,
    settings: { allow_directory: true, allowed_subdirs: ['web'], allowed_actions: permitted ? ['create_mission', 'manage_work_items', 'read_swarm'] : [], allowed_tools: [], external_model_cli: '', allow_agent_control: false },
  }, { method: 'PUT' }) as components['schemas']['LocalProjectResultV1'];
  const { goal } = await api.write('/goals', { expected_version: 0, goal: { project_id: project.id, title: '核查本轨实际冲突和权限', success_criteria: ['错误时保留输入、固定版本及原请求身份'], priority: 1, status: 'active' } }) as components['schemas']['S2GoalResultV1'];
  const content: components['schemas']['S2ProposalContentV1'] = {
    title: '本轨真实错误状态核查', goal: { id: goal.id, revision: goal.revision }, inputs: [],
    scope: { project_id: project.id, settings_revision: settings.project.settings.revision, roots: [join(project.root, 'web')], write_paths: ['src/pages/workspace/proposals'], actions: ['create_mission', 'manage_work_items', 'read_swarm'], tools: [], external_targets: [], max_concurrency: 1 },
    budget: { calls: 3, tokens: null, money_minor: null, currency: '' },
    stops: { deadline: new Date(Date.now() + 86400_000).toISOString(), conditions: ['完成本轨核查后停止'], max_work_items: 2, max_attempts: 2, no_progress_seconds: 300, max_no_progress_rounds: 2 },
    deliverables: ['实际错误与恢复记录'], success_criteria: ['不丢失输入，不自动重发未知操作'],
  };
  const { proposal } = await api.write('/proposals', { expected_version: 0, content }) as components['schemas']['S2ProposalResultV1'];
  return { project, proposal };
}

async function openDraft(page: Page, webURL: string, proposal: components['schemas']['ProposalV1']) {
  await page.goto(`${webURL}/workspace`);
  await page.getByRole('tab', { name: '提案与批准', exact: true }).click();
  await page.getByRole('button', { name: `查看提案 ${proposal.content.title}`, exact: true }).click();
  return page.getByRole('dialog', { name: '提案详情', exact: true });
}

test('真实version_conflict保留编辑内容，人工核对最新记录后保存新版本', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService();
  try {
    const { proposal } = await prepareDraft(api);
    const dialog = await openDraft(page, server.webURL, proposal);
    await dialog.getByRole('button', { name: '编辑新版本', exact: true }).click();
    await dialog.getByLabel('提案标题', { exact: true }).fill('本轨保留的人工编辑');
    const external = await api.write(`/proposals/${proposal.id}`, { expected_version: proposal.version, content: { ...proposal.content, title: '本轨另一实际人类修改' } }, { method: 'PUT' }) as components['schemas']['S2ProposalResultV1'];
    const conflictResponse = page.waitForResponse(response => response.url().endsWith(`/proposals/${proposal.id}`) && response.request().method() === 'PUT');
    await dialog.getByRole('button', { name: '保存新版本', exact: true }).click();
    const conflict = await conflictResponse;
    expect(conflict.status()).toBe(409);
    expect((await conflict.json() as components['schemas']['ErrorV1']).error.code).toBe('version_conflict');
    await expect(dialog.getByRole('alert')).toContainText('记录已更新');
    await expect(dialog.getByLabel('提案标题', { exact: true })).toHaveValue('本轨保留的人工编辑');
    await dialog.getByRole('button', { name: '刷新当前提案', exact: true }).click();
    await expect(dialog.getByLabel('提案标题', { exact: true })).toHaveValue('本轨保留的人工编辑');
    await dialog.locator('summary').filter({ hasText: '核对最新提案' }).click();
    await dialog.getByRole('button', { name: '使用已核对的最新记录版本', exact: true }).click();
    await dialog.getByRole('button', { name: '保存新版本', exact: true }).click();
    await expect(dialog.getByRole('region', { name: '编辑提案新版本' })).not.toBeVisible();
    const current = await api.get(`/proposals/${proposal.id}`) as components['schemas']['S2ProposalResultV1'];
    expect(current.proposal.revision).toBe(3);
    expect(current.proposal.content.title).toBe('本轨保留的人工编辑');
    expect(current.proposal.content.stops.deadline).toBe(proposal.content.stops.deadline);
    expect((await api.get(`/proposals/${proposal.id}?revision=2`) as components['schemas']['S2ProposalResultV1']).proposal.content).toEqual(external.proposal.content);
  } finally { await server.close({ preserveData: true }); }
});

test('实际网络断开后的未知修订跨重开和刷新保持锁且不自动重发', async ({ page, context }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService();
  try {
    const { proposal } = await prepareDraft(api);
    let dialog = await openDraft(page, server.webURL, proposal);
    await dialog.getByRole('button', { name: '编辑新版本', exact: true }).click();
    await dialog.getByLabel('提案标题', { exact: true }).fill('本轨未确认的修订输入');
    const writes: string[] = [];
    page.on('request', request => { if (request.method() === 'PUT' && request.url().endsWith(`/proposals/${proposal.id}`)) writes.push(request.url()); });
    await context.setOffline(true);
    await dialog.getByRole('button', { name: '保存新版本', exact: true }).click();
    await expect(dialog.getByRole('alert').filter({ hasText: '结果未知' })).toBeVisible();
    await expect(dialog.getByLabel('提案标题', { exact: true })).toHaveValue('本轨未确认的修订输入');
    await expect(dialog.getByRole('button', { name: '保存新版本', exact: true })).toBeDisabled();
    await context.setOffline(false);
    await dialog.getByRole('button', { name: '关闭编辑', exact: true }).click();
    await dialog.getByRole('button', { name: '编辑新版本', exact: true }).click();
    await expect(dialog.getByRole('button', { name: '保存新版本', exact: true })).toBeDisabled();
    await page.reload();
    await page.getByRole('button', { name: `查看提案 ${proposal.content.title}`, exact: true }).click();
    dialog = page.getByRole('dialog', { name: '提案详情', exact: true });
    await dialog.getByRole('button', { name: '编辑新版本', exact: true }).click();
    await expect(dialog.getByRole('alert').filter({ hasText: '结果未知' })).toBeVisible();
    await expect(dialog.getByRole('button', { name: '保存新版本', exact: true })).toBeDisabled();
    expect(writes).toHaveLength(0);
    expect((await api.get(`/proposals/${proposal.id}`) as components['schemas']['S2ProposalResultV1']).proposal).toEqual(proposal);
  } finally { await context.setOffline(false); await server.close({ preserveData: true }); }
});

test('实际项目许可拒绝审阅，读取403时保留快照并禁用写入，恢复后重新读取', async ({ page, context }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService();
  try {
    const { proposal } = await prepareDraft(api, false);
    const dialog = await openDraft(page, server.webURL, proposal);
    const deniedResponse = page.waitForResponse(response => response.url().endsWith(`/proposals/${proposal.id}/submit`));
    await dialog.getByRole('button', { name: '提交审阅', exact: true }).click();
    const denied = await deniedResponse;
    expect(denied.status()).toBe(403);
    expect((await denied.json() as components['schemas']['ErrorV1']).error.code).toBe('approval_required');
    await expect(dialog.getByRole('alert')).toContainText('操作许可');
    expect((await api.get(`/proposals/${proposal.id}`) as components['schemas']['S2ProposalResultV1']).proposal.status).toBe('draft');
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_approvals WHERE proposal_id=?', proposal.id)).toEqual([{ count: 0 }]);
    await context.setExtraHTTPHeaders({ Authorization: 'Bearer invalid' });
    const failedRead = page.waitForResponse(response => response.url().endsWith(`/proposals/${proposal.id}`) && response.request().method() === 'GET');
    await dialog.getByRole('button', { name: '刷新当前提案', exact: true }).click();
    expect((await failedRead).status()).toBe(403);
    await expect(dialog.getByRole('alert').filter({ hasText: '数据可能已过期' })).toBeVisible();
    await expect(dialog.getByText(proposal.content.title, { exact: true })).toBeVisible();
    await expect(dialog.getByRole('button', { name: '提交审阅', exact: true })).toBeDisabled();
    await context.setExtraHTTPHeaders({});
    await dialog.getByRole('button', { name: '重试加载', exact: true }).click();
    await expect(dialog.getByRole('alert').filter({ hasText: '数据可能已过期' })).not.toBeVisible();
    await expect(dialog.getByRole('button', { name: '提交审阅', exact: true })).toBeEnabled();
  } finally { await context.setExtraHTTPHeaders({}); await server.close({ preserveData: true }); }
});

test('实际人工拒绝固定提案版本，保留内容且不签发Grant或启动执行', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService();
  try {
    const { project, proposal } = await prepareDraft(api);
    const dialog = await openDraft(page, server.webURL, proposal);
    await dialog.getByRole('button', { name: '提交审阅', exact: true }).click();
    await expect(dialog.getByRole('status', { name: '状态: 审核中', exact: true })).toBeVisible();
    const declinedResponse = page.waitForResponse(response => response.url().endsWith(`/proposals/${proposal.id}/decline`));
    await dialog.getByRole('button', { name: '拒绝当前版本', exact: true }).click();
    const declined = await declinedResponse;
    expect(declined.status()).toBe(200);
    const result = await declined.json() as components['schemas']['S2ProposalResultV1'];
    expect(result.proposal.status).toBe('declined');
    expect(result.proposal.revision).toBe(proposal.revision);
    expect(result.proposal.content).toEqual(proposal.content);
    await expect(dialog.getByRole('status', { name: '状态: 已拒绝', exact: true })).toBeVisible();
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_approvals WHERE proposal_id=?', proposal.id)).toEqual([{ count: 0 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions WHERE project_id=?', project.id)).toEqual([{ count: 0 }]);
    await server.restart();
    const restarted = await humanAPI(server.apiURL);
    expect((await restarted.get(`/proposals/${proposal.id}`) as components['schemas']['S2ProposalResultV1']).proposal).toEqual(result.proposal);
  } finally { await server.close({ preserveData: true }); }
});

test('真实固定资料引用经界面保存，项目移除引用后保留历史且新审阅拒绝', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService();
  try {
    const { project, proposal } = await prepareDraft(api);
    const imported = await importTrackedDocument(api);
    const material = imported.detail.material;
    const space = (await api.get(`/project-spaces/${project.space_id}`) as components['schemas']['ProjectSpaceResultV1']).space;
    const included = (await api.write(`/project-spaces/${project.space_id}/references`, { expected_version: space.version, material_id: material.id, revision: 1 }) as components['schemas']['ProjectSpaceResultV1']).space;
    expect(included.material_refs).toHaveLength(1);
    const dialog = await openDraft(page, server.webURL, proposal);
    await dialog.getByRole('button', { name: '编辑新版本', exact: true }).click();
    const picker = dialog.getByRole('group', { name: '项目已纳入的固定资料版本', exact: true });
    await expect(picker).toContainText(included.material_refs[0].locator);
    await picker.getByRole('checkbox').check();
    const savedResponse = page.waitForResponse(response => response.url().endsWith(`/proposals/${proposal.id}`) && response.request().method() === 'PUT');
    await dialog.getByRole('button', { name: '保存新版本', exact: true }).click();
    const saved = await savedResponse;
    expect(saved.status()).toBe(200);
    const fixed = (await saved.json() as components['schemas']['S2ProposalResultV1']).proposal;
    expect(fixed.content.inputs).toEqual([{ material_id: material.id, revision: 1 }]);
    expect((await api.get(`/proposals/${proposal.id}?revision=1`) as components['schemas']['S2ProposalResultV1']).proposal.content).toEqual(proposal.content);
    await api.write(`/project-spaces/${project.space_id}/references/remove`, { expected_version: included.version, material_id: material.id });
    await page.reload();
    const reopened = page.getByRole('dialog', { name: '提案详情', exact: true });
    await page.getByRole('button', { name: `查看提案 ${proposal.content.title}`, exact: true }).click();
    await reopened.getByRole('button', { name: '编辑新版本', exact: true }).click();
    await expect(reopened.getByRole('alert').filter({ hasText: '以下已保存引用当前未出现在所选项目空间中' })).toBeVisible();
    await reopened.getByRole('button', { name: '关闭编辑', exact: true }).click();
    const deniedResponse = page.waitForResponse(response => response.url().endsWith(`/proposals/${proposal.id}/submit`));
    await reopened.getByRole('button', { name: '提交审阅', exact: true }).click();
    const denied = await deniedResponse;
    expect(denied.status()).toBe(403);
    expect((await denied.json() as components['schemas']['ErrorV1']).error.code).toBe('scope_denied');
    expect((await api.get(`/proposals/${proposal.id}?revision=2`) as components['schemas']['S2ProposalResultV1']).proposal).toEqual(fixed);
    expect(server.query('SELECT COUNT(*) AS count FROM workspace_approvals WHERE proposal_id=?', proposal.id)).toEqual([{ count: 0 }]);
  } finally { await server.close({ preserveData: true }); }
});
