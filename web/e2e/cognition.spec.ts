import { test, expect, type Page } from '@playwright/test';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import { cognitionTestServer } from '../src/api/realHttpTest';
import type { CognitionToday, CognitionFocus, CognitionEvents, CognitionAgents, CognitionFilters, CognitionEvidence, CognitionEvent } from '../src/api/cognitionTypes';
import { formatNumber, formatDate } from '../src/pages/cognition/format';

const columns = ['事件 ID', '类型', '主题', 'Token 数', '范围', '时间（上海）'];
let server: Awaited<ReturnType<typeof startS1Server>>, api: Awaited<ReturnType<typeof humanAPI>>;
test.beforeAll(async () => {
  test.setTimeout(120_000);
  server = await cognitionTestServer(true);
  api = await humanAPI(server.apiURL);
});
test.afterAll(async () => { if (server) await server.close(); });
const go = (route: string) => `${server.webURL}${route}`;
async function read<T>(path: string) { return await api.get(`/cognition${path}`) as T; }
function observe(page: Page) {
  const requests: URL[] = [], writes: string[] = [];
  page.on('request', request => {
    const url = new URL(request.url());
    if (url.pathname.startsWith('/api/v1/cognition/')) requests.push(url);
    if (url.pathname.startsWith('/api/v1/') && request.method() !== 'GET') writes.push(url.pathname);
  });
  return { requests, writes };
}
async function assertEvents(page: Page, events: CognitionEvent[], container = page.locator('.cg-page')) {
  await expect(container.locator('.cg-events tbody tr')).toHaveCount(events.length);
  if (!events.length) { await expect(container.getByText('当前条件下没有事件')).toBeVisible(); return; }
  await expect(container.locator('.cg-events thead th')).toHaveText(columns);
  for (const [index, event] of events.entries()) {
    expect(Object.keys(event).sort()).toEqual(['event_id', 'kind', 'scope', 'tokens', 'topic', 'ts']);
    await expect(container.locator('.cg-events tbody tr').nth(index).locator('td')).toHaveText([event.event_id, event.kind, event.topic || '未标注', formatNumber(event.tokens), event.scope || '未标注', formatDate(event.ts)]);
  }
}

test('认知使用既有外壳、四个路由、刷新和键盘页签', async ({ page }) => {
  await page.goto(go('/attention?fixture=1'));
  await page.getByRole('navigation', { name: '主导航' }).getByRole('button', { name: '认知', exact: true }).click();
  await expect(page).toHaveURL(/\/cognition\/today/);
  await expect(page.getByRole('heading', { name: '认知工作台' })).toBeVisible();
  await expect(page.getByText('当前为示例模式')).not.toBeVisible();
  await expect(page.getByRole('button', { name: /添加资料/ })).not.toBeVisible();
  for (const [name, route] of [['今日', 'today'], ['焦点', 'focus'], ['来源', 'agents'], ['证据', 'evidence']]) {
    await page.getByRole('tab', { name, exact: true }).click();
    await expect(page).toHaveURL(new RegExp(`/cognition/${route}`));
    await page.reload();
    await expect(page.getByRole('tab', { name, exact: true })).toHaveAttribute('aria-selected', 'true');
  }
  await page.getByRole('tab', { name: '证据', exact: true }).press('Home');
  await expect(page).toHaveURL(/\/cognition\/today/);
  await expect(page.getByRole('tab', { name: '今日', exact: true })).toBeFocused();
  await page.getByRole('tab', { name: '今日', exact: true }).press('ArrowRight');
  await expect(page).toHaveURL(/\/cognition\/focus/);
  await expect(page.getByRole('tab', { name: '焦点', exact: true })).toBeFocused();
});

