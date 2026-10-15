import { chromium, expect } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { humanAPI } from '../../tests/s1/api.mjs';
import type { components } from '../src/api/schema';

type S = components['schemas'];
const origin = process.env.ASTROCYTE_TEST_PRODUCT_URL;
const output = process.env.ASTROCYTE_VERIFICATION_OUTPUT_DIR;
if (origin !== 'http://127.0.0.1:64505' || !output) throw new Error('需要明确的同一产品入口和独占验证目录');
const directory = resolve(output);
await mkdir(directory, { recursive: true });
const evidence: {
  started_at: string; completed_at?: string; origin: string; stage: string;
  collections: Record<string, number>; pages: { path: string; status: number; headings: number; alerts: number; overflow: boolean }[];
  public_materials: { id: string; kind: string; revision: number }[];
  commands: string[]; errors: string[];
} = { started_at: new Date().toISOString(), origin, stage: 'started', collections: {}, pages: [], public_materials: [], commands: [], errors: [] };
const path = resolve(directory, 'product-core-readonly.json');
await writeFile(path, JSON.stringify(evidence, null, 2), { flag: 'wx' });
const save = () => writeFile(path, JSON.stringify(evidence, null, 2));
const browser = await chromium.launch();
try {
  const api = await humanAPI(origin);
  evidence.stage = 'human_session';
  const { items: materials } = await api.get('/materials') as S['MaterialListV1'];
  evidence.public_materials = materials.filter(item => /^https:\/\/(arxiv\.org|www\.bilibili\.com|github\.com)\//.test(item.source_locator)).map(item => ({ id: item.id, kind: item.kind, revision: item.current_revision }));
  for (const endpoint of ['materials', 'opportunities', 'project-spaces', 'local-projects', 'goals', 'proposals', 'missions']) {
    const result = await api.get(`/${endpoint}`) as { items: unknown[] };
    evidence.collections[endpoint] = result.items.length;
  }
  await save();
  const context = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await context.newPage();
  page.on('pageerror', error => { evidence.errors.push(error.name); });
  page.on('request', request => {
    const url = new URL(request.url());
    if (url.pathname.startsWith('/api/v1/') && request.method() !== 'GET') evidence.commands.push(`${request.method()} ${url.pathname}`);
  });
  for (const route of ['/', '/attention', '/workspace', '/workspace?tab=proposals', '/swarm', '/workspace/works', '/cognition/today']) {
    evidence.stage = `page:${route}`;
    await save();
    const response = await page.goto(`${origin}${route}`);
    expect(response?.status()).toBe(200);
    await expect(page.getByRole('main')).toBeVisible();
    await expect(page.getByRole('main').getByRole('heading').first()).toBeVisible();
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > innerWidth);
    evidence.pages.push({ path: route, status: response!.status(), headings: await page.getByRole('main').getByRole('heading').count(), alerts: await page.getByRole('main').getByRole('alert').count(), overflow });
    await save();
    expect(overflow).toBe(false);
  }
  expect(evidence.commands).toEqual([]);
  expect(evidence.errors).toEqual([]);
  evidence.stage = 'readonly_completed';
  evidence.completed_at = new Date().toISOString();
  await save();
  console.log(JSON.stringify(evidence));
} finally {
  await save();
  await browser.close();
}
