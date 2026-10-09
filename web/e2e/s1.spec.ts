import { test, expect } from '@playwright/test';
import { startS1Server } from '../../tests/s1/server.mjs';
import { humanAPI, waitJob } from '../../tests/s1/api.mjs';
import { importMaterial } from '../../tests/s1/api.mjs';
import { changedPaperText, paperImport, paperText, videoImport, videoTranscript } from '../../tests/s1/fixtures.mjs';
import { seedCandidate } from '../../tests/s1/scenarios.mjs';

// Each viewport owns its own API/Vite/data directory. Existing S0 tests stay empty.
// Synthetic contract_local inputs use real HTTP/SQLite/objects; no page.route mocks.
test.describe('S1 contract_local with real API and persistence', () => {
  let server: Awaited<ReturnType<typeof startS1Server>>;
  let api: Awaited<ReturnType<typeof humanAPI>>;
  test.beforeAll(async () => {
    test.setTimeout(120_000);
    server = await startS1Server({ browser: true });
    api = await humanAPI(server.apiURL);
  });
  test.afterAll(async () => { if (server) await server.close(); });

  test('fresh browser bootstraps a human session and settles actual empty collections', async ({ page }) => {
    const materialsResponse = page.waitForResponse(response => response.url() === `${server.webURL}/api/v1/materials`);
    await page.goto(`${server.webURL}/attention`);
    expect((await materialsResponse).status()).toBe(200);
    await expect(page.getByRole('heading', { name: '资料沉淀', exact: true })).toBeVisible();
    await expect(page.getByText('暂无素材', { exact: true })).toBeVisible();
    await expect(page.getByText('暂无机会', { exact: true })).toBeVisible();
    await expect(page.getByRole('alert').filter({ hasText: '示例数据模式' })).toHaveCount(0);
    expect((await api.get('/materials')).items).toHaveLength(0);
  });

  test('browser imports a summarize representative export and exposes its retained positions', async ({ page }, testInfo) => {
    const input = videoImport('https://youtu.be/S1LOCAL005a');
    await page.goto(`${server.webURL}/attention`);
    await page.getByRole('main').getByRole('button', { name: '添加资料', exact: true }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('导入方式', { exact: true }).selectOption('summarize');
    await dialog.getByLabel(/视频原始来源/).fill(input.source_locator);
    await dialog.getByLabel(/summarize 导出内容/).fill(input.export_text);
    const receiptResponse = page.waitForResponse(response => response.url().endsWith('/api/v1/materials/imports') && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '导入资料', exact: true }).click();
    const response = await receiptResponse;
    expect(response.status()).toBe(202);
    const receipt = await response.json();
    const job = await waitJob(api, receipt.job_id);
    const detail = await api.get(`/materials/${job.material_id}`);
    expect(detail.material.collection_reason).toBeNull();
    await dialog.getByRole('button', { name: '关闭', exact: true }).click();
    await page.getByRole('button', { name: '↻ 刷新', exact: true }).click();
    const card = page.getByRole('button').filter({ has: page.getByText('contract_local 视频代表导出', { exact: true }) });
    await expect(card).toBeVisible();
    await card.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('dialog')).toContainText(videoTranscript);
    await expect(page.getByRole('dialog')).toContainText('0.000s');
    await expect(page.getByRole('dialog').getByRole('link', { name: /summarize.json/ })).toBeVisible();
    await page.screenshot({ path: testInfo.outputPath('s1-import-source.png'), fullPage: true });
    expect((await api.get('/missions')).items).toHaveLength(0);
  });

  test('view and refresh stay neutral; an explicit keyboard reread increments human attention once', async ({ page }, testInfo) => {
    const input = { ...paperImport('https://example.invalid/contract-local/browser-reread'), title: 'contract_local 浏览器重读材料' };
    const imported = await importMaterial(api, input);
    await importMaterial(api, { ...input, export_text: changedPaperText });
    const id = imported.detail.material.id;
    const baseline = await api.get(`/materials/${id}`);
    const humanCount = baseline.material.human_usage_count;
    if (humanCount === undefined) throw new Error('S1 material omitted its human usage count');
    await page.goto(`${server.webURL}/attention`);
    await page.getByRole('button', { name: '↻ 刷新', exact: true }).click();
    const card = page.getByRole('button').filter({ has: page.getByText(input.title, { exact: true }) });
    await card.click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByLabel('原文版本', { exact: true })).toHaveValue('2');
    await dialog.getByLabel('原文版本', { exact: true }).selectOption('1');
    await expect(dialog.locator('pre').filter({ hasText: paperText })).toHaveText(paperText);
    const detailResponse = page.waitForResponse(response => response.url().endsWith(`/api/v1/materials/${id}`));
    await dialog.getByRole('button', { name: '刷新资料详情', exact: true }).click();
    expect((await detailResponse).status()).toBe(200);
    expect((await api.get(`/materials/${id}`)).material.human_usage_count).toBe(baseline.material.human_usage_count);
    const rereadResponse = page.waitForResponse(response => response.url().endsWith(`/api/v1/materials/${id}/uses`) && response.request().method() === 'POST');
    const reread = dialog.getByRole('button', { name: '主动重读', exact: true });
    await expect(reread).toBeEnabled();
    await reread.focus();
    await page.keyboard.press('Enter');
    expect((await rereadResponse).status()).toBe(200);
    await expect.poll(async () => (await api.get(`/materials/${id}`)).material.human_usage_count).toBe(humanCount + 1);
    await page.screenshot({ path: testInfo.outputPath('s1-human-reread.png'), fullPage: true });
  });

  test('AT04: later records the displayed candidate revision and survives browser/service reconnection', async ({ page }, testInfo) => {
    const seeded = await seedCandidate(api, 'https://example.invalid/contract-local/browser-later');
    const candidate = seeded.candidate.opportunity;
    if (!candidate.title) throw new Error('S1 candidate omitted its title');
    await page.goto(`${server.webURL}/attention`);
    const card = page.locator('li[role="button"]').filter({ has: page.getByText(candidate.title, { exact: true }) });
    await card.click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('反馈类型', { exact: true }).selectOption('later');
    await dialog.getByLabel(/反馈理由/).fill('contract_local 以后再做；保留资料待真实验证');
    const reviewResponse = page.waitForResponse(response => response.url().endsWith(`/api/v1/opportunities/${candidate.id}/reviews`) && response.request().method() === 'POST');
    await dialog.getByRole('button', { name: '保存人工反馈', exact: true }).click();
    expect((await reviewResponse).status()).toBe(201);
    await expect.poll(async () => (await api.get(`/opportunities/${candidate.id}`)).opportunity.state).toBe('deferred');
    const saved = await api.get(`/opportunities/${candidate.id}`);
    expect(saved.reviews).toHaveLength(1);
    expect(saved.reviews[0]).toMatchObject({ revision: candidate.revision, feedback: 'later' });
    expect(saved.reviews.some((review: { feedback: string }) => review.feedback === 'reject')).toBe(false);
    expect((await api.get(`/materials/${seeded.detail.material.id}`)).material.lifecycle).toBe('active');
    await page.screenshot({ path: testInfo.outputPath('s1-later-feedback.png'), fullPage: true });
    await server.restart();
    api = await humanAPI(server.apiURL);
    const reconnectResponse = page.waitForResponse(response => response.url() === `${server.webURL}/api/v1/opportunities`);
    await page.reload();
    expect((await reconnectResponse).status()).toBe(200);
    await card.click();
    await expect(page.getByRole('dialog')).toContainText('contract_local 以后再做；保留资料待真实验证');
    expect(await api.get(`/opportunities/${candidate.id}`)).toEqual(saved);
    expect((await api.get(`/materials/${seeded.detail.material.id}/revisions/1/content`)).text).toBe(paperText);
    expect((await api.get('/missions')).items).toHaveLength(0);
  });
});