test('今日完整保留实际日期、unknown日环比、Sleep、七日计数与来源过滤', async ({ page }) => {
  const { requests, writes } = observe(page);
  const today = await read<CognitionToday>('/today'), filters = await read<CognitionFilters>('/filters');
  expect(today.focus.length).toBeGreaterThan(0);
  expect(today.daily_counts).toHaveLength(7);
  expect(today.focus.some(topic => topic.attention_delta === null)).toBe(true);
  await page.goto(go('/cognition/today'));
  await expect(page.getByText(today.day, { exact: true })).toBeVisible();
  await expect(page.getByTestId('snapshot-time')).toBeVisible();
  for (const topic of today.focus) {
    await expect(page.getByTestId(`today-a-${topic.topic}`)).toHaveText(formatNumber(topic.attention));
    if (topic.attention_delta === null) await expect(page.getByTestId(`today-delta-${topic.topic}`)).toContainText('日环比暂无');
  }
  await expect(page.getByTestId('sleep-status')).toHaveText({ completed: '已完成', running: '运行中', not_produced: '尚未产出' }[today.sleep.status]);
  for (const day of today.daily_counts) await expect(page.getByTestId(`daily-${day.day}`)).toHaveText(formatNumber(day.count));
  const source = filters.sources[0];
  expect(source).toBeTruthy();
  const response = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/cognition/today') && new URL(response.url()).searchParams.get('source') === source);
  await page.getByLabel('来源', { exact: true }).selectOption(source);
  const filtered = await response;
  expect(filtered.status()).toBe(200);
  const selected = await filtered.json() as CognitionToday;
  await assertEvents(page, selected.activity);
  expect(requests.some(url => url.searchParams.get('source') === source)).toBe(true);
  expect(writes).toEqual([]);
});

test('焦点按实际A排序、显示多条确证轨迹，抽屉完整分页且恢复焦点', async ({ page }) => {
  const { requests, writes } = observe(page);
  const focus = await read<CognitionFocus>('/focus');
  expect(focus.topics.length).toBeGreaterThanOrEqual(2);
  expect(focus.series.filter(line => line.points.length).length).toBeGreaterThanOrEqual(2);
  const ordered = [...focus.topics].sort((a, b) => b.attention - a.attention);
  await page.goto(go('/cognition/focus'));
  await expect(page.getByRole('img', { name: '多主题注意力历史与实验轨迹' })).toBeVisible();
  await expect(page.locator('.cg-topic-row strong')).toHaveText(ordered.map(topic => topic.topic));
  const lines = focus.series.filter(line => line.points.length);
  await expect(page.locator('.cg-legend>span')).toHaveCount(lines.length);
  await page.getByRole('button', { name: lines[1].topic, exact: true }).click();
  await expect(page.locator('.cg-legend>span')).toHaveCount(lines.length - 1);
  const candidates = await Promise.all(ordered.map(topic => read<CognitionEvidence>(`/topics/evidence?topic=${encodeURIComponent(topic.topic)}&page_size=20`)));
  const evidence = candidates.find(item => item.total > 20);
  expect(evidence, '需要实际同一主题超过20条完整证据').toBeDefined();
  expect(evidence!.total).toBeGreaterThan(20);
  const trigger = page.locator('.cg-topic-row').filter({ has: page.getByTestId(`focus-a-${evidence!.topic}`) });
  await trigger.click();
  const dialog = page.getByRole('dialog');
  await expect(dialog.getByRole('heading', { name: '按日与类型汇总' })).toBeVisible();
  await assertEvents(page, evidence!.events, dialog);
  await expect(dialog.getByTestId('events-total')).toHaveText(formatNumber(evidence!.total));
  await dialog.getByRole('button', { name: '下一页' }).click();
  const second = await read<CognitionEvidence>(`/topics/evidence?topic=${encodeURIComponent(evidence!.topic)}&page_size=20&page=2`);
  await assertEvents(page, second.events, dialog);
  await expect.poll(() => requests.some(url => url.pathname.endsWith('/topics/evidence') && url.searchParams.get('page') === '2' && url.searchParams.get('topic') === evidence!.topic)).toBe(true);
  await dialog.getByRole('button', { name: '关闭', exact: true }).press('Shift+Tab');
  await expect(dialog.getByRole('button', { name: second.total > 40 ? '下一页' : '上一页', exact: true })).toBeFocused();
  await page.keyboard.press('Escape');
  await expect(dialog).not.toBeVisible();
  await expect(trigger).toBeFocused();
  expect(writes).toEqual([]);
});

