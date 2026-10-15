/// <reference types="node" />
import { test, expect, type Page } from '@playwright/test';
import type { components } from '../src/api/schema';
import { relayIntentService, declareIntent } from './realRelay';
import { relayDisplayState, relayLivenessLabel } from '../src/pages/workspace/relayPresentation';

type S = components['schemas'];
function writes(page: Page) {
  const commands: { action: string; body: Record<string, unknown>; key: string }[] = [];
  page.on('request', request => {
    const path = new URL(request.url()).pathname;
    if (request.method() === 'POST' && path.startsWith('/api/v1/relay/')) commands.push({ action: path.split('/').at(-1)!, body: request.postDataJSON(), key: request.headers()['idempotency-key'] });
  });
  return commands;
}

test('实际注册CLI、默认A确认与权限拒绝保留意图和收据', async ({ page }) => {
  test.setTimeout(120_000);
  const owned = await relayIntentService();
  try {
    const commands = writes(page), dialog = await declareIntent(page, owned);
    await expect(dialog.getByLabel('权限级别', { exact: true })).toHaveValue('A');
    await expect(dialog).toContainText(owned.project.root);
    await expect(dialog.getByRole('button', { name: '确认资料权限并请求接力' })).toBeDisabled();
    await dialog.getByRole('checkbox').check();
    const denied = page.waitForResponse(response => response.url().endsWith('/confirm'));
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    expect((await denied).status()).toBe(403);
    await expect(dialog.getByRole('alert')).toContainText('review_project_permissions');
    expect(commands.map(command => command.action)).toEqual(['intents', 'confirm']);
    const data = await owned.api.get('/relay/tasks') as S['RelayTasksV1'];
    expect(data.tasks).toHaveLength(1);
    expect(data.tasks[0].task.state).toBe('ready');
    expect(data.tasks[0].terminal).toBeNull();
    const receipts = await owned.api.get(`/relay/tasks/${data.tasks[0].task.task_id}/receipts`) as S['RelayReceiptsV1'];
    expect(receipts.receipts.find(receipt => receipt.action === 'dispatch')?.status).toBe('rejected');
    await page.reload();
    await page.getByRole('button', { name: '查看意图与权限', exact: true }).click();
    await expect(page.getByRole('dialog')).toContainText('被拒绝 · 首 RED 保留');
    expect(commands).toHaveLength(2);
  } finally { await owned.server.close(); }
});

test('真实空会话初始化权限拒绝保存prepare收据且不发送首次输入', async ({ page }) => {
  test.setTimeout(120_000);
  const owned = await relayIntentService();
  try {
    const commands = writes(page), dialog = await declareIntent(page, owned);
    const prepare = dialog.getByRole('button', { name: '确认资料权限并初始化会话', exact: true });
    await expect(prepare).toBeDisabled();
    await dialog.getByRole('checkbox').check();
    const denied = page.waitForResponse(response => response.url().endsWith('/prepare'));
    await prepare.click();
    expect((await denied).status()).toBe(403);
    await expect(dialog.getByRole('alert')).toContainText('review_project_permissions');
    expect(commands.map(command => command.action)).toEqual(['intents', 'prepare']);
    expect(commands[1].body).toEqual({ permission_confirmed: true, permission_level: 'A', project_settings_revision: owned.project.settings.revision });
    const { tasks } = await owned.api.get('/relay/tasks') as S['RelayTasksV1'];
    expect(tasks).toHaveLength(1);
    expect(tasks[0].task.state).toBe('ready');
    expect(tasks[0].terminal).toBeNull();
    expect(tasks[0].worktree).toBeNull();
    const { receipts } = await owned.api.get(`/relay/tasks/${tasks[0].task.task_id}/receipts`) as S['RelayReceiptsV1'];
    expect(receipts.find(receipt => receipt.action === 'prepare')).toMatchObject({ status: 'rejected', operation_id: commands[1].key });
    expect(receipts.filter(receipt => receipt.action === 'dispatch')).toEqual([]);
    await page.reload();
    await page.getByRole('button', { name: '查看意图与权限', exact: true }).click();
    await expect(page.getByRole('region', { name: '派发与意图收据', exact: true })).toContainText('初始化会话 · 被拒绝');
    expect(commands).toHaveLength(2);
    expect(owned.server.query('SELECT COUNT(*) AS count FROM local_agent_sessions')).toEqual([{ count: 0 }]);
  } finally { await owned.server.close(); }
});

