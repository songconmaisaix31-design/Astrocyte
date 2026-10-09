/// <reference types="node" />
/** Injected responses verify UI recovery; the opt-in case uses the real URL adapter. */
import { test, expect } from '@playwright/test';
import { Buffer } from 'node:buffer';
import process from 'node:process';
import type { components } from '../src/api/schema';
import { startS1Server } from '../../tests/s1/server.mjs';
import { summarizeEnvironment } from '../../scripts/summarize.mjs';

const videoURL = 'https://www.bilibili.com/video/BV1PReT6EEqR/';

// Explicit negative live check for the current network gate, not video acceptance.
test('real network-blocked video job shows next step and preserves the URL', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_SUMMARIZE_NETWORK_GUARD !== '1', 'Only run with an explicitly confirmed public-source network block.');
  test.setTimeout(180_000);
  const env = await summarizeEnvironment({ LOCALAPPDATA: process.env.LOCALAPPDATA ?? '', ASTROCYTE_ENABLE_SUMMARIZE: 'true', ASTROCYTE_ENABLE_CODEX_DISTILLATION: 'false' });
  const server = await startS1Server({ browser: true, env });
  const modelRequests: string[] = [];
  page.on('request', request => { if (request.method() === 'POST' && request.url().includes('/distillations/jobs')) modelRequests.push(request.url()); });
  try {
    await page.goto(`${server.webURL}/attention`);
    const add = page.locator('main').getByRole('button', { name: '添加资料', exact: true });
    await add.focus();
    await page.keyboard.press('Enter');
    const dialog = page.getByRole('dialog', { name: '添加资料' });
    await dialog.getByLabel('导入方式', { exact: true }).selectOption('summarize_url');
    await dialog.getByLabel('视频链接', { exact: true }).fill(videoURL);
    await dialog.getByLabel('收藏理由（可选）').fill('核对公开来源获取失败，保留原链接');
    const accepted = page.waitForResponse(response => response.url().endsWith('/materials/imports') && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '导入资料', exact: true }).focus();
    await page.keyboard.press('Enter');
    const receipt = await accepted;
    expect(receipt.status()).toBe(202);
    const job: components['schemas']['ImportJobV1'] = await receipt.json();
    const request = receipt.request();
    expect(request.postDataJSON().adapter).toBe('summarize_url');
    expect(request.postDataJSON()).not.toHaveProperty('export_text');
    await page.keyboard.press('Escape');
    await expect(add).toBeFocused();
    const row = page.locator('li').filter({ has: page.locator('strong').filter({ hasText: job.job_id }) });
    await expect(row).toContainText('正文导入');
    let completed: components['schemas']['JobV1'] | undefined;
    await expect.poll(async () => {
      const response = await page.request.get(`${server.webURL}/api/v1/jobs`);
      const jobs: components['schemas']['JobListV1'] = await response.json();
      completed = jobs.items.find(entry => entry.job_id === job.job_id);
      return completed?.status;
    }, { timeout: 120_000, intervals: [1000] }).toBe('failed');
    expect(completed!.error?.required_action).toBe('configure_public_source_network');
    expect(completed!.material_id).toBeFalsy();
    await page.getByRole('button', { name: '刷新队列', exact: true }).click();
    await expect(row).toContainText('真实 DNS 解析与代理配置');
    await expect(row).toContainText('修复后再明确重试');
    await expect(row).not.toContainText('configure_public_source_network');
    await page.screenshot({ path: testInfo.outputPath('real-network-failure.png'), fullPage: true });
    await testInfo.attach('actual-failed-job', { body: JSON.stringify(completed, null, 2), contentType: 'application/json' });
    // Replaying the accepted original command returns its original receipt and
    // cannot start another extraction; this does not define new-form URL reuse.
    const replay = await page.request.post(request.url(), { headers: request.headers(), data: request.postDataJSON() });
    expect(replay.status()).toBe(202);
    expect(await replay.json()).toEqual(job);
    const jobsAfter: components['schemas']['JobListV1'] = await (await page.request.get(`${server.webURL}/api/v1/jobs`)).json();
    expect(jobsAfter.items).toHaveLength(1);
    expect(jobsAfter.items[0].attempts).toBe(completed!.attempts);
    await add.focus();
    await page.keyboard.press('Enter');
    await expect(dialog.getByLabel('视频链接', { exact: true })).toHaveValue(videoURL);
    await expect(dialog.getByLabel('收藏理由（可选）')).toHaveValue('核对公开来源获取失败，保留原链接');
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
    await page.screenshot({ path: testInfo.outputPath('retained-video-input.png'), fullPage: true });
    expect(modelRequests).toEqual([]);
  } finally { await server.close(); }
});

