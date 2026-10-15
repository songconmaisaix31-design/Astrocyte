import { chromium, expect } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { humanAPI, waitJob } from '../../tests/s1/api.mjs';
import { loadPublicPaperSnapshots } from '../../tests/s1/public-inputs.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
const originals = process.env.ASTROCYTE_TEST_PUBLIC_PAPER_INPUTS_DIR;
if (origin !== 'http://127.0.0.1:64505' || !output || originals !== 'C:/Users/DW/orca/workspaces/Astrocyte/standards-backend/.verification') throw new Error('需要同一产品、独占输出目录及四件明确获准公开HTML');
const snapshots = await loadPublicPaperSnapshots(originals);
const directory = resolve(output);
await mkdir(directory, { recursive: true });
const evidence: {
  stage: string; started_at: string; completed_at?: string;
  imports: { source_key: string; version?: string; operation_id?: string; status?: number; job_id?: string; material_id?: string; revision?: number; text_length?: number }[];
  abstract_fulltext_absent?: boolean; distillation_id?: string; distillation_status?: number;
  event_ids: string[]; native_sessions?: number;
} = { stage: 'started', started_at: new Date().toISOString(), imports: [], event_ids: [] };
const path = resolve(directory, 'product-paper-human.json');
await writeFile(path, JSON.stringify(evidence, null, 2), { flag: 'wx' });
const save = () => writeFile(path, JSON.stringify(evidence, null, 2));
const browser = await chromium.launch();
try {
  const api = await humanAPI(origin);
  const context = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await context.newPage();
  page.setDefaultTimeout(30_000);
  for (const name of ['original', 'revised', 'secondary'] as const) {
    const snapshot = snapshots[name];
    const raw = JSON.stringify(snapshot);
    const record: typeof evidence.imports[number] = { source_key: snapshot.source_key, version: snapshot.observed_version };
    evidence.imports.push(record);
    evidence.stage = `reviewing_public_${name}`; await save();
    await page.goto(`${origin}/attention`);
    await page.locator('summary').filter({ hasText: /^插件快照复核导入/ }).click();
    const panel = page.getByRole('region', { name: '插件快照复核导入', exact: true });
    await panel.getByLabel('粘贴插件快照 JSON', { exact: true }).fill(raw);
    await panel.getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel.getByRole('heading', { name: snapshot.title, exact: true })).toBeVisible();
    const waiting = page.waitForResponse(response => response.url().endsWith('/materials/imports') && response.request().method() === 'POST');
    evidence.stage = `importing_public_${name}`; await save();
    await panel.getByRole('button', { name: '按快照正文导入（HTML 正文）', exact: true }).click();
    const response = await waiting;
    record.operation_id = response.request().headers()['idempotency-key'];
    record.status = response.status(); await save();
    expect(response.status()).toBe(202);
    record.job_id = (await response.json() as S['ImportJobV1']).job_id; await save();
    const job = await waitJob(api, record.job_id);
    expect(job.delivery_unknown).not.toBe(true);
    record.material_id = job.material_id!; await save();
    const detail = await api.get(`/materials/${record.material_id}`) as S['MaterialDetailV1'];
    record.revision = detail.material.current_revision;
    const content = await api.get(`/materials/${record.material_id}/revisions/${record.revision}/content`) as S['ContentV1'];
    const expectedText = `# ${snapshot.title}\n\nSource: ${snapshot.source_url}\n\n## Abstract\n\n${snapshot.abstract}\n\n## Full text (browser snapshot)\n\n${snapshot.text}`;
    expect(content.text).toBe(expectedText);
    expect(content.provenance.processor).toBe('paper_snapshot');
    expect(content.provenance.source).toBe(snapshot.source_url);
    record.text_length = content.text.length; await save();
    const attachment = await page.request.get(`${origin}/api/v1/materials/${record.material_id}/revisions/${record.revision}/attachments/paper-snapshot.json`);
    expect(attachment.status()).toBe(200);
    expect(await attachment.text()).toBe(raw);
  }
  expect(evidence.imports[0].material_id).toBe(evidence.imports[1].material_id);
  expect(evidence.imports.map(item => item.revision)).toEqual([1, 2, 1]);
  evidence.stage = 'reviewing_actual_abstract_only'; await save();
  await page.goto(`${origin}/attention`);
  await page.locator('summary').filter({ hasText: /^插件快照复核导入/ }).click();
  const panel = page.getByRole('region', { name: '插件快照复核导入', exact: true });
  await panel.getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify(snapshots.abstract));
  await panel.getByRole('button', { name: '复核快照', exact: true }).click();
  await expect(panel.getByText('快照原文可得状态 · 仅摘要/元数据', { exact: true })).toBeVisible();
  await expect(panel.getByRole('button', { name: '按快照正文导入（HTML 正文）', exact: true })).toHaveCount(0);
  evidence.abstract_fulltext_absent = true; await save();
  const materialID = evidence.imports[2].material_id!;
  const detail = await api.get(`/materials/${materialID}`) as S['MaterialDetailV1'];
  await page.goto(`${origin}/attention`);
  await page.locator('main li[role="button"]').filter({ has: page.getByText(detail.material.title!, { exact: true }) }).click();
  const dialog = page.getByRole('dialog', { name: '素材详情', exact: true });
  await dialog.getByRole('button', { name: '继续沉淀', exact: true }).click();
  await dialog.getByLabel('沉淀层次', { exact: true }).selectOption('content');
  await dialog.getByLabel('本轮问题', { exact: true }).fill('人工摘录本次公开论文摘要，核对固定来源与版本');
  expect(snapshots.secondary.abstract.length).toBeGreaterThan(0);
  await dialog.getByLabel('人工整理结果', { exact: true }).fill(snapshots.secondary.abstract);
  evidence.stage = 'saving_manual_abstract_excerpt'; await save();
  const distilling = page.waitForResponse(response => response.url().endsWith('/distillations') && response.request().method() === 'POST');
  await dialog.getByRole('button', { name: '保存人工沉淀', exact: true }).click();
  const distilled = await distilling;
  evidence.distillation_status = distilled.status(); await save();
  expect(distilled.status()).toBe(200);
  const result = await distilled.json() as S['DistillationResultV1'];
  evidence.distillation_id = result.distillation.id; await save();
  expect(result.distillation.provenance).toMatchObject({ processor: 'human', mode: 'manual' });
  expect(result.distillation.output_text).toBe(snapshots.secondary.abstract);
  expect(result.distillation.input_refs).toEqual([expect.objectContaining({ material_id: materialID, revision: 1 })]);
  const reread = await api.get(`/materials/${materialID}`) as S['MaterialDetailV1'];
  expect(reread.distillations.find(item => item.id === result.distillation.id)).toEqual(result.distillation);
  await page.reload();
  await page.locator('main li[role="button"]').filter({ has: page.getByText(detail.material.title!, { exact: true }) }).click();
  await expect(page.getByRole('dialog')).toContainText(snapshots.secondary.abstract);
  const database = new DatabaseSync('D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite', { readOnly: true });
  try {
    for (const id of [evidence.imports[0].material_id!, materialID, result.distillation.id]) {
      const events = database.prepare('SELECT id,aggregate_id FROM attention_outbox WHERE aggregate_id=?').all(id) as { id: string; aggregate_id: string }[];
      expect(events.length).toBeGreaterThan(0);
      for (const event of events) {
        expect(database.prepare('SELECT aggregate_id FROM attention_outbox WHERE id=?').get(event.id)).toEqual({ aggregate_id: id });
        evidence.event_ids.push(event.id);
      }
    }
    const sessions = database.prepare('SELECT COUNT(*) AS count FROM local_agent_sessions').get() as { count: number };
    evidence.native_sessions = sessions.count;
    expect(sessions.count).toBe(0);
  } finally { database.close(); }
  evidence.stage = 'public_paper_and_manual_excerpt_completed'; evidence.completed_at = new Date().toISOString(); await save();
  console.log(JSON.stringify(evidence));
} finally { await save(); await browser.close(); }