test('真实网络断开的初始化未知锁同时阻止prepare和confirm并跨刷新保留', async ({ page, context }) => {
  test.setTimeout(120_000);
  const owned = await relayIntentService();
  try {
    const commands = writes(page), dialog = await declareIntent(page, owned);
    await dialog.getByRole('checkbox').check();
    await context.setOffline(true);
    await dialog.getByRole('button', { name: '确认资料权限并初始化会话', exact: true }).click();
    await expect(dialog).toContainText('初始化结果未知，禁止重新初始化或派发');
    const saved = await page.evaluate(() => Object.entries(sessionStorage).filter(([key]) => key.startsWith('astrocyte:relay:pending:prepare:')));
    expect(saved).toHaveLength(1);
    expect(saved[0][1]).toMatch(/^[0-9a-f-]{36}$/);
    await context.setOffline(false);
    await dialog.getByRole('button', { name: '读取权限、任务和收据', exact: true }).click();
    await expect(dialog.getByRole('button', { name: '确认资料权限并初始化会话', exact: true })).toBeDisabled();
    await expect(dialog.getByRole('button', { name: '确认资料权限并请求接力', exact: true })).toBeDisabled();
    await page.keyboard.press('Escape');
    await page.getByRole('button', { name: '查看意图与权限', exact: true }).click();
    await expect(page.getByRole('dialog')).toContainText('初始化结果未知，禁止重新初始化或派发');
    await page.reload();
    await page.getByRole('button', { name: '查看意图与权限', exact: true }).click();
    const reopened = page.getByRole('dialog');
    await expect(reopened.getByRole('button', { name: '确认资料权限并初始化会话', exact: true })).toBeDisabled();
    await expect(reopened.getByRole('button', { name: '确认资料权限并请求接力', exact: true })).toBeDisabled();
    expect(commands.filter(command => ['prepare', 'confirm'].includes(command.action))).toEqual([]);
    const { tasks } = await owned.api.get('/relay/tasks') as S['RelayTasksV1'];
    expect(tasks[0].task.state).toBe('ready');
    expect(tasks[0].terminal).toBeNull();
  } finally { await context.setOffline(false); await owned.server.close(); }
});

test('实际HTTP拒绝不成对交接声明，新意图不自动声明交接', async ({ page }) => {
  test.setTimeout(120_000);
  const owned = await relayIntentService();
  try {
    const body = { work_id: owned.work.id, project_id: owned.project.id, agent: 'codex', task: '只核成对声明校验，不启动控制器。' };
    for (const fields of [{ handoff_context_confirmed: true }, { handoff_from_task_id: crypto.randomUUID() }]) {
      const rejected = await owned.api.request('/relay/intents', { method: 'POST', body: { ...body, ...fields } });
      expect(rejected.status).toBe(400);
      expect((await owned.api.get('/relay/tasks') as S['RelayTasksV1']).tasks).toEqual([]);
    }
    const commands = writes(page);
    await declareIntent(page, owned);
    expect(commands).toHaveLength(1);
    expect(commands[0].body).not.toHaveProperty('handoff_from_task_id');
    expect(commands[0].body).not.toHaveProperty('handoff_context_confirmed');
    const data = await owned.api.get('/relay/tasks') as S['RelayTasksV1'];
    expect(data.tasks).toHaveLength(1);
    expect(data.tasks[0].task.state).toBe('ready');
    expect(data.tasks[0].terminal).toBeNull();
  } finally { await owned.server.close(); }
});

