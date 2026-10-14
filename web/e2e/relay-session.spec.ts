/// <reference types="node" />
import { test, expect, type Page, type APIResponse, type Response } from '@playwright/test';
import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import type { RelaySessionEvents, RelaySessionCoverage } from '../src/api/relaySession';
import type { RelayEvent, RelayTaskView } from '../src/api/relay';
import type { CognitionToday } from '../src/api/cognitionTypes';

type PublicFacts = {
  task_id: string; project_id: string; agent_name: string;
  registered_at: string; dispatched_at: string; events: RelayEvent[];
  forbidden_canaries: string[];
};
type Observation = {
  origin: string; agent_name: string; registry_before: string[];
  viewports: { name: string; first_visible_at: string; event_ids: string[] }[];
};
const eventFields = ['event_id', 'kind', 'scope', 'tokens', 'topic', 'ts'];
const liveLabels = { live: '近期活动确证 · live', unverifiable: '活性无法核实 · unverifiable', exited: '终端已退出 · exited' };
let origin: string, facts: PublicFacts, observation: Observation;

test.beforeAll(async () => {
  origin = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_URL || '';
  expect(origin).toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
  const expectedPath = process.env.ASTROCYTE_TEST_RELAY_L1_GETEVENT_PUBLIC_FILE;
  expect(expectedPath, '需要 H 从实际 GetEvent 导出的公开事实文件').toBeTruthy();
  facts = JSON.parse(await readFile(expectedPath!, 'utf8')) as PublicFacts;
  observation = JSON.parse(await readFile(join(process.cwd(), 'test-results/l1-live/first-visible.json'), 'utf8')) as Observation;
  expect(facts.events.length).toBeGreaterThanOrEqual(2);
  expect(facts.forbidden_canaries.length, '需要实际原生输入中的公开 canary 标记').toBeGreaterThan(0);
  expect(observation.origin).toBe(origin);
  expect(observation.agent_name).toBe(facts.agent_name);
  expect(observation.registry_before).not.toContain(facts.agent_name);
  expect(Date.parse(facts.registered_at)).toBeLessThanOrEqual(Date.parse(facts.dispatched_at));
  for (const event of facts.events) expect(Object.keys(event).sort()).toEqual(eventFields);
});

async function read<T>(response: APIResponse | Response): Promise<T> {
  expect(response.status()).toBe(200);
  return await response.json() as T;
}
function sessionResponse(page: Page, query: Record<string, string>) {
  return page.waitForResponse(response => {
    const url = new URL(response.url());
    return url.pathname === '/api/v1/relay/session-events' && Object.entries(query).every(([key, value]) => url.searchParams.get(key) === value);
  });
}
async function enter(page: Page) {
  const response = sessionResponse(page, {});
  await page.goto(`${origin}/cognition/today`);
  expect((await response).status()).toBe(200);
  return page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
}
async function assertRows(page: Page, data: RelaySessionEvents) {
  const region = page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
  await expect(region.locator('tbody tr')).toHaveCount(data.events.length);
  if (data.events.length) await expect(region.locator('thead th')).toHaveText(['事件 ID', '类型', '主题', 'Token 数', '范围', '时间（上海）']);
  for (const [index, event] of data.events.entries()) {
    expect(Object.keys(event).sort()).toEqual(eventFields);
    const committed = facts.events.find(item => item.event_id === event.event_id);
    expect(committed, 'HTTP ID 必须对应 H 的实际 GetEvent 结果').toBeDefined();
    expect(event).toEqual(committed);
    const cells = region.locator('tbody tr').nth(index).locator('td');
    await expect(cells).toHaveCount(6);
    await expect(cells.nth(0)).toHaveText(event.event_id);
    await expect(cells.nth(1)).toHaveText(event.kind);
    await expect(cells.nth(2)).toHaveText(event.topic || '未标注');
    await expect(cells.nth(3)).toHaveText(event.tokens === null ? '暂无' : new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 4 }).format(event.tokens));
    await expect(cells.nth(4)).toHaveText(event.scope || '未标注');
    await expect(cells.nth(5)).toHaveText(new Date(event.ts).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai', hour12: false }));
  }
  const html = await region.evaluate(node => node.outerHTML);
  for (const canary of facts.forbidden_canaries) expect(html).not.toContain(canary);
}
async function assertCoverage(page: Page, item: RelaySessionCoverage) {
  const coverage = page.getByRole('article', { name: `会话覆盖 ${item.task_id}`, exact: true });
  await expect(coverage).toContainText(`· ${item.status}`);
  await expect(coverage).toContainText(`当前文件快照：${item.snapshot_complete ? '已读完当前冻结字节范围' : '尚未读完或无法核实'}`);
  await expect(coverage).toContainText(`完整会话生命周期：${item.session_complete ? '服务报告已核实' : '尚未核实'}`);
  const counts = [['完整行', item.lines], ['已派生', item.projected], ['重复', item.duplicates], ['未知格式', item.unknown], ['错误', item.errors], ['Token 未知', item.unknown_tokens], ['剩余字节', item.remaining_bytes], ['未完成尾部字节', item.tail_bytes]] as const;
  for (const [label, value] of counts) await expect(coverage.locator('dl div').filter({ has: page.locator('dt', { hasText: new RegExp(`^${label}$`) }) }).locator('dd')).toHaveText(String(value));
  if (item.gap || item.unknown || item.errors || item.tail_bytes || item.remaining_bytes) await expect(coverage).toContainText('不能据此宣称会话完整');
}

