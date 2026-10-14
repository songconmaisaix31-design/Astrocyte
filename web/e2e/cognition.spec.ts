import { test, expect, type Page } from '@playwright/test';

// Explicit contract fixtures belong only to the test harness; the product always reads the API.
const topics = [{ topic: '测试主题甲', attention: 3.25, momentum: 0.4, confidence: 0.8 }, { topic: '测试主题乙', attention: 1.5, momentum: -0.1, confidence: 0.6 }];
const events = Array.from({ length: 45 }, (_, index) => ({ event_id: `contract-event-${String(index).padStart(3, '0')}`, kind: index % 2 ? 'code_created' : 'topic_activated', topic: index % 3 ? topics[0].topic : topics[1].topic, tokens: index % 4 ? index * 7 : null, scope: 'project', ts: `2026-10-12T${String(15 - Math.floor(index / 4)).padStart(2, '0')}:${String(59 - index).padStart(2, '0')}:00Z` }));
async function setup(page: Page) {
  const requests: URL[] = [];
  await page.route('**/api/v1/cognition/**', async route => {
    const request = route.request(); expect(request.method()).toBe('GET');
    const url = new URL(request.url()); requests.push(url);
    let response: unknown;
    const suffix = url.pathname.slice('/api/v1/cognition'.length);
    if (suffix === '/health') response = { status: 'ok', snapshot_at: '2026-10-13T08:01:00Z', timezone: 'Asia/Shanghai' };
    else if (suffix === '/filters') response = { topics: topics.map(topic => topic.topic), kinds: ['topic_activated', 'code_created'], sources: ['codex', 'manual'] };
    else if (suffix === '/today') response = { day: '2026-10-13', focus: topics.map(topic => ({ ...topic, attention_delta: null })), activity: [], sleep: { day: '2026-10-12', status: 'not_produced', completed_at: null }, daily_counts: Array.from({ length: 7 }, (_, index) => ({ day: `2026-10-${String(7 + index).padStart(2, '0')}`, count: index === 6 ? 0 : index * 5 })), sources: [{ source: 'codex', count: 0 }, { source: 'manual', count: 0 }] };
    else if (suffix === '/focus') response = { topics: [...topics].reverse(), series: topics.map(topic => ({ topic: topic.topic, points: [{ ts: '2026-10-11T03:00:00Z', attention: topic.attention - 1 }, { ts: '2026-10-12T03:00:00Z', attention: topic.attention }] })) };
    else if (suffix === '/agents') response = { groups: [{ source: 'manual', today: 0, days7: 45, days30: 45, recent: events.slice(0, 3) }] };
    else {
      const topic = url.searchParams.get('topic'), kind = url.searchParams.get('kind'), start = url.searchParams.get('start'), end = url.searchParams.get('end');
      const rows = events.filter(event => (!topic || event.topic === topic) && (!kind || event.kind === kind) && (!start || start <= '2026-10-12') && (!end || end >= '2026-10-12'));
      const size = Number(url.searchParams.get('page_size') || 50), pageNumber = Number(url.searchParams.get('page') || 1);
      response = { events: rows.slice((pageNumber - 1) * size, pageNumber * size).map(event => ({ ...event, raw: 'contract-extra-raw-must-not-render', source: 'extra-event-source-must-not-render' })), total: rows.length, page: pageNumber, page_size: size, ...(suffix === '/topics/evidence' ? { topic, aggregates: [{ day: '2026-10-12', kind: 'topic_activated', count: rows.filter(event => event.kind === 'topic_activated').length }, { day: '2026-10-12', kind: 'code_created', count: rows.filter(event => event.kind === 'code_created').length }] } : {}) };
    }
    await route.fulfill({ json: response });
  });
  return requests;
}
test('cognition enters the existing shell, navigates four routes, survives reload and keyboard tabs', async ({ page }) => {
  await setup(page);
  await page.goto('/attention?fixture=1');
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
test('today preserves the empty current day, unknown delta, actual Sleep and source GET filter', async ({ page }) => {
  const requests = await setup(page);
  await page.goto('/cognition/today');
  await expect(page.getByText('2026-10-13', { exact: true })).toBeVisible();
  await expect(page.getByTestId('snapshot-time')).toContainText('2026/10/13');
  await expect(page.getByTestId('today-a-测试主题甲')).toHaveText('3.25');
  await expect(page.getByTestId('today-delta-测试主题甲')).toContainText('日环比暂无');
  await expect(page.getByTestId('sleep-status')).toHaveText('尚未产出');
  await expect(page.getByTestId('daily-2026-10-13')).toHaveText('0');
  await expect(page.getByText('当前条件下没有事件')).toBeVisible();
  await expect(page.locator('.cg-events')).toHaveCount(0);
  await page.getByLabel('来源', { exact: true }).selectOption('manual');
  await expect.poll(() => requests.some(url => url.pathname.endsWith('/today') && url.searchParams.get('source') === 'manual')).toBe(true);
});
test('focus sorts A, shows multiple measured lines, and pages full evidence drawer with focus restoration', async ({ page }) => {
  const requests = await setup(page);
  await page.goto('/cognition/focus');
  await expect(page.getByRole('img', { name: '多主题注意力历史与实验轨迹' })).toBeVisible();
  await expect(page.locator('.cg-topic-row').first()).toContainText('测试主题甲');
  await expect(page.locator('.cg-legend>span')).toHaveCount(2);
  await page.getByRole('button', { name: '测试主题乙', exact: true }).click();
  await expect(page.locator('.cg-legend>span')).toHaveCount(1);
  const trigger = page.locator('.cg-topic-row').first();
  await trigger.click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible();
  await expect(dialog.getByRole('heading', { name: '按日与类型汇总' })).toBeVisible();
  await expect(dialog.locator('.cg-events thead th')).toHaveCount(6);
  await expect(dialog.locator('.cg-events tbody tr')).toHaveCount(20);
  await expect(dialog.getByTestId('events-total')).toHaveText('30');
  await dialog.getByRole('button', { name: '下一页' }).click();
  await expect(dialog.locator('.cg-events tbody tr')).toHaveCount(10);
  await expect.poll(() => requests.some(url => url.pathname.endsWith('/topics/evidence') && url.searchParams.get('page') === '2' && url.searchParams.get('topic') === '测试主题甲')).toBe(true);
  await expect(dialog).not.toContainText('contract-extra-raw-must-not-render');
  await expect(dialog).not.toContainText('extra-event-source-must-not-render');
  await expect(dialog.getByRole('button', { name: '下一页' })).toBeDisabled();
  await dialog.getByRole('button', { name: '关闭', exact: true }).press('Shift+Tab');
  await expect(dialog.getByRole('button', { name: '上一页' })).toBeFocused();
  await page.keyboard.press('Escape');
  await expect(dialog).not.toBeVisible();
  await expect(trigger).toBeFocused();
});
test('agents keeps source labels, real count meanings and only six event fields', async ({ page }) => {
  await setup(page); await page.goto('/cognition/agents');
  await expect(page.getByRole('heading', { name: 'manual', exact: true })).toBeVisible();
  await expect(page.getByTestId('agent-manual-today')).toHaveText('0');
  await expect(page.getByTestId('agent-manual-days7')).toHaveText('45');
  await expect(page.locator('.cg-events thead th')).toHaveCount(6);
  await expect(page.locator('.cg-page')).not.toContainText('人类行为');
});
test('evidence applies topic kind source inclusive dates, resets paging, and preserves six-column projection', async ({ page }) => {
  const requests = await setup(page); await page.goto('/cognition/evidence');
  await expect(page.getByTestId('events-total')).toHaveText('45');
  await expect(page.locator('.cg-events tbody tr')).toHaveCount(20);
  await page.getByRole('button', { name: '下一页' }).click();
  await expect(page.locator('.cg-pagination')).toContainText('第 2 / 3 页');
  await page.getByLabel('主题', { exact: true }).selectOption('测试主题甲');
  await expect(page.locator('.cg-pagination')).toContainText('第 1 / 2 页');
  await page.getByLabel('类型', { exact: true }).selectOption('code_created');
  await page.getByLabel('来源', { exact: true }).selectOption('manual');
  await page.getByLabel('开始日期', { exact: true }).fill('2026-10-12');
  await page.getByLabel('结束日期', { exact: true }).fill('2026-10-12');
  await expect(page.getByTestId('events-total')).toHaveText('15');
  await expect.poll(() => requests.some(url => url.pathname.endsWith('/events') && ['topic', 'kind', 'source', 'start', 'end'].every(key => url.searchParams.has(key)) && url.searchParams.get('page') === '1')).toBe(true);
  await expect(page.locator('.cg-events thead th')).toHaveText(['事件 ID', '类型', '主题', 'Token 数', '范围', '时间（上海）']);
  await expect(page.locator('.cg-page')).not.toContainText('contract-extra-raw-must-not-render');
  await expect(page.locator('.cg-page')).not.toContainText('extra-event-source-must-not-render');
  await page.getByLabel('每页条数', { exact: true }).selectOption('50');
  await page.getByRole('button', { name: '清除筛选' }).click();
  await expect(page.locator('.cg-events tbody tr')).toHaveCount(45);
  await page.getByLabel('开始日期', { exact: true }).fill('2026-10-13');
  await page.getByLabel('结束日期', { exact: true }).fill('2026-10-13');
  await expect(page.getByTestId('events-total')).toHaveText('0');
  await expect(page.getByText('当前条件下没有事件')).toBeVisible();
});
test('an unavailable snapshot stays an explicit error without fixture fallback, and retry rereads', async ({ page }) => {
  let attempts = 0;
  await page.route('**/api/v1/cognition/**', async route => { attempts++; await route.fulfill({ status: 503, json: { schema_version: 1, error: { code: 'snapshot_unavailable', message: '工作台快照不可用', retryable: true, required_action: '请检查本机快照', request_id: 'contract-error' } } }); });
  await page.goto('/cognition/today?fixture=1');
  await expect(page.locator('.cg-page').getByRole('alert').first()).toBeVisible();
  await expect(page.locator('.cg-focus-card')).toHaveCount(0);
  await expect(page.getByText('示例论文')).not.toBeVisible();
  const before = attempts;
  await page.getByRole('button', { name: '重新读取', exact: false }).click();
  await expect.poll(() => attempts).toBeGreaterThan(before);
});