test('真实拒绝动作事实逐项对应HTTP和DOM，历史事实不授当前控制器归属', async ({ page }) => {
  test.setTimeout(120_000);
  const owned = await relayIntentService();
  try {
    await owned.api.write(`/local-projects/${owned.project.id}/settings`, { expected_revision: owned.project.settings.revision, settings: { external_model_cli: 'codex', allowed_actions: ['observe'] } }, { method: 'PUT' });
    await declareIntent(page, owned);
    const before = await owned.api.get('/relay/tasks') as S['RelayTasksV1'];
    const id = before.tasks[0].task.task_id;
    const rejected = await owned.api.request(`/relay/tasks/${id}/transition`, { method: 'POST', body: { from: 'working', to: 'done', completion_confirmed: true } });
    expect(rejected.status).toBe(409);
    const facts = await owned.api.get(`/relay/tasks/${id}/facts`) as S['RelayFactsV1'];
    expect(facts.current_ownership).toEqual({ status: 'unknown', reason: 'controller_binding_unrecorded' });
    expect(facts.session_relations).toEqual([]);
    expect(facts.facts.some(fact => fact.kind === 'action_result' && fact.status === 'rejected' && fact.reason === 'version_conflict')).toBe(true);
    await page.keyboard.press('Escape');
    const commands = writes(page);
    await page.getByRole('button', { name: '查看意图与权限', exact: true }).click();
    const region = page.getByRole('region', { name: '动作事实与会话关系', exact: true });
    await expect(region).toContainText('当前原控制器归属未知');
    await expect(region).toContainText('controller_binding_unrecorded');
    await expect(region.getByRole('list', { name: '已记录动作事实' }).locator('li')).toHaveCount(facts.facts.length);
    for (const fact of facts.facts) {
      const row = region.getByRole('list', { name: '已记录动作事实' }).locator('li').filter({ hasText: fact.fact_id });
      await expect(row).toContainText(fact.source);
      if (fact.receipt_id) await expect(row).toContainText(fact.receipt_id);
      if (fact.operation_id) await expect(row).toContainText(fact.operation_id);
      if (fact.reason) await expect(row).toContainText(fact.reason);
    }
    await expect(region).toContainText('尚无已记录会话关系');
    expect(await owned.api.get('/relay/tasks')).toEqual(before);
    expect(await owned.api.get(`/relay/tasks/${id}/facts`)).toEqual(facts);
    expect(commands).toEqual([]);
  } finally { await owned.server.close(); }
});

test('实际B与C独立设置保留服务权限摘要，未许可原生操作时拒绝', async ({ page }) => {
  test.setTimeout(120_000);
  const owned = await relayIntentService();
  try {
    await owned.api.write(`/local-projects/${owned.project.id}/settings`, { expected_revision: owned.project.settings.revision, settings: { allow_directory: true, allowed_subdirs: ['AGENTS.md'], expand_references: true, allowed_actions: [] } }, { method: 'PUT' });
    const commands = writes(page), dialog = await declareIntent(page, owned);
    await expect(dialog.getByLabel('权限级别', { exact: true })).toHaveValue('C');
    await expect(dialog.getByLabel('权限级别', { exact: true })).toHaveAttribute('readonly', '');
    await expect(dialog).toContainText('B · 明确目录许可');
    await expect(dialog).toContainText('C · 引用展开许可');
    await expect(dialog).toContainText('AGENTS.md');
    await dialog.getByRole('checkbox').check();
    const denied = page.waitForResponse(response => response.url().endsWith('/confirm'));
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    expect((await denied).status()).toBe(403);
    expect(commands[1].body.permission_level).toBe('C');
  } finally { await owned.server.close(); }
});

test('真实权限版本变化阻断确认并保留原意图', async ({ page }) => {
  test.setTimeout(120_000);
  const owned = await relayIntentService();
  try {
    const commands = writes(page), dialog = await declareIntent(page, owned);
    await dialog.getByRole('checkbox').check();
    await owned.api.write(`/local-projects/${owned.project.id}/settings`, { expected_revision: owned.project.settings.revision, settings: { expand_references: true } }, { method: 'PUT' });
    await dialog.getByRole('button', { name: '读取权限、任务和收据' }).click();
    await expect(dialog.getByRole('alert')).toContainText('任务的权限版本已过期');
    await expect(dialog.getByRole('button', { name: '确认资料权限并请求接力' })).toBeDisabled();
    expect(commands.map(command => command.action)).toEqual(['intents']);
  } finally { await owned.server.close(); }
});

test('真实网络中断后的派发未知锁在读取、重开和刷新后保持', async ({ page, context }) => {
  test.setTimeout(120_000);
  const owned = await relayIntentService();
  try {
    const commands = writes(page), dialog = await declareIntent(page, owned);
    await dialog.getByRole('checkbox').check();
    await context.setOffline(true);
    await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
    await expect(dialog).toContainText('派发结果未知，禁止重派');
    const saved = await page.evaluate(() => Object.entries(sessionStorage).filter(([key]) => key.startsWith('astrocyte:relay:pending:confirm:')));
    expect(saved).toHaveLength(1);
    expect(saved[0][1]).toMatch(/^[0-9a-f-]{36}$/);
    await context.setOffline(false);
    await dialog.getByRole('button', { name: '读取权限、任务和收据' }).click();
    await expect(dialog.getByRole('button', { name: '确认资料权限并请求接力' })).toBeDisabled();
    await page.keyboard.press('Escape');
    await page.getByRole('button', { name: '查看意图与权限', exact: true }).click();
    await expect(page.getByRole('dialog').getByRole('button', { name: '确认资料权限并请求接力' })).toBeDisabled();
    await page.reload();
    await page.getByRole('button', { name: '查看意图与权限', exact: true }).click();
    await expect(page.getByRole('dialog')).toContainText('派发结果未知，禁止重派');
    expect(commands.filter(command => command.action === 'confirm')).toHaveLength(0);
  } finally { await context.setOffline(false); await owned.server.close(); }
});