test('actual registered native Session appears in the same Today URL within three minutes and survives real paging', async ({ page }, info) => {
  const writes: string[] = [], errors: string[] = [], sessionQueries: URL[] = [];
  page.on('request', request => {
    const url = new URL(request.url());
    if (url.pathname.startsWith('/api/v1/') && request.method() !== 'GET') writes.push(request.method());
    if (url.pathname === '/api/v1/relay/session-events') sessionQueries.push(url);
  });
  page.on('pageerror', error => errors.push(error.name));
  const first = observation.viewports.find(item => item.name === info.project.name);
  expect(first, '当前原 viewport 必须有实际浏览器首次可见记录').toBeDefined();
  const elapsed = Date.parse(first!.first_visible_at) - Date.parse(facts.dispatched_at);
  expect(elapsed).toBeGreaterThanOrEqual(0);
  expect(elapsed).toBeLessThanOrEqual(180_000);
  expect(first!.event_ids.some(id => facts.events.some(event => event.event_id === id))).toBe(true);
  const region = await enter(page);
  const registry = await read<{ agents: { name: string }[] }>(await page.request.get(`${origin}/api/v1/relay/agents`));
  expect(registry.agents.some(agent => agent.name === facts.agent_name)).toBe(true);
  const snapshotBefore = await page.request.get(`${origin}/api/v1/cognition/today`);
  expect([200, 503]).toContain(snapshotBefore.status());
  let snapshot: CognitionToday | undefined;
  if (snapshotBefore.status() === 200) {
    snapshot = await snapshotBefore.json() as CognitionToday;
    for (const item of snapshot.focus) for (const [prefix, value] of [['a', item.attention], ['m', item.momentum], ['c', item.confidence]] as const) await expect(page.getByTestId(`today-${prefix}-${item.topic}`)).toHaveText(new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 4 }).format(value));
    for (const item of snapshot.daily_counts) await expect(page.getByTestId(`daily-${item.day}`)).toHaveText(new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 4 }).format(item.count));
    await expect(page.getByTestId('sleep-status')).toHaveText({ completed: '已完成', running: '运行中', not_produced: '尚未产出' }[snapshot.sleep.status]);
  } else {
    await expect(page.getByText('API 503', { exact: false }).first()).toBeVisible();
    await expect(page.locator('[data-testid^="today-a-"]')).toHaveCount(0);
  }
  const valuesBefore = await page.locator('[data-testid^="today-a-"], [data-testid^="today-m-"], [data-testid^="today-c-"], [data-testid^="daily-"], [data-testid="sleep-status"]').allTextContents();
  await region.getByLabel('会话任务 ID', { exact: true }).fill('invalid/task/id');
  await region.getByRole('button', { name: '读取所选事件', exact: true }).click();
  await expect(region).toContainText('查询未提交');
  expect(sessionQueries.some(url => url.searchParams.get('task_id') === 'invalid/task/id')).toBe(false);
  await region.getByLabel('会话任务 ID', { exact: true }).fill(facts.task_id);
  await region.getByLabel('会话每页条数', { exact: true }).selectOption('100');
  let response = sessionResponse(page, { task_id: facts.task_id, page_size: '100', page: '1' });
  await region.getByRole('button', { name: '读取所选事件', exact: true }).click();
  const full = await read<RelaySessionEvents>(await response);
  expect(full.total).toBe(facts.events.length);
  expect(full.events).toEqual(facts.events);
  expect(full.events.some(event => event.tokens === null)).toBe(true);
  expect(full.coverage).toHaveLength(1);
  expect(full.coverage[0].task_id).toBe(facts.task_id);
  expect(full.coverage[0].project_id).toBe(facts.project_id);
  expect(full.coverage[0].gap).not.toBe('');
  expect(full.coverage[0].session_complete).toBe(false);
  await assertRows(page, full); await assertCoverage(page, full.coverage[0]);
  await region.getByLabel('会话每页条数', { exact: true }).selectOption('1');
  response = sessionResponse(page, { task_id: facts.task_id, page_size: '1', page: '1' });
  await region.getByRole('button', { name: '读取所选事件', exact: true }).click();
  const firstPage = await read<RelaySessionEvents>(await response);
  expect(firstPage.events).toEqual(full.events.slice(0, 1));
  await assertRows(page, firstPage);
  response = sessionResponse(page, { task_id: facts.task_id, page_size: '1', page: '2' });
  await region.getByRole('button', { name: '会话下一页', exact: true }).click();
  const secondPage = await read<RelaySessionEvents>(await response);
  expect(secondPage.events).toEqual(full.events.slice(1, 2));
  expect(secondPage.total).toBe(full.total);
  await assertRows(page, secondPage); await assertCoverage(page, secondPage.coverage[0]);
  const start = full.events.at(-1)!.ts, end = full.events[0].ts;
  await region.getByLabel('会话开始时间', { exact: true }).fill(start);
  await region.getByLabel('会话结束时间', { exact: true }).fill(end);
  response = sessionResponse(page, { task_id: facts.task_id, start, end, page: '1' });
  await region.getByRole('button', { name: '读取所选事件', exact: true }).click();
  const bounded = await read<RelaySessionEvents>(await response);
  expect(bounded.total).toBe(full.total);
  expect(bounded.events).toEqual(full.events.slice(0, 1));
  await assertRows(page, bounded);
  const emptyStart = new Date(Date.now() + 86_400_000).toISOString();
  await region.getByLabel('会话开始时间', { exact: true }).fill(emptyStart);
  await region.getByLabel('会话结束时间', { exact: true }).fill('');
  response = sessionResponse(page, { start: emptyStart });
  await region.getByRole('button', { name: '读取所选事件', exact: true }).click();
  const empty = await read<RelaySessionEvents>(await response);
  expect(empty.events).toEqual([]); expect(empty.total).toBe(0);
  await expect(region).toContainText('空结果不证明会话完整');
  await assertCoverage(page, empty.coverage[0]);
  expect(await page.locator('[data-testid^="today-a-"], [data-testid^="today-m-"], [data-testid^="today-c-"], [data-testid^="daily-"], [data-testid="sleep-status"]').allTextContents()).toEqual(valuesBefore);
  const snapshotAfter = await page.request.get(`${origin}/api/v1/cognition/today`);
  expect(snapshotAfter.status()).toBe(snapshotBefore.status());
  if (snapshot) expect(await snapshotAfter.json()).toEqual(snapshot);
  expect(writes).toEqual([]); expect(errors).toEqual([]);
});

