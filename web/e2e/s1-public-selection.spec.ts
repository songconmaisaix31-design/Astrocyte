/// <reference types="node" />
import { test, expect } from '@playwright/test';
import process from 'node:process';
import { resolve, join } from 'node:path';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI, waitJob } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const approvedStore = 'C:/Users/DW/AppData/Local/Temp/astrocyte-s1-4pmMWO';
const originalURLJob = '26MPWE7S3S4FLEUE62BZLKC54T';

test('real metadata recommendation precedes human selection and reuses the original video body without another extraction', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_PUBLIC_SELECTION !== '1', 'Requires the coordinator-assigned single model/browser slot and existing actual public metadata.');
  test.setTimeout(2_100_000);
  const path = process.env.ASTROCYTE_S1_REUSE_OWNED_TEMP;
  const ownedRoot = process.env.ASTROCYTE_S1_REUSE_APPROVED_ROOT;
  expect(path && ownedRoot && resolve(path) === resolve(approvedStore) && resolve(ownedRoot) === resolve(approvedStore)).toBe(true);
  // Any dedupe regression fails without downloading: summarize remains disabled.
  const server = await startS1Server({ browser: true, reuseOwnedTemporary: { path: path!, ownedRoot: ownedRoot! } });
  let api = await humanAPI(server.apiURL);
  const read = async <T,>(endpoint: string) => await api.get(endpoint) as T;
  let recommendationID: string | undefined;
  try {
    const original = await read<S['JobV1']>(`/jobs/${originalURLJob}`);
    expect(original.status).toBe('succeeded');
    const originalDetail = await read<S['MaterialDetailV1']>(`/materials/${original.material_id!}`);
    const originalBody = await read<S['ContentV1']>(`/materials/${original.material_id!}/revisions/${originalDetail.material.current_revision}/content`);
    const originalJobs = server.query("SELECT id, data, payload FROM attention_jobs WHERE json_extract(data,'$.kind')!='source_sync' ORDER BY id");
    const originalRevisions = server.query('SELECT material_id, revision, data FROM attention_material_revisions ORDER BY material_id, revision');
    const originalMaterials = server.query('SELECT id, data FROM attention_materials ORDER BY id');
    const originals = new Set(originalJobs.map(row => String(row.id)));
    const sources = await read<S['TrackingSourceListV1']>('/tracking-sources');
    const source = sources.items.find(item => item.external_id === '2356677875');
    expect(source, 'Run the actual metadata browser test first.').toBeTruthy();
    let listing = await read<S['TrackingSourceResultV1']>(`/tracking-sources/${source!.id}`);
    await expect.poll(async () => { listing = await read<S['TrackingSourceResultV1']>(`/tracking-sources/${source!.id}`); return listing.source.status; }, { timeout: 90_000 }).toBe('succeeded');
    const selectedItem = listing.items.find(item => item.metadata.locator.includes('BV1PReT6EEqR'));
    expect(selectedItem).toBeTruthy();
    expect(selectedItem!.metadata.unavailable_reason).toBe('');
    expect(selectedItem!.stale).toBe(false);
    expect(selectedItem!.recommendation?.status).not.toBe('unknown');
    expect(selectedItem!.recommendation?.error?.code).not.toBe('delivery_unknown');
    const projects = await read<S['LocalProjectListV1']>('/local-projects');
    const project = projects.items.find(item => resolve(item.root) === resolve(join(approvedStore, 'explicit-public-material-project')));
    expect(project?.settings.external_model_cli).toBe('codex');
    expect(project?.settings.allowed_actions).toEqual(expect.arrayContaining(['start', 'stop']));
    // Explicit no-model native connection verification after the cold restart.
    await page.goto(`${server.webURL}/workspace`);
    const panel = page.locator('section').filter({ has: page.getByRole('heading', { name: '本地 Agent 与项目', exact: true }) }).first();
    await panel.getByRole('button').filter({ has: page.getByRole('heading', { name: project!.name, exact: true }) }).click();
    const native = panel.getByRole('region', { name: '项目原生 Agent 操作' });
    await native.getByLabel('项目操作客户端', { exact: true }).selectOption('codex');
    const probe = page.waitForResponse(response => response.url().endsWith('/sessions/probe') && response.request().method() === 'POST');
    await native.getByRole('button', { name: '明确验证原生连接（启动后停止）', exact: true }).click();
    const probeHTTP = await probe;
    expect(probeHTTP.status()).toBe(200);
    expect((await probeHTTP.json() as S['NativeSessionResultV1']).session.stop_confirmed).toBe(true);
    await page.goto(`${server.webURL}/attention`);
    await page.getByLabel('查看公开来源', { exact: true }).selectOption(source!.id);
    const review = page.getByRole('region', { name: '公开来源更新清单' });
    await review.getByLabel('建议使用的项目许可', { exact: true }).selectOption(project!.id);
    const row = review.locator('li').filter({ has: page.locator('a[href*="BV1PReT6EEqR"]') });
    await expect(row).toHaveCount(1);
    if (!selectedItem!.recommendation) {
      await expect(row.getByRole('checkbox')).toBeDisabled();
      const recommendation = page.waitForResponse(response => response.url().endsWith(`/tracking-sources/${source!.id}/recommend`) && response.request().method() === 'POST');
      await row.getByRole('button', { name: '获取此条 Agent 建议', exact: true }).click();
      const recommendedHTTP = await recommendation;
      await testInfo.attach('actual-recommendation-acceptance', { body: await recommendedHTTP.body(), contentType: 'application/json' });
      expect(recommendedHTTP.status()).toBe(200);
      const accepted = await recommendedHTTP.json() as S['TrackingSourceResultV1'];
      expect(accepted.jobs).toHaveLength(1);
      recommendationID = accepted.jobs[0].job_id;
      const queued = await read<S['JobV1']>(`/jobs/${recommendationID}`);
      const terminal = await waitJob(api, recommendationID, 'succeeded', Math.max(1000, Math.min(1_800_000, Date.parse(queued.deadline_at) - Date.now() + 1000)));
      expect(terminal.attempts).toBe(1);
      expect(terminal.delivery_unknown).toBe(false);
      await testInfo.attach('actual-metadata-model-job', { body: JSON.stringify(terminal, null, 2), contentType: 'application/json' });
    } else {
      // A retry of the browser test consumes the same successful recommendation,
      // never replays failed/unknown model work under another request identity.
      expect(selectedItem!.recommendation.status).toBe('succeeded');
    }
    listing = await read<S['TrackingSourceResultV1']>(`/tracking-sources/${source!.id}`);
    const recommendedItem = listing.items.find(item => item.external_id === selectedItem!.external_id)!;
    expect(recommendedItem.recommendation?.status).toBe('succeeded');
    expect(recommendedItem.recommendation?.metadata_revision).toBe(recommendedItem.revision);
    expect(recommendedItem.recommendation?.text.trim()).toBeTruthy();
    await review.getByRole('button', { name: '重载已保存清单', exact: true }).click();
    await expect(row).toContainText(recommendedItem.recommendation!.text);
    if (!recommendedItem.selected) {
      await row.getByRole('checkbox').check();
      const selected = page.waitForResponse(response => response.url().endsWith(`/tracking-sources/${source!.id}/select`) && response.request().method() === 'POST');
      await review.getByRole('button', { name: '获取所选 1 条正文', exact: true }).click();
      const selectionHTTP = await selected;
      await testInfo.attach('actual-human-selected-body-receipt', { body: await selectionHTTP.body(), contentType: 'application/json' });
      expect(selectionHTTP.status()).toBe(200);
      const selection = await selectionHTTP.json() as S['TrackingSourceResultV1'];
      expect(selection.jobs).toHaveLength(1);
      expect(selection.jobs[0].job_id).toBe(originalURLJob);
    }
    const final = await read<S['TrackingSourceResultV1']>(`/tracking-sources/${source!.id}`);
    const chosen = final.items.find(item => item.external_id === selectedItem!.external_id)!;
    expect(chosen.selected).toBe(true);
    expect(chosen.import_job_id).toBe(originalURLJob);
    expect(chosen.material_id).toBe(original.material_id);
    expect(final.items.filter(item => item.selected)).toHaveLength(1);
    expect(final.items.filter(item => item.external_id !== chosen.external_id).every(item => item.recommendation === null && !item.selected && item.material_id === null)).toBe(true);
    const newJobs = server.query('SELECT id, data FROM attention_jobs ORDER BY id').filter(item => !originals.has(String(item.id))).map(item => JSON.parse(String(item.data)) as S['JobV1']);
    expect(newJobs.filter(job => job.kind === 'source_recommendation').length).toBeLessThanOrEqual(1);
    expect(newJobs.filter(job => !['source_sync', 'source_recommendation'].includes(job.kind))).toEqual([]);
    const readOriginalJobRows = () => server.query('SELECT id, data, payload FROM attention_jobs ORDER BY id').filter(row => originals.has(String(row.id)));
    expect(readOriginalJobRows()).toEqual(originalJobs);
    expect(server.query('SELECT material_id, revision, data FROM attention_material_revisions ORDER BY material_id, revision')).toEqual(originalRevisions);
    expect(server.query('SELECT id, data FROM attention_materials ORDER BY id')).toEqual(originalMaterials);
    for (const width of [1280, 1920]) {
      await page.setViewportSize({ width, height: width === 1280 ? 720 : 1080 });
      await row.scrollIntoViewIfNeeded();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      await page.screenshot({ path: testInfo.outputPath(`actual-selected-collection-${width}.png`) });
    }
    await testInfo.attach('actual-selected-source-before-restart', { body: JSON.stringify(final, null, 2), contentType: 'application/json' });
    await server.restart();
    api = await humanAPI(server.apiURL);
    let restarted = await read<S['TrackingSourceResultV1']>(`/tracking-sources/${source!.id}`);
    await expect.poll(async () => { restarted = await read<S['TrackingSourceResultV1']>(`/tracking-sources/${source!.id}`); return restarted.source.status; }, { timeout: 90_000 }).toBe('succeeded');
    expect(restarted.items).toEqual(final.items);
    expect(await read<S['JobV1']>(`/jobs/${originalURLJob}`)).toEqual(original);
    expect(await read<S['ContentV1']>(`/materials/${original.material_id!}/revisions/${originalDetail.material.current_revision}/content`)).toEqual(originalBody);
    expect(readOriginalJobRows()).toEqual(originalJobs);
    expect((await read<S['MissionListV1']>('/missions')).items).toEqual([]);
    await page.reload();
    await page.getByLabel('查看公开来源', { exact: true }).selectOption(source!.id);
    await expect(row.getByRole('button', { name: '查看已入库资料', exact: true })).toBeVisible();
    await testInfo.attach('actual-selected-source-after-restart', { body: JSON.stringify(restarted, null, 2), contentType: 'application/json' });
  } finally {
    try {
      if (recommendationID) await testInfo.attach('actual-final-recommendation', { body: JSON.stringify(await read<S['JobV1']>(`/jobs/${recommendationID}`), null, 2), contentType: 'application/json' });
    } finally { await server.close({ preserveData: true }); }
  }
});