test('来源完整保留实际计数、来源名称与六字段投影', async ({ page }) => {
  const agents = await read<CognitionAgents>('/agents');
  expect(agents.groups.length).toBeGreaterThan(0);
  await page.goto(go('/cognition/agents'));
  for (const group of agents.groups) {
    await expect(page.getByRole('heading', { name: group.source, exact: true })).toBeVisible();
    for (const metric of ['today', 'days7', 'days30'] as const) await expect(page.getByTestId(`agent-${group.source}-${metric}`)).toHaveText(formatNumber(group[metric]));
  }
  for (const table of await page.locator('.cg-events').all()) await expect(table.locator('thead th')).toHaveText(columns);
  await expect(page.locator('.cg-page')).not.toContainText('人类行为');
});

test('证据完整保留过滤、日期、分页重置、每页条数和空日期', async ({ page }) => {
  const { requests, writes } = observe(page);
  const filters = await read<CognitionFilters>('/filters'), initial = await read<CognitionEvents>('/events?page_size=20');
  expect(initial.total).toBeGreaterThan(20);
  expect(filters.topics.length).toBeGreaterThan(0);
  expect(filters.kinds.length).toBeGreaterThan(0);
  expect(filters.sources.length).toBeGreaterThan(0);
  await page.goto(go('/cognition/evidence'));
  await expect(page.getByTestId('events-total')).toHaveText(formatNumber(initial.total));
  await assertEvents(page, initial.events);
  await page.getByRole('button', { name: '下一页' }).click();
  await expect(page.locator('.cg-pagination')).toContainText('第 2 /');
  const topic = filters.topics[0], kind = filters.kinds[0], source = filters.sources[0];
  await page.getByLabel('主题', { exact: true }).selectOption(topic);
  await expect(page.locator('.cg-pagination')).toContainText('第 1 /');
  await page.getByLabel('类型', { exact: true }).selectOption(kind);
  await page.getByLabel('来源', { exact: true }).selectOption(source);
  const day = initial.events[0].ts.slice(0, 10);
  await page.getByLabel('开始日期', { exact: true }).fill(day);
  await page.getByLabel('结束日期', { exact: true }).fill(day);
  await expect.poll(() => requests.some(url => ['topic', 'kind', 'source', 'start', 'end'].every(key => url.searchParams.has(key)) && url.searchParams.get('page') === '1')).toBe(true);
  const params = new URLSearchParams({ topic, kind, source, start: day, end: day, page_size: '20' });
  const selected = await read<CognitionEvents>(`/events?${params}`);
  await expect(page.getByTestId('events-total')).toHaveText(formatNumber(selected.total));
  await assertEvents(page, selected.events);
  await page.getByLabel('每页条数', { exact: true }).selectOption('50');
  await page.getByRole('button', { name: '清除筛选' }).click();
  await assertEvents(page, (await read<CognitionEvents>('/events?page_size=50')).events);
  await page.getByLabel('开始日期', { exact: true }).fill('1900-01-01');
  await page.getByLabel('结束日期', { exact: true }).fill('1900-01-01');
  await expect(page.getByTestId('events-total')).toHaveText('0');
  await expect(page.getByText('当前条件下没有事件')).toBeVisible();
  expect(writes).toEqual([]);
});

test('真实缺少快照时显示错误，人工重试重新读取且没有示例回退', async ({ page }) => {
  test.setTimeout(120_000);
  const unavailable = await startS1Server({ browser: true, env: ({ temporary }) => ({ LOCALAPPDATA: temporary }) });
  try {
    const { requests, writes } = observe(page);
    const response = page.waitForResponse(response => response.url().endsWith('/cognition/health'));
    await page.goto(`${unavailable.webURL}/cognition/today?fixture=1`);
    expect((await response).status()).toBe(503);
    await expect(page.locator('.cg-page').getByRole('alert').first()).toBeVisible();
    await expect(page.locator('.cg-focus-card')).toHaveCount(0);
    await expect(page.getByText('示例论文')).not.toBeVisible();
    const before = requests.length;
    await page.getByRole('button', { name: '↻ 重新读取', exact: true }).click();
    await expect.poll(() => requests.length).toBeGreaterThan(before);
    expect(writes).toEqual([]);
  } finally { await unavailable.close(); }
});
