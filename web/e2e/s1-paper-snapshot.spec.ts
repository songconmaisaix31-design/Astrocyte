/// <reference types="node" />
import { test, expect } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import process from 'node:process';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI, waitJob } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const snapshotDir = 'C:/Users/DW/AppData/Local/Temp/astrocyte-paper-snapshot';

test('real paper_snapshot browser ingest: paste body, source version, dedup, modified body and fresh API', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_PAPER_SNAPSHOT !== '1', 'Requires the controller-assigned real browser slot.');
  test.setTimeout(600_000);
  page.setDefaultTimeout(30_000);

  const fulltext = JSON.parse(await readFile(resolve(snapshotDir, 'fulltext.snapshot.json'), 'utf8'));
  expect(fulltext.schema_version).toBe(1);
  expect(fulltext.content_state).toBe('readable_fulltext');
  expect(fulltext.source_url).toMatch(/^https:\/\//);

  const server = await startS1Server({ browser: true });
  let api = await humanAPI(server.apiURL);
  const errors: string[] = [];
  page.on('pageerror', error => errors.push(error.message));
  let primaryError: unknown;
  console.log(JSON.stringify({ scope: 'task_live_paper_snapshot', directory: server.temporary, api: server.apiURL, web: server.webURL }));

  const read = async <T,>(path: string): Promise<T> => (await api.get(path)) as T;

  async function openPanel() {
    await page.goto(`${server.webURL}/attention`);
    await page.locator('summary').filter({ hasText: /^插件快照复核导入/ }).click();
    return page.locator('section[aria-label="插件快照复核导入"]');
  }

  async function pasteAndImport(panel: Awaited<ReturnType<typeof openPanel>>, snapshot: unknown, button: string) {
    await panel.getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify(snapshot));
    await panel.getByRole('button', { name: '复核快照', exact: true }).click();
    const importBtn = panel.getByRole('button', { name: button, exact: true });
    await expect(importBtn).toBeVisible();
    const accepted = page.waitForResponse(r => r.url().endsWith('/materials/imports') && r.request().method() === 'POST');
    await importBtn.click();
    const response = await accepted;
    expect(response.status()).toBe(202);
    const receipt = (await response.json()) as S['ImportJobV1'];
    return receipt.job_id;
  }

  try {
    // 1. Real full-text snapshot pasted in the browser → paper_snapshot body ingest (no re-fetch).
    const panel = await openPanel();
    const fullJob = await pasteAndImport(panel, fulltext, '按快照正文导入（HTML 正文）');
    const fullJobDone = await waitJob(api, fullJob, 'succeeded', 30_000);
    expect(fullJobDone.material_id).toBeTruthy();
    const fullMaterialId = fullJobDone.material_id!;
    const fullDetail = await read<S['MaterialDetailV1']>(`/materials/${fullMaterialId}`);
    const fullRev = fullDetail.revisions[0];
    expect(fullRev.provenance.processor).toBe('paper_snapshot');
    expect(fullRev.provenance.mode).toMatch(/^browser_snapshot/);
    expect(fullRev.source_key).toBe(fulltext.source_key);
    const fullContent = await read<S['ContentV1']>(`/materials/${fullMaterialId}/revisions/1/content`);
    expect(fullContent.text.length).toBeGreaterThan(10000);
    expect(fullContent.text).toContain(fulltext.title);
    const snapshotAttachment = (fullRev.attachments ?? []).find(a => a.name === 'paper-snapshot.json');
    expect(snapshotAttachment).toBeTruthy();
    const kept = JSON.parse(await readFile(resolve(server.dataDir, 'objects', snapshotAttachment!.object_ref), 'utf8'));
    expect(kept).toEqual(fulltext);

    // 2. Same body re-import dedups to the same material (no new revision).
    const panel2 = await openPanel();
    const dedupJob = await pasteAndImport(panel2, fulltext, '按快照正文导入（HTML 正文）');
    await waitJob(api, dedupJob, 'succeeded', 30_000);
    const dedupDetail = await read<S['MaterialDetailV1']>(`/materials/${fullMaterialId}`);
    expect(dedupDetail.material.id).toBe(fullMaterialId);
    expect(dedupDetail.revisions.length).toBe(1);

    // 3. Same URL with a human-modified body yields a NEW revision (browser user-provided), not a 409.
    const modified = { ...fulltext, text: fulltext.text + '\n\n[Human-provided additional body paragraph for version-change verification.]' };
    const panel3 = await openPanel();
    const modJob = await pasteAndImport(panel3, modified, '按快照正文导入（HTML 正文）');
    await waitJob(api, modJob, 'succeeded', 30_000);
    const modDetail = await read<S['MaterialDetailV1']>(`/materials/${fullMaterialId}`);
    expect(modDetail.revisions.length).toBe(2);
    const modContent = await read<S['ContentV1']>(`/materials/${fullMaterialId}/revisions/2/content`);
    expect(modContent.text).toContain('[Human-provided additional body paragraph');
    expect(modContent.text).toContain(fulltext.title);

    // 4. Fresh API process re-reads the same SQLite/objects.
    await server.restart();
    api = await humanAPI(server.apiURL);
    const refetched = await read<S['MaterialDetailV1']>(`/materials/${fullMaterialId}`);
    expect(refetched.revisions.length).toBe(2);
    const refetchedContent = await read<S['ContentV1']>(`/materials/${fullMaterialId}/revisions/1/content`);
    expect(refetchedContent.content_digest).toBe(fullContent.content_digest);

    // 5. Three sizes + keyboard: the snapshot panel stays usable without overflow.
    await page.goto(`${server.webURL}/attention`);
    const summary = page.locator('summary').filter({ hasText: /^插件快照复核导入/ });
    const sizedPanel = page.locator('section[aria-label="插件快照复核导入"]');
    await summary.click();
    await expect(sizedPanel).toBeVisible();
    for (const width of [1920, 1280, 390]) {
      await page.setViewportSize({ width, height: width === 390 ? 844 : width === 1280 ? 720 : 1080 });
      await expect(sizedPanel).toBeVisible();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      await page.screenshot({ path: testInfo.outputPath(`actual-paper-snapshot-${width}.png`), fullPage: false, animations: 'disabled' });
    }
    // Keyboard: the summary is focusable and Enter toggles it open/closed.
    await summary.focus();
    await page.keyboard.press('Enter');
    await expect(sizedPanel).toBeHidden();
    await page.keyboard.press('Enter');
    await expect(sizedPanel).toBeVisible();
    expect(errors).toEqual([]);

    await writeFile(testInfo.outputPath('actual-fulltext-revision.json'), JSON.stringify(fullRev, null, 2));
    await writeFile(testInfo.outputPath('actual-modified-revision.json'), JSON.stringify(modDetail.revisions[1], null, 2));
  } catch (error) {
    primaryError = error;
    throw error;
  } finally {
    await server.close({ preserveData: true, primaryError });
  }
});