/** 原生动作只在H明确交付的固定公开任务和单执行人窗口内运行。缺少场景即失败，不伪造状态。 */
async function nativeTask(page: Page, scenario: string) {
  const origin = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_URL;
  const id = process.env[`ASTROCYTE_TEST_RELAY_${scenario}_${page.viewportSize()?.width}_TASK_ID`];
  expect(origin, '需要H交付当前实际页面URL').toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
  expect(id, `需要H交付${scenario}本次固定公开任务与执行窗口`).toMatch(/^[0-9a-f-]{36}$/);
  const commands = writes(page);
  await page.goto(`${origin}/workspace/works`);
  const card = page.locator('.relay-tasks > li').filter({ hasText: id! });
  await expect(card).toBeVisible();
  const response = await page.request.get(`${origin}/api/v1/relay/tasks/${id}`);
  expect(response.status()).toBe(200);
  const { view } = await response.json() as S['RelayTaskResultV1'];
  await card.getByRole('button', { name: view.task.state === 'ready' ? '查看意图与权限' : '查看任务、事件和收据', exact: true }).click();
  return { origin: origin!, id: id!, view, dialog: page.getByRole('dialog'), commands };
}

for (const drive of ['terminal', 'structured'] as const) {
  test(`实际EOF登记与${drive}驱动保留正确发送能力`, async ({ page }) => {
    const { dialog, view, commands } = await nativeTask(page, `EOF_${drive.toUpperCase()}`);
    expect(view.terminal?.drive_mode).toBe(drive);
    if (drive === 'terminal') {
      await expect(dialog).toContainText('首条手册与任务发送后关闭标准输入');
      await expect(dialog.getByLabel('发送输入', { exact: true })).toBeDisabled();
      await expect(dialog.getByRole('button', { name: '发送输入', exact: true })).toBeDisabled();
      await dialog.getByRole('button', { name: '读取会话并留收据', exact: true }).click();
      await expect(dialog.getByRole('status').filter({ hasText: '读取会话收据' })).toBeVisible();
      await expect(dialog.getByRole('button', { name: '停止 Agent', exact: true })).toBeEnabled();
      expect(commands.map(command => command.action)).toEqual(['read']);
    } else {
      await expect(dialog).not.toContainText('当前终端登记为一次性输入');
      await dialog.getByLabel('发送输入', { exact: true }).fill('只核对既有固定公开任务，不使用文件或命令工具。');
      await dialog.getByRole('button', { name: '发送输入', exact: true }).click();
      await expect(dialog.getByRole('status').filter({ hasText: '发送输入收据' })).toBeVisible();
      expect(commands.map(command => command.action)).toEqual(['send']);
    }
  });
}

for (const stage of ['initial', 'reopened']) {
  test(`实际${stage}互斥收据只提供人工GET核对`, async ({ page }) => {
    const { dialog, commands, origin, id } = await nativeTask(page, `CONFLICT_${stage.toUpperCase()}`);
    const response = await page.request.get(`${origin}/api/v1/relay/tasks/${id}/receipts`);
    expect(response.status()).toBe(200);
    const receipts = await response.json() as S['RelayReceiptsV1'];
    const rejected = receipts.receipts.find(receipt => receipt.status === 'rejected' && receipt.error?.required_action.startsWith('view_existing_task:'));
    expect(rejected).toBeDefined();
    const existingID = rejected!.error!.required_action.slice('view_existing_task:'.length);
    const existing = dialog.getByRole('region', { name: `已有任务 ${existingID}`, exact: true });
    await expect(existing).toContainText(existingID);
    await existing.getByRole('button', { name: '只读查看已有任务', exact: true }).click();
    await expect(existing).toContainText('工作中 · working');
    expect(commands).toEqual([]);
  });
}

