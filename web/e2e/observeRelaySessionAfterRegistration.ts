/// <reference types="node" />
import { chromium, expect } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { relayEvidenceDirectory } from './relayEvidence';

const origin = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_URL;
const registeredAt = process.env.ASTROCYTE_TEST_RELAY_REGISTERED_AT;
if (!origin || !/^http:\/\/127\.0\.0\.1:\d+$/.test(origin) || !registeredAt || !Number.isFinite(Date.parse(registeredAt))) throw new Error('需要H当前URL与原始实际注册时间');
const directory = relayEvidenceDirectory();
const output = resolve(directory, 'late-first-visible.json');
await mkdir(directory, { recursive: true });
await writeFile(output, '', { flag: 'wx' });
const browser = await chromium.launch();
try {
  const viewports = [{ name: 'chromium-1920', width: 1920, height: 1080 }, { name: 'chromium-1280', width: 1280, height: 720 }];
  const contexts = await Promise.all(viewports.map(viewport => browser.newContext({ viewport: { width: viewport.width, height: viewport.height } })));
  const pages = await Promise.all(contexts.map(context => context.newPage()));
  await Promise.all(pages.map(async page => {
    expect((await page.request.get(`${origin}/api/v1/auth/session`)).status()).toBe(200);
    await page.goto(`${origin}/cognition/today`);
  }));
  console.log('F_LATE_DOM_READY: 保留原登记时间，迟到窗口不用于三分钟PASS。');
  const observations = await Promise.all(pages.map(async (page, index) => {
    const region = page.getByRole('region', { name: '本机 Session 派生事件', exact: true });
    await expect(region.locator('tbody tr').first()).toBeVisible({ timeout: 180_000 });
    const firstVisibleAt = new Date().toISOString();
    return { name: viewports[index].name, first_visible_at: firstVisibleAt, event_ids: await region.locator('tbody tr td:first-child').allTextContents(), elapsed_from_registration_ms: Date.parse(firstVisibleAt) - Date.parse(registeredAt) };
  }));
  await writeFile(output, JSON.stringify({ origin, registered_at: registeredAt, late_verification: true, viewports: observations }, null, 2));
  console.log('F_LATE_FIRST_VISIBLE: 实际两viewport首次可见已记录；原三分钟FAIL保留。');
} finally { await browser.close(); }