test('video URL input and in-flight identity survive keyboard close and reopen', async ({ page }) => {
  let release!: () => void;
  const responseReady = new Promise<void>(resolve => { release = resolve; });
  const writes: { key: string; body: components['schemas']['ImportMaterialRequestV1'] }[] = [];
  await page.route('**/api/v1/jobs', route => route.fulfill({ json: { schema_version: 1, items: [], next_cursor: null } }));
  await page.route('**/api/v1/materials/imports', async route => {
    writes.push({ key: route.request().headers()['idempotency-key'], body: route.request().postDataJSON() });
    await responseReady;
    await route.fulfill({ status: 503, json: { schema_version: 1, error: { code: 'evidence_missing', message: '注入：视频需要登录或没有公开字幕', retryable: false, request_id: 'injected-video', required_action: '提供真实 summarize 导出；登录内容范围需先确认' } } });
  });
  await page.goto('/attention');
  const add = page.locator('main').getByRole('button', { name: '添加资料', exact: true });
  await add.focus();
  await page.keyboard.press('Enter');
  const dialog = page.getByRole('dialog', { name: '添加资料' });
  await dialog.getByLabel('导入方式', { exact: true }).selectOption('summarize_url');
  await expect(dialog.getByLabel('summarize 导出内容')).toHaveCount(0);
  await dialog.getByLabel('视频链接', { exact: true }).fill(videoURL);
  await dialog.getByLabel('收藏理由（可选）').fill('保留这份视频输入');
  await dialog.getByRole('button', { name: '导入资料', exact: true }).focus();
  await page.keyboard.press('Enter');
  await expect.poll(() => writes.length).toBe(1);
  await page.keyboard.press('Escape');
  await expect(add).toBeFocused();
  await page.keyboard.press('Enter');
  await expect(dialog.getByLabel('视频链接', { exact: true })).toHaveValue(videoURL);
  await expect(dialog.getByRole('button', { name: '导入资料', exact: true })).toBeDisabled();
  release();
  await expect(dialog.getByRole('alert')).toContainText('提供真实 summarize 导出');
  await dialog.getByRole('button', { name: '导入资料', exact: true }).click();
  await expect.poll(() => writes.length).toBe(2);
  expect(writes[1]).toEqual(writes[0]);
  expect(writes[0].body.adapter).toBe('summarize_url');
  expect(writes[0].body).not.toHaveProperty('export_text');
  expect(writes[0].body).not.toHaveProperty('local_file_ref');
  await dialog.getByLabel('导入方式', { exact: true }).selectOption('summarize');
  await dialog.getByLabel('选择既有导出文件').setInputFiles({ name: 'captured.md', mimeType: 'text/markdown', buffer: Buffer.from('# Existing Markdown\n真实导出入口的UI恢复检查。') });
  await expect(dialog.getByLabel('summarize 导出内容')).toContainText('真实导出入口');
  await page.keyboard.press('Escape');
  await add.click();
  await expect(dialog.getByLabel('summarize 导出内容')).toContainText('真实导出入口');
  await dialog.getByRole('button', { name: '导入资料', exact: true }).click();
  await expect.poll(() => writes.length).toBe(3);
  expect(writes[2].body.adapter).toBe('summarize');
  expect(writes[2].body.export_text).toContain('真实导出入口');
  expect(writes[2].key).not.toBe(writes[0].key);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.screenshot({ path: `test-results/video-import-recovery-${page.viewportSize()?.width}.png`, fullPage: true });
});

