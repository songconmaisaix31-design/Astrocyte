import { test, expect } from '@playwright/test';
import { copyFile } from 'node:fs/promises';
import { join } from 'node:path';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI, waitJob } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type Detail = components['schemas']['TrackingSourceResultV1'];
// Explicit app/API acceptance only; the host owns the already connected personal browser.
test('approved selected Douyin folder is chosen, saved, deduplicated and retained after cold restart', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_SELECTED_BROWSER !== '1' || testInfo.project.name !== 'chromium-1920', 'Approved selected browser metadata only, once');
  test.setTimeout(240_000);
  const approved = process.env.ASTROCYTE_S1_SELECTED_BROWSER_CONFIG;
  if (!approved) throw new Error('Name the exact root-approved noncredential browser config artifact');
  const server = await startS1Server({ browser: true });
  let primaryError: unknown;
  try {
    await copyFile(approved, join(server.dataDir, 'selected-browser.json'));
    await server.restart();
    let api = await humanAPI(server.apiURL);
    const writes: string[] = [];
    page.on('request', request => { if (request.method() !== 'GET' && request.url().includes('/api/v1/')) writes.push(new URL(request.url()).pathname); });
    await page.goto(`${server.webURL}/attention`);
    await page.getByRole('button', { name: '查看来源更新', exact: true }).click();
    const accounts = page.locator('section').filter({ has: page.getByRole('heading', { name: '账号与更新清单', exact: true }) }).last();
    await accounts.locator('summary').filter({ hasText: /^绑定创作者 \/ 收藏夹$/ }).click();
    await accounts.getByLabel('来源平台', { exact: true }).selectOption('douyin');
    await accounts.getByLabel('追踪内容', { exact: true }).selectOption('favorites');
    await accounts.getByLabel('收藏夹读取范围', { exact: true }).selectOption('browser_selected');
    expect(writes).toEqual([]);
    await expect(accounts.getByLabel('公开收藏夹 ID', { exact: true })).not.toBeVisible();
    await expect(accounts.getByLabel('收藏夹所属账号 ID', { exact: true })).not.toBeVisible();
    const discovery = page.waitForResponse(response => response.url().endsWith('/api/v1/source-collections/discover') && response.request().method() === 'POST');
    await accounts.getByRole('button', { name: '读取已连接浏览器的选定收藏夹', exact: true }).click();
    const discovered = await discovery;
    expect(discovered.status(), await discovered.text()).toBe(200);
    const folders = (await discovered.json() as components['schemas']['SourceCollectionListV1']).items;
    const selected = folders.find(folder => folder.title === '求职');
    if (!selected) throw new Error('The approved 求职 folder was not returned; do not guess another folder');
    expect(selected.access_mode).toBe('browser_selected'); expect(selected.item_count).toBe(3);
    expect(await api.get('/source-collections?platform=douyin&owner_id=self')).toMatchObject({ items: folders });
    await accounts.getByRole('button', { name: /^求职 · 3 条$/ }).click();
    const binding = page.waitForResponse(response => response.url().endsWith('/api/v1/tracking-sources') && response.request().method() === 'POST');
    await accounts.getByRole('button', { name: '绑定此选定浏览器收藏夹', exact: true }).click();
    const bound = await binding;
    expect(bound.status(), await bound.text()).toBe(200);
    const result = await bound.json() as Detail;
    expect(result.source).toMatchObject({ access_mode: 'browser_selected', platform: 'douyin', source_kind: 'favorites', owner_id: selected.owner_id, external_id: selected.external_id });
    const review = accounts.getByRole('region', { name: '来源更新清单', exact: true });
    const sync = async () => {
      await review.getByRole('button', { name: '同步新标题与简介', exact: true }).click();
      const response = page.waitForResponse(response => response.url().endsWith(`/api/v1/tracking-sources/${result.source.id}/sync`) && response.request().method() === 'POST');
      await review.getByRole('button', { name: '确认同步标题清单', exact: true }).click();
      const synced = await response; expect(synced.status(), await synced.text()).toBe(200);
      const receipt = await synced.json() as Detail;
      for (const job of receipt.jobs) await waitJob(api, job.job_id, 'succeeded', 60_000);
      await expect(review).toContainText('已保存标题 3 条');
      return await api.get(`/tracking-sources/${result.source.id}`);
    };
    const first = await sync();
    expect(first.items).toHaveLength(3); expect(first.source.last_error).toBeNull();
    const identities = first.items.map(item => [item.external_id, item.revision]).sort();
    const assertMetadataOnly = (detail: Detail) => { expect(detail.items.map(item => [item.external_id, item.revision]).sort()).toEqual(identities); for (const item of detail.items) { expect(item.selected).toBe(false); expect(item.recommendation).toBeNull(); expect(item.material_id).toBeNull(); expect(item.import_job_id).toBeNull(); } };
    assertMetadataOnly(first); assertMetadataOnly(await sync());
    await expect(review.getByRole('button', { name: '获取本页 Agent 建议', exact: true })).toBeDisabled();
    for (const checkbox of await review.getByRole('checkbox').all()) await expect(checkbox).toBeDisabled();
    await server.restart(); api = await humanAPI(server.apiURL);
    await expect.poll(async () => (await api.get(`/tracking-sources/${result.source.id}`)).source.status, { timeout: 60_000 }).toBe('succeeded');
    assertMetadataOnly(await api.get(`/tracking-sources/${result.source.id}`));
    expect(await api.get('/source-collections?platform=douyin&owner_id=self')).toMatchObject({ items: folders });
    expect((await api.get('/materials')).items).toHaveLength(0);
    const jobs = (await api.get('/jobs')).items;
    expect(jobs.length).toBeGreaterThanOrEqual(2);
    for (const job of jobs) { expect(job.kind).toBe('source_sync'); expect(job.status).toBe('succeeded'); expect(job.attempts).toBe(1); expect(job.delivery_unknown).toBe(false); }
    expect((await api.get('/missions')).items).toHaveLength(0);
    await page.reload(); await page.getByRole('button', { name: '查看来源更新', exact: true }).click();
    await accounts.getByLabel('查看来源', { exact: true }).selectOption(result.source.id);
    await expect(review).toContainText('已保存标题 3 条');
    for (const width of [1920, 1280, 390]) { await page.setViewportSize({ width, height: width === 390 ? 844 : 900 }); expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true); await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' })); await page.screenshot({ path: testInfo.outputPath(`selected-douyin-${width}.png`), fullPage: true, animations: 'disabled' }); }
    await testInfo.attach('selected-folder-metadata-only', { body: JSON.stringify(first, null, 2), contentType: 'application/json' });
    console.log('Selected-folder metadata retained:', server.temporary, '3 items, no body/model/media selection');
  } catch (error) { primaryError = error; throw error; }
  finally { await server.close({ preserveData: true, primaryError }); }
});

