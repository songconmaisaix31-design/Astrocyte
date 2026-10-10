/// <reference types="node" />
import { test, expect } from '@playwright/test';
import process from 'node:process';
import { resolve } from 'node:path';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const owner = '3494358764489275';
const folder = '2356677875';
const approvedStore = 'C:/Users/DW/AppData/Local/Temp/astrocyte-s1-4pmMWO';

// Coordinator-assigned public-network/browser slot. No page.route, model,
// synthetic metadata or media extraction is used. The optional selection phase
// submits at most one new metadata-only model job, reusing the approved store.
test('real selected Bilibili public collection persists metadata, gates selection and deduplicates after restart', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_PUBLIC_COLLECTION !== '1', 'Requires coordinator-assigned public/browser slot.');
  test.setTimeout(process.env.ASTROCYTE_TEST_PUBLIC_SELECTION === '1' ? 2_100_000 : 240_000);
  const reusePath = process.env.ASTROCYTE_S1_REUSE_OWNED_TEMP;
  const ownedRoot = process.env.ASTROCYTE_S1_REUSE_APPROVED_ROOT;
  expect(reusePath && ownedRoot && resolve(reusePath) === resolve(approvedStore) && resolve(ownedRoot) === resolve(approvedStore), 'Use only the exact coordinator-approved previously owned store.').toBe(true);
  // summarize is disabled: selection must reuse the existing actual receipt;
  // a regression cannot accidentally start another download/transcription.
  const server = await startS1Server({ browser: true, reuseOwnedTemporary: { path: reusePath!, ownedRoot: ownedRoot! } });
  let api = await humanAPI(server.apiURL);
  const readSource = async (id: string) => await api.get(`/tracking-sources/${id}`) as S['TrackingSourceResultV1'];
  const originalJobs = server.query("SELECT id, data FROM attention_jobs WHERE json_extract(data,'$.kind')!='source_sync' ORDER BY id");
  const originalMaterials = server.query('SELECT id, data FROM attention_materials ORDER BY id');
  const originalRevisions = server.query('SELECT material_id, revision, data FROM attention_material_revisions ORDER BY material_id, revision');
  const originalSessions = server.query('SELECT id, data FROM local_agent_sessions ORDER BY id');
  let completed = false;
  try {
    await page.goto(`${server.webURL}/attention`);
    const accounts = page.locator('section').filter({ has: page.getByRole('heading', { name: '账号与更新清单', exact: true }) }).first();
    await accounts.getByText('绑定公开创作者 / 收藏夹', { exact: true }).click();
    await accounts.getByLabel('追踪内容', { exact: true }).selectOption('favorites');
    await expect(accounts.getByLabel('收藏夹所属账号 ID', { exact: true })).toHaveValue(owner);
    const collectionsResponse = page.waitForResponse(response => response.url().includes('/api/v1/source-collections?') && response.request().method() === 'GET');
    await accounts.getByRole('button', { name: '查找此账号的公开收藏夹', exact: true }).click();
    const collectionHTTP = await collectionsResponse;
    expect(collectionHTTP.status()).toBe(200);
    const collections = await collectionHTTP.json() as S['SourceCollectionListV1'];
    const chosen = collections.items.find(item => item.external_id === folder);
    expect(chosen, 'The user-selected real public folder must be returned, not replaced with fixtures.').toBeTruthy();
    expect(chosen!.item_count).toBeGreaterThan(0);
    await accounts.getByRole('button', { name: `${chosen!.title || chosen!.external_id} · 内容数量 ${chosen!.item_count}` }).click();
    await expect(accounts.getByLabel('公开收藏夹 ID', { exact: true })).toHaveValue(folder);
    await expect(accounts.getByLabel('公开主页 / 收藏夹链接', { exact: true })).toHaveValue(chosen!.locator);
    const bind = page.waitForResponse(response => response.url().endsWith('/api/v1/tracking-sources') && response.request().method() === 'POST');
    await accounts.getByRole('button', { name: '绑定公开来源', exact: true }).click();
    const bindingHTTP = await bind;
    expect(bindingHTTP.status()).toBe(200);
    const binding = await bindingHTTP.json() as S['TrackingSourceResultV1'];
    const sourceID = binding.source.id;
    const review = accounts.getByRole('region', { name: '公开来源更新清单' });
    await review.getByRole('button', { name: '同步新标题与简介', exact: true }).click();
    await review.getByRole('button', { name: '确认同步标题清单', exact: true }).click();
    let synced = binding;
    await expect.poll(async () => {
      synced = await readSource(sourceID);
      return synced.source.status;
    }, { timeout: 90_000 }).toBe('succeeded');
    expect(synced.source.last_error).toBeNull();
    expect(synced.source.last_success_at).toBeTruthy();
    expect(synced.items.length).toBeGreaterThan(0);
    expect(synced.items.length).toBeLessThanOrEqual(100);
    expect(synced.items.some(item => item.metadata.locator.includes('BV1PReT6EEqR'))).toBe(true);
    await review.getByRole('button', { name: '重载已保存清单', exact: true }).click();
    await expect(review.locator('li')).toHaveCount(synced.items.length);
    for (const [index, item] of synced.items.entries()) if (item.recommendation === null) await expect(review.locator('li').nth(index).getByRole('checkbox')).toBeDisabled();
    await expect(review.getByRole('button', { name: '获取所选 0 条正文', exact: true })).toBeDisabled();
    await expect(review.getByRole('button', { name: '获取此条 Agent 建议', exact: true }).first()).toBeDisabled();
    const beforeRows = server.query('SELECT external_id, revision, data FROM attention_source_items WHERE source_id=? ORDER BY external_id, revision', sourceID);
    expect(beforeRows.length).toBe(synced.items.length);
    const duplicate = page.waitForResponse(response => response.url().endsWith('/api/v1/tracking-sources') && response.request().method() === 'POST');
    await accounts.getByRole('button', { name: '绑定公开来源', exact: true }).click();
    const repeated = await (await duplicate).json() as S['TrackingSourceResultV1'];
    expect(repeated.source.id).toBe(sourceID);
    expect(server.query('SELECT COUNT(*) AS count FROM attention_tracking_sources WHERE external_id=?', folder)[0].count).toBe(1);
    await testInfo.attach('actual-public-collections', { body: JSON.stringify(collections, null, 2), contentType: 'application/json' });
    await testInfo.attach('actual-public-metadata', { body: JSON.stringify(synced, null, 2), contentType: 'application/json' });
    await server.restart();
    api = await humanAPI(server.apiURL);
    let restarted = await readSource(sourceID);
    await expect.poll(async () => { restarted = await readSource(sourceID); return restarted.source.status; }, { timeout: 90_000 }).toBe('succeeded');
    expect(restarted.source.last_error).toBeNull();
    expect(restarted.items).toEqual(synced.items);
    expect(server.query('SELECT external_id, revision, data FROM attention_source_items WHERE source_id=? ORDER BY external_id, revision', sourceID)).toEqual(beforeRows);
    expect(server.query('SELECT id, data FROM attention_materials ORDER BY id')).toEqual(originalMaterials);
    expect(server.query('SELECT material_id, revision, data FROM attention_material_revisions ORDER BY material_id, revision')).toEqual(originalRevisions);
    expect(server.query("SELECT id, data FROM attention_jobs WHERE json_extract(data,'$.kind')!='source_sync' ORDER BY id")).toEqual(originalJobs);
    expect(server.query('SELECT id, data FROM local_agent_sessions ORDER BY id')).toEqual(originalSessions);
    await page.reload();
    await accounts.getByLabel('查看公开来源', { exact: true }).selectOption(sourceID);
    await expect(review.locator('li')).toHaveCount(synced.items.length);
    for (const width of [1280, 1920]) {
      await page.setViewportSize({ width, height: width === 1280 ? 720 : 1080 });
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      await page.screenshot({ path: testInfo.outputPath(`actual-public-collection-${width}.png`) });
    }
    await testInfo.attach('actual-restarted-source', { body: JSON.stringify(restarted, null, 2), contentType: 'application/json' });
    completed = true;
  } finally {
    await testInfo.attach('actual-public-store', { body: JSON.stringify({ temporary: server.temporary, dataDir: server.dataDir, completed }), contentType: 'application/json' });
    await server.close({ preserveData: true });
  }
});