test('selected public video URL imports through the real service without an export or model request', async ({ page }, testInfo) => {
  test.skip(process.env.ASTROCYTE_TEST_SUMMARIZE_URL !== '1', 'Requires the published summarize runtime; explicit real URL extraction only.');
  // W0's approved Reader/Service deadline is 1800s; reserve startup/cleanup
  // separately. Run one project/worker so the selected video is extracted once.
  const jobTimeout = 1800_000;
  test.setTimeout(jobTimeout + 120_000);
  const env = await summarizeEnvironment({ LOCALAPPDATA: process.env.LOCALAPPDATA ?? '', ASTROCYTE_ENABLE_SUMMARIZE: 'true', ASTROCYTE_ENABLE_CODEX_DISTILLATION: 'false' });
  const server = await startS1Server({ browser: true, env });
  const modelRequests: string[] = [];
  page.on('request', request => { if (request.method() === 'POST' && request.url().includes('/distillations/jobs')) modelRequests.push(request.url()); });
  try {
    // The default Playwright dev --ephemeral server disables extraction. The
    // helper owns a separate temporary SQLite/object store and strips ASTRO env.
    await page.goto(`${server.webURL}/attention`);
    await page.locator('main').getByRole('button', { name: '添加资料', exact: true }).click();
    const dialog = page.getByRole('dialog', { name: '添加资料' });
    await dialog.getByLabel('导入方式', { exact: true }).selectOption('summarize_url');
    await dialog.getByLabel('视频链接', { exact: true }).fill(videoURL);
    const accepted = page.waitForResponse(response => response.url().endsWith('/materials/imports') && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '导入资料', exact: true }).click();
    const receipt = await accepted;
    expect(receipt.status()).toBe(202);
    const job: components['schemas']['ImportJobV1'] = await receipt.json();
    expect(receipt.request().postDataJSON().adapter).toBe('summarize_url');
    expect(receipt.request().postDataJSON()).not.toHaveProperty('export_text');
    expect(receipt.request().postDataJSON()).not.toHaveProperty('local_file_ref');
    await page.keyboard.press('Escape');
    const row = page.locator('li').filter({ has: page.locator('strong').filter({ hasText: job.job_id }) });
    await expect(row).toContainText('正文导入');
    let completed: components['schemas']['JobV1'] | undefined;
    await expect.poll(async () => {
      const response = await page.request.get(`${server.webURL}/api/v1/jobs`);
      expect(response.ok()).toBe(true);
      const jobs: components['schemas']['JobListV1'] = await response.json();
      completed = jobs.items.find(entry => entry.job_id === job.job_id);
      return completed?.status;
    }, { timeout: jobTimeout, intervals: [2000] }).toMatch(/succeeded|failed|cancelled/);
    await page.getByRole('button', { name: '刷新队列', exact: true }).click();
    await expect(row).toContainText(completed!.status === 'succeeded' ? '已完成' : '失败');
    await page.screenshot({ path: testInfo.outputPath('video-import-real.png'), fullPage: true });
    await testInfo.attach('actual-video-job', { body: JSON.stringify(completed, null, 2), contentType: 'application/json' });
    expect(completed!.status, JSON.stringify(completed!.error)).toBe('succeeded');
    await row.getByRole('button', { name: '查看作业资料' }).click();
    const detail = page.getByRole('dialog', { name: '素材详情' });
    await expect(detail).toContainText('已保存可读文本');
    await expect(detail).toContainText('尚无此版本的 Codex 内容整理结果');
    await expect(detail.getByText('固定来源', { exact: true }).locator('..')).toContainText(videoURL.replace(/\/$/, ''));
    const original = detail.getByText('保存的原文 / 提取文本', { exact: true }).locator('..').locator('pre');
    await expect(original).not.toHaveText('');
    await expect(original).not.toContainText('此版本未提供可读原文');
    await expect(detail.getByText('处理来源', { exact: true })).toHaveCount(2);
    for (const provenance of await detail.getByText('处理来源', { exact: true }).all()) {
      await expect(provenance.locator('..')).toContainText('summarize');
      await expect(provenance.locator('..')).toContainText(videoURL.replace(/\/$/, ''));
    }
    await detail.getByRole('button', { name: '继续沉淀', exact: true }).click();
    await expect(detail.getByLabel('自动沉淀层次', { exact: true })).toHaveValue('content');
    expect(modelRequests).toEqual([]);
    await page.screenshot({ path: testInfo.outputPath('video-import-content.png'), fullPage: true });
  } finally { await server.close(); }
});