test('真实重复派发409保留原任务，刷新后只读核对已有任务', async ({ page }) => {
  const { dialog, commands, view } = await nativeTask(page, 'DUPLICATE');
  expect(view.task.state).toBe('ready');
  await dialog.getByRole('checkbox').check();
  const refused = page.waitForResponse(response => response.url().endsWith('/confirm'));
  await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
  const response = await refused;
  expect(response.status()).toBe(409);
  const body = await response.json() as S['ErrorV1'];
  expect(body.error.required_action).toMatch(/^view_existing_task:[0-9a-f-]{36}$/);
  const id = body.error.required_action.slice('view_existing_task:'.length);
  await dialog.getByRole('region', { name: `已有任务 ${id}`, exact: true }).getByRole('button', { name: '只读查看已有任务' }).click();
  await expect(dialog.getByRole('region', { name: `已有任务 ${id}`, exact: true })).toContainText('工作中 · working');
  await page.reload();
  await page.locator('.relay-tasks > li').filter({ hasText: view.task.task_id }).getByRole('button', { name: '查看意图与权限' }).click();
  await expect(page.getByRole('dialog').getByRole('region', { name: `已有任务 ${id}`, exact: true })).toBeVisible();
  expect(commands.map(command => command.action)).toEqual(['confirm']);
});

test('实际通用冲突不显示已有任务入口或自动再次确认', async ({ page }) => {
  const { dialog, commands } = await nativeTask(page, 'GENERIC_CONFLICT');
  await dialog.getByRole('checkbox').check();
  const refused = page.waitForResponse(response => response.url().endsWith('/confirm'));
  await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
  const response = await refused;
  expect(response.status()).toBe(409);
  expect((await response.json() as S['ErrorV1']).error.required_action).toBe('refresh_relay_state');
  await expect(dialog.getByRole('button', { name: '只读查看已有任务', exact: true })).toHaveCount(0);
  await dialog.getByRole('button', { name: '读取权限、任务和收据' }).click();
  expect(commands.map(command => command.action)).toEqual(['confirm']);
});

test('真实服务投影在刷新后保留任务、终端、收据及三值活性', async ({ page }) => {
  const { dialog, view, commands } = await nativeTask(page, 'WORKING');
  expect(view.task.state).toBe('working');
  expect(view.terminal).not.toBeNull();
  await expect(dialog).toContainText(relayLivenessLabel[view.terminal!.liveness]);
  await expect(dialog).toContainText(relayDisplayState(view));
  await expect(dialog).toContainText(view.terminal!.terminal_id);
  await expect(dialog).toContainText(view.worktree!.branch);
  await page.keyboard.press('Escape');
  await page.reload();
  await expect(page.locator('.relay-tasks > li').filter({ hasText: view.task.task_id })).toContainText(view.terminal!.terminal_id);
  expect(commands).toEqual([]);
});

test('实际离线发送保留输入、禁止重发且仍允许停止', async ({ page, context }) => {
  const { dialog, view, commands } = await nativeTask(page, 'UNKNOWN_SEND');
  expect(view.task.state).toBe('working');
  const input = '核对既有固定公开任务，保持文件与命令工具关闭。';
  await dialog.getByLabel('发送输入', { exact: true }).fill(input);
  await context.setOffline(true);
  try {
    await dialog.getByRole('button', { name: '发送输入', exact: true }).click();
    await expect(dialog).toContainText('动作结果未知');
    await expect(dialog.getByLabel('发送输入', { exact: true })).toHaveValue(input);
    await context.setOffline(false);
    await dialog.getByRole('button', { name: '只读取任务、事件和收据' }).click();
    await expect(dialog.getByRole('button', { name: '发送输入', exact: true })).toBeDisabled();
    await expect(dialog.getByRole('button', { name: '停止 Agent', exact: true })).toBeEnabled();
    await dialog.getByRole('button', { name: '停止 Agent', exact: true }).click();
    await expect(dialog).toContainText('terminal_exited');
    expect(commands.filter(command => command.action === 'send')).toHaveLength(0);
    expect(commands.filter(command => command.action === 'stop')).toHaveLength(1);
  } finally { await context.setOffline(false); }
});

