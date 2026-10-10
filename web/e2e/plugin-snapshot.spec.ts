import { test, expect, type Page } from '@playwright/test';

/**
 * Plugin snapshot paste-review: human pastes the extension's JSON snapshot
 * (W1 shape: source_url / content_state / host_family / text / pdf_urls),
 * reviews metadata, and chooses an explicit import path. The preserved raw
 * body imports through paper_snapshot (no re-fetch); a public PDF imports
 * through paper_pdf (not arxiv-only). A paywall/restricted or abstract-only
 * snapshot without a body yields no import, never a bypass.
 */

const panel = (page: Page) => page.locator('section[aria-label="插件快照复核导入"]');

async function openPanel(page: Page) {
  await page.locator('summary').filter({ hasText: /^插件快照复核导入/ }).click();
  await expect(panel(page)).toBeVisible();
}

test.describe('插件快照复核导入（示例）', () => {
  test('rejects non-JSON and a snapshot without a source_url', async ({ page }) => {
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill('not json');
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByText(/不是有效的 JSON/)).toBeVisible();
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill('{}');
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByText(/缺少来源链接/)).toBeVisible();
  });

  test('reviews metadata and never enables import in fixture mode', async ({ page }) => {
    const writes: string[] = [];
    page.on('request', request => {
      if (request.url().includes('/api/v1/') && !['GET', 'HEAD'].includes(request.method())) writes.push(request.url());
    });
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify({ schema_version: 1, source_url: 'https://arxiv.org/abs/2504.16054', host_family: 'arxiv', title: '快照标题', authors: ['甲'], arxiv_id: '2504.16054', abstract: '快照摘要', content_state: 'abstract_only' }));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByRole('heading', { name: '快照标题' })).toBeVisible();
    await expect(panel(page).getByText(/不是正文/).first()).toBeVisible();
    await expect(panel(page).getByRole('button', { name: '按 arXiv 官方导入', exact: true })).toBeDisabled();
    expect(writes).toEqual([]);
  });
});

test.describe('插件快照复核导入（真实）', () => {
  test('preserves the raw body and offers paper_snapshot without re-fetch', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify({ schema_version: 1, source_url: 'https://journals.plos.org/plosone/article?id=x', host_family: 'plos', title: 'PLOS 论文', source_key: 'doi:10.1/example', content_state: 'readable_fulltext', text: '浏览器提取的完整正文' }));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByRole('heading', { name: 'PLOS 论文' })).toBeVisible();
    await expect(panel(page).getByRole('button', { name: '按快照正文导入（HTML 正文）', exact: true })).toBeVisible();
    await expect(panel(page).getByText(/浏览器提取的正文/)).toBeVisible();
  });

  test('offers a public PDF via paper_pdf for a non-arXiv host', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify({ schema_version: 1, source_url: 'https://journals.plos.org/p', host_family: 'plos', title: 'PDF 论文', content_state: 'abstract_only', pdf_urls: ['https://journals.plos.org/p/file.pdf'] }));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByRole('heading', { name: 'PDF 论文' })).toBeVisible();
    await expect(panel(page).getByRole('button', { name: '按公共 PDF 导入', exact: true })).toBeVisible();
  });

  test('gates a paywalled snapshot honestly without a bypass', async ({ page }) => {
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify({ schema_version: 1, source_url: 'https://example.com/p', host_family: 'generic', title: '付费墙论文', content_state: 'paywall', warning: 'Access restriction detected' }));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByRole('heading', { name: '付费墙论文' })).toBeVisible();
    await expect(panel(page).getByText(/无绕过/)).toBeVisible();
    await expect(panel(page).getByRole('button', { name: /导入/ })).not.toBeVisible();
  });

  test('a body change on the same URL issues a new idempotency key instead of 409', async ({ page }) => {
    const keys: string[] = [];
    await page.route('**/api/v1/materials/imports', route => {
      keys.push((route.request().headers()['idempotency-key'] ?? '').toLowerCase());
      return route.fulfill({ status: 202, contentType: 'application/json', body: JSON.stringify({ schema_version: 1, job_id: 'job-x', status: 'queued' }) });
    });
    await page.goto('/attention');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    const snap = (body: string) => JSON.stringify({ schema_version: 1, source_url: 'https://journals.plos.org/p', host_family: 'plos', title: '正文变化论文', content_state: 'readable_fulltext', text: body });
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(snap('正文A'));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await panel(page).getByRole('button', { name: '按快照正文导入（HTML 正文）', exact: true }).click();
    await expect(panel(page).getByLabel('粘贴插件快照 JSON', { exact: true })).toHaveValue('');
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(snap('正文B'));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await panel(page).getByRole('button', { name: '按快照正文导入（HTML 正文）', exact: true }).click();
    await expect.poll(() => keys.length).toBe(2);
    expect(keys[0]).not.toBe(keys[1]);
  });
});

test.describe('插件快照复核导入（移动布局）', () => {
  test('renders without horizontal overflow at 390px', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto('/attention?fixture=1');
    await page.waitForLoadState('networkidle');
    await openPanel(page);
    await panel(page).getByLabel('粘贴插件快照 JSON', { exact: true }).fill(JSON.stringify({ schema_version: 1, source_url: 'https://arxiv.org/abs/2504.16054', host_family: 'arxiv', title: '窄屏快照标题', authors: ['甲'] }));
    await panel(page).getByRole('button', { name: '复核快照', exact: true }).click();
    await expect(panel(page).getByRole('heading', { name: '窄屏快照标题' })).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  });
});
