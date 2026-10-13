import { test, expect, type Locator } from '@playwright/test';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import type { CognitionHealth, CognitionToday, CognitionFocus, CognitionAgents, CognitionEvidence, CognitionFilters } from '../src/api/cognitionTypes';

test('real cognition snapshot: four views, read-only interactions and ten visible independent SQL comparisons', async ({ page }, info) => {
  test.skip(process.env.WORKBENCH_REAL_ACCEPTANCE !== '1', 'Requires the explicitly configured local real snapshot; contract fixtures are not live acceptance');
  const evidenceRoot = join(process.env.LOCALAPPDATA!, 'astrocyte-workbench/v0.1/verification/frontend');
  await mkdir(evidenceRoot, { recursive: true });
  const expected = JSON.parse(await readFile(join(evidenceRoot, '..', 'root-sql-expected-2026-10-14.json'), 'utf8')) as { snapshot_at: string; day: string; values: { route: string; metric: string; topic?: string; source?: string; day?: string; value: number; query: Record<string, string> }[] };
  const writes: string[] = [], errors: string[] = [];
  page.on('request', request => { if (new URL(request.url()).pathname.startsWith('/api/') && request.method() !== 'GET') writes.push(request.method()); });
  page.on('pageerror', error => errors.push(error.name));
  await page.goto('/cognition/today');
  await expect(page.getByTestId('snapshot-time')).toBeVisible();
  const health = await (await page.request.get('/api/v1/cognition/health')).json() as CognitionHealth;
  const today = await (await page.request.get('/api/v1/cognition/today')).json() as CognitionToday;
  expect(Date.parse(health.snapshot_at)).toBe(Date.parse(expected.snapshot_at));
  expect(today.day).toBe(expected.day);
  const filters = await (await page.request.get('/api/v1/cognition/filters')).json() as CognitionFilters;
  const values: { route: string; metric: string; topic?: string; source?: string; day?: string; value: number; text: string; query: Record<string, string> }[] = [];
  const capture = async (locator: Locator, metadata: Omit<typeof values[number], 'value' | 'text'>) => {
    await expect(locator).toBeVisible();
    values.push({ ...metadata, value: Number(await locator.getAttribute('data-value')), text: (await locator.innerText()).trim() });
  };
  const screenshot = async (view: string) => {
    await page.evaluate(() => window.scrollTo(0, 0));
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await page.screenshot({ path: join(evidenceRoot, `${info.project.name}-${view}.png`), fullPage: true });
  };
  // Freeze ten UI values; the baseline was exported independently from the snapshot by root.
  for (const item of today.daily_counts.slice(0, 3)) await capture(page.getByTestId(`daily-${item.day}`), { route: '/cognition/today', metric: 'daily_count', day: item.day, query: {} });
  if (today.activity.length === 0) await expect(page.getByText('当前条件下没有事件')).toBeVisible();
  await screenshot('today');
  if (filters.sources.length) {
    const chosen = filters.sources[0];
    const response = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/cognition/today') && new URL(response.url()).searchParams.get('source') === chosen);
    await page.getByLabel('来源', { exact: true }).selectOption(chosen);
    expect((await response).ok()).toBe(true);
  }
  await page.getByRole('tab', { name: '焦点', exact: true }).click();
  await expect(page.getByRole('img', { name: '多主题注意力历史与实验轨迹' })).toBeVisible();
  const focus = await (await page.request.get('/api/v1/cognition/focus')).json() as CognitionFocus;
  const highest = [...focus.topics].sort((a, b) => b.attention - a.attention)[0];
  expect(highest).toBeDefined();
  for (const [prefix, metric] of [['a', 'attention'], ['m', 'momentum'], ['c', 'confidence']]) await capture(page.getByTestId(`focus-${prefix}-${highest.topic}`), { route: '/cognition/focus', metric, topic: highest.topic, query: {} });
  await expect(page.locator('.cg-topic-row').first()).toContainText(highest.topic);
  await screenshot('focus');
  let selectedTopic = highest.topic;
  for (const topic of focus.topics) {
    const evidence = await (await page.request.get(`/api/v1/cognition/topics/evidence?topic=${encodeURIComponent(topic.topic)}&page_size=20`)).json() as CognitionEvidence;
    if (evidence.total > 20) { selectedTopic = topic.topic; break; }
  }
  await page.locator('.cg-topic-row').filter({ has: page.getByTestId(`focus-a-${selectedTopic}`) }).click();
  const dialog = page.getByRole('dialog');
  await expect(dialog.getByRole('heading', { name: '完整事件证据' })).toBeVisible();
  const drawerTotal = Number((await dialog.getByTestId('events-total').innerText()).replaceAll(',', ''));
  if (drawerTotal > 20) {
    await dialog.getByRole('button', { name: '下一页' }).click();
    await expect(dialog.locator('.cg-pagination')).toContainText('第 2 /');
  } else await expect(dialog.getByRole('button', { name: '下一页' })).toBeDisabled();
  if (drawerTotal) await expect(dialog.locator('.cg-events thead th')).toHaveCount(6);
  await screenshot('drawer');
  await page.keyboard.press('Escape');
  await expect(dialog).not.toBeVisible();
  await page.getByRole('tab', { name: '来源', exact: true }).click();
  const agents = await (await page.request.get('/api/v1/cognition/agents')).json() as CognitionAgents;
  expect(agents.groups).toHaveLength(3);
  for (const group of agents.groups) await capture(page.getByTestId(`agent-${group.source}-days30`), { route: '/cognition/agents', metric: 'days30', source: group.source, query: {} });
  await screenshot('agents');
  await page.getByRole('tab', { name: '证据', exact: true }).click();
  await capture(page.getByTestId('events-total'), { route: '/cognition/evidence', metric: 'total', query: { page: '1', page_size: '20' } });
  expect(values).toHaveLength(10);
  for (const value of values) {
    const sql = expected.values.find(item => ['route', 'metric', 'topic', 'source', 'day'].every(key => item[key as keyof typeof item] === value[key as keyof typeof value]));
    expect(sql, 'independent SQL case exists for every visible value').toBeDefined();
    expect(value.value).toBe(sql!.value);
    expect(value.text).toBe(new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 4 }).format(sql!.value));
  }
  await expect(page.locator('.cg-events thead th')).toHaveCount(6);
  await page.getByRole('button', { name: '下一页' }).click();
  await expect(page.locator('.cg-pagination')).toContainText('第 2 /');
  await page.getByLabel('主题', { exact: true }).selectOption(selectedTopic);
  await expect(page.locator('.cg-pagination')).toContainText('第 1 /');
  if (filters.kinds.length) await page.getByLabel('类型', { exact: true }).selectOption(filters.kinds[0]);
  if (filters.sources.length) await page.getByLabel('来源', { exact: true }).selectOption(filters.sources[0]);
  await page.getByLabel('开始日期', { exact: true }).fill(today.day);
  await page.getByLabel('结束日期', { exact: true }).fill(today.day);
  if (today.activity.length === 0) {
    await expect(page.getByTestId('events-total')).toHaveText('0');
    await expect(page.getByText('当前条件下没有事件')).toBeVisible();
  } else await expect(page.getByTestId('events-total')).toBeVisible();
  await page.getByRole('button', { name: '清除筛选' }).click();
  await expect(page.locator('.cg-events tbody tr')).toHaveCount(20);
  await screenshot('evidence');
  const endHealth = await (await page.request.get('/api/v1/cognition/health')).json() as CognitionHealth;
  expect(endHealth.snapshot_at).toBe(health.snapshot_at);
  expect(writes).toEqual([]); expect(errors).toEqual([]);
  await writeFile(join(evidenceRoot, `ui-values-${info.project.name}.json`), JSON.stringify({ snapshot_at: health.snapshot_at, timezone: health.timezone, day: today.day, values, sql_comparisons: 10, maximum_absolute_error: 0, read_only: true, drawer_page_2: drawerTotal > 20 }, null, 2));
});