test('实际人工完成取消、done读取停止与退出后释放均可达', async ({ page }) => {
  const { dialog, commands, view } = await nativeTask(page, 'COMPLETE_RELEASE');
  expect(view.task.state).toBe('working');
  const controls = dialog.getByRole('region', { name: '人工完成与交接' });
  await expect(controls.getByRole('button', { name: '确认任务完成' })).toBeDisabled();
  await controls.getByRole('checkbox').check();
  await controls.getByRole('button', { name: '取消确认' }).click();
  await expect(controls.getByRole('checkbox')).not.toBeChecked();
  expect(commands).toEqual([]);
  await controls.getByRole('checkbox').check();
  await controls.getByRole('button', { name: '确认任务完成' }).click();
  await expect(dialog).toContainText('任务已完成 · done');
  await expect(dialog.getByRole('button', { name: '发送输入', exact: true })).toBeDisabled();
  await expect(dialog.getByRole('button', { name: '读取会话并留收据' })).toBeEnabled();
  await expect(controls.getByRole('button', { name: '确认释放工作区' })).toBeDisabled();
  await dialog.getByRole('button', { name: '读取会话并留收据' }).click();
  await dialog.getByRole('button', { name: '停止 Agent' }).click();
  await expect(dialog).toContainText('终端已退出 · exited');
  await controls.getByRole('checkbox').check();
  await controls.getByRole('button', { name: '确认释放工作区' }).click();
  await expect(dialog).toContainText('已释放 · released');
  expect(commands.map(command => command.action)).toEqual(['transition', 'read', 'stop', 'transition']);
});

test('真实释放工作区由另一注册Agent接续，重新经过权限确认', async ({ page }) => {
  const { dialog, view, commands, origin } = await nativeTask(page, 'ALTERNATE');
  expect(view.task.state).toBe('ready');
  const previousID = process.env[`ASTROCYTE_TEST_RELAY_ALTERNATE_PREVIOUS_${page.viewportSize()?.width}_TASK_ID`];
  expect(previousID).toMatch(/^[0-9a-f-]{36}$/);
  const previousResponse = await page.request.get(`${origin}/api/v1/relay/tasks/${previousID}`);
  expect(previousResponse.status()).toBe(200);
  const previous = (await previousResponse.json() as S['RelayTaskResultV1']).view;
  expect(previous.task.state).toBe('released');
  expect(view.task.agent).not.toBe(previous.task.agent);
  await dialog.getByRole('checkbox').check();
  await dialog.getByRole('button', { name: '确认资料权限并请求接力' }).click();
  await expect(dialog.getByRole('status').filter({ hasText: '派发收据' })).toContainText('动作已接受');
  const currentResponse = await page.request.get(`${origin}/api/v1/relay/tasks/${view.task.task_id}`);
  expect(currentResponse.status()).toBe(200);
  const current = (await currentResponse.json() as S['RelayTaskResultV1']).view;
  expect(current.worktree).toMatchObject({ worktree_id: previous.worktree!.worktree_id, root: previous.worktree!.root, branch: previous.worktree!.branch, created_at: previous.worktree!.created_at });
  expect(commands.map(command => command.action)).toEqual(['confirm']);
});

test('实际事件接入错误由人工核对，发布不表示认知快照更新', async ({ page }) => {
  const { dialog, view, commands } = await nativeTask(page, 'PUBLICATION_ERROR');
  expect(view.event_publication.status).toBe('error');
  await expect(dialog).toContainText('事件接入失败');
  await expect(dialog).toContainText('认知数值仍来自已发布快照');
  expect(commands).toEqual([]);
  await dialog.getByRole('button', { name: '核对事件接入' }).click();
  await expect(dialog).toContainText('已接入本机应用事件');
  await expect(dialog.getByRole('region', { name: '任务收据' })).toContainText('核对事件接入 · 动作已接受');
  expect(commands.map(command => command.action)).toEqual(['reconcile']);
});

test('真实嵌套项目根目录不会作为作品的完全匹配许可', async ({ page }) => {
  test.setTimeout(120_000);
  const owned = await relayIntentService({ nestedRoot: true });
  try {
    const commands = writes(page);
    await page.goto(`${owned.server.webURL}/workspace/works`);
    await page.getByRole('region', { name: '作品寻回册', exact: true }).getByLabel(`${owned.work.name} 的操作`, { exact: true }).getByRole('button', { name: '接力', exact: true }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toContainText('作品尚无根目录完全匹配的获准项目');
    await expect(dialog.getByRole('button', { name: '保存意图并查看权限' })).toBeDisabled();
    expect(commands).toEqual([]);
  } finally { await owned.server.close(); }
});
