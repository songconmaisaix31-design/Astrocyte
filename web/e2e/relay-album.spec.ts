import { test, expect } from '@playwright/test';
import { realService, repositoryRoot } from './realService';
import type { components } from '../src/api/schema';
import { relayAlbumMetrics } from '../src/pages/workspace/relayAlbumMetrics';

test('实际扫描作品保留五项统计、平台筛选、交接手册与三种尺寸布局', async ({ page }) => {
  test.setTimeout(180_000);
  const { server, api } = await realService(true);
  try {
    await expect.poll(async () => ((await api.get('/local-projects/registered')) as components['schemas']['RegisteredProjectDiscoveryResultV1']).snapshot.projects.length, { timeout: 120_000 }).toBeGreaterThan(0);
    const scanned = await api.write('/works/refresh', { roots: [repositoryRoot] }) as components['schemas']['WorksSnapshotResultV1'];
    expect(scanned.snapshot.works).toHaveLength(1);
    const work = scanned.snapshot.works[0];
    const content = '本轨实际扫描资料与原文版本；原生接力单独取得许可。';
    await api.write(`/works/${work.id}/handoff`, { content }, { method: 'PUT' });
    const snapshot = (await api.get('/works') as components['schemas']['WorksSnapshotResultV1']).snapshot;
    const tasks = await api.get('/relay/tasks') as components['schemas']['RelayTasksV1'];
    const writes: string[] = [], errors: string[] = [];
    page.on('request', request => { if (request.method() !== 'GET' && request.url().includes('/api/v1/')) writes.push(request.url()); });
    page.on('pageerror', error => errors.push(error.name));
    await page.goto(`${server.webURL}/workspace/works`);
    const album = page.getByRole('region', { name: '作品寻回册', exact: true });
    await expect(album.getByLabel('作品统计').locator('dd')).toHaveText(relayAlbumMetrics(snapshot, tasks.dispatched_work_ids).map(metric => metric.value === null ? '未知' : String(metric.value)));
    await album.getByLabel('平台', { exact: true }).selectOption(work.platform);
    await expect(album.getByRole('button', { name: /^查看作品 / })).toHaveCount(1);
    await album.getByRole('button', { name: '交接手册', exact: true }).click();
    await expect(page.getByRole('dialog')).toContainText(content);
    await page.keyboard.press('Escape');
    for (const width of [1920, 1280, 390]) {
      await page.setViewportSize({ width, height: 844 });
      await expect(album.getByRole('button', { name: `查看作品 ${work.name}`, exact: true })).toBeVisible();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    }
    expect(writes).toEqual([]);
    expect(errors).toEqual([]);
  } finally { await server.close(); }
});
