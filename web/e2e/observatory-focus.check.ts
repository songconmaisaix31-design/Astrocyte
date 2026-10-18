import { test, expect } from '@playwright/test';
import { fileURLToPath } from 'node:url';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
test.describe('公开焦点历史的新版交互', () => {
  let server: Awaited<ReturnType<typeof startS1Server>>;
  let api: Awaited<ReturnType<typeof humanAPI>>;
  let focus: S['WorkbenchFocusV1'];
  test.beforeAll(async () => {
    test.setTimeout(120_000);
    const reader = process.env.ASTROCYTE_TEST_OBSERVATORY_PUBLIC_READER_ROOT;
    if (!reader) throw new Error('需要主控明确交付的只读公开 Q 发布目录');
    server = await startS1Server({ env: { ASTROCYTE_WEB_DIR: fileURLToPath(new URL('../dist/', import.meta.url)), ASTROCYTE_WORKBENCH_READER_ROOT: reader, OPENVIKING_BASE_URL: '', OPENVIKING_API_KEY: '', OPENVIKING_ACCOUNT: '', OPENVIKING_USER: '', OPENVIKING_ROOT_URI: '' } });
    api = await humanAPI(server.apiURL);
    focus = await api.get('/cognition/focus') as S['WorkbenchFocusV1'];
    expect(focus.topics).toHaveLength(2);
    expect(focus.series).toHaveLength(2);
    for (const line of focus.series) expect(line.points).toHaveLength(3);
  });
  test.afterAll(async () => { if (server) await server.close({ preserveData: true }); });

  test('公开焦点的真实指标、历史点、时间轴与证据保持对应', async ({ page }) => {
    const writes: string[] = [];
    page.on('request', request => { if (request.method() !== 'GET' && new URL(request.url()).pathname.startsWith('/api/v1/')) writes.push(request.url()); });
    await page.goto(`${server.apiURL}/cognition/focus`);
    const chart = page.getByRole('img', { name: '多主题注意力历史与实验轨迹' });
    await expect(chart).toBeVisible();
    await expect(chart.locator('circle')).toHaveCount(focus.series.reduce((total, line) => total + line.points.length, 0));
    for (const topic of focus.topics) {
      for (const [prefix, value] of [['a', topic.attention], ['m', topic.momentum], ['c', topic.confidence]] as const) {
        await expect(page.getByTestId(`focus-${prefix}-${topic.topic}`)).toHaveAttribute('data-value', String(value));
      }
    }
    const times = [...new Set(focus.series.flatMap(line => line.points.map(point => point.ts)))].sort((a, b) => Date.parse(a) - Date.parse(b));
    const slider = page.getByRole('slider', { name: '轨迹时间轴' });
    await slider.focus();
    await page.keyboard.press('Home');
    await expect(slider).toHaveValue('0');
    await expect(chart.locator('circle')).toHaveCount(focus.series.flatMap(line => line.points).filter(point => Date.parse(point.ts) <= Date.parse(times[0])).length);
    await page.getByRole('button', { name: '返回全部', exact: true }).click();
    await expect(chart.locator('circle')).toHaveCount(focus.series.flatMap(line => line.points).length);
    const topic = [...focus.topics].sort((a, b) => b.attention - a.attention)[0];
    const evidence = await api.get(`/cognition/topics/evidence?topic=${encodeURIComponent(topic.topic)}&page=1&page_size=20`) as S['WorkbenchEvidenceV1'];
    await page.getByRole('button', { name: '追溯关联证据', exact: true }).click();
    const dialog = page.getByRole('dialog', { name: `${topic.topic} · 主题证据` });
    await expect(dialog).toBeVisible();
    await expect(dialog.locator('.cg-event-id')).toHaveText(evidence.events.map(event => event.event_id));
    await expect(dialog.getByTestId('events-total')).toHaveAttribute('data-value', String(evidence.total));
    await page.keyboard.press('Escape');
    await expect(dialog).not.toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
    expect(writes).toEqual([]);
  });
});
