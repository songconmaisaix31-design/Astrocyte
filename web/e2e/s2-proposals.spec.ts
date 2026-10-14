import { test, expect, type Locator } from '@playwright/test';
import { realService, registerCurrentProject } from './realService';
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
  await form.getByLabel('允许动作', { exact: true }).fill('start\nclaim\nsubmit\naccept');
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

test('真实提案创建、不可变版本、人工批准与撤销通过HTTP及SQLite重启核查', async ({ page }) => {
  test.setTimeout(240_000);
  const { server, api } = await realService();
  try {
    const project = await registerCurrentProject(api);
    await api.write(`/local-projects/${project.id}/settings`, {
      expected_revision: project.settings.revision,
      settings: { allow_directory: true, allowed_subdirs: ['web'], allowed_actions: [], allowed_tools: [], external_model_cli: '', allow_agent_control: false },
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
    await page.getByRole('button', { name: '查看提案 本轨真实记录核查 r1', exact: true }).click();
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
    await expect(dialog.getByText('审核中', { exact: true })).toBeVisible();
    await dialog.getByLabel('批准有效期', { exact: true }).fill(localMinute(new Date(Date.now() + 3600_000)));
    await dialog.getByRole('checkbox', { name: '我已核对当前固定版本、范围、资源上限和停止条件' }).check();
    const approvedResponse = page.waitForResponse(response => response.url().endsWith(`/proposals/${original.id}/approvals`));
    await dialog.getByRole('button', { name: '批准当前版本', exact: true }).click();
    const approved = await approvedResponse;
    expect(approved.status()).toBe(200);
    const { approval } = await approved.json() as components['schemas']['S2ApprovalResultV1'];
    expect(approval.proposal_revision).toBe(2);
    expect(approval.status).toBe('active');
    const missions = (await api.get('/missions') as components['schemas']['MissionListV1']).items;
    expect(missions.length).toBeLessThanOrEqual(1);
    for (const mission of missions) expect(mission).toMatchObject({ state: 'ready', grant: { id: approval.id } });
    expect(server.query('SELECT COUNT(*) AS count FROM foundation_outbox WHERE event_type=? AND aggregate_id=?', 'workspace.proposal_approved', original.id)).toEqual([{ count: 1 }]);
    expect(server.query('SELECT COUNT(*) AS count FROM local_agent_sessions WHERE project_id=?', project.id)).toEqual([{ count: 0 }]);
    expect((await readdir(server.dataDir)).includes('native-sessions')).toBe(false);
    await dialog.getByLabel('撤销原因', { exact: true }).fill('本轨核查完成，停止后续新动作');
    await dialog.getByRole('button', { name: '撤销批准', exact: true }).click();
    await expect(dialog.getByText('已撤销', { exact: true })).toBeVisible();
    expect(server.query('SELECT epoch, json_extract(data, \'$.status\') AS status FROM workspace_approvals WHERE id=?', approval.id)).toEqual([{ epoch: 2, status: 'revoked' }]);
    expect(server.query('SELECT revision, json_extract(data, \'$.content.title\') AS title FROM workspace_proposal_revisions WHERE id=? ORDER BY revision', original.id)).toEqual([{ revision: 1, title: original.content.title }, { revision: 2, title: '本轨真实记录核查 r2' }]);
    await server.restart();
    const restarted = await humanAPI(server.apiURL);
    expect((await restarted.get(`/approvals/${approval.id}`) as components['schemas']['S2ApprovalResultV1']).approval.status).toBe('revoked');
    expect((await restarted.get(`/proposals/${original.id}?revision=1`) as components['schemas']['S2ProposalResultV1']).proposal.content).toEqual(original.content);
    await page.reload();
    await page.getByRole('button', { name: '查看提案 本轨真实记录核查 r2', exact: true }).click();
    await expect(page.getByRole('dialog').getByText('已撤销', { exact: true })).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  } finally { await server.close({ preserveData: true }); }
});