test('actual Origin refusal hides previously authorized Session rows and fresh bootstrap reads recover', async ({ page }) => {
  const region = await enter(page);
  await expect(region.locator('tbody tr').filter({ hasText: facts.events[0].event_id })).toHaveCount(1);
  await page.setExtraHTTPHeaders({ Origin: 'http://127.0.0.1:1' });
  const rejected = sessionResponse(page, {});
  await region.getByRole('button', { name: '只读刷新会话事件', exact: true }).click();
  expect((await rejected).status()).toBe(403);
  await expect(region).toContainText('当前会话覆盖未知');
  await expect(region.locator('tbody tr')).toHaveCount(0);
  await page.setExtraHTTPHeaders({});
  const recovered = sessionResponse(page, {});
  await region.getByRole('button', { name: '只读刷新会话事件', exact: true }).click();
  expect((await recovered).status()).toBe(200);
  await expect(region.locator('tbody tr').filter({ hasText: facts.events[0].event_id })).toHaveCount(1);
});

test('actual terminal API and task cards retain the observed three-value liveness and incomplete-session boundary', async ({ page }) => {
  await page.goto(`${origin}/workspace/works`);
  const data = await read<{ tasks: RelayTaskView[] }>(await page.request.get(`${origin}/api/v1/relay/tasks`));
  const selected = data.tasks.filter(view => view.task.project_id === facts.project_id && view.terminal);
  expect(selected.length).toBeGreaterThan(0);
  for (const view of selected) {
    expect(Object.keys(liveLabels)).toContain(view.terminal!.liveness);
    const card = page.locator('.relay-tasks > li').filter({ hasText: view.task.task_id });
    await expect(card.locator('.relay-liveness')).toHaveText(liveLabels[view.terminal!.liveness]);
    if (view.terminal!.liveness === 'exited' && view.task.state !== 'done') await expect(card).not.toContainText('任务已完成');
  }
});
