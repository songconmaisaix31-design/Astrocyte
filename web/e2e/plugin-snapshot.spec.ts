import { test, expect, type Page } from '@playwright/test';
import { capturedPublicDocument, realService } from './realService';
import { humanAPI, waitJob } from '../../tests/s1/api.mjs';

const panel = (page: Page) => page.locator('section[aria-label="插件快照复核导入"]');
async function open(page: Page, url = '/attention') {
  await page.goto(url);
  await page.locator('summary').filter({ hasText: /^插件快照复核导入/ }).click();
  await expect(panel(page)).toBeVisible();
}
async function review(page: Page, raw: string) {
  await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(raw);
  await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
}

test('实际页面拒绝非 JSON 与缺少来源的输入', async ({ page }) => {
  await open(page);
  await review(page, 'not json');
  await expect(panel(page).getByText(/不是有效的 JSON/)).toBeVisible();
  await review(page, '{}');
  await expect(panel(page).getByText(/缺少来源链接/)).toBeVisible();
});
test('复核真实已捕获公共来源，不自动发送导入', async ({ page }) => {
  const captured = await capturedPublicDocument();
  const writes: string[] = [];
  page.on('request', request => { if (request.url().includes('/api/v1/') && request.method() !== 'GET') writes.push(request.url()); });
  await open(page);
  await review(page, JSON.stringify({ source_url: captured.url, title: captured.title, content_state: 'readable_fulltext', text: captured.text }));
  await expect(panel(page).getByRole('heading', { name: captured.title, exact: true })).toBeVisible();
  await expect(panel(page).getByRole('button', { name: '按快照正文导入（HTML 正文）', exact: true })).toBeEnabled();
  expect(writes).toEqual([]);
});
test('真实已捕获正文保留完整文本，并使用生产 paper_snapshot 入库', async ({ page }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const captured = await capturedPublicDocument();
    const raw = JSON.stringify({ source_url: captured.url, title: captured.title, host_family: 'generic', content_state: 'readable_fulltext', text: captured.text });
    await open(page, `${server.webURL}/attention`);
    await review(page, raw);
    await expect(panel(page).getByRole('heading', { name: captured.title, exact: true })).toBeVisible();
    await expect(panel(page).getByText(/浏览器提取的正文/)).toBeVisible();
    const response = page.waitForResponse(value => value.url().endsWith('/materials/imports') && value.request().method() === 'POST');
    await panel(page).getByRole('button', { name: '按快照正文导入（HTML 正文）', exact: true }).click();
    const accepted = await response;
    expect(accepted.status()).toBe(202);
    expect(accepted.request().postDataJSON().export_text).toBe(raw);
    const receipt = await accepted.json();
    const job = await waitJob(api, receipt.job_id);
    const content = await api.get(`/materials/${job.material_id}/revisions/1/content`);
    expect(content.text).toBe(captured.text);
    expect(content.provenance.processor).toBe('paper_snapshot');
    expect(content.provenance.source).toBe(captured.url);
  } finally { await server.close({ preserveData: true }); }
});
test('公共 PDF 的非 arXiv 选择逻辑保留为明确参数契约', async () => {
  const { snapshotImportOptions, parsePluginSnapshot } = await import('../src/pages/attention/pluginSnapshot');
  // 本测试不声明 PDF 来源已获取，只验证调用者提供的地址如何进入导入参数。
  const raw = JSON.stringify({ source_url: 'https://example.com/input', host_family: 'generic', content_state: 'abstract_only', pdf_urls: ['https://example.com/input.pdf'] });
  expect(snapshotImportOptions(parsePluginSnapshot(raw), raw)).toEqual([{ adapter: 'paper_pdf', source_locator: 'https://example.com/input.pdf', export_text: null, label: '按公共 PDF 导入' }]);
});
test('付费和受限状态明确输入时禁止构造导入选项', async () => {
  const { snapshotImportOptions, parsePluginSnapshot } = await import('../src/pages/attention/pluginSnapshot');
  for (const content_state of ['paywall', 'restricted']) {
    const raw = JSON.stringify({ source_url: 'https://example.com/input', host_family: 'generic', content_state });
    expect(snapshotImportOptions(parsePluginSnapshot(raw), raw)).toEqual([]);
  }
});
test('修改人工输入标题产生新命令身份，实际服务保留原始正文', async ({ page }) => {
  test.setTimeout(120_000);
  const { server, api } = await realService();
  try {
    const captured = await capturedPublicDocument();
    await open(page, `${server.webURL}/attention`);
    const keys: string[] = [];
    const jobs: string[] = [];
    for (const title of [captured.title, `人工复核标题 · ${captured.title}`]) {
      const raw = JSON.stringify({ source_url: captured.url, host_family: 'generic', title, content_state: 'readable_fulltext', text: captured.text });
      await review(page, raw);
      const response = page.waitForResponse(value => value.url().endsWith('/materials/imports') && value.request().method() === 'POST');
      await panel(page).getByRole('button', { name: '按快照正文导入（HTML 正文）', exact: true }).click();
      const accepted = await response;
      expect(accepted.status()).toBe(202);
      keys.push(accepted.request().headers()['idempotency-key']);
      jobs.push((await accepted.json()).job_id);
      await expect(panel(page).getByLabel('粘贴插件快照 JSON', { exact: true })).toHaveValue('');
      await waitJob(api, jobs.at(-1)!);
    }
    expect(keys).toHaveLength(2);
    expect(keys[0]).not.toBe(keys[1]);
    expect((await api.get('/materials')).items).toHaveLength(1);
    await server.restart();
    const reconnected = await humanAPI(server.apiURL);
    expect((await reconnected.get('/materials')).items).toHaveLength(1);
  } finally { await server.close({ preserveData: true }); }
});
test('390px 显示实际已捕获来源，复核不产生水平溢出', async ({ page }) => {
  const captured = await capturedPublicDocument();
  await page.setViewportSize({ width: 390, height: 844 });
  await open(page);
  await review(page, JSON.stringify({ source_url: captured.url, title: captured.title, content_state: 'readable_fulltext', text: captured.text }));
  await expect(panel(page).getByRole('heading', { name: captured.title, exact: true })).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});
