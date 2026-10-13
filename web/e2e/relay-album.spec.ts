import { test, expect } from '@playwright/test';
import { fixtureWorksSnapshot } from '../src/pages/workspace/worksAlbumPresentation';

test('contract-local album displays five truthful metrics and data-driven platforms', async ({ page }) => {
  const snapshot = fixtureWorksSnapshot();
  snapshot.works[0].platform = 'public-new-cli';
  snapshot.platforms[0].platform = 'public-new-cli';
  await page.route('**/api/v1/works', route => route.fulfill({ json: { schema_version: 1, snapshot } }));
  const writes: string[] = [];
  const errors: string[] = [];
  page.on('request', request => { if (request.method() !== 'GET' && request.url().includes('/api/v1/')) writes.push(request.url()); });
  page.on('pageerror', error => errors.push(error.message));
  await page.goto('/workspace/works');
  const album = page.getByRole('region', { name: '作品寻回册', exact: true });
  await expect(album.getByLabel('作品统计').locator('dd')).toHaveText(['3', '3', '3', '未知', '未知']);
  await album.getByLabel('平台', { exact: true }).selectOption('public-new-cli');
  await expect(album.getByRole('button', { name: /^查看作品 / })).toHaveCount(1);
  await album.getByRole('button', { name: '交接手册', exact: true }).click();
  await expect(page.getByRole('dialog')).toContainText(snapshot.works[0].handoff.content);
  await page.keyboard.press('Escape');
  for (const width of [1920, 1280, 390]) {
    await page.setViewportSize({ width, height: 844 });
    await expect(album.getByRole('button', { name: '查看作品 个人知识库' })).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
  }
  expect(writes).toEqual([]);
  expect(errors).toEqual([]);
});
