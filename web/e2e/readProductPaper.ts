import { chromium, expect } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { humanAPI } from '../../tests/s1/api.mjs';
import { loadPublicPaperSnapshots } from '../../tests/s1/public-inputs.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
const records = process.env.ASTROCYTE_TEST_PRODUCT_RECORDS_DIR;
const originals = process.env.ASTROCYTE_TEST_PUBLIC_PAPER_INPUTS_DIR;
if (origin !== 'http://127.0.0.1:64505' || !output || !records || resolve(records) === resolve(output) || originals !== 'C:/Users/DW/orca/workspaces/Astrocyte/standards-backend/.verification') throw new Error('需要同一产品入口、分开的原记录和输出目录及四件明确获准HTML目录');
const directory = resolve(output);
const original = JSON.parse(await readFile(resolve(records, 'product-paper-human.json'), 'utf8')) as { imports: { source_key: string; status: number; material_id: string; revision: number }[]; distillation_status: number; abstract_fulltext_absent: boolean };
expect(original.imports.map(item => item.status)).toEqual([202, 202, 202]);
expect(original.distillation_status).toBe(200);
expect(original.abstract_fulltext_absent).toBe(true);
const snapshots = await loadPublicPaperSnapshots(originals);
const api = await humanAPI(origin);
const materialID = original.imports[2].material_id;
const detail = await api.get(`/materials/${materialID}`) as S['MaterialDetailV1'];
const candidates = detail.distillations.filter(item => item.question === '人工摘录本次公开论文摘要，核对固定来源与版本');
expect(candidates).toHaveLength(1);
const distillation = candidates[0];
expect(distillation.output_text).toBe(snapshots.secondary.abstract);
expect(distillation.provenance).toMatchObject({ processor: 'human', mode: 'manual' });
expect(distillation.input_refs).toEqual([expect.objectContaining({ material_id: materialID, revision: 1 })]);
const arxiv = await api.get(`/materials/${original.imports[0].material_id}`) as S['MaterialDetailV1'];
expect(arxiv.material.current_revision).toBe(2);
expect(arxiv.revisions.map(item => item.revision).sort()).toEqual([1, 2]);
const database = new DatabaseSync('D:/Astrocyte-agent-assets/20261015/relay/H/shared-product-ad8413/local/astrocyte-workbench/v0.2-access/app-state/state.sqlite', { readOnly: true });
const events: { id: string; type: string; aggregate_id: string }[] = [];
try {
  for (const id of [arxiv.material.id, materialID, distillation.id]) {
    const rows = database.prepare('SELECT id,type,aggregate_id FROM attention_outbox WHERE aggregate_id=?').all(id) as typeof events;
    expect(rows.length).toBeGreaterThan(0);
    for (const event of rows) expect(database.prepare('SELECT aggregate_id FROM attention_outbox WHERE id=?').get(event.id)).toEqual({ aggregate_id: id });
    events.push(...rows);
  }
  expect(events.filter(event => event.aggregate_id === distillation.id).map(event => event.type)).toEqual(['attention.distillation_recorded']);
  expect(database.prepare('SELECT COUNT(*) AS count FROM local_agent_sessions').get()).toEqual({ count: 0 });
} finally { database.close(); }
const browser = await chromium.launch();
const commands: string[] = [];
try {
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  page.on('request', request => {
    if (new URL(request.url()).pathname.startsWith('/api/v1/') && request.method() !== 'GET') commands.push(request.method());
  });
  await page.goto(`${origin}/attention`);
  await page.locator('main li[role="button"]').filter({ has: page.getByText(detail.material.title!, { exact: true }) }).click();
  const dialog = page.getByRole('dialog', { name: '素材详情', exact: true });
  const manual = dialog.getByRole('article').filter({ hasText: distillation.question });
  await expect(manual).toHaveCount(1);
  await expect(manual.locator('pre')).toHaveText(distillation.output_text);
  await expect(manual).toContainText('human');
  const content = await api.get(`/materials/${materialID}/revisions/1/content`) as S['ContentV1'];
  await expect(dialog.getByRole('region', { name: '来源与原文版本', exact: true }).locator('pre')).toHaveText(content.text);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  expect(commands).toEqual([]);
  const result = { material_ids: [arxiv.material.id, materialID], arxiv_versions: [1, 2], manual_distillation_id: distillation.id, provenance: { processor: 'human', mode: 'manual' }, events, native_sessions: 0, commands, completed_at: new Date().toISOString() };
  await writeFile(resolve(directory, 'product-paper-readonly.json'), JSON.stringify(result, null, 2), { flag: 'wx' });
  console.log(JSON.stringify(result));
} finally { await browser.close(); }
